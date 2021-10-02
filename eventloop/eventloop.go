package eventloop

import (
	"unsafe"

	"github.com/Allenxuxu/gev/poller"
	"github.com/Allenxuxu/toolkit/sync/atomic"
	"github.com/Allenxuxu/toolkit/sync/spinlock"
)

var (
	DefaultPacketSize    = 65536
	DefaultBufferSize    = 4096
	DefaultTaskQueueSize = 1024
)

// Socket 接口
type Socket interface {
	HandleEvent(fd int, events poller.Event)
	Close() error
}

// EventLoop 事件循环
type EventLoop struct {
	eventLoopLocal
	// nolint
	// Prevents false sharing on widespread platforms with
	// 128 mod (cache line size) = 0 .
	pad [128 - unsafe.Sizeof(eventLoopLocal{})%128]byte
}

// nolint
type eventLoopLocal struct {
	ConnCunt   atomic.Int64
	needWake   *atomic.Bool
	poll       *poller.Poller
	mu         spinlock.SpinLock
	sockets    map[int]Socket
	packet     []byte
	taskQueueW []func()
	taskQueueR []func()

	UserBuffer *[]byte
}
