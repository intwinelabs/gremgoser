package gremgoser

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// TestRequestPreparation tests the ability to package a query and a set of bindings into a request struct for further manipulation
func TestRequestPreparation(t *testing.T) {
	assert := assert.New(t)

	query := "g.V(x)"
	bindings := map[string]string{"x": "10"}
	rebindings := map[string]string{}
	req := prepareRequest(query, bindings, rebindings)
	assert.NotNil(req)
	assert.IsType(&GremlinRequest{}, req)

	_req := &GremlinRequest{
		RequestId: req.RequestId,
		Op:        "eval",
		Processor: "",
		Args: map[string]interface{}{
			"gremlin":    query,
			"bindings":   bindings,
			"language":   "gremlin-groovy",
			"rebindings": rebindings,
		},
	}

	assert.Equal(_req, req)
}

// TestRequestPackaging tests the ability for gremgoser to format a request using the established Gremlin Server WebSockets protocol for delivery to the server
func TestRequestPackaging(t *testing.T) {
	assert := assert.New(t)

	id, err := uuid.Parse("1d6d02bd-8e56-421d-9438-3bd6d0079ff1")
	assert.IsType(uuid.UUID{}, id)
	assert.Nil(err)
	req := &GremlinRequest{
		RequestId: id,
		Op:        "eval",
		Processor: "",
		Args: map[string]interface{}{
			"gremlin":  "g.V(x)",
			"bindings": map[string]string{"x": "10"},
			"language": "gremlin-groovy",
		},
	}

	msg, err := packageRequest(req)
	if err != nil {
		t.Error(err)
	}

	j, err := json.Marshal(req)
	if err != nil {
		t.Error(err)
	}

	var _msg []byte

	mimetype := []byte("application/vnd.gremlin-v2.0+json")
	_msg = append(mimetype, j...)

	assert.Equal(_msg, msg)
}

// TestRequestDispatch tests the ability for a requester to send a request to the client for writing to Gremlin Server
func TestRequestDispatch(t *testing.T) {
	assert := assert.New(t)

	id, err := uuid.Parse("1d6d02bd-8e56-421d-9438-3bd6d0079ff1")
	assert.IsType(uuid.UUID{}, id)
	assert.Nil(err)
	req := &GremlinRequest{
		RequestId: id,
		Op:        "eval",
		Processor: "",
		Args: map[string]interface{}{
			"gremlin":  "g.V(x)",
			"bindings": map[string]string{"x": "10"},
			"language": "gremlin-groovy",
		},
	}
	c, _ := NewClient(NewClientConfig("ws://127.0.0.1"))
	msg, err := packageRequest(req)
	assert.Nil(err)
	c.dispatchRequest(msg)
	_req := <-c.requests // c.requests is the channel where all requests are sent for writing to Gremlin Server, write workers listen on this channel
	assert.Equal(_req, msg)
}

// TestAuthRequestDispatch tests the ability for a requester to send a request to the client for writing to Gremlin Server
func TestAuthRequestDispatch(t *testing.T) {
	assert := assert.New(t)
	id, err := uuid.Parse("1d6d02bd-8e56-421d-9438-3bd6d0079ff1")
	assert.IsType(uuid.UUID{}, id)
	assert.Nil(err)
	req := prepareAuthRequest(id, "test", "root")

	c, _ := NewClient(NewClientConfig("ws://127.0.0.1"))
	msg, err := packageRequest(req)
	assert.Nil(err)
	c.dispatchRequest(msg)
	_req := <-c.requests // c.requests is the channel where all requests are sent for writing to Gremlin Server, write workers listen on this channel
	assert.Equal(_req, msg)
}

// TestAuthRequestPreparation tests the ability to create successful authentication request
func TestAuthRequestPreparation(t *testing.T) {
	assert := assert.New(t)
	id, err := uuid.Parse("1d6d02bd-8e56-421d-9438-3bd6d0079ff1")
	assert.IsType(uuid.UUID{}, id)
	assert.Nil(err)

	req := prepareAuthRequest(id, "test", "root")
	assert.False(req.RequestId != id || req.Processor != "trasversal" || req.Op != "authentication")
	assert.False(len(req.Args) != 1 || req.Args["sasl"] == "")
}
