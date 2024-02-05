package tgb

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	tg "github.com/mr-linch/go-tg"
	"github.com/stretchr/testify/assert"
)

func testWithClientLocal(
	t *testing.T,
	do func(t *testing.T, ctx context.Context, client *tg.Client),
	handler http.HandlerFunc,
) {
	t.Helper()

	server := httptest.NewServer(handler)
	defer server.Close()

	client := tg.New("12345:secret",
		tg.WithClientServerURL(server.URL),
		tg.WithClientDoer(http.DefaultClient),
	)

	ctx := context.Background()

	do(t, ctx, client)
}

func TestAny(t *testing.T) {
	var (
		filterYes = FilterFunc(func(ctx context.Context, update *Update) (bool, error) {
			return true, nil
		})
		filterNo = FilterFunc(func(ctx context.Context, update *Update) (bool, error) {
			return false, nil
		})
		filterErr = FilterFunc(func(ctx context.Context, update *Update) (bool, error) {
			return false, errors.New("some error")
		})
	)

	allow, err := Any(filterYes, filterNo).Allow(context.Background(), &Update{})
	assert.NoError(t, err)
	assert.True(t, allow)

	allow, err = Any(filterNo, filterNo).Allow(context.Background(), &Update{})
	assert.NoError(t, err)
	assert.False(t, allow)

	allow, err = Any(filterErr, filterYes).Allow(context.Background(), &Update{})
	assert.Error(t, err)
	assert.False(t, allow)
}

func TestAll(t *testing.T) {
	var (
		filterYes = FilterFunc(func(ctx context.Context, update *Update) (bool, error) {
			return true, nil
		})
		filterNo = FilterFunc(func(ctx context.Context, update *Update) (bool, error) {
			return false, nil
		})
		filterErr = FilterFunc(func(ctx context.Context, update *Update) (bool, error) {
			return false, errors.New("some error")
		})
	)

	allow, err := All(filterYes, filterYes).Allow(context.Background(), &Update{})
	assert.NoError(t, err)
	assert.True(t, allow)

	allow, err = All(filterYes, filterNo).Allow(context.Background(), &Update{})
	assert.NoError(t, err)
	assert.False(t, allow)

	allow, err = All(filterYes, filterErr).Allow(context.Background(), &Update{})
	assert.Error(t, err)
	assert.False(t, allow)
}

func TestCommandFilter(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		Name   string
		Filter Filter
		Update *tg.Update
		Allow  bool
		Error  error
	}{
		{
			Name:   "Default",
			Filter: Command("start"),
			Update: &tg.Update{
				Message: &tg.Message{
					Text: "/start azcv 5678",
				},
			},
			Allow: true,
		},
		{
			Name:   "NotMessage",
			Filter: Command("start"),
			Update: &tg.Update{},
			Allow:  false,
		},
		{
			Name:   "ChannelPost",
			Filter: Command("start"),
			Update: &tg.Update{
				ChannelPost: &tg.Message{
					Text: "/start azcv 5678",
				},
			},
			Allow: true,
		},
		{
			Name: "InCaption",
			Filter: Command("start",
				WithCommandIgnoreCaption(false),
			),
			Update: &tg.Update{
				Message: &tg.Message{
					Caption: "/start azcv 5678",
				},
			},
			Allow: true,
		},
		{
			Name: "NoTextOrCaption",
			Filter: Command("start",
				WithCommandIgnoreCaption(false),
			),
			Update: &tg.Update{
				Message: &tg.Message{},
			},
			Allow: false,
		},
		{
			Name:   "BadPrefix",
			Filter: Command("start"),
			Update: &tg.Update{
				Message: &tg.Message{
					Text: "!start azcv 5678",
				},
			},
			Allow: false,
		},
		{
			Name: "CustomPrefix",
			Filter: Command("start",
				WithCommandPrefix("!"),
			),
			Update: &tg.Update{
				Message: &tg.Message{
					Text: "!start azcv 5678",
				},
			},
			Allow: true,
		},
		{
			Name:   "WithSelfMention",
			Filter: Command("start"),
			Update: &tg.Update{
				Message: &tg.Message{
					Text: "/start@go_tg_test_bot azcv 5678",
				},
			},
			Allow: true,
		},
		{
			Name:   "WithNotSelfMention",
			Filter: Command("start"),
			Update: &tg.Update{
				Message: &tg.Message{
					Text: "/start@anybot azcv 5678",
				},
			},
			Allow: false,
		},
		{
			Name:   "NotRegisteredCommand",
			Filter: Command("start"),
			Update: &tg.Update{
				Message: &tg.Message{
					Text: "/help azcv 5678",
				},
			},
			Allow: false,
		},
		{
			Name:   "WithNotSelfMentionAndIgnore",
			Filter: Command("start", WithCommandIgnoreMention(true)),
			Update: &tg.Update{
				Message: &tg.Message{
					Text: "/start@anybot azcv 5678",
				},
			},
			Allow: true,
		},
		{
			Name:   "WithIgnoreCase",
			Filter: Command("start", WithCommandIgnoreCase(false)),
			Update: &tg.Update{
				Message: &tg.Message{
					Text: "/START azcv 5678",
				},
			},
			Allow: false,
		},
		{
			Name:   "WithAlias",
			Filter: Command("start", WithCommandAlias("help")),
			Update: &tg.Update{
				Message: &tg.Message{
					Text: "/help azcv 5678",
				},
			},
			Allow: true,
		},
	} {
		t.Run(test.Name, func(t *testing.T) {
			testWithClientLocal(t, func(t *testing.T, ctx context.Context, client *tg.Client) {
				update := &Update{Update: test.Update, Client: client}

				allow, err := test.Filter.Allow(ctx, update)
				assert.Equal(t, test.Allow, allow)
				assert.Equal(t, test.Error, err)
			}, func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/bot12345:secret/getMe", r.URL.Path)

				w.WriteHeader(http.StatusOK)

				_, _ = w.Write([]byte(`{
					"ok": true,
					"result": {
						"id": 5433024556,
						"is_bot": true,
						"first_name": "go-tg: test bot",
						"username": "go_tg_test_bot",
						"can_join_groups": true,
						"can_read_all_group_messages": true,
						"supports_inline_queries": true
					}
				}`))
			})
		})
	}
}

