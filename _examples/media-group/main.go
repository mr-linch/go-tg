// Package contains example of sending media groups and paid media.
package main

import (
	"context"
	"regexp"
	"strconv"

	_ "embed"

	"github.com/mr-linch/go-tg"
	"github.com/mr-linch/go-tg/_examples/runner"
	"github.com/mr-linch/go-tg/tgb"
)

//go:embed resources/gopher.png
var gopherPNG []byte

func main() {
	runner.Run(tgb.NewRouter().
		Message(func(ctx context.Context, msg *tgb.MessageUpdate) error {
			return msg.Answer(
				"Send a number (1-10) for a media group, or /paid <count> <stars> for paid media.",
			).DoVoid(ctx)
		}, tgb.Command("start")).
		Message(func(ctx context.Context, msg *tgb.MessageUpdate) error {
			// handle /paid <count> <stars> command
			args := msg.Text
			parts := regexp.MustCompile(`\s+`).Split(args, -1)
			if len(parts) < 3 {
				return msg.Answer("usage: /paid <count> <stars>").DoVoid(ctx)
			}

			count, err := strconv.Atoi(parts[1])
			if err != nil || count < 1 || count > 10 {
				return msg.Answer("count should be between 1 and 10").DoVoid(ctx)
			}

			stars, err := strconv.Atoi(parts[2])
			if err != nil || stars < 1 || stars > 25000 {
				return msg.Answer("stars should be between 1 and 25000").DoVoid(ctx)
			}

			media := make([]tg.InputPaidMedia, count)
			for i := range media {
				media[i] = tg.NewInputPaidMediaPhoto(
					tg.NewFileArgUpload(
						tg.NewInputFileBytes("gopher.png", gopherPNG),
					),
				)
			}

			return msg.Client.SendPaidMedia(msg.Chat, stars, media).DoVoid(ctx)
		}, tgb.Command("paid")).
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
			for i := range media {
				media[i] = tg.NewInputMediaPhoto(
					tg.NewFileArgUpload(
						tg.NewInputFileBytes("gopher.png", gopherPNG),
					),
				)
			}

			return msg.AnswerMediaGroup(media).DoVoid(ctx)
		}, tgb.Regexp(regexp.MustCompile(`^(\d+)$`))).
		Message(func(ctx context.Context, msg *tgb.MessageUpdate) error {
			return msg.Answer("send a number (1-10) or use /paid <count> <stars>").DoVoid(ctx)
		}),
	)
}
