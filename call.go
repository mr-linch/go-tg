package tg

import (
	"context"
	"encoding/json"
	"fmt"

	"golang.org/x/exp/maps"
)

type Call[T any] struct {
	client  *Client
	request *Request
}

func (call *Call[T]) MarshalJSON() ([]byte, error) {
	args := make(map[string]string, len(call.request.args)+1)

	args["method"] = call.request.Method

	maps.Copy(args, call.request.args)

	return json.Marshal(args)
}

func (call *Call[T]) Bind(client *Client) {
	call.client = client
}

func (call *Call[T]) Do(ctx context.Context) (result T, err error) {
	if err := call.client.Invoke(ctx, call.request, &result); err != nil {
		return result, err
	}

	return
}

func (call *Call[T]) DoVoid(ctx context.Context) (err error) {
	return call.client.Invoke(ctx, call.request, nil)
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

func (call *CallNoResult) MarshalJSON() ([]byte, error) {
	args := make(map[string]string, len(call.request.args)+1)
	args["method"] = call.request.Method
	maps.Copy(args, call.request.args)
	return json.Marshal(args)
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

func (call *CallNoResult) DoVoid(ctx context.Context) (err error) {
	return call.client.Invoke(ctx, call.request, nil)
}
