package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
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
	flagDebug         bool
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
	flag.BoolVar(&flagDebug, "debug", false, "Debug mode")
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
		tg.WithServer(flagServer),
	)

	me, err := client.Me(ctx)
	if err != nil {
		return fmt.Errorf("get me: %w", err)
	}
	log.Printf("auth as https://t.me/%s", me.Username)

	bot := newBot()

	if flagWebhookURL != "" {
		return runWebhook(ctx, client, bot, flagWebhookURL, flagWebhookListen)
	} else {
		return runPolling(ctx, client, bot)
	}
}

func newBot() *tgb.Bot {

	return tgb.New().
		// handles /start and /help
		Message(tgb.HandlerFunc(func(ctx context.Context, update *tg.Update) error {
			return update.Respond(ctx, tg.NewSendMessageCall(
				update.Message.Chat,
				tg.HTML.Text(
					tg.HTML.Bold("👋 Hi, I'm echo bot!"),
					"",
					tg.HTML.Italic("🚀 Powered by", tg.HTML.Spoiler(tg.HTML.Link("go-tg", "github.com/mr-linch/go-tg"))),
				),
			).ParseMode(tg.HTML))

		}), tgb.Command("start", tgb.WithCommandAlias("help"))).
		// handles gopher image
		Message(tgb.HandlerFunc(func(ctx context.Context, update *tg.Update) error {
			// if update is from webhook we will answer to it here
			if err := update.Respond(ctx, tg.NewSendChatActionCall(
				update.Message.Chat,
				"upload_photo",
			)); err != nil {
				return fmt.Errorf("send chat action: %w", err)
			}

			// second call was send as call to api
			return update.Respond(ctx, tg.NewSendPhotoCall(update.Message.Chat, tg.FileArg{
				Upload: tg.NewInputFileBytes("gopher.png", gopherPNG),
			}))

		}), tgb.Regexp(regexp.MustCompile(`(go|golang|gopher)[$\s+]?`))).
		// handle other messages
		Message(tgb.HandlerFunc(func(ctx context.Context, update *tg.Update) error {
			return update.Respond(ctx, tg.NewCopyMessageCall(
				update.Message.Chat,
				update.Message.Chat,
				update.Message.ID,
			))
		}))

}

func runPolling(ctx context.Context, client *tg.Client, bot *tgb.Bot) error {
	poller := tgb.NewPoller(
		bot,
		client,
	)

	log.Printf("start poller")
	if err := poller.Run(ctx); err != nil {
		return fmt.Errorf("start polling: %w", err)
	}

	return nil
}

func runWebhook(ctx context.Context, client *tg.Client, bot *tgb.Bot, url, listen string) error {
	webhook := tgb.NewWebhook(
		url,
		bot,
		client,
	)

	if err := webhook.Setup(ctx, true); err != nil {
		return fmt.Errorf("webhook: %w", err)
	}

	server := http.Server{
		Handler: webhook,
		Addr:    listen,
	}

	go func() {
		<-ctx.Done()

		log.Printf("shutdown webhook server")

		closeCtx, close := context.WithTimeout(context.Background(), 10*time.Second)
		defer close()

		if err := server.Shutdown(closeCtx); err != nil {
			log.Printf("server shutdown error: %v", err)
		}
	}()

	log.Printf("start webhook server on %s", listen)
	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		return fmt.Errorf("listen and serve: %w", err)
	}

	return nil
}
