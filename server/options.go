package server

import (
	"github.com/derry6/gleafd/pkg/log"
	"github.com/derry6/gleafd/server/segment"
	"github.com/derry6/gleafd/server/snowflake"
)

type Options struct {
	name string
	addr string
	mdws []Midware

	logger log.Logger
	// Segment
	repo segment.Repository
	// Snowflake
	stor snowflake.Storage
}

func newDefaultOptions() *Options {
	return &Options{
		name: "gleafd",
		addr: "127.0.0.1:8090",
		mdws: make([]Midware, 0),
	}
}

type Option func(opts *Options)

func WithLogger(logger log.Logger) Option {
	return func(opts *Options) {
		if logger == nil {
			logger = log.DefaultLogger
		}
		opts.logger = logger
	}
}

func WithName(name string) Option {
	return func(opts *Options) {
		opts.name = name
	}
}

func WithAddr(addr string) Option {
	return func(opts *Options) {
		opts.addr = addr
	}
}

func WithMidwares(mdws []Midware) Option {
	return func(opts *Options) {
		opts.mdws = mdws
	}
}

func WithSegmentRepository(repo segment.Repository) Option {
	return func(opts *Options) {
		opts.repo = repo
	}
}

func WithSnowflakeStorage(stor snowflake.Storage) Option {
	return func(opts *Options) {
		opts.stor = stor
	}
}
