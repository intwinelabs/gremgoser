package gremgoser

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

type dialer interface {
	connect() error
	isConnected() bool
	isDisposed() bool
	write([]byte) error
	read() ([]byte, error)
	close() error
	ping(errs chan error)
}

func (ws *Ws) connect() (err error) {
	var resp *http.Response
	d := websocket.Dialer{
		WriteBufferSize:  8192,
		ReadBufferSize:   8192,
		HandshakeTimeout: 5 * time.Second, // Timeout or else we'll hang forever and never fail on bad hosts.
	}
	ws.conn, resp, err = d.Dial(ws.uri, http.Header{})
	if err != nil {
		// As of 3.2.2 the URL has changed.
		// https://groups.google.com/forum/#!msg/gremlin-users/x4hiHsmTsHM/Xe4GcPtRCAAJ
		ws.uri = ws.uri + "/gremlin"
		ws.conn, resp, err = d.Dial(ws.uri, http.Header{})
	}

	if err != nil && resp != nil {
		err = fmt.Errorf("WS connection error: %s: %s", resp.Status, err)
	}

	if err == nil {
		ws.connected = true
		ws.disposed = false
		ws.conn.SetPongHandler(ws.pongHandler)
	}
	return
}

func (ws *Ws) pongHandler(appData string) error {
	ws.Lock()
	ws.connected = true
	ws.Unlock()
	return nil
}

func (ws *Ws) isConnected() bool {
	return ws.connected
}

func (ws *Ws) isDisposed() bool {
	return ws.disposed
}

func (ws *Ws) write(msg []byte) (err error) {
	ws.conn.SetWriteDeadline(time.Now().Add(ws.writingWait))
	err = ws.conn.WriteMessage(2, msg)
	return
}

func (ws *Ws) read() (msg []byte, err error) {
	ws.conn.SetReadDeadline(time.Now().Add(ws.readingWait))
	_, msg, err = ws.conn.ReadMessage()
	return
}

func (ws *Ws) close() (err error) {
	defer func() {
		close(ws.quit)
		ws.conn.Close()
		ws.disposed = true
	}()

	err = ws.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "")) //Cleanly close the connection with the server
	return
}

func (ws *Ws) ping(errs chan error) {
	ticker := time.NewTicker(ws.pingInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			connected := true
			if err := ws.conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(ws.writingWait)); err != nil {
				errs <- err
				connected = false
			}
			ws.Lock()
			ws.connected = connected
			ws.Unlock()

		case <-ws.quit:
			return
		}
	}
}

// writeWorker works on a loop and dispatches messages as soon as it receives them
func (c *Client) writeWorker(errs chan error, quit chan struct{}) {
	for {
		select {
		case msg := <-c.requests:
			err := c.conn.write(msg)
			if err != nil {
				errs <- err
				c.Errored = true
				break
			}

		case <-quit:
			return
		}
	}
}

// readWorker works on a loop and sorts messages as soon as it receives them
func (c *Client) readWorker(errs chan error, quit chan struct{}) {
	for {
		msg, err := c.conn.read()
		if err != nil {
			errs <- err
			c.Errored = true
			break
		}
		if msg != nil {
			err := c.handleResponse(msg)
			if err != nil {
				errs <- err
			}
			c.verbose("Message handled...")
		}

		select {
		case <-quit:
			return
		default:
			continue
		}
	}
}
