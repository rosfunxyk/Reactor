package connection

import (
	"errors"

	"github.com/rosfunxyk/Reactor/eventloop"
	"github.com/rosfunxyk/Reactor/sync/atomic"
)

type CallBack interface {
	OnMessage(c *Connection, ctx interface{}, data []byte) interface{}
	OnClose(c *Connection)
}

// Connection TCP 连接
type Connection struct {
	fd            int
	connected     atomic.Bool
	buffer        []byte
	outBuffer     []byte // write buffer
	inBuffer      []byte // read buffer
	outBufferLen  atomic.Int64
	inBufferLen   atomic.Int64
	readHandler_  CallBack
	writeHandler_ CallBack
	errorHandler_ CallBack
	connHandler_  CallBack
	loop          *eventloop.EventLoop
	peerAddr      string
	ctx           interface{}
}

var ErrConnectionClosed = errors.New("connection closed")
