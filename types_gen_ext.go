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

var _ PeerID = (ChatID)(0)

func (id ChatID) PeerID() string {
	return strconv.FormatInt(int64(id), 10)
}

// UserID it's unique identifier for Telegram user or bot.
type UserID int64

var _ PeerID = (UserID)(0)

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
	if arg.FileID != "" {
		return string(arg.FileID)
	} else if arg.URL != "" {
		return arg.URL
	} else if arg.addr != "" {
		return arg.addr
	}

	return ""
}

//go:generate go run github.com/mr-linch/go-tg-gen@latest -types-output types_gen.go

func (chat Chat) PeerID() string {
	return chat.ID.PeerID()
}

func (user User) PeerID() string {
	return user.ID.PeerID()
}


// ReplyMarkup generic for keyboards.
//
// Known implementations:
//   - [ReplyKeyboardMarkup]
//   - [InlineKeyboardMarkup]
//   - [ReplyKeyboardRemove]
//   - [ForceReply]
type ReplyMarkup interface {
	isReplyMarkup()
}

var _ ReplyMarkup = (*InlineKeyboardMarkup)(nil)

// NewInlineKeyboardMarkup creates a new InlineKeyboardMarkup.
func NewInlineKeyboardMarkup(rows ...[]InlineKeyboardButton) InlineKeyboardMarkup {
	return InlineKeyboardMarkup{
		InlineKeyboard: rows,
	}
}

func (markup InlineKeyboardMarkup) Ptr() *InlineKeyboardMarkup {
	return &markup
}

// NewInlineButtonURL create inline button
// with http(s):// or tg:// URL to be opened when the button is pressed.
func NewInlineKeyboardButtonURL(text string, url string) InlineKeyboardButton {
	return InlineKeyboardButton{
		Text: text,
		URL:  url,
	}
}

// NewInlineKeyboardButtonCallback creates a new InlineKeyboardButton with specified callback data.
// Query should have length 1-64 bytes.
func NewInlineKeyboardButtonCallback(text string, query string) InlineKeyboardButton {
	return InlineKeyboardButton{
		Text:         text,
		CallbackData: query,
	}
}

type CallbackDataEncoder[T any] interface {
	Encode(data T) (string, error)
}

// NewInlineKeyboardButtonWebApp creates a button that open a web app.
func NewInlineKeyboardButtonWebApp(text string, webApp WebAppInfo) InlineKeyboardButton {
	return InlineKeyboardButton{
		Text:   text,
		WebApp: &webApp,
	}
}

// InlineKeyboardMarkup represents button that open web page with auth data.
func NewInlineKeyboardButtonLoginURL(text string, loginURL LoginURL) InlineKeyboardButton {
	return InlineKeyboardButton{
		Text:     text,
		LoginURL: &loginURL,
	}
}

// NewInlineKeyboardButtonSwitchInlineQuery represents button that
//
//	will prompt the user to select one of their chats,
//
// open that chat and insert the bot's username and the specified inline query in the input field.
func NewInlineKeyboardButtonSwitchInlineQuery(text string, query string) InlineKeyboardButton {
	return InlineKeyboardButton{
		Text:              text,
		SwitchInlineQuery: query,
	}
}

// NewInlineKeyboardButtonSwitchInlineQueryCurrentChat represents button that
// will insert the bot's username and the specified inline query in the current chat's input field
func NewInlineKeyboardButtonSwitchInlineQueryCurrentChat(text string, query string) InlineKeyboardButton {
	return InlineKeyboardButton{
		Text:                         text,
		SwitchInlineQueryCurrentChat: query,
	}
}

// NewInlineKeyboardButtonCallbackGame represents the button which open a game.
func NewInlineKeyboardButtonCallbackGame(text string) InlineKeyboardButton {
	return InlineKeyboardButton{
		Text:         text,
		CallbackGame: &CallbackGame{},
	}
}

// NewInlineKeyboardButtonPay represents a Pay button.
// NOTE: This type of button must always be the first button in the first row and can only be used in invoice messages
func NewInlineKeyboardButtonPay(text string) InlineKeyboardButton {
	return InlineKeyboardButton{
		Text: text,
		Pay:  true,
	}
}

func (markup InlineKeyboardMarkup) isReplyMarkup() {}

