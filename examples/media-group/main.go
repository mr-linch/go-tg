// Package contains example of using tgb.ChatType filter.
package main

import (
	"context"
	"regexp"
	"strconv"

	_ "embed"

	"github.com/mr-linch/go-tg"
	"github.com/mr-linch/go-tg/examples"
	"github.com/mr-linch/go-tg/tgb"
)

var (
	flagToken string
)

var (
	//go:embed resources/gopher.png
	gopherPNG []byte
)

func main() {
	examples.Run(tgb.NewRouter().
		Message(func(ctx context.Context, msg *tgb.MessageUpdate) error {
			// handle /start command

			return msg.Answer("how much items I should send?").DoVoid(ctx)
		}, tgb.Command("start")).
		Message(func(ctx context.Context, msg *tgb.MessageUpdate) error {
			// handle messages matched integer regexp

			count, err := strconv.Atoi(msg.Text)
			if err != nil {
				return err
			}

			if count < 1 || count > 10 {
				return msg.Answer("count should be between 1 and 10").DoVoid(ctx)
			}

			media := make([]tg.InputMedia, count)

			for i := 0; i < count; i++ {
				media[i] = &tg.InputMediaPhoto{
					Media: tg.NewFileArgUpload(
						tg.NewInputFileBytes("gopher.png", gopherPNG),
					),
				}
			}

			return msg.AnswerMediaGroup(
				media,
			).DoVoid(ctx)
		}, tgb.Regexp(regexp.MustCompile(`^(\d+)$`))).
		Message(func(ctx context.Context, msg *tgb.MessageUpdate) error {
			// other message

			return msg.Answer("sorry, I don't understand you, send just number").DoVoid(ctx)
		}),
	)
}
