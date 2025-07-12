package gorlock

import (
	"context"
	"math"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisConfig struct {
	Address        string // address:port
	Database       int    // database number to connect to
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

// lock tries to acquire a lock on the given connection for the key via SETNX as discussed at http://redis.io/commands/setnx, returning true if successful.
// Timeout is given in milliseconds.
func (rl *redlock) lock(key string, timeout time.Duration) (acquired bool, err error) {
	conn := rl.c.Conn()
	defer conn.Close()

	ctx := context.Background()
	result := conn.SetNX(ctx, key, time.Now().Add(timeout).UnixNano(), timeout)

	if result.Val() {
		acquired = true
	} else {
		var expires int64
		expires, err = conn.Get(ctx, key).Int64()
		expireTime := time.Unix(0, expires)
		if err != nil {
			acquired = false
		} else if expireTime.Before(time.Now()) { // try to reset the time
			newExpires := time.Now().Add(timeout).UnixNano()
			var newTime int64
			newTime, err = conn.GetSet(ctx, key, newExpires).Int64()
			if err != nil {
				return
			} else if newTime == expires { // we set it
				acquired = true
			}
		}
	}

	if acquired {
		expire := int(math.Ceil(timeout.Seconds()))
		_, err = conn.Expire(ctx, key, time.Duration(expire)*time.Second).Result()
	}
	return
}

// unlock deletes the given lock key, releasing the lock.
func (rl *redlock) unlock(key string) (err error) {
	conn := rl.c.Conn()
	defer conn.Close()

	_, err = conn.Del(context.Background(), key).Result()
	return err
}

// close Redis
func (rl *redlock) close() (err error) {
	if rl.c != nil {
		return rl.c.Close()
	}
	return
}
