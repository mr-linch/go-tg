package tgb

import (
	"context"
	"testing"

	"github.com/mr-linch/go-tg"
	"github.com/stretchr/testify/assert"
)

func TestBot(t *testing.T) {
	t.Run("Handlers", func(t *testing.T) {
		bot := New()
		ctx := context.Background()

		assert.NotNil(t, bot, "bot should be not nil")

		var isMiddelwareCallled bool

		bot.Use(
			func(next Handler) Handler {
				return HandlerFunc(func(ctx context.Context, update *Update) error {
					assert.NotNil(t, update)

					isMiddelwareCallled = true
					return next.Handle(ctx, update)
				})
			},
		)

		{
			var isMessageHandlerCalled bool

			bot.Message(func(ctx context.Context, msg *MessageUpdate) error {
				assert.NotNil(t, msg.Message)
				isMessageHandlerCalled = true
				return nil
			})

			err := bot.Handle(ctx, &Update{Update: &tg.Update{
				Message: &tg.Message{},
			}})

			assert.NoError(t, err)
			assert.True(t, isMiddelwareCallled, "middleware should be called")
			assert.True(t, isMessageHandlerCalled, "message handler should be called")
		}

		{
			isEditedMessageHandlerCalled := false

			bot.EditedMessage(func(ctx context.Context, msg *MessageUpdate) error {
				assert.NotNil(t, msg.Message)
				isEditedMessageHandlerCalled = true
				return nil
			})

			err := bot.Handle(ctx, &Update{Update: &tg.Update{
				EditedMessage: &tg.Message{},
			}})

			assert.NoError(t, err)
			assert.True(t, isEditedMessageHandlerCalled, "edited message handler should be called")
		}

		{
			isChannelPostHandlerCalled := false

			bot.ChannelPost(func(ctx context.Context, msg *MessageUpdate) error {
				assert.NotNil(t, msg.Message)
				isChannelPostHandlerCalled = true
				return nil
			})

			err := bot.Handle(ctx, &Update{Update: &tg.Update{
				ChannelPost: &tg.Message{},
			}})

			assert.NoError(t, err)
			assert.True(t, isChannelPostHandlerCalled, "channel post handler should be called")
		}

		{
			isEditedChannelPostHandlerCalled := false

			bot.EditedChannelPost(func(ctx context.Context, msg *MessageUpdate) error {
				assert.NotNil(t, msg.Message)
				isEditedChannelPostHandlerCalled = true
				return nil
			})

			err := bot.Handle(ctx, &Update{Update: &tg.Update{
				EditedChannelPost: &tg.Message{},
			}})

			assert.NoError(t, err)
			assert.True(t, isEditedChannelPostHandlerCalled, "edited channel post handler should be called")
		}

		{
			isInlineQueryHandlerCalled := false

			bot.InlineQuery(func(ctx context.Context, msg *InlineQueryUpdate) error {
				assert.NotNil(t, msg.InlineQuery)
				isInlineQueryHandlerCalled = true
				return nil
			})

			err := bot.Handle(ctx, &Update{Update: &tg.Update{
				InlineQuery: &tg.InlineQuery{},
			}})

			assert.NoError(t, err)
			assert.True(t, isInlineQueryHandlerCalled, "inline query handler should be called")
		}

		{
			isChosenInlineResultHandlerCalled := false

			bot.ChosenInlineResult(func(ctx context.Context, msg *ChosenInlineResultUpdate) error {
				assert.NotNil(t, msg.ChosenInlineResult)
				isChosenInlineResultHandlerCalled = true
				return nil
			})

			err := bot.Handle(ctx, &Update{Update: &tg.Update{
				ChosenInlineResult: &tg.ChosenInlineResult{},
			}})

			assert.NoError(t, err)
			assert.True(t, isChosenInlineResultHandlerCalled, "choosen inline result handler should be called")
		}

		{
			isCallbackQueryHandlerCalled := false

			bot.CallbackQuery(func(ctx context.Context, msg *CallbackQueryUpdate) error {
				assert.NotNil(t, msg.CallbackQuery)
				isCallbackQueryHandlerCalled = true
				return nil
			})

			err := bot.Handle(ctx, &Update{Update: &tg.Update{
				CallbackQuery: &tg.CallbackQuery{},
			}})

			assert.NoError(t, err)
			assert.True(t, isCallbackQueryHandlerCalled, "callback query handler should be called")
		}

		{
			isShippingQueryHandlerCalled := false

			bot.ShippingQuery(func(ctx context.Context, msg *ShippingQueryUpdate) error {
				assert.NotNil(t, msg.ShippingQuery)
				isShippingQueryHandlerCalled = true
				return nil
			})

			err := bot.Handle(ctx, &Update{Update: &tg.Update{
				ShippingQuery: &tg.ShippingQuery{},
			}})

			assert.NoError(t, err)
			assert.True(t, isShippingQueryHandlerCalled, "shipping query handler should be called")
		}

		{
			isPreCheckoutQueryHandlerCalled := false

			bot.PreCheckoutQuery(func(ctx context.Context, msg *PreCheckoutQueryUpdate) error {
				assert.NotNil(t, msg.PreCheckoutQuery)
				isPreCheckoutQueryHandlerCalled = true
				return nil
			})

			err := bot.Handle(ctx, &Update{Update: &tg.Update{
				PreCheckoutQuery: &tg.PreCheckoutQuery{},
			}})

			assert.NoError(t, err)
			assert.True(t, isPreCheckoutQueryHandlerCalled, "pre checkout query handler should be called")
		}

		{
			isPollHandlerCalled := false

			bot.Poll(func(ctx context.Context, msg *PollUpdate) error {
				assert.NotNil(t, msg.Poll)
				isPollHandlerCalled = true
				return nil
			})

			err := bot.Handle(ctx, &Update{Update: &tg.Update{
				Poll: &tg.Poll{},
			}})

			assert.NoError(t, err)
			assert.True(t, isPollHandlerCalled, "poll handler should be called")
		}

		{
			isPollAnswerHandlerCalled := false

			bot.PollAnswer(func(ctx context.Context, msg *PollAnswerUpdate) error {
				assert.NotNil(t, msg.PollAnswer)
				isPollAnswerHandlerCalled = true
				return nil
			})

			err := bot.Handle(ctx, &Update{Update: &tg.Update{
				PollAnswer: &tg.PollAnswer{},
			}})

			assert.NoError(t, err)
			assert.True(t, isPollAnswerHandlerCalled, "poll answer handler should be called")

		}

		{
			isMyChatMemberHandlerCalled := false

			bot.MyChatMember(func(ctx context.Context, msg *ChatMemberUpdatedUpdate) error {
				assert.NotNil(t, msg.ChatMemberUpdated)
				isMyChatMemberHandlerCalled = true
				return nil
			})

			err := bot.Handle(ctx, &Update{Update: &tg.Update{
				MyChatMember: &tg.ChatMemberUpdated{},
			}})

			assert.NoError(t, err)
			assert.True(t, isMyChatMemberHandlerCalled, "my chat member handler should be called")
		}

		{
			isChatMemberHandlerCalled := false

			bot.ChatMember(func(ctx context.Context, msg *ChatMemberUpdatedUpdate) error {
				assert.NotNil(t, msg.ChatMemberUpdated)
				isChatMemberHandlerCalled = true
				return nil
			})

			err := bot.Handle(ctx, &Update{Update: &tg.Update{
				ChatMember: &tg.ChatMemberUpdated{},
			}})

			assert.NoError(t, err)
			assert.True(t, isChatMemberHandlerCalled, "chat member handler should be called")
		}

		{
			isChatJoinRequestHandlerCalled := false

			bot.ChatJoinRequest(func(ctx context.Context, msg *ChatJoinRequestUpdate) error {
				assert.NotNil(t, msg.ChatJoinRequest)
				isChatJoinRequestHandlerCalled = true
				return nil
			})

			err := bot.Handle(ctx, &Update{Update: &tg.Update{
				ChatJoinRequest: &tg.ChatJoinRequest{},
			}})

			assert.NoError(t, err)
			assert.True(t, isChatJoinRequestHandlerCalled, "chat join request handler should be called")
		}
	})

	t.Run("FilterNotAllow", func(t *testing.T) {
		isPrivateChatHandlerCalled := false
		isGroupChatHandlerCalled := false
		isGroupAndPrivateChatHandlerCalled := false

		bot := New().
			Message(func(context.Context, *MessageUpdate) error {
				isPrivateChatHandlerCalled = true
				return nil
			}, ChatType(tg.ChatTypePrivate)).
			Message(func(context.Context, *MessageUpdate) error {
				isGroupChatHandlerCalled = true
				return nil
			}, Any(ChatType(tg.ChatTypeGroup), ChatType(tg.ChatTypeSupergroup))).
			Message(func(context.Context, *MessageUpdate) error {
				isGroupAndPrivateChatHandlerCalled = true
				return nil
			}, ChatType(tg.ChatTypePrivate), ChatType(tg.ChatTypeGroup))

		err := bot.Handle(context.Background(), &Update{
			Update: &tg.Update{
				Message: &tg.Message{
					Chat: tg.Chat{
						Type: tg.ChatTypePrivate,
					},
				},
			},
		})

		assert.NoError(t, err)
		assert.True(t, isPrivateChatHandlerCalled, "private chat handler should be called")
		assert.False(t, isGroupChatHandlerCalled, "group chat handler should not be called")
		assert.False(t, isGroupAndPrivateChatHandlerCalled, "group and private chat handler should not be called")

		isPrivateChatHandlerCalled = false
		isGroupChatHandlerCalled = false
		isGroupAndPrivateChatHandlerCalled = false

		err = bot.Handle(context.Background(), &Update{
			Update: &tg.Update{
				Message: &tg.Message{
					Chat: tg.Chat{
						Type: tg.ChatTypeGroup,
					},
				},
			},
		})

		assert.NoError(t, err)
		assert.False(t, isPrivateChatHandlerCalled, "private chat handler should not be called")
		assert.True(t, isGroupChatHandlerCalled, "group chat handler should be called")
		assert.False(t, isGroupAndPrivateChatHandlerCalled, "group and private chat handler should not be called")
	})

}
