package tg

import (
	"encoding"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
		NewButtonRow(
			NewInlineKeyboardButtonURL("text", "https://google.com"),
			NewInlineKeyboardButtonCallback("text", "data"),
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

	assert.EqualValues(t, InlineKeyboardMarkup{
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

	assert.EqualValues(t, &ReplyKeyboardMarkup{
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

func TestNewButtonLayout(t *testing.T) {
	keyboard := NewButtonLayout(1,
		NewInlineKeyboardButtonCallback("1", "1"),
		NewInlineKeyboardButtonCallback("2", "2"),
		NewInlineKeyboardButtonCallback("3", "3"),
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
		assert.EqualValues(t, test.Want, test.Layout.Keyboard())
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
		assert.EqualValues(t, test.Want, test.Layout.Keyboard())
	}
}

func TestNewButtonColumn(t *testing.T) {
	keyboard := NewButtonColumn(
		NewInlineKeyboardButtonCallback("1", "1"),
		NewInlineKeyboardButtonCallback("2", "2"),
		NewInlineKeyboardButtonCallback("3", "3"),
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
		{"audio", InlineQueryResultCachedAudio{}},

		{"document", InlineQueryResultCachedDocument{}},
		{"gif", InlineQueryResultCachedGIF{}},
		{"mpeg4_gif", InlineQueryResultCachedMPEG4GIF{}},
		{"photo", InlineQueryResultCachedPhoto{}},
		{"sticker", InlineQueryResultCachedSticker{}},
		{"video", InlineQueryResultCachedVideo{}},
		{"voice", InlineQueryResultCachedVoice{}},
		{"audio", InlineQueryResultAudio{}},
		{"document", InlineQueryResultDocument{}},
		{"gif", InlineQueryResultGIF{}},
		{"mpeg4_gif", InlineQueryResultMPEG4GIF{}},
		{"photo", InlineQueryResultPhoto{}},
		{"video", InlineQueryResultVideo{}},
		{"voice", InlineQueryResultVoice{}},
		{"article", InlineQueryResultArticle{}},
		{"contact", InlineQueryResultContact{}},
		{"game", InlineQueryResultGame{}},
		{"location", InlineQueryResultLocation{}},
		{"venue", InlineQueryResultVenue{}},
	} {
		t.Run(test.Type, func(t *testing.T) {
			body, err := json.Marshal(test.Result)
			assert.NoError(t, err, "marshal json")

			test.Result.isInlineQueryResult()

			result := struct {
				Type string `json:"type"`
			}{}

			err = json.Unmarshal(body, &result)
			assert.NoError(t, err, "unmarshal json")

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
		&InputMediaPhoto{},
		&InputMediaVideo{},
		&InputMediaAudio{},
		&InputMediaAnimation{},
		&InputMediaDocument{},
	} {
		assert.Implements(t, (*InputMedia)(nil), test)

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
			InputMedia: &InputMediaPhoto{
				Media: FileArg{FileID: "file_id"},
			},
			Want: `{"type":"photo","media":"file_id"}`,
		},
		{
			InputMedia: &InputMediaVideo{
				Media: FileArg{FileID: "file_id"},
			},
			Want: `{"type":"video","media":"file_id"}`,
		},
		{
			InputMedia: &InputMediaAudio{
				Media: FileArg{FileID: "file_id"},
			},
			Want: `{"type":"audio","media":"file_id"}`,
		},
		{
			InputMedia: &InputMediaAnimation{
				Media: FileArg{FileID: "file_id"},
			},
			Want: `{"type":"animation","media":"file_id"}`,
		},
		{
			InputMedia: &InputMediaDocument{
				Media: FileArg{FileID: "file_id"},
			},
			Want: `{"type":"document","media":"file_id"}`,
		},
	} {
		v, err := json.Marshal(test.InputMedia)
		assert.NoError(t, err, "marshal json")
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
				assert.Error(t, err, "marshal json")
			} else {
				assert.NoError(t, err, "marshal json")
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
		{BotCommandScopeDefault{}, `{"type":"default"}`},
		{BotCommandScopeAllPrivateChats{}, `{"type":"all_private_chats"}`},
		{BotCommandScopeAllGroupChats{}, `{"type":"all_group_chats"}`},
		{BotCommandScopeAllChatAdministrators{}, `{"type":"all_chat_administrators"}`},
		{BotCommandScopeChat{}, `{"type":"chat","chat_id":0}`},
		{BotCommandScopeChatAdministrators{}, `{"type":"chat_administrators","chat_id":0}`},
		{BotCommandScopeChatMember{}, `{"type":"chat_member","chat_id":0,"user_id":0}`},
	} {
		v, err := json.Marshal(test.Scope)
		assert.NoError(t, err, "marshal json")
		assert.Equal(t, test.Want, string(v))
		test.Scope.isBotCommandScope()
	}
}

func TestMenuButton(t *testing.T) {
	for _, test := range []struct {
		Scope MenuButton
		Want  string
	}{
		{MenuButtonDefault{}, `{"type":"default"}`},
		{MenuButtonCommands{}, `{"type":"commands"}`},
		{MenuButtonWebApp{}, `{"type":"web_app","text":"","web_app":{"url":""}}`},
	} {
		v, err := json.Marshal(test.Scope)
		assert.NoError(t, err, "marshal json")
		assert.Equal(t, test.Want, string(v))
		test.Scope.isMenuButton()
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
		Date: time.Now().Unix(),
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
		{"test", UpdateTypeUnknown, true},
	} {
		t.Run(test.Text, func(t *testing.T) {
			var typ UpdateType

			err := typ.UnmarshalText([]byte(test.Text))

			if test.Err {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.Want, typ)
			}
		})
	}
}

func TestUpdateType_MarshalText(t *testing.T) {
	v := UpdateTypeEditedMessage

	b, err := v.MarshalText()
	assert.NoError(t, err)
	assert.Equal(t, []byte("edited_message"), b)

	v = UpdateTypeUnknown
	_, err = v.MarshalText()
	assert.Error(t, err)

	output, err := json.Marshal(struct {
		Type []UpdateType `json:"type"`
	}{
		Type: []UpdateType{UpdateTypeCallbackQuery, UpdateTypeChannelPost},
	})

	assert.NoError(t, err)
	assert.Equal(t, `{"type":["callback_query","channel_post"]}`, string(output))

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
		{MessageEntityCustomEmoji, "custom_emoji"},
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
		{MessageEntityCustomEmoji, []byte("custom_emoji"), false},
	} {
		b, err := test.Type.MarshalText()
		if test.Err {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
			assert.Equal(t, test.Want, b)
		}
	}
}

func TestMessageEntityType_UnmarshalText(t *testing.T) {
	for _, test := range []struct {
		Input string
		Want  MessageEntityType
		Err   bool
	}{
		{"unknown", MessageEntityTypeUnknown, true},
		{"mention", MessageEntityTypeMention, false},
		{"hashtag", MessageEntityTypeHashtag, false},
		{"cashtag", MessageEntityTypeCashtag, false},
		{"bot_command", MessageEntityTypeBotCommand, false},
		{"url", MessageEntityTypeURL, false},
		{"email", MessageEntityTypeEmail, false},
		{"phone_number", MessageEntityTypePhoneNumber, false},
		{"bold", MessageEntityTypeBold, false},
		{"italic", MessageEntityTypeItalic, false},
		{"underline", MessageEntityTypeUnderline, false},
		{"strikethrough", MessageEntityTypeStrikethrough, false},
		{"spoiler", MessageEntityTypeSpoiler, false},
		{"code", MessageEntityTypeCode, false},
		{"pre", MessageEntityTypePre, false},
		{"text_link", MessageEntityTypeTextLink, false},
		{"text_mention", MessageEntityTypeTextMention, false},
		{"custom_emoji", MessageEntityCustomEmoji, false},
	} {
		var e MessageEntityType

		err := e.UnmarshalText([]byte(test.Input))
		if test.Err {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
			assert.Equal(t, test.Want, e)
		}
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
	}{
		{StickerTypeUnknown, "unknown"},
		{StickerTypeRegular, "regular"},
		{StickerTypeMask, "mask"},
		{StickerTypeCustomEmoji, "custom_emoji"},
	} {
		b, err := test.Type.MarshalText()
		assert.NoError(t, err)
		assert.Equal(t, test.Want, string(b))
	}
}

func TestStickerType_UnmarshalText(t *testing.T) {
	for _, test := range []struct {
		Input string
		Want  StickerType
	}{
		{"unknown", StickerTypeUnknown},
		{"regular", StickerTypeRegular},
		{"mask", StickerTypeMask},
		{"custom_emoji", StickerTypeCustomEmoji},
	} {
		var e StickerType

		err := e.UnmarshalText([]byte(test.Input))
		assert.NoError(t, err)
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
		Want   string
	}{
		{
			Origin: &MessageOrigin{},
			Want:   "unknown",
		},
		{
			Origin: &MessageOrigin{User: &MessageOriginUser{}},
			Want:   "user",
		},
		{
			Origin: &MessageOrigin{HiddenUser: &MessageOriginHiddenUser{}},
			Want:   "hidden_user",
		},
		{
			Origin: &MessageOrigin{Chat: &MessageOriginChat{}},
			Want:   "chat",
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

		assert.Equal(t, "user", b.Type())
		require.NotNil(t, b.User)
		assert.EqualValues(t, 12345, b.User.Date)
		assert.Equal(t, UserID(1), b.User.SenderUser.ID)
	})

	t.Run("MessageOriginHiddenUser", func(t *testing.T) {
		var b MessageOrigin

		err := b.UnmarshalJSON([]byte(`{"type": "hidden_user", "date": 12345, "sender_user_name": "john doe"}`))
		require.NoError(t, err)

		assert.Equal(t, "hidden_user", b.Type())
		require.NotNil(t, b.HiddenUser)
		assert.EqualValues(t, 12345, b.HiddenUser.Date)
		assert.Equal(t, "john doe", b.HiddenUser.SenderUserName)
	})

	t.Run("MessageOriginChat", func(t *testing.T) {
		var b MessageOrigin

		err := b.UnmarshalJSON([]byte(`{"type": "chat", "date": 12345, "sender_chat": {"id":1}, "author_signature": "john doe"}`))
		require.NoError(t, err)

		assert.Equal(t, "chat", b.Type())
		require.NotNil(t, b.Chat)
		assert.EqualValues(t, 12345, b.Chat.Date)
		assert.Equal(t, ChatID(1), b.Chat.SenderChat.ID)
		assert.Equal(t, "john doe", b.Chat.AuthorSignature)
	})

	t.Run("MessageOriginChannel", func(t *testing.T) {
		var b MessageOrigin

		err := b.UnmarshalJSON([]byte(`{"type": "channel", "date": 12345, "chat": {"id":1}, "message_id": 2, "author_signature": "john doe"}`))
		require.NoError(t, err)

		assert.Equal(t, "channel", b.Type())
		require.NotNil(t, b.Channel)
		assert.EqualValues(t, 12345, b.Channel.Date)
		assert.Equal(t, ChatID(1), b.Channel.Chat.ID)
		assert.Equal(t, 2, b.Channel.MessageID)
		assert.Equal(t, "john doe", b.Channel.AuthorSignature)
	})

	t.Run("Error", func(t *testing.T) {
		var b MessageOrigin

		err := b.UnmarshalJSON([]byte(`{"type": "unknown"`))
		require.Error(t, err)

		err = b.UnmarshalJSON([]byte(`{"type": "unknown", "date": 12345}`))
		require.Error(t, err)
	})
}

func TestReactionType(t *testing.T) {
	t.Run("Emoji", func(t *testing.T) {
		var r ReactionType

		err := r.UnmarshalJSON([]byte(`{"type": "emoji", "emoji": "😀"}`))
		require.NoError(t, err)

		assert.Equal(t, "emoji", r.Type())
		require.NotNil(t, r.Emoji)
		assert.Equal(t, "😀", r.Emoji.Emoji)
	})

	t.Run("CustomEmoji", func(t *testing.T) {
		var r ReactionType

		err := r.UnmarshalJSON([]byte(`{"type": "custom_emoji", "custom_emoji_id": "12345"}`))
		require.NoError(t, err)

		assert.Equal(t, "custom_emoji", r.Type())
		require.NotNil(t, r.CustomEmoji)
		assert.Equal(t, "12345", r.CustomEmoji.CustomEmojiID)
	})

	t.Run("Unknown", func(t *testing.T) {
		var r ReactionType

		err := r.UnmarshalJSON([]byte(`{"type": "unknown"}`))
		require.Error(t, err)
	})
}

func TestReactionType_MarshalJSON(t *testing.T) {
	t.Run("Emoji", func(t *testing.T) {
		r := ReactionType{
			Emoji: &ReactionTypeEmoji{Emoji: "😀"},
		}

		assert.Equal(t, "emoji", r.Type())

		b, err := json.Marshal(r)
		require.NoError(t, err)

		assert.Equal(t, `{"type":"emoji","emoji":"😀"}`, string(b))
	})

	t.Run("CustomEmoji", func(t *testing.T) {
		r := &ReactionType{
			CustomEmoji: &ReactionTypeCustomEmoji{CustomEmojiID: "12345"},
		}

		assert.Equal(t, "custom_emoji", r.Type())

		b, err := json.Marshal(r)
		require.NoError(t, err)

		assert.Equal(t, `{"type":"custom_emoji","custom_emoji_id":"12345"}`, string(b))
	})

	t.Run("Unknown", func(t *testing.T) {
		r := ReactionType{}

		assert.Equal(t, "unknown", r.Type())

		_, err := json.Marshal(r)
		require.Error(t, err)
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
		assert.EqualValues(t, 2, m.InaccessibleMessage.MessageID)
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

func TestWebhookInfo_DateTime(t *testing.T) {
	a := time.Now().Truncate(time.Second)
	b := time.Now().Truncate(time.Second).Add(time.Second)

	w := WebhookInfo{
		LastErrorDate:                a.Unix(),
		LastSynchronizationErrorDate: b.Unix(),
	}

	assert.Equal(t, a, w.LastErrorDateTime())
	assert.Equal(t, b, w.LastSyncronizationErrorDateTime())

}

func TestChat_EmojiStatusExpirationDateTime(t *testing.T) {
	a := time.Now().Truncate(time.Second)

	c := Chat{
		EmojiStatusExpirationDate: a.Unix(),
	}

	assert.Equal(t, a, c.EmojiStatusExpirationDateTime())

}

func TestMessage_DateTime(t *testing.T) {
	a := time.Now().Truncate(time.Second)

	m := Message{
		Date: a.Unix(),
	}

	assert.Equal(t, a, m.DateTime())
}

func TestMessage_EditDateTime(t *testing.T) {
	a := time.Now().Truncate(time.Second)

	m := Message{
		EditDate: a.Unix(),
	}

	assert.Equal(t, a, m.EditDateTime())
}

func TestInaccessibleMessage_DateTime(t *testing.T) {
	a := time.Now().Truncate(time.Second)

	m := InaccessibleMessage{
		Date: a.Unix(),
	}

	assert.Equal(t, a, m.DateTime())
}

func TestMessageOriginUser_DateTime(t *testing.T) {
	a := time.Now().Truncate(time.Second)

	m := MessageOriginUser{
		Date: a.Unix(),
	}

	assert.Equal(t, a, m.DateTime())
}

func TestMessageOriginHiddenUser_DateTime(t *testing.T) {
	a := time.Now().Truncate(time.Second)

	m := MessageOriginHiddenUser{
		Date: a.Unix(),
	}

	assert.Equal(t, a, m.DateTime())
}

func TestMessageOriginChat_DateTime(t *testing.T) {
	a := time.Now().Truncate(time.Second)

	m := MessageOriginChat{
		Date: a.Unix(),
	}

	assert.Equal(t, a, m.DateTime())
}

func TestMessageOriginChannel_DateTime(t *testing.T) {
	a := time.Now().Truncate(time.Second)

	m := MessageOriginChannel{
		Date: a.Unix(),
	}

	assert.Equal(t, a, m.DateTime())
}

func TestPoll_CloseDateTime(t *testing.T) {
	a := time.Now().Truncate(time.Second)

	p := Poll{
		CloseDate: a.Unix(),
	}

	assert.Equal(t, a, p.CloseDateTime())
}

func TestVideoChatScheduled_StartDateTime(t *testing.T) {
	a := time.Now().Truncate(time.Second)

	v := VideoChatScheduled{
		StartDate: a.Unix(),
	}

	assert.Equal(t, a, v.StartDateTime())
}