func TestRegexp(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		Name   string
		Filter Filter
		Update *tg.Update
		Allow  bool
	}{
		{
			Name:   "Message.Text",
			Filter: Regexp(regexp.MustCompile(`go`)),
			Update: &tg.Update{
				Message: &tg.Message{
					Text: "golang",
				},
			},
			Allow: true,
		},
		{
			Name:   "Message.Caption",
			Filter: Regexp(regexp.MustCompile(`go`)),
			Update: &tg.Update{
				Message: &tg.Message{
					Caption: "golang",
				},
			},
			Allow: true,
		},
		{
			Name:   "Message.Poll.Question",
			Filter: Regexp(regexp.MustCompile(`go`)),
			Update: &tg.Update{
				Message: &tg.Message{
					Poll: &tg.Poll{Question: "golang"},
				},
			},
			Allow: true,
		},
		{
			Name:   "Message.CallbackQuery.Data",
			Filter: Regexp(regexp.MustCompile(`go`)),
			Update: &tg.Update{
				CallbackQuery: &tg.CallbackQuery{Data: "golang"},
			},
			Allow: true,
		},
		{
			Name:   "Message.InlineQuery.Query",
			Filter: Regexp(regexp.MustCompile(`go`)),
			Update: &tg.Update{
				InlineQuery: &tg.InlineQuery{Query: "golang"},
			},
			Allow: true,
		},
		{
			Name:   "Message.ChosenInlineResult.Query",
			Filter: Regexp(regexp.MustCompile(`go`)),
			Update: &tg.Update{
				ChosenInlineResult: &tg.ChosenInlineResult{Query: "golang"},
			},
			Allow: true,
		},
		{
			Name:   "Poll.Question",
			Filter: Regexp(regexp.MustCompile(`go`)),
			Update: &tg.Update{
				Poll: &tg.Poll{Question: "golang"},
			},
			Allow: true,
		},
		{
			Name:   "PollAnswer.Question",
			Filter: Regexp(regexp.MustCompile(`go`)),
			Update: &tg.Update{
				PollAnswer: &tg.PollAnswer{},
			},
			Allow: false,
		},
	} {
		t.Run(test.Name, func(t *testing.T) {
			allow, err := test.Filter.Allow(context.Background(), &Update{Update: test.Update})
			assert.Equal(t, test.Allow, allow)
			assert.NoError(t, err)
		})
	}
}

