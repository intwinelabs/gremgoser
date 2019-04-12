package gremgoser

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/intwinelabs/logger"

	"github.com/google/uuid"
)

func newClient(conf *ClientConfig) *Client {
	return &Client{
		conf:             conf,
		requests:         make(chan []byte, 3), // c.requests takes any request and delivers it to the WriteWorker for dispatch to Gremlin Server
		responses:        make(chan []byte, 3), // c.responses takes raw responses from ReadWorker and delivers it for sorting to handelResponse
		results:          &sync.Map{},
		responseNotifier: &sync.Map{},
		respMutex:        &sync.Mutex{}, // c.mutex ensures that sorting is thread safe
	}
}

// NewClient returns a gremgoser client for interaction with the Gremlin Server specified in the host IP.
func NewClient(conf *ClientConfig) (*Client, chan error) {
	errs := make(chan error)

	if conf.URI == "" {
		errs <- ErrorInvalidURI
		return nil, errs
	}

	c := newClient(conf)
	c.errs = errs

	ws := &Ws{
		debug:     conf.Debug,
		verbose:   conf.Verbose,
		uri:       conf.URI,
		connected: false,
		quit:      make(chan struct{}),
		logger:    conf.Logger,
	}

	// check for configs
	if conf.Logger == nil {
		conf.Logger = logger.New()
	}
	if conf.Timeout != 0 {
		ws.timeout = conf.Timeout
	} else {
		ws.timeout = 300 * time.Second
	}
	if conf.PingInterval != 0 {
		ws.pingInterval = conf.PingInterval
	} else {
		ws.pingInterval = 60 * time.Second
	}
	if conf.WritingWait != 0 {
		ws.writingWait = conf.WritingWait
	} else {
		ws.writingWait = 120 * time.Second
	}
	if conf.ReadingWait != 0 {
		ws.readingWait = conf.ReadingWait
	} else {
		ws.readingWait = 120 * time.Second
	}
	c.conn = ws
	c.conf = conf

	// Connects to Gremlin Server
	err := c.conn.connect()
	if err != nil {
		c.debug("error connecting to %s: %s", conf.URI, err)
		errs <- ErrorWSConnection
		return nil, errs
	}

	quit := c.conn.(*Ws).quit

	go c.writeWorker(c.errs, quit)
	go c.readWorker(c.errs, quit)
	go c.conn.ping(c.errs)

	// Force authentication
	if c.conf.AuthReq != nil {
		c.Execute("g.V('__FORCE____AUTH__')", nil, nil)
	}

	return c, c.errs
}

// Reconnect trys to reconnect the underlying ws connection
func (c *Client) Reconnect() {
	if !c.conn.isConnected() {
		err := c.conn.connect()
		if err != nil {
			c.errs <- err
		} else {
			// Force authentication
			if c.conf.AuthReq != nil {
				c.Execute("g.V('__FORCE____AUTH__')", nil, nil)
			}
		}
	}
}

// IsConnected return bool
func (c *Client) IsConnected() bool {
	return c.conn.isConnected()
}

// debug prints to the configured logger if debug is enabled
func (c *Client) debug(frmt string, i ...interface{}) {
	if c.conf.Debug {
		c.conf.Logger.InfoDepth(1, fmt.Sprintf("GREMGOSER: DEBUG: "+frmt, i...))
	}
}

// verbose prints to the configured logger if debug is enabled
func (c *Client) verbose(frmt string, i ...interface{}) {
	if c.conf.Verbose {
		c.conf.Logger.InfoDepth(1, fmt.Sprintf("GREMGOSER: VERBOSE: "+frmt, i...))
	}
}

func (c *Client) executeRequest(query string, bindings, rebindings map[string]string) ([]*GremlinRespData, error) {
	req := prepareRequest(query, bindings, rebindings)
	msg, err := packageRequest(req)
	if err != nil {
		c.debug("error packing request: %s", err)
		return nil, err
	}
	c.debug("packed request: %+v", req)
	id := req.RequestId
	c.responseNotifier.Store(id, make(chan int, 1))
	c.dispatchRequest(msg)
	resp := c.retrieveResponse(id)
	return resp, nil
}

func (c *Client) authenticate(requestId uuid.UUID) (err error) {
	c.conf.AuthReq.RequestId = requestId
	msg, err := packageRequest(c.conf.AuthReq)
	if err != nil {
		c.debug("error authenticating to ws server: %s", err)
		return err
	}
	c.dispatchRequest(msg)
	return err
}

