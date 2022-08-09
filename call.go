package tg

import (
	"context"
)

type Call[T any] struct {
	client  *Client
	request *Request
}

// Request returns a low-level request object for making API calls.
func (call *Call[T]) Request() *Request {
	return call.request
}

func (call *Call[T]) MarshalJSON() ([]byte, error) {
	return call.request.MarshalJSON()
}

func (call *Call[T]) Bind(client *Client) {
	call.client = client
}

func (call *Call[T]) Do(ctx context.Context) (result T, err error) {
	if err := call.client.Do(ctx, call.request, &result); err != nil {
		return result, err
	}

	return
}

func (call *Call[T]) DoVoid(ctx context.Context) (err error) {
	return call.client.Do(ctx, call.request, nil)
}

// BindClient binds Client to the Call.
// It's useful for chaining calls.
//
//	return tg.BindClient(tg.NewGetMeCall(), client).DoVoid(ctx)
func BindClient[C interface {
	Bind(client *Client)
}](
	call C,
	client *Client,
) C {
	call.Bind(client)
	return call
}

type CallNoResult struct {
	client  *Client
	request *Request
}

func (call *CallNoResult) Request() *Request {
	return call.request
}

func (call *CallNoResult) MarshalJSON() ([]byte, error) {
	return call.request.MarshalJSON()
}

func (call *CallNoResult) Bind(client *Client) {
	call.client = client
}

func (call *CallNoResult) DoVoid(ctx context.Context) (err error) {
	return call.client.Do(ctx, call.request, nil)
}
