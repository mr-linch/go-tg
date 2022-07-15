// Package contains example of using tgb.ChatType filter.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
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

	router := tgb.NewRouter().
		Message(func(ctx context.Context, msg *tgb.MessageUpdate) error {
			return msg.Answer("this is private chat response").DoVoid(ctx)
		}, tgb.ChatType(tg.ChatTypePrivate)).
		Message(func(ctx context.Context, msg *tgb.MessageUpdate) error {
			return msg.Answer("this is group chat response").DoVoid(ctx)
		}, tgb.ChatType(tg.ChatTypeGroup, tg.ChatTypeSupergroup))

	return tgb.NewPoller(
		router,
		client,
	).Run(ctx)
}
