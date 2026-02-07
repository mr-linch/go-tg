package methodgen

import (
	"bytes"
	"flag"
	"io"
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mr-linch/go-tg/gen/config"
	"github.com/mr-linch/go-tg/gen/ir"
	"github.com/mr-linch/go-tg/gen/parser"
)

var updateGolden = flag.Bool("update", false, "update golden files")

var testLog = slog.New(slog.NewTextHandler(io.Discard, nil))

func loadTestConfig(t *testing.T) *config.MethodGen {
	t.Helper()
	cfg, err := config.LoadFile("../config.yaml")
	require.NoError(t, err)
	return &cfg.MethodGen
}

var testAPI = &ir.API{
	Types: []ir.Type{
		{Name: "Update"},
		{Name: "Message"},
	},
	Methods: []ir.Method{
		// Method with no required params, returns array
		{
			Name:        "getUpdates",
			Description: []string{"Use this method to receive incoming updates using long polling ([wiki](https://en.wikipedia.org/wiki/Push_technology#Long_polling)).", "Returns an Array of [Update](https://core.telegram.org/bots/api#update) objects."},
			Params: []ir.Param{
				{Name: "offset", TypeExpr: ir.TypeExpr{Types: []ir.TypeRef{{Type: "Integer"}}}, Description: "Identifier of the first update to be returned."},
				{Name: "limit", TypeExpr: ir.TypeExpr{Types: []ir.TypeRef{{Type: "Integer"}}}, Description: "Limits the number of updates to be retrieved."},
				{Name: "timeout", TypeExpr: ir.TypeExpr{Types: []ir.TypeRef{{Type: "Integer"}}}, Description: "Timeout in seconds for long polling."},
				{Name: "allowed_updates", TypeExpr: ir.TypeExpr{Types: []ir.TypeRef{{Type: "String"}}, Array: 1}, Description: "List of update types."},
			},
			Returns: ir.TypeExpr{Types: []ir.TypeRef{{Type: "Update"}}, Array: 1},
		},
		// Method with required PeerID + string, returns Message
		{
			Name:        "sendMessage",
			Description: []string{"Use this method to send text messages. On success, the sent [Message](https://core.telegram.org/bots/api#message) is returned."},
			Params: []ir.Param{
				{Name: "chat_id", TypeExpr: ir.TypeExpr{Types: []ir.TypeRef{{Type: "Integer"}, {Type: "String"}}}, Required: true, Description: "Unique identifier for the target chat or username."},
				{Name: "text", TypeExpr: ir.TypeExpr{Types: []ir.TypeRef{{Type: "String"}}}, Required: true, Description: "Text of the message to be sent. See [formatting options](https://core.telegram.org/bots/api#formatting-options) for more details."},
				{Name: "parse_mode", TypeExpr: ir.TypeExpr{Types: []ir.TypeRef{{Type: "String"}}}, Description: "Mode for parsing entities."},
				{Name: "reply_markup", TypeExpr: ir.TypeExpr{Types: []ir.TypeRef{{Type: "InlineKeyboardMarkup"}, {Type: "ReplyKeyboardMarkup"}}}, Description: "Additional interface options."},
			},
			Returns: ir.TypeExpr{Types: []ir.TypeRef{{Type: "Message"}}},
		},
		// Method returning True (CallNoResult)
		{
			Name:        "deleteWebhook",
			Description: []string{"Use this method to remove webhook integration."},
			Params: []ir.Param{
				{Name: "drop_pending_updates", TypeExpr: ir.TypeExpr{Types: []ir.TypeRef{{Type: "Boolean"}}}, Description: "Pass True to drop all pending updates."},
			},
			Returns: ir.TypeExpr{Types: []ir.TypeRef{{Type: "True"}}},
		},
		// Method with FileArg (file_id in description)
		{
			Name:        "sendPhoto",
			Description: []string{"Use this method to send photos."},
			Params: []ir.Param{
				{Name: "chat_id", TypeExpr: ir.TypeExpr{Types: []ir.TypeRef{{Type: "Integer"}, {Type: "String"}}}, Required: true, Description: "Unique identifier for the target chat."},
				{Name: "photo", TypeExpr: ir.TypeExpr{Types: []ir.TypeRef{{Type: "InputFile"}}}, Required: true, Description: "Photo to send. Pass a file_id as String or upload."},
				{Name: "caption", TypeExpr: ir.TypeExpr{Types: []ir.TypeRef{{Type: "String"}}}, Description: "Photo caption."},
			},
			Returns: ir.TypeExpr{Types: []ir.TypeRef{{Type: "Message"}}},
		},
		// Method with InputFile (no file_id)
		{
			Name:        "setWebhook",
			Description: []string{"Use this method to specify a URL and receive incoming updates."},
			Params: []ir.Param{
				{Name: "url", TypeExpr: ir.TypeExpr{Types: []ir.TypeRef{{Type: "String"}}}, Required: true, Description: "HTTPS URL to send updates to."},
				{Name: "certificate", TypeExpr: ir.TypeExpr{Types: []ir.TypeRef{{Type: "InputFile"}}}, Description: "Upload your public key certificate."},
			},
			Returns: ir.TypeExpr{Types: []ir.TypeRef{{Type: "True"}}},
		},
		// Method with InputMedia slice
		{
			Name:        "sendMediaGroup",
			Description: []string{"Use this method to send a group of photos or videos as an album."},
			Params: []ir.Param{
				{Name: "chat_id", TypeExpr: ir.TypeExpr{Types: []ir.TypeRef{{Type: "Integer"}, {Type: "String"}}}, Required: true, Description: "Unique identifier for the target chat."},
				{Name: "media", TypeExpr: ir.TypeExpr{Types: []ir.TypeRef{{Type: "InputMedia"}}, Array: 1}, Required: true, Description: "Array of media to send."},
			},
			Returns: ir.TypeExpr{Types: []ir.TypeRef{{Type: "Message"}}, Array: 1},
		},
		// Method with UserID param
		{
			Name:        "getUserProfilePhotos",
			Description: []string{"Use this method to get a list of profile pictures."},
			Params: []ir.Param{
				{Name: "user_id", TypeExpr: ir.TypeExpr{Types: []ir.TypeRef{{Type: "Integer64"}}}, Required: true, Description: "Unique identifier of the target user."},
				{Name: "offset", TypeExpr: ir.TypeExpr{Types: []ir.TypeRef{{Type: "Integer"}}}, Description: "Sequential number of the first photo."},
				{Name: "limit", TypeExpr: ir.TypeExpr{Types: []ir.TypeRef{{Type: "Integer"}}}, Description: "Limits the number of photos."},
			},
			Returns: ir.TypeExpr{Types: []ir.TypeRef{{Type: "UserProfilePhotos"}}},
		},
	},
}

