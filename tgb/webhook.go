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

type WebhookOption func(*Webhook)

func WithDropPendingUpdates(dropPendingUpdates bool) WebhookOption {
	return func(webhook *Webhook) {
		webhook.dropPendingUpdates = dropPendingUpdates
	}
}

func WithIP(ip string) WebhookOption {
	return func(webhook *Webhook) {
		webhook.ip = ip
	}
}

func WithWebhookSecuritySubnets(subnets ...netip.Prefix) WebhookOption {
	return func(webhook *Webhook) {
		webhook.securitySubnets = subnets
	}
}

func WithWebhookSecurityToken(token string) WebhookOption {
	return func(webhook *Webhook) {
		webhook.securityToken = token
	}
}

func WithMaxConnections(maxConnections int) WebhookOption {
	return func(webhook *Webhook) {
		webhook.maxConnections = maxConnections
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

func (webhook *Webhook) Setup(ctx context.Context, dropPendingUpdates bool) error {
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
	update := tg.NewUpdateWebhook(webhook.client)
	if err := json.Unmarshal(body, update); err != nil {
		http.Error(w, "failed to parse body", http.StatusBadRequest)
		return
	}

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
