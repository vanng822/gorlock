package gorlock

import (
	"fmt"
	"testing"
	"time"

	"github.com/gpitfield/redlock"
	"github.com/stretchr/testify/assert"
)

func TestRunOk(t *testing.T) {
	assert.Nil(t, Run("run.ok", func() error {
		return nil
	}))
}

func TestRunError(t *testing.T) {
	assert.EqualError(t, Run("run.error", func() error {
		return fmt.Errorf("run is not ok")
	}), "run is not ok")
}

func doBlock(key string, timeout time.Duration, done chan bool) {
	// should use Lock directly
	l, _ := Acquire(key, DefaultSettings)
	go func() {
		time.Sleep(timeout)
		l.Unlock(key)
		done <- true
	}()
}

func TestRunWaitingOk(t *testing.T) {
	key := "runwating.ok"
	done := make(chan bool)
	doBlock(key, LockWaitingDefaultSettings.RetryInterval*2, done)
	assert.Nil(t, RunWaiting(key, func() error {
		return nil
	}))
	<-done
}

func TestRunWaitingError(t *testing.T) {
	key := "runwating.error"
	done := make(chan bool)
	doBlock(key, LockWaitingDefaultSettings.RetryInterval*2, done)
	assert.EqualError(t, RunWaiting(key, func() error {
		return fmt.Errorf("run wating is not ok")
	}), "run wating is not ok")
	<-done
}

func TestLockTimesUp(t *testing.T) {
	tmp := LockWaitingDefaultSettings.RetryTimeout
	defer func() {
		LockWaitingDefaultSettings.RetryTimeout = tmp
	}()
	LockWaitingDefaultSettings.RetryTimeout = 300 * time.Millisecond
	key := "runwating.error"
	done := make(chan bool)
	doBlock(key, 500*time.Millisecond, done)
	assert.EqualError(t, RunWaiting(key, func() error {
		return nil
	}), "Time's up! Can not acquire lock key: runwating.error")
	<-done
}

func TestConnectionError(t *testing.T) {
	tmp := *_redisConfig
	defer func() {
		*_redisConfig = tmp
	}()
	key := "runwating.error"
	done := make(chan bool)
	doBlock(key, 100*time.Millisecond, done)
	_redisConfig = &redlock.RedisConfig{
		Address:        "localhost:6390",
		Database:       1,
		KeyPrefix:      "gorlock",
		ConnectTimeout: 30 * time.Second,
	}
	lock, err := Acquire(key, DefaultSettings)
	assert.Nil(t, lock)
	assert.Error(t, err)
	assert.Regexp(t, "6390: getsockopt: connection refused", err.Error())
	<-done
}

func TestCanNotAcquire(t *testing.T) {
	key := "runwating.error"
	done := make(chan bool)
	doBlock(key, 100*time.Millisecond, done)
	lock, err := Acquire(key, DefaultSettings)
	assert.Nil(t, lock)
	assert.EqualError(t, err, "Can not acquire lock key: runwating.error")
	<-done
}
