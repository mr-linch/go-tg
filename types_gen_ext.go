package tg

import "encoding/json"

//go:generate go run github.com/mr-linch/go-tg-gen@latest -types-output types_gen.go

func (update *Update) Client() *Client {
	return update.client
}

func (update *Update) Bind(client *Client) {
	update.client = client
}

func (update *Update) Respond(v json.Marshaler) {
	// TODO: add fallback to direct call if update is long polling
	update.response = v
}

func (update *Update) Response() json.Marshaler {
	return update.response
}

type InputMedia struct {
}

type InlineQueryResult struct {
}

type BotCommandScope struct {
}

type ReplyMarkup struct{}
