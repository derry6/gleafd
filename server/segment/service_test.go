package segment

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/derry6/gleafd/pkg/log"
)

type testRepo struct {
	segs []*Segment
	sync.RWMutex
}

func (r *testRepo) List(ctx context.Context) (segs []*Segment, err error) {
	r.RLock()
	defer r.RUnlock()
	for _, pSeg := range r.segs {
		seg := *pSeg
		segs = append(segs, &seg)
	}
	return segs, nil
}
func (r *testRepo) Get(ctx context.Context, biztag string) (*Segment, error) {
	r.RLock()
	defer r.RUnlock()
	for _, pSeg := range r.segs {
		if pSeg.BizTag == biztag {
			seg := *pSeg
			return &seg, nil
		}
	}
	return nil, errors.New("biztag not in test repo")
}
func (r *testRepo) UpdateMaxID(ctx context.Context, biztag string) (*Segment, error) {
	r.Lock()
	defer r.Unlock()
	for _, pSeg := range r.segs {
		if pSeg.BizTag == biztag {
			pSeg.MaxID += int64(pSeg.Step)
			seg := *pSeg
			return &seg, nil
		}
	}
	return nil, errors.New("biztag not found in test repo")
}
func (r *testRepo) UpdateMaxIDWithStep(ctx context.Context, biztag string, step int32) (*Segment, error) {
	r.Lock()
	defer r.Unlock()
	for _, pSeg := range r.segs {
		if pSeg.BizTag == biztag {
			pSeg.MaxID += int64(step)
			seg := *pSeg
			return &seg, nil
		}
	}
	return nil, errors.New("biztag not found in test repo")
}
func (r *testRepo) ListBizTags(ctx context.Context) (tags []string, err error) {
	r.RLock()
	defer r.RUnlock()
	for _, pSeg := range r.segs {
		tags = append(tags, pSeg.BizTag)
	}
	return tags, nil
}

func TestServiceGet(t *testing.T) {
	ts := time.Now()
	rand.Seed(ts.UnixNano())

	repo := &testRepo{
		segs: []*Segment{
			&Segment{"biztag1", 1, 500, "", ts},
			&Segment{"biztag2", 2001, 1000, "", ts},
			&Segment{"biztag3", 4001, 4000, "", ts},
			&Segment{"biztag4", 5, 200, "", ts},
		},
	}
	svc := NewService(repo, log.DefaultLogger)

	results := [4][]int64{
		/*biztag1*/ {1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
		/*biztag2*/ {2001, 2002, 2003, 2004, 2005, 2006, 2007, 2008, 2009, 2010},
		/*biztag3*/ {4001, 4002, 4003, 4004, 4005, 4006, 4007, 4008, 4009, 4010},
		/*biztag4*/ {5, 6, 7, 8, 9, 10, 11, 12, 13, 14},
	}

	for i := 0; i < 4; i++ {
		n := rand.Intn(9) + 1
		tag := fmt.Sprintf("biztag%d", i+1)
		ids, err := svc.Get(context.Background(), tag, n)
		if err != nil {
			t.Fatal(err)
		}
		if len(ids) != n {
			t.Fatalf("biztag = %s, i = %d, len = %d, want = %d", tag, i, len(ids), n)
		}
		min := int(math.Min(float64(len(results[i])), float64(len(ids))))
		for x := 0; x < min; x++ {
			if results[i][x] != ids[x] {
				t.Fatalf("biztag = %s i = %d, x = %d, v = %d, want = %d",
					tag, i, x, ids[x], results[i][x])
			}
		}
	}
	svc.Close()
}

func BenchmarkServiceGet(b *testing.B) {
	ts := time.Now()
	rand.Seed(ts.UnixNano())
	repo := &testRepo{segs: []*Segment{&Segment{"biztag1", 1, 1000, "", ts}}}
	svc := NewService(repo, log.DefaultLogger)
	for i := 0; i < b.N; i++ {
		ids, err := svc.Get(context.Background(), "biztag1", 10)
		if err != nil {
			b.Fatal(err)
		}
		if len(ids) != 10 {
			b.Fatalf("len(ids) = %d, want = 10", len(ids))
		}
	}
	svc.Close()
}

func BenchmarkServiceGetParallel(b *testing.B) {
	ts := time.Now()
	rand.Seed(ts.UnixNano())
	repo := &testRepo{segs: []*Segment{&Segment{"biztag1", 1, 1000, "", ts}}}
	svc := NewService(repo, log.DefaultLogger)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			ids, err := svc.Get(context.Background(), "biztag1", 10)
			if err != nil {
				b.Fatal(err)
			}
			if len(ids) != 10 {
				b.Fatalf("len(ids) = %d, want = 10", len(ids))
			}
		}
	})
	svc.Close()
}
