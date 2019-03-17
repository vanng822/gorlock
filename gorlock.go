// Package gorlock is redis lock for doing certain task executed once at the time.
//
// import (
// 	"fmf"
//
// 	"github.com/vanng822/gorlock"
// )
// gorlock.Run("somekey", func() error {
// 	fmt.Println("Doing some job")
// 	return nil
// })
package gorlock

import (
	"fmt"
	"time"

	"github.com/gpitfield/redlock"
)

var (
	_redisConfig               *redlock.RedisConfig
	DefaultSettings            *Settings
	LockWaitingDefaultSettings *Settings
)

func init() {
	_redisConfig = &redlock.RedisConfig{
		Address:        "localhost:6379",
		Database:       1,
		KeyPrefix:      "gorlock",
		ConnectTimeout: 30 * time.Second,
	}
	DefaultSettings = &Settings{
		LockTimeout:   15 * time.Second,
		LockWaiting:   false,
		RetryTimeout:  15 * time.Second,
		RetryInterval: 150 * time.Millisecond,
	}
	LockWaitingDefaultSettings = &Settings{
		LockTimeout:   15 * time.Second,
		LockWaiting:   true,
		RetryTimeout:  15 * time.Second,
		RetryInterval: 150 * time.Millisecond,
	}
}

type Settings struct {
	LockTimeout   time.Duration
	RetryTimeout  time.Duration
	RetryInterval time.Duration
	LockWaiting   bool
}

func SetRedisConfig(config *redlock.RedisConfig) {
	_redisConfig = config
}

func RedisConfig() *redlock.RedisConfig {
	return _redisConfig
}

// New ..
func New() *redlock.Redlock {
	return redlock.New(RedisConfig())
}

// Acquire a lock and returns it, need to unlock it when done
func Acquire(key string, settings *Settings) (*redlock.Redlock, error) {
	var (
		acquired bool
		err      error
		lock     *redlock.Redlock
	)
	lock = New()
	acquired, err = lock.Lock(key, settings.LockTimeout)
	if err != nil {
		return nil, err
	}
	if acquired {
		return lock, nil
	}
	if settings.LockWaiting {
		timesup := time.After(settings.RetryTimeout)
		for {
			select {
			case <-timesup:
				return nil, fmt.Errorf("Time's up! Can not acquire lock key: %s", key)
			default:
				time.Sleep(settings.RetryInterval)
				acquired, err = lock.Lock(key, settings.LockTimeout)
				if err != nil {
					return nil, err
				}
				if acquired {
					return lock, nil
				}
			}
		}
	}
	return nil, fmt.Errorf("Can not acquire lock key: %s", key)
}

// Run executes the job if a lock is acquired
func Run(key string, fn func() error) error {
	lock, err := Acquire(key, DefaultSettings)
	if err != nil {
		return err
	}
	defer lock.Unlock(key)
	return fn()
}

// RunWaiting waits until acquiring a lock
// and execute the job
func RunWaiting(key string, fn func() error) error {
	lock, err := Acquire(key, LockWaitingDefaultSettings)
	if err != nil {
		return err
	}
	defer lock.Unlock(key)
	return fn()
}
