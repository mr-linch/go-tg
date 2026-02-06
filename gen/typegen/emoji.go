package typegen

import (
	"strings"

	"github.com/forPelevin/gomoji"
)

// emojiName returns a snake_case human-readable name for an emoji character
// using the gomoji library's slug database.
// Returns false if the emoji is not recognized.
func emojiName(emoji string) (string, bool) {
	info, err := gomoji.GetInfo(emoji)
	if err != nil {
		return "", false
	}
	return strings.ReplaceAll(info.Slug, "-", "_"), true
}
