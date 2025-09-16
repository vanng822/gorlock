// Package gorlock is redis lock for doing certain task executed once at the time.
//
// import (
//
//	"fmf"
//
//	"github.com/vanng822/gorlock"
//
// )
//
//	gorlock.Run("somekey", func() error {
//		fmt.Println("Doing some job")
//		return nil
//	})
package gorlock

import (
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	defaultRedlock             *redlock
	defaultSettings            *Settings
	lockWaitingDefaultSettings *Settings
)

func init() {
	defaultRedlock = newRedLock(&RedisConfig{
		Address:        "localhost:6379",
		Database:       1,
		ConnectTimeout: 5 * time.Second,
	})
	defaultSettings = &Settings{
		KeyPrefix:     "gorlock",
		LockTimeout:   15 * time.Second,
		LockWaiting:   false,
		RetryTimeout:  5 * time.Second,
		RetryInterval: 150 * time.Millisecond,
	}
	lockWaitingDefaultSettings = &Settings{
		KeyPrefix:     "gorlock",
		LockTimeout:   15 * time.Second,
		LockWaiting:   true,
		RetryTimeout:  5 * time.Second,
		RetryInterval: 150 * time.Millisecond,
	}
}

type Settings struct {
	KeyPrefix     string
	LockTimeout   time.Duration
	RetryTimeout  time.Duration
	RetryInterval time.Duration
	LockWaiting   bool
}

type Gorlock interface {
	Lock(key string) (bool, error)
	Unlock(key string) error
	Close() error
	WithSettings(settings *Settings) Gorlock
	WithRedisClient(redisClient *redis.Client) Gorlock
	Run(key string, fn func() error) error
}

type gorlock struct {
	redlock   *redlock
	settings  *Settings
	isDefault bool
}

// Lock a lock and returns status, need to unlock it when done
func (g *gorlock) Lock(key string) (acquired bool, err error) {
	lockKey := g.prefixedKey(key)
	acquired, err = g.redlock.lock(lockKey, g.settings.LockTimeout)
	if err != nil {
		return false, err
	}
	if acquired {
		return true, nil
	}
	if g.settings.LockWaiting {
		timesup := time.After(g.settings.RetryTimeout)
		for {
			select {
			case <-timesup:
				return false, fmt.Errorf("time's up! Can not acquire lock key: %s", key)
			default:
				time.Sleep(g.settings.RetryInterval)
				acquired, err = g.redlock.lock(lockKey, g.settings.LockTimeout)
				if err != nil {
					return false, err
				}
				if acquired {
					return true, nil
				}
			}
		}
	}
	return false, fmt.Errorf("can not acquire lock key: %s", key)
}

func (g *gorlock) prefixedKey(key string) string {
	if g.settings.KeyPrefix == "" {
		return key
	}
	return fmt.Sprintf("%s:%s", g.settings.KeyPrefix, key)
}

func (g *gorlock) Unlock(key string) (err error) {
	lockKey := g.prefixedKey(key)
	return g.redlock.unlock(lockKey)
}

// Close the gorlock connection
// If the gorlock is created with NewDefault or NewDefaultWaiting, it will not close the connection
func (g *gorlock) Close() (err error) {
	// isDefault true means that we share the default redlock instance
	// don't close it
	if g.redlock != nil && !g.isDefault {
		return g.redlock.close()
	}
	return nil
}

// WithSettings provides a way to set custom settings for the gorlock instance.
func (g *gorlock) WithSettings(settings *Settings) Gorlock {
	g.settings = settings
	return g
}

// For managing the Redis client self
// go-redis seems to be the good choice for redis client in Go
// just expose this to outside world
func (g *gorlock) WithRedisClient(redisClient *redis.Client) Gorlock {
	g.redlock = &redlock{
		c: redisClient,
	}
	// should close the client if self managed the redis client
	g.isDefault = false

	return g
}

func (g *gorlock) Run(key string, fn func() error) error {
	acquired, err := g.Lock(key)
	if err != nil {
		return err
	}
	if !acquired {
		return fmt.Errorf("can not acquire lock key: %s", key)
	}
	defer g.Unlock(key)
	return fn()
}

func New(settings *Settings, redisConfig *RedisConfig) Gorlock {
	return &gorlock{
		redlock:   newRedLock(redisConfig),
		settings:  settings,
		isDefault: false,
	}
}

// New ..
func NewDefault() Gorlock {
	return newDefault()
}

func newDefault() *gorlock {
	return &gorlock{
		redlock:   defaultRedlock,
		settings:  defaultSettings,
		isDefault: true,
	}
}

func NewDefaultWaiting() Gorlock {
	return newDefaultWaiting()
}

func newDefaultWaiting() *gorlock {
	return &gorlock{
		redlock:   defaultRedlock,
		settings:  lockWaitingDefaultSettings,
		isDefault: true,
	}
}

// Run executes the job if a lock is acquired
func Run(key string, fn func() error) error {
	g := newDefault()
	return g.Run(key, fn)
}

// RunWaiting waits until acquiring a lock
// and execute the job
func RunWaiting(key string, fn func() error) error {
	g := newDefaultWaiting()
	return g.Run(key, fn)
}
