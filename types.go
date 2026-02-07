package tg

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

// UnixTime represents a Unix timestamp (seconds since epoch).
// It is stored as int64 and serializes to/from JSON as an integer.
type UnixTime int64

// Time converts the Unix timestamp to time.Time.
// Returns the zero time.Time for UnixTime(0).
func (t UnixTime) Time() time.Time {
	if t == 0 {
		return time.Time{}
	}
	return time.Unix(int64(t), 0)
}

// IsZero reports whether the timestamp is zero (unset).
func (t UnixTime) IsZero() bool {
	return t == 0
}

type ChatID int64

var _ PeerID = ChatID(0)

func (id ChatID) PeerID() string {
	return strconv.FormatInt(int64(id), 10)
}

// UserID it's unique identifier for Telegram user or bot.
type UserID int64

var _ PeerID = UserID(0)

func (id UserID) PeerID() string {
	return strconv.FormatInt(int64(id), 10)
}

// Username represents a Telegram username.
type Username string

// PeerID implements [Peer] interface.
func (un Username) PeerID() string {
	return "@" + string(un)
}

// Link returns a public link with username.
func (un Username) Link() string {
	return "https://t.me/" + string(un)
}

// DeepLink returns a deep link with username.
func (un Username) DeepLink() string {
	return "tg://resolve?domain=" + string(un)
}

// PeerID represents generic Telegram peer.
//
// Known implementations:
//   - [UserID]
//   - [ChatID]
//   - [Username]
//   - [Chat]
//   - [User]
type PeerID interface {
	PeerID() string
}

type FileID string

// FileArg it's union type for different ways of sending files.
type FileArg struct {
	// Send already uploaded file by its file_id.
	FileID FileID

	// Send remote file by URL.
	URL string

	// Upload file
	Upload InputFile

	addr string
}

// NewFileArgUpload creates a new FileArg for uploading a file by content.
func NewFileArgUpload(file InputFile) FileArg {
	return FileArg{
		Upload: file,
	}
}

// NewFileArgURL creates a new FileArg for sending a file by URL.
func NewFileArgURL(url string) FileArg {
	return FileArg{
		URL: url,
	}
}

// NewFileArgID creates a new FileArg for sending a file by file_id.
func NewFileArgID(id FileID) FileArg {
	return FileArg{
		FileID: id,
	}
}

func (arg FileArg) MarshalJSON() ([]byte, error) {
	str := arg.getRef()
	if str != "" {
		return json.Marshal(str)
	}

	return nil, fmt.Errorf("FileArg is not json serializable")
}

// isRef returns true if FileArg is just reference
func (arg *FileArg) isRef() bool {
	return arg.FileID != "" || arg.URL != "" || arg.addr != ""
}

// getRef returns text representation of reference
func (arg *FileArg) getRef() string {
	switch {
	case arg.FileID != "":
		return string(arg.FileID)
	case arg.URL != "":
		return arg.URL
	case arg.addr != "":
		return arg.addr
	default:
		return ""
	}
}

//go:generate go run github.com/mr-linch/go-tg-gen@latest -types-output types_gen.go

func (chat Chat) PeerID() string {
	return chat.ID.PeerID()
}

func (user User) PeerID() string {
	return user.ID.PeerID()
}

// FullName returns the user's full name.
// It combines first name and last name, or returns just first name if last name is empty.
func (user User) FullName() string {
	if user.LastName == "" {
		return user.FirstName
	}
	return user.FirstName + " " + user.LastName
}

// FullName returns the chat's display name.
// For groups, supergroups and channels it returns the title.
// For private chats it combines first name and last name.
func (chat Chat) FullName() string {
	if chat.Title != "" {
		return chat.Title
	}
	if chat.LastName == "" {
		return chat.FirstName
	}
	return chat.FirstName + " " + chat.LastName
}

func (markup InlineKeyboardMarkup) Ptr() *InlineKeyboardMarkup {
	return &markup
}

type CallbackDataEncoder[T any] interface {
	Encode(data T) (string, error)
}

// getMedia returns the media and thumbnail from an InputMedia union.
func (u *InputMedia) getMedia() (media *FileArg, thumb *InputFile) {
	switch {
	case u.Photo != nil:
		return &u.Photo.Media, nil
	case u.Video != nil:
		return &u.Video.Media, u.Video.Thumbnail
	case u.Animation != nil:
		return &u.Animation.Media, u.Animation.Thumbnail
	case u.Audio != nil:
		return &u.Audio.Media, u.Audio.Thumbnail
	case u.Document != nil:
		return &u.Document.Media, u.Document.Thumbnail
	default:
		return nil, nil
	}
}

// getMedia returns the media, thumbnail and cover from an InputPaidMedia union.
func (u *InputPaidMedia) getMedia() (media *FileArg, thumb *InputFile, cover *FileArg) {
	switch {
	case u.Photo != nil:
		return &u.Photo.Media, nil, nil
	case u.Video != nil:
		return &u.Video.Media, u.Video.Thumbnail, u.Video.Cover
	default:
		return nil, nil, nil
	}
}

