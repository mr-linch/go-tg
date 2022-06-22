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
)

type Webhook struct {
	url            string
	handler        Handler
	client         *tg.Client
	ip             string
	maxConnections int

	securitySubnets []netip.Prefix
	securityToken   string
}

var DefaultSubnets = []netip.Prefix{
	netip.MustParsePrefix("149.154.160.0/20"),
	netip.MustParsePrefix("91.108.4.0/22"),
}

type WebhookOption func(*Webhook)

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
		url:     url,
		handler: handler,
		client:  client,

		securitySubnets: DefaultSubnets,
		securityToken:   token,
	}

	for _, option := range options {
		option(webhook)
	}

	return webhook
}

func (webhook *Webhook) Setup(ctx context.Context, dropPendingUpdates bool) error {
	info, err := webhook.client.GetWebhookInfo().Do(ctx)
	if err != nil {
		return fmt.Errorf("get webhook info: %w", err)
	}

	if info.URL == webhook.url &&
		info.MaxConnections == webhook.maxConnections &&
		(webhook.ip == "" || info.IPAddress == webhook.ip) {
		return nil
	}

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

	return setWebhookCall.Do(ctx)
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
	update := &tg.Update{}
	if err := json.Unmarshal(body, update); err != nil {
		http.Error(w, "failed to parse body", http.StatusBadRequest)
		return
	}

	update.Bind(webhook.client)

	// handle update
	if err := webhook.handler.Handle(r.Context(), update); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := update.Response()

	if response != nil {
		body, err := json.Marshal(response)
		log.Printf("response %s", string(body))
		if err != nil {
			log.Printf("failed to marshal response: %s", err)
		}
		_, err = w.Write(body)
		if err != nil {
			log.Printf("failed to write response: %s", err)
		}
	}
}
