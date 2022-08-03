package tg

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBindClient(t *testing.T) {
	call := NewGetMeCall()

	assert.Nil(t, call.client)

	BindClient(call, &Client{})

	assert.NotNil(t, call.client)
}
