package tgb

import (
	"context"
	"fmt"
	"testing"

	"github.com/mr-linch/go-tg"
	"github.com/stretchr/testify/assert"
)

func TestRouter(t *testing.T) {
	t.Run("HandleEmpty", func(t *testing.T) {
		err := NewRouter().Handle(context.Background(), &Update{
			Update: &tg.Update{
				Message: &tg.Message{},
			},
		})
		assert.NoError(t, err)
	})

	t.Run("UpdateAndMessageHanlder", func(t *testing.T) {
		isMessageHandlerCalled := false
		isUpdateHanlderCalled := false
		err := NewRouter().
			Message(func(ctx context.Context, msg *MessageUpdate) error {
				isMessageHandlerCalled = true
				return nil
			}).
			Update(func(ctx context.Context, update *Update) error {
				isUpdateHanlderCalled = true
				return nil
			}).
			Handle(context.Background(), &Update{
				Update: &tg.Update{
					Message: &tg.Message{},
				},
			})

		assert.NoError(t, err)
		assert.False(t, isMessageHandlerCalled)
		assert.True(t, isUpdateHanlderCalled)
	})

	t.Run("UpdateOnlyHandler", func(t *testing.T) {
		isUpdateHanlderCalled := false
		err := NewRouter().
			Update(func(ctx context.Context, update *Update) error {
				isUpdateHanlderCalled = true
				return nil
			}).
			Handle(context.Background(), &Update{
				Update: &tg.Update{
					Message: &tg.Message{},
				},
			})

		assert.NoError(t, err)
		assert.True(t, isUpdateHanlderCalled, "update handler is not called")
	})

	t.Run("UnknownUpdateSubtype", func(t *testing.T) {
		err := NewRouter().Message(func(ctx context.Context, msg *MessageUpdate) error {
			return nil
		}).Handle(context.Background(), &Update{
			Update: &tg.Update{},
		})

		assert.NoError(t, err)
	})
	t.Run("AllowError", func(t *testing.T) {

		err := NewRouter().
			Message(func(ctx context.Context, msg *MessageUpdate) error {
				return nil
			}, FilterFunc(func(ctx context.Context, update *Update) (bool, error) {
				return false, fmt.Errorf("failure")
			})).Handle(context.Background(), &Update{Update: &tg.Update{
			Message: &tg.Message{},
		}})

		assert.EqualError(t, err, "filter error: failure")

	})
	t.Run("Error", func(t *testing.T) {
		handlerErr := fmt.Errorf("handler error")

		router := NewRouter().
			Message(func(ctx context.Context, msg *MessageUpdate) error {
				return handlerErr
			})

		err := router.Handle(context.Background(), &Update{
			Update: &tg.Update{
				Message: &tg.Message{},
			},
		})

		assert.Equal(t, handlerErr, err)

		isErrorHandlerCalled := false

		router.Error(func(ctx context.Context, update *Update, err error) error {
			isErrorHandlerCalled = true
			assert.Equal(t, handlerErr, err)
			return nil
		})

		err = router.Handle(context.Background(), &Update{
			Update: &tg.Update{
				Message: &tg.Message{},
			},
		})

		assert.Nil(t, err)
		assert.True(t, isErrorHandlerCalled)
	})

	t.Run("Handlers", func(t *testing.T) {
		router := NewRouter()
		ctx := context.Background()

		assert.NotNil(t, router, "bot should be not nil")

		var isMiddelwareCallled bool

		router.Use(
			MiddlewareFunc(func(next Handler) Handler {
				return HandlerFunc(func(ctx context.Context, update *Update) error {
					assert.NotNil(t, update)

					isMiddelwareCallled = true
					return next.Handle(ctx, update)
				})
			}),
		)

		{
			var isMessageHandlerCalled bool

			router.Message(func(ctx context.Context, msg *MessageUpdate) error {
				assert.NotNil(t, msg.Message)
				isMessageHandlerCalled = true
				return nil
			})

			err := router.Handle(ctx, &Update{Update: &tg.Update{
				Message: &tg.Message{},
			}})

			assert.NoError(t, err)
			assert.True(t, isMiddelwareCallled, "middleware should be called")
			assert.True(t, isMessageHandlerCalled, "message handler should be called")
		}

		{
			isEditedMessageHandlerCalled := false

			router.EditedMessage(func(ctx context.Context, msg *MessageUpdate) error {
				assert.NotNil(t, msg.Message)
				isEditedMessageHandlerCalled = true
				return nil
			})

			err := router.Handle(ctx, &Update{Update: &tg.Update{
				EditedMessage: &tg.Message{},
			}})

			assert.NoError(t, err)
			assert.True(t, isEditedMessageHandlerCalled, "edited message handler should be called")
		}

		{
			isChannelPostHandlerCalled := false

			router.ChannelPost(func(ctx context.Context, msg *MessageUpdate) error {
				assert.NotNil(t, msg.Message)
				isChannelPostHandlerCalled = true
				return nil
			})

			err := router.Handle(ctx, &Update{Update: &tg.Update{
				ChannelPost: &tg.Message{},
			}})

			assert.NoError(t, err)
			assert.True(t, isChannelPostHandlerCalled, "channel post handler should be called")
		}

		{
			isEditedChannelPostHandlerCalled := false

			router.EditedChannelPost(func(ctx context.Context, msg *MessageUpdate) error {
				assert.NotNil(t, msg.Message)
				isEditedChannelPostHandlerCalled = true
				return nil
			})

			err := router.Handle(ctx, &Update{Update: &tg.Update{
				EditedChannelPost: &tg.Message{},
			}})

			assert.NoError(t, err)
			assert.True(t, isEditedChannelPostHandlerCalled, "edited channel post handler should be called")
		}

		{
			isInlineQueryHandlerCalled := false

			router.InlineQuery(func(ctx context.Context, msg *InlineQueryUpdate) error {
				assert.NotNil(t, msg.InlineQuery)
				isInlineQueryHandlerCalled = true
				return nil
			})

			err := router.Handle(ctx, &Update{Update: &tg.Update{
				InlineQuery: &tg.InlineQuery{},
			}})

			assert.NoError(t, err)
			assert.True(t, isInlineQueryHandlerCalled, "inline query handler should be called")
		}

		{
			isChosenInlineResultHandlerCalled := false

			router.ChosenInlineResult(func(ctx context.Context, msg *ChosenInlineResultUpdate) error {
				assert.NotNil(t, msg.ChosenInlineResult)
				isChosenInlineResultHandlerCalled = true
				return nil
			})

			err := router.Handle(ctx, &Update{Update: &tg.Update{
				ChosenInlineResult: &tg.ChosenInlineResult{},
			}})

			assert.NoError(t, err)
			assert.True(t, isChosenInlineResultHandlerCalled, "chosen inline result handler should be called")
		}

		{
			isCallbackQueryHandlerCalled := false

			router.CallbackQuery(func(ctx context.Context, msg *CallbackQueryUpdate) error {
				assert.NotNil(t, msg.CallbackQuery)
				isCallbackQueryHandlerCalled = true
				return nil
			})

			err := router.Handle(ctx, &Update{Update: &tg.Update{
				CallbackQuery: &tg.CallbackQuery{},
			}})

			assert.NoError(t, err)
			assert.True(t, isCallbackQueryHandlerCalled, "callback query handler should be called")
		}

		{
			isShippingQueryHandlerCalled := false

			router.ShippingQuery(func(ctx context.Context, msg *ShippingQueryUpdate) error {
				assert.NotNil(t, msg.ShippingQuery)
				isShippingQueryHandlerCalled = true
				return nil
			})

			err := router.Handle(ctx, &Update{Update: &tg.Update{
				ShippingQuery: &tg.ShippingQuery{},
			}})

			assert.NoError(t, err)
			assert.True(t, isShippingQueryHandlerCalled, "shipping query handler should be called")
		}

		{
			isPreCheckoutQueryHandlerCalled := false

			router.PreCheckoutQuery(func(ctx context.Context, msg *PreCheckoutQueryUpdate) error {
				assert.NotNil(t, msg.PreCheckoutQuery)
				isPreCheckoutQueryHandlerCalled = true
				return nil
			})

			err := router.Handle(ctx, &Update{Update: &tg.Update{
				PreCheckoutQuery: &tg.PreCheckoutQuery{},
			}})

			assert.NoError(t, err)
			assert.True(t, isPreCheckoutQueryHandlerCalled, "pre checkout query handler should be called")
		}

		{
			isPollHandlerCalled := false

			router.Poll(func(ctx context.Context, msg *PollUpdate) error {
				assert.NotNil(t, msg.Poll)
				isPollHandlerCalled = true
				return nil
			})

			err := router.Handle(ctx, &Update{Update: &tg.Update{
				Poll: &tg.Poll{},
			}})

			assert.NoError(t, err)
			assert.True(t, isPollHandlerCalled, "poll handler should be called")
		}

		{
			isPollAnswerHandlerCalled := false

			router.PollAnswer(func(ctx context.Context, msg *PollAnswerUpdate) error {
				assert.NotNil(t, msg.PollAnswer)
				isPollAnswerHandlerCalled = true
				return nil
			})

			err := router.Handle(ctx, &Update{Update: &tg.Update{
				PollAnswer: &tg.PollAnswer{},
			}})

			assert.NoError(t, err)
			assert.True(t, isPollAnswerHandlerCalled, "poll answer handler should be called")

		}

		{
			isMyChatMemberHandlerCalled := false

			router.MyChatMember(func(ctx context.Context, msg *ChatMemberUpdatedUpdate) error {
				assert.NotNil(t, msg.ChatMemberUpdated)
				isMyChatMemberHandlerCalled = true
				return nil
			})

			err := router.Handle(ctx, &Update{Update: &tg.Update{
				MyChatMember: &tg.ChatMemberUpdated{},
			}})

			assert.NoError(t, err)
			assert.True(t, isMyChatMemberHandlerCalled, "my chat member handler should be called")
		}

		{
			isChatMemberHandlerCalled := false

			router.ChatMember(func(ctx context.Context, msg *ChatMemberUpdatedUpdate) error {
				assert.NotNil(t, msg.ChatMemberUpdated)
				isChatMemberHandlerCalled = true
				return nil
			})

			err := router.Handle(ctx, &Update{Update: &tg.Update{
				ChatMember: &tg.ChatMemberUpdated{},
			}})

			assert.NoError(t, err)
			assert.True(t, isChatMemberHandlerCalled, "chat member handler should be called")
		}

		{
			isChatJoinRequestHandlerCalled := false

			router.ChatJoinRequest(func(ctx context.Context, msg *ChatJoinRequestUpdate) error {
				assert.NotNil(t, msg.ChatJoinRequest)
				isChatJoinRequestHandlerCalled = true
				return nil
			})

			err := router.Handle(ctx, &Update{Update: &tg.Update{
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

		router := NewRouter().
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

		err := router.Handle(context.Background(), &Update{
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

		err = router.Handle(context.Background(), &Update{
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
