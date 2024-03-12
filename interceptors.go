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

type interceptorRetryFloodErrorOpts struct {
	tries         int
	maxRetryAfter time.Duration
	timeAfter     func(time.Duration) <-chan time.Time
}

// InterceptorRetryFloodErrorOption is an option for NewRetryFloodErrorInterceptor.
type InterceptorRetryFloodErrorOption func(*interceptorRetryFloodErrorOpts)

// WithInterceptorRetryFloodErrorTries sets the number of tries.
func WithInterceptorRetryFloodErrorTries(tries int) InterceptorRetryFloodErrorOption {
	return func(o *interceptorRetryFloodErrorOpts) {
		o.tries = tries
	}
}

// WithInterceptorRetryFloodErrorMaxRetryAfter sets the maximum retry after duration.
func WithInterceptorRetryFloodErrorMaxRetryAfter(maxRetryAfter time.Duration) InterceptorRetryFloodErrorOption {
	return func(o *interceptorRetryFloodErrorOpts) {
		o.maxRetryAfter = maxRetryAfter
	}
}

// WithInterceptorRetryFloodErrorTimeAfter sets the time.After function.
func WithInterceptorRetryFloodErrorTimeAfter(timeAfter func(time.Duration) <-chan time.Time) InterceptorRetryFloodErrorOption {
	return func(o *interceptorRetryFloodErrorOpts) {
		o.timeAfter = timeAfter
	}
}

// NewInterceptorRetryFloodError returns a new interceptor that retries the request if the error is flood error.
// With that interceptor, calling of method that hit limit will be look like it will look like the request just takes unusually long.
// Under the hood, multiple HTTP requests are being performed, with the appropriate delays in between.
//
// Default tries is 3, maxRetryAfter is 1 hour, timeAfter is time.After.
// The interceptor will retry the request if the error is flood error with RetryAfter less than maxRetryAfter.
// The interceptor will wait for RetryAfter duration before retrying the request.
// The interceptor will retry the request for tries times.
func NewInterceptorRetryFloodError(opts ...InterceptorRetryFloodErrorOption) Interceptor {
	options := interceptorRetryFloodErrorOpts{
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

type interceptorRetryInternalServerErrorOpts struct {
	tries     int
	delay     time.Duration
	timeAfter func(time.Duration) <-chan time.Time
}

// RetryInternalServerErrorOption is an option for NewRetryInternalServerErrorInterceptor.
type RetryInternalServerErrorOption func(*interceptorRetryInternalServerErrorOpts)

// WithInterceptorRetryInternalServerErrorTries sets the number of tries.
func WithInterceptorRetryInternalServerErrorTries(tries int) RetryInternalServerErrorOption {
	return func(o *interceptorRetryInternalServerErrorOpts) {
		o.tries = tries
	}
}

// WithInterceptorRetryInternalServerErrorDelay sets the delay between tries.
// The delay calculated as delay * 2^i + random jitter, where i is the number of tries.
func WithInterceptorRetryInternalServerErrorDelay(delay time.Duration) RetryInternalServerErrorOption {
	return func(o *interceptorRetryInternalServerErrorOpts) {
		o.delay = delay
	}
}

// WithInterceptorRetryInternalServerErrorTimeAfter sets the time.After function.
func WithInterceptorRetryInternalServerErrorTimeAfter(timeAfter func(time.Duration) <-chan time.Time) RetryInternalServerErrorOption {
	return func(o *interceptorRetryInternalServerErrorOpts) {
		o.timeAfter = timeAfter
	}
}

// NewInterceptorRetryInternalServerError returns a new interceptor that retries the request if the error is internal server error.
//
// With that interceptor, calling of method that hit limit will be look like it will look like the request just takes unusually long.
// Under the hood, multiple HTTP requests are being performed, with the appropriate delays in between.
//
// Default tries is 10, delay is 100ms, timeAfter is time.After.
// The interceptor will retry the request if the error is internal server error.
// The interceptor will wait for delay * 2^i + random jitter before retrying the request, where i is the number of tries.
// The interceptor will retry the request for ten times.
func NewInterceptorRetryInternalServerError(opts ...RetryInternalServerErrorOption) Interceptor {
	options := &interceptorRetryInternalServerErrorOpts{
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

// NewInterceptorDefaultParseMethod returns a new interceptor that sets the parse_method to the request if it is empty.
// Use in combination with NewInterceptorMethodFilter to filter and specify only needed methods.
// Like:
//
//	NewInterceptorMethodFilter(NewInterceptorDefaultParseMethod(tg.HTML), "sendMessage", "editMessageText")
func NewInterceptorDefaultParseMethod(pm ParseMode) Interceptor {
	return func(ctx context.Context, req *Request, dst any, invoker InterceptorInvoker) error {
		if !req.Has("parse_mode") {
			req.Stringer("parse_mode", pm)
		}

		return invoker(ctx, req, dst)
	}
}

// ТewInterceptorMethodFilter returns a new filtering interceptor
// that calls the interceptor only for specified methods.
func NewInterceptorMethodFilter(interceptor Interceptor, methods ...string) Interceptor {
	methodMap := make(map[string]struct{}, len(methods))
	for _, method := range methods {
		methodMap[method] = struct{}{}
	}

	return func(ctx context.Context, req *Request, dst any, invoker InterceptorInvoker) error {
		if _, ok := methodMap[req.Method]; ok {
			return interceptor(ctx, req, dst, invoker)
		}

		return invoker(ctx, req, dst)
	}
}
