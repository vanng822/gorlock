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

type Redlock struct {
	c    *redis.Client
	conf *RedisConfig
}

// NewRedLock returns a redlock instance
func NewRedLock(conf *RedisConfig) (rl *Redlock) {
	c := redis.NewClient(&redis.Options{
		Addr:         conf.Address,
		DB:           conf.Database,
		DialTimeout:  conf.ConnectTimeout,
		ReadTimeout:  conf.ReadTimeout,
		WriteTimeout: conf.WriteTimeout,
	})

	return &Redlock{
		conf: conf,
		c:    c,
	}
}

// Close Redis
func (rl *Redlock) Close() (err error) {
	if rl.c != nil {
		return rl.c.Close()
	}
	return
}
