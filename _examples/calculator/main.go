// Package contains example of using tgb.ChatType filter.
package main

import (
	"context"
	"fmt"
	"regexp"
	"strconv"

	"github.com/mr-linch/go-tg"
	"github.com/mr-linch/go-tg/_examples/runner"
	"github.com/mr-linch/go-tg/tgb"
)

func main() {
	const typingMessage = "use keyboard above for typing..."

	runner.Run(tgb.NewRouter().
		Message(func(ctx context.Context, msg *tgb.MessageUpdate) error {
			return msg.Answer(tg.HTML.Italic(typingMessage)).
				ParseMode(tg.HTML).
				ReplyMarkup(newCalculatorMessageKeyboard()).
				DoVoid(ctx)
		}).
		CallbackQuery(func(ctx context.Context, cbq *tgb.CallbackQueryUpdate) error {
			// handle special case of "=" button
			return cbq.AnswerText("not implemented", true).DoVoid(ctx)
		}, tgb.Regexp(regexp.MustCompile(`=`))).
		CallbackQuery(func(ctx context.Context, cbq *tgb.CallbackQueryUpdate) error {
			// handle other buttons
			var currentText string

			if cbq.Message.IsInaccessible() {
				return cbq.AnswerText(
					tg.HTML.Italic("this keyboard is too old, please /start again"),
					true,
				).DoVoid(ctx)
			}

			msg := cbq.Message.Message

			currentText = msg.Text

			if currentText == typingMessage {
				currentText = ""
			}

			currentText += cbq.Data

			if err := cbq.Answer().DoVoid(ctx); err != nil {
				return fmt.Errorf("answer callback query: %w", err)
			}

			return cbq.Client.EditMessageText(
				msg.Chat.ID,
				msg.ID,
				currentText,
			).ReplyMarkup(newCalculatorMessageKeyboard()).DoVoid(ctx)
		}),
	)
}

func newCalculatorMessageKeyboard() tg.InlineKeyboardMarkup {
	kb := tg.NewInlineKeyboard().
		Callback("+", "+").Callback("-", "-").Callback("*", "*").Callback("/", "/").Row()

	for i := 9; i >= 0; i-- {
		kb.Callback(strconv.Itoa(i), strconv.Itoa(i))
	}

	return kb.Callback(".", ".").Callback("=", "=").Adjust(3).Markup()
}
