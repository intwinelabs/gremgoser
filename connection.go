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
		WriteBufferSize:  8192,
		ReadBufferSize:   8192,
		HandshakeTimeout: 60 * time.Second, // Timeout or else we'll hang forever and never fail on bad hosts.
	}
	ws.conn, resp, err = d.Dial(ws.uri, http.Header{})
	if err != nil {
		ws.verbosef("error dialing websocket connection (%s): %s", ws.uri, err)
	}
	ws.verbosef("dial response: %+v", resp)
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
	}

	return nil
}

func (ws *Ws) pongHandler(appData string) error {
	ws.conn.SetReadDeadline(time.Now().Add(ws.pingInterval + 10))
	ws.Lock()
	ws.connected = true
	ws.Unlock()
	ws.verbosef("received pong message from server")
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
	wwt := time.Now().Add(ws.writingWait)
	ws.verbosef("waiting to write until: %s, msg: %s", wwt.Format("Mon Jan 2 15:04:05 -0700 MST 2006"), msg)
	ws.conn.SetWriteDeadline(wwt)
	err := ws.conn.WriteMessage(2, msg)
	if err == nil {
		ws.verbosef("msg write: %s, msg: %s", time.Now().Format("Mon Jan 2 15:04:05 -0700 MST 2006"), msg)
	}
	return err
}

func (ws *Ws) read() ([]byte, error) {
	if ws.conn == nil {
		return nil, ErrorWSConnectionNil
	}
	rwt := time.Now().Add(ws.readingWait)
	ws.verbosef("waiting to read until: %s", rwt.Format("Mon Jan 2 15:04:05 -0700 MST 2006"))
	ws.conn.SetReadDeadline(rwt)
	_, msg, err := ws.conn.ReadMessage()
	if err == nil {
		ws.verbosef("msg read: %s, msg: %s", time.Now().Format("Mon Jan 2 15:04:05 -0700 MST 2006"), msg)
	}
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
			ws.verbosef("sending ping message to server")
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
			c.verbose("message handled: %s", msg)
		}
		select {
		case <-quit:
			return
		default:
			continue
		}
	}
}

// debugf prints to the configured logger if debug is enabled
func (ws *Ws) debugf(frmt string, i ...interface{}) {
	if ws.debug {
		ws.logger.InfoDepth(1, fmt.Sprintf("GREMGOSER: WS: DEBUG: "+frmt, i...))
	}
}

// verbosef prints to the configured logger if debug is enabled
func (ws *Ws) verbosef(frmt string, i ...interface{}) {
	if ws.verbose {
		ws.logger.InfoDepth(1, fmt.Sprintf("GREMGOSER: WS: VERBOSE: "+frmt, i...))
	}
}
