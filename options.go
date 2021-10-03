package Reactor

import (
	"time"
)

// Options 服务配置
type Options struct {
	Network  string
	Address  string
	NumLoops int
	IdleTime time.Duration
	Strategy LoadBalanceStrategy

	tick                        time.Duration
	wheelSize                   int64
	metricsPath, metricsAddress string
}

// Option ...
type Option func(*Options)

func newOptions(opt ...Option) *Options {
	opts := Options{}

	for _, o := range opt {
		o(&opts)
	}

	if opts.Network == "" {
		opts.Network = "tcp"
	}
	if opts.Address == "" {
		opts.Address = ":1388"
	}
	if opts.tick == 0 {
		opts.tick = 1 * time.Millisecond
	}
	if opts.wheelSize == 0 {
		opts.wheelSize = 1000
	}
	if opts.Strategy == nil {
		opts.Strategy = RoundRobin()
	}

	return &opts
}

// Network [tcp] 暂时只支持tcp
func Network(n string) Option {
	return func(o *Options) {
		o.Network = n
	}
}

// Address server 监听地址
func Address(a string) Option {
	return func(o *Options) {
		o.Address = a
	}
}

// NumLoops work eventloop 的数量
func NumLoops(n int) Option {
	return func(o *Options) {
		o.NumLoops = n
	}
}

// IdleTime 最大空闲时间（秒）
func IdleTime(t time.Duration) Option {
	return func(o *Options) {
		o.IdleTime = t
	}
}

func LoadBalance(strategy LoadBalanceStrategy) Option {
	return func(o *Options) {
		o.Strategy = strategy
	}
}

func MetricsServer(path, address string) Option {
	return func(o *Options) {
		o.metricsPath = path
		o.metricsAddress = address
	}
}
