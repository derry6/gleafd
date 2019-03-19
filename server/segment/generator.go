package segment

import (
	"context"
	"sync/atomic"
	"time"
)

type generator struct {
	svc        *Service
	biztag     string
	waits      chan *Segment
	next       chan int64
	closed     int32
	inited     int32
	closeC     chan struct{}
	minStep    int32
	curStep    int32
	lastUpdate time.Time
	total      int64
}

func newGenerator(svc *Service, biztag string, waits chan *Segment) *generator {
	return &generator{
		biztag: biztag,
		waits:  waits,
		next:   make(chan int64, 100),
		svc:    svc,
		closeC: make(chan struct{}),
	}
}

func (g *generator) get(ctx context.Context) (int64, error) {
	if atomic.LoadInt32(&g.closed) != 0 {
		return 0, ErrClosed
	}
	// 第一次获取可能会有些延时
	if atomic.CompareAndSwapInt32(&g.inited, 0, 1) {
		g.svc.notifyUpdate(g.biztag, g.curStep, g.waits)
		g.lastUpdate = time.Now()
	}
	select {
	case id, ok := <-g.next:
		if !ok {
			return 0, ErrClosed
		}
		return id, nil
	case <-ctx.Done():
		return 0, ctx.Err()
	}
}

func (g *generator) buildOneSegment(seg *Segment) {
	if g.minStep == 0 {
		g.minStep = seg.Step
	}
	start := seg.MaxID - int64(seg.Step)
	end := seg.MaxID
	pct75 := int64(float64(seg.Step)*0.75 + float64(start))

	for i := start; i < end; i++ {
		if i == pct75 { // 使用超过65%时，通知updater获取新号段
			now := time.Now()
			duration := now.Sub(g.lastUpdate)
			g.svc.logger.Infow("Updating", "biztag", g.biztag,
				"start", start, "end", end, "pct75", pct75, "step", g.curStep, "duration", duration)
			if duration <= 10*time.Minute {
				// 少于五分钟增大step
				step := seg.Step * 2
				if step > 1000000 {
					step = 1000000
				}
				g.curStep = step
			} else if duration >= 20*time.Minute {
				// 超过20分钟减小step
				step := g.curStep / 2
				if step < seg.Step {
					step = seg.Step
				}
				g.curStep = step
			} else {
				g.curStep = seg.Step
			}
			g.svc.notifyUpdate(g.biztag, g.curStep, g.waits)
			g.lastUpdate = time.Now()
		}
		select {
		case <-g.closeC: // FIXME:
			return
		case g.next <- i:
		}
		g.total++
		if g.total%1000000 == 0 {
			g.svc.logger.Infow("Generated", "biztag", g.biztag, "curid", i, "total", g.total)
		}
	}
}

// 每个biztag使用一个单独的routine负责
func (g *generator) run() {
	defer func() {
		close(g.next)
	}()
	for {
		select {
		case <-g.closeC:
			return
		case item, ok := <-g.waits:
			if !ok {
				return
			}
			g.buildOneSegment(item)
		}
	}
}

func (g *generator) stop() {
	if atomic.CompareAndSwapInt32(&g.closed, 0, 1) {
		close(g.closeC)
	}
}
