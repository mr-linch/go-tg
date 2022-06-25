package tgb

import (
	"context"
	"testing"

	tg "github.com/mr-linch/go-tg"
	"github.com/stretchr/testify/assert"
)

func TestHandlerFunc_Handle(t *testing.T) {
	called := false

	handler := HandlerFunc(func(ctx context.Context, update *tg.Update) error {
		called = true
		return nil
	}).Handle

	assert.NoError(t, handler(context.Background(), &tg.Update{}))
	assert.True(t, called)
}
