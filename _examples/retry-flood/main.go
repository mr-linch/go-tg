package main

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/mr-linch/go-tg"
	"github.com/mr-linch/go-tg/examples"
	"github.com/mr-linch/go-tg/tgb"
)

func main() {
	pm := tg.HTML

	onStart := func(ctx context.Context, msg *tgb.MessageUpdate) error {
		return msg.Answer(pm.Text(
			"👋 Hi, I'm retry flood demo, send me /spam command for start.",
			"🔁 I will retry when receive flood wait error",
			"Stop spam with shutdown bot service",
		)).DoVoid(ctx)
	}

	onSpam := func(ctx context.Context, mu *tgb.MessageUpdate) error {
		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				var wg sync.WaitGroup

				for i := 0; i < 5; i++ {
					wg.Add(1)
					go func() {
						defer wg.Done()
						if err := mu.Answer(pm.Text("🔁 spamming...")).DoVoid(ctx); err != nil {
							log.Printf("answer: %v", err)
						}
					}()
				}

				wg.Wait()
			}
		}
	}

	examples.Run(tgb.NewRouter().
		Message(onSpam, tgb.Command("spam")).
		ChannelPost(onSpam, tgb.Command("spam")).
		Message(onStart).
		ChannelPost(onStart).
		Error(func(ctx context.Context, update *tgb.Update, err error) error {
			log.Printf("error in handler: %v", err)
			return nil
		}),

		tg.WithClientInterceptors(
			tg.Interceptor(func(ctx context.Context, req *tg.Request, dst any, invoker tg.InterceptorInvoker) error {
				defer func(started time.Time) {
					log.Printf("request: %s took: %s", req.Method, time.Since(started))
				}(time.Now())
				return invoker(ctx, req, dst)
			}),
			tg.NewInterceptorRetryFloodError(
				// we override the default timeAfter function to log the retry flood delay
				tg.WithInterceptorRetryFloodErrorTimeAfter(func(sleep time.Duration) <-chan time.Time {
					log.Printf("retry flood error after %s", sleep)
					return time.After(sleep)
				}),
			),
		),
	)
}
