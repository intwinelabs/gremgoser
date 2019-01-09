package gremgo

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"reflect"
	"sync"
	"time"

	"github.com/iancoleman/strcase"

	"github.com/google/uuid"
)

// Client is a container for the gremgo client.
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

// Dial returns a gremgo client for interaction with the Gremlin Server specified in the host IP.
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
	auth := c.conn.getAuth()
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

// ExecuteFile takes a file path to a Gremlin script, sends it to Gremlin Server, and returns the result.
func (c *Client) ExecuteFile(path string, bindings, rebindings map[string]string) (resp interface{}, err error) {
	if c.conn.isDisposed() {
		return nil, errors.New("you cannot write on a disposed connection")
	}
	d, err := ioutil.ReadFile(path) // Read script from file
	if err != nil {
		log.Println(err)
		return
	}
	query := string(d)
	resp, err = c.executeRequest(query, bindings, rebindings)
	return
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

	for i := 0; i < d.NumField(); i++ {
		name := strcase.ToLowerCamel(d.Type().Field(i).Name)
		val := d.Field(i).Interface()
		switch val.(type) {
		case uuid.UUID:
			q = fmt.Sprintf("%s.property('%s', '%s')", q, name, val.(uuid.UUID).String())
		case string:
			q = fmt.Sprintf("%s.property('%s', '%s')", q, name, val)
		case bool, uint, uint8, uint16, uint32, uint64, int, int8, int16, int32, int64, float32, float64:
			q = fmt.Sprintf("%s.property('%s', %v)", q, name, val)
		default:
			return nil, fmt.Errorf("Type of interface field is not supported: %T", val)
		}
	}

	fmt.Println(q)
	resp, err = c.Execute(q, map[string]string{}, map[string]string{})
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
	fmt.Println(q)
	resp, err = c.Execute(q, map[string]string{}, map[string]string{})

	return
}

// AddEById takes a label, from UUID and to UUID then creates a edge between the two vertex in the graph
func (c *Client) AddEById(label string, from, to uuid.UUID) (resp interface{}, err error) {
	if c.conn.isDisposed() {
		return nil, errors.New("you cannot write on a disposed connection")
	}

	q := fmt.Sprintf("g.V('%s').addE('%s').to(g.V('%s'))", from.String(), label, to.String())
	fmt.Println(q)
	resp, err = c.Execute(q, map[string]string{}, map[string]string{})

	return
}
