package config

import (
	"os"
	"testing"

	"github.com/mr-linch/go-tg/gen/ir"
	"github.com/mr-linch/go-tg/gen/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApplyEnums_FieldMapped(t *testing.T) {
	api := &ir.API{
		Types: []ir.Type{
			{
				Name: "Chat",
				Fields: []ir.Field{
					{Name: "type", Enum: []string{"private", "group", "supergroup", "channel"}},
				},
			},
		},
	}

	cfg := &Config{
		Parser: Parser{
			Enums: []EnumDef{
				{Name: "ChatType", Fields: []string{"Chat.type"}},
			},
		},
	}

	err := cfg.ApplyEnums(api)
	require.NoError(t, err)
	require.Len(t, api.Enums, 1)

	assert.Equal(t, "ChatType", api.Enums[0].Name)
	assert.Equal(t, []string{"private", "group", "supergroup", "channel"}, api.Enums[0].Values)
	assert.Equal(t, []string{"Chat.type"}, api.Enums[0].Fields)
}

func TestApplyEnums_UpdateType(t *testing.T) {
	api := &ir.API{
		Types: []ir.Type{
			{
				Name: "Update",
				Fields: []ir.Field{
					{Name: "update_id"},
					{Name: "message", Optional: true},
					{Name: "edited_message", Optional: true},
					{Name: "channel_post", Optional: true},
				},
			},
		},
	}

	cfg := &Config{
		Parser: Parser{
			Enums: []EnumDef{
				{Name: "UpdateType", Expr: `filter(Fields("Update"), {.Optional}) | map({.Name})`},
			},
		},
	}

	err := cfg.ApplyEnums(api)
	require.NoError(t, err)
	require.Len(t, api.Enums, 1)

	assert.Equal(t, "UpdateType", api.Enums[0].Name)
	assert.Equal(t, []string{"message", "edited_message", "channel_post"}, api.Enums[0].Values)
}

func TestApplyEnums_SubtypeConsts(t *testing.T) {
	api := &ir.API{
		Types: []ir.Type{
			{
				Name:     "MyUnion",
				Subtypes: []string{"MyUnionFoo", "MyUnionBar"},
			},
			{
				Name: "MyUnionFoo",
				Fields: []ir.Field{
					{Name: "type", Const: "foo"},
				},
			},
			{
				Name: "MyUnionBar",
				Fields: []ir.Field{
					{Name: "type", Const: "bar"},
				},
			},
		},
	}

	cfg := &Config{
		Parser: Parser{
			Enums: []EnumDef{
				{Name: "MyUnionType", Expr: `SubtypeConsts("MyUnion", "type")`},
			},
		},
	}

	err := cfg.ApplyEnums(api)
	require.NoError(t, err)
	require.Len(t, api.Enums, 1)

	assert.Equal(t, "MyUnionType", api.Enums[0].Name)
	assert.Equal(t, []string{"foo", "bar"}, api.Enums[0].Values)
}

func TestApplyEnums_InvalidExpr(t *testing.T) {
	api := &ir.API{}

	cfg := &Config{
		Parser: Parser{
			Enums: []EnumDef{
				{Name: "Bad", Expr: `invalid syntax !!!`},
			},
		},
	}

	err := cfg.ApplyEnums(api)
	require.Error(t, err)
	assert.Contains(t, err.Error(), `enum "Bad"`)
}

func TestApplyEnums_ExprEmptyValues(t *testing.T) {
	api := &ir.API{}

	cfg := &Config{
		Parser: Parser{
			Enums: []EnumDef{
				{Name: "Missing", Expr: `Fields("NonExistent") | map({.Name})`},
			},
		},
	}

	err := cfg.ApplyEnums(api)
	require.Error(t, err)
	assert.Contains(t, err.Error(), `enum "Missing"`)
	assert.Contains(t, err.Error(), "produced no values")
}

