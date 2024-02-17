// Package contains simple echo bot, that demonstrates how to use handlers, filters and file uploads.
package main

import (
	"context"
	"fmt"
	"regexp"
	"time"

	_ "embed"

	"github.com/mr-linch/go-tg"
	"github.com/mr-linch/go-tg/examples"
	"github.com/mr-linch/go-tg/tgb"
)

var (
	//go:embed resources/gopher.png
	gopherPNG []byte
)

func main() {
	examples.Run(tgb.NewRouter().
		Message(func(ctx context.Context, msg *tgb.MessageUpdate) error {
			// handles /start and /help
			return msg.Answer(
				tg.HTML.Text(
					tg.HTML.Bold("👋 Hi, I'm echo bot!"),
					tg.HTML.Line("Send me a message and I will echo it back to you. Also you can send me a reaction and I will react with the same emoji."),
					tg.HTML.Italic("🚀 Powered by", tg.HTML.Spoiler("go-tg")),
				),
			).ParseMode(tg.HTML).LinkPreviewOptions(tg.LinkPreviewOptions{
				URL:              "https://github.com/mr-linch/go-tg",
				PreferLargeMedia: true,
			}).DoVoid(ctx)

		}, tgb.Command("start", tgb.WithCommandAlias("help"))).
		Message(func(ctx context.Context, msg *tgb.MessageUpdate) error {
			// handles gopher image
			if err := msg.Update.Reply(ctx, msg.AnswerChatAction(tg.ChatActionUploadPhoto)); err != nil {
				return fmt.Errorf("answer chat action: %w", err)
			}

			time.Sleep(time.Second)

			return msg.AnswerPhoto(tg.NewFileArgUpload(
				tg.NewInputFileBytes("gopher.png", gopherPNG),
			)).DoVoid(ctx)

		}, tgb.Regexp(regexp.MustCompile(`(?mi)(go|golang|gopher)[$\s+]?`))).
		Message(func(ctx context.Context, msg *tgb.MessageUpdate) error {
			// handle other messages
			return msg.Update.Reply(ctx, msg.Copy(msg.Chat))
		}).
		MessageReaction(func(ctx context.Context, reaction *tgb.MessageReactionUpdate) error {
			// sets same reaction to the message
			answer := tg.NewSetMessageReactionCall(reaction.Chat, reaction.MessageID).Reaction(reaction.NewReaction)
			return reaction.Update.Reply(ctx, answer)
		}),
	)
}
