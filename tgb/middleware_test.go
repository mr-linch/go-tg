package tgb

import (
	"context"
	"testing"

	tg "github.com/mr-linch/go-tg"
	"github.com/stretchr/testify/assert"
)

func TestChain_Append(t *testing.T) {
	old := chain{}
	new := old.Append(func(h Handler) Handler { return h })

	assert.Len(t, old, 0)
	assert.Len(t, new, 1)
}

func TestChain_Then(t *testing.T) {
	calls := []int{}

	chain := chain{
		func(h Handler) Handler {
			calls = append(calls, 1)
			return h
		},
		func(h Handler) Handler {
			calls = append(calls, 2)
			return h
		},
	}

	handler := chain.Then(HandlerFunc(func(ctx context.Context, update *tg.Update) error {
		calls = append(calls, 3)
		return nil
	}))

	err := handler.Handle(context.Background(), &tg.Update{})
	assert.NoError(t, err)

	assert.Equal(t, []int{2, 1, 3}, calls)

}
