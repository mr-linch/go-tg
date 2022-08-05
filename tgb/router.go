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

	errorHandler ErrorHandler
}

// NewRouter creates new Bot.
func NewRouter() *Router {
	return &Router{
		chain:         chain{},
		typedHandlers: map[tg.UpdateType][]Handler{},
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

var errFilterNoAllow = fmt.Errorf("no filter match")

func filterMiddleware(filter Filter) Middleware {
	return MiddlewareFunc(func(next Handler) Handler {
		return HandlerFunc(func(ctx context.Context, update *Update) error {
			if filter == nil {
				return next.Handle(ctx, update)
			}

			allow, err := filter.Allow(ctx, update)
			if err != nil {
				return err
			}

			if allow {
				return next.Handle(ctx, update)
			}

			return errFilterNoAllow
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

// Message register handlers for Update with not empty Message field.
func (bot *Router) Message(handler MessageHandler, filters ...Filter) *Router {
	return bot.register(tg.UpdateTypeMessage, handler, filters...)
}

// EditedMessage register handlers for Update with not empty EditedMessage field.
func (bot *Router) EditedMessage(handler MessageHandler, filters ...Filter) *Router {
	return bot.register(tg.UpdateTypeEditedMessage, handler, filters...)
}

// ChannelPost register handlers for Update with not empty ChannelPost field.
func (bot *Router) ChannelPost(handler MessageHandler, filters ...Filter) *Router {
	return bot.register(tg.UpdateTypeChannelPost, handler, filters...)
}

// EditedChannelPost register handlers for Update with not empty EditedChannelPost field.
func (bot *Router) EditedChannelPost(handler MessageHandler, filters ...Filter) *Router {
	return bot.register(tg.UpdateTypeEditedChannelPost, handler, filters...)
}

// InlineQuery register handlers for Update with not empty InlineQuery field.
func (bot *Router) InlineQuery(handler InlineQueryHandler, filters ...Filter) *Router {
	return bot.register(tg.UpdateTypeInlineQuery, handler, filters...)
}

// ChosenInlineResult register handlers for Update with not empty ChosenInlineResult field.
func (bot *Router) ChosenInlineResult(handler ChosenInlineResultHandler, filters ...Filter) *Router {
	return bot.register(tg.UpdateTypeChosenInlineResult, handler, filters...)
}

// CallbackQuery register handlers for Update with not empty CallbackQuery field.
func (bot *Router) CallbackQuery(handler CallbackQueryHandler, filters ...Filter) *Router {
	return bot.register(tg.UpdateTypeCallbackQuery, handler, filters...)
}

// ShippingQuery register handlers for Update with not empty ShippingQuery field.
func (bot *Router) ShippingQuery(handler ShippingQueryHandler, filters ...Filter) *Router {
	return bot.register(tg.UpdateTypeShippingQuery, handler, filters...)
}

// PreCheckoutQuery register handlers for Update with not empty PreCheckoutQuery field.
func (bot *Router) PreCheckoutQuery(handler PreCheckoutQueryHandler, filters ...Filter) *Router {
	return bot.register(tg.UpdateTypePreCheckoutQuery, handler, filters...)
}

// Poll register handlers for Update with not empty Poll field.
func (bot *Router) Poll(handler PollHandler, filters ...Filter) *Router {
	return bot.register(tg.UpdateTypePoll, handler, filters...)
}

// PollAnswer register handlers for Update with not empty PollAnswer field.
func (bot *Router) PollAnswer(handler PollAnswerHandler, filters ...Filter) *Router {
	return bot.register(tg.UpdateTypePollAnswer, handler, filters...)
}

// MyChatMember register handlers for Update with not empty MyChatMember field.
func (bot *Router) MyChatMember(handler ChatMemberUpdatedHandler, filters ...Filter) *Router {
	return bot.register(tg.UpdateTypeMyChatMember, handler, filters...)
}

// ChatMember register handlers for Update with not empty ChatMember field.
func (bot *Router) ChatMember(handler ChatMemberUpdatedHandler, filters ...Filter) *Router {
	return bot.register(tg.UpdateTypeChatMember, handler, filters...)
}

// ChatJoinRequest register handlers for Update with not empty ChatJoinRequest field.
func (bot *Router) ChatJoinRequest(handler ChatJoinRequestHandler, filters ...Filter) *Router {
	return bot.register(tg.UpdateTypeChatJoinRequest, handler, filters...)
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

// Handle handles an Update.
func (bot *Router) Handle(ctx context.Context, update *Update) error {
	group := append([]Handler{}, bot.updateHandlers...)

	typed, ok := bot.typedHandlers[update.Type()]
	if !ok {
		return nil
	}

	group = append(group, typed...)

	for _, handler := range group {
		err := handler.Handle(ctx, update)
		if errors.Is(err, errFilterNoAllow) {
			continue
		}
		return err
	}

	return nil
}
