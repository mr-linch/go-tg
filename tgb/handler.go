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
		update.BusinessMessage,
		update.EditedBusinessMessage,
	); msg != nil {
		return handler(ctx, &MessageUpdate{
			Message: msg,
			Update:  update,
			Client:  update.Client,
		})
	}

	return nil
}

// InlineQueryHandler it's typed handler for InlineQuery.
// Impliment Handler interface.
type InlineQueryHandler func(context.Context, *InlineQueryUpdate) error

func (handler InlineQueryHandler) Handle(ctx context.Context, update *Update) error {
	return handler(ctx, &InlineQueryUpdate{
		InlineQuery: update.InlineQuery,
		Update:      update,
		Client:      update.Client,
	})
}

// ChosenInlineResultHandler it's typed handler for ChosenInlineResult.
// Impliment Handler interface.
type ChosenInlineResultHandler func(context.Context, *ChosenInlineResultUpdate) error

func (handler ChosenInlineResultHandler) Handle(ctx context.Context, update *Update) error {
	return handler(ctx, &ChosenInlineResultUpdate{
		ChosenInlineResult: update.ChosenInlineResult,
		Update:             update,
		Client:             update.Client,
	})
}

// CallbackQueryHandler it's typed handler for CallbackQuery.
type CallbackQueryHandler func(context.Context, *CallbackQueryUpdate) error

func (handler CallbackQueryHandler) Handle(ctx context.Context, update *Update) error {
	return handler(ctx, &CallbackQueryUpdate{
		CallbackQuery: update.CallbackQuery,
		Update:        update,
		Client:        update.Client,
	})
}

// ShippingQueryHandler it's typed handler for ShippingQuery.
type ShippingQueryHandler func(context.Context, *ShippingQueryUpdate) error

func (handler ShippingQueryHandler) Handle(ctx context.Context, update *Update) error {
	return handler(ctx, &ShippingQueryUpdate{
		ShippingQuery: update.ShippingQuery,
		Update:        update,
		Client:        update.Client,
	})
}

// PreCheckoutQueryHandler it's typed handler for PreCheckoutQuery.
type PreCheckoutQueryHandler func(context.Context, *PreCheckoutQueryUpdate) error

func (handler PreCheckoutQueryHandler) Handle(ctx context.Context, update *Update) error {
	return handler(ctx, &PreCheckoutQueryUpdate{
		PreCheckoutQuery: update.PreCheckoutQuery,
		Update:           update,
		Client:           update.Client,
	})
}

// PollHandler it's typed handler for Poll.
type PollHandler func(context.Context, *PollUpdate) error

func (handler PollHandler) Handle(ctx context.Context, update *Update) error {
	return handler(ctx, &PollUpdate{
		Poll:   update.Poll,
		Update: update,
		Client: update.Client,
	})
}

// PollAnswerHandler it's typed handler for PollAnswer.
type PollAnswerHandler func(context.Context, *PollAnswerUpdate) error

func (handler PollAnswerHandler) Handle(ctx context.Context, update *Update) error {
	return handler(ctx, &PollAnswerUpdate{
		PollAnswer: update.PollAnswer,
		Update:     update,
		Client:     update.Client,
	})
}

// UpdateHandler it's typed handler for ChatMemberUpdate subtype.
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

// ChatJoinRequestHandler it's typed handler for ChatJoinRequest.
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

// MessageReactionHandler it's typed handler for [MessageReactionUpdate].
type MessageReactionHandler func(context.Context, *MessageReactionUpdate) error

func (handler MessageReactionHandler) Handle(ctx context.Context, update *Update) error {
	return handler(ctx, &MessageReactionUpdate{
		MessageReactionUpdated: update.MessageReaction,
		Update:                 update,
		Client:                 update.Client,
	})
}

// MessageReactionCountHandler it's typed handler for [MessageReactionCountUpdate].
type MessageReactionCountHandler func(context.Context, *MessageReactionCountUpdate) error

func (handler MessageReactionCountHandler) Handle(ctx context.Context, update *Update) error {
	return handler(ctx, &MessageReactionCountUpdate{
		MessageReactionCountUpdated: update.MessageReactionCount,
		Update:                      update,
		Client:                      update.Client,
	})
}

// ChatBoostHandler it's typed handler for [ChatBoostUpdate].
type ChatBoostHandler func(context.Context, *ChatBoostUpdate) error

func (handler ChatBoostHandler) Handle(ctx context.Context, update *Update) error {
	return handler(ctx, &ChatBoostUpdate{
		ChatBoostUpdated: update.ChatBoost,
		Update:           update,
		Client:           update.Client,
	})
}

// RemovedChatBoostHandler it's typed handler for [RemovedChatBoostUpdate].
type RemovedChatBoostHandler func(context.Context, *RemovedChatBoostUpdate) error

func (handler RemovedChatBoostHandler) Handle(ctx context.Context, update *Update) error {
	return handler(ctx, &RemovedChatBoostUpdate{
		ChatBoostRemoved: update.RemovedChatBoost,
		Update:           update,
		Client:           update.Client,
	})
}

// BusinessConnection it's typed handler for [BusinessConnectionUpdate].
type BusinessConnectionHandler func(context.Context, *BusinessConnectionUpdate) error

func (handler BusinessConnectionHandler) Handle(ctx context.Context, update *Update) error {
	return handler(ctx, &BusinessConnectionUpdate{
		BusinessConnection: update.BusinessConnection,
		Update:             update,
		Client:             update.Client,
	})
}

// DeletedBusinessMessageHandler it's typed handler for [DeletedBusinessMessage]
type DeletedBusinessMessageHandler func(context.Context, *DeletedBusinessMessagesUpdate) error

func (handler DeletedBusinessMessageHandler) Handle(ctx context.Context, update *Update) error {
	return handler(ctx, &DeletedBusinessMessagesUpdate{
		BusinessMessagesDeleted: update.DeletedBusinessMessages,
		Update:                  update,
		Client:                  update.Client,
	})
}
