package tg

import "encoding/json"

// Button define generic button interface.
type Button interface {
	InlineKeyboardButton | KeyboardButton
}

// NewButtonRow it's generic helper for create keyboards in functional way.
func NewButtonRow[T Button](buttons ...T) []T {
	return buttons
}

// Deprecated: Use [InlineKeyboard] or [ReplyKeyboard] instead.
type ButtonLayout[T Button] struct {
	buttons  [][]T
	rowWidth int
}

// Deprecated: Use [NewInlineKeyboard] or [NewReplyKeyboard] with [InlineKeyboard.Adjust] instead.
func NewButtonColumn[T Button](buttons ...T) [][]T {
	result := make([][]T, 0, len(buttons))

	for _, button := range buttons {
		result = append(result, []T{button})
	}

	return result
}

// Deprecated: Use [NewInlineKeyboard] or [NewReplyKeyboard] instead.
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

// Row add new row with no respect for row width.
func (layout *ButtonLayout[T]) Row(buttons ...T) *ButtonLayout[T] {
	layout.buttons = append(layout.buttons, buttons)
	return layout
}

// keyboard is a generic core for building button keyboards.
type keyboard[T Button] struct {
	rows    [][]T
	current []T
}

func (b *keyboard[T]) addButtons(buttons ...T) {
	b.current = append(b.current, buttons...)
}

func (b *keyboard[T]) addRow() {
	if len(b.current) == 0 {
		return
	}
	b.rows = append(b.rows, b.current)
	b.current = nil
}

func (b *keyboard[T]) doAdjust(sizes ...int) {
	buttons := b.current
	b.current = nil

	if len(buttons) == 0 {
		return
	}

	if len(sizes) == 0 {
		sizes = []int{1}
	}

	i, si := 0, 0
	for i < len(buttons) {
		size := sizes[si%len(sizes)]
		end := min(i+size, len(buttons))
		b.rows = append(b.rows, buttons[i:end])
		i = end
		si++
	}
}

func (b *keyboard[T]) build() [][]T {
	if len(b.current) > 0 {
		b.rows = append(b.rows, b.current)
		b.current = nil
	}
	return b.rows
}

// InlineKeyboard is a builder for inline keyboards with a fluent API.
//
// Both [InlineKeyboard.Row] and [InlineKeyboard.Adjust] commit uncommitted buttons
// added via [InlineKeyboard.Button] or shorthand methods like [InlineKeyboard.Callback].
// Previously committed rows are never affected.
//
// InlineKeyboard implements [ReplyMarkup] and can be passed directly to .ReplyMarkup().
// Use [InlineKeyboard.Markup] when you need the underlying [InlineKeyboardMarkup] value.
type InlineKeyboard struct {
	keyboard[InlineKeyboardButton]
}

// NewInlineKeyboard creates a new [InlineKeyboard] builder.
func NewInlineKeyboard() *InlineKeyboard {
	return &InlineKeyboard{}
}

// Button adds pre-built [InlineKeyboardButton] values to the current (uncommitted) row.
func (b *InlineKeyboard) Button(buttons ...InlineKeyboardButton) *InlineKeyboard {
	b.addButtons(buttons...)
	return b
}

// Row commits the current buttons as a completed row.
// Subsequent buttons go into a new row.
func (b *InlineKeyboard) Row() *InlineKeyboard {
	b.addRow()
	return b
}

// Adjust redistributes uncommitted buttons into rows following
// a repeating size pattern. For example, Adjust(2,1) produces
// rows of 2, 1, 2, 1, … buttons. Empty sizes default to [1].
// Previously committed rows are not affected.
func (b *InlineKeyboard) Adjust(sizes ...int) *InlineKeyboard {
	b.doAdjust(sizes...)
	return b
}

// Markup flushes any remaining uncommitted buttons as the last row
// and returns the resulting [InlineKeyboardMarkup].
func (b *InlineKeyboard) Markup() InlineKeyboardMarkup {
	return InlineKeyboardMarkup{
		InlineKeyboard: b.build(),
	}
}

