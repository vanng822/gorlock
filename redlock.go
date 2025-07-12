package gorlock

import (
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisConfig struct {
	Address        string // address:port
	Database       int    // database number to connect to
	KeyPrefix      string // optional prefix applied to keys
	ConnectTimeout time.Duration
	ReadTimeout    time.Duration
	WriteTimeout   time.Duration
}

type redlock struct {
	c    *redis.Client
	conf *RedisConfig
}

// newRedLock returns a redlock instance
func newRedLock(conf *RedisConfig) (rl *redlock) {
	c := redis.NewClient(&redis.Options{
		Addr:         conf.Address,
		DB:           conf.Database,
		DialTimeout:  conf.ConnectTimeout,
		ReadTimeout:  conf.ReadTimeout,
		WriteTimeout: conf.WriteTimeout,
	})

	return &redlock{
		conf: conf,
		c:    c,
	}
}

// close Redis
func (rl *redlock) close() (err error) {
	if rl.c != nil {
		return rl.c.Close()
	}
	return
}
