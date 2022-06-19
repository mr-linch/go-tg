package tg

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseMode_String(t *testing.T) {
	tests := []struct {
		mode ParseMode
		want string
	}{
		{
			ParseModeMarkdown,
			"Markdown",
		},
		{
			ParseModeHTML,
			"HTML",
		},
		{
			ParseModeMarkdownV2,
			"MarkdownV2",
		},
		{
			ParseMode(12),
			"",
		},
	}
	for _, tt := range tests {
		got := tt.mode.String()
		assert.Equal(t, tt.want, got)
	}
}

func TestPeerIDImpl(t *testing.T) {
	for _, test := range []struct {
		PeerID PeerID
		Want   string
	}{
		{UserID(1), "1"},
		{ChatID(1), "1"},
		// {&User{ID: UserID(1)}, "1"},
	} {
		assert.Equal(t, test.Want, test.PeerID.PeerID())
	}
}
