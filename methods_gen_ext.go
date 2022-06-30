package tg

import "context"

//go:generate go run github.com/mr-linch/go-tg-gen@latest -methods-output methods_gen.go

// Me returns cached current bot info.
func (client *Client) Me(ctx context.Context) (User, error) {
	if client.me == nil {
		user, err := client.GetMe().Do(ctx)
		if err != nil {
			return User{}, err
		}
		client.me = &user
	}
	return *client.me, nil
}

// SendMediaGroupCall reprenesents a call to the sendMediaGroup method.
// Use this method to send a group of photos, videos, documents or audios as an album
// Documents and audio files can be only grouped in an album with messages of the same type
// On success, an array of Messages that were sent is returned.
type SendMediaGroupCall struct {
	Call[[]Message]
}

// NewSendMediaGroupCall constructs a new SendMediaGroupCall with required parameters.
// chatId - Unique identifier for the target chat or username of the target channel (in the format @channelusername)
// media - A JSON-serialized array describing messages to be sent, must include 2-10 items
func NewSendMediaGroupCall(chatId PeerID, media []InputMedia) *SendMediaGroupCall {
	return &SendMediaGroupCall{
		Call[[]Message]{
			request: NewRequest("sendMediaGroup").
				PeerID("chat_id", chatId).
				InputMediaSlice(media),
		},
	}
}

// SendMediaGroupCall constructs a new SendMediaGroupCall with required parameters.
func (client *Client) SendMediaGroup(chatId PeerID, media []InputMedia) *SendMediaGroupCall {
	return callWithClient(
		client,
		NewSendMediaGroupCall(chatId, media),
	)
}

// ChatId Unique identifier for the target chat or username of the target channel (in the format @channelusername)
func (call *SendMediaGroupCall) ChatId(chatId PeerID) *SendMediaGroupCall {
	call.request.PeerID("chat_id", chatId)
	return call
}

// Media A JSON-serialized array describing messages to be sent, must include 2-10 items
func (call *SendMediaGroupCall) Media(media []InputMedia) *SendMediaGroupCall {
	call.request.JSON("media", media)
	return call
}

// DisableNotification Sends messages silently. Users will receive a notification with no sound.
func (call *SendMediaGroupCall) DisableNotification(disableNotification bool) *SendMediaGroupCall {
	call.request.Bool("disable_notification", disableNotification)
	return call
}

// ProtectContent Protects the contents of the sent messages from forwarding and saving
func (call *SendMediaGroupCall) ProtectContent(protectContent bool) *SendMediaGroupCall {
	call.request.Bool("protect_content", protectContent)
	return call
}

// ReplyToMessageId If the messages are a reply, ID of the original message
func (call *SendMediaGroupCall) ReplyToMessageId(replyToMessageId int) *SendMediaGroupCall {
	call.request.Int("reply_to_message_id", replyToMessageId)
	return call
}

// AllowSendingWithoutReply Pass True, if the message should be sent even if the specified replied-to message is not found
func (call *SendMediaGroupCall) AllowSendingWithoutReply(allowSendingWithoutReply bool) *SendMediaGroupCall {
	call.request.Bool("allow_sending_without_reply", allowSendingWithoutReply)
	return call
}
