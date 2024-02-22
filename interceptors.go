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

// NewRetryFloodErrorInterceptor returns a new interceptor that retries the request if the error is flood error.
func NewRetryFloodErrorInterceptor(tries int, maxRetryAfter time.Duration) Interceptor {
	return func(ctx context.Context, req *Request, dst any, invoker InterceptorInvoker) error {
	LOOP:
		for i := 0; i < tries; i++ {
			err := invoker(ctx, req, dst)
			if err == nil {
				return nil
			}

			var tgErr *Error
			if errors.As(err, &tgErr) && tgErr.Code == http.StatusTooManyRequests && tgErr.Parameters != nil {
				if tgErr.Parameters.RetryAfterDuration() > maxRetryAfter {
					return err
				}

				select {
				case <-time.After(tgErr.Parameters.RetryAfterDuration()):
					continue LOOP
				case <-ctx.Done():
					return ctx.Err()
				}
			}

			return err
		}

		return nil
	}
}

// NewRetryInternalServerErrorInterceptor returns a new interceptor that retries the request if the error is internal server error.
func NewRetryInternalServerErrorInterceptor(tries int, delay time.Duration) Interceptor {
	return func(ctx context.Context, req *Request, dst any, invoker InterceptorInvoker) error {
	LOOP:
		for i := 0; i < tries; i++ {
			err := invoker(ctx, req, dst)
			if err == nil {
				return nil
			}

			var tgErr *Error
			if errors.As(err, &tgErr) && tgErr.Code == http.StatusInternalServerError {
				// do backoff delay
				backoffDelay := delay * time.Duration(math.Pow(2, float64(i)))
				jitter := time.Duration(rand.Int63n(int64(backoffDelay)))

				select {
				case <-time.After(backoffDelay + jitter):
					continue LOOP
				case <-ctx.Done():
					return ctx.Err()
				}
			}

			return err
		}

		return nil
	}
}