func TestChatType(t *testing.T) {
	for _, test := range []struct {
		Name   string
		Filter Filter
		Update *tg.Update
		Allow  bool
	}{
		{
			"Message",
			ChatType(tg.ChatTypePrivate),
			&tg.Update{
				Message: &tg.Message{
					Chat: tg.Chat{Type: tg.ChatTypePrivate},
				},
			},
			true,
		},
		{
			"EditedMessage",
			ChatType(tg.ChatTypePrivate),
			&tg.Update{
				EditedMessage: &tg.Message{
					Chat: tg.Chat{Type: tg.ChatTypePrivate},
				},
			},
			true,
		},
		{
			"ChannelPost",
			ChatType(tg.ChatTypeChannel),
			&tg.Update{
				ChannelPost: &tg.Message{
					Chat: tg.Chat{Type: tg.ChatTypeChannel},
				},
			},
			true,
		},
		{
			"EditedChannelPost",
			ChatType(tg.ChatTypeChannel),
			&tg.Update{
				EditedChannelPost: &tg.Message{
					Chat: tg.Chat{Type: tg.ChatTypeChannel},
				},
			},
			true,
		},
		{
			"CallbackQuery",
			ChatType(tg.ChatTypePrivate),
			&tg.Update{
				CallbackQuery: &tg.CallbackQuery{
					Message: &tg.MaybeInaccessibleMessage{
						Message: &tg.Message{
							Chat: tg.Chat{Type: tg.ChatTypePrivate},
						},
					},
				},
			},
			true,
		},
		{
			"CallbackQueryNoChat",
			ChatType(tg.ChatTypePrivate),
			&tg.Update{
				CallbackQuery: &tg.CallbackQuery{
					Message: nil,
				},
			},
			false,
		},
		{
			"InlineQuery",
			ChatType(tg.ChatTypeSender),
			&tg.Update{
				InlineQuery: &tg.InlineQuery{
					ChatType: tg.ChatTypeSender,
				},
			},
			true,
		},
		{
			"MyChatMember",
			ChatType(tg.ChatTypeSupergroup),
			&tg.Update{
				MyChatMember: &tg.ChatMemberUpdated{
					Chat: tg.Chat{Type: tg.ChatTypeSupergroup},
				},
			},
			true,
		},
		{
			"ChatMember",
			ChatType(tg.ChatTypeSupergroup),
			&tg.Update{
				ChatMember: &tg.ChatMemberUpdated{
					Chat: tg.Chat{Type: tg.ChatTypeSupergroup},
				},
			},
			true,
		},
		{
			"ChatJoinRequest",
			ChatType(tg.ChatTypeSupergroup),
			&tg.Update{
				ChatJoinRequest: &tg.ChatJoinRequest{
					Chat: tg.Chat{Type: tg.ChatTypeSupergroup},
				},
			},
			true,
		},
		{
			"ShippingQuery",
			ChatType(tg.ChatTypeSupergroup),
			&tg.Update{
				ShippingQuery: &tg.ShippingQuery{},
			},
			false,
		},
	} {
		t.Run(test.Name, func(t *testing.T) {
			allow, err := test.Filter.Allow(context.Background(), &Update{Update: test.Update})
			assert.Equal(t, test.Allow, allow)
			assert.NoError(t, err)
		})
	}
}

func TestMessageType(t *testing.T) {
	for _, test := range []struct {
		Name    string
		Update  *Update
		Allowed []tg.MessageType
		Want    bool
	}{
		{
			Name: "CallbackQueryShouldBeNotAllowed",
			Update: &Update{Update: &tg.Update{
				CallbackQuery: &tg.CallbackQuery{},
			}},
			Allowed: []tg.MessageType{tg.MessageTypeText},
			Want:    false,
		},
		{
			Name: "MessageWithTextShouldBeAllowed",
			Update: &Update{Update: &tg.Update{
				Message: &tg.Message{
					Text: "text",
				},
			}},
			Allowed: []tg.MessageType{tg.MessageTypeText},
			Want:    true,
		},
		{
			Name: "MessageWithPhotoForTextFilterShouldBeNotAllowed",
			Update: &Update{Update: &tg.Update{
				Message: &tg.Message{
					Photo: []tg.PhotoSize{{}},
				},
			}},
			Allowed: []tg.MessageType{tg.MessageTypeText},
			Want:    false,
		},
	} {
		ctx := context.Background()

		t.Run(test.Name, func(t *testing.T) {
			filter := MessageType(test.Allowed...)

			allow, err := filter.Allow(ctx, test.Update)
			assert.Equal(t, test.Want, allow)
			assert.NoError(t, err)
		})
	}
}

