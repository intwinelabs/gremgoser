package gremgoser

import (
	"testing"
	"time"

	"github.com/intwinelabs/logger"
	"github.com/stretchr/testify/assert"
)

func TestSetAuthentication(t *testing.T) {
	assert := assert.New(t)

	user := "foo"
	pass := "bar"
	u := "ws://127.0.0.1"
	conf := NewClientConfig(u)
	conf.SetAuthentication(user, pass)
	assert.Equal(nil, conf.AuthReq)
}

func TestSetDebug(t *testing.T) {
	assert := assert.New(t)

	u := "ws://127.0.0.1"
	conf := NewClientConfig(u)
	conf.SetDebug()
	assert.Equal(true, conf.Debug)
}

func TestSetVerbose(t *testing.T) {
	assert := assert.New(t)

	u := "ws://127.0.0.1"
	conf := NewClientConfig(u)
	conf.SetVerbose()
	assert.Equal(true, conf.Verbose)
}
func TestSetTimeout(t *testing.T) {
	assert := assert.New(t)

	u := "ws://127.0.0.1"
	conf := NewClientConfig(u)
	conf.SetTimeout(22)
	timeout := time.Duration(22000000000)
	assert.Equal(timeout, conf.Timeout)
}

func TestSetPingInterval(t *testing.T) {
	assert := assert.New(t)

	u := "ws://127.0.0.1"
	conf := NewClientConfig(u)
	conf.SetPingInterval(22)
	interval := time.Duration(22000000000)
	assert.Equal(interval, conf.PingInterval)
}

func TestSetWritingWait(t *testing.T) {
	assert := assert.New(t)

	u := "ws://127.0.0.1"
	conf := NewClientConfig(u)
	conf.SetWritingWait(22)
	wait := time.Duration(22000000000)
	assert.Equal(wait, conf.WritingWait)
}

func TestSetReadingWait(t *testing.T) {
	assert := assert.New(t)

	u := "ws://127.0.0.1"
	conf := NewClientConfig(u)
	conf.SetReadingWait(22)
	wait := time.Duration(22000000000)
	assert.Equal(wait, conf.ReadingWait)
}

func TestSetLogger(t *testing.T) {
	assert := assert.New(t)

	log := logger.New()
	u := "ws://127.0.0.1"
	conf := NewClientConfig(u)
	conf.SetLogger(log)
	assert.Equal(log, conf.Logger)
}
