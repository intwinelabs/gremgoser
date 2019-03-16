package gremgoser

import (
	"time"

	"github.com/google/uuid"
	"github.com/intwinelabs/logger"
)

// NewClientConfig returns a default client config
func NewClientConfig(uri string) *ClientConfig {
	return &ClientConfig{
		URI:          uri,
		Timeout:      5 * time.Second,
		PingInterval: 60 * time.Second,
		WritingWait:  15 * time.Second,
		ReadingWait:  15 * time.Second,
	}
}

// SetAuthentication sets on dialer credentials for authentication
func (conf *ClientConfig) SetAuthentication(username string, password string) {
	conf.AuthReq = prepareAuthRequest(uuid.New(), username, password)
}

// SetDebug sets the debug flag
func (conf *ClientConfig) SetDebug() {
	conf.Debug = true
}

// SetVerbose sets the verbose flag
func (conf *ClientConfig) SetVerbose() {
	conf.Verbose = true
}

// SetTimeout sets the dial timeout
func (conf *ClientConfig) SetTimeout(seconds int) {
	conf.Timeout = time.Duration(seconds) * time.Second
}

// SetPingInterval sets the interval of ping sending for know is
// connection is alive and in consequence the client is connected
func (conf *ClientConfig) SetPingInterval(seconds int) {
	conf.PingInterval = time.Duration(seconds) * time.Second
}

// SetWritingWait sets the time for waiting that writing occur
func (conf *ClientConfig) SetWritingWait(seconds int) {
	conf.WritingWait = time.Duration(seconds) * time.Second
}

// SetReadingWait sets the time for waiting that reading occur
func (conf *ClientConfig) SetReadingWait(seconds int) {
	conf.ReadingWait = time.Duration(seconds) * time.Second
}

// SetLogger sets the default logger
func (conf *ClientConfig) SetLogger(logger *logger.Logger) {
	conf.Logger = logger
}