func TestTextFuncFilter(t *testing.T) {
	newUpdateMsg := func(text string) *Update {
		return &Update{Update: &tg.Update{
			Message: &tg.Message{
				Text: text,
			},
		}}
	}

	for _, test := range []struct {
		Name   string
		Filter Filter
		Update *Update
		Allow  bool
	}{
		{
			Name:   "NoText",
			Filter: TextEqual(""),
			Update: newUpdateMsg(""),
		},
		{
			Name:   "TextEqual/Allow",
			Filter: TextEqual("text"),
			Update: newUpdateMsg("text"),
			Allow:  true,
		},
		{
			Name:   "TextEqual/Disallow",
			Filter: TextEqual("Text"),
			Update: newUpdateMsg("txet"),
			Allow:  false,
		},
		{
			Name:   "TextEqual/CaseIgnore/Allow",
			Filter: TextEqual("Text", WithTextFuncIgnoreCase(true)),
			Update: newUpdateMsg("text"),
			Allow:  true,
		},
		{
			Name:   "TextEqual/CaseIgnore/Allow/UTF8",
			Filter: TextEqual("Привіт", WithTextFuncIgnoreCase(true)),
			Update: newUpdateMsg("привіт"),
			Allow:  true,
		},
		{
			Name:   "TextEqual/CaseIgnore/Disallow",
			Filter: TextEqual("Tex t", WithTextFuncIgnoreCase(true)),
			Update: newUpdateMsg("text"),
			Allow:  false,
		},
		{
			Name:   "TextHasPrefix/Allow",
			Filter: TextHasPrefix("foo"),
			Update: newUpdateMsg("foobar"),
			Allow:  true,
		},
		{
			Name:   "TextHasPrefix/Disallow",
			Filter: TextHasPrefix("bar"),
			Update: newUpdateMsg("foobar"),
			Allow:  false,
		},
		{
			Name:   "TextHasPrefix/CaseIgnore/Allow",
			Filter: TextHasPrefix("foo", WithTextFuncIgnoreCase(true)),
			Update: newUpdateMsg("Foobar"),
			Allow:  true,
		},
		{
			Name:   "TextHasPrefix/CaseIgnore/Allow/UTF8",
			Filter: TextHasPrefix("При", WithTextFuncIgnoreCase(true)),
			Update: newUpdateMsg("привіт"),
			Allow:  true,
		},
		{
			Name:   "TextHasPrefix/CaseIgnore/Disallow/UTF8",
			Filter: TextHasPrefix("Хай", WithTextFuncIgnoreCase(true)),
			Update: newUpdateMsg("привіт"),
			Allow:  false,
		},
		{
			Name:   "TestHasSuffix/Allow",
			Filter: TextHasSuffix("аша"),
			Update: newUpdateMsg("привіташа"),
			Allow:  true,
		},
		{
			Name:   "TestHasSuffix/Disallow",
			Filter: TextHasSuffix("привіт"),
			Update: newUpdateMsg("привіташа"),
			Allow:  false,
		},
		{
			Name:   "TestHasSuffix/CaseIgnore/Allow",
			Filter: TextHasSuffix("аша", WithTextFuncIgnoreCase(true)),
			Update: newUpdateMsg("ПривітАша"),
			Allow:  true,
		},
		{
			Name:   "TextContains/Allow",
			Filter: TextContains("каш"),
			Update: newUpdateMsg("акашка"),
			Allow:  true,
		},
		{
			Name:   "TextContains/Disallow",
			Filter: TextContains("каш"),
			Update: newUpdateMsg("саш"),
			Allow:  false,
		},
		{
			Name:   "TextContains/CaseIgnore/Allow",
			Filter: TextContains("каш", WithTextFuncIgnoreCase(true)),
			Update: newUpdateMsg("Каша"),
			Allow:  true,
		},
		{
			Name:   "TextContains/CaseIgnore/Disallow",
			Filter: TextContains("каш", WithTextFuncIgnoreCase(true)),
			Update: newUpdateMsg("Саша"),
			Allow:  false,
		},
		{
			Name:   "TextIn/Allow",
			Filter: TextIn([]string{"1", "2", "3"}),
			Update: newUpdateMsg("2"),
			Allow:  true,
		},
		{
			Name:   "TextIn/Disallow",
			Filter: TextIn([]string{"1", "2", "3"}),
			Update: newUpdateMsg("4"),
			Allow:  false,
		},
		{
			Name:   "TextIn/CaseIgnore/Allow",
			Filter: TextIn([]string{"A", "B", "C"}, WithTextFuncIgnoreCase(true)),
			Update: newUpdateMsg("b"),
			Allow:  true,
		},
		{
			Name:   "TextIn/CaseIgnore/Disallow",
			Filter: TextIn([]string{"A", "B", "C"}, WithTextFuncIgnoreCase(true)),
			Update: newUpdateMsg("f"),
			Allow:  false,
		},
	} {
		t.Run(test.Name, func(t *testing.T) {
			allow, err := test.Filter.Allow(context.Background(), test.Update)
			assert.Equal(t, test.Allow, allow)
			assert.NoError(t, err)
		})
	}
}

