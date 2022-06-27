package tg

import (
	"encoding/json"
	"fmt"
	"strconv"
)

type ChatID int64

var _ PeerID = (ChatID)(0)

func (id ChatID) PeerID() string {
	return strconv.FormatInt(int64(id), 10)
}

// ChatType represents enum of possible chat types.
type ChatType int8

const (
	// ChatTypePrivate represents one-to-one chat.
	ChatTypePrivate ChatType = iota + 1
	// ChatTypeGroup represents group chats.
	ChatTypeGroup
	// ChatTypeSupergroup supergroup chats.
	ChatTypeSupergroup
	// ChatTypeChannel represents channels
	ChatTypeChannel
	// ChatTypeSender for a private chat with the inline query sender
	ChatTypeSender
)

func (chatType ChatType) String() string {
	if chatType < ChatTypePrivate || chatType > ChatTypeSender {
		return "unknown"
	}

	return [...]string{"private", "group", "supergroup", "channel", "sender"}[chatType-1]
}

func (chatType ChatType) MarshalJSON() ([]byte, error) {
	return json.Marshal(chatType.String())
}

func (chatType *ChatType) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	switch s {
	case "private":
		*chatType = ChatTypePrivate
	case "group":
		*chatType = ChatTypeGroup
	case "supergroup":
		*chatType = ChatTypeSupergroup
	case "channel":
		*chatType = ChatTypeChannel
	case "sender":
		*chatType = ChatTypeSender
	default:
		return fmt.Errorf("unknown chat type: %s", s)
	}

	return nil
}

// UserID it's unique identifier for Telegram user or bot.
type UserID int64

var _ PeerID = (UserID)(0)

func (id UserID) PeerID() string {
	return strconv.FormatInt(int64(id), 10)
}

type Username string

func (un Username) PeerID() string {
	return "@" + string(un)
}

type PeerID interface {
	PeerID() string
}

// MessageID it's unique identifier for a message in a chat.
type MessageID int

type FileID string

type FileArg struct {
	FileID FileID
	Upload InputFile
}

//go:generate go run github.com/mr-linch/go-tg-gen@latest -types-output types_gen.go

func (chat Chat) PeerID() string {
	return chat.ID.PeerID()
}

type InputMedia struct {
}

type BotCommandScope struct {
}

type CallbackGame struct{}

// ReplyMarkup represents a custom keyboard.
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

// NewInlineKeyboardButtonWebApp creates a button that open a web app.
func NewInlineKeyboardButtonWebApp(text string, webApp WebAppInfo) InlineKeyboardButton {
	return InlineKeyboardButton{
		Text:   text,
		WebApp: &webApp,
	}
}

// InlineKeyboardMarkup represents button that open web page with auth data.
func NewInlineKeyboardButtonLoginURL(text string, loginURL LoginUrl) InlineKeyboardButton {
	return InlineKeyboardButton{
		Text:     text,
		LoginURL: &loginURL,
	}
}

// NewInlineKeyboardButtonSwitchInlineQuery represents button that
//  will prompt the user to select one of their chats,
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

type InlineQueryResult interface {
	isInlineQueryResult()
}

func (InlineQueryResultCachedAudio) isInlineQueryResult() {}
func (result InlineQueryResultCachedAudio) MarshalJSON() ([]byte, error) {
	result.Type = "audio"
	type alias InlineQueryResultCachedAudio
	return json.Marshal(alias(result))
}

func (InlineQueryResultCachedDocument) isInlineQueryResult() {}
func (result InlineQueryResultCachedDocument) MarshalJSON() ([]byte, error) {
	result.Type = "document"
	type alias InlineQueryResultCachedDocument
	return json.Marshal(alias(result))
}

func (InlineQueryResultCachedGif) isInlineQueryResult() {}
func (result InlineQueryResultCachedGif) MarshalJSON() ([]byte, error) {
	result.Type = "gif"
	type alias InlineQueryResultCachedGif
	return json.Marshal(alias(result))
}

func (InlineQueryResultCachedMpeg4Gif) isInlineQueryResult() {}
func (result InlineQueryResultCachedMpeg4Gif) MarshalJSON() ([]byte, error) {
	result.Type = "mpeg4_gif"
	type alias InlineQueryResultCachedMpeg4Gif
	return json.Marshal(alias(result))
}

func (InlineQueryResultCachedPhoto) isInlineQueryResult() {}
func (result InlineQueryResultCachedPhoto) MarshalJSON() ([]byte, error) {
	result.Type = "photo"
	type alias InlineQueryResultCachedPhoto
	return json.Marshal(alias(result))
}

