package tgb

import (
	"context"
	"errors"

	tg "github.com/mr-linch/go-tg"
)

type Handler interface {
	Handle(ctx context.Context, update *tg.Update) error
}

type HandlerFunc func(ctx context.Context, update *tg.Update) error

func (handler HandlerFunc) Handle(ctx context.Context, update *tg.Update) error {
	return handler(ctx, update)
}

type MessageHandler func(context.Context, *tg.Message) error

func (handler MessageHandler) Handle(ctx context.Context, update *tg.Update) error {
	if msg := firstNotNil(
		update.Message,
		update.EditedMessage,
		update.ChannelPost,
		update.EditedChannelPost,
	); msg != nil {
		return handler(ctx, msg)
	}

	return errors.New("no message in Update")
}

type InlineQueryHandler func(context.Context, *tg.InlineQuery) error

func (handler InlineQueryHandler) Handle(ctx context.Context, update *tg.Update) error {
	if update.InlineQuery != nil {
		return handler(ctx, update.InlineQuery)
	}

	return errors.New("no inline query in Update")
}

type ChosenInlineResultHandler func(context.Context, *tg.ChosenInlineResult) error

func (handler ChosenInlineResultHandler) Handle(ctx context.Context, update *tg.Update) error {
	if update.ChosenInlineResult != nil {
		return handler(ctx, update.ChosenInlineResult)
	}

	return errors.New("no chosen inline query in Update")
}

type CallbackQueryHandler func(context.Context, *tg.CallbackQuery) error

func (handler CallbackQueryHandler) Handle(ctx context.Context, update *tg.Update) error {
	if update.CallbackQuery != nil {
		return handler(ctx, update.CallbackQuery)
	}

	return errors.New("no callback query in Update")
}

type ShippingQueryHandler func(context.Context, *tg.ShippingQuery) error

func (handler ShippingQueryHandler) Handle(ctx context.Context, update *tg.Update) error {
	if update.ShippingQuery != nil {
		return handler(ctx, update.ShippingQuery)
	}

	return errors.New("no shipping query in Update")
}

type PreCheckoutQueryHandler func(context.Context, *tg.PreCheckoutQuery) error

func (handler PreCheckoutQueryHandler) Handle(ctx context.Context, update *tg.Update) error {
	if update.PreCheckoutQuery != nil {
		return handler(ctx, update.PreCheckoutQuery)
	}

	return errors.New("no precheckout query in Update")
}

type PollHandler func(context.Context, *tg.Poll) error

func (handler PollHandler) Handle(ctx context.Context, update *tg.Update) error {
	if update.Poll != nil {
		return handler(ctx, update.Poll)
	}

	return errors.New("no poll in Update")
}

type PollAnswerHandler func(context.Context, *tg.PollAnswer) error

func (handler PollAnswerHandler) Handle(ctx context.Context, update *tg.Update) error {
	if update.PollAnswer != nil {
		return handler(ctx, update.PollAnswer)
	}

	return errors.New("no poll answer in Update")
}

type ChatMemberUpdatedHandler func(context.Context, *tg.ChatMemberUpdated) error

func (handler ChatMemberUpdatedHandler) Handle(ctx context.Context, update *tg.Update) error {
	if updated := firstNotNil(
		update.MyChatMember,
		update.ChatMember,
	); updated != nil {
		return handler(ctx, updated)
	}

	return errors.New("no ChatMemberUpdated in Update")
}

type ChatJoinRequestHandler func(context.Context, *tg.ChatJoinRequest) error

func (handler ChatJoinRequestHandler) Handle(ctx context.Context, update *tg.Update) error {
	if update.ChatJoinRequest != nil {
		return handler(ctx, update.ChatJoinRequest)
	}

	return errors.New("no chat join request in Update")
}

func firstNotNil[T any](fields ...*T) *T {
	for _, field := range fields {
		if field != nil {
			return field
		}
	}

	return nil
}
