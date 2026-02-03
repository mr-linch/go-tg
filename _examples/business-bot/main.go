// Package contains simple echo bot, that demonstrates how to use handlers, filters and file uploads.
package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/mr-linch/go-tg"
	"github.com/mr-linch/go-tg/_examples/runner"
	"github.com/mr-linch/go-tg/tgb"
)

var pm = tg.HTML

func main() {
	runner.Run(tgb.NewRouter().
		Message(func(ctx context.Context, mu *tgb.MessageUpdate) error {
			me, err := mu.Client.GetMe().Do(ctx)
			if err != nil {
				return fmt.Errorf("get me: %w", err)
			}

			if !me.CanConnectToBusiness {
				return mu.Answer("Bussines features is not enabled for current bot. Enable it via @BotFather").DoVoid(ctx)
			}

			if mu.From != nil && !mu.From.IsPremium {
				return mu.Answer("Bussines features works only for Telegram Premium users. Purchase subscription before use that bot.").DoVoid(ctx)
			}

			return mu.Answer("Connect bot in Telegram Bussines settings of your account").DoVoid(ctx)
		}, tgb.Command("start")).
		BusinessMessage(func(ctx context.Context, mu *tgb.MessageUpdate) error {
			if strings.Contains(mu.Text, "ping") {
				return mu.Answer("Pong!").BusinessConnectionID(mu.BusinessConnectionID).DoVoid(ctx)
			}

			log.Printf("New business message #%d: %s", mu.ID, mu.Text)

			return nil
		}).
		EditedBusinessMessage(func(ctx context.Context, mu *tgb.MessageUpdate) error {
			log.Printf("Edited business message: %#v", mu.Message)

			return nil
		}).
		BusinessConnection(func(ctx context.Context, bcu *tgb.BusinessConnectionUpdate) error {
			log.Printf("New business connection: %#v", bcu.BusinessConnection)

			lines := []string{}

			if bcu.IsEnabled {
				lines = append(lines, "🤝 Business connection estabilished")
			} else {
				lines = append(lines, "❌ Business connection closed")
			}

			lines = append(lines, "")

			lines = append(lines, pm.Line(
				pm.Bold("ID: "), pm.Code(bcu.BusinessConnection.ID),
			))

			canReply := bcu.Rights != nil && bcu.Rights.CanReply
			lines = append(lines, pm.Line(
				pm.Bold("Can Reply? "), pm.Code(fmt.Sprintf("%t", canReply)),
			))

			return bcu.Update.Reply(ctx,
				tg.NewSendMessageCall(
					bcu.User,
					pm.Text(lines...),
				).
					ParseMode(pm),
			)
		}),
	)
}
