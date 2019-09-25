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

	return c, c.errs
}

// Reconnect tries to reconnect the underlying ws connection
func (c *Client) Reconnect() {
	if !c.conn.isConnected() {
		err := c.conn.connect()
		if err != nil {
			c.errs <- err
		}
	}
}

// IsConnected return bool
func (c *Client) IsConnected() bool {
	return c.conn.isConnected()
}

// debug prints to the configured logger if debug is enabled
func (c *Client) debug(frmt string, i ...interface{}) {
	if c.conf.Debug && c.conf.Logger != nil {
		c.conf.Logger.InfoDepth(1, fmt.Sprintf("GREMGOSER: DEBUG: "+frmt, i...))
	}
}

// verbose prints to the configured logger if verbose is enabled
func (c *Client) verbose(frmt string, i ...interface{}) {
	if c.conf.Verbose && c.conf.Logger != nil {
		c.conf.Logger.InfoDepth(1, fmt.Sprintf("GREMGOSER: VERBOSE: "+frmt, i...))
	}
}

// veryVerbose prints to the configured logger if very verbose is enabled
func (c *Client) veryVerbose(frmt string, i ...interface{}) {
	if c.conf.VeryVerbose && c.conf.Logger != nil {
		c.conf.Logger.InfoDepth(1, fmt.Sprintf("GREMGOSER: VERY VERBOSE: "+frmt, i...))
	}
}