func TestGenerate(t *testing.T) {
	cfg := loadTestConfig(t)

	var buf bytes.Buffer
	err := Generate(testAPI, &buf, cfg, testLog, Options{Package: "tg"})
	require.NoError(t, err)

	golden := "testdata/methods_gen.golden.go"
	if *updateGolden {
		err := os.MkdirAll("testdata", 0o750)
		require.NoError(t, err)
		err = os.WriteFile(golden, buf.Bytes(), 0o600)
		require.NoError(t, err)
	}

	expected, err := os.ReadFile(golden)
	require.NoError(t, err)
	assert.Equal(t, string(expected), buf.String())
}

func TestGenerate_CallNoResult(t *testing.T) {
	api := &ir.API{
		Methods: []ir.Method{
			{
				Name:        "deleteWebhook",
				Description: []string{"Use this method to remove webhook integration."},
				Returns:     ir.TypeExpr{Types: []ir.TypeRef{{Type: "True"}}},
			},
		},
	}

	var buf bytes.Buffer
	err := Generate(api, &buf, &config.MethodGen{}, testLog, Options{Package: "tg"})
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "type DeleteWebhookCall struct {")
	assert.Contains(t, output, "CallNoResult")
	assert.NotContains(t, output, "Call[")
}

func TestGenerate_ArrayReturn(t *testing.T) {
	api := &ir.API{
		Methods: []ir.Method{
			{
				Name:    "getUpdates",
				Returns: ir.TypeExpr{Types: []ir.TypeRef{{Type: "Update"}}, Array: 1},
			},
		},
	}

	var buf bytes.Buffer
	err := Generate(api, &buf, &config.MethodGen{}, testLog, Options{Package: "tg"})
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "Call[[]Update]")
}

