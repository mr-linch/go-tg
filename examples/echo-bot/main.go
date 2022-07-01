// Package contains simple echo bot, that demonstrates how to use handlers, filters and file uploads.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"regexp"
	"syscall"
	"time"

	_ "embed"

	"github.com/mr-linch/go-tg"
	"github.com/mr-linch/go-tg/tgb"
)

var (
	flagToken         string
	flagServer        string
	flagWebhookURL    string
	flagWebhookListen string
)

var (
	//go:embed resources/gopher.png
	gopherPNG []byte
)

func main() {
	flag.StringVar(&flagToken, "token", "", "Telegram Bot API token")
	flag.StringVar(&flagServer, "server", "https://api.telegram.org", "Telegram Bot API server")
	flag.StringVar(&flagWebhookURL, "webhook-url", "", "Telegram Bot API webhook URL, if not provdide run in longpoll mode")
	flag.StringVar(&flagWebhookListen, "webhook-listen", ":8000", "Telegram Bot API webhook URL")
	flag.Parse()

	ctx := context.Background()

	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt, os.Kill, syscall.SIGTERM)
	defer cancel()

	if err := run(ctx); err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context) error {
	if flagToken == "" {
		return fmt.Errorf("token is required")
	}

	client := tg.New(flagToken,
		tg.WithClient(flagServer),
	)

	me, err := client.Me(ctx)
	if err != nil {
		return fmt.Errorf("get me: %w", err)
	}
	log.Printf("auth as https://t.me/%s", me.Username)

	bot := newBot()

	if flagWebhookURL != "" {
		return tgb.NewWebhook(
			bot,
			client,
			flagWebhookURL,
			tgb.WithDropPendingUpdates(true),
			tgb.WithWebhookLogger(log.Default()),
		).Run(
			ctx,
			flagWebhookListen,
		)
	} else {
		return tgb.NewPoller(
			bot,
			client,
			tgb.WithPollerLogger(log.Default()),
		).Run(ctx)
	}
}

func newBot() *tgb.Bot {
	return tgb.New().
		// handles /start and /help
		Message(func(ctx context.Context, msg *tgb.MessageUpdate) error {
			return msg.Answer(
				tg.HTML.Text(
					tg.HTML.Bold("👋 Hi, I'm echo bot!"),
					"",
					tg.HTML.Italic("🚀 Powered by", tg.HTML.Spoiler(tg.HTML.Link("go-tg", "github.com/mr-linch/go-tg"))),
				),
			).ParseMode(tg.HTML).DoVoid(ctx)
		}, tgb.Command("start", tgb.WithCommandAlias("help"))).
		// handles gopher image
		Message(func(ctx context.Context, msg *tgb.MessageUpdate) error {
			if err := msg.Update.Respond(ctx, msg.AnswerChatAction(tg.ChatActionUploadPhoto)); err != nil {
				return fmt.Errorf("answer chat action: %w", err)
			}

			time.Sleep(time.Second)

			return msg.AnswerPhoto(tg.FileArg{
				Upload: tg.NewInputFileBytes("gopher.png", gopherPNG),
			}).DoVoid(ctx)

		}, tgb.Regexp(regexp.MustCompile(`(?mi)(go|golang|gopher)[$\s+]?`))).
		// handle other messages
		Message(func(ctx context.Context, msg *tgb.MessageUpdate) error {
			return msg.Copy(msg.Chat).DoVoid(ctx)
		})

}
