package gorlock

import (
	"fmt"
	"testing"

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
