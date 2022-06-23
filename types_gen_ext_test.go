package tg

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPeerIDImpl(t *testing.T) {
	for _, test := range []struct {
		PeerID PeerID
		Want   string
	}{
		{UserID(1), "1"},
		{ChatID(1), "1"},
		{&Chat{ID: ChatID(1)}, "1"},
	} {
		assert.Equal(t, test.Want, test.PeerID.PeerID())
	}
}

func TestChatType_String(t *testing.T) {
	for _, test := range []struct {
		ChatType ChatType
		Want     string
	}{
		{ChatTypePrivate, "private"},
		{ChatTypeGroup, "group"},
		{ChatTypeSupergroup, "supergroup"},
		{ChatTypeChannel, "channel"},
		{ChatTypeSender, "sender"},
		{ChatType(-1), "unknown"},
	} {
		assert.Equal(t, test.Want, test.ChatType.String())
	}
}

func TestChatType_MarshalJSON(t *testing.T) {
	type sample struct {
		Type ChatType `json:"type"`
	}
	for _, test := range []struct {
		Sample sample
		Want   string
	}{
		{sample{ChatTypePrivate}, `{"type":"private"}`},
		{sample{ChatTypeGroup}, `{"type":"group"}`},
		{sample{ChatTypeSupergroup}, `{"type":"supergroup"}`},
		{sample{ChatTypeChannel}, `{"type":"channel"}`},
		{sample{ChatTypeSender}, `{"type":"sender"}`},
		{sample{ChatType(-1)}, `{"type":"unknown"}`},
	} {
		actual, err := json.Marshal(test.Sample)
		assert.NoError(t, err)

		assert.Equal(t, test.Want, string(actual))
	}
}

func TestChatType_UnmarshalJSON(t *testing.T) {
	type sample struct {
		Type ChatType `json:"type"`
	}
	for _, test := range []struct {
		Input  string
		Sample sample
		Want   ChatType
		Err    bool
	}{
		{`{"type": "private"}`, sample{}, ChatTypePrivate, false},
		{`{"type": "group"}`, sample{}, ChatTypeGroup, false},
		{`{"type": "supergroup"}`, sample{}, ChatTypeSupergroup, false},
		{`{"type": "channel"}`, sample{}, ChatTypeChannel, false},
		{`{"type": "test"}`, sample{}, ChatType(-1), true},
		{`{"type": "sender"}`, sample{}, ChatTypeSender, false},
		{`{"type": {}}`, sample{}, ChatType(-1), true},
	} {
		if test.Err {
			assert.Error(t, json.Unmarshal([]byte(test.Input), &test.Sample))
		} else {
			assert.NoError(t, json.Unmarshal([]byte(test.Input), &test.Sample))
			assert.Equal(t, test.Want, test.Sample.Type)
		}
	}
}

func TestInlineReplyMarkup(t *testing.T) {
	actual := NewInlineKeyboardMarkup(
		NewInlineKeyboardRow(
			NewInlineKeyboardButtonURL("text", "https://google.com"),
			NewInlineKeyboardButtonCallback("text", "data"),
			NewInlineKeyboardButtonWebApp("text", WebAppInfo{}),
			NewInlineKeyboardButtonLoginURL("text", LoginUrl{
				URL: "https://google.com",
			}),
			NewInlineKeyboardButtonSwitchInlineQuery("text", "query"),
			NewInlineKeyboardButtonSwitchInlineQueryCurrentChat("text", "query"),
			NewInlineKeyboardButtonCallbackGame("text"),
			NewInlineKeyboardButtonPay("text"),
		),
	)

	actual.isReplyMarkup()

	assert.EqualValues(t, &InlineKeyboardMarkup{
		InlineKeyboard: [][]InlineKeyboardButton{
			{
				{Text: "text", URL: "https://google.com"},
				{Text: "text", CallbackData: "data"},
				{Text: "text", WebApp: &WebAppInfo{}},
				{Text: "text", LoginURL: &LoginUrl{URL: "https://google.com"}},
				{Text: "text", SwitchInlineQuery: "query"},
				{Text: "text", SwitchInlineQueryCurrentChat: "query"},
				{Text: "text", CallbackGame: &CallbackGame{}},
				{Text: "text", Pay: true},
			},
		},
	}, actual)
}

func TestReplyKeyboardMarkup(t *testing.T) {
	actual := NewReplyKeyboardMarkup(
		NewReplyKeyboardRow(
			NewKeyboardButton("text"),
			NewKeyboardButtonRequestContact("text"),
			NewKeyboardButtonRequestLocation("text"),
			NewKeyboardButtonRequestPoll("text", KeyboardButtonPollType{}),
			NewKeyboardButtonWebApp("text", WebAppInfo{}),
		),
	).WithResizeKeyboardMarkup().
		WithOneTimeKeyboardMarkup().
		WithInputFieldPlaceholder("text").
		WithSelective()

	actual.isReplyMarkup()

	assert.EqualValues(t, &ReplyKeyboardMarkup{
		Keyboard: [][]KeyboardButton{
			{
				{Text: "text"},
				{Text: "text", RequestContact: true},
				{Text: "text", RequestLocation: true},
				{Text: "text", RequestPoll: &KeyboardButtonPollType{}},
				{Text: "text", WebApp: &WebAppInfo{}},
			},
		},
		ResizeKeyboard:        true,
		OneTimeKeyboard:       true,
		InputFieldPlaceholder: "text",
		Selective:             true,
	}, actual)
}

func TestReplyKeyboardRemove(t *testing.T) {
	actual := NewReplyKeyboardRemove().WithSelective()

	actual.isReplyMarkup()

	assert.EqualValues(t, &ReplyKeyboardRemove{
		RemoveKeyboard: true,
		Selective:      true,
	}, actual)
}

func TestForceReplay(t *testing.T) {
	actual := NewForceReply().WithSelective().WithInputFieldPlaceholder("test")

	actual.isReplyMarkup()

	assert.EqualValues(t, &ForceReply{
		ForceReply:            true,
		Selective:             true,
		InputFieldPlaceholder: "test",
	}, actual)
}
