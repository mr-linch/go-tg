// Package contains example of using tgb.ChatType filter.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"syscall"

	_ "embed"

	"github.com/mr-linch/go-tg"
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

	router := tgb.NewRouter().
		Message(func(ctx context.Context, msg *tgb.MessageUpdate) error {
			return msg.Answer("how much items I should send?").DoVoid(ctx)
		}, tgb.Command("start")).
		Message(func(ctx context.Context, msg *tgb.MessageUpdate) error {
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
					Media: tg.FileArg{
						Upload: tg.NewInputFileBytes("gopher.png", gopherPNG),
					},
				}
			}

			return msg.AnswerMediaGroup(
				media,
			).DoVoid(ctx)
		}, tgb.Regexp(regexp.MustCompile(`^(\d+)$`))).
		// other message
		Message(func(ctx context.Context, msg *tgb.MessageUpdate) error {
			return msg.Answer("sorry, I don't understand you, send just number").DoVoid(ctx)
		})

	return tgb.NewPoller(
		router,
		client,
		tgb.WithPollerLogger(log.Default()),
	).Run(ctx)
}
