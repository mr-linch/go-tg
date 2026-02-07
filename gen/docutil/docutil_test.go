package docutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var testTypes = map[string]bool{
	"Message": true,
	"Update":  true,
	"User":    true,
}

var testMethods = map[string]string{
	"getMe":       "Client.GetMe",
	"sendMessage": "Client.SendMessage",
	"getFile":     "Client.GetFile",
}

func TestConvertLinks(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "no links",
			input: "Plain text with no links.",
			want:  "Plain text with no links.",
		},
		{
			name:  "known type becomes doc link",
			input: "Note that the [Message](https://core.telegram.org/bots/api#message) object in this field.",
			want:  "Note that the [Message] object in this field.",
		},
		{
			name:  "known method becomes doc link",
			input: "Returned only in [getMe](https://core.telegram.org/bots/api#getme).",
			want:  "Returned only in [Client.GetMe].",
		},
		{
			name:  "unknown link gets url definition",
			input: "See [formatting options](https://core.telegram.org/bots/api#formatting-options) for details.",
			want: "See [formatting options] for details.\n\n" +
				"[formatting options]: https://core.telegram.org/bots/api#formatting-options",
		},
		{
			name:  "mixed types methods and external",
			input: "The [Message](https://core.telegram.org/bots/api#message) is returned by [sendMessage](https://core.telegram.org/bots/api#sendmessage). See [more info](https://example.com).",
			want: "The [Message] is returned by [Client.SendMessage]. See [more info].\n\n" +
				"[more info]: https://example.com",
		},
		{
			name:  "duplicate external links deduplicated",
			input: "[vCard](https://en.wikipedia.org/wiki/VCard) and another [vCard](https://en.wikipedia.org/wiki/VCard) ref.",
			want: "[vCard] and another [vCard] ref.\n\n" +
				"[vCard]: https://en.wikipedia.org/wiki/VCard",
		},
		{
			name:  "multiple different external links",
			input: "A [foo](https://example.com/foo) and [bar](https://example.com/bar).",
			want: "A [foo] and [bar].\n\n" +
				"[foo]: https://example.com/foo\n" +
				"[bar]: https://example.com/bar",
		},
		{
			name:  "all doc links no definitions needed",
			input: "Uses [Message](https://core.telegram.org/bots/api#message) and [User](https://core.telegram.org/bots/api#user).",
			want:  "Uses [Message] and [User].",
		},
		{
			name:  "descriptive text resolved by url anchor to method",
			input: "Until the chat is [unbanned](https://core.telegram.org/bots/api#sendmessage).",
			want:  "Until the chat is [Client.SendMessage].",
		},
		{
			name:  "descriptive text resolved by url anchor to type",
			input: "Returns a [result](https://core.telegram.org/bots/api#message) object.",
			want:  "Returns a [Message] object.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ConvertLinks(tt.input, testTypes, testMethods)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestExtractLinks(t *testing.T) {
	input := "The [Message](https://core.telegram.org/bots/api#message) via [getFile](https://core.telegram.org/bots/api#getfile). See [wiki](https://en.wikipedia.org/wiki/Test)."

	converted, linkDefs := ExtractLinks(input, testTypes, testMethods)

	assert.Equal(t, "The [Message] via [Client.GetFile]. See [wiki].", converted)
	assert.Equal(t, []string{"[wiki]: https://en.wikipedia.org/wiki/Test"}, linkDefs)
}
