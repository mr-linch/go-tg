package tgb

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestChain_Append(t *testing.T) {
	old := chain{}
	new := old.Append(MiddlewareFunc(func(h Handler) Handler { return h }))

	assert.Len(t, old, 0)
	assert.Len(t, new, 1)
}

func TestChain_Then(t *testing.T) {
	calls := []int{}

	chain := chain{
		MiddlewareFunc(func(h Handler) Handler {
			calls = append(calls, 1)
			return h
		}),
		MiddlewareFunc(func(h Handler) Handler {
			calls = append(calls, 2)
			return h
		}),
	}

	handler := chain.Then(HandlerFunc(func(ctx context.Context, update *Update) error {
		calls = append(calls, 3)
		return nil
	}))

	err := handler.Handle(context.Background(), &Update{})
	assert.NoError(t, err)

	assert.Equal(t, []int{2, 1, 3}, calls)

}
