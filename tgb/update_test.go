package tgb

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/mr-linch/go-tg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var _ UpdateReply = (*MockUpdateReply)(nil)

type MockUpdateReply struct {
	mock.Mock
}

func (mock *MockUpdateReply) Bind(client *tg.Client) {
	mock.Called(client)
}

func (mock *MockUpdateReply) MarshalJSON() ([]byte, error) {
	args := mock.Called()
	return args.Get(0).([]byte), args.Error(1)
}

func (mock *MockUpdateReply) DoVoid(ctx context.Context) error {
	args := mock.Called(ctx)
	return args.Error(0)
}

func TestUpdate_Reply(t *testing.T) {
	t.Run("NoWebhook", func(t *testing.T) {
		client := &tg.Client{}
		updateReply := &MockUpdateReply{}

		updateReply.On("Bind", client).Return()
		updateReply.On("DoVoid", mock.Anything).Return(nil)

		update := &Update{
			Client: client,
		}

		err := update.Reply(context.Background(), updateReply)
		assert.NoError(t, err)

		updateReply.AssertExpectations(t)
	})

	t.Run("NoWebhookUseRespond", func(t *testing.T) {
		client := &tg.Client{}
		updateReply := &MockUpdateReply{}

		updateReply.On("Bind", client).Return()
		updateReply.On("DoVoid", mock.Anything).Return(nil)

		update := &Update{
			Client: client,
		}

		err := update.Respond(context.Background(), updateReply)
		assert.NoError(t, err)

		updateReply.AssertExpectations(t)
	})

	t.Run("Webhook", func(t *testing.T) {
		updateReply := &MockUpdateReply{}

		updateReply.On("MarshalJSON", mock.Anything).Return([]byte{}, nil)

		update := &Update{
			webhookReply: make(chan json.Marshaler, 1),
		}

		err := update.Reply(context.Background(), updateReply)
		assert.NoError(t, err)

		obj := <-update.webhookReply

		assert.NotNil(t, obj)

		_, err = obj.MarshalJSON()
		assert.NoError(t, err)

		updateReply.AssertExpectations(t)

	})

}

type EncoderCollect struct {
	Args  map[string]string
	Files map[string]tg.InputFile
}

func (encoder *EncoderCollect) WriteString(k, v string) error {
	encoder.Args[k] = v
	return nil
}

func (encoder *EncoderCollect) WriteFile(k string, v tg.InputFile) error {
	encoder.Files[k] = v
	return nil
}

