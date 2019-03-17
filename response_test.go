package gremgoser

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

/*
Dummy responses for mocking
*/

var id, _ = uuid.Parse("1d6d02bd-8e56-421d-9438-3bd6d0079ff1")

var dummySuccessfulResponse = []byte(`{"result":{"data":[{"id": 2,"label": "person","type": "vertex","properties": [
  {"id": 2, "value": "vadas", "label": "name"},
  {"id": 3, "value": 27, "label": "age"}]}
  ], "meta":{}},
 "requestId":"1d6d02bd-8e56-421d-9438-3bd6d0079ff1",
 "status":{"code":200,"attributes":{},"message":""}}`)

var dummyNeedAuthenticationResponse = []byte(`{"result":{},
 "requestId":"1d6d02bd-8e56-421d-9438-3bd6d0079ff1",
 "status":{"code":407,"attributes":{},"message":""}}`)

var dummyPartialResponse1 = []byte(`{"result":{"data":[{"id": 2,"label": "person","type": "vertex","properties": [
  {"id": 2, "value": "vadas", "label": "name"},
  {"id": 3, "value": 27, "label": "age"}]},
  ], "meta":{}},
 "requestId":"1d6d02bd-8e56-421d-9438-3bd6d0079ff1",
 "status":{"code":206,"attributes":{},"message":""}}`)

var dummyPartialResponse2 = []byte(`{"result":{"data":[{"id": 4,"label": "person","type": "vertex","properties": [
  {"id": 5, "value": "quant", "label": "name"},
  {"id": 6, "value": 54, "label": "age"}]},
  ], "meta":{}},
 "requestId":"1d6d02bd-8e56-421d-9438-3bd6d0079ff1",
 "status":{"code":200,"attributes":{},"message":""}}`)

var dummySuccessfulResponseMarshalled = &GremlinResponse{
	RequestId: id,
	Status:    GremlinStatus{Code: 200},
	Result:    GremlinResult{Data: []*GremlinData{&GremlinData{Label: "testData"}}},
}

var dummyNeedAuthenticationResponseMarshalled = &GremlinResponse{
	RequestId: id,
	Status:    GremlinStatus{Code: 407},
	Result:    GremlinResult{Data: []*GremlinData{}},
}

var dummyPartialResponse1Marshalled = &GremlinResponse{
	RequestId: id,
	Status:    GremlinStatus{Code: 206}, // Code 206 indicates that the response is not the terminating response in a sequence of responses
	Result:    GremlinResult{Data: []*GremlinData{}},
}

var dummyPartialResponse2Marshalled = &GremlinResponse{
	RequestId: id,
	Status:    GremlinStatus{Code: 200},
	Result:    GremlinResult{Data: []*GremlinData{}},
}

// TestResponseHandling tests the overall response handling mechanism of gremgo
func TestResponseHandling(t *testing.T) {
	assert := assert.New(t)

	c, _ := NewClient(NewClientConfig("ws://127.0.0.1"))

	err := c.handleResponse(dummySuccessfulResponse)

	assert.Nil(err)
	_resp := dummySuccessfulResponseMarshalled.Result.Data
	resp := c.retrieveResponse(dummySuccessfulResponseMarshalled.RequestId)
	assert.Equal(_resp, resp)
}

func TestResponseAuthHandling(t *testing.T) {
	assert := assert.New(t)

	conf := NewClientConfig("ws://127.0.0.1")
	conf.SetAuthentication("test", "pass")
	c, _ := NewClient(conf)

	c.handleResponse(dummyNeedAuthenticationResponse)

	req := conf.AuthReq

	sampleAuthRequest, err := packageRequest(req)
	assert.Nil(err)

	authRequest := <-c.requests //Simulate that client send auth challenge to server

	assert.Equal(authRequest, sampleAuthRequest)

	c.handleResponse(dummySuccessfulResponse) //If authentication is successful the server returns the origin petition

	var expectedSuccessful []interface{}
	expectedSuccessful = append(expectedSuccessful, dummySuccessfulResponseMarshalled.Result.Data)

	_req := c.retrieveResponse(dummySuccessfulResponseMarshalled.RequestId)
	assert.Equal(_req, req)
}