// MenuButtonOneOf is an alias for MenuButton for backward compatibility.
//
// Deprecated: Use MenuButton directly.
type MenuButtonOneOf = MenuButton

// IsInaccessible returns true if message is inaccessible.
func (msg *Message) IsInaccessible() bool {
	return msg.Date.IsZero()
}

// TextOrCaption returns the message text or caption, whichever is set.
// For text messages it returns Text, for media messages it returns Caption.
func (msg *Message) TextOrCaption() string {
	if msg.Text != "" {
		return msg.Text
	}
	return msg.Caption
}

// TextOrCaptionEntities returns text entities or caption entities, whichever applies.
// For text messages it returns Entities, for media messages it returns CaptionEntities.
func (msg *Message) TextOrCaptionEntities() []MessageEntity {
	if msg.Text != "" {
		return msg.Entities
	}
	return msg.CaptionEntities
}

// FileID returns the file ID from whichever media field is set.
// For photos, it returns the file ID of the largest size.
// Returns empty FileID if the message contains no media.
func (msg *Message) FileID() FileID {
	switch {
	case len(msg.Photo) > 0:
		return msg.Photo[len(msg.Photo)-1].FileID
	case msg.Animation != nil:
		return msg.Animation.FileID
	case msg.Audio != nil:
		return msg.Audio.FileID
	case msg.Document != nil:
		return msg.Document.FileID
	case msg.Video != nil:
		return msg.Video.FileID
	case msg.VideoNote != nil:
		return msg.VideoNote.FileID
	case msg.Voice != nil:
		return msg.Voice.FileID
	case msg.Sticker != nil:
		return msg.Sticker.FileID
	}
	return ""
}

// Msg returns message from whever possible.
// It returns nil if message is not found.
func (update *Update) Msg() *Message {
	if update == nil {
		return nil
	}
	switch {
	case update.Message != nil:
		return update.Message
	case update.EditedMessage != nil:
		return update.EditedMessage
	case update.ChannelPost != nil:
		return update.ChannelPost
	case update.EditedChannelPost != nil:
		return update.EditedChannelPost
	case update.CallbackQuery != nil && update.CallbackQuery.Message != nil && update.CallbackQuery.Message.Message != nil:
		return update.CallbackQuery.Message.Message
	case update.BusinessMessage != nil:
		return update.BusinessMessage
	case update.EditedBusinessMessage != nil:
		return update.EditedBusinessMessage
	default:
		return nil
	}
}

// Chat returns chat from wherever possible.
// It returns nil if no chat can be determined from this update.
func (update *Update) Chat() *Chat {
	if update == nil {
		return nil
	}
	if msg := update.Msg(); msg != nil {
		return &msg.Chat
	}

	switch {
	case update.MessageReaction != nil:
		return &update.MessageReaction.Chat
	case update.MessageReactionCount != nil:
		return &update.MessageReactionCount.Chat
	case update.ChatMember != nil:
		return &update.ChatMember.Chat
	case update.MyChatMember != nil:
		return &update.MyChatMember.Chat
	case update.ChatJoinRequest != nil:
		return &update.ChatJoinRequest.Chat
	case update.DeletedBusinessMessages != nil:
		return &update.DeletedBusinessMessages.Chat
	case update.ChatBoost != nil:
		return &update.ChatBoost.Chat
	case update.RemovedChatBoost != nil:
		return &update.RemovedChatBoost.Chat
	case update.PollAnswer != nil && update.PollAnswer.VoterChat != nil:
		return update.PollAnswer.VoterChat
	}

	return nil
}

// User returns the user from wherever possible.
// It returns nil if no user can be determined from this update.
func (update *Update) User() *User {
	if update == nil {
		return nil
	}
	if msg := update.Msg(); msg != nil {
		return msg.From
	}

	switch {
	case update.CallbackQuery != nil:
		return &update.CallbackQuery.From
	case update.InlineQuery != nil:
		return &update.InlineQuery.From
	case update.ChosenInlineResult != nil:
		return &update.ChosenInlineResult.From
	case update.ShippingQuery != nil:
		return &update.ShippingQuery.From
	case update.PreCheckoutQuery != nil:
		return &update.PreCheckoutQuery.From
	case update.PurchasedPaidMedia != nil:
		return &update.PurchasedPaidMedia.From
	case update.MyChatMember != nil:
		return &update.MyChatMember.From
	case update.ChatMember != nil:
		return &update.ChatMember.From
	case update.ChatJoinRequest != nil:
		return &update.ChatJoinRequest.From
	case update.MessageReaction != nil:
		return update.MessageReaction.User
	case update.PollAnswer != nil:
		return update.PollAnswer.User
	case update.BusinessConnection != nil:
		return &update.BusinessConnection.User
	}

	return nil
}