// Execute formats a raw Gremlin query, sends it to Gremlin Server, and returns the result.
func (c *Client) Execute(query string, bindings, rebindings map[string]string) ([]*GremlinRespData, error) {
	c.verbose("connection: %+v", c.conn)
	if c.conn.isDisposed() {
		return nil, ErrorConnectionDisposed
	}
	c.verbose("query: %s", query)
	resp, err := c.executeRequest(query, bindings, rebindings)
	c.verbose("response: %+v", resp)
	return resp, err
}

// Get formats a raw Gremlin query, sends it to Gremlin Server, and populates the passed []interface.
func (c *Client) Get(query string, bindings map[string]string, ptr interface{}) error {
	if c.conn.isDisposed() {
		return ErrorConnectionDisposed
	}
	var strct reflect.Value
	if reflect.ValueOf(ptr).Kind() != reflect.Ptr {
		return errors.New("the passed interface is not a ptr")
	} else if reflect.ValueOf(ptr).Elem().Kind() != reflect.Slice {
		return errors.New("the passed interface is not a slice")
	} else {
		strct = reflect.ValueOf(ptr).Elem()
	}

	var respSlice []*GremlinData
	respDataSlice, err := c.executeRequest(query, bindings, nil)
	if err != nil {
		return err
	}

	// if the return is empty return
	if len(respDataSlice) == 0 {
		return nil
	}

	// if the returndata is GraphSON cast to GremlinData
	// we try to marshal the response data slice
	obj, err := json.Marshal(respDataSlice)
	if err != nil {
		c.debug("err marshaling resp data slice: %s", err)
		return nil
	}
	if _, ok := (*respDataSlice[0])["properties"]; ok {
		err := json.Unmarshal(obj, &respSlice)
		if err != nil {
			c.debug("err unmarshaling response slice: %s", err)
			return err
		}
	} else {
		err := json.Unmarshal(obj, &ptr)
		if err != nil {
			c.debug("err unmarshaling response slice: %s", err)
			return err
		}
	}

	// get the underlying struct type
	sType := reflect.TypeOf(strct.Interface()).Elem()

	// create new slice to later copy back
	lenRespSlice := len(respSlice)
	sSlice := reflect.MakeSlice(reflect.SliceOf(sType), lenRespSlice, lenRespSlice+1)
	// iterate over the GremlinData respSlice
	for j, innerItem := range respSlice {
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
					f.Set(reflect.ValueOf(innerItem.Id))
				} else { // it is a property and we have to looks at the properties map
					props := innerItem.Properties
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
							} else if f.Kind() == reflect.Struct { // take JSON string and unmarshal into struct
								_v, ok := v.(string)
								if ok {
									s := reflect.New(f.Type()).Interface()
									json.Unmarshal([]byte(_v), s)
									f.Set(reflect.ValueOf(s).Elem())
								}
							} else if f.Kind() == reflect.Slice { // take JSON string and unmarshal into slice
								_v, ok := v.(string)
								if ok {
									sSlice := reflect.SliceOf(f.Type().Elem())
									s := reflect.New(sSlice)
									json.Unmarshal([]byte(_v), s.Interface())
									f.Set(s.Elem())
								}
							} else if f.Kind() == reflect.Ptr {
								return errors.New("gremgoser does not currently support root level pointers")
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
		// update slice
		sSlice.Index(j).Set(s.Elem())
	}
	// Copy the new slice to the passed data slice
	strct.Set(sSlice)

	return nil
}

// Close closes the underlying connection and marks the client as closed.
func (c *Client) Close() {
	if c.conn != nil {
		c.conn.close()
	}
}

// AddV takes a label and a interface and adds it a vertex to the graph
func (c *Client) AddV(label string, data interface{}) ([]*GremlinRespData, error) {
	c.verbose("passed interface: %s", spew.Sdump(data))
	if c.conn.isDisposed() {
		return nil, ErrorConnectionDisposed
	}
	d := getValue(data)

	id := d.FieldByName("Id")
	if !id.IsValid() {
		return nil, ErrorInterfaceHasNoIdField
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
			return nil, fmt.Errorf("gremgoser: interface field graph tag does not contain a tag option type, field type: %T", val)
		} else if opts.Contains("string") {
			q = fmt.Sprintf("%s.property('%s', '%s')", q, name, escapeString(fmt.Sprintf("%s", val)))
		} else if opts.Contains("bool") || opts.Contains("number") {
			q = fmt.Sprintf("%s.property('%s', %v)", q, name, val)
		} else if opts.Contains("struct") || opts.Contains("[]struct") {
			jsonBytes, err := json.Marshal(val)
			if err != nil {
				return nil, err
			}
			q = fmt.Sprintf("%s.property('%s', '%s')", q, name, jsonBytes)
		} else if opts.Contains("[]string") {
			s := reflect.ValueOf(val)
			for i := 0; i < s.Len(); i++ {
				q = fmt.Sprintf("%s.property('%s', '%s')", q, name, escapeString(fmt.Sprintf("%s", s.Index(i).Interface())))
			}
		} else if opts.Contains("[]bool") || opts.Contains("[]number") {
			s := reflect.ValueOf(val)
			for i := 0; i < s.Len(); i++ {
				q = fmt.Sprintf("%s.property('%s', %v)", q, name, s.Index(i).Interface())
			}
		}
	}

	if tagLength == 0 {
		return nil, ErrorInterfaceHasNoIdField
	}

	return c.Execute(q, nil, nil)
}

