package tg

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
)

type Doer interface {
	Do(r *http.Request) (*http.Response, error)
}

// Client is Telegram Bot API client structure.
// Create new client with NewClient function.
type Client struct {
	// bot api token
	token string

	// bot api server base url,
	// default values is https://api.telegram.org
	server string

	callURL     string
	downloadURL string

	// http client,
	// default values is http.DefaultClient
	doer Doer

	// contains cached bot info
	me     *User
	meLock sync.Mutex

	interceptors []Interceptor
	invoker      InterceptorInvoker
}

// ClientOption is a function that sets some option for Client.
type ClientOption func(*Client)

// WithClientServerURL sets custom server url for Client.
func WithClientServerURL(server string) ClientOption {
	return func(c *Client) {
		c.server = server
	}
}

// WithClientDoer sets custom http client for Client.
func WithClientDoer(doer Doer) ClientOption {
	return func(c *Client) {
		c.doer = doer
	}
}

// WithClientTestEnv switches bot to test environment.
// See https://core.telegram.org/bots/webapps#using-bots-in-the-test-environment
func WithClientTestEnv() ClientOption {
	return func(c *Client) {
		c.callURL = "%s/bot%s/test/%s"
	}
}

// WithClientInterceptor adds interceptor to client.
func WithClientInterceptors(ints ...Interceptor) ClientOption {
	return func(c *Client) {
		c.interceptors = append(c.interceptors, ints...)
	}
}

// New creates new Client with given token and options.
func New(token string, options ...ClientOption) *Client {
	c := &Client{
		token:  token,
		server: "https://api.telegram.org",

		callURL:     "%s/bot%s/%s",
		downloadURL: "%s/file/bot%s/%s",

		doer: http.DefaultClient,
	}

	for _, option := range options {
		option(c)
	}

	c.invoker = c.buildInvoker()

	return c
}

func (client *Client) buildInvoker() InterceptorInvoker {
	invoker := client.invoke

	for i := len(client.interceptors) - 1; i >= 0; i-- {
		invoker = func(next InterceptorInvoker, interceptor Interceptor) InterceptorInvoker {
			return func(ctx context.Context, req *Request, dst any) error {
				return interceptor(ctx, req, dst, next)
			}
		}(invoker, client.interceptors[i])
	}

	return invoker
}

func (client *Client) Token() string {
	return client.token
}

// Execute request at low-level
func (client *Client) execute(ctx context.Context, r *Request) (*Response, error) {
	if len(r.files) > 0 {
		return client.executeStreaming(
			ctx,
			func(w io.Writer) httpEncoder { return newMultipartEncoder(w) },
			r,
		)
	}

	return client.executeSimple(
		ctx,
		func(w io.Writer) httpEncoder { return newURLEncodedEncoder(w) },
		r,
	)
}

func (client *Client) buildCallURL(token, method string) string {
	return fmt.Sprintf(client.callURL, client.server, token, method)
}

func (client *Client) buildDownloadURL(token, path string) string {
	return fmt.Sprintf(client.downloadURL, client.server, token, path)
}

func (client *Client) buildHTTPRequest(
	r *Request,
	body io.Reader,
	contentType string,
) (*http.Request, error) {
	url := client.buildCallURL(client.token, r.Method)

	req, err := http.NewRequest(http.MethodPost, url, body)
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}

	// set content type
	req.Header.Set("Content-Type", contentType)

	return req, nil
}

func (client *Client) executeSimple(
	ctx context.Context,
	newEncoder func(io.Writer) httpEncoder,
	r *Request,
) (*Response, error) {
	buf := &bytes.Buffer{}

	encoder := newEncoder(buf)

	if err := r.Encode(encoder); err != nil {
		return nil, fmt.Errorf("encode: %w", err)
	}

	if err := encoder.Close(); err != nil {
		return nil, fmt.Errorf("encoder close: %w", err)
	}

	req, err := client.buildHTTPRequest(
		r,
		buf,
		encoder.ContentType(),
	)
	if err != nil {
		return nil, fmt.Errorf("build http request: %w", err)
	}

	res, err := client.executeHTTPRequest(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("execute http request: %w", err)
	}

	return res, nil
}

func (client *Client) executeHTTPRequest(ctx context.Context, r *http.Request) (*Response, error) {
	r = r.WithContext(ctx)

	// execute request
	res, err := client.doer.Do(r)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer res.Body.Close()

	// TODO: handle status and content type

	// read content
	content, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	response := &Response{
		StatusCode: res.StatusCode,
	}

	// unmarshal content
	if err := json.Unmarshal(content, &response); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	return response, nil
}

func (client *Client) executeStreaming(
	ctx context.Context,
	newEncoder func(io.Writer) httpEncoder,
	r *Request,
) (*Response, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	pr, pw := io.Pipe()

	encoder := newEncoder(pw)

	resChan := make(chan *Response)
	errChan := make(chan error)

	// upload
	go func() {
		defer pw.Close()
		defer encoder.Close()

		if err := r.Encode(encoder); err != nil {
			errChan <- err
		}
	}()

	// send
	go func() {
		req, err := client.buildHTTPRequest(r, pr, encoder.ContentType())
		if err != nil {
			errChan <- fmt.Errorf("build http request: %w", err)
			return
		}

		res, err := client.executeHTTPRequest(ctx, req)
		if err != nil {
			errChan <- fmt.Errorf("execute http request: %w", err)
			return
		}

		resChan <- res
	}()

	select {
	case err := <-errChan:
		return nil, err
	case res := <-resChan:
		return res, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (client *Client) invoke(ctx context.Context, req *Request, dst any) error {
	res, err := client.execute(ctx, req)
	if err != nil {
		return fmt.Errorf("execute: %w", err)
	}

	if !res.Ok {
		return &Error{
			Code:       res.ErrorCode,
			Message:    res.Description,
			Parameters: res.Parameters,
		}
	}

	if dst != nil {
		if err := json.Unmarshal(res.Result, dst); err != nil {
			return fmt.Errorf("unmarshal: %w", err)
		}
	}

	return nil
}

func (client *Client) Do(ctx context.Context, req *Request, dst interface{}) error {
	return client.invoker(ctx, req, dst)
}

// Download file by path from Client.GetFile method.
// Don't forget to close ReadCloser.
func (client *Client) Download(ctx context.Context, path string) (io.ReadCloser, error) {
	url := client.buildDownloadURL(client.token, path)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}

	res, err := client.doer.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}

	if res.StatusCode != http.StatusOK {
		defer res.Body.Close()

		tgResponse := &Response{}

		if err := json.NewDecoder(res.Body).Decode(tgResponse); err != nil {
			return nil, fmt.Errorf("unmarshal: %w", err)
		}

		return nil, &Error{
			Code:       tgResponse.ErrorCode,
			Message:    tgResponse.Description,
			Parameters: tgResponse.Parameters,
		}
	}

	return res.Body, nil
}
