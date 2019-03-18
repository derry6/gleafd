package snowflake

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/derry6/gleafd/pkg/log"
)

type Service struct {
	md     Metadata
	stor   Storage
	fs     chan Factory // 使用chan 代替使用锁
	logger log.Logger
	wg     sync.WaitGroup
	closed int32 // 退出标记
	closeC chan struct{}
}

func (s *Service) start() error {
	// 第一次启动
	if err := s.update(); err != nil {
		return err
	}
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		if err := s.run(); err != nil {
		}
	}()
	return nil
}

func (s *Service) isValidMachineID(id int) bool {
	return id >= 0 && id < 1023
}

func (s *Service) init() error {
	// 从storage中读取
	md, err := s.stor.GetOrNew(context.Background(), s.md.Name, s.md.Addr)
	if err != nil {
		return err
	}
	if s.isValidMachineID(md.MachineID) {
		s.md.MachineID = md.MachineID
		// 检查时间
		if md.Timestamp > s.nowMs() {
			return fmt.Errorf("last update time greate than current time")
		}
		return s.start()
	} else {
		return fmt.Errorf("invalid machine id: %v", s.md.MachineID)
	}
}

func (s *Service) run() error {
	timer := time.NewTicker(3 * time.Second)
	for {
		select {
		case <-s.closeC:
			return errors.New("service closed")
		case <-timer.C:
			if err := s.update(); err != nil {
			}
		}
	}
}

func (s *Service) nowMs() int64 {
	return time.Now().UnixNano() / 1000000
}

func (s *Service) update() error {
	now := s.nowMs()
	if s.md.Timestamp > now {
		return nil
	}
	s.md.Timestamp = now
	return s.stor.Update(context.Background(), s.md)
}

func (s *Service) Close() error {
	if atomic.CompareAndSwapInt32(&s.closed, 0, 1) {
		close(s.closeC)
		close(s.fs)
		s.wg.Wait()
	}
	return nil
}

func (s *Service) Get(ctx context.Context, biztag string, count int) (ids []int64, err error) {
	if atomic.LoadInt32(&s.closed) == 1 {
		return nil, errors.New("service closed")
	}
	gen := func() (id int64, er error) {
		var f Factory
		select {
		case f = <-s.fs:
		case <-ctx.Done():
			return 0, ctx.Err()
		}
		defer func() {
			select {
			case s.fs <- f:
			case <-ctx.Done():
				id = 0
				er = ctx.Err()
				return
			}
		}()
		return f.Next()
	}
	for i := 0; i < count; i++ {
		id, err := gen()
		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func NewService(name, addr string, storage Storage, logger log.Logger) *Service {
	s := &Service{
		md: Metadata{
			Name: name,
			Addr: addr,
		},
		stor:   storage,
		closeC: make(chan struct{}),
		fs:     make(chan Factory, 1),
		logger: logger,
	}
	if err := s.init(); err != nil {
		logger.Fatalw("New snowflake service", "err", err)
	}
	f, _ := NewFactory(s.md.MachineID)
	s.fs <- f
	return s
}
