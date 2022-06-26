package tgb

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/netip"

	tg "github.com/mr-linch/go-tg"
	"github.com/tomasen/realip"
	"golang.org/x/exp/slices"
)

type Webhook struct {
	url                string
	handler            Handler
	client             *tg.Client
	ip                 string
	maxConnections     int
	dropPendingUpdates bool
	allowedUpdates     []string

	securitySubnets []netip.Prefix
	securityToken   string
}

var defaultSubnets = []netip.Prefix{
	netip.MustParsePrefix("149.154.160.0/20"),
	netip.MustParsePrefix("91.108.4.0/22"),
}

// WebhookOption used to configure the Webhook
type WebhookOption func(*Webhook)

// WithDropPendingUpdates drop pending updates (if pending > 0 only)
func WithDropPendingUpdates(dropPendingUpdates bool) WebhookOption {
	return func(webhook *Webhook) {
		webhook.dropPendingUpdates = dropPendingUpdates
	}
}

// WithWebhookIP the fixed IP address which will be used to
// send webhook requests instead of the IP address resolved through DNS
func WithWebhookIP(ip string) WebhookOption {
	return func(webhook *Webhook) {
		webhook.ip = ip
	}
}

// WithWebhookSecuritySubnets sets list of subnets which are allowed to send webhook requests.
func WithWebhookSecuritySubnets(subnets ...netip.Prefix) WebhookOption {
	return func(webhook *Webhook) {
		webhook.securitySubnets = subnets
	}
}

// WithWebhookSecurityToken sets the security token which is used to validate the webhook requests.
// By default the token is generated from the client token via sha256.
// 1-256 characters. Only characters A-Z, a-z, 0-9, _ and - are allowed.
// The header is useful to ensure that the request comes from a webhook set by you.
func WithWebhookSecurityToken(token string) WebhookOption {
	return func(webhook *Webhook) {
		webhook.securityToken = token
	}
}

// WithWebhookMaxConnections sets the maximum number of concurrent connections.
// By default is 40
func WithWebhookMaxConnections(maxConnections int) WebhookOption {
	return func(webhook *Webhook) {
		webhook.maxConnections = maxConnections
	}
}

// WithWebhookAllowedUpdates sets the list of allowed updates.
// By default all update types except chat_member (default).
// If not specified, the previous setting will be used.
// Please note that this parameter doesn't affect updates created before the call to the setWebhook,
// so unwanted updates may be received for a short period of time.
func WithWebhookAllowedUpdates(updates ...string) WebhookOption {
	return func(webhook *Webhook) {
		webhook.allowedUpdates = updates
	}
}

func NewWebhook(url string, handler Handler, client *tg.Client, options ...WebhookOption) *Webhook {
	securityToken := sha256.Sum256([]byte(client.Token()))
	token := fmt.Sprintf("%x", securityToken)

	webhook := &Webhook{
		url:            url,
		handler:        handler,
		client:         client,
		maxConnections: defaultMaxConnections,

		dropPendingUpdates: false,

		allowedUpdates:  []string{},
		securitySubnets: defaultSubnets,
		securityToken:   token,
	}

	for _, option := range options {
		option(webhook)
	}

	return webhook
}

const defaultMaxConnections = 40

func (webhook *Webhook) Setup(ctx context.Context) error {
	info, err := webhook.client.GetWebhookInfo().Do(ctx)
	if err != nil {
		return fmt.Errorf("get webhook info: %w", err)
	}

	if info.URL != webhook.url ||
		info.MaxConnections != webhook.maxConnections ||
		(len(info.AllowedUpdates) > 0 && !slices.Equal(info.AllowedUpdates, webhook.allowedUpdates)) ||
		(webhook.ip != "" && info.IPAddress != webhook.ip) ||
		(info.PendingUpdateCount > 0 && webhook.dropPendingUpdates) {

		setWebhookCall := webhook.client.SetWebhook(webhook.url)

		if webhook.maxConnections > 0 {
			setWebhookCall = setWebhookCall.MaxConnections(webhook.maxConnections)
		}

		if webhook.ip != "" {
			setWebhookCall = setWebhookCall.IpAddress(webhook.ip)
		}

		if webhook.securityToken != "" {
			setWebhookCall = setWebhookCall.SecretToken(webhook.securityToken)
		}

		if webhook.dropPendingUpdates {
			setWebhookCall = setWebhookCall.DropPendingUpdates(true)
		}

		if webhook.allowedUpdates != nil {
			setWebhookCall = setWebhookCall.AllowedUpdates(webhook.allowedUpdates)
		}

		return setWebhookCall.Do(ctx)
	}

	return nil

}

func (webhook *Webhook) isAllowedIP(ip netip.Addr) bool {
	if len(webhook.securitySubnets) == 0 {
		return true
	}

	for _, net := range webhook.securitySubnets {
		if net.Contains(ip) {
			return true
		}
	}

	return false
}

const securityTokenHeader = "X-Telegram-Bot-Api-Secret-Token"

func (webhook *Webhook) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// check method
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// check IP
	ip, err := netip.ParseAddr(realip.FromRequest(r))
	if err != nil {
		http.Error(w, "invalid ip", http.StatusBadRequest)
		return
	}

	if !webhook.isAllowedIP(ip) {
		http.Error(w, "security check failed", http.StatusUnauthorized)
		return
	}

	// check token
	if webhook.securityToken != "" {
		if r.Header.Get(securityTokenHeader) != webhook.securityToken {
			http.Error(w, "security check failed", http.StatusUnauthorized)
			return
		}
	}

	// check content type
	if r.Header.Get("Content-Type") != "application/json" {
		http.Error(w, "content type not supported", http.StatusUnsupportedMediaType)
		return
	}

	// read body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read body", http.StatusBadRequest)
		return
	}

	// parse body
	update := tg.NewUpdateWebhook()
	if err := json.Unmarshal(body, update); err != nil {
		http.Error(w, "failed to parse body", http.StatusBadRequest)
		return
	}

	update.Bind(webhook.client)

	// handle update
	if err := webhook.handler.Handle(r.Context(), update); err != nil {
		log.Printf("failed to handle update: %v", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := update.Response()

	if response != nil {
		body, err := json.Marshal(response)

		if err != nil {
			log.Printf("failed to marshal response: %s", err)
			return
		}

		_, err = w.Write(body)
		if err != nil {
			log.Printf("failed to write response: %s", err)
			return
		}
	}
}