func TestGenerate_RequiredParams(t *testing.T) {
	api := &ir.API{
		Methods: []ir.Method{
			{
				Name: "sendMessage",
				Params: []ir.Param{
					{Name: "chat_id", TypeExpr: ir.TypeExpr{Types: []ir.TypeRef{{Type: "Integer"}, {Type: "String"}}}, Required: true},
					{Name: "text", TypeExpr: ir.TypeExpr{Types: []ir.TypeRef{{Type: "String"}}}, Required: true},
					{Name: "parse_mode", TypeExpr: ir.TypeExpr{Types: []ir.TypeRef{{Type: "String"}}}},
				},
				Returns: ir.TypeExpr{Types: []ir.TypeRef{{Type: "Message"}}},
			},
		},
	}

	cfg := loadTestConfig(t)
	var buf bytes.Buffer
	err := Generate(api, &buf, cfg, testLog, Options{Package: "tg"})
	require.NoError(t, err)

	output := buf.String()
	// Constructor has both required params
	assert.Contains(t, output, "func NewSendMessageCall(chatID PeerID, text string)")
	// Optional param not in constructor
	assert.NotContains(t, output, "NewSendMessageCall(chatID PeerID, text string, parseMode")
	// But has setter (ParseMode type from rule match)
	assert.Contains(t, output, "func (call *SendMessageCall) ParseMode(parseMode ParseMode)")
}

func TestGenerate_PeerID(t *testing.T) {
	api := &ir.API{
		Methods: []ir.Method{
			{
				Name: "sendMessage",
				Params: []ir.Param{
					{Name: "chat_id", TypeExpr: ir.TypeExpr{Types: []ir.TypeRef{{Type: "Integer"}, {Type: "String"}}}, Required: true},
				},
				Returns: ir.TypeExpr{Types: []ir.TypeRef{{Type: "Message"}}},
			},
		},
	}

	cfg := loadTestConfig(t)
	var buf bytes.Buffer
	err := Generate(api, &buf, cfg, testLog, Options{Package: "tg"})
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "chatID PeerID")
	assert.Contains(t, output, "PeerID(\"chat_id\", chatID)")
}

func TestGenerate_FileArg(t *testing.T) {
	api := &ir.API{
		Methods: []ir.Method{
			{
				Name: "sendPhoto",
				Params: []ir.Param{
					{Name: "chat_id", TypeExpr: ir.TypeExpr{Types: []ir.TypeRef{{Type: "Integer"}, {Type: "String"}}}, Required: true},
					{Name: "photo", TypeExpr: ir.TypeExpr{Types: []ir.TypeRef{{Type: "InputFile"}}}, Required: true, Description: "Pass a file_id or upload."},
				},
				Returns: ir.TypeExpr{Types: []ir.TypeRef{{Type: "Message"}}},
			},
		},
	}

	cfg := loadTestConfig(t)
	var buf bytes.Buffer
	err := Generate(api, &buf, cfg, testLog, Options{Package: "tg"})
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "photo FileArg")
	assert.Contains(t, output, "File(\"photo\", photo)")
}

func TestGenerate_InputFile(t *testing.T) {
	api := &ir.API{
		Methods: []ir.Method{
			{
				Name: "setWebhook",
				Params: []ir.Param{
					{Name: "url", TypeExpr: ir.TypeExpr{Types: []ir.TypeRef{{Type: "String"}}}, Required: true},
					{Name: "certificate", TypeExpr: ir.TypeExpr{Types: []ir.TypeRef{{Type: "InputFile"}}}, Description: "Upload your certificate."},
				},
				Returns: ir.TypeExpr{Types: []ir.TypeRef{{Type: "True"}}},
			},
		},
	}

	cfg := loadTestConfig(t)
	var buf bytes.Buffer
	err := Generate(api, &buf, cfg, testLog, Options{Package: "tg"})
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "certificate InputFile")
	assert.Contains(t, output, "JSON(\"certificate\", certificate)")
}

