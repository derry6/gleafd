package snowflake

import (
	"errors"
	"math/rand"
	"time"
)

var (
	epoch = time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC).UnixNano() / 1000000
)

var (
	ErrClockMoveBackwards = errors.New("system clock move backwards")
	ErrInvalidMachineID   = errors.New("invalid machine id")
)

// 1+41+10+12
const (
	MachineIDBits  uint8 = 10
	SeqBits        uint8 = 12
	MachineIDShift       = SeqBits
	TimeShift            = SeqBits + MachineIDBits
	MachineIDMax   int   = -1 ^ (-1 << MachineIDBits)
	SeqMask        int32 = -1 ^ (-1 << SeqBits)
)

type Factory interface {
	Next() (int64, error)
}

type factory struct {
	machineID int
	lastTs    int64
	seq       int32
	rnd       *rand.Rand
}

// 不支持多routine并发
func NewFactory(machineID int) (Factory, error) {
	if machineID < 0 || MachineIDMax < machineID {
		return nil, ErrInvalidMachineID
	}
	nano := time.Now().UTC().UnixNano()
	rnd := rand.New(rand.NewSource(nano))
	return &factory{
		machineID: machineID,
		lastTs:    0,
		seq:       0,
		rnd:       rnd,
	}, nil
}

func (sf *factory) Next() (int64, error) {
	ts := sf.nowMs()
	if ts < sf.lastTs { // 时钟回溯的问题
		offset := sf.lastTs - ts
		if offset > 5 {
			return int64(0), ErrClockMoveBackwards
		}
		// 等待两倍的时间差
		time.Sleep(time.Duration(offset<<1) * time.Millisecond)
		ts = sf.nowMs()
		if ts < sf.lastTs {
			return int64(0), ErrClockMoveBackwards
		}
	}
	// 同一毫秒内，随机数递增
	if ts == sf.lastTs {
		sf.seq = (sf.seq + 1) & SeqMask
		if sf.seq == 0 {
			ts = sf.waitNextTs()
		}
	} else {
		// 每一个毫秒开始是选择一个0到9随机数
		// 避免出现个位数为0的ID太多。
		sf.seq = int32(sf.rnd.Intn(10))
	}
	return sf.buildFinalId(ts)
}

func (sf *factory) buildFinalId(now int64) (int64, error) {
	sf.lastTs = now
	timestamp := (now - epoch) << TimeShift
	machineID := int64(sf.machineID << MachineIDShift)
	n := timestamp | machineID | int64(sf.seq)
	return int64(n), nil
}

func (sf *factory) waitNextTs() int64 {
	t := sf.nowMs()
	for t <= sf.lastTs {
		time.Sleep(100 * time.Microsecond) // sleep 100us
		t = sf.nowMs()
	}
	return t
}

func (sf *factory) nowMs() int64 {
	return time.Now().UnixNano() / 1000000
}