func TestMessageUpdateHelpers(t *testing.T) {
	msg := MessageUpdate{
		Message: &tg.Message{
			ID: 1,
			Chat: tg.Chat{
				ID: tg.ChatID(123),
			},
		},
	}

	for _, test := range []struct {
		Name           string
		Request        *tg.Request
		ExceptedMethod string
		ExpectedArgs   map[string]string
		ExpectedFiles  map[string]tg.InputFile
	}{
		{
			Name:           "Answer",
			Request:        msg.Answer("Hello").Request(),
			ExceptedMethod: "sendMessage",
			ExpectedArgs: map[string]string{
				"chat_id": "123",
				"text":    "Hello",
			},
		},
		{
			Name: "AnswerPhoto",
			Request: msg.AnswerPhoto(tg.FileArg{
				FileID: "file_id",
			}).Request(),
			ExceptedMethod: "sendPhoto",
			ExpectedArgs: map[string]string{
				"chat_id": "123",
				"photo":   "file_id",
			},
		},
		{
			Name: "AnswerAudio",
			Request: msg.AnswerAudio(tg.FileArg{
				FileID: "file_id",
			}).Request(),
			ExceptedMethod: "sendAudio",
			ExpectedArgs: map[string]string{
				"chat_id": "123",
				"audio":   "file_id",
			},
		},
		{
			Name: "AnswerAnimation",
			Request: msg.AnswerAnimation(tg.FileArg{
				FileID: "file_id",
			}).Request(),
			ExceptedMethod: "sendAnimation",
			ExpectedArgs: map[string]string{
				"chat_id":   "123",
				"animation": "file_id",
			},
		},
		{
			Name: "AnswerVideo",
			Request: msg.AnswerVideo(tg.FileArg{
				FileID: "file_id",
			}).Request(),
			ExceptedMethod: "sendVideo",
			ExpectedArgs: map[string]string{
				"chat_id": "123",
				"video":   "file_id",
			},
		},
		{
			Name: "AnswerVoice",
			Request: msg.AnswerVoice(tg.FileArg{
				FileID: "file_id",
			}).Request(),
			ExceptedMethod: "sendVoice",
			ExpectedArgs: map[string]string{
				"chat_id": "123",
				"voice":   "file_id",
			},
		},
		{
			Name: "AnswerDocument",
			Request: msg.AnswerDocument(tg.FileArg{
				FileID: "file_id",
			}).Request(),
			ExceptedMethod: "sendDocument",
			ExpectedArgs: map[string]string{
				"chat_id":  "123",
				"document": "file_id",
			},
		},
		{
			Name: "AnswerVideoNote",
			Request: msg.AnswerVideoNote(tg.FileArg{
				FileID: "file_id",
			}).Request(),
			ExceptedMethod: "sendVideoNote",
			ExpectedArgs: map[string]string{
				"chat_id":    "123",
				"video_note": "file_id",
			},
		},
		{
			Name:           "AnswerLocation",
			Request:        msg.AnswerLocation(1, 1).Request(),
			ExceptedMethod: "sendLocation",
			ExpectedArgs: map[string]string{
				"chat_id":   "123",
				"latitude":  "1",
				"longitude": "1",
			},
		},
		{
			Name:           "AnswerVenue",
			Request:        msg.AnswerVenue(1, 1, "title", "address").Request(),
			ExceptedMethod: "sendVenue",
			ExpectedArgs: map[string]string{
				"chat_id":   "123",
				"latitude":  "1",
				"longitude": "1",
				"title":     "title",
				"address":   "address",
			},
		},
		{
			Name:           "AnswerContact",
			Request:        msg.AnswerContact("1234", "sasha").Request(),
			ExceptedMethod: "sendContact",
			ExpectedArgs: map[string]string{
				"chat_id":      "123",
				"phone_number": "1234",
				"first_name":   "sasha",
			},
		},
		{
			Name: "AnswerSticker",
			Request: msg.AnswerSticker(tg.FileArg{
				FileID: "file_id",
			}).Request(),
			ExceptedMethod: "sendSticker",
			ExpectedArgs: map[string]string{
				"chat_id": "123",
				"sticker": "file_id",
			},
		},
		{
			Name:           "AnswerPoll",
			Request:        msg.AnswerPoll("question", []tg.InputPollOption{{Text: "1"}}).Request(),
			ExceptedMethod: "sendPoll",
			ExpectedArgs: map[string]string{
				"chat_id":  "123",
				"question": "question",
				"options":  "[{\"text\":\"1\"}]",
			},
		},
		{
			Name:           "AnswerDice",
			Request:        msg.AnswerDice("🎰").Request(),
			ExceptedMethod: "sendDice",
			ExpectedArgs: map[string]string{
				"chat_id": "123",
				"emoji":   "🎰",
			},
		},
		{
			Name:           "AnswerChatAction",
			Request:        msg.AnswerChatAction(tg.ChatActionUploadPhoto).Request(),
			ExceptedMethod: "sendChatAction",
			ExpectedArgs: map[string]string{
				"chat_id": "123",
				"action":  "upload_photo",
			},
		},
		{
			Name:           "Forward",
			Request:        msg.Forward(tg.ChatID(2)).Request(),
			ExceptedMethod: "forwardMessage",
			ExpectedArgs: map[string]string{
				"chat_id":      "2",
				"message_id":   "1",
				"from_chat_id": "123",
			},
		},
		{
			Name:           "Copy",
			Request:        msg.Copy(tg.ChatID(2)).Request(),
			ExceptedMethod: "copyMessage",
			ExpectedArgs: map[string]string{
				"chat_id":      "2",
				"message_id":   "1",
				"from_chat_id": "123",
			},
		},
		{
			Name:           "EditText",
			Request:        msg.EditText("text").Request(),
			ExceptedMethod: "editMessageText",
			ExpectedArgs: map[string]string{
				"chat_id":    "123",
				"message_id": "1",
				"text":       "text",
			},
		},
		{
			Name:           "EditCaption",
			Request:        msg.EditCaption("text").Request(),
			ExceptedMethod: "editMessageCaption",
			ExpectedArgs: map[string]string{
				"chat_id":    "123",
				"message_id": "1",
				"caption":    "text",
			},
		},
		{
			Name: "EditReplyMarkup",
			Request: msg.EditReplyMarkup(tg.NewInlineKeyboardMarkup(
				tg.NewButtonRow(
					tg.NewInlineKeyboardButtonCallback("1", "1"),
				),
			)).Request(),
			ExceptedMethod: "editMessageReplyMarkup",
			ExpectedArgs: map[string]string{
				"chat_id":      "123",
				"message_id":   "1",
				"reply_markup": "{\"inline_keyboard\":[[{\"text\":\"1\",\"callback_data\":\"1\"}]]}",
			},
		},
		{
			Name:           "React",
			Request:        msg.React(tg.ReactionTypeEmojiThumbsUp).Request(),
			ExceptedMethod: "setMessageReaction",
			ExpectedArgs: map[string]string{
				"chat_id":    "123",
				"message_id": "1",
				"reaction":   `[{"type":"emoji","emoji":"👍"}]`,
			},
		},
		{
			Name: "AnswerMediaGroup",
			Request: msg.AnswerMediaGroup([]tg.InputMedia{
				tg.NewInputMediaPhoto(tg.InputMediaPhoto{
					Media: tg.FileArg{
						FileID: "file_id",
					},
				}),
			}).Request(),
			ExceptedMethod: "sendMediaGroup",
			ExpectedArgs: map[string]string{
				"chat_id": "123",
				"media":   "[{\"type\":\"photo\",\"media\":\"file_id\"}]",
			},
		},
	} {
		t.Run(test.Name, func(t *testing.T) {
			encoder := EncoderCollect{
				Args:  make(map[string]string),
				Files: make(map[string]tg.InputFile),
			}

			err := test.Request.Encode(&encoder)
			assert.NoError(t, err)

			assert.Equal(t, test.ExceptedMethod, test.Request.Method)

			assert.Equal(t, test.ExpectedArgs, encoder.Args)

			for k, v := range test.ExpectedFiles {
				assert.Equal(t, v, encoder.Files[k])
			}
		})
	}

}

