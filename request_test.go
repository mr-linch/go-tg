package tg

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewRequest(t *testing.T) {
	r := NewRequest("getMe")

	assert.Equal(t, "getMe", r.Method)
}

func TestRequest_String(t *testing.T) {
	r := NewRequest("getMe")

	r.String("foo", "bar")

	assert.Equal(t, "bar", r.args["foo"])
}
