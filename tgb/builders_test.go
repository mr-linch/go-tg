package tgb

import (
	"testing"

	"github.com/mr-linch/go-tg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTextMessageBuilder(t *testing.T) {
	t.Run("AsSend", func(t *testing.T) {
		client := &tg.Client{}
		lpo := tg.LinkPreviewOptions{IsDisabled: true}
		entities := []tg.MessageEntity{{Type: tg.MessageEntityTypeBold, Offset: 0, Length: 4}}
		replyMarkup := tg.NewInlineKeyboardMarkup()
		pm := tg.HTML

		call := NewTextMessageCallBuilder("text").
			Client(client).
			Text("text2").
			LinkPreviewOptions(lpo).
			Entities(entities).
			ReplyMarkup(replyMarkup).
			ParseMode(pm).
			AsSend(tg.ChatID(1))

		assert.Equal(t, "sendMessage", call.Request().Method)

		arg, ok := call.Request().GetArg("chat_id")
		require.True(t, ok)
		assert.Equal(t, "1", arg)

		arg, ok = call.Request().GetArg("text")
		require.True(t, ok)
		assert.Equal(t, "text2", arg)

		jsonArg, ok := call.Request().GetJSON("link_preview_options")
		require.True(t, ok)
		assert.Equal(t, lpo, jsonArg)

		jsonArg, ok = call.Request().GetJSON("entities")
		require.True(t, ok)
		assert.Equal(t, entities, jsonArg)

		jsonArg, ok = call.Request().GetJSON("reply_markup")
		require.True(t, ok)
		assert.Equal(t, replyMarkup, jsonArg)

		arg, ok = call.Request().GetArg("parse_mode")
		require.True(t, ok)
		assert.Equal(t, "HTML", arg)
	})

	t.Run("AsEditText", func(t *testing.T) {
		client := &tg.Client{}
		lpo := tg.LinkPreviewOptions{IsDisabled: true}
		entities := []tg.MessageEntity{{Type: tg.MessageEntityTypeBold, Offset: 0, Length: 4}}
		replyMarkup := tg.NewInlineKeyboardMarkup()
		pm := tg.HTML

		call := NewTextMessageCallBuilder("text").
			Client(client).
			LinkPreviewOptions(lpo).
			Entities(entities).
			ReplyMarkup(replyMarkup).
			ParseMode(pm).
			AsEditText(tg.ChatID(1), 2)

		assert.Equal(t, "editMessageText", call.Request().Method)

		arg, ok := call.Request().GetArg("chat_id")
		require.True(t, ok)
		assert.Equal(t, "1", arg)

		arg, ok = call.Request().GetArg("text")
		require.True(t, ok)
		assert.Equal(t, "text", arg)

		arg, ok = call.Request().GetArg("message_id")
		require.True(t, ok)
		assert.Equal(t, "2", arg)

		jsonArg, ok := call.Request().GetJSON("link_preview_options")
		require.True(t, ok)
		assert.Equal(t, lpo, jsonArg)

		jsonArg, ok = call.Request().GetJSON("entities")
		require.True(t, ok)
		assert.Equal(t, entities, jsonArg)

		jsonArg, ok = call.Request().GetJSON("reply_markup")
		require.True(t, ok)
		assert.Equal(t, replyMarkup, jsonArg)

		arg, ok = call.Request().GetArg("parse_mode")
		require.True(t, ok)
		assert.Equal(t, "HTML", arg)
	})

	t.Run("AsEditTextFromCBQ", func(t *testing.T) {
		client := &tg.Client{}
		lpo := tg.LinkPreviewOptions{IsDisabled: true}
		entities := []tg.MessageEntity{{Type: tg.MessageEntityTypeBold, Offset: 0, Length: 4}}
		replyMarkup := tg.NewInlineKeyboardMarkup()
		pm := tg.HTML

		call := NewTextMessageCallBuilder("text").
			Client(client).
			LinkPreviewOptions(lpo).
			Entities(entities).
			ReplyMarkup(replyMarkup).
			ParseMode(pm).
			AsEditTextFromCBQ(
				&tg.CallbackQuery{
					Message: &tg.MaybeInaccessibleMessage{
						InaccessibleMessage: &tg.InaccessibleMessage{
							Chat:      tg.Chat{ID: 1},
							MessageID: 2,
						},
					},
				},
			)

		assert.Equal(t, "editMessageText", call.Request().Method)

		arg, ok := call.Request().GetArg("chat_id")
		require.True(t, ok)
		assert.Equal(t, "1", arg)

		arg, ok = call.Request().GetArg("text")
		require.True(t, ok)
		assert.Equal(t, "text", arg)

		arg, ok = call.Request().GetArg("message_id")
		require.True(t, ok)
		assert.Equal(t, "2", arg)

		jsonArg, ok := call.Request().GetJSON("link_preview_options")
		require.True(t, ok)
		assert.Equal(t, lpo, jsonArg)

		jsonArg, ok = call.Request().GetJSON("entities")
		require.True(t, ok)
		assert.Equal(t, entities, jsonArg)

		jsonArg, ok = call.Request().GetJSON("reply_markup")
		require.True(t, ok)
		assert.Equal(t, replyMarkup, jsonArg)

		arg, ok = call.Request().GetArg("parse_mode")
		require.True(t, ok)
		assert.Equal(t, "HTML", arg)
	})

	t.Run("AsEditTextFromMsg", func(t *testing.T) {
		client := &tg.Client{}
		lpo := tg.LinkPreviewOptions{IsDisabled: true}
		entities := []tg.MessageEntity{{Type: tg.MessageEntityTypeBold, Offset: 0, Length: 4}}
		replyMarkup := tg.NewInlineKeyboardMarkup()
		pm := tg.HTML

		call := NewTextMessageCallBuilder("text").
			Client(client).
			LinkPreviewOptions(lpo).
			Entities(entities).
			ReplyMarkup(replyMarkup).
			ParseMode(pm).
			AsEditTextFromMsg(&tg.Message{
				Chat: tg.Chat{ID: 1},
				ID:   2,
			})

		assert.Equal(t, "editMessageText", call.Request().Method)

		arg, ok := call.Request().GetArg("chat_id")
		require.True(t, ok)
		assert.Equal(t, "1", arg)

		arg, ok = call.Request().GetArg("text")
		require.True(t, ok)
		assert.Equal(t, "text", arg)

		arg, ok = call.Request().GetArg("message_id")
		require.True(t, ok)
		assert.Equal(t, "2", arg)

		jsonArg, ok := call.Request().GetJSON("link_preview_options")
		require.True(t, ok)
		assert.Equal(t, lpo, jsonArg)

		jsonArg, ok = call.Request().GetJSON("entities")
		require.True(t, ok)
		assert.Equal(t, entities, jsonArg)

		jsonArg, ok = call.Request().GetJSON("reply_markup")
		require.True(t, ok)
		assert.Equal(t, replyMarkup, jsonArg)

		arg, ok = call.Request().GetArg("parse_mode")
		require.True(t, ok)
		assert.Equal(t, "HTML", arg)
	})

	t.Run("AsEditTextInline", func(t *testing.T) {
		client := &tg.Client{}
		lpo := tg.LinkPreviewOptions{IsDisabled: true}
		entities := []tg.MessageEntity{{Type: tg.MessageEntityTypeBold, Offset: 0, Length: 4}}
		replyMarkup := tg.NewInlineKeyboardMarkup()
		pm := tg.HTML

		call := NewTextMessageCallBuilder("text").
			Client(client).
			LinkPreviewOptions(lpo).
			Entities(entities).
			ReplyMarkup(replyMarkup).
			ParseMode(pm).
			AsEditTextInline("inline")

		assert.Equal(t, "editMessageText", call.Request().Method)

		arg, ok := call.Request().GetArg("inline_message_id")
		require.True(t, ok)
		assert.Equal(t, "inline", arg)

		arg, ok = call.Request().GetArg("text")
		require.True(t, ok)
		assert.Equal(t, "text", arg)

		jsonArg, ok := call.Request().GetJSON("link_preview_options")
		require.True(t, ok)
		assert.Equal(t, lpo, jsonArg)

		jsonArg, ok = call.Request().GetJSON("entities")
		require.True(t, ok)
		assert.Equal(t, entities, jsonArg)

		jsonArg, ok = call.Request().GetJSON("reply_markup")
		require.True(t, ok)
		assert.Equal(t, replyMarkup, jsonArg)

		arg, ok = call.Request().GetArg("parse_mode")
		require.True(t, ok)
		assert.Equal(t, "HTML", arg)
	})

	t.Run("AsEditTextReplyMarkup", func(t *testing.T) {
		client := &tg.Client{}
		replyMarkup := tg.NewInlineKeyboardMarkup()

		call := NewTextMessageCallBuilder("text").
			Client(client).
			ReplyMarkup(replyMarkup).
			AsEditReplyMarkup(tg.ChatID(1), 2)

		assert.Equal(t, "editMessageReplyMarkup", call.Request().Method)

		jsonArg, ok := call.Request().GetJSON("reply_markup")
		require.True(t, ok)
		assert.Equal(t, replyMarkup, jsonArg)

		arg, ok := call.Request().GetArg("chat_id")
		require.True(t, ok)
		assert.Equal(t, "1", arg)

		arg, ok = call.Request().GetArg("message_id")
		require.True(t, ok)
		assert.Equal(t, "2", arg)
	})

	t.Run("AsEditTextReplyMarkupFromCBQ", func(t *testing.T) {
		client := &tg.Client{}
		replyMarkup := tg.NewInlineKeyboardMarkup()

		call := NewTextMessageCallBuilder("text").
			Client(client).
			ReplyMarkup(replyMarkup).
			AsEditReplyMarkupFromCBQ(
				&tg.CallbackQuery{
					Message: &tg.MaybeInaccessibleMessage{
						InaccessibleMessage: &tg.InaccessibleMessage{
							Chat:      tg.Chat{ID: 1},
							MessageID: 2,
						},
					},
				},
			)

		assert.Equal(t, "editMessageReplyMarkup", call.Request().Method)

		jsonArg, ok := call.Request().GetJSON("reply_markup")
		require.True(t, ok)
		assert.Equal(t, replyMarkup, jsonArg)

		arg, ok := call.Request().GetArg("chat_id")
		require.True(t, ok)
		assert.Equal(t, "1", arg)

		arg, ok = call.Request().GetArg("message_id")
		require.True(t, ok)
		assert.Equal(t, "2", arg)
	})

	t.Run("AsEditTextReplyMarkupFromMsg", func(t *testing.T) {
		client := &tg.Client{}
		replyMarkup := tg.NewInlineKeyboardMarkup()

		call := NewTextMessageCallBuilder("text").
			Client(client).
			ReplyMarkup(replyMarkup).
			AsEditReplyMarkupFromMsg(tg.Message{
				Chat: tg.Chat{ID: 1},
				ID:   2,
			})

		assert.Equal(t, "editMessageReplyMarkup", call.Request().Method)

		jsonArg, ok := call.Request().GetJSON("reply_markup")
		require.True(t, ok)
		assert.Equal(t, replyMarkup, jsonArg)

		arg, ok := call.Request().GetArg("chat_id")
		require.True(t, ok)
		assert.Equal(t, "1", arg)

		arg, ok = call.Request().GetArg("message_id")
		require.True(t, ok)
		assert.Equal(t, "2", arg)
	})

	t.Run("AsEditReplyMarkupInline", func(t *testing.T) {
		client := &tg.Client{}
		replyMarkup := tg.NewInlineKeyboardMarkup()

		call := NewTextMessageCallBuilder("text").
			Client(client).
			ReplyMarkup(replyMarkup).
			AsEditReplyMarkupInline("inline")

		assert.Equal(t, "editMessageReplyMarkup", call.Request().Method)

		jsonArg, ok := call.Request().GetJSON("reply_markup")
		require.True(t, ok)
		assert.Equal(t, replyMarkup, jsonArg)

		arg, ok := call.Request().GetArg("inline_message_id")
		require.True(t, ok)
		assert.Equal(t, "inline", arg)
	})
}
