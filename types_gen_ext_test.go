package tg

import (
	"encoding"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnixTime(t *testing.T) {
	t.Run("Time", func(t *testing.T) {
		assert.Equal(t, time.Time{}, UnixTime(0).Time())
		assert.Equal(t, time.Unix(1234567890, 0), UnixTime(1234567890).Time())
		assert.Equal(t, time.Unix(-1, 0), UnixTime(-1).Time())
	})

	t.Run("IsZero", func(t *testing.T) {
		assert.True(t, UnixTime(0).IsZero())
		assert.False(t, UnixTime(1).IsZero())
		assert.False(t, UnixTime(-1).IsZero())
	})

	t.Run("JSON", func(t *testing.T) {
		type s struct {
			Date     UnixTime `json:"date"`
			Optional UnixTime `json:"optional,omitempty"`
		}

		// Unmarshal
		var v s
		err := json.Unmarshal([]byte(`{"date":1234567890}`), &v)
		require.NoError(t, err)
		assert.Equal(t, UnixTime(1234567890), v.Date)
		assert.True(t, v.Optional.IsZero())

		// Marshal
		data, err := json.Marshal(s{Date: UnixTime(1234567890)})
		require.NoError(t, err)
		assert.Equal(t, `{"date":1234567890}`, string(data))

		// Marshal with omitempty: zero value is omitted
		data, err = json.Marshal(s{Date: UnixTime(42)})
		require.NoError(t, err)
		assert.Equal(t, `{"date":42}`, string(data))
	})
}

func TestPeerIDImpl(t *testing.T) {
	for _, test := range []struct {
		PeerID PeerID
		Want   string
	}{
		{UserID(1), "1"},
		{ChatID(1), "1"},
		{&Chat{ID: ChatID(1)}, "1"},
		{&User{ID: UserID(1)}, "1"},
	} {
		assert.Equal(t, test.Want, test.PeerID.PeerID())
	}
}

func TestUsername_PeerID(t *testing.T) {
	assert.Equal(t, "@username", Username("username").PeerID())
}

func TestUsername_Link(t *testing.T) {
	assert.Equal(t, "https://t.me/username", Username("username").Link())
}

func TestUsername_DeepLink(t *testing.T) {
	assert.Equal(t, "tg://resolve?domain=username", Username("username").DeepLink())
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

func TestChatAction_String(t *testing.T) {
	for _, test := range []struct {
		ChatAction ChatAction
		Want       string
	}{
		{ChatActionTyping, "typing"},
		{ChatActionUploadPhoto, "upload_photo"},
		{ChatActionUploadVideo, "upload_video"},
		{ChatActionRecordVideo, "record_video"},
		{ChatActionRecordVoice, "record_voice"},
		{ChatActionUploadVoice, "upload_voice"},
		{ChatActionUploadDocument, "upload_document"},
		{ChatActionChooseSticker, "choose_sticker"},
		{ChatActionFindLocation, "find_location"},
		{ChatActionRecordVideoNote, "record_video_note"},
		{ChatActionUploadVideoNote, "upload_video_note"},
		{ChatAction(-1), "unknown"},
	} {
		assert.Equal(t, test.Want, test.ChatAction.String())
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
		require.NoError(t, err)

		assert.Equal(t, test.Want, string(actual))
	}
}

func TestChatType_UnmarshalJSON(t *testing.T) {
	type sample struct {
		Type ChatType `json:"type"`
	}
	tests := []struct {
		Input string
		Want  ChatType
		Err   bool
	}{
		{`{"type": "private"}`, ChatTypePrivate, false},
		{`{"type": "group"}`, ChatTypeGroup, false},
		{`{"type": "supergroup"}`, ChatTypeSupergroup, false},
		{`{"type": "channel"}`, ChatTypeChannel, false},
		{`{"type": "unknown_future"}`, ChatType(0), false}, // forward compatibility
		{`{"type": "sender"}`, ChatTypeSender, false},
		{`{"type": {}}`, ChatType(-1), true}, // invalid JSON type
	}
	for _, tt := range tests {
		var s sample
		if tt.Err {
			require.Error(t, json.Unmarshal([]byte(tt.Input), &s))
		} else {
			require.NoError(t, json.Unmarshal([]byte(tt.Input), &s))
			assert.Equal(t, tt.Want, s.Type)
		}
	}
}

func TestInlineReplyMarkup(t *testing.T) {
	actual := NewInlineKeyboardMarkup(
		NewButtonRow(
			NewInlineKeyboardButtonURL("text", "https://google.com"),
			NewInlineKeyboardButtonCallbackData("text", "data"),
			NewInlineKeyboardButtonWebApp("text", WebAppInfo{}),
			NewInlineKeyboardButtonLoginURL("text", LoginURL{
				URL: "https://google.com",
			}),
			NewInlineKeyboardButtonSwitchInlineQuery("text", "query"),
			NewInlineKeyboardButtonSwitchInlineQueryCurrentChat("text", "query"),
			NewInlineKeyboardButtonCallbackGame("text"),
			NewInlineKeyboardButtonPay("text"),
		),
	)

	actual.isReplyMarkup()

	assert.Equal(t, InlineKeyboardMarkup{
		InlineKeyboard: [][]InlineKeyboardButton{
			{
				{Text: "text", URL: "https://google.com"},
				{Text: "text", CallbackData: "data"},
				{Text: "text", WebApp: &WebAppInfo{}},
				{Text: "text", LoginURL: &LoginURL{URL: "https://google.com"}},
				{Text: "text", SwitchInlineQuery: "query"},
				{Text: "text", SwitchInlineQueryCurrentChat: "query"},
				{Text: "text", CallbackGame: &CallbackGame{}},
				{Text: "text", Pay: true},
			},
		},
	}, actual)

	assert.Equal(t, actual, *actual.Ptr())
}

func TestReplyKeyboardMarkup(t *testing.T) {
	actual := NewReplyKeyboardMarkup(
		NewButtonRow(
			NewKeyboardButton("text"),
			NewKeyboardButtonRequestContact("text"),
			NewKeyboardButtonRequestLocation("text"),
			NewKeyboardButtonRequestPoll("text", KeyboardButtonPollType{}),
			NewKeyboardButtonWebApp("text", WebAppInfo{}),
			NewKeyboardButtonRequestChat("test", KeyboardButtonRequestChat{RequestID: 1}),
			NewKeyboardButtonRequestUsers("text", KeyboardButtonRequestUsers{RequestID: 1}),
		),
	).WithResizeKeyboardMarkup().
		WithOneTimeKeyboardMarkup().
		WithInputFieldPlaceholder("text").
		WithSelective()

	actual.isReplyMarkup()

	assert.Equal(t, &ReplyKeyboardMarkup{
		Keyboard: [][]KeyboardButton{
			{
				{Text: "text"},
				{Text: "text", RequestContact: true},
				{Text: "text", RequestLocation: true},
				{Text: "text", RequestPoll: &KeyboardButtonPollType{}},
				{Text: "text", WebApp: &WebAppInfo{}},
				{Text: "test", RequestChat: &KeyboardButtonRequestChat{RequestID: 1}},
				{Text: "text", RequestUsers: &KeyboardButtonRequestUsers{RequestID: 1}},
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

	assert.Equal(t, &ReplyKeyboardRemove{
		RemoveKeyboard: true,
		Selective:      true,
	}, actual)
}

func TestForceReplay(t *testing.T) {
	actual := NewForceReply().WithSelective().WithInputFieldPlaceholder("test")

	actual.isReplyMarkup()

	assert.Equal(t, &ForceReply{
		ForceReply:            true,
		Selective:             true,
		InputFieldPlaceholder: "test",
	}, actual)
}

func TestNewButtonLayout(t *testing.T) {
	keyboard := NewButtonLayout(1,
		NewInlineKeyboardButtonCallbackData("1", "1"),
		NewInlineKeyboardButtonCallbackData("2", "2"),
		NewInlineKeyboardButtonCallbackData("3", "3"),
	).Keyboard()

	assert.Equal(t, [][]InlineKeyboardButton{
		{{Text: "1", CallbackData: "1"}},
		{{Text: "2", CallbackData: "2"}},
		{{Text: "3", CallbackData: "3"}},
	}, keyboard)
}

func TestButtonLayout_Add(t *testing.T) {
	for _, test := range []struct {
		Layout *ButtonLayout[KeyboardButton]
		Want   [][]KeyboardButton
	}{
		{
			Layout: NewButtonLayout[KeyboardButton](3).
				Add(NewKeyboardButton("text")),
			Want: [][]KeyboardButton{
				{{Text: "text"}},
			},
		},
		{
			Layout: NewButtonLayout[KeyboardButton](3).
				Add(NewKeyboardButton("text"), NewKeyboardButton("text"), NewKeyboardButton("text")),
			Want: [][]KeyboardButton{
				{{Text: "text"}, {Text: "text"}, {Text: "text"}},
			},
		},
		{
			Layout: NewButtonLayout[KeyboardButton](3).
				Add(NewKeyboardButton("text"), NewKeyboardButton("text"), NewKeyboardButton("text"), NewKeyboardButton("text")),
			Want: [][]KeyboardButton{
				{{Text: "text"}, {Text: "text"}, {Text: "text"}},
				{{Text: "text"}},
			},
		},
		{
			Layout: NewButtonLayout[KeyboardButton](3).
				Add(NewKeyboardButton("text"), NewKeyboardButton("text"), NewKeyboardButton("text"), NewKeyboardButton("text")).
				Add(NewKeyboardButton("text")),
			Want: [][]KeyboardButton{
				{{Text: "text"}, {Text: "text"}, {Text: "text"}},
				{{Text: "text"}},
				{{Text: "text"}},
			},
		},
	} {
		assert.Equal(t, test.Want, test.Layout.Keyboard())
	}
}

func TestButtonLayout_Row(t *testing.T) {
	keyboard := NewButtonLayout(1,
		NewKeyboardButton("1"),
		NewKeyboardButton("2"),
		NewKeyboardButton("3"),
	).Row(
		NewKeyboardButton("4"),
		NewKeyboardButton("5"),
		NewKeyboardButton("6"),
		NewKeyboardButton("7"),
	).Keyboard()

	assert.Equal(t, [][]KeyboardButton{
		{{Text: "1"}},
		{{Text: "2"}},
		{{Text: "3"}},
		{{Text: "4"}, {Text: "5"}, {Text: "6"}, {Text: "7"}},
	}, keyboard)
}

func TestButtonLayout_Insert(t *testing.T) {
	for _, test := range []struct {
		Layout *ButtonLayout[KeyboardButton]
		Want   [][]KeyboardButton
	}{
		{
			Layout: NewButtonLayout[KeyboardButton](3).
				Insert(NewKeyboardButton("text")),
			Want: [][]KeyboardButton{
				{{Text: "text"}},
			},
		},
		{
			Layout: NewButtonLayout[KeyboardButton](3).
				Insert(NewKeyboardButton("text"), NewKeyboardButton("text"), NewKeyboardButton("text")),
			Want: [][]KeyboardButton{
				{{Text: "text"}, {Text: "text"}, {Text: "text"}},
			},
		},
		{
			Layout: NewButtonLayout[KeyboardButton](3).
				Insert(NewKeyboardButton("text"), NewKeyboardButton("text"), NewKeyboardButton("text"), NewKeyboardButton("text")),
			Want: [][]KeyboardButton{
				{{Text: "text"}, {Text: "text"}, {Text: "text"}},
				{{Text: "text"}},
			},
		},
		{
			Layout: NewButtonLayout[KeyboardButton](3).
				Insert(NewKeyboardButton("1"), NewKeyboardButton("2")).
				Insert(NewKeyboardButton("3")),
			Want: [][]KeyboardButton{
				{{Text: "1"}, {Text: "2"}, {Text: "3"}},
			},
		},
		{
			Layout: NewButtonLayout[KeyboardButton](3).
				Insert(NewKeyboardButton("1"), NewKeyboardButton("2")).
				Insert(NewKeyboardButton("3")),
			Want: [][]KeyboardButton{
				{{Text: "1"}, {Text: "2"}, {Text: "3"}},
			},
		},
		{
			Layout: NewButtonLayout[KeyboardButton](3).
				Add(NewKeyboardButton("1"), NewKeyboardButton("2"), NewKeyboardButton("3")).
				Insert(NewKeyboardButton("4")).
				Add(NewKeyboardButton("5")),
			Want: [][]KeyboardButton{
				{{Text: "1"}, {Text: "2"}, {Text: "3"}},
				{{Text: "4"}},
				{{Text: "5"}},
			},
		},
	} {
		assert.Equal(t, test.Want, test.Layout.Keyboard())
	}
}

func TestNewButtonColumn(t *testing.T) {
	keyboard := NewButtonColumn(
		NewInlineKeyboardButtonCallbackData("1", "1"),
		NewInlineKeyboardButtonCallbackData("2", "2"),
		NewInlineKeyboardButtonCallbackData("3", "3"),
	)

	assert.Equal(t, [][]InlineKeyboardButton{
		{{Text: "1", CallbackData: "1"}},
		{{Text: "2", CallbackData: "2"}},
		{{Text: "3", CallbackData: "3"}},
	}, keyboard)
}

func TestInlineQueryResultMarshalJSON(t *testing.T) {
	for _, test := range []struct {
		Type   string
		Result InlineQueryResult
	}{
		{"audio", NewInlineQueryResultCachedAudio(InlineQueryResultCachedAudio{})},
		{"document", NewInlineQueryResultCachedDocument(InlineQueryResultCachedDocument{})},
		{"gif", NewInlineQueryResultCachedGIF(InlineQueryResultCachedGIF{})},
		{"mpeg4_gif", NewInlineQueryResultCachedMPEG4GIF(InlineQueryResultCachedMPEG4GIF{})},
		{"photo", NewInlineQueryResultCachedPhoto(InlineQueryResultCachedPhoto{})},
		{"sticker", NewInlineQueryResultCachedSticker(InlineQueryResultCachedSticker{})},
		{"video", NewInlineQueryResultCachedVideo(InlineQueryResultCachedVideo{})},
		{"voice", NewInlineQueryResultCachedVoice(InlineQueryResultCachedVoice{})},
		{"audio", NewInlineQueryResultAudio(InlineQueryResultAudio{})},
		{"document", NewInlineQueryResultDocument(InlineQueryResultDocument{})},
		{"gif", NewInlineQueryResultGIF(InlineQueryResultGIF{})},
		{"mpeg4_gif", NewInlineQueryResultMPEG4GIF(InlineQueryResultMPEG4GIF{})},
		{"photo", NewInlineQueryResultPhoto(InlineQueryResultPhoto{})},
		{"video", NewInlineQueryResultVideo(InlineQueryResultVideo{})},
		{"voice", NewInlineQueryResultVoice(InlineQueryResultVoice{})},
		{"article", NewInlineQueryResultArticle(InlineQueryResultArticle{})},
		{"contact", NewInlineQueryResultContact(InlineQueryResultContact{})},
		{"game", NewInlineQueryResultGame(InlineQueryResultGame{})},
		{"location", NewInlineQueryResultLocation(InlineQueryResultLocation{})},
		{"venue", NewInlineQueryResultVenue(InlineQueryResultVenue{})},
	} {
		t.Run(test.Type, func(t *testing.T) {
			body, err := json.Marshal(test.Result)
			require.NoError(t, err, "marshal json")

			result := struct {
				Type string `json:"type"`
			}{}

			err = json.Unmarshal(body, &result)
			require.NoError(t, err, "unmarshal json")

			assert.Equal(t, test.Type, result.Type)
		})
	}
}

func TestInputMessageContent(t *testing.T) {
	for _, test := range []InputMessageContent{
		InputTextMessageContent{},
		InputLocationMessageContent{},
		InputVenueMessageContent{},
		InputContactMessageContent{},
		InputInvoiceMessageContent{},
	} {
		assert.Implements(t, (*InputMessageContent)(nil), test)
		test.isInputMessageContent()
	}
}

func TestInputMedia_getMedia(t *testing.T) {
	for _, test := range []InputMedia{
		NewInputMediaPhoto(InputMediaPhoto{}),
		NewInputMediaVideo(InputMediaVideo{}),
		NewInputMediaAudio(InputMediaAudio{}),
		NewInputMediaAnimation(InputMediaAnimation{}),
		NewInputMediaDocument(InputMediaDocument{}),
	} {
		media, _ := test.getMedia()
		assert.NotNil(t, media)
	}
}

func TestInputMedia_MarshalJSON(t *testing.T) {
	for _, test := range []struct {
		InputMedia InputMedia
		Want       string
	}{
		{
			InputMedia: NewInputMediaPhoto(InputMediaPhoto{
				Media: FileArg{FileID: "file_id"},
			}),
			Want: `{"type":"photo","media":"file_id"}`,
		},
		{
			InputMedia: NewInputMediaVideo(InputMediaVideo{
				Media: FileArg{FileID: "file_id"},
			}),
			Want: `{"type":"video","media":"file_id"}`,
		},
		{
			InputMedia: NewInputMediaAudio(InputMediaAudio{
				Media: FileArg{FileID: "file_id"},
			}),
			Want: `{"type":"audio","media":"file_id"}`,
		},
		{
			InputMedia: NewInputMediaAnimation(InputMediaAnimation{
				Media: FileArg{FileID: "file_id"},
			}),
			Want: `{"type":"animation","media":"file_id"}`,
		},
		{
			InputMedia: NewInputMediaDocument(InputMediaDocument{
				Media: FileArg{FileID: "file_id"},
			}),
			Want: `{"type":"document","media":"file_id"}`,
		},
	} {
		v, err := json.Marshal(test.InputMedia)
		require.NoError(t, err, "marshal json")
		assert.Equal(t, test.Want, string(v))
	}
}

func TestFileArg_MarshalJSON(t *testing.T) {
	for _, test := range []struct {
		Name    string
		FileArg FileArg
		Want    string
		Err     bool
	}{
		{
			Name:    "FileID",
			FileArg: FileArg{FileID: "file_id"},
			Want:    `"file_id"`,
		},
		{
			Name:    "FileURL",
			FileArg: FileArg{URL: "file_url"},
			Want:    `"file_url"`,
		},
		{
			Name:    "FileAddr",
			FileArg: FileArg{addr: "addr"},
			Want:    `"addr"`,
		},
		{
			Name: "FileUpload",
			FileArg: FileArg{
				Upload: InputFile{},
			},
			Err: true,
		},
	} {
		t.Run(test.Name, func(t *testing.T) {
			v, err := json.Marshal(test.FileArg)

			if test.Err {
				require.Error(t, err, "marshal json")
			} else {
				require.NoError(t, err, "marshal json")
				assert.Equal(t, test.Want, string(v))
			}
		})
	}
}

func TestFileArg_getString(t *testing.T) {
	for _, test := range []struct {
		FileArg FileArg
		Want    string
	}{
		{
			FileArg: FileArg{FileID: "file_id"},
			Want:    "file_id",
		},
		{
			FileArg: FileArg{URL: "file_url"},
			Want:    "file_url",
		},
		{
			FileArg: FileArg{addr: "addr"},
			Want:    "addr",
		},
	} {
		assert.Equal(t, test.Want, test.FileArg.getRef())
	}
}

func TestNewFileArgUpload(t *testing.T) {
	f := NewFileArgUpload(InputFile{Name: "file_name"})
	assert.Equal(t, "file_name", f.Upload.Name)
}

func TestNewFileArgURL(t *testing.T) {
	f := NewFileArgURL("https://picsum.photos/500")
	assert.Equal(t, "https://picsum.photos/500", f.URL)
}

func TestNewFileArgID(t *testing.T) {
	f := NewFileArgID("file_id")
	assert.Equal(t, FileID("file_id"), f.FileID)
}

func TestBotCommandScope(t *testing.T) {
	for _, test := range []struct {
		Scope BotCommandScope
		Want  string
	}{
		{NewBotCommandScopeDefault(BotCommandScopeDefault{}), `{"type":"default"}`},
		{NewBotCommandScopeAllPrivateChats(BotCommandScopeAllPrivateChats{}), `{"type":"all_private_chats"}`},
		{NewBotCommandScopeAllGroupChats(BotCommandScopeAllGroupChats{}), `{"type":"all_group_chats"}`},
		{NewBotCommandScopeAllChatAdministrators(BotCommandScopeAllChatAdministrators{}), `{"type":"all_chat_administrators"}`},
		{NewBotCommandScopeChat(BotCommandScopeChat{}), `{"type":"chat","chat_id":0}`},
		{NewBotCommandScopeChatAdministrators(BotCommandScopeChatAdministrators{}), `{"type":"chat_administrators","chat_id":0}`},
		{NewBotCommandScopeChatMember(BotCommandScopeChatMember{}), `{"type":"chat_member","chat_id":0,"user_id":0}`},
	} {
		v, err := json.Marshal(test.Scope)
		require.NoError(t, err, "marshal json")
		assert.Equal(t, test.Want, string(v))
	}
}

func TestMenuButton(t *testing.T) {
	for _, test := range []struct {
		Button MenuButton
		Want   string
	}{
		{NewMenuButtonDefault(MenuButtonDefault{}), `{"type":"default"}`},
		{NewMenuButtonCommands(MenuButtonCommands{}), `{"type":"commands"}`},
		{NewMenuButtonWebApp(MenuButtonWebApp{}), `{"type":"web_app","text":"","web_app":{"url":""}}`},
	} {
		v, err := json.Marshal(test.Button)
		require.NoError(t, err, "marshal json")
		assert.Equal(t, test.Want, string(v))
	}
}

func TestMessage_Type(t *testing.T) {
	for _, test := range []struct {
		Message *Message
		Want    MessageType
	}{
		{
			Message: &Message{},
			Want:    MessageTypeUnknown,
		},
		{
			Message: &Message{Text: "hello"},
			Want:    MessageTypeText,
		},
		{
			Message: &Message{Animation: &Animation{}},
			Want:    MessageTypeAnimation,
		},
		{
			Message: &Message{Audio: &Audio{}},
			Want:    MessageTypeAudio,
		},
		{
			Message: &Message{Document: &Document{}},
			Want:    MessageTypeDocument,
		},
		{
			Message: &Message{Photo: []PhotoSize{{}}},
			Want:    MessageTypePhoto,
		},
		{
			Message: &Message{Sticker: &Sticker{}},
			Want:    MessageTypeSticker,
		},
		{
			Message: &Message{Video: &Video{}},
			Want:    MessageTypeVideo,
		},
		{
			Message: &Message{VideoNote: &VideoNote{}},
			Want:    MessageTypeVideoNote,
		},
		{
			Message: &Message{Voice: &Voice{}},
			Want:    MessageTypeVoice,
		},
		{
			Message: &Message{Contact: &Contact{}},
			Want:    MessageTypeContact,
		},
		{
			Message: &Message{Dice: &Dice{}},
			Want:    MessageTypeDice,
		},
		{
			Message: &Message{Game: &Game{}},
			Want:    MessageTypeGame,
		},
		{
			Message: &Message{Poll: &Poll{}},
			Want:    MessageTypePoll,
		},
		{
			Message: &Message{Venue: &Venue{}},
			Want:    MessageTypeVenue,
		},
		{
			Message: &Message{Location: &Location{}},
			Want:    MessageTypeLocation,
		},
		{
			Message: &Message{NewChatMembers: []User{{}}},
			Want:    MessageTypeNewChatMembers,
		},
		{
			Message: &Message{LeftChatMember: &User{}},
			Want:    MessageTypeLeftChatMember,
		},
		{
			Message: &Message{NewChatTitle: "hello"},
			Want:    MessageTypeNewChatTitle,
		},
		{
			Message: &Message{NewChatPhoto: []PhotoSize{{}}},
			Want:    MessageTypeNewChatPhoto,
		},
		{
			Message: &Message{DeleteChatPhoto: true},
			Want:    MessageTypeDeleteChatPhoto,
		},
		{
			Message: &Message{GroupChatCreated: true},
			Want:    MessageTypeGroupChatCreated,
		},
		{
			Message: &Message{SupergroupChatCreated: true},
			Want:    MessageTypeSupergroupChatCreated,
		},
		{
			Message: &Message{ChannelChatCreated: true},
			Want:    MessageTypeChannelChatCreated,
		},
		{
			Message: &Message{MessageAutoDeleteTimerChanged: &MessageAutoDeleteTimerChanged{}},
			Want:    MessageTypeMessageAutoDeleteTimerChanged,
		},
		{
			Message: &Message{MigrateToChatID: -10023123123},
			Want:    MessageTypeMigrateToChatID,
		},
		{
			Message: &Message{MigrateFromChatID: -10023123123},
			Want:    MessageTypeMigrateFromChatID,
		},
		{
			Message: &Message{PinnedMessage: &MaybeInaccessibleMessage{
				Message: &Message{},
			}},
			Want: MessageTypePinnedMessage,
		},
		{
			Message: &Message{Invoice: &Invoice{}},
			Want:    MessageTypeInvoice,
		},
		{
			Message: &Message{SuccessfulPayment: &SuccessfulPayment{}},
			Want:    MessageTypeSuccessfulPayment,
		},
		{
			Message: &Message{UsersShared: &UsersShared{}},
			Want:    MessageTypeUsersShared,
		},
		{
			Message: &Message{ChatShared: &ChatShared{}},
			Want:    MessageTypeChatShared,
		},
		{
			Message: &Message{ConnectedWebsite: "telegram.me"},
			Want:    MessageTypeConnectedWebsite,
		},
		{
			Message: &Message{PassportData: &PassportData{}},
			Want:    MessageTypePassportData,
		},
		{
			Message: &Message{ProximityAlertTriggered: &ProximityAlertTriggered{}},
			Want:    MessageTypeProximityAlertTriggered,
		},
		{
			Message: &Message{VideoChatScheduled: &VideoChatScheduled{}},
			Want:    MessageTypeVideoChatScheduled,
		},
		{
			Message: &Message{VideoChatStarted: &VideoChatStarted{}},
			Want:    MessageTypeVideoChatStarted,
		},
		{
			Message: &Message{VideoChatEnded: &VideoChatEnded{}},
			Want:    MessageTypeVideoChatEnded,
		},
		{
			Message: &Message{VideoChatParticipantsInvited: &VideoChatParticipantsInvited{}},
			Want:    MessageTypeVideoChatParticipantsInvited,
		},
		{
			Message: &Message{WebAppData: &WebAppData{}},
			Want:    MessageTypeWebAppData,
		},
	} {
		assert.Equal(t, test.Want, test.Message.Type())
	}
}

func TestUpdateType_String(t *testing.T) {
	for _, test := range []struct {
		Type UpdateType
		Want string
	}{
		{UpdateTypeUnknown, "unknown"},
		{UpdateTypeMessage, "message"},
		{UpdateTypeEditedMessage, "edited_message"},
		{UpdateTypeChannelPost, "channel_post"},
		{UpdateTypeEditedChannelPost, "edited_channel_post"},
		{UpdateTypeInlineQuery, "inline_query"},
		{UpdateTypeChosenInlineResult, "chosen_inline_result"},
		{UpdateTypeCallbackQuery, "callback_query"},
		{UpdateTypeShippingQuery, "shipping_query"},
		{UpdateTypePreCheckoutQuery, "pre_checkout_query"},
		{UpdateTypePoll, "poll"},
		{UpdateTypePollAnswer, "poll_answer"},
		{UpdateTypeMyChatMember, "my_chat_member"},
		{UpdateTypeChatMember, "chat_member"},
		{UpdateTypeChatJoinRequest, "chat_join_request"},
		{UpdateTypeMessageReaction, "message_reaction"},
		{UpdateTypeMessageReactionCount, "message_reaction_count"},
		{UpdateTypeChatBoost, "chat_boost"},
		{UpdateTypeRemovedChatBoost, "removed_chat_boost"},
	} {
		assert.Equal(t, test.Want, test.Type.String(), "update type: %s", test.Want)
	}
}

func TestMessage_IsInaccessible(t *testing.T) {
	accessible := &Message{
		Date: UnixTime(time.Now().Unix()),
	}

	inaccessible := &Message{
		Date: 0,
	}

	assert.False(t, accessible.IsInaccessible())
	assert.True(t, inaccessible.IsInaccessible())
}

func TestUpdateType_UnmarshalText(t *testing.T) {
	for _, test := range []struct {
		Text string
		Want UpdateType
		Err  bool
	}{
		{"message", UpdateTypeMessage, false},
		{"edited_message", UpdateTypeEditedMessage, false},
		{"channel_post", UpdateTypeChannelPost, false},
		{"edited_channel_post", UpdateTypeEditedChannelPost, false},
		{"inline_query", UpdateTypeInlineQuery, false},
		{"chosen_inline_result", UpdateTypeChosenInlineResult, false},
		{"callback_query", UpdateTypeCallbackQuery, false},
		{"shipping_query", UpdateTypeShippingQuery, false},
		{"pre_checkout_query", UpdateTypePreCheckoutQuery, false},
		{"poll", UpdateTypePoll, false},
		{"poll_answer", UpdateTypePollAnswer, false},
		{"my_chat_member", UpdateTypeMyChatMember, false},
		{"chat_member", UpdateTypeChatMember, false},
		{"chat_join_request", UpdateTypeChatJoinRequest, false},
		{"message_reaction", UpdateTypeMessageReaction, false},
		{"message_reaction_count", UpdateTypeMessageReactionCount, false},
		{"chat_boost", UpdateTypeChatBoost, false},
		{"removed_chat_boost", UpdateTypeRemovedChatBoost, false},
		{"business_connection", UpdateTypeBusinessConnection, false},
		{"business_message", UpdateTypeBusinessMessage, false},
		{"edited_business_message", UpdateTypeEditedBusinessMessage, false},
		{"deleted_business_messages", UpdateTypeDeletedBusinessMessages, false},
		{"test", UpdateTypeUnknown, false}, // unknown values set to Unknown
	} {
		t.Run(test.Text, func(t *testing.T) {
			var typ UpdateType

			err := typ.UnmarshalText([]byte(test.Text))

			if test.Err {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, test.Want, typ)
			}
		})
	}
}

func TestUpdateType_MarshalText(t *testing.T) {
	v := UpdateTypeEditedMessage

	b, err := v.MarshalText()
	require.NoError(t, err)
	assert.Equal(t, []byte("edited_message"), b)

	v = UpdateTypeUnknown
	_, err = v.MarshalText()
	require.Error(t, err)

	output, err := json.Marshal(struct {
		Type []UpdateType `json:"type"`
	}{
		Type: []UpdateType{UpdateTypeCallbackQuery, UpdateTypeChannelPost},
	})

	require.NoError(t, err)
	assert.JSONEq(t, `{"type":["callback_query","channel_post"]}`, string(output))
}

func TestUpdate_Type(t *testing.T) {
	for _, test := range []struct {
		Update *Update
		Want   UpdateType
	}{
		{
			Update: &Update{},
			Want:   UpdateTypeUnknown,
		},
		{
			Update: &Update{Message: &Message{}},
			Want:   UpdateTypeMessage,
		},
		{
			Update: &Update{EditedMessage: &Message{}},
			Want:   UpdateTypeEditedMessage,
		},
		{
			Update: &Update{ChannelPost: &Message{}},
			Want:   UpdateTypeChannelPost,
		},
		{
			Update: &Update{EditedChannelPost: &Message{}},
			Want:   UpdateTypeEditedChannelPost,
		},
		{
			Update: &Update{InlineQuery: &InlineQuery{}},
			Want:   UpdateTypeInlineQuery,
		},
		{
			Update: &Update{ChosenInlineResult: &ChosenInlineResult{}},
			Want:   UpdateTypeChosenInlineResult,
		},
		{
			Update: &Update{CallbackQuery: &CallbackQuery{}},
			Want:   UpdateTypeCallbackQuery,
		},
		{
			Update: &Update{ShippingQuery: &ShippingQuery{}},
			Want:   UpdateTypeShippingQuery,
		},
		{
			Update: &Update{PreCheckoutQuery: &PreCheckoutQuery{}},
			Want:   UpdateTypePreCheckoutQuery,
		},
		{
			Update: &Update{Poll: &Poll{}},
			Want:   UpdateTypePoll,
		},
		{
			Update: &Update{PollAnswer: &PollAnswer{}},
			Want:   UpdateTypePollAnswer,
		},
		{
			Update: &Update{MyChatMember: &ChatMemberUpdated{}},
			Want:   UpdateTypeMyChatMember,
		},
		{
			Update: &Update{ChatJoinRequest: &ChatJoinRequest{}},
			Want:   UpdateTypeChatJoinRequest,
		},
		{
			Update: &Update{ChatMember: &ChatMemberUpdated{}},
			Want:   UpdateTypeChatMember,
		},
		{
			Update: &Update{MessageReaction: &MessageReactionUpdated{}},
			Want:   UpdateTypeMessageReaction,
		},
		{
			Update: &Update{MessageReactionCount: &MessageReactionCountUpdated{}},
			Want:   UpdateTypeMessageReactionCount,
		},
		{
			Update: &Update{ChatBoost: &ChatBoostUpdated{}},
			Want:   UpdateTypeChatBoost,
		},
		{
			Update: &Update{RemovedChatBoost: &ChatBoostRemoved{}},
			Want:   UpdateTypeRemovedChatBoost,
		},
		{
			Update: &Update{BusinessConnection: &BusinessConnection{}},
			Want:   UpdateTypeBusinessConnection,
		},
		{
			Update: &Update{BusinessMessage: &Message{}},
			Want:   UpdateTypeBusinessMessage,
		},
		{
			Update: &Update{EditedBusinessMessage: &Message{}},
			Want:   UpdateTypeEditedBusinessMessage,
		},
		{
			Update: &Update{DeletedBusinessMessages: &BusinessMessagesDeleted{}},
			Want:   UpdateTypeDeletedBusinessMessages,
		},
	} {
		assert.Equal(t, test.Want, test.Update.Type())
	}
}

func TestMessageEntityType_String(t *testing.T) {
	for _, test := range []struct {
		Type MessageEntityType
		Want string
	}{
		{MessageEntityTypeUnknown, "unknown"},
		{MessageEntityTypeMention, "mention"},
		{MessageEntityTypeHashtag, "hashtag"},
		{MessageEntityTypeCashtag, "cashtag"},
		{MessageEntityTypeBotCommand, "bot_command"},
		{MessageEntityTypeURL, "url"},
		{MessageEntityTypeEmail, "email"},
		{MessageEntityTypePhoneNumber, "phone_number"},
		{MessageEntityTypeBold, "bold"},
		{MessageEntityTypeItalic, "italic"},
		{MessageEntityTypeUnderline, "underline"},
		{MessageEntityTypeStrikethrough, "strikethrough"},
		{MessageEntityTypeSpoiler, "spoiler"},
		{MessageEntityTypeCode, "code"},
		{MessageEntityTypePre, "pre"},
		{MessageEntityTypeTextLink, "text_link"},
		{MessageEntityTypeTextMention, "text_mention"},
		{MessageEntityTypeCustomEmoji, "custom_emoji"},
		{MessageEntityTypeBlockquote, "blockquote"},
	} {
		assert.Equal(t, test.Want, test.Type.String())
	}
}

func TestMessageEntityType_MarshalText(t *testing.T) {
	for _, test := range []struct {
		Type encoding.TextMarshaler
		Want []byte
		Err  bool
	}{
		{MessageEntityTypeUnknown, nil, true},
		{MessageEntityTypeMention, []byte("mention"), false},
		{MessageEntityTypeHashtag, []byte("hashtag"), false},
		{MessageEntityTypeCashtag, []byte("cashtag"), false},
		{MessageEntityTypeBotCommand, []byte("bot_command"), false},
		{MessageEntityTypeURL, []byte("url"), false},
		{MessageEntityTypeEmail, []byte("email"), false},
		{MessageEntityTypePhoneNumber, []byte("phone_number"), false},
		{MessageEntityTypeBold, []byte("bold"), false},
		{MessageEntityTypeItalic, []byte("italic"), false},
		{MessageEntityTypeUnderline, []byte("underline"), false},
		{MessageEntityTypeStrikethrough, []byte("strikethrough"), false},
		{MessageEntityTypeSpoiler, []byte("spoiler"), false},
		{MessageEntityTypeCode, []byte("code"), false},
		{MessageEntityTypePre, []byte("pre"), false},
		{MessageEntityTypeTextLink, []byte("text_link"), false},
		{MessageEntityTypeTextMention, []byte("text_mention"), false},
		{MessageEntityTypeCustomEmoji, []byte("custom_emoji"), false},
		{MessageEntityTypeBlockquote, []byte("blockquote"), false},
	} {
		b, err := test.Type.MarshalText()
		if test.Err {
			assert.Error(t, err)
		} else {
			require.NoError(t, err)
			assert.Equal(t, test.Want, b)
		}
	}
}

func TestMessageEntityType_UnmarshalText(t *testing.T) {
	for _, test := range []struct {
		Input string
		Want  MessageEntityType
	}{
		{"unknown_value", MessageEntityTypeUnknown},
		{"mention", MessageEntityTypeMention},
		{"hashtag", MessageEntityTypeHashtag},
		{"cashtag", MessageEntityTypeCashtag},
		{"bot_command", MessageEntityTypeBotCommand},
		{"url", MessageEntityTypeURL},
		{"email", MessageEntityTypeEmail},
		{"phone_number", MessageEntityTypePhoneNumber},
		{"bold", MessageEntityTypeBold},
		{"italic", MessageEntityTypeItalic},
		{"underline", MessageEntityTypeUnderline},
		{"strikethrough", MessageEntityTypeStrikethrough},
		{"spoiler", MessageEntityTypeSpoiler},
		{"code", MessageEntityTypeCode},
		{"pre", MessageEntityTypePre},
		{"text_link", MessageEntityTypeTextLink},
		{"text_mention", MessageEntityTypeTextMention},
		{"custom_emoji", MessageEntityTypeCustomEmoji},
		{"blockquote", MessageEntityTypeBlockquote},
		{"expandable_blockquote", MessageEntityTypeExpandableBlockquote},
	} {
		var e MessageEntityType

		err := e.UnmarshalText([]byte(test.Input))
		require.NoError(t, err)
		assert.Equal(t, test.Want, e)
	}
}

func TestMessageEntity_Extract(t *testing.T) {
	text := "Lorem Ipsum - це текст-\"риба\", що використовується в друкарстві та дизайні. Lorem Ipsum є, фактично, стандартною \"рибою\" аж з XVI сторіччя, коли невідомий друкар взяв шрифтову гранку та склав на ній підбірку зразків шрифтів. Пишить мені на hey@lipsum.com"

	boldEntity := MessageEntity{
		Type:   MessageEntityTypeBold,
		Offset: 0,
		Length: 11,
	}

	assert.Equal(t, "Lorem Ipsum", boldEntity.Extract(text))

	emailEntity := MessageEntity{
		Type:   MessageEntityTypeEmail,
		Offset: 240,
		Length: 14,
	}

	assert.Equal(t, "hey@lipsum.com", emailEntity.Extract(text))
}

func TestUpdate_Msg(t *testing.T) {
	msg := &Message{ID: 1}

	for _, test := range []struct {
		Update  *Update
		Message *Message
	}{
		{nil, nil},
		{&Update{}, nil},
		{&Update{Message: msg}, msg},
		{&Update{EditedMessage: msg}, msg},
		{&Update{ChannelPost: msg}, msg},
		{&Update{EditedChannelPost: msg}, msg},
		{&Update{CallbackQuery: &CallbackQuery{Message: &MaybeInaccessibleMessage{
			Message: msg,
		}}}, msg},
		{&Update{CallbackQuery: &CallbackQuery{}}, nil},
		{&Update{BusinessMessage: msg}, msg},
		{&Update{EditedBusinessMessage: msg}, msg},
	} {
		assert.Equal(t, test.Message, test.Update.Msg())
	}
}

func TestUpdate_Chat(t *testing.T) {
	chat := Chat{ID: 1}

	for _, test := range []struct {
		Update *Update
		Chat   *Chat
	}{
		{nil, nil},
		{&Update{InlineQuery: &InlineQuery{}}, nil},
		{&Update{Message: &Message{Chat: chat}}, &chat},
		{&Update{ChatMember: &ChatMemberUpdated{Chat: chat}}, &chat},
		{&Update{MyChatMember: &ChatMemberUpdated{Chat: chat}}, &chat},
		{&Update{ChatJoinRequest: &ChatJoinRequest{Chat: chat}}, &chat},
	} {
		assert.Equal(t, test.Chat, test.Update.Chat())
	}
}

func TestStickerType_MarshalText(t *testing.T) {
	for _, test := range []struct {
		Type StickerType
		Want string
		Err  bool
	}{
		{StickerTypeUnknown, "", true},
		{StickerTypeRegular, "regular", false},
		{StickerTypeMask, "mask", false},
		{StickerTypeCustomEmoji, "custom_emoji", false},
	} {
		b, err := test.Type.MarshalText()
		if test.Err {
			assert.Error(t, err)
		} else {
			require.NoError(t, err)
			assert.Equal(t, test.Want, string(b))
		}
	}
}

func TestStickerType_UnmarshalText(t *testing.T) {
	for _, test := range []struct {
		Input string
		Want  StickerType
	}{
		{"some_unknown_value", StickerTypeUnknown},
		{"regular", StickerTypeRegular},
		{"mask", StickerTypeMask},
		{"custom_emoji", StickerTypeCustomEmoji},
	} {
		var e StickerType

		err := e.UnmarshalText([]byte(test.Input))
		require.NoError(t, err)
		assert.Equal(t, test.Want, e)
	}
}

func TestMenuButtonOneOf_UnmarshalJSON(t *testing.T) {
	t.Run("Commands", func(t *testing.T) {
		var b MenuButtonOneOf

		err := b.UnmarshalJSON([]byte(`{"type": "commands"}`))
		require.NoError(t, err)

		require.NotNil(t, b.Commands)
		assert.Equal(t, "commands", b.Commands.Type)
	})

	t.Run("Default", func(t *testing.T) {
		var b MenuButtonOneOf

		err := b.UnmarshalJSON([]byte(`{"type": "default"}`))
		require.NoError(t, err)

		require.NotNil(t, b.Default)
		assert.Equal(t, "default", b.Default.Type)
	})

	t.Run("WebApp", func(t *testing.T) {
		var b MenuButtonOneOf

		err := b.UnmarshalJSON([]byte(`{"type": "web_app", "text": "12345"}`))
		require.NoError(t, err)

		require.NotNil(t, b.WebApp)
		assert.Equal(t, "web_app", b.WebApp.Type)
		assert.Equal(t, "12345", b.WebApp.Text)
	})
}

func TestMessageOrigin_Type(t *testing.T) {
	for _, test := range []struct {
		Origin *MessageOrigin
		Want   MessageOriginType
	}{
		{
			Origin: &MessageOrigin{},
			Want:   0,
		},
		{
			Origin: &MessageOrigin{User: &MessageOriginUser{}},
			Want:   MessageOriginTypeUser,
		},
		{
			Origin: &MessageOrigin{HiddenUser: &MessageOriginHiddenUser{}},
			Want:   MessageOriginTypeHiddenUser,
		},
		{
			Origin: &MessageOrigin{Chat: &MessageOriginChat{}},
			Want:   MessageOriginTypeChat,
		},
	} {
		assert.Equal(t, test.Want, test.Origin.Type())
	}
}

func TestMessageOrigin_UnmarshalJSON(t *testing.T) {
	t.Run("MessageOriginUser", func(t *testing.T) {
		var b MessageOrigin

		err := b.UnmarshalJSON([]byte(`{"type": "user", "date": 12345, "sender_user": {"id":1}}`))
		require.NoError(t, err)

		assert.Equal(t, MessageOriginTypeUser, b.Type())
		require.NotNil(t, b.User)
		assert.EqualValues(t, 12345, b.User.Date)
		assert.Equal(t, UserID(1), b.User.SenderUser.ID)
	})

	t.Run("MessageOriginHiddenUser", func(t *testing.T) {
		var b MessageOrigin

		err := b.UnmarshalJSON([]byte(`{"type": "hidden_user", "date": 12345, "sender_user_name": "john doe"}`))
		require.NoError(t, err)

		assert.Equal(t, MessageOriginTypeHiddenUser, b.Type())
		require.NotNil(t, b.HiddenUser)
		assert.EqualValues(t, 12345, b.HiddenUser.Date)
		assert.Equal(t, "john doe", b.HiddenUser.SenderUserName)
	})

	t.Run("MessageOriginChat", func(t *testing.T) {
		var b MessageOrigin

		err := b.UnmarshalJSON([]byte(`{"type": "chat", "date": 12345, "sender_chat": {"id":1}, "author_signature": "john doe"}`))
		require.NoError(t, err)

		assert.Equal(t, MessageOriginTypeChat, b.Type())
		require.NotNil(t, b.Chat)
		assert.EqualValues(t, 12345, b.Chat.Date)
		assert.Equal(t, ChatID(1), b.Chat.SenderChat.ID)
		assert.Equal(t, "john doe", b.Chat.AuthorSignature)
	})

	t.Run("MessageOriginChannel", func(t *testing.T) {
		var b MessageOrigin

		err := b.UnmarshalJSON([]byte(`{"type": "channel", "date": 12345, "chat": {"id":1}, "message_id": 2, "author_signature": "john doe"}`))
		require.NoError(t, err)

		assert.Equal(t, MessageOriginTypeChannel, b.Type())
		require.NotNil(t, b.Channel)
		assert.EqualValues(t, 12345, b.Channel.Date)
		assert.Equal(t, ChatID(1), b.Channel.Chat.ID)
		assert.Equal(t, 2, b.Channel.MessageID)
		assert.Equal(t, "john doe", b.Channel.AuthorSignature)
	})

	t.Run("MalformedJSON", func(t *testing.T) {
		var b MessageOrigin

		err := b.UnmarshalJSON([]byte(`{"type": "unknown"`))
		require.Error(t, err)
	})

	t.Run("Unknown", func(t *testing.T) {
		var b MessageOrigin

		err := b.UnmarshalJSON([]byte(`{"type": "future_origin", "date": 12345}`))
		require.NoError(t, err)

		assert.True(t, b.IsUnknown())
		require.NotNil(t, b.Unknown)
		assert.Equal(t, "future_origin", b.Unknown.Type)
		assert.Equal(t, MessageOriginType(0), b.Type())
	})
}

func TestMaybeInaccessibleMessage(t *testing.T) {
	t.Run("InaccessibleMessage", func(t *testing.T) {
		var m MaybeInaccessibleMessage

		err := m.UnmarshalJSON([]byte(`{"chat": {"id": 1}, "message_id": 2, "date": 0}`))
		require.NoError(t, err)

		assert.True(t, m.IsInaccessible())
		assert.Equal(t, ChatID(1), m.Chat().ID)
		assert.Equal(t, 2, m.MessageID())
		require.NotNil(t, m.InaccessibleMessage)
		assert.Equal(t, ChatID(1), m.InaccessibleMessage.Chat.ID)
		assert.Equal(t, 2, m.InaccessibleMessage.MessageID)
		assert.EqualValues(t, 0, m.InaccessibleMessage.Date)
	})

	t.Run("Message", func(t *testing.T) {
		var m MaybeInaccessibleMessage

		err := m.UnmarshalJSON([]byte(`{"message_id": 2, "date": 1234, "chat": {"id": 1}}`))
		require.NoError(t, err)

		assert.Equal(t, ChatID(1), m.Chat().ID)

		assert.True(t, m.IsAccessible())
		assert.Equal(t, 2, m.MessageID())
		require.NotNil(t, m.Message)
		assert.Equal(t, 2, m.Message.ID)
		assert.EqualValues(t, 1234, m.Message.Date)
	})

	t.Run("UnmarshalError", func(t *testing.T) {
		var m MaybeInaccessibleMessage

		err := m.UnmarshalJSON([]byte(`{"chat": {"id": 1}`))
		require.Error(t, err)
	})
}
