package tg

import (
	"context"
	"fmt"
)

type Call[T any] struct {
	client  *Client
	request *Request
}

func (call *Call[T]) Bind(client *Client) {
	call.client = client
}

func (call *Call[T]) Do(ctx context.Context) (result *T, err error) {
	if err := call.client.Invoke(ctx, call.request, &result); err != nil {
		return nil, err
	}

	return
}

func callWithClient[B interface {
	Bind(client *Client)
}](
	client *Client,
	b B,
) B {
	b.Bind(client)
	return b
}

type CallNoResult struct {
	client  *Client
	request *Request
}

func (call *CallNoResult) Bind(client *Client) {
	call.client = client
}

func (call *CallNoResult) Do(ctx context.Context) (err error) {
	var result bool

	if err := call.client.Invoke(ctx, call.request, &result); err != nil {
		return err
	}

	if !result {
		return fmt.Errorf("call returns not True")
	}

	return
}
