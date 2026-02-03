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

	// Code
	Code(v ...string) string

	// Preformated text
	Pre(v ...string) string

	// Blockquote
	Blockquote(v ...string) string

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

		bold:       parseModeTag{"<b>", "</b>"},
		italic:     parseModeTag{"<i>", "</i>"},
		underline:  parseModeTag{"<u>", "</u>"},
		strike:     parseModeTag{"<s>", "</s>"},
		spoiler:    parseModeTag{"<tg-spoiler>", "</tg-spoiler>"},
		code:       parseModeTag{"<code>", "</code>"},
		pre:        parseModeTag{"<pre>", "</pre>"},
		blockquote: parseModeTag{"<blockquote>", "</blockquote>"},

		linkTemplate: `<a href="{url}">{title}</a>`,

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

		linkTemplate: `[{title}]({url})`,

		escape: regexpReplacer(regexp.MustCompile(`([_*\x60\[])`), `\$1`),
	}

	// MD2 is ParseMode that uses MarkdownV2 tags.
	MD2 ParseMode = parseMode{
		name:      "MarkdownV2",
		separator: " ",

		bold:       parseModeTag{"*", "*"},
		italic:     parseModeTag{"_", "_"},
		underline:  parseModeTag{"__", "__"},
		strike:     parseModeTag{"~", "~"},
		spoiler:    parseModeTag{"||", "||"},
		code:       parseModeTag{"`", "`"},
		pre:        parseModeTag{"```", "```"},
		blockquote: parseModeTag{">", ""},

		linkTemplate: `[{title}]({url})`,

		escape: regexpReplacer(regexp.MustCompile(`([_*\[\]()~\x60>#\+\-=|{}.!])`), `\$1`),
	}
)

type parseMode struct {
	separator string

	name string

	bold         parseModeTag
	italic       parseModeTag
	underline    parseModeTag
	strike       parseModeTag
	spoiler      parseModeTag
	code         parseModeTag
	pre          parseModeTag
	blockquote   parseModeTag
	linkTemplate string

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

// Code
func (pm parseMode) Code(v ...string) string {
	return pm.code.wrap(strings.Join(v, pm.separator))
}

// Preformated text
func (pm parseMode) Pre(v ...string) string {
	return pm.pre.wrap(strings.Join(v, pm.separator))
}

// Blockquote
func (pm parseMode) Blockquote(v ...string) string {
	return pm.blockquote.wrap(strings.Join(v, pm.separator))
}

func (pm parseMode) Escape(v string) string {
	return pm.escape(v)
}