func (*InlineKeyboard) isReplyMarkup() {}

// MarshalJSON implements [json.Marshaler].
// It serializes the builder as an [InlineKeyboardMarkup].
func (b *InlineKeyboard) MarshalJSON() ([]byte, error) {
	return json.Marshal(b.Markup())
}

// Callback adds a callback button with the given text and data.
func (b *InlineKeyboard) Callback(text, callbackData string) *InlineKeyboard {
	b.addButtons(InlineKeyboardButton{Text: text, CallbackData: callbackData})
	return b
}

// URL adds a URL button.
func (b *InlineKeyboard) URL(text, url string) *InlineKeyboard {
	b.addButtons(InlineKeyboardButton{Text: text, URL: url})
	return b
}

// WebApp adds a Web App button.
func (b *InlineKeyboard) WebApp(text, url string) *InlineKeyboard {
	b.addButtons(InlineKeyboardButton{Text: text, WebApp: &WebAppInfo{URL: url}})
	return b
}

// LoginURL adds a login URL button.
func (b *InlineKeyboard) LoginURL(text string, loginURL LoginURL) *InlineKeyboard {
	b.addButtons(InlineKeyboardButton{Text: text, LoginURL: &loginURL})
	return b
}

// SwitchInlineQuery adds a switch inline query button.
func (b *InlineKeyboard) SwitchInlineQuery(text, query string) *InlineKeyboard {
	b.addButtons(InlineKeyboardButton{Text: text, SwitchInlineQuery: query})
	return b
}

// SwitchInlineQueryCurrentChat adds a switch inline query button for the current chat.
func (b *InlineKeyboard) SwitchInlineQueryCurrentChat(text, query string) *InlineKeyboard {
	b.addButtons(InlineKeyboardButton{Text: text, SwitchInlineQueryCurrentChat: query})
	return b
}

// SwitchInlineQueryChosenChat adds a switch inline query button with chat selection.
func (b *InlineKeyboard) SwitchInlineQueryChosenChat(text string, chosen SwitchInlineQueryChosenChat) *InlineKeyboard {
	b.addButtons(InlineKeyboardButton{Text: text, SwitchInlineQueryChosenChat: &chosen})
	return b
}

// CopyText adds a copy-text button.
func (b *InlineKeyboard) CopyText(text string, copyText CopyTextButton) *InlineKeyboard {
	b.addButtons(InlineKeyboardButton{Text: text, CopyText: &copyText})
	return b
}

// Pay adds a pay button.
func (b *InlineKeyboard) Pay(text string) *InlineKeyboard {
	b.addButtons(InlineKeyboardButton{Text: text, Pay: true})
	return b
}

// CallbackGame adds a callback game button.
func (b *InlineKeyboard) CallbackGame(text string) *InlineKeyboard {
	b.addButtons(InlineKeyboardButton{Text: text, CallbackGame: &CallbackGame{}})
	return b
}

// ReplyKeyboard is a builder for reply keyboards with a fluent API.
//
// Both [ReplyKeyboard.Row] and [ReplyKeyboard.Adjust] commit uncommitted buttons
// added via [ReplyKeyboard.Button] or shorthand methods like [ReplyKeyboard.Text].
// Previously committed rows are never affected.
//
// ReplyKeyboard implements [ReplyMarkup] and can be passed directly to .ReplyMarkup().
// Use [ReplyKeyboard.Markup] when you need the underlying [ReplyKeyboardMarkup] value.
type ReplyKeyboard struct {
	keyboard[KeyboardButton]
	resize      bool
	oneTime     bool
	persistent  bool
	selective   bool
	placeholder string
}

// NewReplyKeyboard creates a new [ReplyKeyboard] builder.
func NewReplyKeyboard() *ReplyKeyboard {
	return &ReplyKeyboard{}
}

// Button adds pre-built [KeyboardButton] values to the current (uncommitted) row.
func (b *ReplyKeyboard) Button(buttons ...KeyboardButton) *ReplyKeyboard {
	b.addButtons(buttons...)
	return b
}

