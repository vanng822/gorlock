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
		KeyPrefix:      "gorlock",
		ConnectTimeout: 30 * time.Second,
	})
	defaultSettings = &Settings{
		LockTimeout:   15 * time.Second,
		LockWaiting:   false,
		RetryTimeout:  15 * time.Second,
		RetryInterval: 150 * time.Millisecond,
	}
	lockWaitingDefaultSettings = &Settings{
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

type Gorlock interface {
	Acquire(key string) (bool, error)
	Unlock(key string) error
	Close() error
}

type gorlock struct {
	redlock   *redlock
	settings  *Settings
	isDefault bool
}

// Acquire a lock and returns status, need to unlock it when done
func (g *gorlock) Acquire(key string) (acquired bool, err error) {
	acquired, err = g.redlock.lock(key, g.settings.LockTimeout)
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
				acquired, err = g.redlock.lock(key, g.settings.LockTimeout)
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

func (g *gorlock) Unlock(key string) (err error) {
	return g.redlock.unlock(key)
}

// Close the gorlock connection
// If the gorlock is created with NewDefault or NewDefaultWaiting, it will not close the connection
func (g *gorlock) Close() (err error) {
	if g.redlock != nil && !g.isDefault {
		return g.redlock.close()
	}
	return nil
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
	return &gorlock{
		redlock:   defaultRedlock,
		settings:  defaultSettings,
		isDefault: true,
	}
}

func NewDefaultWaiting() Gorlock {
	return &gorlock{
		redlock:   defaultRedlock,
		settings:  lockWaitingDefaultSettings,
		isDefault: true,
	}
}

// Run executes the job if a lock is acquired
func Run(key string, fn func() error) error {
	g := NewDefault()
	acquired, err := g.Acquire(key)
	if err != nil {
		return err
	}
	if !acquired {
		return fmt.Errorf("can not acquire lock key: %s", key)
	}
	defer g.Unlock(key)
	return fn()
}

// RunWaiting waits until acquiring a lock
// and execute the job
func RunWaiting(key string, fn func() error) error {
	g := NewDefaultWaiting()
	acquired, err := g.Acquire(key)
	if err != nil {
		return err
	}
	if !acquired {
		return fmt.Errorf("can not acquire lock key: %s", key)
	}
	defer g.Unlock(key)
	return fn()
}
