package gremgoser

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

func (c *Client) handleResponse(msg []byte) error {
	resp, err := marshalResponse(msg)
	if err != nil && err != Error407Authenticate {
		c.debug("error handling response: %s", err)
		return err
	}

	c.verbose("handling response: %+v", resp)

	if resp.Status.Code == 407 { //Server request authentication
		return c.authenticate(resp.RequestId)
	}

	c.saveResponse(resp)
	return nil
}

// marshalResponse creates a response struct for every incoming response for further manipulation
func marshalResponse(msg []byte) (*GremlinResponse, error) {
	resp := &GremlinResponse{}
	err := json.Unmarshal(msg, resp)
	if err != nil {
		return resp, err
	}

	err = responseDetectError(resp.Status.Code)
	if err != nil {
		return resp, err
	}
	return resp, nil
}

// saveResponse makes the response available for retrieval by the requester. Mutexes are used for thread safety.
func (c *Client) saveResponse(resp *GremlinResponse) {
	c.respMutex.Lock()
	var container []*GremlinData
	existingData, ok := c.results.Load(resp.RequestId) // Retrieve old data container (for requests with multiple responses)
	if ok {
		container = existingData.([]*GremlinData)
	}
	c.verbose("RequestId: %s, existing data: %+v", resp.RequestId, container)
	for _, val := range resp.Result.Data {
		container = append(container, val) //iterate over new items
	}
	c.verbose("RequestId: %s, new data: %+v", resp.RequestId, container)
	c.results.Store(resp.RequestId, container) // Add new data to buffer for future retrieval
	respNotifier, _ := c.responseNotifier.LoadOrStore(resp.RequestId, make(chan int, 1))
	if resp.Status.Code != 206 {
		respNotifier.(chan int) <- 1
	}
	c.respMutex.Unlock()
}

// retrieveResponse retrieves the response saved by saveResponse.
func (c *Client) retrieveResponse(id uuid.UUID) []*GremlinData {
	data := []*GremlinData{}
	resp, _ := c.responseNotifier.Load(id)
	timeout := make(chan bool, 1)
	go func() {
		time.Sleep(c.conf.ReadingWait)
		timeout <- true
	}()
	select {
	case n := <-resp.(chan int):
		if n == 1 {
			if dataI, ok := c.results.Load(id); ok {
				data = dataI.([]*GremlinData)
				close(resp.(chan int))
				c.responseNotifier.Delete(id)
				c.deleteResponse(id)
			}
		}
	case <-timeout:
		// the read from resp ch has timed out
		c.debug("timeout on response")
		return nil
	}
	return data
}

// deleteRespones deletes the response from the container. Used for cleanup purposes by requester.
func (c *Client) deleteResponse(id uuid.UUID) {
	c.results.Delete(id)
	return
}

// responseDetectError detects any possible errors in responses from Gremlin Server and generates an error for each code
func responseDetectError(code int) error {
	switch code {
	case 200:
		return nil
	case 204:
		return nil
	case 206:
		return nil
	case 401:
		return Error401Unauthorized
	case 407:
		return Error407Authenticate
	case 498:
		return Error498MalformedRequest
	case 499:
		return Error499InvalidRequestArguments
	case 500:
		return Error500ServerError
	case 597:
		return Error597ScriptEvaluationError
	case 598:
		return Error598ServerTimeout
	case 599:
		return Error599ServerSerializationError
	default:
		return ErrorUnknownCode
	}
	return nil
}