// Row commits the current buttons as a completed row.
// Subsequent buttons go into a new row.
func (b *ReplyKeyboard) Row() *ReplyKeyboard {
	b.addRow()
	return b
}

// Adjust redistributes uncommitted buttons into rows following
// a repeating size pattern. For example, Adjust(2,1) produces
// rows of 2, 1, 2, 1, … buttons. Empty sizes default to [1].
// Previously committed rows are not affected.
func (b *ReplyKeyboard) Adjust(sizes ...int) *ReplyKeyboard {
	b.doAdjust(sizes...)
	return b
}

// Resize requests clients to resize the keyboard vertically for optimal fit.
func (b *ReplyKeyboard) Resize() *ReplyKeyboard {
	b.resize = true
	return b
}

// OneTime requests clients to hide the keyboard as soon as it's been used.
func (b *ReplyKeyboard) OneTime() *ReplyKeyboard {
	b.oneTime = true
	return b
}

// Persistent requests clients to always show the keyboard.
func (b *ReplyKeyboard) Persistent() *ReplyKeyboard {
	b.persistent = true
	return b
}

// Selective shows the keyboard to specific users only.
func (b *ReplyKeyboard) Selective() *ReplyKeyboard {
	b.selective = true
	return b
}

// Placeholder sets the placeholder text shown in the input field; 1-64 characters.
func (b *ReplyKeyboard) Placeholder(text string) *ReplyKeyboard {
	b.placeholder = text
	return b
}

// Markup flushes any remaining uncommitted buttons as the last row
// and returns the resulting [ReplyKeyboardMarkup].
func (b *ReplyKeyboard) Markup() *ReplyKeyboardMarkup {
	return &ReplyKeyboardMarkup{
		Keyboard:              b.build(),
		ResizeKeyboard:        b.resize,
		OneTimeKeyboard:       b.oneTime,
		IsPersistent:          b.persistent,
		Selective:             b.selective,
		InputFieldPlaceholder: b.placeholder,
	}
}

func (*ReplyKeyboard) isReplyMarkup() {}

// MarshalJSON implements [json.Marshaler].
// It serializes the builder as a [ReplyKeyboardMarkup].
func (b *ReplyKeyboard) MarshalJSON() ([]byte, error) {
	return json.Marshal(b.Markup())
}

// Text adds a simple text button.
func (b *ReplyKeyboard) Text(text string) *ReplyKeyboard {
	b.addButtons(KeyboardButton{Text: text})
	return b
}

// RequestContact adds a button that sends the user's phone number.
func (b *ReplyKeyboard) RequestContact(text string) *ReplyKeyboard {
	b.addButtons(KeyboardButton{Text: text, RequestContact: true})
	return b
}

// RequestLocation adds a button that sends the user's current location.
func (b *ReplyKeyboard) RequestLocation(text string) *ReplyKeyboard {
	b.addButtons(KeyboardButton{Text: text, RequestLocation: true})
	return b
}

// RequestPoll adds a button that lets the user create and send a poll.
func (b *ReplyKeyboard) RequestPoll(text string, pollType KeyboardButtonPollType) *ReplyKeyboard {
	b.addButtons(KeyboardButton{Text: text, RequestPoll: &pollType})
	return b
}

// RequestUsers adds a button that lets the user select users.
func (b *ReplyKeyboard) RequestUsers(text string, request KeyboardButtonRequestUsers) *ReplyKeyboard {
	b.addButtons(KeyboardButton{Text: text, RequestUsers: &request})
	return b
}

// RequestChat adds a button that lets the user select a chat.
func (b *ReplyKeyboard) RequestChat(text string, request KeyboardButtonRequestChat) *ReplyKeyboard {
	b.addButtons(KeyboardButton{Text: text, RequestChat: &request})
	return b
}

// WebApp adds a Web App button.
func (b *ReplyKeyboard) WebApp(text, url string) *ReplyKeyboard {
	b.addButtons(KeyboardButton{Text: text, WebApp: &WebAppInfo{URL: url}})
	return b
}