// NewReplyKeyboardMarkup creates a new ReplyKeyboardMarkup.
func NewReplyKeyboardMarkup(rows ...[]KeyboardButton) *ReplyKeyboardMarkup {
	return &ReplyKeyboardMarkup{
		Keyboard: rows,
	}
}

var _ ReplyMarkup = (*ReplyKeyboardMarkup)(nil)

// WithResizeKeyboard requests clients to resize the keyboard vertically for optimal fit (e.g., make the keyboard smaller if there are just two rows of buttons).
// Defaults to false, in which case the custom keyboard is always of the same height as the app's standard keyboard.
func (markup *ReplyKeyboardMarkup) WithResizeKeyboardMarkup() *ReplyKeyboardMarkup {
	markup.ResizeKeyboard = true
	return markup
}

// WithOneTimeKeyboard  requests clients to hide the keyboard as soon as it's been used.
// The keyboard will still be available, but clients will automatically display the
// usual letter-keyboard in the chat - the user can press a special button in
// the input field to see the custom keyboard again.
// Defaults to false.
func (markup *ReplyKeyboardMarkup) WithOneTimeKeyboardMarkup() *ReplyKeyboardMarkup {
	markup.OneTimeKeyboard = true
	return markup
}

// WithInputFieldPlaceholder sets the placeholder to be shown in the input field when the keyboard is active;
// 1-64 characters
func (markup *ReplyKeyboardMarkup) WithInputFieldPlaceholder(placeholder string) *ReplyKeyboardMarkup {
	markup.InputFieldPlaceholder = placeholder
	return markup
}

// Use this parameter if you want to show the keyboard to specific users only.
func (markup *ReplyKeyboardMarkup) WithSelective() *ReplyKeyboardMarkup {
	markup.Selective = true
	return markup
}

// NewKeyboardButton creates a plain reply keyboard button.
func NewKeyboardButton(text string) KeyboardButton {
	return KeyboardButton{
		Text: text,
	}
}

// NewKeyboardButtonRequestContact creates a reply keyboard button that request a contact from user.
// Available in private chats only.
func NewKeyboardButtonRequestContact(text string) KeyboardButton {
	return KeyboardButton{
		Text:           text,
		RequestContact: true,
	}
}

// NewKeyboardButtonRequestLocation creates a reply keyboard button that request a location from user.
// Available in private chats only.
func NewKeyboardButtonRequestLocation(text string) KeyboardButton {
	return KeyboardButton{
		Text:            text,
		RequestLocation: true,
	}
}

// NewKeyboardButtonRequestPoll creates a reply keyboard button that request a poll from user.
// Available in private chats only.
func NewKeyboardButtonRequestPoll(text string, poll KeyboardButtonPollType) KeyboardButton {
	return KeyboardButton{
		Text:        text,
		RequestPoll: &poll,
	}
}

// NewKeyboardButtonWebApp create a reply keyboard button that open a web app.
func NewKeyboardButtonWebApp(text string, webApp WebAppInfo) KeyboardButton {
	return KeyboardButton{
		Text:   text,
		WebApp: &webApp,
	}
}

// NewKeyboardButtonRequestUsers creates a reply keyboard button that request a user from user.
// Available in private chats only.
func NewKeyboardButtonRequestUsers(text string, config KeyboardButtonRequestUsers) KeyboardButton {
	return KeyboardButton{
		Text:         text,
		RequestUsers: &config,
	}
}

// NewKeyboardButtonRequestChats creates a reply keyboard button that request a chat from user.
// Available in private chats only.
func NewKeyboardButtonRequestChat(text string, config KeyboardButtonRequestChat) KeyboardButton {
	return KeyboardButton{
		Text:        text,
		RequestChat: &config,
	}
}

func (markup ReplyKeyboardMarkup) isReplyMarkup() {}

var _ ReplyMarkup = (*ReplyKeyboardRemove)(nil)

// NewReplyKeyboardRemove creates a new ReplyKeyboardRemove.
func NewReplyKeyboardRemove() *ReplyKeyboardRemove {
	return &ReplyKeyboardRemove{
		RemoveKeyboard: true,
	}
}