// TestResponseMarshalling tests the ability to marshal a response into a designated response struct for further manipulation
func TestResponseMarshalling(t *testing.T) {
	assert := assert.New(t)

	resp, err := marshalResponse(dummySuccessfulResponse)
	assert.Nil(err)
	assert.False(dummySuccessfulResponseMarshalled.RequestId != resp.RequestId || dummySuccessfulResponseMarshalled.Status.Code != resp.Status.Code)

}

// TestResponseSortingSingleResponse tests the ability for sortResponse to save a response received from Gremlin Server
func TestResponseSortingSingleResponse(t *testing.T) {
	assert := assert.New(t)

	c, _ := NewClient(NewClientConfig("ws://127.0.0.1"))

	c.saveResponse(dummySuccessfulResponseMarshalled)

	res, ok := c.results.Load(dummySuccessfulResponseMarshalled.RequestId)
	assert.True(ok)
	assert.Equal(dummySuccessfulResponseMarshalled.Result.Data, res)
}

// TestResponseSortingMultipleResponse tests the ability for the sortResponse function to categorize and group responses that are sent in a stream
func TestResponseSortingMultipleResponse(t *testing.T) {
	assert := assert.New(t)

	c, _ := NewClient(NewClientConfig("ws://127.0.0.1"))

	c.saveResponse(dummyPartialResponse1Marshalled)
	c.saveResponse(dummyPartialResponse2Marshalled)

	var expected []*GremlinData
	expected = append(expected, dummyPartialResponse1Marshalled.Result.Data...)
	expected = append(expected, dummyPartialResponse2Marshalled.Result.Data...)

	results, ok := c.results.Load(dummyPartialResponse1Marshalled.RequestId)
	assert.True(ok)
	assert.Equal(expected, results)
}

// TestResponseRetrieval tests the ability for a requester to retrieve the response for a specified requestId generated when sending the request
func TestResponseRetrieval(t *testing.T) {
	assert := assert.New(t)
	c, _ := NewClient(NewClientConfig("ws://127.0.0.1"))

	c.saveResponse(dummyPartialResponse1Marshalled)
	c.saveResponse(dummyPartialResponse2Marshalled)

	resp := c.retrieveResponse(dummyPartialResponse1Marshalled.RequestId)

	var expected []*GremlinData
	expected = append(expected, dummyPartialResponse1Marshalled.Result.Data...)
	expected = append(expected, dummyPartialResponse2Marshalled.Result.Data...)

	assert.Equal(resp, expected)
}

// TestResponseDeletion tests the ability for a requester to clean up after retrieving a response after delivery to a client
func TestResponseDeletion(t *testing.T) {
	assert := assert.New(t)
	c, _ := NewClient(NewClientConfig("ws://127.0.0.1"))

	c.saveResponse(dummyPartialResponse1Marshalled)
	c.saveResponse(dummyPartialResponse2Marshalled)

	c.deleteResponse(dummyPartialResponse1Marshalled.RequestId)

	_, ok := c.results.Load(dummyPartialResponse1Marshalled.RequestId)
	assert.False(ok)
}

var codes = []struct {
	code int
}{
	{200},
	{204},
	{206},
	{401},
	{407},
	{498},
	{499},
	{500},
	{597},
	{598},
	{599},
	{3434}, // Testing unknown error code
}

// Tests detection of errors and if an error is generated for a specific error code
func TestResponseErrorDetection(t *testing.T) {
	assert := assert.New(t)
	for _, co := range codes {
		err := responseDetectError(co.code)
		switch {
		case co.code == 200:
			assert.Nil(err)
		case co.code == 204:
			assert.Nil(err)
		case co.code == 206:
			assert.Nil(err)
		default:
			assert.Nil(err)
		}
	}
}
