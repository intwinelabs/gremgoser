package gremgoser

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSetAuthentication(t *testing.T) {
	assert := assert.New(t)

	user := "foo"
	pass := "bar"
	u := "ws://127.0.0.1"
	ws := NewDialer(u)
	dc := SetAuthentication(user, pass)
	dc(ws)
	assert.Equal(user, ws.auth.username)
	assert.Equal(pass, ws.auth.password)
}

func TestSetTimeout(t *testing.T) {
	assert := assert.New(t)

	u := "ws://127.0.0.1"
	ws := NewDialer(u)
	dc := SetTimeout(22)
	dc(ws)
	timeout := time.Duration(22000000000)
	assert.Equal(timeout, ws.timeout)
}

func TestSetPingInterval(t *testing.T) {
	assert := assert.New(t)

	u := "ws://127.0.0.1"
	ws := NewDialer(u)
	dc := SetPingInterval(22)
	dc(ws)
	interval := time.Duration(22000000000)
	assert.Equal(interval, ws.pingInterval)
}

func TestSetWritingWait(t *testing.T) {
	assert := assert.New(t)

	u := "ws://127.0.0.1"
	ws := NewDialer(u)
	dc := SetWritingWait(22)
	dc(ws)
	wait := time.Duration(22000000000)
	assert.Equal(wait, ws.writingWait)
}

func TestSetReadingWait(t *testing.T) {
	assert := assert.New(t)

	u := "ws://127.0.0.1"
	ws := NewDialer(u)
	dc := SetReadingWait(22)
	dc(ws)
	wait := time.Duration(22000000000)
	assert.Equal(wait, ws.readingWait)
}
