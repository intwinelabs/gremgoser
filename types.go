package gremgoser

import (
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/intwinelabs/logger"
)

var (
	ErrorInterfaceHasNoIdField       = errors.New("gremgoser: the passed interface must have an Id field")
	ErrorNoGraphTags                 = errors.New("gremgoser: the passed interface has no graph tags")
	ErrorUnsupportedPropertyMap      = errors.New("gremgoser: unsupported property map")
	ErrorCannotCastProperty          = errors.New("gremgoser: passed property cannot be cast")
	ErrorWSConnection                = errors.New("gremgoser: error connecting to websocket")
	ErrorWSConnectionNil             = errors.New("gremgoser: error websocket connection nil")
	ErrorConnectionDisposed          = errors.New("gremgoser: you cannot write on a disposed connection")
	ErrorInvalidURI                  = errors.New("gremgoser: invalid uri supplied in config")
	ErrorNoAuth                      = errors.New("gremgoser: client does not have a secure dialer for authentication with the server")
	Error401Unauthorized             = errors.New("gremgoser: UNAUTHORIZED")
	Error407Authenticate             = errors.New("gremgoser: AUTHENTICATE")
	Error498MalformedRequest         = errors.New("gremgoser: MALFORMED REQUEST")
	Error499InvalidRequestArguments  = errors.New("gremgoser: INVALID REQUEST ARGUMENTS")
	Error500ServerError              = errors.New("gremgoser: SERVER ERROR")
	Error597ScriptEvaluationError    = errors.New("gremgoser: SCRIPT EVALUATION ERROR")
	Error598ServerTimeout            = errors.New("gremgoser: SERVER TIMEOUT")
	Error599ServerSerializationError = errors.New("gremgoser: SERVER SERIALIZATION ERROR")
	ErrorUnknownCode                 = errors.New("gremgoser: UNKNOWN ERROR")
)

// ClientConfig configs a client
type ClientConfig struct {
	URI          string
	AuthReq      *GremlinRequest
	Debug        bool
	Verbose      bool
	Timeout      time.Duration
	PingInterval time.Duration
	WritingWait  time.Duration
	ReadingWait  time.Duration
	Logger       *logger.Logger
}

// Client is a container for the gremgoser client.
type Client struct {
	conf             *ClientConfig
	conn             dialer
	requests         chan []byte
	responses        chan []byte
	errs             chan error
	results          *sync.Map
	responseNotifier *sync.Map // responseNotifier notifies the requester that a response has arrived for the request
	respMutex        *sync.Mutex
	Errored          bool
}

// Ws is the dialer for a WebSocket connection
type Ws struct {
	uri          string
	conn         *websocket.Conn
	disposed     bool
	connected    bool
	pingInterval time.Duration
	writingWait  time.Duration
	readingWait  time.Duration
	timeout      time.Duration
	quit         chan struct{}
	sync.RWMutex
}

// GremlinRequest is a container for all evaluation request parameters to be sent to the Gremlin Server.
type GremlinRequest struct {
	RequestId uuid.UUID              `json:"requestId,string"`
	Op        string                 `json:"op"`
	Processor string                 `json:"processor"`
	Args      map[string]interface{} `json:"args"`
}

type GremlinResponse struct {
	RequestId uuid.UUID     `json:"requestId,string"`
	Status    GremlinStatus `json:"status"`
	Result    GremlinResult `json:"result"`
}

type GremlinStatus struct {
	Code       int                     `json:"code"`
	Attributes GremlinStatusAttributes `json:"attributes"`
	Message    string                  `json:"message"`
}

type GremlinStatusAttributes struct {
	XMsStatusCode         int     `json:"x-ms-status-code"`
	XMsRequestCharge      float32 `json:"x-ms-request-charge"`
	XMsTotalRequestCharge float32 `json:"x-ms-total-request-charge"`
}

type GremlinResult struct {
	Data []*GremlinData `json:"data"`
	Meta interface{}    `json:"meta"`
}

type GremlinData struct {
	Id         uuid.UUID              `json:"id"`
	Label      string                 `json:"label"`
	Type       string                 `json:"type"`
	InVLablel  string                 `json:"inVLabel"`
	OutVLablel string                 `json:"outVLabel"`
	InV        uuid.UUID              `json"inV"`
	OutV       uuid.UUID              `json"outV"`
	Properties map[string]interface{} `json:"properties"`
}

type GremlinProperty struct {
	Id    uuid.UUID   `json:"id"`
	Value interface{} `json:"value"`
}
