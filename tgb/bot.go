package tgb

import (
	"context"
	"fmt"
)

// ErrorHandler define interface for error handling in Bot.
// See Bot.Error for more information.
type ErrorHandler func(ctx context.Context, update *Update, err error) error

type registeredHandler struct {
	Handler Handler
	Filter  Filter
}

// Bot is a router for incoming Updates.
// tg.Update should be wrapped into tgb.Update with binded Client and Update.
type Bot struct {
	chain                     chain
	messageHandler            []*registeredHandler
	editedMessageHandler      []*registeredHandler
	channelPostHandler        []*registeredHandler
	editedChannelPostHandler  []*registeredHandler
	inlineQueryHandler        []*registeredHandler
	chosenInlineResultHandler []*registeredHandler
	callbackQueryHandler      []*registeredHandler
	shippingQueryHandler      []*registeredHandler
	preCheckoutQueryHandler   []*registeredHandler
	pollHandler               []*registeredHandler
	pollAnswerHandler         []*registeredHandler
	myChatMemberHandler       []*registeredHandler
	chatMemberHandler         []*registeredHandler
	chatJoinRequestHandler    []*registeredHandler
	updateHandler             []*registeredHandler
	errorHandler              ErrorHandler
}

// New creates new Bot.
func New() *Bot {
	return &Bot{
		chain: chain{},
	}
}

func compactFilter(filters ...Filter) Filter {
	if len(filters) == 1 {
		return filters[0]
	} else if len(filters) > 1 {
		return All(filters...)
	}
	return nil
}

// Use add middleware to chain handlers.
// Should be called before any other register handler.
func (bot *Bot) Use(mws ...Middleware) *Bot {
	bot.chain = bot.chain.Append(mws...)
	return bot
}

// Message register handlers for Update with not empty Message field.
func (bot *Bot) Message(handler MessageHandler, filters ...Filter) *Bot {
	bot.messageHandler = append(bot.messageHandler, &registeredHandler{
		Handler: bot.chain.Then(handler),
		Filter:  compactFilter(filters...),
	})
	return bot
}

// EditedMessage register handlers for Update with not empty EditedMessage field.
func (bot *Bot) EditedMessage(handler MessageHandler, filters ...Filter) *Bot {
	bot.editedMessageHandler = append(bot.editedMessageHandler, &registeredHandler{
		Handler: bot.chain.Then(handler),
		Filter:  compactFilter(filters...),
	})
	return bot
}

// ChannelPost register handlers for Update with not empty ChannelPost field.
func (bot *Bot) ChannelPost(handler MessageHandler, filters ...Filter) *Bot {
	bot.channelPostHandler = append(bot.channelPostHandler, &registeredHandler{
		Handler: bot.chain.Then(handler),
		Filter:  compactFilter(filters...),
	})
	return bot
}

// EditedChannelPost register handlers for Update with not empty EditedChannelPost field.
func (bot *Bot) EditedChannelPost(handler MessageHandler, filters ...Filter) *Bot {
	bot.editedChannelPostHandler = append(bot.editedChannelPostHandler, &registeredHandler{
		Handler: bot.chain.Then(handler),
		Filter:  compactFilter(filters...),
	})
	return bot
}

// InlineQuery register handlers for Update with not empty InlineQuery field.
func (bot *Bot) InlineQuery(handler InlineQueryHandler, filters ...Filter) *Bot {
	bot.inlineQueryHandler = append(bot.inlineQueryHandler, &registeredHandler{
		Handler: bot.chain.Then(handler),
		Filter:  compactFilter(filters...),
	})
	return bot
}

// ChosenInlineResult register handlers for Update with not empty ChosenInlineResult field.
func (bot *Bot) ChosenInlineResult(handler ChosenInlineResultHandler, filters ...Filter) *Bot {
	bot.chosenInlineResultHandler = append(bot.chosenInlineResultHandler, &registeredHandler{
		Handler: bot.chain.Then(handler),
		Filter:  compactFilter(filters...),
	})
	return bot
}

// CallbackQuery register handlers for Update with not empty CallbackQuery field.
func (bot *Bot) CallbackQuery(handler CallbackQueryHandler, filters ...Filter) *Bot {
	bot.callbackQueryHandler = append(bot.callbackQueryHandler, &registeredHandler{
		Handler: bot.chain.Then(handler),
		Filter:  compactFilter(filters...),
	})
	return bot
}

// ShippingQuery register handlers for Update with not empty ShippingQuery field.
func (bot *Bot) ShippingQuery(handler ShippingQueryHandler, filters ...Filter) *Bot {
	bot.shippingQueryHandler = append(bot.shippingQueryHandler, &registeredHandler{
		Handler: bot.chain.Then(handler),
		Filter:  compactFilter(filters...),
	})
	return bot
}

