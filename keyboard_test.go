package tg

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInlineKeyboard(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		kb := NewInlineKeyboard()
		m := kb.Markup()
		assert.Nil(t, m.InlineKeyboard)
	})

	t.Run("SingleRow", func(t *testing.T) {
		kb := NewInlineKeyboard().
			Callback("A", "a").Callback("B", "b")
		m := kb.Markup()
		assert.Equal(t, [][]InlineKeyboardButton{
			{
				{Text: "A", CallbackData: "a"},
				{Text: "B", CallbackData: "b"},
			},
		}, m.InlineKeyboard)
	})

	t.Run("MultipleRows", func(t *testing.T) {
		kb := NewInlineKeyboard().
			Callback("A", "a").Row().
			Callback("B", "b").Row().
			Callback("C", "c")
		m := kb.Markup()
		assert.Equal(t, [][]InlineKeyboardButton{
			{{Text: "A", CallbackData: "a"}},
			{{Text: "B", CallbackData: "b"}},
			{{Text: "C", CallbackData: "c"}},
		}, m.InlineKeyboard)
	})

	t.Run("RowAtEnd", func(t *testing.T) {
		kb := NewInlineKeyboard().
			Callback("A", "a").Row()
		m := kb.Markup()
		assert.Equal(t, [][]InlineKeyboardButton{
			{{Text: "A", CallbackData: "a"}},
		}, m.InlineKeyboard)
	})

	t.Run("AdjustOne", func(t *testing.T) {
		kb := NewInlineKeyboard().
			Callback("A", "a").Callback("B", "b").Callback("C", "c").
			Adjust(1)
		m := kb.Markup()
		assert.Equal(t, [][]InlineKeyboardButton{
			{{Text: "A", CallbackData: "a"}},
			{{Text: "B", CallbackData: "b"}},
			{{Text: "C", CallbackData: "c"}},
		}, m.InlineKeyboard)
	})

	t.Run("AdjustTwo", func(t *testing.T) {
		kb := NewInlineKeyboard().
			Callback("A", "a").Callback("B", "b").Callback("C", "c").
			Adjust(2)
		m := kb.Markup()
		assert.Equal(t, [][]InlineKeyboardButton{
			{
				{Text: "A", CallbackData: "a"},
				{Text: "B", CallbackData: "b"},
			},
			{{Text: "C", CallbackData: "c"}},
		}, m.InlineKeyboard)
	})

	t.Run("AdjustRepeatingPattern", func(t *testing.T) {
		kb := NewInlineKeyboard().
			Callback("1", "1").Callback("2", "2").Callback("3", "3").
			Callback("4", "4").Callback("5", "5").Callback("6", "6").
			Adjust(3, 1)
		m := kb.Markup()
		assert.Equal(t, [][]InlineKeyboardButton{
			{
				{Text: "1", CallbackData: "1"},
				{Text: "2", CallbackData: "2"},
				{Text: "3", CallbackData: "3"},
			},
			{{Text: "4", CallbackData: "4"}},
			{
				{Text: "5", CallbackData: "5"},
				{Text: "6", CallbackData: "6"},
			},
		}, m.InlineKeyboard)
	})

	t.Run("AdjustDoesNotAffectCommittedRows", func(t *testing.T) {
		kb := NewInlineKeyboard().
			Callback("Fixed", "f").Row().
			Callback("A", "a").Callback("B", "b").Callback("C", "c").
			Adjust(2)
		m := kb.Markup()
		assert.Equal(t, [][]InlineKeyboardButton{
			{{Text: "Fixed", CallbackData: "f"}},
			{
				{Text: "A", CallbackData: "a"},
				{Text: "B", CallbackData: "b"},
			},
			{{Text: "C", CallbackData: "c"}},
		}, m.InlineKeyboard)
	})

	t.Run("AdjustWithRemainder", func(t *testing.T) {
		kb := NewInlineKeyboard().
			Callback("1", "1").Callback("2", "2").Callback("3", "3").
			Callback("4", "4").Callback("5", "5").
			Adjust(3)
		m := kb.Markup()
		assert.Equal(t, [][]InlineKeyboardButton{
			{
				{Text: "1", CallbackData: "1"},
				{Text: "2", CallbackData: "2"},
				{Text: "3", CallbackData: "3"},
			},
			{
				{Text: "4", CallbackData: "4"},
				{Text: "5", CallbackData: "5"},
			},
		}, m.InlineKeyboard)
	})

	t.Run("AdjustNoArgs", func(t *testing.T) {
		kb := NewInlineKeyboard().
			Callback("A", "a").Callback("B", "b").
			Adjust()
		m := kb.Markup()
		assert.Equal(t, [][]InlineKeyboardButton{
			{{Text: "A", CallbackData: "a"}},
			{{Text: "B", CallbackData: "b"}},
		}, m.InlineKeyboard)
	})

	t.Run("AdjustEmpty", func(t *testing.T) {
		kb := NewInlineKeyboard().
			Callback("A", "a").Row().
			Adjust(2)
		m := kb.Markup()
		assert.Equal(t, [][]InlineKeyboardButton{
			{{Text: "A", CallbackData: "a"}},
		}, m.InlineKeyboard)
	})

	t.Run("StaticDynamicStatic", func(t *testing.T) {
		kb := NewInlineKeyboard().
			Callback("A", "a").Callback("B", "b").Callback("C", "c").Row()

		for i, name := range []string{"I1", "I2", "I3", "I4", "I5", "I6"} {
			kb.Callback(name, "item:"+string(rune('0'+i)))
		}
		kb.Adjust(4)

		kb.Callback("Back", "back")

		m := kb.Markup()
		assert.Equal(t, [][]InlineKeyboardButton{
			{
				{Text: "A", CallbackData: "a"},
				{Text: "B", CallbackData: "b"},
				{Text: "C", CallbackData: "c"},
			},
			{
				{Text: "I1", CallbackData: "item:0"},
				{Text: "I2", CallbackData: "item:1"},
				{Text: "I3", CallbackData: "item:2"},
				{Text: "I4", CallbackData: "item:3"},
			},
			{
				{Text: "I5", CallbackData: "item:4"},
				{Text: "I6", CallbackData: "item:5"},
			},
			{{Text: "Back", CallbackData: "back"}},
		}, m.InlineKeyboard)
	})

	t.Run("MultipleAdjustBlocks", func(t *testing.T) {
		kb := NewInlineKeyboard()

		kb.Callback("A", "a").Callback("B", "b").Callback("C", "c").Callback("D", "d")
		kb.Adjust(2)

		kb.Callback("X", "x").Callback("Y", "y").Callback("Z", "z")
		kb.Adjust(3)

		m := kb.Markup()
		assert.Equal(t, [][]InlineKeyboardButton{
			{
				{Text: "A", CallbackData: "a"},
				{Text: "B", CallbackData: "b"},
			},
			{
				{Text: "C", CallbackData: "c"},
				{Text: "D", CallbackData: "d"},
			},
			{
				{Text: "X", CallbackData: "x"},
				{Text: "Y", CallbackData: "y"},
				{Text: "Z", CallbackData: "z"},
			},
		}, m.InlineKeyboard)
	})

	t.Run("ButtonWithPreBuilt", func(t *testing.T) {
		btn := NewInlineKeyboardButtonCallbackData("Pre", "pre")
		kb := NewInlineKeyboard().
			Button(btn).
			Callback("After", "after")
		m := kb.Markup()
		assert.Equal(t, [][]InlineKeyboardButton{
			{
				{Text: "Pre", CallbackData: "pre"},
				{Text: "After", CallbackData: "after"},
			},
		}, m.InlineKeyboard)
	})

	t.Run("AllButtonTypes", func(t *testing.T) {
		kb := NewInlineKeyboard().
			Callback("cb", "data").Row().
			URL("url", "https://example.com").Row().
			WebApp("webapp", "https://example.com/app").Row().
			LoginURL("login", LoginURL{URL: "https://example.com/login"}).Row().
			SwitchInlineQuery("siq", "query").Row().
			SwitchInlineQueryCurrentChat("siqcc", "query2").Row().
			SwitchInlineQueryChosenChat("siqc", SwitchInlineQueryChosenChat{AllowUserChats: true}).Row().
			CopyText("copy", CopyTextButton{Text: "copied"}).Row().
			Pay("pay").Row().
			CallbackGame("game")

		m := kb.Markup()
		require.Len(t, m.InlineKeyboard, 10)
		assert.Equal(t, "data", m.InlineKeyboard[0][0].CallbackData)
		assert.Equal(t, "https://example.com", m.InlineKeyboard[1][0].URL)
		assert.Equal(t, "https://example.com/app", m.InlineKeyboard[2][0].WebApp.URL)
		assert.Equal(t, "https://example.com/login", m.InlineKeyboard[3][0].LoginURL.URL)
		assert.Equal(t, "query", m.InlineKeyboard[4][0].SwitchInlineQuery)
		assert.Equal(t, "query2", m.InlineKeyboard[5][0].SwitchInlineQueryCurrentChat)
		assert.True(t, m.InlineKeyboard[6][0].SwitchInlineQueryChosenChat.AllowUserChats)
		assert.Equal(t, "copied", m.InlineKeyboard[7][0].CopyText.Text)
		assert.True(t, m.InlineKeyboard[8][0].Pay)
		assert.NotNil(t, m.InlineKeyboard[9][0].CallbackGame)
	})

	t.Run("MarshalJSON", func(t *testing.T) {
		kb := NewInlineKeyboard().
			Callback("A", "a").Callback("B", "b").Row().
			Callback("C", "c")

		kbJSON, err := json.Marshal(kb)
		require.NoError(t, err)

		markupJSON, err := json.Marshal(kb.Markup())
		require.NoError(t, err)

		assert.JSONEq(t, string(markupJSON), string(kbJSON))
	})

	t.Run("ReplyMarkupInterface", func(t *testing.T) {
		var _ ReplyMarkup = NewInlineKeyboard()
	})
}