func TestApplyEnums_FieldNotFound(t *testing.T) {
	api := &ir.API{
		Types: []ir.Type{
			{Name: "Chat", Fields: []ir.Field{{Name: "id"}}},
		},
	}

	cfg := &Config{
		Parser: Parser{
			Enums: []EnumDef{
				{Name: "Bad", Fields: []string{"Chat.nonexistent"}},
			},
		},
	}

	err := cfg.ApplyEnums(api)
	require.Error(t, err)
	assert.Contains(t, err.Error(), `enum "Bad"`)
	assert.Contains(t, err.Error(), "not found")
}

func TestApplyEnums_TypeNotFound(t *testing.T) {
	api := &ir.API{}

	cfg := &Config{
		Parser: Parser{
			Enums: []EnumDef{
				{Name: "Bad", Fields: []string{"NonExistent.type"}},
			},
		},
	}

	err := cfg.ApplyEnums(api)
	require.Error(t, err)
	assert.Contains(t, err.Error(), `enum "Bad"`)
	assert.Contains(t, err.Error(), "not found")
}

func TestApplyEnums_FullDoc(t *testing.T) {
	f, err := os.Open("../parser/testdata/index.html")
	require.NoError(t, err)
	defer func() { require.NoError(t, f.Close()) }()

	api, err := parser.Parse(f)
	require.NoError(t, err)

	cfg, err := LoadFile("../config.yaml")
	require.NoError(t, err)

	err = cfg.ApplyEnums(api)
	require.NoError(t, err)

	enumByName := func(name string) *ir.Enum {
		for i := range api.Enums {
			if api.Enums[i].Name == name {
				return &api.Enums[i]
			}
		}
		return nil
	}

	// Field-mapped: ChatType
	chatType := enumByName("ChatType")
	require.NotNil(t, chatType, "ChatType enum not found")
	assert.Equal(t, []string{"private", "group", "supergroup", "channel"}, chatType.Values)

	// Field-mapped: StickerType
	stickerType := enumByName("StickerType")
	require.NotNil(t, stickerType, "StickerType enum not found")
	assert.Equal(t, []string{"regular", "mask", "custom_emoji"}, stickerType.Values)

	// Expr: UpdateType
	updateType := enumByName("UpdateType")
	require.NotNil(t, updateType, "UpdateType enum not found")
	assert.Contains(t, updateType.Values, "message")
	assert.Contains(t, updateType.Values, "edited_message")
	assert.Contains(t, updateType.Values, "channel_post")
	assert.Greater(t, len(updateType.Values), 10)

	// Expr: MessageOriginType (union discriminator)
	moType := enumByName("MessageOriginType")
	require.NotNil(t, moType, "MessageOriginType enum not found")
	assert.Equal(t, []string{"user", "hidden_user", "chat", "channel"}, moType.Values)

	// Expr: ChatMemberStatus (union discriminator with "status" field)
	cmStatus := enumByName("ChatMemberStatus")
	require.NotNil(t, cmStatus, "ChatMemberStatus enum not found")
	assert.Contains(t, cmStatus.Values, "creator")
	assert.Contains(t, cmStatus.Values, "administrator")
	assert.Contains(t, cmStatus.Values, "member")
	assert.Contains(t, cmStatus.Values, "kicked")

	// Expr: ReactionTypeType
	rtType := enumByName("ReactionTypeType")
	require.NotNil(t, rtType, "ReactionTypeType enum not found")
	assert.Contains(t, rtType.Values, "emoji")
	assert.Contains(t, rtType.Values, "custom_emoji")
}

func TestLoadFile(t *testing.T) {
	cfg, err := LoadFile("../config.yaml")
	require.NoError(t, err)
	assert.Greater(t, len(cfg.Parser.Enums), 5)
	assert.Greater(t, len(cfg.TypeGen.Exclude), 10)
}

func TestTypeGen_IsExcluded(t *testing.T) {
	tg := &TypeGen{
		Exclude: []string{"ChatID", "UserID"},
	}
	assert.True(t, tg.IsExcluded("ChatID"))
	assert.True(t, tg.IsExcluded("UserID"))
	assert.False(t, tg.IsExcluded("Message"))
}
