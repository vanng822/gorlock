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

func testingDoBlock(key string, timeout time.Duration, done chan bool) {
	// should use Lock directly
	lock := New()
	lock.Lock(key, 2*timeout)
	go func() {
		time.Sleep(timeout)
		lock.Unlock(key)
		done <- true
	}()
}

func TestRunWaitingOk(t *testing.T) {
	key := "runwating.ok"
	done := make(chan bool)
	defer func() {
		<-done
	}()
	testingDoBlock(key, LockWaitingDefaultSettings.RetryInterval*2, done)
	assert.Nil(t, RunWaiting(key, func() error {
		return nil
	}))
}

func TestRunWaitingError(t *testing.T) {
	key := "runwating.error"
	done := make(chan bool)
	defer func() {
		<-done
	}()
	testingDoBlock(key, LockWaitingDefaultSettings.RetryInterval*2, done)
	assert.EqualError(t, RunWaiting(key, func() error {
		return fmt.Errorf("run wating is not ok")
	}), "run wating is not ok")

}

func TestLockTimesUp(t *testing.T) {
	tmp := LockWaitingDefaultSettings.RetryTimeout
	defer func() {
		LockWaitingDefaultSettings.RetryTimeout = tmp
	}()
	LockWaitingDefaultSettings.RetryTimeout = 300 * time.Millisecond
	key := "runwating.error"
	done := make(chan bool)
	defer func() {
		<-done
	}()
	testingDoBlock(key, 500*time.Millisecond, done)
	assert.EqualError(t, RunWaiting(key, func() error {
		return nil
	}), "Time's up! Can not acquire lock key: runwating.error")
}

func TestConnectionError(t *testing.T) {
	tmp := *_redisConfig
	defer func() {
		*_redisConfig = tmp
	}()
	key := "runwating.error"
	done := make(chan bool)
	defer func() {
		<-done
	}()
	testingDoBlock(key, 100*time.Millisecond, done)
	_redisConfig = &redlock.RedisConfig{
		Address:        "localhost:6390",
		Database:       1,
		KeyPrefix:      "gorlock",
		ConnectTimeout: 30 * time.Second,
	}
	lock, err := Acquire(key, DefaultSettings)
	assert.Nil(t, lock)
	assert.Error(t, err)
	assert.Regexp(t, "6390: connect: connection refused", err.Error())
}

func TestCanNotAcquire(t *testing.T) {
	key := "runwating.error"
	done := make(chan bool)
	defer func() {
		<-done
	}()
	testingDoBlock(key, 100*time.Millisecond, done)
	lock, err := Acquire(key, DefaultSettings)
	assert.Nil(t, lock)
	assert.EqualError(t, err, "Can not acquire lock key: runwating.error")
}

func TestRunConnectionError(t *testing.T) {
	tmp := *_redisConfig
	defer func() {
		*_redisConfig = tmp
	}()
	key := "runwating.error"
	done := make(chan bool)
	defer func() {
		<-done
	}()
	testingDoBlock(key, 100*time.Millisecond, done)
	SetRedisConfig(&redlock.RedisConfig{
		Address:        "localhost:6390",
		Database:       1,
		KeyPrefix:      "gorlock",
		ConnectTimeout: 30 * time.Second,
	})
	err := Run(key, func() error {
		return nil
	})
	assert.Error(t, err)
	assert.Regexp(t, "6390: connect: connection refused", err.Error())
}
