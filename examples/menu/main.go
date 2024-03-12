// Package contains simple echo bot, that demonstrates how to use handlers, filters and file uploads.
package main

import (
	"context"
	"strconv"
	"time"

	"github.com/mr-linch/go-tg"
	"github.com/mr-linch/go-tg/examples"
	"github.com/mr-linch/go-tg/tgb"
)

var pm = tg.HTML

func newMenuMainMessage() *tgb.TextMessageCallBuilder {
	return tgb.NewTextMessageCallBuilder(
		pm.Text(
			pm.Bold("👋 Hi, I'm demo of", pm.Code("tg.TextMessageCallBuilder")),
			"",
			pm.Italic("Use attached keyboard or commands to navigate"),
		),
	).ReplyMarkup(tg.NewInlineKeyboardMarkup(
		tg.NewButtonRow(
			tg.NewInlineKeyboardButtonCallback("menu 1", "menu_1"),
			tg.NewInlineKeyboardButtonCallback("menu 2", "menu_2"),
			tg.NewInlineKeyboardButtonCallback("menu 3", "menu_3"),
		),
	)).ParseMode(pm)
}

func newSubmenu(n int) *tgb.TextMessageCallBuilder {
	return tgb.NewTextMessageCallBuilder(
		pm.Text(
			pm.Bold("Menu ", strconv.Itoa(n)),
			"",
			pm.Bold("Now:", " ", pm.Code(time.Now().Format(time.RFC3339))),
			"",
			pm.Italic("Use attached keyboard or commands to navigate"),
		),
	).ReplyMarkup(
		tg.NewInlineKeyboardMarkup(
			tg.NewButtonRow(
				tg.NewInlineKeyboardButtonCallback("go back", "menu_main"),
				tg.NewInlineKeyboardButtonCallback("refresh", "menu_"+strconv.Itoa(n)),
			),
		),
	).ParseMode(pm)
}

func main() {
	examples.Run(tgb.NewRouter().
		// start message and cbq handlers
		Message(func(ctx context.Context, msg *tgb.MessageUpdate) error {
			return msg.Update.Reply(ctx, newMenuMainMessage().AsSend(msg.Chat))
		}, tgb.Command("start")).
		CallbackQuery(func(ctx context.Context, cbq *tgb.CallbackQueryUpdate) error {
			return cbq.Update.Reply(ctx, newMenuMainMessage().AsEditTextFromCBQ(cbq.CallbackQuery))
		}, tgb.TextEqual("menu_main")).

		// switch menu handlers
		CallbackQuery(func(ctx context.Context, cbq *tgb.CallbackQueryUpdate) error {
			_ = cbq.Update.Reply(ctx, cbq.Answer())

			menuNum := cbq.Data[len("menu_"):]

			n, err := strconv.Atoi(menuNum)
			if err != nil {
				return cbq.AnswerText("invalid menu number", true).DoVoid(ctx)
			}

			return cbq.Update.Reply(ctx, newSubmenu(n).AsEditTextFromCBQ(cbq.CallbackQuery))
		}, tgb.Any(tgb.TextHasPrefix("menu_"))),
	)
}
