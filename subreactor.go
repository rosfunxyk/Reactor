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

type example struct{}

func (s *example) OnConnect(c *connection.Connection) {
	//log.Println(" OnConnect ： ", c.PeerAddr())
}
func (s *example) OnMessage(c *connection.Connection, ctx interface{}, data []byte) (out []byte) {
	//log.Println("OnMessage")
	out = data
	return
}

func (s *example) OnClose(c *connection.Connection) {
	//log.Println("OnClose")
}