func (c *Client) executeRequest(query string, bindings, rebindings map[string]interface{}) ([]*GremlinRespData, error) {
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
func (c *Client) Execute(query string, bindings, rebindings map[string]interface{}) ([]*GremlinRespData, error) {
	c.verbose("connection: %+v", c.conn)
	if c.conn.isDisposed() {
		return nil, ErrorConnectionDisposed
	}
	c.verbose("query: %s", query)
	resp, err := c.executeRequest(query, bindings, rebindings)
	c.verbose("response: %+v", spew.Sprint(resp))
	return resp, err
}

// Get formats a raw Gremlin query, sends it to Gremlin Server, and populates the passed []interface.
func (c *Client) Get(query string, bindings map[string]interface{}, ptr interface{}) error {
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
	// we try to unmarshal the response data slice
	obj, err := json.Marshal(respDataSlice)
	if err != nil {
		c.debug("err marshaling resp data slice: %s", err)
		return nil
	}
	decoder := json.NewDecoder(bytes.NewReader(obj))
	decoder.UseNumber()
	if _, ok := (*respDataSlice[0])["properties"]; ok {
		err := decoder.Decode(&respSlice)
		//err := json.Unmarshal(obj, &respSlice)
		if err != nil {
			c.debug("err unmarshaling response slice: %s", err)
			return err
		}
	} else {
		err := decoder.Decode(&ptr)
		//err := json.Unmarshal(obj, &ptr)
		if err != nil {
			c.debug("err unmarshaling response slice: %s", err)
			return err
		}
		return nil
	}

	c.veryVerbose("Response Data Slice: %s", spew.Sdump(respSlice))

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
			c.veryVerbose("Struct Field ==> Name: %s, Opts: %s", name, opts)
			if len(name) == 0 && len(opts) == 0 {
				continue
			}
			// get the current field
			f := s.Elem().Field(i)
			// check if we can modify
			if f.CanSet() {
				isPtr, isSlice := false, false
				var kind reflect.Kind
				// if it's a pointer or slice get the underlying interface
				if f.Kind() == reflect.Ptr {
					isPtr = true
					kind = f.Type().Elem().Kind()
				} else if f.Kind() == reflect.Slice {
					isSlice = true
					kind = f.Type().Elem().Kind()
				} else {
					kind = f.Kind()
				}
				_ = isSlice
				c.veryVerbose("Struct Field Type: %s", kind)
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
							switch kind {
							case reflect.String: // Set as string
								vString, ok := v.(string)
								if ok {
									if isPtr {
										f.Set(reflect.ValueOf(&vString))
									} else {
										f.SetString(vString)
									}
								}
							case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64: // Set as int
								vNumber, ok := v.(json.Number)
								if ok {
									vInt, _ := vNumber.Int64()
									if !f.OverflowInt(vInt) {
										if isPtr {
											f.Set(reflect.ValueOf(&vInt))
										} else {
											f.SetInt(vInt)
										}
									}
								}
							case reflect.Float32, reflect.Float64: // Set as float
								vNumber, ok := v.(json.Number)
								if ok {
									vFloat, _ := vNumber.Float64()
									if !f.OverflowFloat(vFloat) {
										if isPtr {
											f.Set(reflect.ValueOf(&vFloat))
										} else {
											f.SetFloat(vFloat)
										}
									}
								}
							case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64: // Set as uint
								vNumber, ok := v.(json.Number)
								if ok {
									vInt, _ := vNumber.Int64()
									vUint := uint64(vInt)
									if !f.OverflowUint(vUint) {
										if isPtr {
											f.Set(reflect.ValueOf(&vUint))
										} else {
											f.SetUint(vUint)
										}
									}
								}
							case reflect.Bool: // Set as bool
								vBool, ok := v.(bool)
								if ok {
									if isPtr {
										f.Set(reflect.ValueOf(&vBool))
									} else {
										f.SetBool(vBool)
									}
								}
							case reflect.Struct, reflect.Map: // take JSON string and unmarshal into struct or map
								vString, ok := v.(string)
								if ok {
									s := reflect.New(f.Type()).Interface()
									json.Unmarshal([]byte(vString), s)
									if isPtr {
										f.Set(reflect.ValueOf(s))
									} else {
										f.Set(reflect.ValueOf(s).Elem())
									}
								}
							}
							// this is a special case
							if isSlice && opts.Contains("struct") { // take JSON string and unmarshal into slice
								vString, ok := v.(string) // the data is stored as a struct string in the graph
								if ok {
									sSlice := reflect.SliceOf(f.Type().Elem())
									s := reflect.New(sSlice)
									json.Unmarshal([]byte(vString), s.Interface())
									if isPtr {
										sInterface := s.Interface()
										f.Set(reflect.ValueOf(sInterface))
									} else {
										f.Set(s.Elem())
									}
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
								switch kind {
								case reflect.String: // Set as string
									vString, ok := v.(string)
									if ok {
										pSlice.Index(i).SetString(vString)
									}
								case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64: // Set as int
									vNumber, ok := v.(json.Number)
									if ok {
										vInt, _ := vNumber.Int64()
										if !pSlice.Index(i).OverflowInt(vInt) {
											pSlice.Index(i).SetInt(vInt)
										}
									}
								case reflect.Float32, reflect.Float64: // Set as float
									vNumber, ok := v.(json.Number)
									if ok {
										vFloat, _ := vNumber.Float64()
										if !pSlice.Index(i).OverflowFloat(vFloat) {
											pSlice.Index(i).SetFloat(vFloat)
										}
									}
								case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64: // Set as uint
									vNumber, ok := v.(json.Number)
									if ok {
										vInt, _ := vNumber.Int64()
										vUint := uint64(vInt)
										if !pSlice.Index(i).OverflowUint(vUint) {
											pSlice.Index(i).SetUint(vUint)
										}
									}
								case reflect.Bool: // Set as bool
									vBool, ok := v.(bool)
									if ok {
										pSlice.Index(i).SetBool(vBool)
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

	c.veryVerbose("Interface de-serialized: %+v", spew.Sdump(ptr))

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
		if reflect.TypeOf(val).Kind() == reflect.Ptr {
			val = reflect.ValueOf(val).Elem().Interface()
		}
		if len(opts) == 0 {
			return nil, fmt.Errorf("gremgoser: interface field graph tag does not contain a tag option type, field type: %T", val)
		} else if opts.Contains("string") || opts.Contains("partitionKey") {
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
		} else if opts.Contains("partitionKey") {
			q = fmt.Sprintf("%s.has('%s', '%s')", q, name, escapeString(fmt.Sprintf("%s", val)))
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
	return c.Execute(q, nil, nil)
}

// AddEById takes a label, from UUID and to UUID then creates a edge between the two vertex in the graph
func (c *Client) AddEById(label string, from, to uuid.UUID) ([]*GremlinRespData, error) {
	if c.conn.isDisposed() {
		return nil, ErrorConnectionDisposed
	}
	q := fmt.Sprintf("g.V('%s').addE('%s').to(g.V('%s'))", from.String(), label, to.String())
	return c.Execute(q, nil, nil)
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
	return c.Execute(q, nil, nil)
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
	return c.Execute(q, nil, nil)
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
	return c.Execute(q, nil, nil)
}

// DropEById takes a label, from UUID and to UUID then drops the edge between the two vertex in the graph
func (c *Client) DropEById(label string, from, to uuid.UUID) ([]*GremlinRespData, error) {
	if c.conn.isDisposed() {
		return nil, ErrorConnectionDisposed
	}
	q := fmt.Sprintf("g.V('%s').outE('%s').and(inV().is('%s')).drop()", from.String(), label, to.String())
	return c.Execute(q, nil, nil)
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
