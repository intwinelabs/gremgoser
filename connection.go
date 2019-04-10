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

func (ws *Ws) connect() error {
	var resp *http.Response
	var err error
	d := websocket.Dialer{
		WriteBufferSize:   4092,
		ReadBufferSize:    4092,
		HandshakeTimeout:  5 * time.Second, // Timeout or else we'll hang forever and never fail on bad hosts.
		EnableCompression: true,
	}
	ws.conn, resp, err = d.Dial(ws.uri, http.Header{})
	if err != nil {
		// As of 3.2.2 the URL has changed.
		// https://groups.google.com/forum/#!msg/gremlin-users/x4hiHsmTsHM/Xe4GcPtRCAAJ
		ws.uri = ws.uri + "/gremlin"
		ws.conn, resp, err = d.Dial(ws.uri, http.Header{})
	}

	if err != nil && resp == nil {
		return ErrorWSConnection
	}

	if err == nil {
		ws.connected = true
		ws.conn.SetPongHandler(ws.pongHandler)
		if ws.debug {
			go func() {
				con := ws.conn.UnderlyingConn()
				for {
					info, err := GetsockoptTCPInfo(&con)
					ws.debugf("tcpinfo: %+v, err: %+v", info, err)
					time.Sleep(1 * time.Second)
				}
			}()
		}
	}

	return nil
}

func (ws *Ws) pongHandler(appData string) error {
	ws.Lock()
	ws.connected = true
	ws.Unlock()
	ws.debugf("received pong message from server")
	return nil
}

func (ws *Ws) isConnected() bool {
	return ws.connected
}

func (ws *Ws) isDisposed() bool {
	return ws.disposed
}

func (ws *Ws) write(msg []byte) error {
	if ws.conn == nil {
		return ErrorWSConnectionNil
	}
	ws.conn.SetWriteDeadline(time.Now().Add(ws.writingWait))
	return ws.conn.WriteMessage(2, msg)
}

func (ws *Ws) read() ([]byte, error) {
	if ws.conn == nil {
		return nil, ErrorWSConnectionNil
	}
	ws.conn.SetReadDeadline(time.Now().Add(ws.readingWait))
	_, msg, err := ws.conn.ReadMessage()
	return msg, err
}

func (ws *Ws) close() error {
	if ws.conn == nil {
		return ErrorWSConnectionNil
	}
	defer func() {
		close(ws.quit)
		ws.conn.Close()
		ws.disposed = true
	}()

	err := ws.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "")) //Cleanly close the connection with the server
	return err
}

func (ws *Ws) ping(errs chan error) {
	var isConnected bool
	ticker := time.NewTicker(ws.pingInterval)
	defer ticker.Stop()
	for {
		if ws.conn == nil {
			errs <- ErrorWSConnectionNil
		}
		select {
		case <-ticker.C:
			isConnected = true
			err := ws.conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(ws.writingWait))
			if err != nil {
				errs <- err
				isConnected = false
			}
			ws.debugf("sending ping message to server")
			ws.Lock()
			ws.connected = isConnected
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

// debug prints to the configured logger if debug is enabled
func (ws *Ws) debugf(frmt string, i ...interface{}) {
	if ws.debug {
		ws.logger.InfoDepth(1, fmt.Sprintf("gremgoser: ws: DEBUG: "+frmt, i...))
	}
}