// PreCheckoutQuery register handlers for Update with not empty PreCheckoutQuery field.
func (bot *Bot) PreCheckoutQuery(handler PreCheckoutQueryHandler, filters ...Filter) *Bot {
	bot.preCheckoutQueryHandler = append(bot.preCheckoutQueryHandler, &registeredHandler{
		Handler: bot.chain.Then(handler),
		Filter:  compactFilter(filters...),
	})
	return bot
}

// Poll register handlers for Update with not empty Poll field.
func (bot *Bot) Poll(handler PollHandler, filters ...Filter) *Bot {
	bot.pollHandler = append(bot.pollHandler, &registeredHandler{
		Handler: bot.chain.Then(handler),
		Filter:  compactFilter(filters...),
	})
	return bot
}

// PollAnswer register handlers for Update with not empty PollAnswer field.
func (bot *Bot) PollAnswer(handler PollAnswerHandler, filters ...Filter) *Bot {
	bot.pollAnswerHandler = append(bot.pollAnswerHandler, &registeredHandler{
		Handler: bot.chain.Then(handler),
		Filter:  compactFilter(filters...),
	})
	return bot
}

// MyChatMember register handlers for Update with not empty MyChatMember field.
func (bot *Bot) MyChatMember(handler ChatMemberUpdatedHandler, filters ...Filter) *Bot {
	bot.myChatMemberHandler = append(bot.myChatMemberHandler, &registeredHandler{
		Handler: bot.chain.Then(handler),
		Filter:  compactFilter(filters...),
	})
	return bot
}

// ChatMember register handlers for Update with not empty ChatMember field.
func (bot *Bot) ChatMember(handler ChatMemberUpdatedHandler, filters ...Filter) *Bot {
	bot.chatMemberHandler = append(bot.chatMemberHandler, &registeredHandler{
		Handler: bot.chain.Then(handler),
		Filter:  compactFilter(filters...),
	})
	return bot
}

// ChatJoinRequest register handlers for Update with not empty ChatJoinRequest field.
func (bot *Bot) ChatJoinRequest(handler ChatJoinRequestHandler, filters ...Filter) *Bot {
	bot.chatJoinRequestHandler = append(bot.chatJoinRequestHandler, &registeredHandler{
		Handler: bot.chain.Then(handler),
		Filter:  compactFilter(filters...),
	})
	return bot
}

// Error registers a handler for errors.
// If any error occurs in the chain, it will be passed to that handler.
// By default, errors are returned back by handler method.
// You can customize this behavior by passing a custom error handler.
func (bot *Bot) Error(handler ErrorHandler) *Bot {
	bot.errorHandler = handler
	return bot
}

// Update registers a generic Update handler.
// It will be called as typed handlers only in filters match the update.
// First check Update handler, then typed.
func (bot *Bot) Update(handler HandlerFunc, filters ...Filter) *Bot {
	bot.updateHandler = append(bot.updateHandler, &registeredHandler{
		Handler: bot.chain.Then(handler),
		Filter:  compactFilter(filters...),
	})
	return bot
}

func (bot *Bot) pickAndHandle(ctx context.Context, update *Update, group []*registeredHandler) error {
	for _, item := range group {
		if item.Filter != nil {
			allow, err := item.Filter.Allow(ctx, update)
			if err != nil {
				return fmt.Errorf("filter %T: %w", item.Filter, err)
			}
			if !allow {
				continue
			}
		}

		return item.Handler.Handle(ctx, update)
	}

	return nil
}

// Handle handles an Update.
func (bot *Bot) Handle(ctx context.Context, update *Update) error {
	group := append([]*registeredHandler{}, bot.updateHandler...)

	switch {
	case update.Message != nil:
		group = append(group, bot.messageHandler...)
	case update.EditedMessage != nil:
		group = append(group, bot.editedMessageHandler...)
	case update.ChannelPost != nil:
		group = append(group, bot.channelPostHandler...)
	case update.EditedChannelPost != nil:
		group = append(group, bot.editedChannelPostHandler...)
	case update.InlineQuery != nil:
		group = append(group, bot.inlineQueryHandler...)
	case update.ChosenInlineResult != nil:
		group = append(group, bot.chosenInlineResultHandler...)
	case update.CallbackQuery != nil:
		group = append(group, bot.callbackQueryHandler...)
	case update.ShippingQuery != nil:
		group = append(group, bot.shippingQueryHandler...)
	case update.PreCheckoutQuery != nil:
		group = append(group, bot.preCheckoutQueryHandler...)
	case update.Poll != nil:
		group = append(group, bot.pollHandler...)
	case update.PollAnswer != nil:
		group = append(group, bot.pollAnswerHandler...)
	case update.MyChatMember != nil:
		group = append(group, bot.myChatMemberHandler...)
	case update.ChatMember != nil:
		group = append(group, bot.chatMemberHandler...)
	case update.ChatJoinRequest != nil:
		group = append(group, bot.chatJoinRequestHandler...)
	default:
		return nil
	}

	if err := bot.pickAndHandle(ctx, update, group); err != nil {
		if bot.errorHandler != nil {
			return bot.errorHandler(ctx, update, err)
		}
		return err
	}

	return nil
}
