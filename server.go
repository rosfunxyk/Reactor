package Reactor

import (
	"Project/Reactor/sync"
	"errors"
	"runtime"

	"Project/Reactor/connection"
	"Project/Reactor/eventloop"
	"Project/Reactor/listener"
	"Project/Reactor/log"
	"Project/Reactor/sync/atomic"

	"golang.org/x/sys/unix"
)

// Handler Server 注册接口
type Handler interface {
	connection.CallBack
	OnConnect(c *connection.Connection)
}

// Server gev Server
type Server struct {
	loop       *eventloop.EventLoop
	thread_num int
	workLoops  []*subreactor
	callback   Handler

	opts    *Options
	running atomic.Bool
}

// NewServer 创建 Server
func NewServer(handler Handler, opts ...Option) (server *Server, err error) {
	if handler == nil {
		return nil, errors.New("handler is nil")
	}
	options := newOptions(opts...)
	server = new(Server)
	server.callback = handler
	server.opts = options
	server.loop, err = eventloop.New()
	if err != nil {
		_ = server.loop.Stop()
		return nil, err
	}

	l, err := listener.New(server.opts.Network, server.opts.Address, server.loop, server.handleNewConnection)
	if err != nil {
		return nil, err
	}
	if err = server.loop.AddSocketAndEnableRead(l.Fd(), l); err != nil {
		return nil, err
	}

	if server.opts.NumLoops <= 0 {
		server.opts.NumLoops = runtime.NumCPU()
	}

	wloops := make([]*subreactor, server.opts.NumLoops)
	for i := 0; i < server.opts.NumLoops; i++ {
		l, err := eventloop.New()
		if err != nil {
			for j := 0; j < i; j++ {
				_ = wloops[j].Stop()
			}
			return nil, err
		}
		wloops[i] = l
	}
	server.workLoops = wloops

	return
}

func (s *Server) handleNewConnection(fd int, sa unix.Sockaddr) {
	loop := s.opts.Strategy(s.workLoops)

	c := connection.New(fd, loop, sa, s.callback)

	loop.QueueInLoop(func() {
		s.callback.OnConnect(c)
		if err := loop.AddSocketAndEnableRead(fd, c); err != nil {
			log.Error("[AddSocketAndEnableRead]", err)
		}
	})
}

// Start 启动 Server
func (s *Server) Start() {
	sw := sync.WaitGroupWrapper{}

	length := len(s.workLoops)
	for i := 0; i < length; i++ {
		sw.AddAndRun(s.workLoops[i].RunLoop)
	}

	sw.AddAndRun(s.loop.RunLoop)
	s.running.Set(true)
	sw.Wait()
}