func TestCallbackQueryUpdateHelpers(t *testing.T) {
	cbq := CallbackQueryUpdate{
		CallbackQuery: &tg.CallbackQuery{
			ID: "1028493893",
		},
	}

	for _, test := range []struct {
		Name           string
		Request        *tg.Request
		ExceptedMethod string
		ExpectedArgs   map[string]string
		ExpectedFiles  map[string]tg.InputFile
	}{
		{
			Name:           "Answer",
			Request:        cbq.Answer().Request(),
			ExceptedMethod: "answerCallbackQuery",
			ExpectedArgs: map[string]string{
				"callback_query_id": cbq.ID,
			},
		},
		{
			Name:           "AnswerText",
			Request:        cbq.AnswerText("text", true).Request(),
			ExceptedMethod: "answerCallbackQuery",
			ExpectedArgs: map[string]string{
				"callback_query_id": cbq.ID,
				"text":              "text",
				"show_alert":        "true",
			},
		},
		{
			Name:           "AnswerURL",
			Request:        cbq.AnswerURL("https://t.me/bot?start=123").Request(),
			ExceptedMethod: "answerCallbackQuery",
			ExpectedArgs: map[string]string{
				"callback_query_id": cbq.ID,
				"url":               "https://t.me/bot?start=123",
			},
		},
	} {
		t.Run(test.Name, func(t *testing.T) {
			encoder := EncoderCollect{
				Args:  make(map[string]string),
				Files: make(map[string]tg.InputFile),
			}

			err := test.Request.Encode(&encoder)
			assert.NoError(t, err)

			assert.Equal(t, test.ExceptedMethod, test.Request.Method)

			assert.Equal(t, test.ExpectedArgs, encoder.Args)

			for k, v := range test.ExpectedFiles {
				assert.Equal(t, v, encoder.Files[k])
			}
		})
	}
}

func TestInlineQueryUpdateHelpers(t *testing.T) {
	iq := InlineQueryUpdate{
		InlineQuery: &tg.InlineQuery{
			ID: "1028493893",
		},
	}

	for _, test := range []struct {
		Name           string
		Request        *tg.Request
		ExceptedMethod string
		ExpectedArgs   map[string]string
		ExpectedFiles  map[string]tg.InputFile
	}{
		{
			Name:           "Answer",
			Request:        iq.Answer([]tg.InlineQueryResult{}).Request(),
			ExceptedMethod: "answerInlineQuery",
			ExpectedArgs: map[string]string{
				"inline_query_id": iq.ID,
				"results":         "[]",
			},
		},
	} {
		t.Run(test.Name, func(t *testing.T) {
			encoder := EncoderCollect{
				Args:  make(map[string]string),
				Files: make(map[string]tg.InputFile),
			}

			err := test.Request.Encode(&encoder)
			assert.NoError(t, err)

			assert.Equal(t, test.ExceptedMethod, test.Request.Method)

			assert.Equal(t, test.ExpectedArgs, encoder.Args)

			for k, v := range test.ExpectedFiles {
				assert.Equal(t, v, encoder.Files[k])
			}
		})
	}
}

