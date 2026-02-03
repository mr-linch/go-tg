package readmegen

import (
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/mr-linch/go-tg/gen/ir"
)

// badgeBlockPattern matches the auto-generated badge block.
var badgeBlockPattern = regexp.MustCompile(
	`(?s)<!-- auto-generated: Telegram Bot API badge -->.*?<!-- end: auto-generated -->`,
)

// badgeLinePattern matches shields.io Telegram Bot API version badge (legacy, unwrapped).
// Example: ![Telegram Bot API](https://img.shields.io/badge/Telegram%20Bot%20API-7.2-blue?logo=telegram)
var badgeLinePattern = regexp.MustCompile(
	`(?m)^!\[Telegram Bot API\]\(https://img\.shields\.io/badge/Telegram%20Bot%20API-[0-9]+\.[0-9]+-blue\?logo=telegram\)$`,
)

// UpdateVersion updates the Telegram Bot API version badge and text in README.md.
func UpdateVersion(path string, api *ir.API) error {
	if api.Version == "" {
		return nil // Nothing to update
	}

	info, err := os.Stat(path)
	if err != nil {
		return err
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	// Build badge block with version and date in shields.io format.
	// Date needs URL encoding (spaces -> %20, commas -> %2C).
	badgeText := fmt.Sprintf("%s (from %s)", api.Version, api.ReleaseDate)
	encodedText := url.PathEscape(badgeText)

	// Build anchor from date: "December 31, 2025" -> "december-31-2025"
	anchor := strings.ToLower(api.ReleaseDate)
	anchor = strings.ReplaceAll(anchor, " ", "-")
	anchor = strings.ReplaceAll(anchor, ",", "")

	badgeBlock := fmt.Sprintf(
		"<!-- auto-generated: Telegram Bot API badge -->\n"+
			"[![Telegram Bot API](https://img.shields.io/badge/Telegram%%20Bot%%20API-%s-blue?logo=telegram)](https://core.telegram.org/bots/api#%s)\n"+
			"<!-- end: auto-generated -->",
		encodedText, anchor,
	)

	var updated []byte
	switch {
	case badgeBlockPattern.Match(content):
		// Replace existing badge block.
		updated = badgeBlockPattern.ReplaceAll(content, []byte(badgeBlock))
	case badgeLinePattern.Match(content):
		// Wrap legacy badge line with auto-generated markers.
		updated = badgeLinePattern.ReplaceAll(content, []byte(badgeBlock))
	default:
		updated = content
	}

	return os.WriteFile(path, updated, info.Mode())
}
