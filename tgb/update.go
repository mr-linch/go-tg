package tgb

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/mr-linch/go-tg"
)

// Update wraps around a tg.Update.
// Also contains Client which is used to send responses.
type Update struct {
	*tg.Update
	Client *tg.Client

	webhookReplyLock sync.Mutex
	webhookReply     chan json.Marshaler
	webhookReplySent bool
}

func newUpdateWebhook(update *tg.Update, client *tg.Client) *Update {
	return &Update{
		Update: update,
		Client: client,

		webhookReply:     make(chan json.Marshaler),
		webhookReplySent: false,
	}
}

// UpdateReply defines interface for responding to an update via Webhook.
type UpdateReply interface {
	json.Marshaler
	DoVoid(ctx context.Context) error
	Bind(client *tg.Client)
}

// Deprecated: use UpdateReply instead.
type UpdateRespond = UpdateReply

// Reply to Webhook, if possible or make usual call via Client.
func (update *Update) Reply(ctx context.Context, v UpdateReply) error {
	update.webhookReplyLock.Lock()
	defer update.webhookReplyLock.Unlock()

	if update.webhookReply != nil && !update.webhookReplySent {
		update.webhookReplySent = true
		update.webhookReply <- v
		return nil
	}

	return tg.BindClient(v, update.Client).DoVoid(ctx)
}

// Deprecated: use Reply instead.
func (update *Update) Respond(ctx context.Context, v UpdateRespond) error {
	return update.Reply(ctx, v)
}

func (update *Update) disableWebhookReply() {
	update.webhookReplyLock.Lock()
	defer update.webhookReplyLock.Unlock()

	update.webhookReplySent = true
}
