package tgb

import (
	"context"
	"encoding/json"

	"github.com/mr-linch/go-tg"
)

// Update wraps around a tg.Update.
// Also contains Client which is used to send responses.
type Update struct {
	*tg.Update
	Client *tg.Client

	webhookResponse chan json.Marshaler
}

func (update *Update) isWebhook() bool {
	return update.webhookResponse != nil
}

// UpdateRespond defines interface for responding to an update via Webhook.
type UpdateRespond interface {
	json.Marshaler
	DoVoid(ctx context.Context) error
	Bind(client *tg.Client)
}

// Respond to Webhook, if possible or make usual call via Client.
func (update *Update) Respond(ctx context.Context, v UpdateRespond) error {
	if update.isWebhook() {
		update.webhookResponse <- v
		return nil
	}

	v.Bind(update.Client)

	return v.DoVoid(ctx)
}

// MessageUpdate it's extend wrapper around tg.Message.
type MessageUpdate struct {
	*tg.Message
	Client *tg.Client
	Update *Update
}

// Answer calls sendMessage with pre-defined chatID to incoming message chat.
func (msg *MessageUpdate) Answer(text string) *tg.SendMessageCall {
	return msg.Client.SendMessage(msg.Chat, text)
}

// AnswerPhoto calls sendPhoto with pre-defined chatID to incoming message chat.
func (msg *MessageUpdate) AnswerPhoto(photo tg.FileArg) *tg.SendPhotoCall {
	return msg.Client.SendPhoto(msg.Chat, photo)
}

// AnswerAudio calls sendAudio with pre-defined chatID to incoming message chat.
func (msg *MessageUpdate) AnswerAudio(audio tg.FileArg) *tg.SendAudioCall {
	return msg.Client.SendAudio(msg.Chat, audio)
}

// AnswerAnimation calls sendAnimation with pre-defined chatID to incoming message chat.
func (msg *MessageUpdate) AnswerAnimation(animation tg.FileArg) *tg.SendAnimationCall {
	return msg.Client.SendAnimation(msg.Chat, animation)
}

// AnswerDocument calls sendDocument with pre-defined chatID to incoming message chat.
func (msg *MessageUpdate) AnswerDocument(document tg.FileArg) *tg.SendDocumentCall {
	return msg.Client.SendDocument(msg.Chat, document)
}

// AnswerVideo calls sendVideo with pre-defined chatID to incoming message chat.
func (msg *MessageUpdate) AnswerVideo(video tg.FileArg) *tg.SendVideoCall {
	return msg.Client.SendVideo(msg.Chat, video)
}

// AnswerVoice calls sendVoice with pre-defined chatID to incoming message chat.
func (msg *MessageUpdate) AnswerVoice(voice tg.FileArg) *tg.SendVoiceCall {
	return msg.Client.SendVoice(msg.Chat, voice)
}

// AnswerVideoNote calls sendVideoNote with pre-defined chatID to incoming message chat.
func (msg *MessageUpdate) AnswerVideoNote(videoNote tg.FileArg) *tg.SendVideoNoteCall {
	return msg.Client.SendVideoNote(msg.Chat, videoNote)
}

// AnswerLocation calls sendLocation with pre-defined chatID to incoming message chat.
func (msg *MessageUpdate) AnswerLocation(latitude float64, longitude float64) *tg.SendLocationCall {
	return msg.Client.SendLocation(msg.Chat, latitude, longitude)
}

// AnswerVenue calls sendVenue with pre-defined chatID to incoming message chat.
func (msg *MessageUpdate) AnswerVenue(latitude float64, longitude float64, title string, address string) *tg.SendVenueCall {
	return msg.Client.SendVenue(msg.Chat, latitude, longitude, title, address)
}

// AnswerContact calls sendContact with pre-defined chatID to incoming message chat.
func (msg *MessageUpdate) AnswerContact(phoneNumber string, firstName string) *tg.SendContactCall {
	return msg.Client.SendContact(msg.Chat, phoneNumber, firstName)
}

// AnswerSticker calls sendSticker with pre-defined chatID to incoming message chat.
func (msg *MessageUpdate) AnswerSticker(sticker tg.FileArg) *tg.SendStickerCall {
	return msg.Client.SendSticker(msg.Chat, sticker)
}

// AnswerPoll calls sendPoll with pre-defined chatID to incoming message chat.
func (msg *MessageUpdate) AnswerPoll(question string, options []string) *tg.SendPollCall {
	return msg.Client.SendPoll(msg.Chat, question, options)
}

