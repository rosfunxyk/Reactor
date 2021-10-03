package Reactor

import (
	"Project/Reactor/connection"
	"Project/Reactor/eventloop"
	"Project/Reactor/sync/atomic"
)

type subreactor struct {
	loop     *eventloop.EventLoop
	callback Handler

	opts    *Options
	running atomic.Bool
}

func (s *subreactor) OnConnect(c *connection.Connection) {
	//log.Println(" OnConnect ： ", c.PeerAddr())
}
func (s *subreactor) OnMessage(c *connection.Connection, ctx interface{}, data []byte) (out []byte) {
	//log.Println("OnMessage")
	out = data
	return
}

func (s *subreactor) OnClose(c *connection.Connection) {
	//log.Println("OnClose")
}

func (s *subreactor) Stop() {
	//log.Println("OnClose")
	s.loop.Stop()
}

// New 创建一个 EventLoop
func New(handler Handler, opts ...Option) (subreactor *subreactor, err error) {
	if err != nil {
		return nil, err
	}
	subreactor.loop, err = eventloop.New()
	return subreactor, err
}
