package tgb

import (
	"context"

	tg "github.com/mr-linch/go-tg"
)

type Handler interface {
	Handle(ctx context.Context, update *tg.Update) error
}

type HanlderFunc func(ctx context.Context, update *tg.Update) error

func (handler HanlderFunc) Handle(ctx context.Context, update *tg.Update) error {
	return handler(ctx, update)
}
