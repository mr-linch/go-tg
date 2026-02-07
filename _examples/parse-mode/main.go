// Package contains example that demonstrates all ParseMode formatting features
// with an inline keyboard to switch between HTML, MarkdownV2 and Markdown modes.
package main

import (
	"context"
	"fmt"

	"github.com/mr-linch/go-tg"
	"github.com/mr-linch/go-tg/_examples/runner"
	"github.com/mr-linch/go-tg/tgb"
)

type parseModeCallback struct {
	Mode string
}

var parseModeFilter = tgb.NewCallbackDataFilter[parseModeCallback]("pm")

var modes = []struct {
	key  string
	mode tg.ParseMode
}{
	{"html", tg.HTML},
	{"md2", tg.MD2},
	{"md", tg.MD},
}

func resolveMode(key string) tg.ParseMode {
	for _, m := range modes {
		if m.key == key {
			return m.mode
		}
	}
	return tg.HTML
}

func newKeyboard(active string) tg.InlineKeyboardMarkup {
	var buttons []tg.InlineKeyboardButton
	for _, m := range modes {
		label := m.mode.String()
		if m.key == active {
			label = "✓ " + label
		}
		buttons = append(buttons, parseModeFilter.MustButton(label, parseModeCallback{Mode: m.key}))
	}
	return tg.NewInlineKeyboardMarkup(buttons)
}

func newMessageBuilder(modeKey string, user tg.User) *tgb.TextMessageCallBuilder {
	pm := resolveMode(modeKey)

	// Usually you should use HTML parse mode where only user input needs escaping.
	// Escapef is handy when the mode is switchable at runtime: it escapes the format
	// string (so literal . | ! etc. become safe for MarkdownV2) while %s args
	// (already wrapped in Bold, Italic, etc.) pass through unchanged.

	text := pm.Text(
		pm.Escapef("%s: %s %s", pm.Bold("Formatting Demo"), pm.Italic(pm.String()), pm.CustomEmoji("⭐", "5463392464314315076")),
		"",

		pm.Bold("Text Styles:"),
		pm.Escapef("%s | %s | %s | %s", pm.Bold("bold"), pm.Italic("italic"), pm.Underline("underline"), pm.Strike("strike")),
		pm.Escapef("%s | %s", pm.Bold(pm.Italic("bold italic")), pm.Spoiler("hidden spoiler")),
		"",

		pm.Bold("Code:"),
		pm.Escapef("Inline: %s", pm.Code("fmt.Println()")),
		pm.PreLanguage("go", fmt.Sprintf("func main() {\n    fmt.Println(%q)\n}", "Hello from "+pm.String())),
		"",

		pm.Bold("Links & Mentions:"),
		pm.Link("Telegram Bot API", "https://core.telegram.org/bots/api"),
		pm.Mention(pm.Escape(user.FirstName), tg.UserID(user.ID)),
		"",

		pm.Bold("Quotes:"),
		pm.Blockquote(pm.Text(
			pm.Escape("This is a blockquote."),
			pm.Italic(pm.Escape("It can span multiple lines.")),
		)),
		"",
		pm.ExpandableBlockquote(pm.Text(
			pm.Bold("Expandable quote"),
			pm.Escape("Click to expand and see more content."),
			pm.Escape("This part is hidden by default."),
			pm.Escape("Telegram collapses long blockquotes."),
			pm.Italic(pm.Escape("Line 4: italic text inside expandable quote.")),
			pm.Bold(pm.Escape("Line 5: bold text inside expandable quote.")),
			pm.Escapef("Line 6: %s inside expandable quote.", pm.Code("inline code")),
			pm.Escape("Line 7: the last line of the expandable blockquote."),
		)),
	)

	return tgb.NewTextMessageCallBuilder(text).
		ParseMode(pm).
		ReplyMarkup(newKeyboard(modeKey))
}

func main() {
	runner.Run(tgb.NewRouter().
		Message(func(ctx context.Context, msg *tgb.MessageUpdate) error {
			return msg.Update.Reply(ctx, newMessageBuilder("html", *msg.From).AsSend(msg.Chat))
		}, tgb.Command("start", tgb.WithCommandAlias("help"))).
		CallbackQuery(parseModeFilter.Handler(func(ctx context.Context, cbq *tgb.CallbackQueryUpdate, cbd parseModeCallback) error {
			go cbq.Answer().DoVoid(ctx)

			if cbq.Message.IsInaccessible() {
				return nil
			}

			return cbq.Update.Reply(ctx, newMessageBuilder(cbd.Mode, cbq.From).AsEditTextFromCBQ(cbq.CallbackQuery))
		}), parseModeFilter.Filter()),
	)
}
