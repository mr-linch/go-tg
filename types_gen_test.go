package tg

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInlineKeyboardButtonConstructors(t *testing.T) {
	for _, test := range []struct {
		Name   string
		Button InlineKeyboardButton
		Want   InlineKeyboardButton
	}{
		{
			Name:   "URL",
			Button: NewInlineKeyboardButtonURL("text", "https://example.com"),
			Want:   InlineKeyboardButton{Text: "text", URL: "https://example.com"},
		},
		{
			Name:   "CallbackData",
			Button: NewInlineKeyboardButtonCallbackData("text", "data"),
			Want:   InlineKeyboardButton{Text: "text", CallbackData: "data"},
		},
		{
			Name:   "WebApp",
			Button: NewInlineKeyboardButtonWebApp("text", WebAppInfo{URL: "https://example.com/app"}),
			Want:   InlineKeyboardButton{Text: "text", WebApp: &WebAppInfo{URL: "https://example.com/app"}},
		},
		{
			Name: "LoginURL",
			Button: NewInlineKeyboardButtonLoginURL("text", LoginURL{
				URL:                "https://example.com/login",
				ForwardText:        "forward",
				BotUsername:        "bot",
				RequestWriteAccess: true,
			}),
			Want: InlineKeyboardButton{Text: "text", LoginURL: &LoginURL{
				URL:                "https://example.com/login",
				ForwardText:        "forward",
				BotUsername:        "bot",
				RequestWriteAccess: true,
			}},
		},
		{
			Name:   "SwitchInlineQuery",
			Button: NewInlineKeyboardButtonSwitchInlineQuery("text", "query"),
			Want:   InlineKeyboardButton{Text: "text", SwitchInlineQuery: "query"},
		},
		{
			Name:   "SwitchInlineQueryCurrentChat",
			Button: NewInlineKeyboardButtonSwitchInlineQueryCurrentChat("text", "query"),
			Want:   InlineKeyboardButton{Text: "text", SwitchInlineQueryCurrentChat: "query"},
		},
		{
			Name:   "SwitchInlineQueryChosenChat",
			Button: NewInlineKeyboardButtonSwitchInlineQueryChosenChat("text", SwitchInlineQueryChosenChat{Query: "query", AllowUserChats: true}),
			Want:   InlineKeyboardButton{Text: "text", SwitchInlineQueryChosenChat: &SwitchInlineQueryChosenChat{Query: "query", AllowUserChats: true}},
		},
		{
			Name:   "CopyText",
			Button: NewInlineKeyboardButtonCopyText("text", CopyTextButton{Text: "copy this"}),
			Want:   InlineKeyboardButton{Text: "text", CopyText: &CopyTextButton{Text: "copy this"}},
		},
		{
			Name:   "CallbackGame",
			Button: NewInlineKeyboardButtonCallbackGame("text"),
			Want:   InlineKeyboardButton{Text: "text", CallbackGame: &CallbackGame{}},
		},
		{
			Name:   "Pay",
			Button: NewInlineKeyboardButtonPay("text"),
			Want:   InlineKeyboardButton{Text: "text", Pay: true},
		},
	} {
		t.Run(test.Name, func(t *testing.T) {
			assert.Equal(t, test.Want, test.Button)
		})
	}
}

func TestKeyboardButtonConstructors(t *testing.T) {
	for _, test := range []struct {
		Name   string
		Button KeyboardButton
		Want   KeyboardButton
	}{
		{
			Name:   "Base",
			Button: NewKeyboardButton("text"),
			Want:   KeyboardButton{Text: "text"},
		},
		{
			Name:   "RequestUsers",
			Button: NewKeyboardButtonRequestUsers("text", KeyboardButtonRequestUsers{RequestID: 1, UserIsBot: true}),
			Want:   KeyboardButton{Text: "text", RequestUsers: &KeyboardButtonRequestUsers{RequestID: 1, UserIsBot: true}},
		},
		{
			Name:   "RequestChat",
			Button: NewKeyboardButtonRequestChat("text", KeyboardButtonRequestChat{RequestID: 2, ChatIsChannel: true}),
			Want:   KeyboardButton{Text: "text", RequestChat: &KeyboardButtonRequestChat{RequestID: 2, ChatIsChannel: true}},
		},
		{
			Name:   "RequestContact",
			Button: NewKeyboardButtonRequestContact("text"),
			Want:   KeyboardButton{Text: "text", RequestContact: true},
		},
		{
			Name:   "RequestLocation",
			Button: NewKeyboardButtonRequestLocation("text"),
			Want:   KeyboardButton{Text: "text", RequestLocation: true},
		},
		{
			Name:   "RequestPoll",
			Button: NewKeyboardButtonRequestPoll("text", KeyboardButtonPollType{Type: "quiz"}),
			Want:   KeyboardButton{Text: "text", RequestPoll: &KeyboardButtonPollType{Type: "quiz"}},
		},
		{
			Name:   "WebApp",
			Button: NewKeyboardButtonWebApp("text", WebAppInfo{URL: "https://example.com/app"}),
			Want:   KeyboardButton{Text: "text", WebApp: &WebAppInfo{URL: "https://example.com/app"}},
		},
	} {
		t.Run(test.Name, func(t *testing.T) {
			assert.Equal(t, test.Want, test.Button)
		})
	}
}

func TestInlineQueryResultsButtonConstructors(t *testing.T) {
	for _, test := range []struct {
		Name   string
		Button InlineQueryResultsButton
		Want   InlineQueryResultsButton
	}{
		{
			Name:   "StartParameter",
			Button: NewInlineQueryResultsButtonStartParameter("text", "start-param"),
			Want:   InlineQueryResultsButton{Text: "text", StartParameter: "start-param"},
		},
		{
			Name:   "WebApp",
			Button: NewInlineQueryResultsButtonWebApp("text", WebAppInfo{URL: "https://example.com/app"}),
			Want:   InlineQueryResultsButton{Text: "text", WebApp: &WebAppInfo{URL: "https://example.com/app"}},
		},
	} {
		t.Run(test.Name, func(t *testing.T) {
			assert.Equal(t, test.Want, test.Button)
		})
	}
}
