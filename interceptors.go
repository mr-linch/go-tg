package tg

import (
	"context"
	"errors"
	"math"
	"math/rand"
	"net/http"
	"time"
)

type InterceptorInvoker func(ctx context.Context, req *Request, dst any) error

// Interceptor is a function that intercepts request and response.
type Interceptor func(ctx context.Context, req *Request, dst any, invoker InterceptorInvoker) error

type retryFloodErrorOpts struct {
	tries         int
	maxRetryAfter time.Duration
	timeAfter     func(time.Duration) <-chan time.Time
}

// RetryFloodErrorOption is an option for NewRetryFloodErrorInterceptor.
type RetryFloodErrorOption func(*retryFloodErrorOpts)

// WithRetryFloodErrorTries sets the number of tries.
func WithRetryFloodErrorTries(tries int) RetryFloodErrorOption {
	return func(o *retryFloodErrorOpts) {
		o.tries = tries
	}
}

// WithRetryFloodErrorMaxRetryAfter sets the maximum retry after duration.
func WithRetryFloodErrorMaxRetryAfter(maxRetryAfter time.Duration) RetryFloodErrorOption {
	return func(o *retryFloodErrorOpts) {
		o.maxRetryAfter = maxRetryAfter
	}
}

// WithRetryFloodErrorTimeAfter sets the time.After function.
func WithRetryFloodErrorTimeAfter(timeAfter func(time.Duration) <-chan time.Time) RetryFloodErrorOption {
	return func(o *retryFloodErrorOpts) {
		o.timeAfter = timeAfter
	}
}

// NewRetryFloodErrorInterceptor returns a new interceptor that retries the request if the error is flood error.
func NewRetryFloodErrorInterceptor(opts ...RetryFloodErrorOption) Interceptor {
	options := retryFloodErrorOpts{
		tries:         3,
		maxRetryAfter: time.Hour,
		timeAfter:     time.After,
	}

	for _, o := range opts {
		o(&options)
	}

	return func(ctx context.Context, req *Request, dst any, invoker InterceptorInvoker) error {
		var err error
	LOOP:
		for i := 0; i < options.tries; i++ {
			err = invoker(ctx, req, dst)
			if err == nil {
				return nil
			}

			var tgErr *Error
			if errors.As(err, &tgErr) && tgErr.Code == http.StatusTooManyRequests && tgErr.Parameters != nil {
				if tgErr.Parameters.RetryAfterDuration() > options.maxRetryAfter {
					return err
				}

				select {
				case <-options.timeAfter(tgErr.Parameters.RetryAfterDuration()):
					continue LOOP
				case <-ctx.Done():
					return ctx.Err()
				}
			}

			break
		}

		return err
	}
}

type retryInternalServerErrorOpts struct {
	tries     int
	delay     time.Duration
	timeAfter func(time.Duration) <-chan time.Time
}

// RetryInternalServerErrorOption is an option for NewRetryInternalServerErrorInterceptor.
type RetryInternalServerErrorOption func(*retryInternalServerErrorOpts)

// WithRetryInternalServerErrorTries sets the number of tries.
func WithRetryInternalServerErrorTries(tries int) RetryInternalServerErrorOption {
	return func(o *retryInternalServerErrorOpts) {
		o.tries = tries
	}
}

// WithRetryInternalServerErrorDelay sets the delay between tries.
// The delay calculated as delay * 2^i + random jitter, where i is the number of tries.
func WithRetryInternalServerErrorDelay(delay time.Duration) RetryInternalServerErrorOption {
	return func(o *retryInternalServerErrorOpts) {
		o.delay = delay
	}
}

// WithRetryInternalServerErrorTimeAfter sets the time.After function.
func WithRetryInternalServerErrorTimeAfter(timeAfter func(time.Duration) <-chan time.Time) RetryInternalServerErrorOption {
	return func(o *retryInternalServerErrorOpts) {
		o.timeAfter = timeAfter
	}
}

// NewRetryInternalServerErrorInterceptor returns a new interceptor that retries the request if the error is internal server error.
func NewRetryInternalServerErrorInterceptor(opts ...RetryInternalServerErrorOption) Interceptor {
	options := &retryInternalServerErrorOpts{
		tries:     10,
		delay:     time.Millisecond * 100,
		timeAfter: time.After,
	}

	for _, o := range opts {
		o(options)
	}

	return func(ctx context.Context, req *Request, dst any, invoker InterceptorInvoker) error {
		var err error
	LOOP:
		for i := 0; i < options.tries; i++ {
			err = invoker(ctx, req, dst)
			if err == nil {
				return nil
			}

			var tgErr *Error
			if errors.As(err, &tgErr) && tgErr.Code == http.StatusInternalServerError {
				// do backoff delay
				backoffDelay := options.delay * time.Duration(math.Pow(2, float64(i)))
				jitter := time.Duration(rand.Int63n(int64(backoffDelay)))

				select {
				case <-options.timeAfter(backoffDelay + jitter):
					continue LOOP
				case <-ctx.Done():
					return ctx.Err()
				}
			}

			break
		}

		return err
	}
}
