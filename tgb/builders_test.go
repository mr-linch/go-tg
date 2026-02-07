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
			BusinessConnectionID("biz123").
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

		arg, ok = call.Request().GetArg("business_connection_id")
		require.True(t, ok)
		assert.Equal(t, "biz123", arg)
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
			BusinessConnectionID("biz123").
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

		arg, ok = call.Request().GetArg("business_connection_id")
		require.True(t, ok)
		assert.Equal(t, "biz123", arg)
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
			BusinessConnectionID("biz123").
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

		arg, ok = call.Request().GetArg("business_connection_id")
		require.True(t, ok)
		assert.Equal(t, "biz123", arg)
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
			BusinessConnectionID("biz123").
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

		arg, ok = call.Request().GetArg("business_connection_id")
		require.True(t, ok)
		assert.Equal(t, "biz123", arg)
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
			BusinessConnectionID("biz123").
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

		arg, ok = call.Request().GetArg("business_connection_id")
		require.True(t, ok)
		assert.Equal(t, "biz123", arg)
	})

	t.Run("AsEditTextReplyMarkup", func(t *testing.T) {
		client := &tg.Client{}
		replyMarkup := tg.NewInlineKeyboardMarkup()

		call := NewTextMessageCallBuilder("text").
			Client(client).
			ReplyMarkup(replyMarkup).
			BusinessConnectionID("biz123").
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

		arg, ok = call.Request().GetArg("business_connection_id")
		require.True(t, ok)
		assert.Equal(t, "biz123", arg)
	})

	t.Run("AsEditTextReplyMarkupFromCBQ", func(t *testing.T) {
		client := &tg.Client{}
		replyMarkup := tg.NewInlineKeyboardMarkup()

		call := NewTextMessageCallBuilder("text").
			Client(client).
			ReplyMarkup(replyMarkup).
			BusinessConnectionID("biz123").
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

		arg, ok = call.Request().GetArg("business_connection_id")
		require.True(t, ok)
		assert.Equal(t, "biz123", arg)
	})

	t.Run("AsEditTextReplyMarkupFromMsg", func(t *testing.T) {
		client := &tg.Client{}
		replyMarkup := tg.NewInlineKeyboardMarkup()

		call := NewTextMessageCallBuilder("text").
			Client(client).
			ReplyMarkup(replyMarkup).
			BusinessConnectionID("biz123").
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

		arg, ok = call.Request().GetArg("business_connection_id")
		require.True(t, ok)
		assert.Equal(t, "biz123", arg)
	})

	t.Run("AsEditReplyMarkupInline", func(t *testing.T) {
		client := &tg.Client{}
		replyMarkup := tg.NewInlineKeyboardMarkup()

		call := NewTextMessageCallBuilder("text").
			Client(client).
			ReplyMarkup(replyMarkup).
			BusinessConnectionID("biz123").
			AsEditReplyMarkupInline("inline")

		assert.Equal(t, "editMessageReplyMarkup", call.Request().Method)

		jsonArg, ok := call.Request().GetJSON("reply_markup")
		require.True(t, ok)
		assert.Equal(t, replyMarkup, jsonArg)

		arg, ok := call.Request().GetArg("inline_message_id")
		require.True(t, ok)
		assert.Equal(t, "inline", arg)

		arg, ok = call.Request().GetArg("business_connection_id")
		require.True(t, ok)
		assert.Equal(t, "biz123", arg)
	})
}

func newTestMediaBuilder() *MediaMessageCallBuilder {
	return NewMediaMessageCallBuilder("caption").
		Client(&tg.Client{}).
		ParseMode(tg.HTML).
		CaptionEntities([]tg.MessageEntity{{Type: tg.MessageEntityTypeBold, Offset: 0, Length: 7}}).
		ShowCaptionAboveMedia(true).
		ReplyMarkup(tg.NewInlineKeyboardMarkup()).
		BusinessConnectionID("biz123")
}

func assertMediaCommonFields(t *testing.T, call interface {
	Request() *tg.Request
}, method string, hasShowCaptionAboveMedia bool,
) {
	t.Helper()

	assert.Equal(t, method, call.Request().Method)

	arg, ok := call.Request().GetArg("caption")
	require.True(t, ok)
	assert.Equal(t, "caption", arg)

	arg, ok = call.Request().GetArg("parse_mode")
	require.True(t, ok)
	assert.Equal(t, "HTML", arg)

	jsonArg, ok := call.Request().GetJSON("caption_entities")
	require.True(t, ok)
	assert.Equal(t, []tg.MessageEntity{{Type: tg.MessageEntityTypeBold, Offset: 0, Length: 7}}, jsonArg)

	if hasShowCaptionAboveMedia {
		arg, ok = call.Request().GetArg("show_caption_above_media")
		require.True(t, ok)
		assert.Equal(t, "true", arg)
	} else {
		_, ok = call.Request().GetArg("show_caption_above_media")
		assert.False(t, ok)
	}

	jsonArg, ok = call.Request().GetJSON("reply_markup")
	require.True(t, ok)
	assert.Equal(t, tg.NewInlineKeyboardMarkup(), jsonArg)

	arg, ok = call.Request().GetArg("business_connection_id")
	require.True(t, ok)
	assert.Equal(t, "biz123", arg)
}