func (InlineQueryResultCachedSticker) isInlineQueryResult() {}
func (result InlineQueryResultCachedSticker) MarshalJSON() ([]byte, error) {
	result.Type = "sticker"
	type alias InlineQueryResultCachedSticker
	return json.Marshal(alias(result))
}

func (InlineQueryResultCachedVideo) isInlineQueryResult() {}
func (result InlineQueryResultCachedVideo) MarshalJSON() ([]byte, error) {
	result.Type = "video"
	type alias InlineQueryResultCachedVideo
	return json.Marshal(alias(result))
}

func (InlineQueryResultCachedVoice) isInlineQueryResult() {}
func (result InlineQueryResultCachedVoice) MarshalJSON() ([]byte, error) {
	result.Type = "voice"
	type alias InlineQueryResultCachedVoice
	return json.Marshal(alias(result))
}

func (InlineQueryResultAudio) isInlineQueryResult() {}
func (result InlineQueryResultAudio) MarshalJSON() ([]byte, error) {
	result.Type = "audio"
	type alias InlineQueryResultAudio
	return json.Marshal(alias(result))
}

func (InlineQueryResultDocument) isInlineQueryResult() {}
func (result InlineQueryResultDocument) MarshalJSON() ([]byte, error) {
	result.Type = "document"
	type alias InlineQueryResultDocument
	return json.Marshal(alias(result))
}

func (InlineQueryResultGif) isInlineQueryResult() {}
func (result InlineQueryResultGif) MarshalJSON() ([]byte, error) {
	result.Type = "gif"
	type alias InlineQueryResultGif
	return json.Marshal(alias(result))
}

func (InlineQueryResultMpeg4Gif) isInlineQueryResult() {}
func (result InlineQueryResultMpeg4Gif) MarshalJSON() ([]byte, error) {
	result.Type = "mpeg4_gif"
	type alias InlineQueryResultMpeg4Gif
	return json.Marshal(alias(result))
}

func (InlineQueryResultPhoto) isInlineQueryResult() {}
func (result InlineQueryResultPhoto) MarshalJSON() ([]byte, error) {
	result.Type = "photo"
	type alias InlineQueryResultPhoto
	return json.Marshal(alias(result))
}

func (InlineQueryResultVideo) isInlineQueryResult() {}
func (result InlineQueryResultVideo) MarshalJSON() ([]byte, error) {
	result.Type = "video"
	type alias InlineQueryResultVideo
	return json.Marshal(alias(result))
}

func (InlineQueryResultVoice) isInlineQueryResult() {}
func (result InlineQueryResultVoice) MarshalJSON() ([]byte, error) {
	result.Type = "voice"
	type alias InlineQueryResultVoice
	return json.Marshal(alias(result))
}

func (InlineQueryResultArticle) isInlineQueryResult() {}
func (result InlineQueryResultArticle) MarshalJSON() ([]byte, error) {
	result.Type = "article"
	type alias InlineQueryResultArticle
	return json.Marshal(alias(result))
}

func (InlineQueryResultContact) isInlineQueryResult() {}
func (result InlineQueryResultContact) MarshalJSON() ([]byte, error) {
	result.Type = "contact"
	type alias InlineQueryResultContact
	return json.Marshal(alias(result))
}

func (InlineQueryResultGame) isInlineQueryResult() {}
func (result InlineQueryResultGame) MarshalJSON() ([]byte, error) {
	result.Type = "game"
	type alias InlineQueryResultGame
	return json.Marshal(alias(result))
}

func (InlineQueryResultLocation) isInlineQueryResult() {}
func (result InlineQueryResultLocation) MarshalJSON() ([]byte, error) {
	result.Type = "location"
	type alias InlineQueryResultLocation
	return json.Marshal(alias(result))
}

func (InlineQueryResultVenue) isInlineQueryResult() {}
func (result InlineQueryResultVenue) MarshalJSON() ([]byte, error) {
	result.Type = "venue"
	type alias InlineQueryResultVenue
	return json.Marshal(alias(result))
}

type InputMessageContent interface {
	isInputMessageContent()
}

func (content InputTextMessageContent) isInputMessageContent()     {}
func (content InputLocationMessageContent) isInputMessageContent() {}
func (content InputVenueMessageContent) isInputMessageContent()    {}
func (content InputContactMessageContent) isInputMessageContent()  {}
func (content InputInvoiceMessageContent) isInputMessageContent()  {}
