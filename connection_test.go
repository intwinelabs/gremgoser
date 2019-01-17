package gremgoser

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
)

var upgrader = websocket.Upgrader{}

func echo(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer c.Close()
	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			break
		}
		err = c.WriteMessage(mt, message)
		if err != nil {
			break
		}
	}
}

func TestWsConnection(t *testing.T) {
	assert := assert.New(t)

	// Create test server with the echo handler.
	s := httptest.NewServer(http.HandlerFunc(echo))
	defer s.Close()

	// Convert http://127.0.0.1 to ws://127.0.0.
	u := "ws" + strings.TrimPrefix(s.URL, "http")

	// test connecting to the mock server
	ws := NewDialer(u)
	err := ws.connect()
	assert.Nil(err)

	// test to see if connectend
	connected := ws.isConnected()
	assert.True(connected)

	// test to see if dispoased
	disposed := ws.isDisposed()
	assert.False(disposed)

	// test write to ws
	err = ws.write([]byte("foo"))
	assert.Nil(err)

	// test read
	resp, err := ws.read()
	assert.Nil(err)
	_resp := []byte("foo")
	assert.Equal(_resp, resp)

	// test get auth
	auth, err := ws.getAuth()
	assert.Nil(auth)
	assert.Equal(NoAuth, err)

	// TODO test ping

	// test close
	err = ws.close()
	assert.Nil(err)
}