// SenderChat returns the sender chat from wherever possible.
// This is set when a message is sent on behalf of a chat (e.g. anonymous admin),
// or when a reaction is performed by an anonymous chat.
// It returns nil if no sender chat can be determined.
func (update *Update) SenderChat() *Chat {
	if update == nil {
		return nil
	}
	if msg := update.Msg(); msg != nil {
		return msg.SenderChat
	}

	switch {
	case update.MessageReaction != nil:
		return update.MessageReaction.ActorChat
	case update.PollAnswer != nil:
		return update.PollAnswer.VoterChat
	}

	return nil
}

// MsgID returns the message ID from wherever possible.
// It returns 0 if no message ID can be determined from this update.
func (update *Update) MsgID() int {
	if update == nil {
		return 0
	}
	if msg := update.Msg(); msg != nil {
		return msg.ID
	}

	switch {
	case update.MessageReaction != nil:
		return update.MessageReaction.MessageID
	case update.MessageReactionCount != nil:
		return update.MessageReactionCount.MessageID
	}

	return 0
}

// ChatID returns the chat ID from wherever possible.
// Unlike Chat().ID, this also covers BusinessConnection.UserChatID
// where no Chat object is available.
// It returns 0 if no chat ID can be determined from this update.
func (update *Update) ChatID() ChatID {
	if update == nil {
		return 0
	}
	if chat := update.Chat(); chat != nil {
		return chat.ID
	}

	if update.BusinessConnection != nil {
		return update.BusinessConnection.UserChatID
	}

	return 0
}

// InlineMessageID returns the inline message ID from wherever possible.
// It returns an empty string if no inline message ID can be determined.
func (update *Update) InlineMessageID() string {
	if update == nil {
		return ""
	}
	switch {
	case update.CallbackQuery != nil:
		return update.CallbackQuery.InlineMessageID
	case update.ChosenInlineResult != nil:
		return update.ChosenInlineResult.InlineMessageID
	}

	return ""
}

// BusinessConnectionID returns the business connection ID from wherever possible.
// It returns an empty string if no business connection ID can be determined.
func (update *Update) BusinessConnectionID() string {
	if update == nil {
		return ""
	}
	switch {
	case update.BusinessConnection != nil:
		return update.BusinessConnection.ID
	case update.DeletedBusinessMessages != nil:
		return update.DeletedBusinessMessages.BusinessConnectionID
	default:
		if msg := update.Msg(); msg != nil {
			return msg.BusinessConnectionID
		}
		return ""
	}
}

// Extract entitie value from plain text.
func (me MessageEntity) Extract(text string) string {
	return string([]rune(text)[me.Offset : me.Offset+me.Length])
}

// This object describes a message that can be inaccessible to the bot.
// It can be one of:
//   - [Message]
//   - [InaccessibleMessage]
type MaybeInaccessibleMessage struct {
	Message             *Message
	InaccessibleMessage *InaccessibleMessage
}

// IsInaccessible returns true if message is inaccessible.
func (mim *MaybeInaccessibleMessage) IsInaccessible() bool {
	return mim.InaccessibleMessage != nil
}

// IsAccessible returns true if message is accessible.
func (mim *MaybeInaccessibleMessage) IsAccessible() bool {
	return mim.Message != nil
}

func (mim *MaybeInaccessibleMessage) Chat() Chat {
	if mim.InaccessibleMessage != nil {
		return mim.InaccessibleMessage.Chat
	}

	return mim.Message.Chat
}

func (mim *MaybeInaccessibleMessage) MessageID() int {
	if mim.InaccessibleMessage != nil {
		return mim.InaccessibleMessage.MessageID
	}

	return mim.Message.ID
}

func (mim *MaybeInaccessibleMessage) UnmarshalJSON(v []byte) error {
	var partial struct {
		Date UnixTime `json:"date"`
	}

	if err := json.Unmarshal(v, &partial); err != nil {
		return fmt.Errorf("unmarshal MaybeInaccessibleMessage partial: %w", err)
	}

	if partial.Date.IsZero() {
		mim.InaccessibleMessage = &InaccessibleMessage{}
		return json.Unmarshal(v, mim.InaccessibleMessage)
	} else {
		mim.Message = &Message{}
		return json.Unmarshal(v, mim.Message)
	}
}

// RetryAfterDuration returns duration for retry after.
func (rp *ResponseParameters) RetryAfterDuration() time.Duration {
	return time.Duration(rp.RetryAfter) * time.Second
}

// TopicIconColor represents the color of a forum topic icon in RGB format.
type TopicIconColor int

const (
	TopicIconColorBlue   TopicIconColor = 0x6FB9F0
	TopicIconColorYellow TopicIconColor = 0xFFD67E
	TopicIconColorPurple TopicIconColor = 0xCB86DB
	TopicIconColorGreen  TopicIconColor = 0x8EEE98
	TopicIconColorPink   TopicIconColor = 0xFF93B2
	TopicIconColorRed    TopicIconColor = 0xFB6F5F
)
