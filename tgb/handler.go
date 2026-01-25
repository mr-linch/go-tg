package tgb

import (
	"context"
)

// Handler define generic Update handler.
type Handler interface {
	Handle(ctx context.Context, update *Update) error
}

// HandlerFunc define functional handler.
type HandlerFunc func(ctx context.Context, update *Update) error

// Handle implements Handler interface.
func (handler HandlerFunc) Handle(ctx context.Context, update *Update) error {
	return handler(ctx, update)
}
