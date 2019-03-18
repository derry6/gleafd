package server

import (
	"context"
	"time"

	"github.com/derry6/gleafd/pkg/log"
)

type Midware func(svc Service) Service

type LoggingMidware struct {
	logger log.Logger
	Service
}

func (m *LoggingMidware) GetSegments(ctx context.Context, biztag string, count int) (ids []int64, err error) {
	defer func(begin time.Time) {
		m.logger.Infow("GetSegments",
			"biztag", biztag,
			"count", count,
			"results", ids,
			"err", err,
			"elapsed", time.Now().Sub(begin),
		)
	}(time.Now())
	ids, err = m.Service.GetSegments(ctx, biztag, count)
	return
}

func (m *LoggingMidware) GetSnowflakes(ctx context.Context, biztag string, count int) (ids []int64, err error) {
	defer func(begin time.Time) {
		m.logger.Infow("GetSnowflakes",
			"biztag", biztag,
			"count", count,
			"results", ids,
			"err", err,
			"elapsed", time.Now().Sub(begin),
		)
	}(time.Now())
	ids, err = m.Service.GetSnowflakes(ctx, biztag, count)
	return
}

func (m *LoggingMidware) HealthCheck(ctx context.Context, name string) (status int, err error) {
	defer func(begin time.Time) {
		m.logger.Infow("HealthCheck",
			"name", name,
			"status", status,
			"err", err,
			"elapsed", time.Now().Sub(begin),
		)
	}(time.Now())
	status, err = m.Service.HealthCheck(ctx, name)
	return
}

func Logging(logger log.Logger, svc Service) Service {
	m := &LoggingMidware{
		Service: svc,
		logger:  logger,
	}
	return m
}
