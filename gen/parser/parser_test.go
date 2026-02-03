package parser

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/html"

	"github.com/mr-linch/go-tg/gen/ir"
)

func TestIsTypeName(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"User", true},
		{"MessageEntity", true},
		{"ChatFullInfo", true},
		{"InputMedia2", true},
		{"sendMessage", false},
		{"getUpdates", false},
		{"December 31, 2025", false},
		{"A", false},
		{"", false},
		{"Do I need a Local Bot API Server", false},
		{"Making requests when getting updates", false},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.want, isTypeName(tt.input))
		})
	}
}

func TestIsMethodName(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"sendMessage", true},
		{"getUpdates", true},
		{"deleteWebhook", true},
		{"setWebhook", true},
		{"User", false},
		{"MessageEntity", false},
		{"a", false},
		{"close", true},
		{"send", true},
		{"", false},
		{"some thing", false},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.want, isMethodName(tt.input))
		})
	}
}

func TestExtractText(t *testing.T) {
	tests := []struct {
		name string
		html string
		want string
	}{
		{
			name: "plain text",
			html: `<td>hello world</td>`,
			want: "hello world",
		},
		{
			name: "with link",
			html: `<td><a href="#user">User</a></td>`,
			want: "User",
		},
		{
			name: "with em",
			html: `<td><em>Optional</em>. Some description</td>`,
			want: "Optional. Some description",
		},
		{
			name: "nested nodes",
			html: `<td>Array of <a href="#msg">MessageEntity</a></td>`,
			want: "Array of MessageEntity",
		},
		{
			name: "br converts to space",
			html: `<p>First line.<br>Second line.</p>`,
			want: "First line. Second line.",
		},
		{
			name: "skips i elements",
			html: `<h4><a class="anchor"><i class="anchor-icon"></i></a>Update</h4>`,
			want: "Update",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := parseHTML(t, tt.html)
			got := extractText(node)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestExtractDescription(t *testing.T) {
	tests := []struct {
		name string
		html string
		want string
	}{
		{
			name: "internal link",
			html: `<p>See <a href="#sending-files">More information on Sending Files »</a></p>`,
			want: "See [More information on Sending Files »](https://core.telegram.org/bots/api#sending-files)",
		},
		{
			name: "external link",
			html: `<p>Visit <a href="https://example.com">Example</a> for details.</p>`,
			want: "Visit [Example](https://example.com) for details.",
		},
		{
			name: "plain text unchanged",
			html: `<p>No links here.</p>`,
			want: "No links here.",
		},
		{
			name: "br converts to space",
			html: `<p>First.<br>Second.</p>`,
			want: "First. Second.",
		},
		{
			name: "multiple links",
			html: `<p>Use <a href="#method1">method1</a> or <a href="#method2">method2</a>.</p>`,
			want: "Use [method1](https://core.telegram.org/bots/api#method1) or [method2](https://core.telegram.org/bots/api#method2).",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := parseHTML(t, tt.html)
			p := findFirstElement(node, "p")
			require.NotNil(t, p)
			got := extractDescription(p)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestParseTypeCell(t *testing.T) {
	tests := []struct {
		name string
		html string
		want ir.TypeExpr
	}{
		{
			name: "simple primitive",
			html: `<table><tbody><tr><td>Integer</td></tr></tbody></table>`,
			want: ir.TypeExpr{Types: []ir.TypeRef{{Type: "Integer"}}},
		},
		{
			name: "linked type",
			html: `<table><tbody><tr><td><a href="#user">User</a></td></tr></tbody></table>`,
			want: ir.TypeExpr{Types: []ir.TypeRef{{Type: "User", Ref: "user"}}},
		},
		{
			name: "array of linked type",
			html: `<table><tbody><tr><td>Array of <a href="#messageentity">MessageEntity</a></td></tr></tbody></table>`,
			want: ir.TypeExpr{Types: []ir.TypeRef{{Type: "MessageEntity", Ref: "messageentity"}}, Array: 1},
		},
		{
			name: "nested array",
			html: `<table><tbody><tr><td>Array of Array of <a href="#photosize">PhotoSize</a></td></tr></tbody></table>`,
			want: ir.TypeExpr{Types: []ir.TypeRef{{Type: "PhotoSize", Ref: "photosize"}}, Array: 2},
		},
		{
			name: "primitive union",
			html: `<table><tbody><tr><td>Integer or String</td></tr></tbody></table>`,
			want: ir.TypeExpr{Types: []ir.TypeRef{{Type: "Integer"}, {Type: "String"}}},
		},
		{
			name: "linked with primitive union",
			html: `<table><tbody><tr><td><a href="#inputfile">InputFile</a> or String</td></tr></tbody></table>`,
			want: ir.TypeExpr{Types: []ir.TypeRef{{Type: "InputFile", Ref: "inputfile"}, {Type: "String"}}},
		},
		{
			name: "float number",
			html: `<table><tbody><tr><td>Float number</td></tr></tbody></table>`,
			want: ir.TypeExpr{Types: []ir.TypeRef{{Type: "Float"}}},
		},
		{
			name: "array of string",
			html: `<table><tbody><tr><td>Array of String</td></tr></tbody></table>`,
			want: ir.TypeExpr{Types: []ir.TypeRef{{Type: "String"}}, Array: 1},
		},
		{
			name: "array of multi-link union",
			html: `<table><tbody><tr><td>Array of <a href="#inputmediaaudio">InputMediaAudio</a>, <a href="#inputmediadocument">InputMediaDocument</a>, <a href="#inputmediaphoto">InputMediaPhoto</a> and <a href="#inputmediavideo">InputMediaVideo</a></td></tr></tbody></table>`,
			want: ir.TypeExpr{
				Types: []ir.TypeRef{
					{Type: "InputMediaAudio", Ref: "inputmediaaudio"},
					{Type: "InputMediaDocument", Ref: "inputmediadocument"},
					{Type: "InputMediaPhoto", Ref: "inputmediaphoto"},
					{Type: "InputMediaVideo", Ref: "inputmediavideo"},
				},
				Array: 1,
			},
		},
		{
			name: "boolean",
			html: `<table><tbody><tr><td>Boolean</td></tr></tbody></table>`,
			want: ir.TypeExpr{Types: []ir.TypeRef{{Type: "Boolean"}}},
		},
		{
			name: "true",
			html: `<table><tbody><tr><td>True</td></tr></tbody></table>`,
			want: ir.TypeExpr{Types: []ir.TypeRef{{Type: "True"}}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := parseHTML(t, tt.html)
			td := findFirstElement(node, "td")
			require.NotNil(t, td, "could not find <td> in test HTML")
			got := parseTypeCell(td)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestParseReturnType(t *testing.T) {
	tests := []struct {
		name string
		html string
		want ir.TypeExpr
	}{
		{
			name: "returns True on success",
			html: `<p>Use this method to do something. Returns <em>True</em> on success.</p>`,
			want: ir.TypeExpr{Types: []ir.TypeRef{{Type: "True"}}},
		},
		{
			name: "True is returned",
			html: `<p>Does something. <em>True</em> is returned on success.</p>`,
			want: ir.TypeExpr{Types: []ir.TypeRef{{Type: "True"}}},
		},
		{
			name: "returns linked type",
			html: `<p>On success, the sent <a href="#message">Message</a> is returned.</p>`,
			want: ir.TypeExpr{Types: []ir.TypeRef{{Type: "Message", Ref: "message"}}},
		},
		{
			name: "returns array of type",
			html: `<p>Returns an Array of <a href="#update">Update</a> objects.</p>`,
			want: ir.TypeExpr{Types: []ir.TypeRef{{Type: "Update", Ref: "update"}}, Array: 1},
		},
		{
			name: "last link wins",
			html: `<p>Contains a <a href="#update">Update</a>. Returns a <a href="#webhookinfo">WebhookInfo</a> object.</p>`,
			want: ir.TypeExpr{Types: []ir.TypeRef{{Type: "WebhookInfo", Ref: "webhookinfo"}}},
		},
		{
			name: "no type found",
			html: `<p>Some description without types.</p>`,
			want: ir.TypeExpr{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := parseHTML(t, tt.html)
			var nodes []*html.Node
			collectTestNodes(node, &nodes)
			got := parseReturnType(nodes)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestIsSubtypesList(t *testing.T) {
	tests := []struct {
		name  string
		html  string
		want  []string
		valid bool
	}{
		{
			name:  "valid subtypes list",
			html:  `<ul><li><a href="#messageoriginuser">MessageOriginUser</a></li><li><a href="#messageoriginchat">MessageOriginChat</a></li></ul>`,
			want:  []string{"MessageOriginUser", "MessageOriginChat"},
			valid: true,
		},
		{
			name:  "not subtypes - has text content",
			html:  `<ul><li>Some text and <a href="#thing">Thing</a></li></ul>`,
			want:  nil,
			valid: false,
		},
		{
			name:  "not subtypes - no links",
			html:  `<ul><li>Just text</li></ul>`,
			want:  nil,
			valid: false,
		},
		{
			name:  "not subtypes - external links",
			html:  `<ul><li><a href="https://example.com">Example</a></li></ul>`,
			want:  nil,
			valid: false,
		},
		{
			name:  "not subtypes - non-type name",
			html:  `<ul><li><a href="#something">something</a></li></ul>`,
			want:  nil,
			valid: false,
		},
		{
			name:  "empty ul",
			html:  `<ul></ul>`,
			want:  nil,
			valid: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := parseHTML(t, tt.html)
			ul := findFirstElement(node, "ul")
			require.NotNil(t, ul, "could not find <ul> in test HTML")
			got, ok := isSubtypesList(ul)
			assert.Equal(t, tt.valid, ok)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestIsInt64Description(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"This number may have more than 32 significant bits and some programming languages may have difficulty/silent defects in interpreting it. But it has at most 52 significant bits, so a 64-bit integer or double-precision float type are safe for storing this identifier.", true},
		{"at most 52 significant bits", true},
		{"Just a regular description", false},
		{"", false},
	}
	for _, tt := range tests {
		t.Run(tt.input[:min(len(tt.input), 30)], func(t *testing.T) {
			assert.Equal(t, tt.want, isInt64Description(tt.input))
		})
	}
}

func TestParse_SimpleType(t *testing.T) {
	doc := wrapDoc(`
<h4><a class="anchor" name="photo" href="#photo"><i class="anchor-icon"></i></a>Photo</h4>
<p>This object represents a photo.</p>
<table class="table">
<thead><tr><th>Field</th><th>Type</th><th>Description</th></tr></thead>
<tbody>
<tr><td>file_id</td><td>String</td><td>Identifier for this file</td></tr>
<tr><td>width</td><td>Integer</td><td>Photo width</td></tr>
<tr><td>caption</td><td>String</td><td><em>Optional</em>. Caption for the photo</td></tr>
<tr><td>sender</td><td><a href="#user">User</a></td><td><em>Optional</em>. Sender of the photo</td></tr>
</tbody>
</table>
`)
	api, err := Parse(strings.NewReader(doc))
	require.NoError(t, err)
	require.Len(t, api.Types, 1)

	typ := api.Types[0]
	assert.Equal(t, "Photo", typ.Name)
	assert.Equal(t, "This object represents a photo.", typ.Description)
	require.Len(t, typ.Fields, 4)

	assert.Equal(t, "file_id", typ.Fields[0].Name)
	assert.Equal(t, ir.TypeExpr{Types: []ir.TypeRef{{Type: "String"}}}, typ.Fields[0].TypeExpr)
	assert.False(t, typ.Fields[0].Optional)

	assert.Equal(t, "width", typ.Fields[1].Name)
	assert.Equal(t, ir.TypeExpr{Types: []ir.TypeRef{{Type: "Integer"}}}, typ.Fields[1].TypeExpr)
	assert.False(t, typ.Fields[1].Optional)

	assert.Equal(t, "caption", typ.Fields[2].Name)
	assert.True(t, typ.Fields[2].Optional)

	assert.Equal(t, "sender", typ.Fields[3].Name)
	assert.Equal(t, ir.TypeExpr{Types: []ir.TypeRef{{Type: "User", Ref: "user"}}}, typ.Fields[3].TypeExpr)
	assert.True(t, typ.Fields[3].Optional)
}

func TestParse_UnionType(t *testing.T) {
	doc := wrapDoc(`
<h4><a class="anchor" name="messageorigin" href="#messageorigin"><i class="anchor-icon"></i></a>MessageOrigin</h4>
<p>This object describes the origin of a message. It can be one of</p>
<ul>
<li><a href="#messageoriginuser">MessageOriginUser</a></li>
<li><a href="#messageoriginhiddenuser">MessageOriginHiddenUser</a></li>
<li><a href="#messageoriginchat">MessageOriginChat</a></li>
</ul>
`)
	api, err := Parse(strings.NewReader(doc))
	require.NoError(t, err)
	require.Len(t, api.Types, 1)

	typ := api.Types[0]
	assert.Equal(t, "MessageOrigin", typ.Name)
	assert.Empty(t, typ.Fields)
	assert.Equal(t, []string{"MessageOriginUser", "MessageOriginHiddenUser", "MessageOriginChat"}, typ.Subtypes)
}

func TestParse_TypeNoTable(t *testing.T) {
	doc := wrapDoc(`
<h4><a class="anchor" name="inputfile" href="#inputfile"><i class="anchor-icon"></i></a>InputFile</h4>
<p>This object represents the contents of a file to be uploaded.</p>
`)
	api, err := Parse(strings.NewReader(doc))
	require.NoError(t, err)
	require.Len(t, api.Types, 1)

	typ := api.Types[0]
	assert.Equal(t, "InputFile", typ.Name)
	assert.Equal(t, "This object represents the contents of a file to be uploaded.", typ.Description)
	assert.Empty(t, typ.Fields)
	assert.Empty(t, typ.Subtypes)
}

func TestParse_Integer64Field(t *testing.T) {
	doc := wrapDoc(`
<h4><a class="anchor" name="user" href="#user"><i class="anchor-icon"></i></a>User</h4>
<p>Represents a user.</p>
<table class="table">
<thead><tr><th>Field</th><th>Type</th><th>Description</th></tr></thead>
<tbody>
<tr><td>id</td><td>Integer</td><td>Unique identifier. This number may have more than 32 significant bits and some programming languages may have difficulty/silent defects in interpreting it. But it has at most 52 significant bits, so a 64-bit integer or double-precision float type are safe for storing this identifier.</td></tr>
</tbody>
</table>
`)
	api, err := Parse(strings.NewReader(doc))
	require.NoError(t, err)
	require.Len(t, api.Types, 1)
	require.Len(t, api.Types[0].Fields, 1)

	assert.Equal(t, ir.TypeExpr{Types: []ir.TypeRef{{Type: "Integer64"}}}, api.Types[0].Fields[0].TypeExpr)
}

func TestParse_Method(t *testing.T) {
	doc := wrapDoc(`
<h4><a class="anchor" name="sendmessage" href="#sendmessage"><i class="anchor-icon"></i></a>sendMessage</h4>
<p>Use this method to send text messages. On success, the sent <a href="#message">Message</a> is returned.</p>
<table class="table">
<thead><tr><th>Parameter</th><th>Type</th><th>Required</th><th>Description</th></tr></thead>
<tbody>
<tr><td>chat_id</td><td>Integer or String</td><td>Yes</td><td>Unique identifier for the target chat</td></tr>
<tr><td>text</td><td>String</td><td>Yes</td><td>Text of the message</td></tr>
<tr><td>parse_mode</td><td>String</td><td>Optional</td><td>Mode for parsing entities</td></tr>
</tbody>
</table>
`)
	api, err := Parse(strings.NewReader(doc))
	require.NoError(t, err)
	require.Len(t, api.Methods, 1)

	m := api.Methods[0]
	assert.Equal(t, "sendMessage", m.Name)
	assert.Equal(t, ir.TypeExpr{Types: []ir.TypeRef{{Type: "Message", Ref: "message"}}}, m.Returns)
	require.Len(t, m.Params, 3)

	assert.Equal(t, "chat_id", m.Params[0].Name)
	assert.Equal(t, ir.TypeExpr{Types: []ir.TypeRef{{Type: "Integer"}, {Type: "String"}}}, m.Params[0].TypeExpr)
	assert.True(t, m.Params[0].Required)

	assert.Equal(t, "text", m.Params[1].Name)
	assert.True(t, m.Params[1].Required)

	assert.Equal(t, "parse_mode", m.Params[2].Name)
	assert.False(t, m.Params[2].Required)
}

func TestParse_MethodNoParams(t *testing.T) {
	doc := wrapDoc(`
<h4><a class="anchor" name="getwebhookinfo" href="#getwebhookinfo"><i class="anchor-icon"></i></a>getWebhookInfo</h4>
<p>Use this method to get current webhook status. Requires no parameters. On success, returns a <a href="#webhookinfo">WebhookInfo</a> object.</p>
`)
	api, err := Parse(strings.NewReader(doc))
	require.NoError(t, err)
	require.Len(t, api.Methods, 1)

	m := api.Methods[0]
	assert.Equal(t, "getWebhookInfo", m.Name)
	assert.Empty(t, m.Params)
	assert.Equal(t, ir.TypeExpr{Types: []ir.TypeRef{{Type: "WebhookInfo", Ref: "webhookinfo"}}}, m.Returns)
}

func TestParse_MethodReturnsBool(t *testing.T) {
	doc := wrapDoc(`
<h4><a class="anchor" name="deletewebhook" href="#deletewebhook"><i class="anchor-icon"></i></a>deleteWebhook</h4>
<p>Use this method to remove webhook. Returns <em>True</em> on success.</p>
<table class="table">
<thead><tr><th>Parameter</th><th>Type</th><th>Required</th><th>Description</th></tr></thead>
<tbody>
<tr><td>drop_pending_updates</td><td>Boolean</td><td>Optional</td><td>Pass True to drop all pending updates</td></tr>
</tbody>
</table>
`)
	api, err := Parse(strings.NewReader(doc))
	require.NoError(t, err)
	require.Len(t, api.Methods, 1)

	m := api.Methods[0]
	assert.Equal(t, "deleteWebhook", m.Name)
	assert.Equal(t, ir.TypeExpr{Types: []ir.TypeRef{{Type: "True"}}}, m.Returns)
}

func TestParse_MethodReturnsArray(t *testing.T) {
	doc := wrapDoc(`
<h4><a class="anchor" name="getupdates" href="#getupdates"><i class="anchor-icon"></i></a>getUpdates</h4>
<p>Use this method to receive incoming updates. Returns an Array of <a href="#update">Update</a> objects.</p>
<table class="table">
<thead><tr><th>Parameter</th><th>Type</th><th>Required</th><th>Description</th></tr></thead>
<tbody>
<tr><td>offset</td><td>Integer</td><td>Optional</td><td>Identifier of the first update</td></tr>
</tbody>
</table>
`)
	api, err := Parse(strings.NewReader(doc))
	require.NoError(t, err)
	require.Len(t, api.Methods, 1)

	m := api.Methods[0]
	assert.Equal(t, "getUpdates", m.Name)
	assert.Equal(t, ir.TypeExpr{Types: []ir.TypeRef{{Type: "Update", Ref: "update"}}, Array: 1}, m.Returns)
}

func TestParse_MethodReturnsLastLink(t *testing.T) {
	doc := wrapDoc(`
<h4><a class="anchor" name="getWebhookInfo" href="#getWebhookInfo"><i class="anchor-icon"></i></a>getWebhookInfo</h4>
<p>Use this method to get current webhook status. Contains a <a href="#update">Update</a>. Returns a <a href="#webhookinfo">WebhookInfo</a> object.</p>
`)
	api, err := Parse(strings.NewReader(doc))
	require.NoError(t, err)
	require.Len(t, api.Methods, 1)

	m := api.Methods[0]
	assert.Equal(t, ir.TypeExpr{Types: []ir.TypeRef{{Type: "WebhookInfo", Ref: "webhookinfo"}}}, m.Returns)
}

func TestParse_SkipsNonAPI(t *testing.T) {
	doc := wrapDoc(`
<h4><a class="anchor" name="december-31-2025" href="#december-31-2025"><i class="anchor-icon"></i></a>December 31, 2025</h4>
<p>Bot API 9.3</p>
<h4><a class="anchor" name="do-i-need" href="#do-i-need"><i class="anchor-icon"></i></a>Do I need a Local Bot API Server</h4>
<p>Some FAQ text.</p>
<h4><a class="anchor" name="making-requests" href="#making-requests"><i class="anchor-icon"></i></a>Making requests when getting updates</h4>
<p>Some explanation.</p>
<h4><a class="anchor" name="user" href="#user"><i class="anchor-icon"></i></a>User</h4>
<p>This object represents a user.</p>
<table class="table">
<thead><tr><th>Field</th><th>Type</th><th>Description</th></tr></thead>
<tbody>
<tr><td>id</td><td>Integer</td><td>User id</td></tr>
</tbody>
</table>
`)
	api, err := Parse(strings.NewReader(doc))
	require.NoError(t, err)
	assert.Len(t, api.Types, 1)
	assert.Equal(t, "User", api.Types[0].Name)
	assert.Empty(t, api.Methods)
}

func TestParse_NestedArray(t *testing.T) {
	doc := wrapDoc(`
<h4><a class="anchor" name="replykeyboardmarkup" href="#replykeyboardmarkup"><i class="anchor-icon"></i></a>ReplyKeyboardMarkup</h4>
<p>This object represents a custom keyboard.</p>
<table class="table">
<thead><tr><th>Field</th><th>Type</th><th>Description</th></tr></thead>
<tbody>
<tr><td>keyboard</td><td>Array of Array of <a href="#keyboardbutton">KeyboardButton</a></td><td>Array of button rows</td></tr>
</tbody>
</table>
`)
	api, err := Parse(strings.NewReader(doc))
	require.NoError(t, err)
	require.Len(t, api.Types, 1)
	require.Len(t, api.Types[0].Fields, 1)

	f := api.Types[0].Fields[0]
	assert.Equal(t, "keyboard", f.Name)
	assert.Equal(t, 2, f.TypeExpr.Array)
	assert.Equal(t, []ir.TypeRef{{Type: "KeyboardButton", Ref: "keyboardbutton"}}, f.TypeExpr.Types)
}

func TestParse_UnionTypeInField(t *testing.T) {
	doc := wrapDoc(`
<h4><a class="anchor" name="mytype" href="#mytype"><i class="anchor-icon"></i></a>MyType</h4>
<p>A type with union field.</p>
<table class="table">
<thead><tr><th>Field</th><th>Type</th><th>Description</th></tr></thead>
<tbody>
<tr><td>reply_markup</td><td><a href="#inlinekeyboardmarkup">InlineKeyboardMarkup</a> or <a href="#replykeyboardmarkup">ReplyKeyboardMarkup</a></td><td><em>Optional</em>. Additional interface options</td></tr>
</tbody>
</table>
`)
	api, err := Parse(strings.NewReader(doc))
	require.NoError(t, err)
	require.Len(t, api.Types, 1)
	require.Len(t, api.Types[0].Fields, 1)

	f := api.Types[0].Fields[0]
	assert.Equal(t, "reply_markup", f.Name)
	assert.True(t, f.Optional)
	assert.Equal(t, ir.TypeExpr{Types: []ir.TypeRef{
		{Type: "InlineKeyboardMarkup", Ref: "inlinekeyboardmarkup"},
		{Type: "ReplyKeyboardMarkup", Ref: "replykeyboardmarkup"},
	}}, f.TypeExpr)
}

func TestParse_FieldNameStripNEW(t *testing.T) {
	doc := wrapDoc(`
<h4><a class="anchor" name="mytype" href="#mytype"><i class="anchor-icon"></i></a>MyType</h4>
<p>A type with new field.</p>
<table class="table">
<thead><tr><th>Field</th><th>Type</th><th>Description</th></tr></thead>
<tbody>
<tr><td>new_field NEW</td><td>String</td><td>A newly added field</td></tr>
</tbody>
</table>
`)
	api, err := Parse(strings.NewReader(doc))
	require.NoError(t, err)
	require.Len(t, api.Types, 1)
	require.Len(t, api.Types[0].Fields, 1)

	assert.Equal(t, "new_field", api.Types[0].Fields[0].Name)
}

func TestParse_BrSplitsDescription(t *testing.T) {
	doc := wrapDoc(`
<h4><a class="anchor" name="answerinlinequery" href="#answerinlinequery"><i class="anchor-icon"></i></a>answerInlineQuery</h4>
<p>Use this method to send answers to an inline query. On success, <em>True</em> is returned.<br>No more than <strong>50</strong> results per query are allowed.</p>
<table class="table">
<thead><tr><th>Parameter</th><th>Type</th><th>Required</th><th>Description</th></tr></thead>
<tbody>
<tr><td>inline_query_id</td><td>String</td><td>Yes</td><td>Unique identifier for the answered query</td></tr>
</tbody>
</table>
`)
	api, err := Parse(strings.NewReader(doc))
	require.NoError(t, err)
	require.Len(t, api.Methods, 1)

	m := api.Methods[0]
	require.Len(t, m.Description, 2)
	assert.Equal(t, "Use this method to send answers to an inline query. On success, True is returned.", m.Description[0])
	assert.Equal(t, "No more than 50 results per query are allowed.", m.Description[1])
}

func TestParse_BlockquoteInMethod(t *testing.T) {
	doc := wrapDoc(`
<h4><a class="anchor" name="getUpdates" href="#getUpdates"><i class="anchor-icon"></i></a>getUpdates</h4>
<p>Use this method to receive updates. Returns an Array of <a href="#update">Update</a> objects.</p>
<blockquote><p><strong>Notes</strong><br>1. This method will not work if an outgoing webhook is set up.</p></blockquote>
<table class="table">
<thead><tr><th>Parameter</th><th>Type</th><th>Required</th><th>Description</th></tr></thead>
<tbody>
<tr><td>offset</td><td>Integer</td><td>Optional</td><td>Identifier of the first update</td></tr>
</tbody>
</table>
`)
	api, err := Parse(strings.NewReader(doc))
	require.NoError(t, err)
	require.Len(t, api.Methods, 1)

	m := api.Methods[0]
	require.Len(t, m.Description, 3)
	assert.Contains(t, m.Description[0], "Use this method")
	// <br> splits into separate description entries
	assert.Equal(t, "Notes", m.Description[1])
	assert.Equal(t, "1. This method will not work if an outgoing webhook is set up.", m.Description[2])
	// No leading/trailing whitespace
	for i, d := range m.Description {
		assert.Equal(t, strings.TrimSpace(d), d, "description[%d] has leading/trailing whitespace", i)
	}
}

func TestParse_FloatNumberField(t *testing.T) {
	doc := wrapDoc(`
<h4><a class="anchor" name="location" href="#location"><i class="anchor-icon"></i></a>Location</h4>
<p>This object represents a point on the map.</p>
<table class="table">
<thead><tr><th>Field</th><th>Type</th><th>Description</th></tr></thead>
<tbody>
<tr><td>longitude</td><td>Float number</td><td>Longitude as defined by the sender</td></tr>
<tr><td>latitude</td><td>Float number</td><td>Latitude as defined by the sender</td></tr>
</tbody>
</table>
`)
	api, err := Parse(strings.NewReader(doc))
	require.NoError(t, err)
	require.Len(t, api.Types, 1)
	require.Len(t, api.Types[0].Fields, 2)

	assert.Equal(t, ir.TypeExpr{Types: []ir.TypeRef{{Type: "Float"}}}, api.Types[0].Fields[0].TypeExpr)
	assert.Equal(t, ir.TypeExpr{Types: []ir.TypeRef{{Type: "Float"}}}, api.Types[0].Fields[1].TypeExpr)
}

func TestExtractConst(t *testing.T) {
	tests := []struct {
		name string
		html string
		want string
	}{
		{
			name: "always quoted",
			html: `<td>Type of the message origin, always "user"</td>`,
			want: "user",
		},
		{
			name: "always quoted underscore",
			html: `<td>Type of the message origin, always "hidden_user"</td>`,
			want: "hidden_user",
		},
		{
			name: "must be em",
			html: `<td>Scope type, must be <em>default</em></td>`,
			want: "default",
		},
		{
			name: "must be em underscore",
			html: `<td>Scope type, must be <em>chat_member</em></td>`,
			want: "chat_member",
		},
		{
			name: "no pattern",
			html: `<td>Unique identifier for the target chat</td>`,
			want: "",
		},
		{
			name: "quoted but not always",
			html: `<td>Type of the chat, can be either "private" or "group"</td>`,
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := parseHTML(t, `<table><tbody><tr>`+tt.html+`</tr></tbody></table>`)
			td := findFirstElement(doc, "td")
			require.NotNil(t, td)
			assert.Equal(t, tt.want, extractConst(td))
		})
	}
}

func TestExtractEnum(t *testing.T) {
	tests := []struct {
		name string
		desc string
		want []string
	}{
		{
			name: "can be either",
			desc: `Type of the chat, can be either "private", "group", "supergroup" or "channel"`,
			want: []string{"private", "group", "supergroup", "channel"},
		},
		{
			name: "currently one of",
			desc: `Type of the sticker, currently one of "regular", "mask", "custom_emoji"`,
			want: []string{"regular", "mask", "custom_emoji"},
		},
		{
			name: "must be one of",
			desc: `Format of the sticker, must be one of "static", "animated", "video"`,
			want: []string{"static", "animated", "video"},
		},
		{
			name: "One of capitalized",
			desc: `One of "forehead", "eyes", "mouth", or "chin".`,
			want: []string{"forehead", "eyes", "mouth", "chin"},
		},
		{
			name: "can be one of",
			desc: `State of the suggested post. Currently, it can be one of "pending", "approved", "declined".`,
			want: []string{"pending", "approved", "declined"},
		},
		{
			name: "no pattern",
			desc: `Unique identifier for the target chat`,
			want: nil,
		},
		{
			name: "always pattern not enum",
			desc: `Type of the message origin, always "user"`,
			want: nil,
		},
		{
			name: "single quoted value not enum",
			desc: `Must be one of "only_one"`,
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, extractEnum(tt.desc))
		})
	}
}

func TestIsTimestampDescription(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"Date the message was sent in Unix time", true},
		{"Unix time for the most recent error", true},
		{"Point in time (Unix timestamp) when poll closes", true},
		{"Date when the user will be unbanned; Unix time.", true},
		{"unix time of the event", true},
		{"Just a regular description", false},
		{"Unique identifier for the chat", false},
		{"", false},
	}
	for _, tt := range tests {
		t.Run(tt.input[:min(len(tt.input), 40)], func(t *testing.T) {
			assert.Equal(t, tt.want, isTimestampDescription(tt.input))
		})
	}
}

func TestParse_TimestampField(t *testing.T) {
	doc := wrapDoc(`
<h4><a class="anchor" name="message" href="#message"><i class="anchor-icon"></i></a>Message</h4>
<p>This object represents a message.</p>
<table class="table">
<thead><tr><th>Field</th><th>Type</th><th>Description</th></tr></thead>
<tbody>
<tr><td>date</td><td>Integer</td><td>Date the message was sent in Unix time</td></tr>
<tr><td>edit_date</td><td>Integer</td><td><em>Optional</em>. Date the message was last edited in Unix time</td></tr>
<tr><td>message_id</td><td>Integer</td><td>Unique message identifier</td></tr>
</tbody>
</table>
`)
	api, err := Parse(strings.NewReader(doc))
	require.NoError(t, err)
	require.Len(t, api.Types, 1)

	dateField := findField(t, api.Types[0], "date")
	assert.Equal(t, ir.TypeExpr{Types: []ir.TypeRef{{Type: "Integer"}}}, dateField.TypeExpr)

	editDate := findField(t, api.Types[0], "edit_date")
	assert.Equal(t, ir.TypeExpr{Types: []ir.TypeRef{{Type: "Integer"}}}, editDate.TypeExpr)

	msgID := findField(t, api.Types[0], "message_id")
	assert.Equal(t, ir.TypeExpr{Types: []ir.TypeRef{{Type: "Integer"}}}, msgID.TypeExpr)
}

func TestExtractDefault(t *testing.T) {
	tests := []struct {
		name string
		desc string
		want string
	}{
		{"simple number", "Defaults to 100", "100"},
		{"zero", "Defaults to 0", "0"},
		{"false", "Defaults to false", "false"},
		{"true", "Defaults to true", "true"},
		{"trailing dot", "Defaults to 40.", "40"},
		{"trailing comma", "Defaults to 40,", "40"},
		{"mid sentence", "Values between 1-100. Defaults to 100", "100"},
		{"lowercase defaults", "defaults to 5", "5"},
		{"no match", "No default value here", ""},
		{"empty", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, extractDefault(tt.desc))
		})
	}
}

func TestParse_ParamDefault(t *testing.T) {
	doc := wrapDoc(`
<h4><a class="anchor" name="getupdates" href="#getupdates"><i class="anchor-icon"></i></a>getUpdates</h4>
<p>Use this method to receive updates. Returns an Array of <a href="#update">Update</a> objects.</p>
<table class="table">
<thead><tr><th>Parameter</th><th>Type</th><th>Required</th><th>Description</th></tr></thead>
<tbody>
<tr><td>offset</td><td>Integer</td><td>Optional</td><td>Identifier of the first update</td></tr>
<tr><td>limit</td><td>Integer</td><td>Optional</td><td>Limits the number of updates. Values between 1-100 are accepted. Defaults to 100</td></tr>
<tr><td>timeout</td><td>Integer</td><td>Optional</td><td>Timeout in seconds for long polling. Defaults to 0, i.e. usual short polling.</td></tr>
</tbody>
</table>
`)
	api, err := Parse(strings.NewReader(doc))
	require.NoError(t, err)
	require.Len(t, api.Methods, 1)

	m := api.Methods[0]
	offset := findParam(t, m, "offset")
	assert.Empty(t, offset.Default)

	limit := findParam(t, m, "limit")
	assert.Equal(t, "100", limit.Default)

	timeout := findParam(t, m, "timeout")
	assert.Equal(t, "0", timeout.Default)
}

func TestParse_ConstField(t *testing.T) {
	doc := wrapDoc(`
<h4><a class="anchor" name="messageoriginuser" href="#messageoriginuser"><i class="anchor-icon"></i></a>MessageOriginUser</h4>
<p>The message was originally sent by a known user.</p>
<table class="table">
<thead><tr><th>Field</th><th>Type</th><th>Description</th></tr></thead>
<tbody>
<tr><td>type</td><td>String</td><td>Type of the message origin, always "user"</td></tr>
<tr><td>date</td><td>Integer</td><td>Date the message was sent</td></tr>
</tbody>
</table>
`)
	api, err := Parse(strings.NewReader(doc))
	require.NoError(t, err)
	require.Len(t, api.Types, 1)

	typeField := findField(t, api.Types[0], "type")
	assert.Equal(t, "user", typeField.Const)
	assert.Nil(t, typeField.Enum)

	dateField := findField(t, api.Types[0], "date")
	assert.Empty(t, dateField.Const)
	assert.Nil(t, dateField.Enum)
}

func TestParse_ConstFieldEm(t *testing.T) {
	doc := wrapDoc(`
<h4><a class="anchor" name="botcommandscopedefault" href="#botcommandscopedefault"><i class="anchor-icon"></i></a>BotCommandScopeDefault</h4>
<p>Represents the default scope of bot commands.</p>
<table class="table">
<thead><tr><th>Field</th><th>Type</th><th>Description</th></tr></thead>
<tbody>
<tr><td>type</td><td>String</td><td>Scope type, must be <em>default</em></td></tr>
</tbody>
</table>
`)
	api, err := Parse(strings.NewReader(doc))
	require.NoError(t, err)
	require.Len(t, api.Types, 1)

	typeField := findField(t, api.Types[0], "type")
	assert.Equal(t, "default", typeField.Const)
}

func TestParse_EnumField(t *testing.T) {
	doc := wrapDoc(`
<h4><a class="anchor" name="chat" href="#chat"><i class="anchor-icon"></i></a>Chat</h4>
<p>This object represents a chat.</p>
<table class="table">
<thead><tr><th>Field</th><th>Type</th><th>Description</th></tr></thead>
<tbody>
<tr><td>id</td><td>Integer</td><td>Unique identifier for the chat</td></tr>
<tr><td>type</td><td>String</td><td>Type of the chat, can be either "private", "group", "supergroup" or "channel"</td></tr>
</tbody>
</table>
`)
	api, err := Parse(strings.NewReader(doc))
	require.NoError(t, err)
	require.Len(t, api.Types, 1)

	typeField := findField(t, api.Types[0], "type")
	assert.Empty(t, typeField.Const)
	assert.Equal(t, []string{"private", "group", "supergroup", "channel"}, typeField.Enum)
}

func TestParse_FullDoc(t *testing.T) {
	f, err := os.Open("testdata/index.html")
	require.NoError(t, err)
	defer func() { require.NoError(t, f.Close()) }()

	api, err := Parse(f)
	require.NoError(t, err)

	// Verify reasonable counts
	assert.Greater(t, len(api.Types), 200, "expected > 200 types, got %d", len(api.Types))
	assert.Greater(t, len(api.Methods), 100, "expected > 100 methods, got %d", len(api.Methods))

	// Spot-check: Update type
	update := findType(t, api, "Update")
	updateID := findField(t, update, "update_id")
	assert.Equal(t, ir.TypeExpr{Types: []ir.TypeRef{{Type: "Integer"}}}, updateID.TypeExpr)
	assert.False(t, updateID.Optional)
	msg := findField(t, update, "message")
	assert.True(t, msg.Optional)
	assert.Equal(t, "message", msg.TypeExpr.Types[0].Ref)

	// Spot-check: User type with Integer64
	user := findType(t, api, "User")
	userID := findField(t, user, "id")
	assert.Equal(t, ir.TypeExpr{Types: []ir.TypeRef{{Type: "Integer64"}}}, userID.TypeExpr)
	lastName := findField(t, user, "last_name")
	assert.True(t, lastName.Optional)

	// Spot-check: Message type with array field
	message := findType(t, api, "Message")
	entities := findField(t, message, "entities")
	assert.Equal(t, 1, entities.TypeExpr.Array)
	assert.Equal(t, "MessageEntity", entities.TypeExpr.Types[0].Type)
	assert.Equal(t, "messageentity", entities.TypeExpr.Types[0].Ref)

	// Spot-check: ReplyKeyboardMarkup with nested array
	rkm := findType(t, api, "ReplyKeyboardMarkup")
	keyboard := findField(t, rkm, "keyboard")
	assert.Equal(t, 2, keyboard.TypeExpr.Array)
	assert.Equal(t, "KeyboardButton", keyboard.TypeExpr.Types[0].Type)

	// Spot-check: MessageOrigin is union type
	mo := findType(t, api, "MessageOrigin")
	assert.Empty(t, mo.Fields)
	assert.Contains(t, mo.Subtypes, "MessageOriginUser")
	assert.Contains(t, mo.Subtypes, "MessageOriginChannel")

	// Spot-check: InputFile has no fields or subtypes
	inputFile := findType(t, api, "InputFile")
	assert.Empty(t, inputFile.Fields)
	assert.Empty(t, inputFile.Subtypes)
	assert.NotEmpty(t, inputFile.Description)

	// Spot-check: sendMessage method
	sendMsg := findMethod(t, api, "sendMessage")
	chatID := findParam(t, sendMsg, "chat_id")
	assert.Equal(t, ir.TypeExpr{Types: []ir.TypeRef{{Type: "Integer"}, {Type: "String"}}}, chatID.TypeExpr)
	assert.True(t, chatID.Required)
	assert.Equal(t, "Message", sendMsg.Returns.Types[0].Type)
	assert.Equal(t, "message", sendMsg.Returns.Types[0].Ref)

	// Spot-check: getUpdates returns Array of Update
	getUpdates := findMethod(t, api, "getUpdates")
	assert.Equal(t, 1, getUpdates.Returns.Array)
	assert.Equal(t, "Update", getUpdates.Returns.Types[0].Type)
	assert.Equal(t, "update", getUpdates.Returns.Types[0].Ref)

	// Spot-check: deleteWebhook returns True
	deleteWH := findMethod(t, api, "deleteWebhook")
	assert.Equal(t, ir.TypeExpr{Types: []ir.TypeRef{{Type: "True"}}}, deleteWH.Returns)

	// Spot-check: getWebhookInfo has no params
	getWHI := findMethod(t, api, "getWebhookInfo")
	assert.Empty(t, getWHI.Params)
	assert.Equal(t, "WebhookInfo", getWHI.Returns.Types[0].Type)
	assert.Equal(t, "webhookinfo", getWHI.Returns.Types[0].Ref)

	// Spot-check: Const fields (discriminators)
	messageOriginUser := findType(t, api, "MessageOriginUser")
	moTypeField := findField(t, messageOriginUser, "type")
	assert.Equal(t, "user", moTypeField.Const)

	botCmdDefault := findType(t, api, "BotCommandScopeDefault")
	scopeType := findField(t, botCmdDefault, "type")
	assert.Equal(t, "default", scopeType.Const)

	// Spot-check: Enum fields (value sets)
	chat := findType(t, api, "Chat")
	chatType := findField(t, chat, "type")
	assert.Equal(t, []string{"private", "group", "supergroup", "channel"}, chatType.Enum)
	assert.Empty(t, chatType.Const)

	sticker := findType(t, api, "Sticker")
	stickerType := findField(t, sticker, "type")
	assert.Equal(t, []string{"regular", "mask", "custom_emoji"}, stickerType.Enum)

	// Spot-check: Timestamp fields stay as Integer (handled by typegen rules)
	dateField := findField(t, message, "date")
	assert.Equal(t, "Integer", dateField.TypeExpr.Types[0].Type)
	editDate := findField(t, message, "edit_date")
	assert.Equal(t, "Integer", editDate.TypeExpr.Types[0].Type)

	// Spot-check: Param defaults
	limitParam := findParam(t, getUpdates, "limit")
	assert.Equal(t, "100", limitParam.Default)
	timeoutParam := findParam(t, getUpdates, "timeout")
	assert.Equal(t, "0", timeoutParam.Default)
	offsetParam := findParam(t, getUpdates, "offset")
	assert.Empty(t, offsetParam.Default)

	// Verify no method descriptions have leading/trailing whitespace
	for _, m := range api.Methods {
		for i, d := range m.Description {
			assert.Equal(t, strings.TrimSpace(d), d, "method %q description[%d] has leading/trailing whitespace", m.Name, i)
		}
	}

	// Verify no type descriptions have leading/trailing whitespace
	for _, typ := range api.Types {
		assert.Equal(t, strings.TrimSpace(typ.Description), typ.Description, "type %q description has leading/trailing whitespace", typ.Name)
	}
}

// --- Test helpers ---

func wrapDoc(body string) string {
	return `<!DOCTYPE html><html><head></head><body><div id="dev_page_content">` + body + `</div></body></html>`
}

func parseHTML(t *testing.T, s string) *html.Node {
	t.Helper()
	doc, err := html.Parse(strings.NewReader(s))
	require.NoError(t, err)
	return doc
}

func findFirstElement(n *html.Node, tag string) *html.Node {
	if n.Type == html.ElementNode && n.Data == tag {
		return n
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if found := findFirstElement(c, tag); found != nil {
			return found
		}
	}
	return nil
}

// collectTestNodes collects <p> and <blockquote> elements for return type testing.
func collectTestNodes(n *html.Node, result *[]*html.Node) {
	if n.Type == html.ElementNode && (n.Data == "p" || n.Data == "blockquote") {
		*result = append(*result, n)
		return
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		collectTestNodes(c, result)
	}
}

func findType(t *testing.T, api *ir.API, name string) ir.Type {
	t.Helper()
	for _, typ := range api.Types {
		if typ.Name == name {
			return typ
		}
	}
	t.Fatalf("type %q not found", name)
	return ir.Type{}
}

func findMethod(t *testing.T, api *ir.API, name string) ir.Method {
	t.Helper()
	for _, m := range api.Methods {
		if m.Name == name {
			return m
		}
	}
	t.Fatalf("method %q not found", name)
	return ir.Method{}
}

func findField(t *testing.T, typ ir.Type, name string) ir.Field {
	t.Helper()
	for _, f := range typ.Fields {
		if f.Name == name {
			return f
		}
	}
	t.Fatalf("field %q not found in type %q", name, typ.Name)
	return ir.Field{}
}

func findParam(t *testing.T, m ir.Method, name string) ir.Param {
	t.Helper()
	for _, p := range m.Params {
		if p.Name == name {
			return p
		}
	}
	t.Fatalf("param %q not found in method %q", name, m.Name)
	return ir.Param{}
}