// UpdateV takes a interface and updates the vertex in the graph
func (c *Client) UpdateV(data interface{}) ([]*GremlinRespData, error) {
	c.verbose("passed interface: %s", spew.Sdump(data))
	if c.conn.isDisposed() {
		return nil, ErrorConnectionDisposed
	}
	d := getValue(data)

	id := d.FieldByName("Id")
	if !id.IsValid() {
		return nil, ErrorInterfaceHasNoIdField
	}

	q := fmt.Sprintf("g.V('%s')", id)

	tagLength := 0

	for i := 0; i < d.NumField(); i++ {
		tag := d.Type().Field(i).Tag.Get("graph")
		name, opts := parseTag(tag)
		if (len(name) == 0 && len(opts) == 0) || d.Type().Field(i).Name == "Id" {
			continue
		}
		tagLength++
		val := d.Field(i).Interface()
		if len(opts) == 0 {
			return nil, fmt.Errorf("gremgoser: interface field graph tag does not contain a tag option type, field type: %T", val)
		} else if opts.Contains("string") {
			q = fmt.Sprintf("%s.property('%s', '%s')", q, name, escapeString(fmt.Sprintf("%s", val)))
		} else if opts.Contains("bool") || opts.Contains("number") {
			q = fmt.Sprintf("%s.property('%s', %v)", q, name, val)
		} else if opts.Contains("struct") || opts.Contains("[]struct") {
			jsonBytes, err := json.Marshal(val)
			if err != nil {
				return nil, err
			}
			q = fmt.Sprintf("%s.property('%s', '%s')", q, name, jsonBytes)
		} else if opts.Contains("[]string") {
			// drop the properties
			q = fmt.Sprintf("%s.sideEffect(properties('%s').drop())", q, name)
			s := reflect.ValueOf(val)
			for i := 0; i < s.Len(); i++ {
				q = fmt.Sprintf("%s.property(list, '%s', '%s')", q, name, escapeString(fmt.Sprintf("%s", s.Index(i).Interface())))
			}
		} else if opts.Contains("[]bool") || opts.Contains("[]number") {
			// drop the properties
			q = fmt.Sprintf("%s.sideEffect(properties('%s').drop())", q, name)
			s := reflect.ValueOf(val)
			for i := 0; i < s.Len(); i++ {
				q = fmt.Sprintf("%s.property(list, '%s', %v)", q, name, s.Index(i).Interface())
			}
		}
	}

	if tagLength == 0 {
		return nil, ErrorInterfaceHasNoIdField
	}

	return c.Execute(q, nil, nil)
}

// DropV takes a interface and drops the vertex from the graph
func (c *Client) DropV(data interface{}) ([]*GremlinRespData, error) {
	c.verbose("passed interface: %s", spew.Sdump(data))
	if c.conn.isDisposed() {
		return nil, ErrorConnectionDisposed
	}
	d := getValue(data)

	id := d.FieldByName("Id")
	if !id.IsValid() {
		return nil, ErrorInterfaceHasNoIdField
	}

	q := fmt.Sprintf("g.V('%s').drop()", id)
	return c.Execute(q, nil, nil)
}

// AddE takes a label, from UUID and to UUID then creates a edge between the two vertex in the graph
func (c *Client) AddE(label string, from, to interface{}) ([]*GremlinRespData, error) {
	if c.conn.isDisposed() {
		return nil, ErrorConnectionDisposed
	}
	df := getValue(from)
	fid := df.FieldByName("Id")
	if !fid.IsValid() {
		return nil, ErrorInterfaceHasNoIdField
	}

	dt := getValue(to)
	tid := dt.FieldByName("Id")
	if !tid.IsValid() {
		return nil, ErrorInterfaceHasNoIdField
	}

	q := fmt.Sprintf("g.V('%s').addE('%s').to(g.V('%s'))", fid.Interface(), label, tid.Interface())
	return c.Execute(q, map[string]string{}, map[string]string{})
}