func TestReplyKeyboard(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		kb := NewReplyKeyboard()
		m := kb.Markup()
		assert.Nil(t, m.Keyboard)
	})

	t.Run("TextButtons", func(t *testing.T) {
		kb := NewReplyKeyboard().
			Text("A").Text("B").Row().
			Text("C")
		m := kb.Markup()
		assert.Equal(t, [][]KeyboardButton{
			{{Text: "A"}, {Text: "B"}},
			{{Text: "C"}},
		}, m.Keyboard)
	})

	t.Run("Options", func(t *testing.T) {
		kb := NewReplyKeyboard().
			Text("A").
			Resize().OneTime().Persistent().Selective().Placeholder("Choose...")
		m := kb.Markup()
		assert.True(t, m.ResizeKeyboard)
		assert.True(t, m.OneTimeKeyboard)
		assert.True(t, m.IsPersistent)
		assert.True(t, m.Selective)
		assert.Equal(t, "Choose...", m.InputFieldPlaceholder)
	})

	t.Run("Adjust", func(t *testing.T) {
		kb := NewReplyKeyboard().
			Text("1").Text("2").Text("3").Text("4").Text("5").
			Adjust(2)
		m := kb.Markup()
		assert.Equal(t, [][]KeyboardButton{
			{{Text: "1"}, {Text: "2"}},
			{{Text: "3"}, {Text: "4"}},
			{{Text: "5"}},
		}, m.Keyboard)
	})

	t.Run("SpecialButtons", func(t *testing.T) {
		kb := NewReplyKeyboard().
			RequestContact("Phone").Row().
			RequestLocation("Location").Row().
			RequestPoll("Poll", KeyboardButtonPollType{Type: PollTypeQuiz}).Row().
			RequestUsers("Users", KeyboardButtonRequestUsers{RequestID: 1}).Row().
			RequestChat("Chat", KeyboardButtonRequestChat{RequestID: 2}).Row().
			WebApp("App", "https://example.com")
		m := kb.Markup()
		require.Len(t, m.Keyboard, 6)
		assert.True(t, m.Keyboard[0][0].RequestContact)
		assert.True(t, m.Keyboard[1][0].RequestLocation)
		assert.Equal(t, PollTypeQuiz, m.Keyboard[2][0].RequestPoll.Type)
		assert.Equal(t, 1, m.Keyboard[3][0].RequestUsers.RequestID)
		assert.Equal(t, 2, m.Keyboard[4][0].RequestChat.RequestID)
		assert.Equal(t, "https://example.com", m.Keyboard[5][0].WebApp.URL)
	})

	t.Run("MarshalJSON", func(t *testing.T) {
		kb := NewReplyKeyboard().
			Text("A").Text("B").
			Resize().OneTime()

		kbJSON, err := json.Marshal(kb)
		require.NoError(t, err)

		markupJSON, err := json.Marshal(kb.Markup())
		require.NoError(t, err)

		assert.JSONEq(t, string(markupJSON), string(kbJSON))
	})

	t.Run("ReplyMarkupInterface", func(t *testing.T) {
		var _ ReplyMarkup = NewReplyKeyboard()
	})
}
