package tg

import (
	"encoding"
	"fmt"
	"html"
	"regexp"
	"strings"
)

type ParseMode interface {
	encoding.TextMarshaler
	fmt.Stringer

	// Change separator for next calls
	Sep(v string) ParseMode

	// Text joins the given strings with new line seprator
	Text(v ...string) string

	// Line joins the given strings with space seprator
	Line(v ...string) string

	// Escapef escapes the format string and then applies fmt.Sprintf with the given args.
	// Useful for MarkdownV2 where literal characters like . | ! need escaping,
	// while %s args (e.g. from Bold, Italic) pass through unchanged.
	Escapef(format string, args ...any) string

	// Bold
	Bold(v ...string) string

	// Italic
	Italic(v ...string) string

	// Underline
	Underline(v ...string) string

	// Deleted (Striketrough)
	Strike(v ...string) string

	// Spoiler
	Spoiler(v ...string) string

	// Link
	Link(title, url string) string

	// Mention creates an inline mention of a user by ID
	Mention(name string, userID UserID) string

	// CustomEmoji inserts a custom emoji by ID with fallback emoji text
	CustomEmoji(emoji, emojiID string) string

	// Code
	Code(v ...string) string

	// Preformated text
	Pre(v ...string) string

	// PreLanguage creates a pre-formatted code block with language specification
	PreLanguage(language string, v ...string) string

	// Blockquote
	Blockquote(v ...string) string

	// ExpandableBlockquote creates a collapsible blockquote
	ExpandableBlockquote(v ...string) string

	// Escape
	Escape(v string) string
}

func regexpReplacer(re *regexp.Regexp, repl string) func(string) string {
	return func(s string) string {
		return re.ReplaceAllString(s, repl)
	}
}

var (
	// HTML is ParseMode that uses HTML tags
	HTML ParseMode = parseMode{
		name:      "HTML",
		separator: " ",

		bold:                 parseModeTag{"<b>", "</b>"},
		italic:               parseModeTag{"<i>", "</i>"},
		underline:            parseModeTag{"<u>", "</u>"},
		strike:               parseModeTag{"<s>", "</s>"},
		spoiler:              parseModeTag{"<tg-spoiler>", "</tg-spoiler>"},
		code:                 parseModeTag{"<code>", "</code>"},
		pre:                  parseModeTag{"<pre>", "</pre>"},
		blockquote:           parseModeTag{"<blockquote>", "</blockquote>"},
		expandableBlockquote: parseModeTag{"<blockquote expandable>", "</blockquote>"},

		linkTemplate:        `<a href="{url}">{title}</a>`,
		customEmojiTemplate: `<tg-emoji emoji-id="{id}">{emoji}</tg-emoji>`,
		preLanguageTemplate: `<pre><code class="language-{lang}">{code}</code></pre>`,

		escape: html.EscapeString,
	}

	// MD is ParseMode that uses Markdown tags.
	// Warning: this is legacy mode, use MarkdownV2 (MD2) instead.
	MD ParseMode = parseMode{
		name:      "Markdown",
		separator: " ",

		bold:   parseModeTag{"*", "*"},
		italic: parseModeTag{"_", "_"},
		code:   parseModeTag{"`", "`"},
		pre:    parseModeTag{"```", "```"},

		linkTemplate:        `[{title}]({url})`,
		customEmojiTemplate: `{emoji}`,
		preLanguageTemplate: "```{lang}\n{code}```",

		escape: regexpReplacer(regexp.MustCompile(`([_*\x60\[])`), `\$1`),
	}

	// MD2 is ParseMode that uses MarkdownV2 tags.
	MD2 ParseMode = parseMode{
		name:      "MarkdownV2",
		separator: " ",

		bold:                 parseModeTag{"*", "*"},
		italic:               parseModeTag{"_", "_"},
		underline:            parseModeTag{"__", "__"},
		strike:               parseModeTag{"~", "~"},
		spoiler:              parseModeTag{"||", "||"},
		code:                 parseModeTag{"`", "`"},
		pre:                  parseModeTag{"```", "```"},
		blockquote:           parseModeTag{">", ""},
		expandableBlockquote: parseModeTag{">", "||"},
		lineBlockquote:       true,

		linkTemplate:        `[{title}]({url})`,
		customEmojiTemplate: `![{emoji}](tg://emoji?id={id})`,
		preLanguageTemplate: "```{lang}\n{code}```",

		escape: regexpReplacer(regexp.MustCompile(`([_*\[\]()~\x60>#\+\-=|{}.!])`), `\$1`),
	}
)

