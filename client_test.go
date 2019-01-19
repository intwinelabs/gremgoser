package gremgoser

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewDialer(t *testing.T) {
	assert := assert.New(t)

	// no conf
	u := "ws://127.0.0.1"
	ws := NewDialer(u)
	assert.IsType(&Ws{}, ws)
	assert.Equal(u, ws.host)
	timeout := time.Duration(5000000000)
	assert.Equal(timeout, ws.timeout)

	// with conf
	user := "foo"
	pass := "bar"
	conf := SetAuthentication(user, pass)
	ws = NewDialer(u, conf)
	assert.IsType(&Ws{}, ws)
	assert.Equal(u, ws.host)
	assert.Equal(timeout, ws.timeout)
}

func TestDial(t *testing.T) {
	assert := assert.New(t)

	// Create test server with the mock handler.
	s := httptest.NewServer(http.HandlerFunc(mock))
	defer s.Close()

	// Convert http://127.0.0.1 to ws://127.0.0.
	u := "ws" + strings.TrimPrefix(s.URL, "http")

	// test connecting to the mock server
	ws := NewDialer(u)

	// setup err channel
	errs := make(chan error)
	go func(chan error) {
		err := <-errs
		assert.Nil(err)
	}(errs)

	c, err := Dial(ws, errs)
	assert.Nil(err)
	assert.NotNil(c)
	assert.IsType(Client{}, c)
}

func TestExecute(t *testing.T) {
	assert := assert.New(t)

	// Create test server with the mock handler.
	s := httptest.NewServer(http.HandlerFunc(mock))
	defer s.Close()

	// Convert http://127.0.0.1 to ws://127.0.0.
	u := "ws" + strings.TrimPrefix(s.URL, "http")

	// test connecting to the mock server
	ws := NewDialer(u)

	// setup err channel
	errs := make(chan error)
	go func(chan error) {
		err := <-errs
		assert.Nil(err)
	}(errs)

	c, err := Dial(ws, errs)
	assert.Nil(err)
	assert.NotNil(c)
	assert.IsType(Client{}, c)

	// test query execution
	q := "g.V()"
	resp, err := c.Execute(q, nil, nil)
	assert.Nil(err)
	assert.NotNil(resp)
}
