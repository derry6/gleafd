package snowflake

import (
	"context"
	"fmt"
	"strings"

	"github.com/derry6/gleafd/pkg/log"

	"github.com/gomodule/redigo/redis"
)

type Storage interface {
	// 查找指定的metadata,不存在则创建并返回合适的machineID
	GetOrNew(ctx context.Context, name, addr string) (md Metadata, err error)
	// 获取所有的metadata
	List(ctx context.Context) (mds []Metadata, err error)
	// 更新时间戳
	Update(ctx context.Context, md Metadata) (err error)
}

type redisStorage struct {
	rds    *redis.Pool
	logger log.Logger
}

// gleafd/snowflakes/name/ip:port -> workerId  timestamp
func (storage *redisStorage) prefix() string {
	return "gleafd/snowflakes"
}

func (storage *redisStorage) key(name, addr string) string {
	return fmt.Sprintf("%s/%s/%s", storage.prefix(), name, addr)
}

func (storage *redisStorage) GetOrNew(ctx context.Context, name, addr string) (md Metadata, err error) {
	k := storage.key(name, addr)
	c := storage.rds.Get()
	defer c.Close()

	exists, err := redis.Bool(c.Do("EXISTS", k))
	if err != nil {
		return md, err
	}
	if !exists {
		// 不存在
		machineId, err := redis.Int(c.Do("INCR", "gleafd_machineid_gen"))
		if err != nil {
			return md, err
		}
		storage.logger.Warnw("Snowflake service creating", "name", name, "addr", addr, "machineId", machineId)
		_, err = c.Do("HMSET", k, "machineid", machineId, "timestamp", 0)
		if err != nil {
			return md, err
		}
		md.Name = name
		md.Addr = addr
		md.MachineID = machineId
		md.Timestamp = 0
		return md, nil
	}

	vals, err := redis.Int64s(c.Do("HMGET", k, "machineid", "timestmap"))
	if err != nil {
		return md, err
	}
	md.Name = name
	md.Addr = addr
	md.MachineID = int(vals[0])
	md.Timestamp = vals[1]
	return md, nil
}

func (storage *redisStorage) List(ctx context.Context) (mds []Metadata, err error) {
	c := storage.rds.Get()
	defer c.Close()

	keys, err := redis.Strings(c.Do("KEYS", storage.prefix()+"*"))
	if err != nil {
		if err == redis.ErrNil {
			return mds, nil
		}
		return nil, err
	}
	for _, k := range keys {
		vals, err := redis.Int64Map(c.Do("HMGET", k, "machineid", "timestamp"))
		if err != nil {
			if err == redis.ErrNil {
				continue
			}
			return nil, err
		}
		parts := strings.Split(k, "/")
		if len(parts) != 4 {
			continue
		}
		mid, _ := vals["machineid"]
		ts, _ := vals["timestamp"]
		mds = append(mds, Metadata{
			Name:      parts[2],
			Addr:      parts[3],
			MachineID: int(mid),
			Timestamp: ts})
	}
	return mds, nil
}

func (storage *redisStorage) Update(ctx context.Context, md Metadata) (err error) {
	c := storage.rds.Get()
	defer c.Close()
	_, err = c.Do("HMSET",
		storage.key(md.Name, md.Addr), "machineid", md.MachineID, "timestamp", md.Timestamp)
	return err
}

func NewRedisStorage(p *redis.Pool, logger log.Logger) Storage {
	return &redisStorage{rds: p, logger: logger}
}