type parseMode struct {
	separator string

	name string

	bold                 parseModeTag
	italic               parseModeTag
	underline            parseModeTag
	strike               parseModeTag
	spoiler              parseModeTag
	code                 parseModeTag
	pre                  parseModeTag
	blockquote           parseModeTag
	expandableBlockquote parseModeTag
	linkTemplate         string
	customEmojiTemplate  string
	preLanguageTemplate  string
	lineBlockquote       bool

	escape func(string) string
}

type parseModeTag struct {
	start string
	end   string
}

func (pmt parseModeTag) wrap(content string) string {
	return pmt.start + content + pmt.end
}

func (pm parseMode) Text(v ...string) string {
	return strings.Join(v, "\n")
}

func (pm parseMode) Line(v ...string) string {
	return strings.Join(v, " ")
}

func (pm parseMode) Escapef(format string, args ...any) string {
	return fmt.Sprintf(pm.escape(format), args...)
}

func (pm parseMode) MarshalText() ([]byte, error) {
	return []byte(pm.String()), nil
}

func (pm parseMode) String() string {
	return pm.name
}

func (pm parseMode) Sep(sep string) ParseMode {
	pm.separator = sep
	return pm
}

// Bold
func (pm parseMode) Bold(v ...string) string {
	return pm.bold.wrap(strings.Join(v, pm.separator))
}

// Italic
func (pm parseMode) Italic(v ...string) string {
	return pm.italic.wrap(strings.Join(v, pm.separator))
}

// Underline. Warning: this is not supported by Markdown, use MarkdownV2.
func (pm parseMode) Underline(v ...string) string {
	return pm.underline.wrap(strings.Join(v, pm.separator))
}

// Strike. Warning: this is not supported by Markdown, use MarkdownV2.
func (pm parseMode) Strike(v ...string) string {
	return pm.strike.wrap(strings.Join(v, pm.separator))
}

// Spoiler. Warning: this is not supported by Markdown, use MarkdownV2.
func (pm parseMode) Spoiler(v ...string) string {
	return pm.spoiler.wrap(strings.Join(v, pm.separator))
}

// Link
func (pm parseMode) Link(title, url string) string {
	return strings.NewReplacer(
		"{title}", title,
		"{url}", url,
	).Replace(pm.linkTemplate)
}

// Mention creates an inline mention of a user by ID
func (pm parseMode) Mention(name string, userID UserID) string {
	return pm.Link(name, fmt.Sprintf("tg://user?id=%d", userID))
}

// CustomEmoji inserts a custom emoji by ID with fallback emoji text
func (pm parseMode) CustomEmoji(emoji, emojiID string) string {
	return strings.NewReplacer(
		"{emoji}", emoji,
		"{id}", emojiID,
	).Replace(pm.customEmojiTemplate)
}

// Code
func (pm parseMode) Code(v ...string) string {
	return pm.code.wrap(strings.Join(v, pm.separator))
}

// Preformated text
func (pm parseMode) Pre(v ...string) string {
	return pm.pre.wrap(strings.Join(v, pm.separator))
}

// PreLanguage creates a pre-formatted code block with language specification
func (pm parseMode) PreLanguage(language string, v ...string) string {
	code := strings.Join(v, pm.separator)
	return strings.NewReplacer(
		"{lang}", language,
		"{code}", code,
	).Replace(pm.preLanguageTemplate)
}

// Blockquote
func (pm parseMode) Blockquote(v ...string) string {
	content := strings.Join(v, pm.separator)
	if pm.lineBlockquote {
		return prefixLines(content, pm.blockquote.start)
	}
	return pm.blockquote.wrap(content)
}

// ExpandableBlockquote creates a collapsible blockquote
func (pm parseMode) ExpandableBlockquote(v ...string) string {
	content := strings.Join(v, pm.separator)
	if pm.lineBlockquote {
		return prefixLines(content, pm.expandableBlockquote.start) + pm.expandableBlockquote.end
	}
	return pm.expandableBlockquote.wrap(content)
}

func (pm parseMode) Escape(v string) string {
	return pm.escape(v)
}

func prefixLines(content, prefix string) string {
	lines := strings.Split(content, "\n")
	for i := range lines {
		lines[i] = prefix + lines[i]
	}
	return strings.Join(lines, "\n")
}
