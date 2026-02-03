package tgb

import tg "github.com/mr-linch/go-tg"

// TextMessageCallBuilder contains all common fields for methods sendText, editMessageText, editMessageReplyMarkup.
// It's useful for building different calls with the same params.
//
// Example:
//
//	newMenuBuilder(...).Client(client).AsSend(chat)
//	newMenuBuilder(...).Client(client).AsEditTextFromCBQ(cbq.CallbackQuery)
//
// Also can be sended as reply to webhook
//
//	msg.Update.Reply(ctx, newMenuBuilder(...).AsEditTextFromCBQ(cbq.CallbackQuery))
type TextMessageCallBuilder struct {
	text               string
	replyMarkup        tg.ReplyMarkup
	linkPreviewOptions *tg.LinkPreviewOptions
	entities           []tg.MessageEntity
	parseMode          tg.ParseMode
	client             *tg.Client
}

// NewTextMessageCallBuilder creates new TextMessageCallBuilder with specified text.
func NewTextMessageCallBuilder(text string) *TextMessageCallBuilder {
	return &TextMessageCallBuilder{
		text: text,
	}
}

// Client sets client for the message.
func (b *TextMessageCallBuilder) Client(client *tg.Client) *TextMessageCallBuilder {
	b.client = client
	return b
}

// Text sets text for the message.
func (b *TextMessageCallBuilder) Text(text string) *TextMessageCallBuilder {
	b.text = text
	return b
}

// ReplyMarkup sets reply markup for the message.
func (b *TextMessageCallBuilder) ReplyMarkup(markup tg.ReplyMarkup) *TextMessageCallBuilder {
	b.replyMarkup = markup
	return b
}

// LinkPreviewOptions sets link preview options for the message.
func (b *TextMessageCallBuilder) LinkPreviewOptions(options tg.LinkPreviewOptions) *TextMessageCallBuilder {
	b.linkPreviewOptions = &options
	return b
}

// Entities sets entities for the message.
func (b *TextMessageCallBuilder) Entities(entities []tg.MessageEntity) *TextMessageCallBuilder {
	b.entities = entities
	return b
}

// ParseMode sets parse mode for the message.
func (b *TextMessageCallBuilder) ParseMode(mode tg.ParseMode) *TextMessageCallBuilder {
	b.parseMode = mode
	return b
}

// AsSend returns call sendMessage with specified peer.
func (b *TextMessageCallBuilder) AsSend(peer tg.PeerID) *tg.SendMessageCall {
	call := tg.NewSendMessageCall(peer, b.text)

	if b.replyMarkup != nil {
		call.ReplyMarkup(b.replyMarkup)
	}

	if b.linkPreviewOptions != nil {
		call.LinkPreviewOptions(*b.linkPreviewOptions)
	}

	if b.entities != nil {
		call.Entities(b.entities)
	}

	if b.parseMode != nil {
		call.ParseMode(b.parseMode)
	}

	if b.client != nil {
		call.Bind(b.client)
	}

	return call
}

// AsEditText returns call editTextMessage with prepopulated fields.
func (b *TextMessageCallBuilder) AsEditText(peer tg.PeerID, id int) *tg.EditMessageTextCall {
	call := tg.NewEditMessageTextCall(peer, id, b.text)

	if b.replyMarkup != nil {
		if v, ok := b.replyMarkup.(tg.InlineKeyboardMarkup); ok {
			call.ReplyMarkup(v)
		}
	}

	if b.linkPreviewOptions != nil {
		call.LinkPreviewOptions(*b.linkPreviewOptions)
	}

	if b.entities != nil {
		call.Entities(b.entities)
	}

	if b.parseMode != nil {
		call.ParseMode(b.parseMode)
	}

	if b.client != nil {
		call.Bind(b.client)
	}

	return call
}

// AsEditTextFromCBQ wraps AsEditText with callback as argument.
// It's useful if you have an object of CallbackQuery and want to edit it.
func (b *TextMessageCallBuilder) AsEditTextFromCBQ(callback *tg.CallbackQuery) *tg.EditMessageTextCall {
	return b.AsEditText(callback.Message.Chat(), callback.Message.MessageID())
}

// AsEditTextFromMsg wraps AsEditText with message as argument.
// It's useful if you have an object of Message and want to edit it.
func (b *TextMessageCallBuilder) AsEditTextFromMsg(msg *tg.Message) *tg.EditMessageTextCall {
	return b.AsEditText(msg.Chat, msg.ID)
}

// AsEditTextInline returns call editTextMessage with by inline message id.
func (b *TextMessageCallBuilder) AsEditTextInline(id string) *tg.EditMessageTextCall {
	call := tg.NewEditMessageTextInlineCall(id, b.text)

	if b.replyMarkup != nil {
		if v, ok := b.replyMarkup.(tg.InlineKeyboardMarkup); ok {
			call.ReplyMarkup(v)
		}
	}

	if b.linkPreviewOptions != nil {
		call.LinkPreviewOptions(*b.linkPreviewOptions)
	}

	if b.entities != nil {
		call.Entities(b.entities)
	}

	if b.parseMode != nil {
		call.ParseMode(b.parseMode)
	}

	if b.client != nil {
		call.Bind(b.client)
	}

	return call
}

// AsEditReplyMarkup returns call editReplyMarkup with prepopulated fields.
func (b *TextMessageCallBuilder) AsEditReplyMarkup(peer tg.PeerID, id int) *tg.EditMessageReplyMarkupCall {
	call := tg.NewEditMessageReplyMarkupCall(peer, id)

	if v, ok := b.replyMarkup.(tg.InlineKeyboardMarkup); ok {
		call.ReplyMarkup(v)
	}

	if b.client != nil {
		call.Bind(b.client)
	}

	return call
}

// AsEditReplyMarkupFromCBQ wraps AsEditReplyMarkup with callback as argument.
// It's useful if you have an object of CallbackQuery and want to edit it.
func (b *TextMessageCallBuilder) AsEditReplyMarkupFromCBQ(callback *tg.CallbackQuery) *tg.EditMessageReplyMarkupCall {
	return b.AsEditReplyMarkup(callback.Message.Chat(), callback.Message.MessageID())
}

// AsEditReplyMarkupFromMsg wraps AsEditReplyMarkup with message as argument.
// It's useful if you have an object of Message and want to edit it.
func (b *TextMessageCallBuilder) AsEditReplyMarkupFromMsg(msg tg.Message) *tg.EditMessageReplyMarkupCall {
	return b.AsEditReplyMarkup(msg.Chat, msg.ID)
}

// AsEditReplyMarkupInline returns call editReplyMarkup with by inline message id.
func (b *TextMessageCallBuilder) AsEditReplyMarkupInline(id string) *tg.EditMessageReplyMarkupCall {
	call := tg.NewEditMessageReplyMarkupInlineCall(id)

	if v, ok := b.replyMarkup.(tg.InlineKeyboardMarkup); ok {
		call.ReplyMarkup(v)
	}

	if b.client != nil {
		call.Bind(b.client)
	}

	return call
}
