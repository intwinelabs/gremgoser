package gremgoser

import (
	"errors"
	"fmt"
	"log"
	"reflect"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Client is a container for the gremgoser client.
type Client struct {
	conn             dialer
	requests         chan []byte
	responses        chan []byte
	results          *sync.Map
	responseNotifier *sync.Map // responseNotifier notifies the requester that a response has arrived for the request
	respMutex        *sync.Mutex
	Errored          bool
}

// NewDialer returns a WebSocket dialer to use when connecting to Gremlin Server
func NewDialer(host string, configs ...DialerConfig) (dialer *Ws) {
	dialer = &Ws{
		timeout:      5 * time.Second,
		pingInterval: 60 * time.Second,
		writingWait:  15 * time.Second,
		readingWait:  15 * time.Second,
		connected:    false,
		quit:         make(chan struct{}),
	}

	for _, conf := range configs {
		conf(dialer)
	}

	dialer.host = host
	return dialer
}

func newClient() (c Client) {
	c.requests = make(chan []byte, 3)  // c.requests takes any request and delivers it to the WriteWorker for dispatch to Gremlin Server
	c.responses = make(chan []byte, 3) // c.responses takes raw responses from ReadWorker and delivers it for sorting to handelResponse
	c.results = &sync.Map{}
	c.responseNotifier = &sync.Map{}
	c.respMutex = &sync.Mutex{} // c.mutex ensures that sorting is thread safe
	return
}

// Dial returns a gremgoser client for interaction with the Gremlin Server specified in the host IP.
func Dial(conn dialer, errs chan error) (c Client, err error) {
	c = newClient()
	c.conn = conn

	// Connects to Gremlin Server
	err = conn.connect()
	if err != nil {
		return
	}

	quit := conn.(*Ws).quit

	go c.writeWorker(errs, quit)
	go c.readWorker(errs, quit)
	go conn.ping(errs)

	return
}

func (c *Client) executeRequest(query string, bindings, rebindings map[string]string) (resp interface{}, err error) {
	req, id, err := prepareRequest(query, bindings, rebindings)
	if err != nil {
		return
	}

	msg, err := packageRequest(req)
	if err != nil {
		log.Println(err)
		return
	}
	c.responseNotifier.Store(id, make(chan int, 1))
	c.dispatchRequest(msg)
	resp = c.retrieveResponse(id)
	return
}

func (c *Client) authenticate(requestId string) (err error) {
	auth, err := c.conn.getAuth()
	if err != nil {
		return err
	}
	req, err := prepareAuthRequest(requestId, auth.username, auth.password)
	if err != nil {
		return
	}

	msg, err := packageRequest(req)
	if err != nil {
		log.Println(err)
		return
	}

	c.dispatchRequest(msg)
	return
}

// Execute formats a raw Gremlin query, sends it to Gremlin Server, and returns the result.
func (c *Client) Execute(query string, bindings, rebindings map[string]string) (resp interface{}, err error) {
	if c.conn.isDisposed() {
		return nil, errors.New("you cannot write on a disposed connection")
	}
	resp, err = c.executeRequest(query, bindings, rebindings)
	return
}

// Get formats a raw Gremlin query, sends it to Gremlin Server, and populates the passed []interface.
func (c *Client) Get(query string, ptr interface{}) (err error) {
	if c.conn.isDisposed() {
		return errors.New("you cannot write on a disposed connection")
	}

	var strct reflect.Value
	if reflect.ValueOf(ptr).Kind() != reflect.Ptr {
		return errors.New("the passed interface is not a ptr")
	} else if reflect.ValueOf(ptr).Elem().Kind() != reflect.Slice {
		return errors.New("the passed interface is not a slice")
	} else {
		strct = reflect.ValueOf(ptr).Elem()
	}

	resp, err := c.executeRequest(query, nil, nil)
	if err != nil {
		return err
	}
	// get the underlying struct type
	sType := reflect.TypeOf(strct.Interface()).Elem()

	// test if we can cast resp as slice of interface
	if respSlice, ok := resp.([]interface{}); ok {
		// iterate of the interface in the slice
		for _, item := range respSlice {
			// create new slice to later copy back
			lenRespSlice := len(respSlice)
			sSlice := reflect.MakeSlice(reflect.SliceOf(sType), lenRespSlice, lenRespSlice+1)
			// test if we can cast resp as slice of interface
			if respSliceSlice, ok := item.([]interface{}); ok {
				// iterate over the inner slice
				for j, innerItem := range respSliceSlice {
					// check if we can cast as a map
					if respMap, ok := innerItem.(map[string]interface{}); ok {
						// create a new struct to populate
						s := reflect.New(sType)
						// check for Id field
						ok := s.Elem().FieldByName("Id").IsValid()
						if !ok {
							return errors.New("the passed interface must have an Id field")
						}
						uuidType := s.Elem().FieldByName("Id").Kind()
						// iterate over fields and populate
						for i := 0; i < s.Elem().NumField(); i++ {
							// get graph tag for field
							tag := sType.Field(i).Tag.Get("graph")
							name, opts := parseTag(tag)
							if len(name) == 0 && len(opts) == 0 {
								continue
							}

							// get the current field
							f := s.Elem().Field(i)

							// check if we can modify
							if f.CanSet() {
								if f.Kind() == uuidType { // if its the Id field we look in the base response map
									// create a UUID
									id, err := uuid.Parse(respMap[name].(string))
									if err != nil {
										return err
									}
									f.Set(reflect.ValueOf(id))
								} else { // it is a property and we have to looks at the properties map
									props, ok := respMap["properties"].(map[string]interface{})
									if ok {
										// check if the key is in the map
										if _, ok := props[name]; ok {
											// get the properties slice
											propSlice := reflect.ValueOf(props[name])
											propSliceLen := propSlice.Len()
											// check the length, if its 1 we set it as a single value otherwise we need to create a slice
											if propSliceLen == 1 {
												// get the value of the property we are looking for
												v, err := getPropertyValue(propSlice.Index(0).Interface())
												if err != nil {
													return err
												}
												if f.Kind() == reflect.String { // Set as string
													_v, ok := v.(string)
													if ok {
														f.SetString(_v)
													}
												} else if f.Kind() == reflect.Int || f.Kind() == reflect.Int8 || f.Kind() == reflect.Int16 || f.Kind() == reflect.Int32 || f.Kind() == reflect.Int64 { // Set as int
													_v, ok := v.(float64)
													__v := int64(_v)
													if ok {

														if !f.OverflowInt(__v) {
															f.SetInt(__v)
														}
													}
												} else if f.Kind() == reflect.Float32 || f.Kind() == reflect.Float64 { // Set as float
													_v, ok := v.(float64)
													if ok {
														if !f.OverflowFloat(_v) {
															f.SetFloat(_v)
														}
													}
												} else if f.Kind() == reflect.Uint || f.Kind() == reflect.Uint8 || f.Kind() == reflect.Uint16 || f.Kind() == reflect.Uint32 || f.Kind() == reflect.Uint64 { // Set as uint
													_v, ok := v.(float64)
													__v := uint64(_v)
													if ok {

														if !f.OverflowUint(__v) {
															f.SetUint(__v)
														}
													}
												} else if f.Kind() == reflect.Bool { // Set as bool
													_v, ok := v.(bool)
													if ok {
														f.SetBool(_v)
													}
												}
											} else if propSliceLen > 1 { // we need to creates slices for the properties
												pSlice := reflect.MakeSlice(reflect.SliceOf(f.Type().Elem()), propSliceLen, propSliceLen+1)
												// now we iterate over the properties
												for i := 0; i < propSliceLen; i++ {
													// get the value of the property we are looking for
													v, err := getPropertyValue(propSlice.Index(i).Interface())
													if err != nil {
														return err
													}
													if f.Type().Elem().Kind() == reflect.String { // Set as string
														_v, ok := v.(string)
														if ok {
															pSlice.Index(i).SetString(_v)
														}
													} else if f.Type().Elem().Kind() == reflect.Int || f.Type().Elem().Kind() == reflect.Int8 || f.Type().Elem().Kind() == reflect.Int16 || f.Type().Elem().Kind() == reflect.Int32 || f.Type().Elem().Kind() == reflect.Int64 { // Set as int
														_v, ok := v.(float64)
														__v := int64(_v)
														if ok {

															if !pSlice.Index(i).OverflowInt(__v) {
																pSlice.Index(i).SetInt(__v)
															}
														}
													} else if f.Type().Elem().Kind() == reflect.Float32 || f.Type().Elem().Kind() == reflect.Float64 { // Set as float
														_v, ok := v.(float64)
														if ok {
															if !pSlice.Index(i).OverflowFloat(_v) {
																pSlice.Index(i).SetFloat(_v)
															}
														}
													} else if f.Type().Elem().Kind() == reflect.Uint || f.Type().Elem().Kind() == reflect.Uint8 || f.Type().Elem().Kind() == reflect.Uint16 || f.Type().Elem().Kind() == reflect.Uint32 || f.Type().Elem().Kind() == reflect.Uint64 { // Set as uint
														_v, ok := v.(float64)
														__v := uint64(_v)
														if ok {

															if !pSlice.Index(i).OverflowUint(__v) {
																pSlice.Index(i).SetUint(__v)
															}
														}
													} else if f.Type().Elem().Kind() == reflect.Bool { // Set as bool
														_v, ok := v.(bool)
														if ok {
															pSlice.Index(i).SetBool(_v)
														}
													}
												}
												// set the field to the created slice
												f.Set(pSlice)
											}
										}
									}
								}
							}
						}
						// update slice
						sSlice.Index(j).Set(s.Elem())
					}
				}
				// Copy the new slice to the passed data slice
				strct.Set(sSlice)
			}
		}
	}
	return
}

// getProprtyValue takes a property map slice and return the value
func getPropertyValue(prop interface{}) (interface{}, error) {
	propMap, ok := prop.(map[string]interface{})
	if ok {
		return propMap["value"], nil
	}
	return nil, errors.New("passed property cannot be cast")
}

// Close closes the underlying connection and marks the client as closed.
func (c *Client) Close() {
	if c.conn != nil {
		c.conn.close()
	}
}

// AddV takes a label and a interface and adds it a vertex to the graph
func (c *Client) AddV(label string, data interface{}) (resp interface{}, err error) {
	if c.conn.isDisposed() {
		return nil, errors.New("you cannot write on a disposed connection")
	}

	d := reflect.ValueOf(data)

	id := d.FieldByName("Id")
	if !id.IsValid() {
		return nil, errors.New("the passed interface must have an Id field")
	}

	q := fmt.Sprintf("g.addV('%s')", label)

	tagLength := 0

	for i := 0; i < d.NumField(); i++ {
		tag := d.Type().Field(i).Tag.Get("graph")
		name, opts := parseTag(tag)
		if len(name) == 0 && len(opts) == 0 {
			continue
		}
		tagLength++
		val := d.Field(i).Interface()
		if len(opts) == 0 {
			return nil, fmt.Errorf("interface field tag does not contain a tag option type, field type: %T", val)
		} else if opts.Contains("string") {
			q = fmt.Sprintf("%s.property('%s', '%s')", q, name, val)
		} else if opts.Contains("bool") || opts.Contains("number") {
			q = fmt.Sprintf("%s.property('%s', %v)", q, name, val)
		} else if opts.Contains("[]string") {
			s := reflect.ValueOf(val)
			for i := 0; i < s.Len(); i++ {
				q = fmt.Sprintf("%s.property('%s', '%s')", q, name, s.Index(i).Interface())
			}
		} else if opts.Contains("[]bool") || opts.Contains("[]number") {
			s := reflect.ValueOf(val)
			for i := 0; i < s.Len(); i++ {
				q = fmt.Sprintf("%s.property('%s', %v)", q, name, s.Index(i).Interface())
			}
		}
	}

	if tagLength == 0 {
		return nil, fmt.Errorf("interface of type: %T, does not contain any graph tags", data)
	}

	resp, err = c.Execute(q, nil, nil)
	return

}

// AddE takes a label, from UUID and to UUID then creates a edge between the two vertex in the graph
func (c *Client) AddE(label string, from, to interface{}) (resp interface{}, err error) {
	if c.conn.isDisposed() {
		return nil, errors.New("you cannot write on a disposed connection")
	}

	df := reflect.ValueOf(from)
	fid := df.FieldByName("Id")
	if !fid.IsValid() {
		return nil, errors.New("the passed from interface must have an Id field")
	}

	dt := reflect.ValueOf(to)
	tid := dt.FieldByName("Id")
	if !tid.IsValid() {
		return nil, errors.New("the passed to interface must have an Id field")
	}

	q := fmt.Sprintf("g.V('%s').addE('%s').to(g.V('%s'))", fid.Interface().(uuid.UUID).String(), label, tid.Interface().(uuid.UUID).String())
	resp, err = c.Execute(q, map[string]string{}, map[string]string{})

	return
}

// AddEById takes a label, from UUID and to UUID then creates a edge between the two vertex in the graph
func (c *Client) AddEById(label string, from, to uuid.UUID) (resp interface{}, err error) {
	if c.conn.isDisposed() {
		return nil, errors.New("you cannot write on a disposed connection")
	}

	q := fmt.Sprintf("g.V('%s').addE('%s').to(g.V('%s'))", from.String(), label, to.String())
	resp, err = c.Execute(q, map[string]string{}, map[string]string{})

	return
}

// AddEWithProps takes a label, from UUID and to UUID then creates a edge between the two vertex in the graph
func (c *Client) AddEWithProps(label string, from, to interface{}, props map[string]interface{}) (resp interface{}, err error) {
	if c.conn.isDisposed() {
		return nil, errors.New("you cannot write on a disposed connection")
	}

	df := reflect.ValueOf(from)
	fid := df.FieldByName("Id")
	if !fid.IsValid() {
		return nil, errors.New("the passed from interface must have an Id field")
	}

	dt := reflect.ValueOf(to)
	tid := dt.FieldByName("Id")
	if !tid.IsValid() {
		return nil, errors.New("the passed to interface must have an Id field")
	}

	q := fmt.Sprintf("g.V('%s').addE('%s').to(g.V('%s'))", fid.Interface().(uuid.UUID).String(), label, tid.Interface().(uuid.UUID).String())
	p, err := buildProps(props)
	if err != nil {
		return
	}
	q = q + p
	resp, err = c.Execute(q, map[string]string{}, map[string]string{})

	return
}

// AddEWithPropsById takes a label, from UUID and to UUID then creates a edge between the two vertex in the graph
func (c *Client) AddEWithPropsById(label string, from, to uuid.UUID, props map[string]interface{}) (resp interface{}, err error) {
	if c.conn.isDisposed() {
		return nil, errors.New("you cannot write on a disposed connection")
	}

	q := fmt.Sprintf("g.V('%s').addE('%s').to(g.V('%s'))", from.String(), label, to.String())
	p, err := buildProps(props)
	if err != nil {
		return
	}
	q = q + p
	resp, err = c.Execute(q, map[string]string{}, map[string]string{})

	return
}

func buildProps(props map[string]interface{}) (string, error) {
	q := ""

	for k, v := range props {
		t := reflect.ValueOf(v).Kind()
		if t == reflect.String {
			q = fmt.Sprintf("%s.property('%s', '%s')", q, k, v)
		} else if t == reflect.Bool || t == reflect.Int || t == reflect.Int8 || t == reflect.Int16 || t == reflect.Int32 || t == reflect.Int64 || t == reflect.Uint || t == reflect.Uint8 || t == reflect.Uint16 || t == reflect.Uint32 || t == reflect.Uint64 || t == reflect.Float32 || t == reflect.Float64 {
			q = fmt.Sprintf("%s.property('%s', %v)", q, k, v)
		} else if t == reflect.Slice {
			s := reflect.ValueOf(v)
			for i := 0; i < s.Len(); i++ {
				if _, ok := s.Index(i).Interface().(string); ok {
					q = fmt.Sprintf("%s.property('%s', '%s')", q, k, s.Index(i).Interface())
				} else {
					q = fmt.Sprintf("%s.property('%s', %v)", q, k, s.Index(i).Interface())
				}
			}
		} else {
			return "", errors.New("unsupported property map")
		}
	}

	return q, nil

}
