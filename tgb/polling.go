package tgb

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	tg "github.com/mr-linch/go-tg"
)

// Poller is a long polling update deliverer.
type Poller struct {
	client  *tg.Client
	handler Handler

	handlerTimeout time.Duration
	timeout        time.Duration
	retryAfter     time.Duration

	wg sync.WaitGroup
}

type PollerOption func(*Poller)

// WithHandlerTimeout sets the timeout for Handler exectution.
func WithPollerHandlerTimeout(timeout time.Duration) PollerOption {
	return func(poller *Poller) {
		poller.handlerTimeout = timeout
	}
}

// WithPollerTimeout sets the timeout for polling.
func WithPollerTimeout(timeout time.Duration) PollerOption {
	return func(poller *Poller) {
		poller.timeout = timeout
	}
}

// WithPollerRetryAfter sets the retry after for polling.
func WithPollerRetryAfter(retryAfter time.Duration) PollerOption {
	return func(poller *Poller) {
		poller.retryAfter = retryAfter
	}
}

func NewPoller(handler Handler, client *tg.Client, opts ...PollerOption) *Poller {
	poller := &Poller{
		client:  client,
		handler: handler,

		timeout:    time.Second * 5,
		retryAfter: time.Second * 5,
	}

	for _, opt := range opts {
		opt(poller)
	}

	return poller
}

func (poller *Poller) removeWebhookIfSet(ctx context.Context) error {
	info, err := poller.client.GetWebhookInfo().Do(ctx)
	if err != nil {
		return fmt.Errorf("get webhook info: %w", err)
	}

	if info.URL != "" {
		if err := poller.client.DeleteWebhook().Do(ctx); err != nil {
			return fmt.Errorf("delete webhook: %w", err)
		}
	}

	return nil
}

func (poller *Poller) processUpdates(ctx context.Context, updates []tg.Update) {
	for i := range updates {
		poller.wg.Add(1)

		go func(i int) {
			defer poller.wg.Done()

			if poller.handlerTimeout > 0 {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, poller.handlerTimeout)
				defer cancel()
			}

			update := &updates[i]
			update.Bind(poller.client)

			_ = poller.handler.Handle(ctx, update)
		}(i)
	}
}

func (poller *Poller) Run(ctx context.Context) error {
	if err := poller.removeWebhookIfSet(ctx); err != nil {
		return fmt.Errorf("remove webhook if set: %w", err)
	}

	var offset int

	for {
		select {
		case <-ctx.Done():
			poller.wg.Wait()
			return nil
		default:
			updates, err := poller.client.
				GetUpdates().
				Offset(offset).
				Limit(100).
				Timeout(int(poller.timeout.Seconds())).
				Do(ctx)

			if err != nil && !errors.Is(err, context.Canceled) {
				if poller.retryAfter > 0 {
					time.Sleep(poller.retryAfter)
				}
				continue
			}

			if len(updates) > 0 {
				offset = updates[len(updates)-1].ID + 1
				go poller.processUpdates(ctx, updates)
			}
		}
	}

}