// AddEById takes a label, from UUID and to UUID then creates a edge between the two vertex in the graph
func (c *Client) AddEById(label string, from, to uuid.UUID) ([]*GremlinRespData, error) {
	if c.conn.isDisposed() {
		return nil, ErrorConnectionDisposed
	}
	q := fmt.Sprintf("g.V('%s').addE('%s').to(g.V('%s'))", from.String(), label, to.String())
	return c.Execute(q, map[string]string{}, map[string]string{})
}

// AddEWithProps takes a label, from UUID and to UUID then creates a edge between the two vertex in the graph
func (c *Client) AddEWithProps(label string, from, to interface{}, props map[string]interface{}) ([]*GremlinRespData, error) {
	if c.conn.isDisposed() {
		return nil, ErrorConnectionDisposed
	}
	df := getValue(from)
	fid := df.FieldByName("Id")
	if !fid.IsValid() {
		return nil, ErrorInterfaceHasNoIdField
	}

	dt := getValue(to)
	tid := dt.FieldByName("Id")
	if !tid.IsValid() {
		return nil, ErrorInterfaceHasNoIdField
	}

	q := fmt.Sprintf("g.V('%s').addE('%s').to(g.V('%s'))", fid.Interface().(uuid.UUID).String(), label, tid.Interface().(uuid.UUID).String())
	p, err := buildProps(props)
	if err != nil {
		return nil, err
	}
	q = q + p
	return c.Execute(q, map[string]string{}, map[string]string{})
}

// AddEWithPropsById takes a label, from UUID and to UUID then creates a edge between the two vertex in the graph
func (c *Client) AddEWithPropsById(label string, from, to uuid.UUID, props map[string]interface{}) ([]*GremlinRespData, error) {
	if c.conn.isDisposed() {
		return nil, ErrorConnectionDisposed
	}
	q := fmt.Sprintf("g.V('%s').addE('%s').to(g.V('%s'))", from.String(), label, to.String())
	p, err := buildProps(props)
	if err != nil {
		return nil, err
	}
	q = q + p
	return c.Execute(q, map[string]string{}, map[string]string{})
}

// DropE takes a label, from UUID and to UUID then drops the edge between the two vertex in the graph
func (c *Client) DropE(label string, from, to interface{}) ([]*GremlinRespData, error) {
	if c.conn.isDisposed() {
		return nil, ErrorConnectionDisposed
	}
	df := getValue(from)
	fid := df.FieldByName("Id")
	if !fid.IsValid() {
		return nil, ErrorInterfaceHasNoIdField
	}

	dt := getValue(to)
	tid := dt.FieldByName("Id")
	if !tid.IsValid() {
		return nil, ErrorInterfaceHasNoIdField
	}

	q := fmt.Sprintf("g.V('%s').outE('%s').and(inV().is('%s')).drop()", fid.Interface(), label, tid.Interface())
	return c.Execute(q, map[string]string{}, map[string]string{})
}

// DropEById takes a label, from UUID and to UUID then drops the edge between the two vertex in the graph
func (c *Client) DropEById(label string, from, to uuid.UUID) ([]*GremlinRespData, error) {
	if c.conn.isDisposed() {
		return nil, ErrorConnectionDisposed
	}
	q := fmt.Sprintf("g.V('%s').outE('%s').and(inV().is('%s')).drop()", from.String(), label, to.String())
	return c.Execute(q, map[string]string{}, map[string]string{})
}

// getProprtyValue takes a property map slice and return the value
func getPropertyValue(prop interface{}) (interface{}, error) {
	propMap, ok := prop.(map[string]interface{})
	if ok {
		if val, ok := propMap["value"]; ok {
			return val, nil
		}
	}
	return nil, ErrorCannotCastProperty
}

// buildProps takes a map of string to interfaces to be used as properties on a edge
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
				q = fmt.Sprintf("%s.property('%s', '%s')", q, k, escapeString(fmt.Sprintf("%s", s.Index(i).Interface())))
			}
		} else {
			return "", ErrorUnsupportedPropertyMap
		}
	}
	return q, nil
}

// getValue returns the underlying reflect.Value
func getValue(data interface{}) reflect.Value {
	var d reflect.Value
	if reflect.ValueOf(data).Kind() != reflect.Ptr {
		d = reflect.ValueOf(data)
	} else {
		d = reflect.ValueOf(data).Elem()
	}
	return d
}

// escapeString takes a string escapes
func escapeString(str string) string {
	var buf bytes.Buffer
	for _, char := range str {
		switch char {
		case '\'', '"', '\\':
			buf.WriteRune('\\')
		}
		buf.WriteRune(char)
	}
	return buf.String()
}
