package segment

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/derry6/gleafd/pkg/log"
)

var (
	ErrClosed         = errors.New("service closed")
	ErrBizTagNotFound = errors.New("biztag not found")
)

type waitItem struct {
	biztag string
	result chan *Segment
	step   int32
}

type Service struct {
	repo   Repository            // 仓储
	gs     map[string]*generator // 保存所有的generators
	gsMu   sync.RWMutex
	waits  chan waitItem // 等待更新的BizTags
	logger log.Logger
	closed int32 // 是否关闭Flag
	closeC chan struct{}
	wg     sync.WaitGroup
}

// 查找对应biztag的generator
func (s *Service) findGenerator(biztag string) (*generator, error) {
	s.gsMu.RLock()
	defer s.gsMu.RUnlock()
	g, ok := s.gs[biztag]
	if !ok {
		return nil, ErrBizTagNotFound
	}
	return g, nil
}

// 获取所有的generator
func (s *Service) getGenerators() (gs []*generator) {
	s.gsMu.RLock()
	defer s.gsMu.RUnlock()
	for _, g := range s.gs {
		gs = append(gs, g)
	}
	return gs
}

// 获取当前所有的biztags
func (s *Service) getBizTagsUnsafe() (tags []string) {
	for k, _ := range s.gs {
		tags = append(tags, k)
	}
	return tags
}

// 从数据库更新所有的biztags
func (s *Service) handleBizTagsUpdated(newTags []string) (added []string, removed []string) {
	s.gsMu.RLock()
	defer s.gsMu.RUnlock()
	oldTags := s.getBizTagsUnsafe()
	// cache new biztag
	for _, newTag := range newTags {
		found := false
		for _, oldTag := range oldTags {
			if oldTag == newTag {
				found = true
				break
			}
		}
		if !found { // 这是从数据库取出的新的biztag,需要创建对应的biztag
			added = append(added, newTag)
		}
	}
	for _, oldTag := range oldTags {
		found := false
		for _, newTag := range newTags {
			if newTag == oldTag {
				found = true
				break
			}
		}
		if !found {
			removed = append(removed, oldTag)
		}
	}
	return added, removed
}

// 从数据库中读取所有的biztags, 每分钟更新一次
func (s *Service) updateBizTagsFromRepo() error {
	tags, err := s.repo.ListBizTags(context.Background())
	if err != nil {
		return err
	}
	added, removed := s.handleBizTagsUpdated(tags)
	if len(added) > 0 {
		s.logger.Infow("Segment biztags added", "tags", added)
	}
	if len(removed) > 0 {
		s.logger.Infow("Segment biztags removed", "tags", removed)
	}
	// 删除对应的generators
	uscMap := make(map[string]chan *Segment)

	s.gsMu.Lock()
	for _, biztag := range removed {
		g, ok := s.gs[biztag]
		if ok {
			delete(s.gs, biztag)
		}
		if ok {
			g.stop()
		}
	}
	// 创建相应的generators
	for _, biztag := range added {
		// generator读
		usc := make(chan *Segment, 1)
		g := newGenerator(s, biztag, usc)
		s.wg.Add(1)
		go func() {
			defer s.wg.Done()
			g.run()
		}()
		uscMap[biztag] = usc
		s.gs[biztag] = g
	}
	s.gsMu.Unlock()
	return nil
}

func (s *Service) notifyUpdate(biztag string, step int32, result chan *Segment) {
	// waitUpdateBizTags/closeC 生命周期跟Service相同
	select {
	case <-s.closeC:
		return
	case s.waits <- waitItem{biztag, result, step}:
	}
}

func (s *Service) update(ws waitItem) error {
	var (
		seg *Segment
		err error
	)
	ctx := context.Background()
	if ws.step <= 0 {
		seg, err = s.repo.UpdateMaxID(ctx, ws.biztag)
	} else {
		seg, err = s.repo.UpdateMaxIDWithStep(ctx, ws.biztag, ws.step)
	}
	if err != nil {
		return err
	}
	// 如果usc被关闭？
	select {
	case <-s.closeC:
		return ErrClosed
	case ws.result <- seg:
		return nil
	}
}

// 负责从数据库中拉取数据
func (s *Service) run() error {
	timer := time.NewTicker(time.Minute)
	for {
		select {
		case <-s.closeC:
			return ErrClosed
		case item, ok := <-s.waits:
			if !ok {
				return ErrClosed
			}
			if err := s.update(item); err != nil {
			}
		case <-timer.C:
			s.updateBizTagsFromRepo()
		}
	}
}

func (s *Service) init() error {
	return s.updateBizTagsFromRepo()
}

func (s *Service) Get(ctx context.Context, biztag string, count int) (ids []int64, err error) {
	g, err := s.findGenerator(biztag)
	if err != nil {
		return nil, err
	}
	for i := 0; i < count; i++ {
		id, err := g.get(ctx)
		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func (s *Service) Close() error {
	if atomic.CompareAndSwapInt32(&s.closed, 0, 1) {
		// stopping all generators
		gs := s.getGenerators()
		for _, g := range gs {
			g.stop()
		}
		close(s.closeC)
		close(s.waits)
		s.wg.Wait()
	}
	return nil
}

func NewService(repo Repository, logger log.Logger) *Service {
	s := &Service{
		repo:   repo,
		gs:     make(map[string]*generator),
		waits:  make(chan waitItem, 100),
		closeC: make(chan struct{}),
		logger: logger,
	}
	if err := s.init(); err != nil {
		logger.Fatalw("New segment service", "err", err)
	}
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		if err := s.run(); err != nil {
			return
		}
	}()
	return s
}