func TestGenerate_InputMediaSlice(t *testing.T) {
	t.Run("DirectType_NoDiscriminator", func(t *testing.T) {
		// Without discriminator union type definition, []InputMedia stays as-is
		api := &ir.API{
			Methods: []ir.Method{
				{
					Name: "sendMediaGroup",
					Params: []ir.Param{
						{Name: "chat_id", TypeExpr: ir.TypeExpr{Types: []ir.TypeRef{{Type: "Integer"}, {Type: "String"}}}, Required: true},
						{Name: "media", TypeExpr: ir.TypeExpr{Types: []ir.TypeRef{{Type: "InputMedia"}}, Array: 1}, Required: true},
					},
					Returns: ir.TypeExpr{Types: []ir.TypeRef{{Type: "Message"}}, Array: 1},
				},
			},
		}

		cfg := loadTestConfig(t)
		var buf bytes.Buffer
		err := Generate(api, &buf, cfg, testLog, Options{Package: "tg"})
		require.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, "media []InputMedia")
		assert.Contains(t, output, "InputMediaSlice(\"media\", media)")
	})

	t.Run("DirectType_WithDiscriminator", func(t *testing.T) {
		// With discriminator union, []InputMedia stays as []InputMedia (not variadic)
		api := &ir.API{
			Types: []ir.Type{
				{Name: "InputMedia", Subtypes: []string{"InputMediaPhoto", "InputMediaVideo"}},
				{Name: "InputMediaPhoto", Fields: []ir.Field{
					{Name: "type", Const: "photo", TypeExpr: ir.TypeExpr{Types: []ir.TypeRef{{Type: "String"}}}},
					{Name: "media", TypeExpr: ir.TypeExpr{Types: []ir.TypeRef{{Type: "String"}}}},
				}},
				{Name: "InputMediaVideo", Fields: []ir.Field{
					{Name: "type", Const: "video", TypeExpr: ir.TypeExpr{Types: []ir.TypeRef{{Type: "String"}}}},
					{Name: "media", TypeExpr: ir.TypeExpr{Types: []ir.TypeRef{{Type: "String"}}}},
				}},
			},
			Methods: []ir.Method{
				{
					Name: "sendMediaGroup",
					Params: []ir.Param{
						{Name: "chat_id", TypeExpr: ir.TypeExpr{Types: []ir.TypeRef{{Type: "Integer"}, {Type: "String"}}}, Required: true},
						{Name: "media", TypeExpr: ir.TypeExpr{Types: []ir.TypeRef{{Type: "InputMedia"}}, Array: 1}, Required: true},
					},
					Returns: ir.TypeExpr{Types: []ir.TypeRef{{Type: "Message"}}, Array: 1},
				},
			},
		}

		cfg := loadTestConfig(t)
		var buf bytes.Buffer
		err := Generate(api, &buf, cfg, testLog, Options{Package: "tg"})
		require.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, "media []InputMedia")
		assert.Contains(t, output, "InputMediaSlice(\"media\", media)")
	})

	t.Run("UnionSubtypes", func(t *testing.T) {
		// Real API spec lists individual subtypes, not the parent union
		api := &ir.API{
			Types: []ir.Type{
				{Name: "InputMedia", Subtypes: []string{"InputMediaAudio", "InputMediaDocument", "InputMediaPhoto", "InputMediaVideo"}},
			},
			Methods: []ir.Method{
				{
					Name: "sendMediaGroup",
					Params: []ir.Param{
						{Name: "chat_id", TypeExpr: ir.TypeExpr{Types: []ir.TypeRef{{Type: "Integer"}, {Type: "String"}}}, Required: true},
						{Name: "media", TypeExpr: ir.TypeExpr{Types: []ir.TypeRef{
							{Type: "InputMediaAudio"},
							{Type: "InputMediaDocument"},
							{Type: "InputMediaPhoto"},
							{Type: "InputMediaVideo"},
						}, Array: 1}, Required: true},
					},
					Returns: ir.TypeExpr{Types: []ir.TypeRef{{Type: "Message"}}, Array: 1},
				},
			},
		}

		cfg := loadTestConfig(t)
		var buf bytes.Buffer
		err := Generate(api, &buf, cfg, testLog, Options{Package: "tg"})
		require.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, "media []InputMedia")
		assert.Contains(t, output, "InputMediaSlice(\"media\", media)")
	})
}

func TestGenerate_InputMediaSlice_HeterogeneousUnions(t *testing.T) {
	api := &ir.API{
		Types: []ir.Type{
			{Name: "InputMedia", Subtypes: []string{"InputMediaPhoto", "InputMediaVideo"}},
			{Name: "InputPaidMedia", Subtypes: []string{"InputPaidMediaPhoto", "InputPaidMediaVideo"}},
		},
		Methods: []ir.Method{
			{
				Name: "hypotheticalMethod",
				Params: []ir.Param{
					{Name: "chat_id", TypeExpr: ir.TypeExpr{Types: []ir.TypeRef{{Type: "Integer"}, {Type: "String"}}}, Required: true},
					{Name: "media", TypeExpr: ir.TypeExpr{Types: []ir.TypeRef{
						{Type: "InputMediaPhoto"},
						{Type: "InputPaidMediaVideo"},
					}, Array: 1}, Required: true},
				},
				Returns: ir.TypeExpr{Types: []ir.TypeRef{{Type: "Message"}}},
			},
		},
	}

	cfg := loadTestConfig(t)
	var buf bytes.Buffer
	err := Generate(api, &buf, cfg, testLog, Options{Package: "tg"})
	require.NoError(t, err)

	output := buf.String()
	// Mixed subtypes from different unions must fall back to any + JSON
	assert.Contains(t, output, "media any")
	assert.Contains(t, output, "JSON(\"media\", media)")
	assert.NotContains(t, output, "InputMediaSlice")
	assert.NotContains(t, output, "InputPaidMediaSlice")
}

