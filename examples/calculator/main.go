// Package contains example of using tgb.ChatType filter.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/mr-linch/go-tg"
	"github.com/mr-linch/go-tg/tgb"
)

var (
	flagToken string
)

func main() {
	flag.StringVar(&flagToken, "token", "", "Telegram Bot API token")
	flag.Parse()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill, syscall.SIGTERM)
	defer cancel()

	if err := run(ctx); err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context) error {
	if flagToken == "" {
		return fmt.Errorf("token is required")
	}

	client := tg.New(flagToken)

	me, err := client.Me(ctx)
	if err != nil {
		return fmt.Errorf("get me: %w", err)
	}
	log.Printf("auth as https://t.me/%s", me.Username)

	const typingMessage = "use keyboard above for typing..."

	bot := tgb.New().
		Message(tgb.HandlerFunc(func(ctx context.Context, update *tg.Update) error {

			return update.Respond(ctx, tg.NewSendMessageCall(
				update.Message.Chat,
				tg.HTML.Italic(typingMessage),
			).ReplyMarkup(newKeyboard()).ParseMode(tg.HTML))
		})).
		CallbackQuery(tgb.HandlerFunc(func(ctx context.Context, update *tg.Update) error {
			cbq := update.CallbackQuery

			var currentText string

			if cbq.Message == nil {
				return update.Respond(ctx, tg.NewAnswerCallbackQueryCall(
					cbq.ID,
				).Text("this keyboard is too old, please /start again").ShowAlert(true))
			}

			currentText = cbq.Message.Text

			if currentText == typingMessage {
				currentText = ""
			}

			currentText += cbq.Data

			return update.Respond(ctx, tg.NewEditMessageTextCall(
				cbq.Message.Chat,
				cbq.Message.ID,
				currentText,
			).ReplyMarkup(newKeyboard()))
		}))

	return tgb.NewPoller(
		bot,
		client,
	).Run(ctx)
}

func newKeyboard() tg.InlineKeyboardMarkup {
	layout := tg.NewButtonLayout[tg.InlineKeyboardButton](3).Row(
		tg.NewInlineKeyboardButtonCallback("+", "+"),
		tg.NewInlineKeyboardButtonCallback("-", "-"),
		tg.NewInlineKeyboardButtonCallback("*", "*"),
		tg.NewInlineKeyboardButtonCallback("/", "/"),
	)

	for i := 9; i >= 0; i-- {
		text := strconv.Itoa(i)
		layout.Insert(
			tg.NewInlineKeyboardButtonCallback(text, text),
		)
	}

	layout.Insert(
		tg.NewInlineKeyboardButtonCallback(".", "."),
		tg.NewInlineKeyboardButtonCallback("=", "="),
	)

	return tg.NewInlineKeyboardMarkup(layout.Keyboard()...)
}
