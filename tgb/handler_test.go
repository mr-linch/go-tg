package tgb

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHandlerFunc_Handle(t *testing.T) {
	called := false

	handler := HandlerFunc(func(ctx context.Context, update *Update) error {
		called = true
		return nil
	}).Handle

	assert.NoError(t, handler(context.Background(), &Update{}))
	assert.True(t, called)
}