func TestMessageEntity(t *testing.T) {
	for _, test := range []struct {
		Name   string
		Update *Update
		Filter Filter
		Allow  bool
	}{
		{
			Name: "NotMessage",
			Update: &Update{Update: &tg.Update{
				CallbackQuery: &tg.CallbackQuery{},
			}},
			Filter: MessageEntity(tg.MessageEntityTypeEmail),
			Allow:  false,
		},
		{
			Name: "MessageWithoutEntity",
			Update: &Update{Update: &tg.Update{
				Message: &tg.Message{
					Text: "text",
				},
			}},
			Filter: MessageEntity(tg.MessageEntityTypeEmail),
			Allow:  false,
		},
		{
			Name: "MessageWithoutSpecifiedEntity",
			Update: &Update{Update: &tg.Update{
				Message: &tg.Message{
					Text: "test@test.com",
					Entities: []tg.MessageEntity{
						{
							Type:   tg.MessageEntityTypeEmail,
							Offset: 0,
							Length: 13,
						},
					},
				},
			}},
			Filter: MessageEntity(tg.MessageEntityTypeHashtag),
			Allow:  false,
		},
		{
			Name: "MessageWithSpecifiedEntity",
			Update: &Update{Update: &tg.Update{
				Message: &tg.Message{
					Text: "test@test.com",
					Entities: []tg.MessageEntity{
						{
							Type:   tg.MessageEntityTypeEmail,
							Offset: 0,
							Length: 13,
						},
					},
				},
			}},
			Filter: MessageEntity(tg.MessageEntityTypeEmail),
			Allow:  true,
		},
		{
			Name: "MessageWithSpecifiedCaptionEntity",
			Update: &Update{Update: &tg.Update{
				Message: &tg.Message{
					Caption: "test@test.com",
					CaptionEntities: []tg.MessageEntity{
						{
							Type:   tg.MessageEntityTypeEmail,
							Offset: 0,
							Length: 13,
						},
					},
				},
			}},
			Filter: MessageEntity(tg.MessageEntityTypeEmail, tg.MessageEntityTypeBold),
			Allow:  true,
		},
		{
			Name: "PollWithSpecifiedEntity",
			Update: &Update{Update: &tg.Update{
				Poll: &tg.Poll{
					Explanation: "test@test.com",
					ExplanationEntities: []tg.MessageEntity{
						{
							Type:   tg.MessageEntityTypeEmail,
							Offset: 0,
							Length: 13,
						},
					},
				},
			}},
			Filter: MessageEntity(tg.MessageEntityTypeEmail),
			Allow:  true,
		},
		{
			Name: "MessageGameWithSpecifiedEntity",
			Update: &Update{Update: &tg.Update{
				Message: &tg.Message{
					Game: &tg.Game{
						Text: "test@test.com",
						TextEntities: []tg.MessageEntity{
							{
								Type:   tg.MessageEntityTypeEmail,
								Offset: 0,
								Length: 13,
							},
						},
					},
				},
			}},
			Filter: MessageEntity(tg.MessageEntityTypeEmail),
			Allow:  true,
		},
		{
			Name: "MessagePollWithSpecifiedEntity",
			Update: &Update{Update: &tg.Update{
				Message: &tg.Message{
					Poll: &tg.Poll{
						Explanation: "test@test.com",
						ExplanationEntities: []tg.MessageEntity{
							{
								Type:   tg.MessageEntityTypeEmail,
								Offset: 0,
								Length: 13,
							},
						},
					},
				},
			}},
			Filter: MessageEntity(tg.MessageEntityTypeEmail),
			Allow:  true,
		},
	} {
		t.Run(test.Name, func(t *testing.T) {
			allow, err := test.Filter.Allow(context.Background(), test.Update)

			assert.Equal(t, test.Allow, allow)
			assert.NoError(t, err)
		})
	}
}

func TestNot(t *testing.T) {
	constFilter := func(v bool, err error) Filter {
		return FilterFunc(func(ctx context.Context, update *Update) (bool, error) {
			return v, err
		})
	}

	trueFilter := constFilter(true, nil)

	allow, err := Not(trueFilter).Allow(context.Background(), &Update{})
	assert.False(t, allow)
	assert.NoError(t, err)

	falseFilter := constFilter(false, nil)

	allow, err = Not(falseFilter).Allow(context.Background(), &Update{})
	assert.True(t, allow)
	assert.NoError(t, err)

	errFilter := constFilter(true, errors.New("test"))
	allow, err = Not(errFilter).Allow(context.Background(), &Update{})
	assert.False(t, allow)
	assert.Error(t, err)
}
