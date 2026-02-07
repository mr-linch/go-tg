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

// Msg returns message from whever possible.
// It returns nil if message is not found.
func (update *Update) Msg() *Message {
	switch {
	case update == nil:
		return nil
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

// Chat returns chat from whever possible.
func (update *Update) Chat() *Chat {
	if update == nil {
		return nil
	}

	if msg := update.Msg(); msg != nil {
		return &msg.Chat
	}

	switch {
	case update.ChatMember != nil:
		return &update.ChatMember.Chat
	case update.MyChatMember != nil:
		return &update.MyChatMember.Chat
	case update.ChatJoinRequest != nil:
		return &update.ChatJoinRequest.Chat
	}

	return nil
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
