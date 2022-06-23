package tg

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
)

type ChatID int64

var _ PeerID = (ChatID)(0)

func (id ChatID) PeerID() string {
	return strconv.FormatInt(int64(id), 10)
}

type ChatType int8

const (
	ChatTypePrivate ChatType = iota + 1
	ChatTypeGroup
	ChatTypeSupergroup
	ChatTypeChannel
)

func (chatType ChatType) String() string {
	if chatType < ChatTypePrivate || chatType > ChatTypeChannel {
		return "unknown"
	}

	return [...]string{"private", "group", "supergroup", "channel"}[chatType-1]
}

func (chatType ChatType) MarshalJSON() ([]byte, error) {
	return json.Marshal(chatType.String())
}

func (chatType *ChatType) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	switch s {
	case "private":
		*chatType = ChatTypePrivate
	case "group":
		*chatType = ChatTypeGroup
	case "supergroup":
		*chatType = ChatTypeSupergroup
	case "channel":
		*chatType = ChatTypeChannel
	default:
		return fmt.Errorf("unknown chat type: %s", s)
	}

	return nil
}

// UserID it's unique identifier for Telegram user or bot.
type UserID int64

var _ PeerID = (UserID)(0)

func (id UserID) PeerID() string {
	return strconv.FormatInt(int64(id), 10)
}

type Username string

func (un Username) PeerID() string {
	return "@" + string(un)
}

type PeerID interface {
	PeerID() string
}

// MessageID it's unique identifier for a message in a chat.
type MessageID int

type FileID string

type FileArg struct {
	FileID FileID
	Upload InputFile
}

//go:generate go run github.com/mr-linch/go-tg-gen@latest -types-output types_gen.go

func (chat Chat) PeerID() string {
	return chat.ID.PeerID()
}

func (update *Update) Client() *Client {
	return update.client
}

func (update *Update) Bind(client *Client) {
	update.client = client
}

type UpdateRespond interface {
	json.Marshaler
	DoNoResult(ctx context.Context) error
	Bind(client *Client)
}

func NewUpdateWebhook(client *Client) *Update {
	return &Update{
		client:    client,
		isWebhook: true,
	}
}

func (update *Update) Respond(ctx context.Context, v UpdateRespond) error {
	if update.isWebhook && update.response == nil {
		update.response = v
		return nil
	}

	v.Bind(update.client)

	return v.DoNoResult(ctx)
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