func TestGenerate_InputMedia(t *testing.T) {
	t.Run("ScalarWithoutUnionType", func(t *testing.T) {
		// Without union type definition, InputMedia is treated as plain type
		api := &ir.API{
			Methods: []ir.Method{
				{
					Name: "editMessageMedia",
					Params: []ir.Param{
						{Name: "chat_id", TypeExpr: ir.TypeExpr{Types: []ir.TypeRef{{Type: "Integer"}, {Type: "String"}}}, Required: true},
						{Name: "media", TypeExpr: ir.TypeExpr{Types: []ir.TypeRef{{Type: "InputMedia"}}}, Required: true},
					},
					Returns: ir.TypeExpr{Types: []ir.TypeRef{{Type: "Message"}}},
				},
			},
		}

		cfg := loadTestConfig(t)
		var buf bytes.Buffer
		err := Generate(api, &buf, cfg, testLog, Options{Package: "tg"})
		require.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, "media InputMedia")
		assert.Contains(t, output, "InputMedia(\"media\", media)")
	})

	t.Run("ScalarWithDiscriminatorUnion", func(t *testing.T) {
		// With discriminator union type, scalar param uses Class interface
		api := &ir.API{
			Types: []ir.Type{
				{Name: "InputMedia", Subtypes: []string{"InputMediaPhoto", "InputMediaVideo"}},
				{Name: "InputMediaPhoto", Fields: []ir.Field{
					{Name: "type", Const: "photo", TypeExpr: ir.TypeExpr{Types: []ir.TypeRef{{Type: "String"}}}},
					{Name: "media", TypeExpr: ir.TypeExpr{Types: []ir.TypeRef{{Type: "String"}}}},
				}},
				{Name: "InputMediaVideo", Fields: []ir.Field{
					{Name: "type", Const: "video", TypeExpr: ir.TypeExpr{Types: []ir.TypeRef{{Type: "String"}}}},
					{Name: "media", TypeExpr: ir.TypeExpr{Types: []ir.TypeRef{{Type: "String"}}}},
				}},
			},
			Methods: []ir.Method{
				{
					Name: "editMessageMedia",
					Params: []ir.Param{
						{Name: "chat_id", TypeExpr: ir.TypeExpr{Types: []ir.TypeRef{{Type: "Integer"}, {Type: "String"}}}, Required: true},
						{Name: "media", TypeExpr: ir.TypeExpr{Types: []ir.TypeRef{{Type: "InputMedia"}}}, Required: true},
					},
					Returns: ir.TypeExpr{Types: []ir.TypeRef{{Type: "Message"}}},
				},
			},
		}

		cfg := loadTestConfig(t)
		var buf bytes.Buffer
		err := Generate(api, &buf, cfg, testLog, Options{Package: "tg"})
		require.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, "media InputMediaClass")
		assert.Contains(t, output, `InputMedia("media", media.AsInputMedia())`)
	})
}

func TestGenerate_InputPaidMediaSlice(t *testing.T) {
	// Without discriminator union, stays as plain slice
	api := &ir.API{
		Methods: []ir.Method{
			{
				Name: "sendPaidMedia",
				Params: []ir.Param{
					{Name: "chat_id", TypeExpr: ir.TypeExpr{Types: []ir.TypeRef{{Type: "Integer"}, {Type: "String"}}}, Required: true},
					{Name: "media", TypeExpr: ir.TypeExpr{Types: []ir.TypeRef{{Type: "InputPaidMedia"}}, Array: 1}, Required: true},
				},
				Returns: ir.TypeExpr{Types: []ir.TypeRef{{Type: "Message"}}},
			},
		},
	}

	cfg := loadTestConfig(t)
	var buf bytes.Buffer
	err := Generate(api, &buf, cfg, testLog, Options{Package: "tg"})
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "media []InputPaidMedia")
	assert.Contains(t, output, "InputPaidMediaSlice(\"media\", media)")
}

