package server

import (
	"context"

	"github.com/derry6/gleafd/server/segment"
	"github.com/derry6/gleafd/server/snowflake"
)

type Service interface {
	GetSegments(ctx context.Context, biztag string, count int) (ids []int64, err error)
	GetSnowflakes(ctx context.Context, biztag string, count int) (ids []int64, err error)
	HealthCheck(ctx context.Context, name string) (status int, err error)
	Close() error
}

type gleafService struct {
	name    string
	segsvc  *segment.Service
	snowsvc *snowflake.Service
}

func (glfs *gleafService) GetSegments(ctx context.Context, biztag string, count int) (ids []int64, err error) {
	if glfs.segsvc == nil {
		return nil, ErrServiceDisabled
	}
	return glfs.segsvc.Get(ctx, biztag, count)
}

func (glfs *gleafService) GetSnowflakes(ctx context.Context, biztag string, count int) (ids []int64, err error) {
	if glfs.snowsvc == nil {
		return nil, ErrServiceDisabled
	}
	return glfs.snowsvc.Get(ctx, biztag, count)
}

func (glfs *gleafService) HealthCheck(ctx context.Context, name string) (status int, err error) {
	return 1, nil
}

func (glfs *gleafService) Close() (err error) {
	if glfs.snowsvc != nil {
		glfs.snowsvc.Close()
	}
	if glfs.segsvc != nil {
		glfs.segsvc.Close()
	}
	return nil
}

func NewService(opts ...Option) Service {
	sopts := newDefaultOptions()
	for _, o := range opts {
		o(sopts)
	}
	glfsvc := &gleafService{name: sopts.name}

	if sopts.repo != nil {
		// segment service
		segsvc := segment.NewService(sopts.repo, sopts.logger)
		glfsvc.segsvc = segsvc
	}
	if sopts.stor != nil {
		// snowflake service
		snowsvc := snowflake.NewService(sopts.name, sopts.addr, sopts.stor, sopts.logger)
		glfsvc.snowsvc = snowsvc
	}
	var s Service = glfsvc
	for _, mdw := range sopts.mdws {
		s = mdw(s)
	}
	return s
}
