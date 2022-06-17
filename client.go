package tg

import "net/http"

// Client is Telegram Bot API client structure.
// Create new client with NewClient function.
type Client struct {
	// bot api token
	token string

	// bot api server base url,
	// default values is https://api.telegram.org
	server string

	// http client,
	// default values is http.DefaultClient
	doer *http.Client
}

// ClientOption is a function that sets some option for Client.
type ClientOption func(*Client)

// WithServer sets custom server url for Client.
func WithServer(server string) ClientOption {
	return func(c *Client) {
		c.server = server
	}
}

// WithDoer sets custom http client for Client.
func WithDoer(doer *http.Client) ClientOption {
	return func(c *Client) {
		c.doer = doer
	}
}

// New creates new Client with given token and options.
func New(token string, options ...ClientOption) *Client {
	c := &Client{
		token:  token,
		server: "https://api.telegram.org",
		doer:   http.DefaultClient,
	}

	for _, option := range options {
		option(c)
	}

	return c
}