func TestShippingQueryUpdateHelpers(t *testing.T) {
	sq := ShippingQueryUpdate{
		ShippingQuery: &tg.ShippingQuery{
			ID: "1028493893",
		},
	}

	for _, test := range []struct {
		Name           string
		Request        *tg.Request
		ExceptedMethod string
		ExpectedArgs   map[string]string
		ExpectedFiles  map[string]tg.InputFile
	}{
		{
			Name:           "Answer",
			Request:        sq.Answer(true).Request(),
			ExceptedMethod: "answerShippingQuery",
			ExpectedArgs: map[string]string{
				"shipping_query_id": sq.ID,
				"ok":                "true",
			},
		},
	} {
		t.Run(test.Name, func(t *testing.T) {
			encoder := EncoderCollect{
				Args:  make(map[string]string),
				Files: make(map[string]tg.InputFile),
			}

			err := test.Request.Encode(&encoder)
			assert.NoError(t, err)

			assert.Equal(t, test.ExceptedMethod, test.Request.Method)

			assert.Equal(t, test.ExpectedArgs, encoder.Args)

			for k, v := range test.ExpectedFiles {
				assert.Equal(t, v, encoder.Files[k])
			}
		})
	}
}

func TestPreCheckoutQueryUpdateHelpers(t *testing.T) {
	pcq := PreCheckoutQueryUpdate{
		PreCheckoutQuery: &tg.PreCheckoutQuery{
			ID: "1028493893",
		},
	}

	for _, test := range []struct {
		Name           string
		Request        *tg.Request
		ExceptedMethod string
		ExpectedArgs   map[string]string
		ExpectedFiles  map[string]tg.InputFile
	}{
		{
			Name:           "Answer",
			Request:        pcq.Answer(true).Request(),
			ExceptedMethod: "answerPreCheckoutQuery",
			ExpectedArgs: map[string]string{
				"pre_checkout_query_id": pcq.ID,
				"ok":                    "true",
			},
		},
	} {
		t.Run(test.Name, func(t *testing.T) {
			encoder := EncoderCollect{
				Args:  make(map[string]string),
				Files: make(map[string]tg.InputFile),
			}

			err := test.Request.Encode(&encoder)
			assert.NoError(t, err)

			assert.Equal(t, test.ExceptedMethod, test.Request.Method)

			assert.Equal(t, test.ExpectedArgs, encoder.Args)

			for k, v := range test.ExpectedFiles {
				assert.Equal(t, v, encoder.Files[k])
			}
		})
	}
}

func TestChatJoinRequestUpdateHelpers(t *testing.T) {
	cjr := ChatJoinRequestUpdate{
		ChatJoinRequest: &tg.ChatJoinRequest{
			Chat: tg.Chat{
				ID: -12345,
			},
			From: tg.User{
				ID: 12345,
			},
		},
	}

	for _, test := range []struct {
		Name           string
		Request        *tg.Request
		ExceptedMethod string
		ExpectedArgs   map[string]string
		ExpectedFiles  map[string]tg.InputFile
	}{
		{
			Name:           "Approve",
			Request:        cjr.Approve().Request(),
			ExceptedMethod: "approveChatJoinRequest",
			ExpectedArgs: map[string]string{
				"chat_id": "-12345",
				"user_id": "12345",
			},
		},
		{
			Name:           "Decline",
			Request:        cjr.Decline().Request(),
			ExceptedMethod: "declineChatJoinRequest",
			ExpectedArgs: map[string]string{
				"chat_id": "-12345",
				"user_id": "12345",
			},
		},
	} {
		t.Run(test.Name, func(t *testing.T) {
			encoder := EncoderCollect{
				Args:  make(map[string]string),
				Files: make(map[string]tg.InputFile),
			}

			err := test.Request.Encode(&encoder)
			assert.NoError(t, err)

			assert.Equal(t, test.ExceptedMethod, test.Request.Method)

			assert.Equal(t, test.ExpectedArgs, encoder.Args)

			for k, v := range test.ExpectedFiles {
				assert.Equal(t, v, encoder.Files[k])
			}
		})
	}
}
