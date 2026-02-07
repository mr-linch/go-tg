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
	text                 string
	replyMarkup          tg.ReplyMarkup
	linkPreviewOptions   *tg.LinkPreviewOptions
	entities             []tg.MessageEntity
	parseMode            tg.ParseMode
	client               *tg.Client
	businessConnectionID string
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

// BusinessConnectionID sets business connection ID for the message.
func (b *TextMessageCallBuilder) BusinessConnectionID(id string) *TextMessageCallBuilder {
	b.businessConnectionID = id
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

	if b.businessConnectionID != "" {
		call.BusinessConnectionID(b.businessConnectionID)
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

	if b.businessConnectionID != "" {
		call.BusinessConnectionID(b.businessConnectionID)
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

	if b.businessConnectionID != "" {
		call.BusinessConnectionID(b.businessConnectionID)
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

	if b.businessConnectionID != "" {
		call.BusinessConnectionID(b.businessConnectionID)
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

	if b.businessConnectionID != "" {
		call.BusinessConnectionID(b.businessConnectionID)
	}

	if b.client != nil {
		call.Bind(b.client)
	}

	return call
}

// MediaMessageCallBuilder contains common fields for caption-based media methods:
// sendPhoto, sendVideo, sendAudio, sendDocument, sendAnimation, sendVoice, and editMessageCaption.
// It's useful for building different calls with the same caption params.
type MediaMessageCallBuilder struct {
	caption               string
	parseMode             tg.ParseMode
	captionEntities       []tg.MessageEntity
	showCaptionAboveMedia bool
	replyMarkup           tg.ReplyMarkup
	businessConnectionID  string
	client                *tg.Client
}

// NewMediaMessageCallBuilder creates new MediaMessageCallBuilder with specified caption.
func NewMediaMessageCallBuilder(caption string) *MediaMessageCallBuilder {
	return &MediaMessageCallBuilder{
		caption: caption,
	}
}

// Client sets client for the message.
func (b *MediaMessageCallBuilder) Client(client *tg.Client) *MediaMessageCallBuilder {
	b.client = client
	return b
}

// Caption sets caption for the message.
func (b *MediaMessageCallBuilder) Caption(caption string) *MediaMessageCallBuilder {
	b.caption = caption
	return b
}

// ParseMode sets parse mode for the caption.
func (b *MediaMessageCallBuilder) ParseMode(mode tg.ParseMode) *MediaMessageCallBuilder {
	b.parseMode = mode
	return b
}

// CaptionEntities sets entities for the caption.
func (b *MediaMessageCallBuilder) CaptionEntities(entities []tg.MessageEntity) *MediaMessageCallBuilder {
	b.captionEntities = entities
	return b
}

// ShowCaptionAboveMedia sets whether to show caption above media.
// Only applied to photo, video, animation, and editCaption calls.
func (b *MediaMessageCallBuilder) ShowCaptionAboveMedia(show bool) *MediaMessageCallBuilder {
	b.showCaptionAboveMedia = show
	return b
}

// ReplyMarkup sets reply markup for the message.
func (b *MediaMessageCallBuilder) ReplyMarkup(markup tg.ReplyMarkup) *MediaMessageCallBuilder {
	b.replyMarkup = markup
	return b
}

// BusinessConnectionID sets business connection ID for the message.
func (b *MediaMessageCallBuilder) BusinessConnectionID(id string) *MediaMessageCallBuilder {
	b.businessConnectionID = id
	return b
}

// AsSendPhoto returns call sendPhoto with specified peer and photo.
func (b *MediaMessageCallBuilder) AsSendPhoto(peer tg.PeerID, photo tg.FileArg) *tg.SendPhotoCall {
	call := tg.NewSendPhotoCall(peer, photo)

	if b.caption != "" {
		call.Caption(b.caption)
	}

	if b.parseMode != nil {
		call.ParseMode(b.parseMode)
	}

	if b.captionEntities != nil {
		call.CaptionEntities(b.captionEntities)
	}

	if b.showCaptionAboveMedia {
		call.ShowCaptionAboveMedia(b.showCaptionAboveMedia)
	}

	if b.replyMarkup != nil {
		call.ReplyMarkup(b.replyMarkup)
	}

	if b.businessConnectionID != "" {
		call.BusinessConnectionID(b.businessConnectionID)
	}

	if b.client != nil {
		call.Bind(b.client)
	}

	return call
}

// AsSendVideo returns call sendVideo with specified peer and video.
func (b *MediaMessageCallBuilder) AsSendVideo(peer tg.PeerID, video tg.FileArg) *tg.SendVideoCall {
	call := tg.NewSendVideoCall(peer, video)

	if b.caption != "" {
		call.Caption(b.caption)
	}

	if b.parseMode != nil {
		call.ParseMode(b.parseMode)
	}

	if b.captionEntities != nil {
		call.CaptionEntities(b.captionEntities)
	}

	if b.showCaptionAboveMedia {
		call.ShowCaptionAboveMedia(b.showCaptionAboveMedia)
	}

	if b.replyMarkup != nil {
		call.ReplyMarkup(b.replyMarkup)
	}

	if b.businessConnectionID != "" {
		call.BusinessConnectionID(b.businessConnectionID)
	}

	if b.client != nil {
		call.Bind(b.client)
	}

	return call
}

// AsSendAudio returns call sendAudio with specified peer and audio.
func (b *MediaMessageCallBuilder) AsSendAudio(peer tg.PeerID, audio tg.FileArg) *tg.SendAudioCall {
	call := tg.NewSendAudioCall(peer, audio)

	if b.caption != "" {
		call.Caption(b.caption)
	}

	if b.parseMode != nil {
		call.ParseMode(b.parseMode)
	}

	if b.captionEntities != nil {
		call.CaptionEntities(b.captionEntities)
	}

	if b.replyMarkup != nil {
		call.ReplyMarkup(b.replyMarkup)
	}

	if b.businessConnectionID != "" {
		call.BusinessConnectionID(b.businessConnectionID)
	}

	if b.client != nil {
		call.Bind(b.client)
	}

	return call
}

// AsSendDocument returns call sendDocument with specified peer and document.
func (b *MediaMessageCallBuilder) AsSendDocument(peer tg.PeerID, document tg.FileArg) *tg.SendDocumentCall {
	call := tg.NewSendDocumentCall(peer, document)

	if b.caption != "" {
		call.Caption(b.caption)
	}

	if b.parseMode != nil {
		call.ParseMode(b.parseMode)
	}

	if b.captionEntities != nil {
		call.CaptionEntities(b.captionEntities)
	}

	if b.replyMarkup != nil {
		call.ReplyMarkup(b.replyMarkup)
	}

	if b.businessConnectionID != "" {
		call.BusinessConnectionID(b.businessConnectionID)
	}

	if b.client != nil {
		call.Bind(b.client)
	}

	return call
}

// AsSendAnimation returns call sendAnimation with specified peer and animation.
func (b *MediaMessageCallBuilder) AsSendAnimation(peer tg.PeerID, animation tg.FileArg) *tg.SendAnimationCall {
	call := tg.NewSendAnimationCall(peer, animation)

	if b.caption != "" {
		call.Caption(b.caption)
	}

	if b.parseMode != nil {
		call.ParseMode(b.parseMode)
	}

	if b.captionEntities != nil {
		call.CaptionEntities(b.captionEntities)
	}

	if b.showCaptionAboveMedia {
		call.ShowCaptionAboveMedia(b.showCaptionAboveMedia)
	}

	if b.replyMarkup != nil {
		call.ReplyMarkup(b.replyMarkup)
	}

	if b.businessConnectionID != "" {
		call.BusinessConnectionID(b.businessConnectionID)
	}

	if b.client != nil {
		call.Bind(b.client)
	}

	return call
}

// AsSendVoice returns call sendVoice with specified peer and voice.
func (b *MediaMessageCallBuilder) AsSendVoice(peer tg.PeerID, voice tg.FileArg) *tg.SendVoiceCall {
	call := tg.NewSendVoiceCall(peer, voice)

	if b.caption != "" {
		call.Caption(b.caption)
	}

	if b.parseMode != nil {
		call.ParseMode(b.parseMode)
	}

	if b.captionEntities != nil {
		call.CaptionEntities(b.captionEntities)
	}

	if b.replyMarkup != nil {
		call.ReplyMarkup(b.replyMarkup)
	}

	if b.businessConnectionID != "" {
		call.BusinessConnectionID(b.businessConnectionID)
	}

	if b.client != nil {
		call.Bind(b.client)
	}

	return call
}

// AsEditCaption returns call editMessageCaption with prepopulated fields.
func (b *MediaMessageCallBuilder) AsEditCaption(peer tg.PeerID, id int) *tg.EditMessageCaptionCall {
	call := tg.NewEditMessageCaptionCall(peer, id, b.caption)

	if b.parseMode != nil {
		call.ParseMode(b.parseMode)
	}

	if b.captionEntities != nil {
		call.CaptionEntities(b.captionEntities)
	}

	if b.showCaptionAboveMedia {
		call.ShowCaptionAboveMedia(b.showCaptionAboveMedia)
	}

	if b.replyMarkup != nil {
		if v, ok := b.replyMarkup.(tg.InlineKeyboardMarkup); ok {
			call.ReplyMarkup(v)
		}
	}

	if b.businessConnectionID != "" {
		call.BusinessConnectionID(b.businessConnectionID)
	}

	if b.client != nil {
		call.Bind(b.client)
	}

	return call
}

// AsEditCaptionFromCBQ wraps AsEditCaption with callback as argument.
func (b *MediaMessageCallBuilder) AsEditCaptionFromCBQ(callback *tg.CallbackQuery) *tg.EditMessageCaptionCall {
	return b.AsEditCaption(callback.Message.Chat(), callback.Message.MessageID())
}

// AsEditCaptionFromMsg wraps AsEditCaption with message as argument.
func (b *MediaMessageCallBuilder) AsEditCaptionFromMsg(msg *tg.Message) *tg.EditMessageCaptionCall {
	return b.AsEditCaption(msg.Chat, msg.ID)
}

// InputMediaPhoto returns InputMedia photo with caption fields from the builder.
func (b *MediaMessageCallBuilder) NewInputMediaPhoto(photo tg.FileArg) tg.InputMedia {
	return tg.InputMedia{Photo: &tg.InputMediaPhoto{
		Media:                 photo,
		Caption:               b.caption,
		ParseMode:             b.parseMode,
		CaptionEntities:       b.captionEntities,
		ShowCaptionAboveMedia: b.showCaptionAboveMedia,
	}}
}

// InputMediaVideo returns InputMedia video with caption fields from the builder.
func (b *MediaMessageCallBuilder) NewInputMediaVideo(video tg.FileArg) tg.InputMedia {
	return tg.InputMedia{Video: &tg.InputMediaVideo{
		Media:                 video,
		Caption:               b.caption,
		ParseMode:             b.parseMode,
		CaptionEntities:       b.captionEntities,
		ShowCaptionAboveMedia: b.showCaptionAboveMedia,
	}}
}

// InputMediaAnimation returns InputMedia animation with caption fields from the builder.
func (b *MediaMessageCallBuilder) NewInputMediaAnimation(animation tg.FileArg) tg.InputMedia {
	return tg.InputMedia{Animation: &tg.InputMediaAnimation{
		Media:                 animation,
		Caption:               b.caption,
		ParseMode:             b.parseMode,
		CaptionEntities:       b.captionEntities,
		ShowCaptionAboveMedia: b.showCaptionAboveMedia,
	}}
}

// InputMediaAudio returns InputMedia audio with caption fields from the builder.
func (b *MediaMessageCallBuilder) NewInputMediaAudio(audio tg.FileArg) tg.InputMedia {
	return tg.InputMedia{Audio: &tg.InputMediaAudio{
		Media:           audio,
		Caption:         b.caption,
		ParseMode:       b.parseMode,
		CaptionEntities: b.captionEntities,
	}}
}

// InputMediaDocument returns InputMedia document with caption fields from the builder.
func (b *MediaMessageCallBuilder) NewInputMediaDocument(document tg.FileArg) tg.InputMedia {
	return tg.InputMedia{Document: &tg.InputMediaDocument{
		Media:           document,
		Caption:         b.caption,
		ParseMode:       b.parseMode,
		CaptionEntities: b.captionEntities,
	}}
}

// AsEditMedia returns call editMessageMedia with prepopulated fields.
func (b *MediaMessageCallBuilder) AsEditMedia(peer tg.PeerID, id int, media tg.InputMedia) *tg.EditMessageMediaCall {
	return b.applyEditMedia(tg.NewEditMessageMediaCall(peer, id, media))
}

// AsEditMediaFromCBQ wraps AsEditMedia with callback as argument.
func (b *MediaMessageCallBuilder) AsEditMediaFromCBQ(callback *tg.CallbackQuery, media tg.InputMedia) *tg.EditMessageMediaCall {
	return b.AsEditMedia(callback.Message.Chat(), callback.Message.MessageID(), media)
}

// AsEditMediaFromMsg wraps AsEditMedia with message as argument.
func (b *MediaMessageCallBuilder) AsEditMediaFromMsg(msg *tg.Message, media tg.InputMedia) *tg.EditMessageMediaCall {
	return b.AsEditMedia(msg.Chat, msg.ID, media)
}

// AsEditMediaInline returns call editMessageMedia by inline message id.
func (b *MediaMessageCallBuilder) AsEditMediaInline(id string, media tg.InputMedia) *tg.EditMessageMediaCall {
	return b.applyEditMedia(tg.NewEditMessageMediaInlineCall(id, media))
}

func (b *MediaMessageCallBuilder) applyEditMedia(call *tg.EditMessageMediaCall) *tg.EditMessageMediaCall {
	if b.replyMarkup != nil {
		if v, ok := b.replyMarkup.(tg.InlineKeyboardMarkup); ok {
			call.ReplyMarkup(v)
		}
	}

	if b.businessConnectionID != "" {
		call.BusinessConnectionID(b.businessConnectionID)
	}

	if b.client != nil {
		call.Bind(b.client)
	}

	return call
}

// AsEditCaptionInline returns call editMessageCaption by inline message id.
func (b *MediaMessageCallBuilder) AsEditCaptionInline(id string) *tg.EditMessageCaptionCall {
	call := tg.NewEditMessageCaptionInlineCall(id, b.caption)

	if b.parseMode != nil {
		call.ParseMode(b.parseMode)
	}

	if b.captionEntities != nil {
		call.CaptionEntities(b.captionEntities)
	}

	if b.showCaptionAboveMedia {
		call.ShowCaptionAboveMedia(b.showCaptionAboveMedia)
	}

	if b.replyMarkup != nil {
		if v, ok := b.replyMarkup.(tg.InlineKeyboardMarkup); ok {
			call.ReplyMarkup(v)
		}
	}

	if b.businessConnectionID != "" {
		call.BusinessConnectionID(b.businessConnectionID)
	}

	if b.client != nil {
		call.Bind(b.client)
	}

	return call
}