func TestMediaMessageBuilder(t *testing.T) {
	t.Run("AsSendPhoto", func(t *testing.T) {
		call := newTestMediaBuilder().AsSendPhoto(tg.ChatID(1), tg.NewFileArgURL("https://example.com/photo.jpg"))
		assertMediaCommonFields(t, call, "sendPhoto", true)
	})

	t.Run("AsSendVideo", func(t *testing.T) {
		call := newTestMediaBuilder().AsSendVideo(tg.ChatID(1), tg.NewFileArgURL("https://example.com/video.mp4"))
		assertMediaCommonFields(t, call, "sendVideo", true)
	})

	t.Run("AsSendAudio", func(t *testing.T) {
		call := newTestMediaBuilder().AsSendAudio(tg.ChatID(1), tg.NewFileArgURL("https://example.com/audio.mp3"))
		assertMediaCommonFields(t, call, "sendAudio", false)
	})

	t.Run("AsSendDocument", func(t *testing.T) {
		call := newTestMediaBuilder().AsSendDocument(tg.ChatID(1), tg.NewFileArgURL("https://example.com/doc.pdf"))
		assertMediaCommonFields(t, call, "sendDocument", false)
	})

	t.Run("AsSendAnimation", func(t *testing.T) {
		call := newTestMediaBuilder().AsSendAnimation(tg.ChatID(1), tg.NewFileArgURL("https://example.com/anim.gif"))
		assertMediaCommonFields(t, call, "sendAnimation", true)
	})

	t.Run("AsSendVoice", func(t *testing.T) {
		call := newTestMediaBuilder().AsSendVoice(tg.ChatID(1), tg.NewFileArgURL("https://example.com/voice.ogg"))
		assertMediaCommonFields(t, call, "sendVoice", false)
	})

	t.Run("AsEditCaption", func(t *testing.T) {
		call := newTestMediaBuilder().AsEditCaption(tg.ChatID(1), 2)
		assertMediaCommonFields(t, call, "editMessageCaption", true)

		arg, ok := call.Request().GetArg("chat_id")
		require.True(t, ok)
		assert.Equal(t, "1", arg)

		arg, ok = call.Request().GetArg("message_id")
		require.True(t, ok)
		assert.Equal(t, "2", arg)
	})

	t.Run("AsEditCaptionFromCBQ", func(t *testing.T) {
		call := newTestMediaBuilder().AsEditCaptionFromCBQ(
			&tg.CallbackQuery{
				Message: &tg.MaybeInaccessibleMessage{
					InaccessibleMessage: &tg.InaccessibleMessage{
						Chat:      tg.Chat{ID: 1},
						MessageID: 2,
					},
				},
			},
		)
		assertMediaCommonFields(t, call, "editMessageCaption", true)

		arg, ok := call.Request().GetArg("chat_id")
		require.True(t, ok)
		assert.Equal(t, "1", arg)

		arg, ok = call.Request().GetArg("message_id")
		require.True(t, ok)
		assert.Equal(t, "2", arg)
	})

	t.Run("AsEditCaptionFromMsg", func(t *testing.T) {
		call := newTestMediaBuilder().AsEditCaptionFromMsg(&tg.Message{
			Chat: tg.Chat{ID: 1},
			ID:   2,
		})
		assertMediaCommonFields(t, call, "editMessageCaption", true)

		arg, ok := call.Request().GetArg("chat_id")
		require.True(t, ok)
		assert.Equal(t, "1", arg)

		arg, ok = call.Request().GetArg("message_id")
		require.True(t, ok)
		assert.Equal(t, "2", arg)
	})

	t.Run("AsEditCaptionInline", func(t *testing.T) {
		call := newTestMediaBuilder().AsEditCaptionInline("inline")
		assertMediaCommonFields(t, call, "editMessageCaption", true)

		arg, ok := call.Request().GetArg("inline_message_id")
		require.True(t, ok)
		assert.Equal(t, "inline", arg)
	})

	t.Run("InputMediaPhoto", func(t *testing.T) {
		b := newTestMediaBuilder()
		media := b.NewInputMediaPhoto(tg.NewFileArgURL("https://example.com/photo.jpg"))

		require.NotNil(t, media.Photo)
		assert.Equal(t, "caption", media.Photo.Caption)
		assert.Equal(t, "HTML", media.Photo.ParseMode.String())
		assert.True(t, media.Photo.ShowCaptionAboveMedia)
	})

	t.Run("InputMediaVideo", func(t *testing.T) {
		b := newTestMediaBuilder()
		media := b.NewInputMediaVideo(tg.NewFileArgURL("https://example.com/video.mp4"))

		require.NotNil(t, media.Video)
		assert.Equal(t, "caption", media.Video.Caption)
		assert.True(t, media.Video.ShowCaptionAboveMedia)
	})

	t.Run("InputMediaAudio", func(t *testing.T) {
		b := newTestMediaBuilder()
		media := b.NewInputMediaAudio(tg.NewFileArgURL("https://example.com/audio.mp3"))

		require.NotNil(t, media.Audio)
		assert.Equal(t, "caption", media.Audio.Caption)
	})

	t.Run("InputMediaDocument", func(t *testing.T) {
		b := newTestMediaBuilder()
		media := b.NewInputMediaDocument(tg.NewFileArgURL("https://example.com/doc.pdf"))

		require.NotNil(t, media.Document)
		assert.Equal(t, "caption", media.Document.Caption)
	})

	t.Run("InputMediaAnimation", func(t *testing.T) {
		b := newTestMediaBuilder()
		media := b.NewInputMediaAnimation(tg.NewFileArgURL("https://example.com/anim.gif"))

		require.NotNil(t, media.Animation)
		assert.Equal(t, "caption", media.Animation.Caption)
		assert.True(t, media.Animation.ShowCaptionAboveMedia)
	})

	t.Run("AsEditMedia", func(t *testing.T) {
		b := newTestMediaBuilder()
		media := b.NewInputMediaPhoto(tg.NewFileArgURL("https://example.com/photo.jpg"))
		call := b.AsEditMedia(tg.ChatID(1), 2, media)

		assert.Equal(t, "editMessageMedia", call.Request().Method)

		arg, ok := call.Request().GetArg("chat_id")
		require.True(t, ok)
		assert.Equal(t, "1", arg)

		arg, ok = call.Request().GetArg("message_id")
		require.True(t, ok)
		assert.Equal(t, "2", arg)

		jsonArg, ok := call.Request().GetJSON("reply_markup")
		require.True(t, ok)
		assert.Equal(t, tg.NewInlineKeyboardMarkup(), jsonArg)

		arg, ok = call.Request().GetArg("business_connection_id")
		require.True(t, ok)
		assert.Equal(t, "biz123", arg)
	})

	t.Run("AsEditMediaFromCBQ", func(t *testing.T) {
		b := newTestMediaBuilder()
		media := b.NewInputMediaPhoto(tg.NewFileArgURL("https://example.com/photo.jpg"))
		call := b.AsEditMediaFromCBQ(
			&tg.CallbackQuery{
				Message: &tg.MaybeInaccessibleMessage{
					InaccessibleMessage: &tg.InaccessibleMessage{
						Chat:      tg.Chat{ID: 1},
						MessageID: 2,
					},
				},
			},
			media,
		)

		assert.Equal(t, "editMessageMedia", call.Request().Method)

		arg, ok := call.Request().GetArg("chat_id")
		require.True(t, ok)
		assert.Equal(t, "1", arg)
	})

	t.Run("AsEditMediaFromMsg", func(t *testing.T) {
		b := newTestMediaBuilder()
		media := b.NewInputMediaPhoto(tg.NewFileArgURL("https://example.com/photo.jpg"))
		call := b.AsEditMediaFromMsg(
			&tg.Message{Chat: tg.Chat{ID: 1}, ID: 2},
			media,
		)

		assert.Equal(t, "editMessageMedia", call.Request().Method)

		arg, ok := call.Request().GetArg("chat_id")
		require.True(t, ok)
		assert.Equal(t, "1", arg)
	})

	t.Run("AsEditMediaInline", func(t *testing.T) {
		b := newTestMediaBuilder()
		media := b.NewInputMediaPhoto(tg.NewFileArgURL("https://example.com/photo.jpg"))
		call := b.AsEditMediaInline("inline", media)

		assert.Equal(t, "editMessageMedia", call.Request().Method)

		arg, ok := call.Request().GetArg("inline_message_id")
		require.True(t, ok)
		assert.Equal(t, "inline", arg)
	})
}