func TestGenerate_ParseMode(t *testing.T) {
	api := &ir.API{
		Methods: []ir.Method{
			{
				Name: "sendMessage",
				Params: []ir.Param{
					{Name: "parse_mode", TypeExpr: ir.TypeExpr{Types: []ir.TypeRef{{Type: "String"}}}},
				},
				Returns: ir.TypeExpr{Types: []ir.TypeRef{{Type: "Message"}}},
			},
		},
	}

	cfg := loadTestConfig(t)
	var buf bytes.Buffer
	err := Generate(api, &buf, cfg, testLog, Options{Package: "tg"})
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "parseMode ParseMode")
	assert.Contains(t, output, "Stringer(\"parse_mode\", parseMode)")
}

func TestGenerate_UserID(t *testing.T) {
	api := &ir.API{
		Methods: []ir.Method{
			{
				Name: "getUserProfilePhotos",
				Params: []ir.Param{
					{Name: "user_id", TypeExpr: ir.TypeExpr{Types: []ir.TypeRef{{Type: "Integer64"}}}, Required: true},
				},
				Returns: ir.TypeExpr{Types: []ir.TypeRef{{Type: "UserProfilePhotos"}}},
			},
		},
	}

	cfg := loadTestConfig(t)
	var buf bytes.Buffer
	err := Generate(api, &buf, cfg, testLog, Options{Package: "tg"})
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "userID UserID")
	assert.Contains(t, output, "UserID(\"user_id\", userID)")
}

func TestGenerate_FullAPI(t *testing.T) {
	// Test with full Telegram API to ensure no panics or errors.
	f, err := os.Open("../parser/testdata/index.html")
	require.NoError(t, err)
	defer f.Close()

	api, err := parser.Parse(f)
	require.NoError(t, err)

	cfg := loadTestConfig(t)
	var buf bytes.Buffer
	err = Generate(api, &buf, cfg, testLog, Options{Package: "tg"})
	require.NoError(t, err)

	// Basic sanity checks
	output := buf.String()
	assert.Contains(t, output, "package tg")
	assert.Contains(t, output, "GetUpdatesCall")
	assert.Contains(t, output, "SendMessageCall")
	assert.Contains(t, output, "SendPhotoCall")

	// Verify InputMedia methods use correct request methods
	assert.Contains(t, output, "InputMediaSlice(\"media\", media)")
	assert.Contains(t, output, `InputMedia("media", media.AsInputMedia())`)
	assert.Contains(t, output, "InputPaidMediaSlice(\"media\", media)")

	// Scalar discriminator union params use Class interfaces
	assert.Contains(t, output, "media InputMediaClass")
	// Slice discriminator union params stay as concrete slices
	assert.Contains(t, output, "media []InputMedia")
}

func TestGenerate_PackageOption(t *testing.T) {
	api := &ir.API{
		Methods: []ir.Method{
			{Name: "getMe", Returns: ir.TypeExpr{Types: []ir.TypeRef{{Type: "User"}}}},
		},
	}

	var buf bytes.Buffer
	err := Generate(api, &buf, &config.MethodGen{}, testLog, Options{Package: "telegram"})
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "package telegram")
}

func TestGenerate_NamingConventions(t *testing.T) {
	api := &ir.API{
		Methods: []ir.Method{
			{
				Name: "sendMessage",
				Params: []ir.Param{
					{Name: "from_chat_id", TypeExpr: ir.TypeExpr{Types: []ir.TypeRef{{Type: "Integer"}, {Type: "String"}}}, Required: true},
					{Name: "message_thread_id", TypeExpr: ir.TypeExpr{Types: []ir.TypeRef{{Type: "Integer"}}}},
				},
				Returns: ir.TypeExpr{Types: []ir.TypeRef{{Type: "Message"}}},
			},
		},
	}

	cfg := loadTestConfig(t)
	var buf bytes.Buffer
	err := Generate(api, &buf, cfg, testLog, Options{Package: "tg"})
	require.NoError(t, err)

	output := buf.String()
	// PascalCase setter name
	assert.Contains(t, output, "func (call *SendMessageCall) FromChatID(fromChatID PeerID)")
	// camelCase arg name with initialism
	assert.Contains(t, output, "fromChatID PeerID")
	// PascalCase setter with initialism
	assert.Contains(t, output, "func (call *SendMessageCall) MessageThreadID(messageThreadID int)")
}
