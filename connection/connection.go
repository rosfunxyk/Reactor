package connection

import (
	"Project/Reactor/eventloop"
	"Project/Reactor/log"
	"Project/Reactor/sync/atomic"
	"errors"
	"fmt"
	"net"
	"strconv"

	"github.com/Allenxuxu/gev/poller"
	"golang.org/x/sys/unix"
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
	callBack      CallBack
	loop          *eventloop.EventLoop
	peerAddr      string
	ctx           interface{}
}

var ErrConnectionClosed = errors.New("connection closed")

// New 创建 Connection
func New(fd int,
	loop *eventloop.EventLoop,
	sa unix.Sockaddr,
	callBack CallBack) *Connection {
	conn := &Connection{
		fd:        fd,
		peerAddr:  sockAddrToString(sa),
		outBuffer: []byte{},
		inBuffer:  []byte{},
		callBack:  callBack,
		loop:      loop,
		buffer:    []byte{},
	}
	conn.connected.Set(true)

	return conn
}

func (c *Connection) UserBuffer() *[]byte {
	return c.loop.UserBuffer
}

// PeerAddr 获取客户端地址信息
func (c *Connection) PeerAddr() string {
	return c.peerAddr
}

// Connected 是否已连接
func (c *Connection) Connected() bool {
	return c.connected.Get()
}

// Send 用来在非 loop 协程发送
func (c *Connection) Send(data interface{}, opts ...Option) error {
	if !c.connected.Get() {
		return ErrConnectionClosed
	}

	opt := Options{}
	for _, o := range opts {
		o(&opt)
	}

	c.loop.QueueInLoop(func() {
		if c.connected.Get() {
			c.sendInLoop(data)

			if opt.sendInLoopFinish != nil {
				opt.sendInLoopFinish(data)
			}
		}
	})
	return nil
}

// Close 关闭连接
func (c *Connection) Close() error {
	if !c.connected.Get() {
		return ErrConnectionClosed
	}

	c.loop.QueueInLoop(func() {
		c.handleClose(c.fd)
	})
	return nil
}

// ShutdownWrite 关闭可写端，等待读取完接收缓冲区所有数据
func (c *Connection) ShutdownWrite() error {
	return unix.Shutdown(c.fd, unix.SHUT_WR)
}

// ReadBufferLength read buffer 当前积压的数据长度
func (c *Connection) ReadBufferLength() int64 {
	return c.inBufferLen.Get()
}

// WriteBufferLength write buffer 当前积压的数据长度
func (c *Connection) WriteBufferLength() int64 {
	return c.outBufferLen.Get()
}

// HandleEvent 内部使用，event loop 回调
func (c *Connection) HandleEvent(fd int, events poller.Event) {

	if events&poller.EventErr != 0 {
		c.handleClose(fd)
		return
	}

	if len(c.outBuffer) == 0 {
		if events&poller.EventWrite != 0 {
			// if return true, it means closed
			if c.handleWrite(fd) {
				return
			}

			if len(c.outBuffer) == 0 {
				c.outBuffer = []byte{}
			}
		}
	} else if events&poller.EventRead != 0 {
		// if return true, it means closed
		if c.handleRead(fd) {
			return
		}

		if len(c.inBuffer) == 0 {
			c.inBuffer = []byte{}
		}
	}

	c.inBufferLen.Swap(int64(len(c.inBuffer)))
	c.outBufferLen.Swap(int64(len(c.outBuffer)))
}

func (c *Connection) handleRead(fd int) (closed bool) {
	// TODO 避免这次内存拷贝
	buf := c.loop.PacketBuf()
	n, err := unix.Read(c.fd, c.outBuffer)
	if n == 0 || err != nil {
		if err != unix.EAGAIN {
			c.handleClose(fd)
			closed = true
		}
		return
	}

	if len(buf) != 0 {
		closed = c.sendInLoop(buf)
	}
	return
}

func (c *Connection) handleWrite(fd int) (closed bool) {
	_, err := unix.Write(c.fd, first)
	if err != nil {
		if err == unix.EAGAIN {
			return
		}
		c.handleClose(fd)
		closed = true
		return
	}

	return
}

func (c *Connection) handleClose(fd int) {
	if c.connected.Get() {
		c.connected.Set(false)
		c.loop.DeleteFdInLoop(fd)

		if err := unix.Close(fd); err != nil {
			log.Error("[close fd]", err)
		}

	}
}

func (c *Connection) sendInLoop(data []byte) (closed bool) {
	_, err := unix.Write(c.fd, data)
	if err != nil && err != unix.EAGAIN {
		c.handleClose(c.fd)
		closed = true
		return
	}

	return
}

func sockAddrToString(sa unix.Sockaddr) string {
	switch sa := (sa).(type) {
	case *unix.SockaddrInet4:
		return net.JoinHostPort(net.IP(sa.Addr[:]).String(), strconv.Itoa(sa.Port))
	case *unix.SockaddrInet6:
		return net.JoinHostPort(net.IP(sa.Addr[:]).String(), strconv.Itoa(sa.Port))
	default:
		return fmt.Sprintf("(unknown - %T)", sa)
	}
}
