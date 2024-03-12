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
	client         *tg.Client
	handler        Handler
	logger         Logger
	handlerTimeout time.Duration
	timeout        time.Duration
	retryAfter     time.Duration
	limit          int
	allowedUpdates []tg.UpdateType

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

// WithPollerLimit sets the limit for batch size.
func WithPollerLimit(limit int) PollerOption {
	return func(poller *Poller) {
		poller.limit = limit
	}
}

// WithPollerAllowedUpdates sets the allowed updates.
func WithPollerAllowedUpdates(allowedUpdates ...tg.UpdateType) PollerOption {
	return func(poller *Poller) {
		poller.allowedUpdates = allowedUpdates
	}
}

// WithPollerLogger sets the logger for the poller.
func WithPollerLogger(logger Logger) PollerOption {
	return func(poller *Poller) {
		poller.logger = logger
	}
}

const defaultPollerLimit = 100

func NewPoller(handler Handler, client *tg.Client, opts ...PollerOption) *Poller {
	poller := &Poller{
		client:  client,
		handler: handler,

		timeout:    time.Second * 5,
		retryAfter: time.Second * 5,

		allowedUpdates: []tg.UpdateType{},

		limit: defaultPollerLimit,
	}

	for _, opt := range opts {
		opt(poller)
	}

	return poller
}

func (poller *Poller) log(format string, args ...interface{}) {
	if poller.logger != nil {
		poller.logger.Printf("tgb.Poller: "+format, args...)
	}
}

func (poller *Poller) removeWebhookIfSet(ctx context.Context) error {
	info, err := poller.client.GetWebhookInfo().Do(ctx)
	if err != nil {
		return fmt.Errorf("get webhook info: %w", err)
	}

	if info.URL != "" {
		poller.log("removing webhook...")
		if err := poller.client.DeleteWebhook().DoVoid(ctx); err != nil {
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

			err := poller.handler.Handle(ctx, &Update{
				Update: update,
				Client: poller.client,
			})

			if err != nil {
				poller.log("error handling update: %v", err)
			}
		}(i)
	}
}

func (poller *Poller) Run(ctx context.Context) error {
	if err := poller.removeWebhookIfSet(ctx); err != nil {
		return fmt.Errorf("remove webhook if set: %w", err)
	}

	var offset int

	defer func() {
		poller.log("shutdown...")
		poller.wg.Wait()
	}()

	for {
		select {
		case <-ctx.Done():
			return nil
		default:

			call := poller.client.
				GetUpdates().
				Offset(offset).
				Timeout(int(poller.timeout.Seconds())).
				AllowedUpdates(poller.allowedUpdates)

			if poller.limit != defaultPollerLimit {
				call = call.Limit(poller.limit)
			}

			updates, err := call.Do(ctx)

			if err != nil && !errors.Is(err, context.Canceled) {
				poller.log("error '%s' when getting updates, retrying in %v...", err, poller.retryAfter)

				if poller.retryAfter > 0 {
					select {
					case <-time.After(poller.retryAfter):
					case <-ctx.Done():
						return nil
					}
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
