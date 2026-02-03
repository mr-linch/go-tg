package tg

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestError_Error(t *testing.T) {
	t.Run("WithoutParameters", func(t *testing.T) {
		err := &Error{
			Code:       123,
			Message:    "test",
			Parameters: nil,
		}

		assert.EqualError(t, err, "123: test")
	})

	t.Run("WithParameters", func(t *testing.T) {
		err := &Error{
			Code:    123,
			Message: "test",
			Parameters: &ResponseParameters{
				MigrateToChatID: 12345,
			},
		}

		assert.EqualError(t, err, "123: test ({MigrateToChatID:12345 RetryAfter:0})")
	})
}

func TestError_Contains(t *testing.T) {
	err := &Error{
		Code:    123,
		Message: "test",
	}

	assert.False(t, err.Contains("Test"))
}
