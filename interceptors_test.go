package tg

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewRetryFloodErrorInterceptor(t *testing.T) {
	t.Run("TestNoError", func(t *testing.T) {
		var calls int

		invoker := InterceptorInvoker(func(ctx context.Context, req *Request, dst any) error {
			calls++
			return nil
		})

		interceptor := NewRetryFloodErrorInterceptor()

		err := interceptor(context.Background(), &Request{}, nil, invoker)

		assert.NoError(t, err, "should no return error")
		assert.Equal(t, 1, calls, "should call invoker once")
	})

	t.Run("NoTgError", func(t *testing.T) {
		var calls int

		invoker := InterceptorInvoker(func(ctx context.Context, req *Request, dst any) error {
			calls++
			return errors.New("test")
		})

		interceptor := NewRetryFloodErrorInterceptor()

		err := interceptor(context.Background(), &Request{}, nil, invoker)

		assert.Error(t, err, "should return error")
		assert.Equal(t, 1, calls, "should call invoker once")
	})

	t.Run("Retry", func(t *testing.T) {
		var calls int

		invoker := InterceptorInvoker(func(ctx context.Context, req *Request, dst any) error {
			calls++
			return &Error{Code: 429, Parameters: &ResponseParameters{RetryAfter: 1}}
		})

		var timeAfterCalls int

		interceptor := NewRetryFloodErrorInterceptor(
			WithRetryFloodErrorTries(3),
			WithRetryFloodErrorMaxRetryAfter(time.Second*2),
			WithRetryFloodErrorTimeAfter(func(time.Duration) <-chan time.Time {
				timeAfterCalls++
				result := make(chan time.Time, 1)
				result <- time.Now()
				return result
			}),
		)

		err := interceptor(context.Background(), &Request{}, nil, invoker)

		assert.Error(t, err, "should return error")
		assert.Equal(t, 3, calls, "should call invoker 3 times")
		assert.Equal(t, 3, timeAfterCalls, "should call timeAfter 3 times")
	})

	t.Run("MaxRetryAfter", func(t *testing.T) {
		var calls int

		invoker := InterceptorInvoker(func(ctx context.Context, req *Request, dst any) error {
			calls++
			return &Error{Code: 429, Parameters: &ResponseParameters{RetryAfter: 2}}
		})

		var timeAfterCalls int

		interceptor := NewRetryFloodErrorInterceptor(
			WithRetryFloodErrorTries(3),
			WithRetryFloodErrorMaxRetryAfter(time.Second),
			WithRetryFloodErrorTimeAfter(func(time.Duration) <-chan time.Time {
				timeAfterCalls++
				result := make(chan time.Time, 1)
				result <- time.Now()
				return result
			}),
		)

		err := interceptor(context.Background(), &Request{}, nil, invoker)

		assert.Error(t, err, "should return error")
		assert.Equal(t, 1, calls, "should call invoker once")
		assert.Equal(t, 0, timeAfterCalls, "should call timeAfter once")
	})

	t.Run("Timeout", func(t *testing.T) {
		var calls int

		invoker := InterceptorInvoker(func(ctx context.Context, req *Request, dst any) error {
			calls++
			return &Error{Code: 429, Parameters: &ResponseParameters{RetryAfter: 1}}
		})

		ctx, cancel := context.WithTimeout(context.Background(), time.Nanosecond)
		defer cancel()

		interceptor := NewRetryFloodErrorInterceptor(
			WithRetryFloodErrorTries(10),
			WithRetryFloodErrorMaxRetryAfter(time.Second*2),
		)

		err := interceptor(ctx, &Request{}, nil, invoker)

		assert.Error(t, err, "should return error")
		assert.Equal(t, 1, calls, "should call invoker once")
	})
}

func TestNewRetryInternalServerErrorInterceptor(t *testing.T) {
	t.Run("TestNoError", func(t *testing.T) {
		var calls int

		invoker := InterceptorInvoker(func(ctx context.Context, req *Request, dst any) error {
			calls++
			return nil
		})

		interceptor := NewRetryInternalServerErrorInterceptor()

		err := interceptor(context.Background(), &Request{}, nil, invoker)

		assert.NoError(t, err, "should no return error")
		assert.Equal(t, 1, calls, "should call invoker once")
	})

	t.Run("NoTgError", func(t *testing.T) {
		var calls int

		invoker := InterceptorInvoker(func(ctx context.Context, req *Request, dst any) error {
			calls++
			return errors.New("test")
		})

		interceptor := NewRetryInternalServerErrorInterceptor()

		err := interceptor(context.Background(), &Request{}, nil, invoker)

		assert.Error(t, err, "should return error")
		assert.Equal(t, 1, calls, "should call invoker once")
	})

	t.Run("Retry", func(t *testing.T) {
		var calls int

		invoker := InterceptorInvoker(func(ctx context.Context, req *Request, dst any) error {
			calls++
			return &Error{Code: 500}
		})

		var timeAfterCalls int

		interceptor := NewRetryInternalServerErrorInterceptor(
			WithRetryInternalServerErrorTries(3),
			WithRetryInternalServerErrorDelay(time.Millisecond),
			WithRetryInternalServerErrorTimeAfter(func(time.Duration) <-chan time.Time {
				defer func() { timeAfterCalls++ }()

				result := make(chan time.Time, 1)
				result <- time.Now()
				return result
			}),
		)

		err := interceptor(context.Background(), &Request{}, nil, invoker)

		assert.Error(t, err, "should return error")
		assert.Equal(t, 3, calls, "should call invoker 3 times")
		assert.Equal(t, 3, timeAfterCalls, "should call timeAfter 3 times")
	})

	t.Run("Timeout", func(t *testing.T) {
		var calls int

		invoker := InterceptorInvoker(func(ctx context.Context, req *Request, dst any) error {
			calls++
			return &Error{Code: 500}
		})

		ctx, cancel := context.WithTimeout(context.Background(), time.Nanosecond)
		defer cancel()

		interceptor := NewRetryInternalServerErrorInterceptor(
			WithRetryInternalServerErrorTries(10),
			WithRetryInternalServerErrorDelay(time.Millisecond),
		)

		err := interceptor(ctx, &Request{}, nil, invoker)

		assert.Error(t, err, "should return error")
		assert.Equal(t, 1, calls, "should call invoker once")
	})
}