// AnswerDice calls sendDice with pre-defined chatID to incoming message chat.
func (msg *MessageUpdate) AnswerDice(emoji string) *tg.SendDiceCall {
	return msg.Client.SendDice(msg.Chat).Emoji(emoji)
}

// AnswerChatAction calls sendChatAction with pre-defined chatID to incoming message chat.
func (msg *MessageUpdate) AnswerChatAction(action tg.ChatAction) *tg.SendChatActionCall {
	return msg.Client.SendChatAction(msg.Chat, action)
}

// Forward incoming message to another chat.
func (msg *MessageUpdate) Forward(to tg.PeerID) *tg.ForwardMessageCall {
	return msg.Client.ForwardMessage(to, msg.Chat, msg.ID)
}

// Copy incoming message to another chat.
func (msg *MessageUpdate) Copy(to tg.PeerID) *tg.CopyMessageCall {
	return msg.Client.CopyMessage(to, msg.Chat, msg.ID)
}

// EditText of incoming message.
func (msg *MessageUpdate) EditText(text string) *tg.EditMessageTextCall {
	return msg.Client.EditMessageText(msg.Chat, msg.ID, text)
}

// EditCaption of incoming message.
func (msg *MessageUpdate) EditCaption(caption string) *tg.EditMessageCaptionCall {
	return msg.Client.EditMessageCaption(msg.Chat, msg.ID, caption)
}

// EditReplyMarkup of incoming message.
func (msg *MessageUpdate) EditReplyMarkup(markup tg.InlineKeyboardMarkup) *tg.EditMessageReplyMarkupCall {
	return msg.Client.EditMessageReplyMarkup(msg.Chat, msg.ID).ReplyMarkup(markup)
}

type CallbackQueryUpdate struct {
	*tg.CallbackQuery

	Update *Update
	Client *tg.Client
}

// Answer without response (just hide loading icon)
func (cbq *CallbackQueryUpdate) Answer() *tg.AnswerCallbackQueryCall {
	return cbq.Client.AnswerCallbackQuery(cbq.ID)
}

// AnswerText with text response and optional alert
func (cbq *CallbackQueryUpdate) AnswerText(text string, alert bool) *tg.AnswerCallbackQueryCall {
	return cbq.Client.AnswerCallbackQuery(cbq.ID).Text(text).ShowAlert(alert)
}

// AnswerURL with URL response and optional.
// URL has limitations, see CallbackQuery.Url for more details.
func (cbq *CallbackQueryUpdate) AnswerURL(url string) *tg.AnswerCallbackQueryCall {
	return cbq.Client.AnswerCallbackQuery(cbq.ID).Url(url)
}

type InlineQueryUpdate struct {
	*tg.InlineQuery
	Update *Update
	Client *tg.Client
}

func (iq *InlineQueryUpdate) Answer(results []tg.InlineQueryResult) *tg.AnswerInlineQueryCall {
	return iq.Client.AnswerInlineQuery(iq.ID, results)
}

type ChosenInlineResultUpdate struct {
	*tg.ChosenInlineResult
	Update *Update
	Client *tg.Client
}

type ShippingQueryUpdate struct {
	*tg.ShippingQuery
	Update *Update
	Client *tg.Client
}

func (sq *ShippingQueryUpdate) Answer(ok bool) *tg.AnswerShippingQueryCall {
	return sq.Client.AnswerShippingQuery(sq.ID, ok)
}

type PreCheckoutQueryUpdate struct {
	*tg.PreCheckoutQuery
	Update *Update
	Client *tg.Client
}

func (pcq *PreCheckoutQueryUpdate) Answer(ok bool) *tg.AnswerPreCheckoutQueryCall {
	return pcq.Client.AnswerPreCheckoutQuery(pcq.ID, ok)
}

type PollUpdate struct {
	*tg.Poll
	Update *Update
	Client *tg.Client
}

type PollAnswerUpdate struct {
	*tg.PollAnswer
	Update *Update
	Client *tg.Client
}

type ChatMemberUpdatedUpdate struct {
	*tg.ChatMemberUpdated
	Update *Update
	Client *tg.Client
}

type ChatJoinRequestUpdate struct {
	*tg.ChatJoinRequest
	Update *Update
	Client *tg.Client
}

// Approve join request
func (joinRequest *ChatJoinRequestUpdate) Approve() *tg.ApproveChatJoinRequestCall {
	return joinRequest.Client.ApproveChatJoinRequest(joinRequest.Chat, joinRequest.From.ID)
}

// Decline join request
func (joinRequest *ChatJoinRequestUpdate) Decline() *tg.DeclineChatJoinRequestCall {
	return joinRequest.Client.DeclineChatJoinRequest(joinRequest.Chat, joinRequest.From.ID)
}
