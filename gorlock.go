package gorlock

import (
	"fmt"
	"time"

	"github.com/gpitfield/redlock"
)

var (
	_redisConfig    *redlock.RedisConfig
	DefaultSettings *Settings
)

func init() {
	_redisConfig = &redlock.RedisConfig{
		Address:        "localhost:6379",
		Database:       1,
		KeyPrefix:      "gorlock",
		ConnectTimeout: 30 * time.Second,
	}
	DefaultSettings = &Settings{
		LockTimeout: 15 * time.Second,
	}
}

type Settings struct {
	LockTimeout time.Duration
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
func Acquire(key string) (*redlock.Redlock, error) {
	lock := New()
	acquired, err := lock.Lock(key, DefaultSettings.LockTimeout)
	if err != nil {
		return nil, err
	}
	if !acquired {
		return nil, fmt.Errorf("Can not acquire lock key: %s", key)
	}
	return lock, nil
}

// Run ..
func Run(key string, fn func() error) error {
	lock, err := Acquire(key)
	if err != nil {
		return err
	}
	defer lock.Unlock(key)
	return fn()
}
