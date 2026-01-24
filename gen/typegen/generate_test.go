package typegen

import (
	"bytes"
	"flag"
	"io"
	"log/slog"
	"os"
	"testing"

	"github.com/mr-linch/go-tg/gen/config"
	"github.com/mr-linch/go-tg/gen/ir"
	"github.com/mr-linch/go-tg/gen/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var updateGolden = flag.Bool("update", false, "update golden files")

var testLog = slog.New(slog.NewTextHandler(io.Discard, nil))

func loadTestConfig(t *testing.T) *config.TypeGen {
	t.Helper()
	cfg, err := config.LoadFile("../config.yaml")
	require.NoError(t, err)
	return &cfg.TypeGen
}

var testAPI = &ir.API{
	Types: []ir.Type{
		{
			Name:        "Update",
			Description: "This object represents an incoming update.",
			Fields: []ir.Field{
				{
					Name:        "update_id",
					TypeExpr:    ir.TypeExpr{Types: []ir.TypeRef{{Type: "Integer"}}},
					Description: "The update's unique identifier.",
				},
				{
					Name:        "message",
					TypeExpr:    ir.TypeExpr{Types: []ir.TypeRef{{Type: "Message"}}},
					Optional:    true,
					Description: "Optional. New incoming message.",
				},
			},
		},
		{
			Name:        "WebhookInfo",
			Description: "Describes the current status of a webhook.",
			Fields: []ir.Field{
				{
					Name:        "url",
					TypeExpr:    ir.TypeExpr{Types: []ir.TypeRef{{Type: "String"}}},
					Description: "Webhook URL.",
				},
				{
					Name:        "last_error_date",
					TypeExpr:    ir.TypeExpr{Types: []ir.TypeRef{{Type: "Integer"}}},
					Optional:    true,
					Description: "Optional. Unix time for the most recent error.",
				},
				{
					Name:        "allowed_updates",
					TypeExpr:    ir.TypeExpr{Types: []ir.TypeRef{{Type: "String"}}, Array: 1},
					Optional:    true,
					Description: "Optional. A list of update types the bot is subscribed to.",
				},
			},
		},
		{
			Name:        "User",
			Description: "This object represents a Telegram user or bot.",
			Fields: []ir.Field{
				{
					Name:        "id",
					TypeExpr:    ir.TypeExpr{Types: []ir.TypeRef{{Type: "Integer64"}}},
					Description: "Unique identifier for this user or bot.",
				},
				{
					Name:        "is_bot",
					TypeExpr:    ir.TypeExpr{Types: []ir.TypeRef{{Type: "Boolean"}}},
					Description: "True, if this user is a bot.",
				},
				{
					Name:        "username",
					TypeExpr:    ir.TypeExpr{Types: []ir.TypeRef{{Type: "String"}}},
					Optional:    true,
					Description: "Optional. User's or bot's username.",
				},
			},
		},
		{
			Name:        "Message",
			Description: "This object represents a message.",
			Fields: []ir.Field{
				{
					Name:        "message_id",
					TypeExpr:    ir.TypeExpr{Types: []ir.TypeRef{{Type: "Integer"}}},
					Description: "Unique message identifier.",
				},
				{
					Name:        "date",
					TypeExpr:    ir.TypeExpr{Types: []ir.TypeRef{{Type: "Integer"}}},
					Description: "Date the message was sent in Unix time.",
				},
				{
					Name:        "photo",
					TypeExpr:    ir.TypeExpr{Types: []ir.TypeRef{{Type: "PhotoSize"}}, Array: 1},
					Optional:    true,
					Description: "Optional. Available sizes of the photo.",
				},
				{
					Name:        "migrate_to_chat_id",
					TypeExpr:    ir.TypeExpr{Types: []ir.TypeRef{{Type: "Integer64"}}},
					Optional:    true,
					Description: "Optional. The group has been migrated to a supergroup.",
				},
			},
		},
		// Type for testing file_id expr rule
		{
			Name:        "ChatPhoto",
			Description: "This object represents a chat photo.",
			Fields: []ir.Field{
				{
					Name:        "small_file_id",
					TypeExpr:    ir.TypeExpr{Types: []ir.TypeRef{{Type: "String"}}},
					Description: "File identifier of small photo.",
				},
				{
					Name:        "big_file_id",
					TypeExpr:    ir.TypeExpr{Types: []ir.TypeRef{{Type: "String"}}},
					Description: "File identifier of big photo.",
				},
			},
		},
		// Type for testing MPEG4 naming and plural IDs
		{
			Name:        "InlineQueryResultMpeg4Gif",
			Description: "Represents a link to a video animation.",
			Fields: []ir.Field{
				{
					Name:        "mpeg4_url",
					TypeExpr:    ir.TypeExpr{Types: []ir.TypeRef{{Type: "String"}}},
					Description: "A valid URL for the MPEG4 file.",
				},
				{
					Name:        "mpeg4_width",
					TypeExpr:    ir.TypeExpr{Types: []ir.TypeRef{{Type: "Integer"}}},
					Optional:    true,
					Description: "Optional. Video width.",
				},
				{
					Name:        "text_parse_mode",
					TypeExpr:    ir.TypeExpr{Types: []ir.TypeRef{{Type: "String"}}},
					Optional:    true,
					Description: "Optional. Mode for parsing entities in the text.",
				},
			},
		},
		{
			Name:        "PollAnswer",
			Description: "This object represents an answer of a user in a non-anonymous poll.",
			Fields: []ir.Field{
				{
					Name:        "user_id",
					TypeExpr:    ir.TypeExpr{Types: []ir.TypeRef{{Type: "Integer64"}}},
					Description: "Unique identifier of the user that answered.",
				},
				{
					Name:        "option_ids",
					TypeExpr:    ir.TypeExpr{Types: []ir.TypeRef{{Type: "Integer"}}, Array: 1},
					Description: "0-based identifiers of chosen answer options.",
				},
			},
		},
		// Type for testing Username suffix rule
		{
			Name:        "LoginURL",
			Description: "This object represents a parameter of the inline keyboard button.",
			Fields: []ir.Field{
				{
					Name:        "url",
					TypeExpr:    ir.TypeExpr{Types: []ir.TypeRef{{Type: "String"}}},
					Description: "An HTTPS URL.",
				},
				{
					Name:        "bot_username",
					TypeExpr:    ir.TypeExpr{Types: []ir.TypeRef{{Type: "String"}}},
					Optional:    true,
					Description: "Optional. Username of a bot.",
				},
			},
		},
		// Type for testing FileArg rule (file_id + attach://)
		{
			Name:        "InputMediaPhoto",
			Description: "Represents a photo to be sent.",
			Fields: []ir.Field{
				{
					Name:        "media",
					TypeExpr:    ir.TypeExpr{Types: []ir.TypeRef{{Type: "String"}}},
					Description: "File to send. Pass a file_id to send a file that exists on the Telegram servers, pass an HTTP URL, or pass \"attach://<file_attach_name>\".",
				},
			},
		},
		// Type for testing FileArg and InputFile rules together
		{
			Name:        "InputPaidMediaVideo",
			Description: "The paid media to send is a video.",
			Fields: []ir.Field{
				{
					Name:        "media",
					TypeExpr:    ir.TypeExpr{Types: []ir.TypeRef{{Type: "String"}}},
					Description: "File to send. Pass a file_id to send a file that exists on the Telegram servers, pass an HTTP URL, or pass \"attach://<file_attach_name>\".",
				},
				{
					Name:        "thumbnail",
					TypeExpr:    ir.TypeExpr{Types: []ir.TypeRef{{Type: "String"}}},
					Optional:    true,
					Description: "Optional. Thumbnail of the file sent; the thumbnail should be in JPEG format. The file must be uploaded using multipart/form-data under the name specified in \"attach://<file_attach_name>\".",
				},
				{
					Name:        "cover",
					TypeExpr:    ir.TypeExpr{Types: []ir.TypeRef{{Type: "String"}}},
					Optional:    true,
					Description: "Optional. Cover for the video. Pass a file_id to send a file that exists on the Telegram servers, pass an HTTP URL, or pass \"attach://<file_attach_name>\".",
				},
			},
		},
		// Type for testing InputFile rule (required, upload-only)
		{
			Name:        "InputProfilePhotoStatic",
			Description: "A static profile photo in the .JPG format.",
			Fields: []ir.Field{
				{
					Name:        "photo",
					TypeExpr:    ir.TypeExpr{Types: []ir.TypeRef{{Type: "String"}}},
					Description: "The static profile photo. The photo must be uploaded using multipart/form-data under the name specified in \"attach://<file_attach_name>\".",
				},
			},
		},
		// Union type: BackgroundFill (not excluded)
		{
			Name:        "BackgroundFill",
			Description: "This object describes the background fill.",
			Subtypes:    []string{"BackgroundFillSolid", "BackgroundFillGradient"},
		},
		{
			Name:        "BackgroundFillSolid",
			Description: "The background is filled with a solid color.",
			Fields: []ir.Field{
				{
					Name:        "type",
					TypeExpr:    ir.TypeExpr{Types: []ir.TypeRef{{Type: "String"}}},
					Description: "Type of the background fill, always \"solid\".",
					Const:       "solid",
				},
				{
					Name:        "color",
					TypeExpr:    ir.TypeExpr{Types: []ir.TypeRef{{Type: "Integer"}}},
					Description: "The color of the background fill.",
				},
			},
		},
		{
			Name:        "BackgroundFillGradient",
			Description: "The background is a gradient fill.",
			Fields: []ir.Field{
				{
					Name:        "type",
					TypeExpr:    ir.TypeExpr{Types: []ir.TypeRef{{Type: "String"}}},
					Description: "Type of the background fill, always \"gradient\".",
					Const:       "gradient",
				},
				{
					Name:        "top_color",
					TypeExpr:    ir.TypeExpr{Types: []ir.TypeRef{{Type: "Integer"}}},
					Description: "Top color of the gradient.",
				},
			},
		},
		// Excluded union type: MessageOrigin
		{
			Name:        "MessageOrigin",
			Description: "This object describes the origin of a message.",
			Subtypes:    []string{"MessageOriginUser", "MessageOriginChat"},
		},
		// Union without discriminator (excluded)
		{
			Name:        "MaybeInaccessibleMessage",
			Description: "This object describes a message that can be inaccessible.",
			Subtypes:    []string{"Message", "InaccessibleMessage"},
		},
	},
}

func TestGenerate(t *testing.T) {
	cfg := loadTestConfig(t)

	var buf bytes.Buffer
	err := Generate(testAPI, &buf, cfg, testLog, Options{})
	require.NoError(t, err)

	got := buf.String()

	if *updateGolden {
		err = os.WriteFile("testdata/types_gen.golden.go", buf.Bytes(), 0o644)
		require.NoError(t, err)
		t.Log("golden file updated")
		return
	}

	golden, err := os.ReadFile("testdata/types_gen.golden.go")
	require.NoError(t, err)

	assert.Equal(t, string(golden), got)
}

func TestGenerate_ExcludesTypes(t *testing.T) {
	cfg := loadTestConfig(t)

	var buf bytes.Buffer
	err := Generate(testAPI, &buf, cfg, testLog, Options{})
	require.NoError(t, err)

	got := buf.String()
	assert.NotContains(t, got, "type MessageOrigin struct")
	assert.NotContains(t, got, "type MaybeInaccessibleMessage struct")
}

func TestGenerate_UnionTypes(t *testing.T) {
	cfg := loadTestConfig(t)

	var buf bytes.Buffer
	err := Generate(testAPI, &buf, cfg, testLog, Options{})
	require.NoError(t, err)

	got := buf.String()
	assert.Contains(t, got, "type BackgroundFill struct")
	assert.Contains(t, got, "Solid *BackgroundFillSolid")
	assert.Contains(t, got, "Gradient *BackgroundFillGradient")
	assert.Contains(t, got, "func (u *BackgroundFill) UnmarshalJSON")
	assert.Contains(t, got, `case "solid":`)
	assert.Contains(t, got, `case "gradient":`)
}

func TestGenerate_UnixTimeFields(t *testing.T) {
	cfg := loadTestConfig(t)

	var buf bytes.Buffer
	err := Generate(testAPI, &buf, cfg, testLog, Options{})
	require.NoError(t, err)

	got := buf.String()
	assert.Contains(t, got, "LastErrorDate UnixTime")
	assert.Contains(t, got, "Date UnixTime")
	// No DateTime helper methods should be generated
	assert.NotContains(t, got, "DateTime()")
}

func TestGenerate_NameOverrides(t *testing.T) {
	cfg := loadTestConfig(t)

	var buf bytes.Buffer
	err := Generate(testAPI, &buf, cfg, testLog, Options{})
	require.NoError(t, err)

	got := buf.String()
	assert.Contains(t, got, `ID int `+"`"+`json:"update_id"`+"`")
	assert.Contains(t, got, `ID int `+"`"+`json:"message_id"`+"`")
}

func TestGenerate_TypeOverrides(t *testing.T) {
	cfg := loadTestConfig(t)

	var buf bytes.Buffer
	err := Generate(testAPI, &buf, cfg, testLog, Options{})
	require.NoError(t, err)

	got := buf.String()
	assert.Contains(t, got, "ID UserID")
	assert.Contains(t, got, "Username Username")
	assert.Contains(t, got, "AllowedUpdates []UpdateType")
}

func TestGenerate_FieldTypeRules(t *testing.T) {
	cfg := loadTestConfig(t)

	var buf bytes.Buffer
	err := Generate(testAPI, &buf, cfg, testLog, Options{})
	require.NoError(t, err)

	got := buf.String()
	// file_id suffix rule: small_file_id and big_file_id get FileID type
	assert.Contains(t, got, "SmallFileID FileID")
	assert.Contains(t, got, "BigFileID FileID")
	// chat_id suffix rule: migrate_to_chat_id gets ChatID type
	assert.Contains(t, got, "MigrateToChatID ChatID")
	// user_id rule: user_id gets UserID type
	assert.Contains(t, got, "UserID UserID")
	// parse_mode suffix rule: text_parse_mode gets ParseMode type
	assert.Contains(t, got, "TextParseMode ParseMode")
	// username rule: username gets Username type (scalar, no pointer even if optional)
	assert.Contains(t, got, "Username Username")
	// username suffix rule: bot_username gets Username type (scalar)
	assert.Contains(t, got, "BotUsername Username")
	// FileArg rule: media with file_id + attach:// gets FileArg (required, no pointer)
	assert.Contains(t, got, "Media FileArg")
	// InputFile rule: thumbnail without file_id but with attach:// gets *InputFile (optional)
	assert.Contains(t, got, "Thumbnail *InputFile")
	// FileArg rule: cover with file_id + attach:// gets *FileArg (optional, pointer)
	assert.Contains(t, got, "Cover *FileArg")
	// InputFile rule: required upload-only field gets InputFile (no pointer)
	assert.Contains(t, got, "Photo InputFile")
}

func TestGenerate_NamingConventions(t *testing.T) {
	cfg := loadTestConfig(t)

	var buf bytes.Buffer
	err := Generate(testAPI, &buf, cfg, testLog, Options{})
	require.NoError(t, err)

	got := buf.String()
	// MPEG4 initialism normalization
	assert.Contains(t, got, "type InlineQueryResultMPEG4GIF struct")
	assert.Contains(t, got, "MPEG4URL string")
	assert.Contains(t, got, "MPEG4Width int")
	// Plural initialism: ids → IDs
	assert.Contains(t, got, "OptionIDs []int")
}

func TestGenerate_PackageOption(t *testing.T) {
	cfg := loadTestConfig(t)

	var buf bytes.Buffer
	err := Generate(testAPI, &buf, cfg, testLog, Options{Package: "mypackage"})
	require.NoError(t, err)

	got := buf.String()
	assert.Contains(t, got, "package mypackage")
	assert.NotContains(t, got, "package tg")
}

func TestGenerate_FullAPI(t *testing.T) {
	cfg := loadTestConfig(t)

	f, err := os.Open("../parser/testdata/index.html")
	require.NoError(t, err)
	defer f.Close()

	api, err := parser.Parse(f)
	require.NoError(t, err)

	var buf bytes.Buffer
	err = Generate(api, &buf, cfg, testLog, Options{})
	require.NoError(t, err)

	got := buf.String()

	// Struct types are generated
	assert.Contains(t, got, "type Update struct")
	assert.Contains(t, got, "type Message struct")
	assert.Contains(t, got, "type User struct")
	assert.Contains(t, got, "type Chat struct")

	// Union types are generated with UnmarshalJSON
	assert.Contains(t, got, "type ChatMember struct")
	assert.Contains(t, got, "func (u *ChatMember) UnmarshalJSON")
	assert.Contains(t, got, "type BackgroundFill struct")
	assert.Contains(t, got, "func (u *BackgroundFill) UnmarshalJSON")

	// Excluded union types are absent
	assert.NotContains(t, got, "type MessageOrigin struct")
	assert.NotContains(t, got, "type ReactionType struct")

	// Excluded types are absent
	assert.NotContains(t, got, "type MaybeInaccessibleMessage struct")
	assert.NotContains(t, got, "type ForumTopicClosed struct")

	// Field type rules (expr-based): file_id suffix → FileID
	assert.Contains(t, got, "SmallFileID FileID")
	assert.Contains(t, got, "BigFileID FileID")

	// Field type rules (expr-based): chat_id suffix → ChatID
	assert.Contains(t, got, "MigrateToChatID ChatID")
	assert.Contains(t, got, "LinkedChatID ChatID")

	// Type overrides: ChatFullInfo.id → ChatID
	assert.Contains(t, got, "type ChatFullInfo struct")

	// Naming: MPEG4 initialism
	assert.Contains(t, got, "type InlineQueryResultMPEG4GIF struct")
	assert.Contains(t, got, "MPEG4URL string")

	// Naming: plural IDs
	assert.Contains(t, got, "OptionIDs []int")

	// Naming: VCard override
	assert.Contains(t, got, "VCard string")

	// Field type rules: user_id → UserID
	assert.Contains(t, got, "UserID UserID")

	// Field type rules: text_parse_mode → ParseMode
	assert.Contains(t, got, "TextParseMode ParseMode")

	// Field type rules: username → Username
	assert.Contains(t, got, "Username Username")

	// Field type rules: FileArg (media with file_id + attach://)
	assert.Contains(t, got, "Media FileArg")

	// Field type rules: *InputFile (optional thumbnail with attach:// only)
	assert.Contains(t, got, "Thumbnail *InputFile")
}
