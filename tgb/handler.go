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

// MessageHandler it's typed handler for Message.
// Impliment Handler interface.
type MessageHandler func(context.Context, *MessageUpdate) error

func (handler MessageHandler) Handle(ctx context.Context, update *Update) error {
	if msg := firstNotNil(
		update.Message,
		update.EditedMessage,
		update.ChannelPost,
		update.EditedChannelPost,
	); msg != nil {
		return handler(ctx, &MessageUpdate{
			Message: msg,
			Update:  update,
			Client:  update.Client,
		})
	}

	return nil
}

type InlineQueryHandler func(context.Context, *InlineQueryUpdate) error

func (handler InlineQueryHandler) Handle(ctx context.Context, update *Update) error {
	return handler(ctx, &InlineQueryUpdate{
		InlineQuery: update.InlineQuery,
		Update:      update,
		Client:      update.Client,
	})
}

type ChosenInlineResultHandler func(context.Context, *ChosenInlineResultUpdate) error

func (handler ChosenInlineResultHandler) Handle(ctx context.Context, update *Update) error {
	return handler(ctx, &ChosenInlineResultUpdate{
		ChosenInlineResult: update.ChosenInlineResult,
		Update:             update,
		Client:             update.Client,
	})
}

type CallbackQueryHandler func(context.Context, *CallbackQueryUpdate) error

func (handler CallbackQueryHandler) Handle(ctx context.Context, update *Update) error {
	return handler(ctx, &CallbackQueryUpdate{
		CallbackQuery: update.CallbackQuery,
		Update:        update,
		Client:        update.Client,
	})
}

type ShippingQueryHandler func(context.Context, *ShippingQueryUpdate) error

func (handler ShippingQueryHandler) Handle(ctx context.Context, update *Update) error {
	return handler(ctx, &ShippingQueryUpdate{
		ShippingQuery: update.ShippingQuery,
		Update:        update,
		Client:        update.Client,
	})
}

type PreCheckoutQueryHandler func(context.Context, *PreCheckoutQueryUpdate) error

func (handler PreCheckoutQueryHandler) Handle(ctx context.Context, update *Update) error {
	return handler(ctx, &PreCheckoutQueryUpdate{
		PreCheckoutQuery: update.PreCheckoutQuery,
		Update:           update,
		Client:           update.Client,
	})
}

type PollHandler func(context.Context, *PollUpdate) error

func (handler PollHandler) Handle(ctx context.Context, update *Update) error {
	return handler(ctx, &PollUpdate{
		Poll:   update.Poll,
		Update: update,
		Client: update.Client,
	})
}

type PollAnswerHandler func(context.Context, *PollAnswerUpdate) error

func (handler PollAnswerHandler) Handle(ctx context.Context, update *Update) error {
	return handler(ctx, &PollAnswerUpdate{
		PollAnswer: update.PollAnswer,
		Update:     update,
		Client:     update.Client,
	})
}

type ChatMemberUpdatedHandler func(context.Context, *ChatMemberUpdatedUpdate) error

func (handler ChatMemberUpdatedHandler) Handle(ctx context.Context, update *Update) error {
	if updated := firstNotNil(
		update.MyChatMember,
		update.ChatMember,
	); updated != nil {
		return handler(ctx, &ChatMemberUpdatedUpdate{
			ChatMemberUpdated: updated,
			Update:            update,
			Client:            update.Client,
		})
	}

	return nil
}

type ChatJoinRequestHandler func(context.Context, *ChatJoinRequestUpdate) error

func (handler ChatJoinRequestHandler) Handle(ctx context.Context, update *Update) error {
	return handler(ctx, &ChatJoinRequestUpdate{
		ChatJoinRequest: update.ChatJoinRequest,
		Update:          update,
		Client:          update.Client,
	})
}

func firstNotNil[T any](fields ...*T) *T {
	for _, field := range fields {
		if field != nil {
			return field
		}
	}

	return nil
}
