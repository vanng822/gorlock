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

func doBlock(key string, timeout time.Duration) {
	l, _ := Acquire(key, DefaultSettings)
	go func() {
		time.Sleep(timeout)
		l.Unlock(key)
	}()
}

func TestRunWaitingOk(t *testing.T) {
	key := "runwating.ok"
	doBlock(key, LockWaitingDefaultSettings.RetryInterval*2)
	assert.Nil(t, RunWaiting(key, func() error {
		return nil
	}))
}

func TestRunWaitingError(t *testing.T) {
	key := "runwating.error"
	doBlock(key, LockWaitingDefaultSettings.RetryInterval*2)
	assert.EqualError(t, RunWaiting(key, func() error {
		return fmt.Errorf("run wating is not ok")
	}), "run wating is not ok")
}

func TestLockTimesUp(t *testing.T) {
	LockWaitingDefaultSettings.RetryTimeout = 300 * time.Millisecond
	key := "runwating.error"
	doBlock(key, 500*time.Millisecond)
	assert.EqualError(t, RunWaiting(key, func() error {
		return nil
	}), "Time's up! Can not acquire lock key: runwating.error")
}
