package tgb

import (
	"context"
	"fmt"

	tg "github.com/mr-linch/go-tg"
)

type registeredHandler struct {
	Handler Handler
	Filter  Filter
}

type Bot struct {
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
}

func New() *Bot {
	return &Bot{}
}

func compactFilter(filters ...Filter) Filter {
	if len(filters) == 1 {
		return filters[0]
	} else if len(filters) > 1 {
		return All(filters...)
	}
	return nil
}

func (bot *Bot) Message(handler Handler, filters ...Filter) *Bot {
	bot.messageHandler = append(bot.messageHandler, &registeredHandler{
		Handler: handler,
		Filter:  compactFilter(filters...),
	})
	return bot
}

func (bot *Bot) EditedMessage(handler Handler, filters ...Filter) *Bot {
	bot.editedMessageHandler = append(bot.editedMessageHandler, &registeredHandler{
		Handler: handler,
		Filter:  compactFilter(filters...),
	})
	return bot
}

func (bot *Bot) ChannelPost(handler Handler, filters ...Filter) *Bot {
	bot.channelPostHandler = append(bot.channelPostHandler, &registeredHandler{
		Handler: handler,
		Filter:  compactFilter(filters...),
	})
	return bot
}

func (bot *Bot) EditedChannelPost(handler Handler, filters ...Filter) *Bot {
	bot.editedChannelPostHandler = append(bot.editedChannelPostHandler, &registeredHandler{
		Handler: handler,
		Filter:  compactFilter(filters...),
	})
	return bot
}

func (bot *Bot) InlineQuery(handler Handler, filters ...Filter) *Bot {
	bot.inlineQueryHandler = append(bot.inlineQueryHandler, &registeredHandler{
		Handler: handler,
		Filter:  compactFilter(filters...),
	})
	return bot
}

func (bot *Bot) ChosenInlineResult(handler Handler, filters ...Filter) *Bot {
	bot.chosenInlineResultHandler = append(bot.chosenInlineResultHandler, &registeredHandler{
		Handler: handler,
		Filter:  compactFilter(filters...),
	})
	return bot
}

func (bot *Bot) CallbackQuery(handler Handler, filters ...Filter) *Bot {
	bot.callbackQueryHandler = append(bot.callbackQueryHandler, &registeredHandler{
		Handler: handler,
		Filter:  compactFilter(filters...),
	})
	return bot
}

func (bot *Bot) ShippingQuery(handler Handler, filters ...Filter) *Bot {
	bot.shippingQueryHandler = append(bot.shippingQueryHandler, &registeredHandler{
		Handler: handler,
		Filter:  compactFilter(filters...),
	})
	return bot
}

func (bot *Bot) PreCheckoutQuery(handler Handler, filters ...Filter) *Bot {
	bot.preCheckoutQueryHandler = append(bot.preCheckoutQueryHandler, &registeredHandler{
		Handler: handler,
		Filter:  compactFilter(filters...),
	})
	return bot
}

func (bot *Bot) Poll(handler Handler, filters ...Filter) *Bot {
	bot.pollHandler = append(bot.pollHandler, &registeredHandler{
		Handler: handler,
		Filter:  compactFilter(filters...),
	})
	return bot
}

func (bot *Bot) PollAnswer(handler Handler, filters ...Filter) *Bot {
	bot.pollAnswerHandler = append(bot.pollAnswerHandler, &registeredHandler{
		Handler: handler,
		Filter:  compactFilter(filters...),
	})
	return bot
}

func (bot *Bot) MyChatMember(handler Handler, filters ...Filter) *Bot {
	bot.myChatMemberHandler = append(bot.myChatMemberHandler, &registeredHandler{
		Handler: handler,
		Filter:  compactFilter(filters...),
	})
	return bot
}

func (bot *Bot) ChatMember(handler Handler, filters ...Filter) *Bot {
	bot.chatMemberHandler = append(bot.chatMemberHandler, &registeredHandler{
		Handler: handler,
		Filter:  compactFilter(filters...),
	})
	return bot
}

func (bot *Bot) ChatJoinRequest(handler Handler, filters ...Filter) *Bot {
	bot.chatJoinRequestHandler = append(bot.chatJoinRequestHandler, &registeredHandler{
		Handler: handler,
		Filter:  compactFilter(filters...),
	})
	return bot
}

func (bot *Bot) pickAndHandle(ctx context.Context, update *tg.Update, group []*registeredHandler) error {
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

func (bot *Bot) Handle(ctx context.Context, update *tg.Update) error {
	var group []*registeredHandler

	switch {
	case update.Message != nil:
		group = bot.messageHandler
	case update.EditedMessage != nil:
		group = bot.editedMessageHandler
	case update.ChannelPost != nil:
		group = bot.channelPostHandler
	case update.EditedChannelPost != nil:
		group = bot.editedChannelPostHandler
	case update.InlineQuery != nil:
		group = bot.inlineQueryHandler
	case update.ChosenInlineResult != nil:
		group = bot.chosenInlineResultHandler
	case update.CallbackQuery != nil:
		group = bot.callbackQueryHandler
	case update.ShippingQuery != nil:
		group = bot.shippingQueryHandler
	case update.PreCheckoutQuery != nil:
		group = bot.preCheckoutQueryHandler
	case update.Poll != nil:
		group = bot.pollHandler
	case update.PollAnswer != nil:
		group = bot.pollAnswerHandler
	case update.MyChatMember != nil:
		group = bot.myChatMemberHandler
	case update.ChatMember != nil:
		group = bot.chatMemberHandler
	case update.ChatJoinRequest != nil:
		group = bot.chatJoinRequestHandler
	default:
		return nil
	}

	return bot.pickAndHandle(ctx, update, group)
}
