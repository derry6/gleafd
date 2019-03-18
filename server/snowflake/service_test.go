package snowflake

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/derry6/gleafd/pkg/log"
)

type testStorage struct {
	metadatas map[string]*Metadata
	sync.RWMutex
	machindId int32
}

func (s *testStorage) key(name, addr string) string {
	return fmt.Sprintf("%s@%s", name, addr)
}

// 查找指定的metadata,不存在则创建并返回合适的machineID
func (r *testStorage) GetOrNew(ctx context.Context, name, addr string) (md Metadata, err error) {
	r.RLock()
	m, ok := r.metadatas[r.key(name, addr)]
	r.RUnlock()
	if ok {
		return *m, nil
	}
	md.Name = name
	md.Addr = addr
	md.MachineID = int(atomic.AddInt32(&r.machindId, 1))
	r.Lock()
	defer r.Unlock()
	r.metadatas[r.key(name, addr)] = &Metadata{Name: name, Addr: addr, MachineID: md.MachineID}
	return md, nil
}

// 获取所有的metadata
func (r *testStorage) List(ctx context.Context) (mds []Metadata, err error) {
	r.RLock()
	defer r.RUnlock()
	for _, v := range r.metadatas {
		mds = append(mds, *v)
	}
	return mds, nil
}

// 更新时间戳
func (r *testStorage) Update(ctx context.Context, md Metadata) (err error) {
	r.RLock()
	defer r.RUnlock()
	m, ok := r.metadatas[r.key(md.Name, md.Addr)]
	if ok {
		m.Timestamp = md.Timestamp
		return nil
	}
	return errors.New("not exists")
}

var testStore = &testStorage{
	metadatas: make(map[string]*Metadata),
	machindId: -1,
}

func TestServiceNew(t *testing.T) {
	svc := NewService("gleafd0", "127.0.0.1:8090", testStore, log.DefaultLogger)
	defer svc.Close()
	t.Logf("machineId = %d", svc.md.MachineID)
	if svc.md.MachineID != 0 {
		t.Fatalf("machine id = %d want = 0", svc.md.MachineID)
	}
}

func TestServiceGet(t *testing.T) {
	svc := NewService("gleafd0", "127.0.0.1:8090", testStore, log.DefaultLogger)
	defer svc.Close()
	var lastMax int64
	// 单个routine是递增ID
	for i := 0; i < 10; i++ {
		ids, err := svc.Get(context.Background(), "example", i+1)
		if err != nil {
			t.Fatal(err)
		}
		if len(ids) != i+1 {
			t.Fatalf("len(ids) = %d, want = %d", len(ids), i+1)
		}
		if ids[0] <= lastMax {
			t.Fatalf("id[0] = %d, lastMax = %d: id <= lastMax", ids[0], lastMax)
		}
		lastMax = ids[len(ids)-1]
		var prv int64
		for _, id := range ids {
			if id <= prv {
				t.Fatalf("id = %d, prv = %d: id <= prv", id, lastMax)
			}
			prv = id
		}
	}
}
