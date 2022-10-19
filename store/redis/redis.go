package redis

import (
	"github.com/go-redis/redis"
	"github.com/hellobchain/sensitivewordfilter/store"
	"sync/atomic"
)

type redisConf struct {
	address string
}

type RedisStore struct {
	version uint64
	client  *redis.Client
}

func NewRedisStore(redisConf *redisConf) store.SensitivewordStore {
	redisOptions := &redis.Options{
		Addr: redisConf.address,
	}
	client := redis.NewClient(redisOptions)
	return &RedisStore{
		client: client,
	}

}

// Write
func (ms *RedisStore) Write(words ...string) error {
	if len(words) == 0 {
		return nil
	}
	for i, l := 0, len(words); i < l; i++ {
		err := ms.client.Set(words[i], nil, 0).Err()
		if err != nil {
			return err
		}
	}
	atomic.AddUint64(&ms.version, 1)
	return nil
}

// Read
func (ms *RedisStore) Read() <-chan string {
	panic("implement me")
}

// ReadAll ReadAll
func (ms *RedisStore) ReadAll() ([]string, error) {
	panic("implement me")
}

// Remove Remove
func (ms *RedisStore) Remove(words ...string) error {
	panic("implement me")
}

// Version Version
func (ms *RedisStore) Version() uint64 {
	return ms.version
}
