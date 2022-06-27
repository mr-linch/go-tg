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
		Message(func(ctx context.Context, msg *tgb.MessageUpdate) error {
			return msg.Answer(tg.HTML.Italic(typingMessage)).
				ParseMode(tg.HTML).
				ReplyMarkup(newKeyboard()).
				DoVoid(ctx)
		}).
		CallbackQuery(func(ctx context.Context, cbq *tgb.CallbackQueryUpdate) error {

			var currentText string

			if cbq.Message == nil {
				return cbq.AnswerText(
					tg.HTML.Italic("this keyboard is too old, please /start again"),
					true,
				).DoVoid(ctx)
			}

			currentText = cbq.Message.Text

			if currentText == typingMessage {
				currentText = ""
			}

			currentText += cbq.Data

			if err := cbq.Answer().DoVoid(ctx); err != nil {
				return fmt.Errorf("answer callback query: %w", err)
			}

			return cbq.Client.EditMessageText(
				cbq.Message.Chat.ID,
				cbq.Message.ID,
				currentText,
			).ReplyMarkup(newKeyboard()).DoVoid(ctx)
		})

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