// WithSelective set it if you want to remove the keyboard for specific users only.
func (markup *ReplyKeyboardRemove) WithSelective() *ReplyKeyboardRemove {
	markup.Selective = true
	return markup
}

func (markup ReplyKeyboardRemove) isReplyMarkup() {}

var _ ReplyMarkup = (*ForceReply)(nil)

// NewForceReply creates a new ForceReply.
func NewForceReply() *ForceReply {
	return &ForceReply{
		ForceReply: true,
	}
}

// WithSelective set it if you want to force reply for specific users only.
func (markup *ForceReply) WithSelective() *ForceReply {
	markup.Selective = true
	return markup
}

// WithInputFieldPlaceholder sets the placeholder to be shown in the input field when the reply is active;
// 1-64 characters
func (markup *ForceReply) WithInputFieldPlaceholder(placeholder string) *ForceReply {
	markup.InputFieldPlaceholder = placeholder
	return markup
}

func (markup ForceReply) isReplyMarkup() {}

// NewButtonRow it's generic helper for create keyboards in functional way.
func NewButtonRow[T Button](buttons ...T) []T {
	return buttons
}

// Button define generic button interface
type Button interface {
	InlineKeyboardButton | KeyboardButton
}

// ButtonLayout it's build for fixed width keyboards.
type ButtonLayout[T Button] struct {
	buttons  [][]T
	rowWidth int
}

// NewButtonColumn returns keyboard from a single column of Button.
func NewButtonColumn[T Button](buttons ...T) [][]T {
	result := make([][]T, 0, len(buttons))

	for _, button := range buttons {
		result = append(result, []T{button})
	}

	return result
}

// NewButtonLayout creates layout with specified width.
// Buttons will be added via Insert method.
func NewButtonLayout[T Button](rowWidth int, buttons ...T) *ButtonLayout[T] {
	layout := &ButtonLayout[T]{
		rowWidth: rowWidth,
		buttons:  make([][]T, 0),
	}

	return layout.Insert(buttons...)
}

// Keyboard returns result of building.
func (layout *ButtonLayout[T]) Keyboard() [][]T {
	return layout.buttons
}

// Insert buttons to last row if possible, or create new and insert.
func (layout *ButtonLayout[T]) Insert(buttons ...T) *ButtonLayout[T] {
	for _, button := range buttons {
		layout.insert(button)
	}

	return layout
}

func (layout *ButtonLayout[T]) insert(button T) *ButtonLayout[T] {
	if len(layout.buttons) > 0 && len(layout.buttons[len(layout.buttons)-1]) < layout.rowWidth {
		layout.buttons[len(layout.buttons)-1] = append(layout.buttons[len(layout.buttons)-1], button)
	} else {
		layout.buttons = append(layout.buttons, []T{button})
	}
	return layout
}

// Add accepts any number of buttons,
// always starts adding from a new row
// and adds a row when it reaches the set width.
func (layout *ButtonLayout[T]) Add(buttons ...T) *ButtonLayout[T] {
	row := make([]T, 0, layout.rowWidth)

	for _, button := range buttons {
		if len(row) == layout.rowWidth {
			layout.buttons = append(layout.buttons, row)
			row = make([]T, 0, layout.rowWidth)
		}

		row = append(row, button)
	}

	if len(row) > 0 {
		layout.buttons = append(layout.buttons, row)
	}

	return layout
}

// Row add new row with no respect for row width
func (layout *ButtonLayout[T]) Row(buttons ...T) *ButtonLayout[T] {
	layout.buttons = append(layout.buttons, buttons)
	return layout
}

// InputMessageContent it's generic interface for all types of input message content.
//
// Known implementations:
//   - [InputTextMessageContent]
//   - [InputLocationMessageContent]
//   - [InputVenueMessageContent]
//   - [InputContactMessageContent]
//   - [InputInvoiceMessageContent]
type InputMessageContent interface {
	isInputMessageContent()
}

func (content InputTextMessageContent) isInputMessageContent()     {}
func (content InputLocationMessageContent) isInputMessageContent() {}
func (content InputVenueMessageContent) isInputMessageContent()    {}
func (content InputContactMessageContent) isInputMessageContent()  {}
func (content InputInvoiceMessageContent) isInputMessageContent()  {}

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

// MenuButtonOneOf is an alias for MenuButton for backward compatibility.
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
