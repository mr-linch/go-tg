// Package contains example of using tgb.ChatType filter.
package main

import (
	"context"
	"fmt"
	"regexp"
	"strconv"

	"github.com/mr-linch/go-tg"
	"github.com/mr-linch/go-tg/examples"
	"github.com/mr-linch/go-tg/tgb"
)

func main() {
	const typingMessage = "use keyboard above for typing..."

	examples.Run(tgb.NewRouter().
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
	layout := tg.NewButtonLayout[tg.InlineKeyboardButton](3).Row(
		tg.NewInlineKeyboardButtonCallbackData("+", "+"),
		tg.NewInlineKeyboardButtonCallbackData("-", "-"),
		tg.NewInlineKeyboardButtonCallbackData("*", "*"),
		tg.NewInlineKeyboardButtonCallbackData("/", "/"),
	)

	for i := 9; i >= 0; i-- {
		text := strconv.Itoa(i)
		layout.Insert(
			tg.NewInlineKeyboardButtonCallbackData(text, text),
		)
	}

	layout.Insert(
		tg.NewInlineKeyboardButtonCallbackData(".", "."),
		tg.NewInlineKeyboardButtonCallbackData("=", "="),
	)

	return tg.NewInlineKeyboardMarkup(layout.Keyboard()...)
}
