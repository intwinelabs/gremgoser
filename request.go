package gremgoser

import (
	"encoding/base64"
	"encoding/json"

	"github.com/google/uuid"
)

type requester interface {
	prepare() error
	getID() string
	getRequest() *GremlinRequest
}

// prepareRequest packages a query and binding into the format that Gremlin Server accepts
func prepareRequest(query string, bindings, rebindings map[string]interface{}) *GremlinRequest {
	req := &GremlinRequest{}
	req.RequestId = uuid.New()
	req.Op = "eval"
	req.Processor = ""

	req.Args = make(map[string]interface{})
	req.Args["language"] = "gremlin-groovy"
	req.Args["gremlin"] = query
	req.Args["bindings"] = bindings
	req.Args["rebindings"] = rebindings

	return req
}

// prepareAuthRequest creates a ws request for Gremlin Server
func prepareAuthRequest(requestId uuid.UUID, username, password string) *GremlinRequest {
	req := &GremlinRequest{}
	req.RequestId = requestId
	req.Op = "authentication"
	req.Processor = "traversal"

	var simpleAuth []byte
	user := []byte(username)
	pass := []byte(password)

	simpleAuth = append(simpleAuth, 0)
	simpleAuth = append(simpleAuth, user...)
	simpleAuth = append(simpleAuth, 0)
	simpleAuth = append(simpleAuth, pass...)

	req.Args = make(map[string]interface{})
	req.Args["sasl"] = base64.StdEncoding.EncodeToString(simpleAuth)

	return req
}

// formatMessage takes a request type and formats it into being able to be delivered to Gremlin Server
func packageRequest(req *GremlinRequest) ([]byte, error) {
	msg := []byte{}
	j, err := json.Marshal(req) // Formats request into byte format
	if err != nil {
		return msg, err
	}
	mimeType := []byte("!application/vnd.gremlin-v2.0+json")
	msg = append(mimeType, j...)

	return msg, nil
}

// dispatchRequest sends the request for writing to the remote Gremlin Server
func (c *Client) dispatchRequest(msg []byte) {
	c.verbose("dispatching request: %s", msg)
	c.requests <- msg
}
