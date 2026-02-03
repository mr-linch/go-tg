package tgb

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandlerFunc_Handle(t *testing.T) {
	called := false

	handler := HandlerFunc(func(ctx context.Context, update *Update) error {
		called = true
		return nil
	}).Handle

	require.NoError(t, handler(context.Background(), &Update{}))
	assert.True(t, called)
}
