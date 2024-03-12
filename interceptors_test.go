package tg

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewInterceptorRetryFloodError(t *testing.T) {
	t.Run("TestNoError", func(t *testing.T) {
		var calls int

		invoker := InterceptorInvoker(func(ctx context.Context, req *Request, dst any) error {
			calls++
			return nil
		})

		interceptor := NewInterceptorRetryFloodError()

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

		interceptor := NewInterceptorRetryFloodError()

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

		interceptor := NewInterceptorRetryFloodError(
			WithInterceptorRetryFloodErrorTries(3),
			WithInterceptorRetryFloodErrorMaxRetryAfter(time.Second*2),
			WithInterceptorRetryFloodErrorTimeAfter(func(time.Duration) <-chan time.Time {
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

		interceptor := NewInterceptorRetryFloodError(
			WithInterceptorRetryFloodErrorTries(3),
			WithInterceptorRetryFloodErrorMaxRetryAfter(time.Second),
			WithInterceptorRetryFloodErrorTimeAfter(func(time.Duration) <-chan time.Time {
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

		interceptor := NewInterceptorRetryFloodError(
			WithInterceptorRetryFloodErrorTries(10),
			WithInterceptorRetryFloodErrorMaxRetryAfter(time.Second*2),
		)

		err := interceptor(ctx, &Request{}, nil, invoker)

		assert.Error(t, err, "should return error")
		assert.Equal(t, 1, calls, "should call invoker once")
	})
}

func TestNewInterceptorRetryInternalServerError(t *testing.T) {
	t.Run("TestNoError", func(t *testing.T) {
		var calls int

		invoker := InterceptorInvoker(func(ctx context.Context, req *Request, dst any) error {
			calls++
			return nil
		})

		interceptor := NewInterceptorRetryInternalServerError()

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

		interceptor := NewInterceptorRetryInternalServerError()

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

		interceptor := NewInterceptorRetryInternalServerError(
			WithInterceptorRetryInternalServerErrorTries(3),
			WithInterceptorRetryInternalServerErrorDelay(time.Millisecond),
			WithInterceptorRetryInternalServerErrorTimeAfter(func(time.Duration) <-chan time.Time {
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

		interceptor := NewInterceptorRetryInternalServerError(
			WithInterceptorRetryInternalServerErrorTries(10),
			WithInterceptorRetryInternalServerErrorDelay(time.Millisecond),
		)

		err := interceptor(ctx, &Request{}, nil, invoker)

		assert.Error(t, err, "should return error")
		assert.Equal(t, 1, calls, "should call invoker once")
	})
}

func TestNewInterceptorMethodFilter(t *testing.T) {
	t.Run("InWhitelist", func(t *testing.T) {
		req := NewRequest("sendMessage")

		var calls int

		interceptor := Interceptor(func(ctx context.Context, req *Request, dst any, invoker InterceptorInvoker) error {
			calls++
			return invoker(ctx, req, dst)
		})

		interceptor = NewInterceptorMethodFilter(interceptor, "sendMessage")

		err := interceptor(context.Background(), req, nil, InterceptorInvoker(func(ctx context.Context, req *Request, dst any) error {
			return nil
		}))

		assert.NoError(t, err, "should no return error")
		assert.Equal(t, 1, calls, "should call invoker once")
	})

	t.Run("NotInWhitelist", func(t *testing.T) {
		req := NewRequest("editMessageText")

		var calls int

		interceptor := Interceptor(func(ctx context.Context, req *Request, dst any, invoker InterceptorInvoker) error {
			calls++
			return invoker(ctx, req, dst)
		})

		interceptor = NewInterceptorMethodFilter(interceptor, "sendMessage")

		err := interceptor(context.Background(), req, nil, InterceptorInvoker(func(ctx context.Context, req *Request, dst any) error {
			return nil
		}))

		assert.NoError(t, err, "should no return error")
		assert.Equal(t, 0, calls, "should call invoker once")
	})
}

func TestNewInterceptorDefaultParseMethod(t *testing.T) {
	t.Run("Ok", func(t *testing.T) {
		req := NewRequest("sendMessage")
		dst := &Response{}

		var calls int

		invoker := InterceptorInvoker(func(ctx context.Context, req *Request, dst any) error {
			calls++
			assert.Equal(t, HTML.String(), req.args["parse_mode"], "should set parse_mode to HTML")
			return nil
		})

		interceptor := NewInterceptorDefaultParseMethod(HTML)

		err := interceptor(context.Background(), req, dst, invoker)

		assert.NoError(t, err, "should no return error")
		assert.Equal(t, 1, calls, "should call invoker once")
	})
}
