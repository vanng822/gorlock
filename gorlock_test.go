package gorlock

import (
	"fmt"
	"testing"
	"time"

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
	g := NewDefault()
	g.Acquire(key)
	go func() {
		time.Sleep(timeout)
		g.Unlock(key)
		done <- true
	}()
}

func TestRunWaitingOk(t *testing.T) {
	key := "runwating.ok"
	done := make(chan bool)
	defer func() {
		<-done
	}()
	testingDoBlock(key, lockWaitingDefaultSettings.RetryInterval*2, done)
	assert.Nil(t, RunWaiting(key, func() error {
		return nil
	}))
}

func TestRunWaitingError(t *testing.T) {
	key := "runwaiting.error"
	done := make(chan bool)
	defer func() {
		<-done
	}()
	testingDoBlock(key, lockWaitingDefaultSettings.RetryInterval*5, done)
	assert.EqualError(t, RunWaiting(key, func() error {
		return fmt.Errorf("run wating is not ok")
	}), "run wating is not ok")

}

func TestCanAcquire(t *testing.T) {
	key := "acquire.ok"
	waitingDefaultSettings := &Settings{
		LockTimeout:   15 * time.Second,
		LockWaiting:   true,
		RetryTimeout:  100 * time.Millisecond,
		RetryInterval: 20 * time.Millisecond,
	}

	redisConfig := &RedisConfig{
		Address:        "localhost:6379",
		Database:       1,
		KeyPrefix:      "gorlock",
		ConnectTimeout: 30 * time.Second,
	}

	g := New(waitingDefaultSettings, redisConfig)
	defer g.Close()
	lock, err := g.Acquire(key)
	assert.True(t, lock)
	assert.NoError(t, err)
}

func TestLockTimesUp(t *testing.T) {
	key := "runwating.error.timesup"
	done := make(chan bool)
	defer func() {
		<-done
	}()
	testingDoBlock(key, 200*time.Millisecond, done)

	waitingDefaultSettings := &Settings{
		LockTimeout:   15 * time.Second,
		LockWaiting:   true,
		RetryTimeout:  100 * time.Millisecond,
		RetryInterval: 20 * time.Millisecond,
	}

	redisConfig := &RedisConfig{
		Address:        "localhost:6379",
		Database:       1,
		KeyPrefix:      "gorlock",
		ConnectTimeout: 30 * time.Second,
	}

	g := New(waitingDefaultSettings, redisConfig)
	defer g.Close()
	acquired, err := g.Acquire(key)
	assert.False(t, acquired)
	assert.EqualError(t, err, "time's up! Can not acquire lock key: runwating.error.timesup")
}

func TestConnectionError(t *testing.T) {
	key := "runwaiting.error"
	done := make(chan bool)
	defer func() {
		<-done
	}()
	testingDoBlock(key, 100*time.Millisecond, done)
	redisConfig := &RedisConfig{
		Address:        "localhost:6390",
		Database:       1,
		KeyPrefix:      "gorlock",
		ConnectTimeout: 30 * time.Second,
	}
	g := New(defaultSettings, redisConfig)
	acquired, err := g.Acquire(key)
	assert.False(t, acquired)
	assert.Error(t, err)
	assert.Regexp(t, "6390: connect: connection refused", err.Error())
}

func TestCanNotAcquire(t *testing.T) {
	key := "acquire.error"
	done := make(chan bool)
	defer func() {
		<-done
	}()
	testingDoBlock(key, 100*time.Millisecond, done)
	g := NewDefault()
	defer g.Close()
	lock, err := g.Acquire(key)
	assert.False(t, lock)
	assert.EqualError(t, err, "can not acquire lock key: acquire.error")
}

func TestAcquireConnectionError(t *testing.T) {
	key := "runwaiting.error.connection"
	done := make(chan bool)
	defer func() {
		<-done
	}()
	testingDoBlock(key, 100*time.Millisecond, done)
	g := New(defaultSettings, &RedisConfig{
		Address:        "localhost:6390",
		Database:       1,
		KeyPrefix:      "gorlock",
		ConnectTimeout: 30 * time.Second,
	})
	acquired, err := g.Acquire(key)
	assert.False(t, acquired)
	assert.Error(t, err)
	assert.Regexp(t, "6390: connect: connection refused", err.Error())
}
