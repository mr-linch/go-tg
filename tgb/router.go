package tgb

import (
	"context"
	"errors"
	"fmt"

	"github.com/mr-linch/go-tg"
)

// ErrorHandler define interface for error handling in Bot.
// See Bot.Error for more information.
type ErrorHandler func(ctx context.Context, update *Update, err error) error

// Router is a router for incoming Updates.
// tg.Update should be wrapped into tgb.Update with binded Client and Update.
type Router struct {
	chain chain

	typedHandlers  map[tg.UpdateType][]Handler
	updateHandlers []Handler

	defaultHandler Handler
	errorHandler   ErrorHandler
}

// NewRouter creates new Bot.
func NewRouter() *Router {
	return &Router{
		chain:         chain{},
		typedHandlers: map[tg.UpdateType][]Handler{},
		defaultHandler: HandlerFunc(func(ctx context.Context, update *Update) error {
			return nil
		}),
	}
}

func compactFilters(filters ...Filter) Filter {
	if len(filters) == 1 {
		return filters[0]
	} else if len(filters) > 1 {
		return All(filters...)
	}
	return nil
}

// ErrFilterNoAllow is returned when filter doesn't allow to handle Update.
var ErrFilterNoAllow = fmt.Errorf("filter no allow")

func filterMiddleware(filter Filter) Middleware {
	return MiddlewareFunc(func(next Handler) Handler {
		return HandlerFunc(func(ctx context.Context, update *Update) error {
			if filter == nil {
				return next.Handle(ctx, update)
			}

			allow, err := filter.Allow(ctx, update)
			if err != nil {
				return fmt.Errorf("filter error: %w", err)
			}

			if allow {
				return next.Handle(ctx, update)
			}

			return ErrFilterNoAllow
		})
	})
}

// Use add middleware to chain handlers.
// Should be called before any other register handler.
func (bot *Router) Use(mws ...Middleware) *Router {
	bot.chain = bot.chain.Append(mws...)
	return bot
}

func (bot *Router) register(typ tg.UpdateType, handler Handler, filters ...Filter) *Router {
	filter := compactFilters(filters...)

	bot.typedHandlers[typ] = append(bot.typedHandlers[typ],
		bot.chain.Append(filterMiddleware(filter)).Then(handler),
	)

	return bot
}

// Error registers a handler for errors.
// If any error occurs in the chain, it will be passed to that handler.
// By default, errors are returned back by handler method.
// You can customize this behavior by passing a custom error handler.
func (bot *Router) Error(handler ErrorHandler) *Router {
	bot.errorHandler = handler
	return bot
}

// Update registers a generic Update handler.
// It will be called as typed handlers only in filters match the update.
// First check Update handler, then typed.
func (bot *Router) Update(handler HandlerFunc, filters ...Filter) *Router {
	fitler := compactFilters(filters...)

	bot.updateHandlers = append(bot.updateHandlers,
		bot.chain.Append(filterMiddleware(fitler)).Then(handler),
	)

	return bot
}

func (bot *Router) getDefaultHandler() Handler {
	return bot.chain.Then(bot.defaultHandler)
}

// Handle handles an Update.
func (bot *Router) Handle(ctx context.Context, update *Update) error {
	group := append([]Handler{}, bot.updateHandlers...)

	typed, ok := bot.typedHandlers[update.Type()]
	if ok {
		group = append(group, typed...)
	}

	// If no handlers found, use default handler.
	group = append(group, bot.getDefaultHandler())

	for _, handler := range group {
		err := handler.Handle(ctx, update)
		if errors.Is(err, ErrFilterNoAllow) {
			continue
		} else if err != nil && bot.errorHandler != nil {
			return bot.errorHandler(ctx, update, err)
		}

		return err
	}

	return nil
}
