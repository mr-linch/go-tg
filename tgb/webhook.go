package tgb

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/netip"
	"time"

	tg "github.com/mr-linch/go-tg"
	"github.com/tomasen/realip"
	"golang.org/x/exp/slices"
)

type Webhook struct {
	url     string
	handler Handler
	client  *tg.Client
	logger  Logger

	ip                 string
	maxConnections     int
	dropPendingUpdates bool
	allowedUpdates     []tg.UpdateType

	securitySubnets []netip.Prefix
	securityToken   string

	isSetup bool
}

var defaultSubnets = []netip.Prefix{
	netip.MustParsePrefix("149.154.160.0/20"),
	netip.MustParsePrefix("91.108.4.0/22"),
}

// WebhookOption used to configure the Webhook
type WebhookOption func(*Webhook)

// WithWebhookLogger sets the logger which will be used to log the webhook related errors.
func WithWebhookLogger(logger Logger) WebhookOption {
	return func(webhook *Webhook) {
		webhook.logger = logger
	}
}

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
func WithWebhookAllowedUpdates(updates ...tg.UpdateType) WebhookOption {
	return func(webhook *Webhook) {
		webhook.allowedUpdates = updates
	}
}

func NewWebhook(handler Handler, client *tg.Client, url string, options ...WebhookOption) *Webhook {
	securityToken := sha256.Sum256([]byte(client.Token()))
	token := hex.EncodeToString(securityToken[:])

	webhook := &Webhook{
		url:            url,
		handler:        handler,
		client:         client,
		maxConnections: defaultMaxConnections,

		dropPendingUpdates: false,

		allowedUpdates:  []tg.UpdateType{},
		securitySubnets: defaultSubnets,
		securityToken:   token,
	}

	for _, option := range options {
		option(webhook)
	}

	return webhook
}

func (webhook *Webhook) log(format string, args ...any) {
	if webhook.logger != nil {
		webhook.logger.Printf("tgb.Webhook: "+format, args...)
	}
}

const defaultMaxConnections = 40

func (webhook *Webhook) Setup(ctx context.Context) (err error) {
	defer func() {
		webhook.isSetup = err == nil
	}()

	info, err := webhook.client.GetWebhookInfo().Do(ctx)
	if err != nil {
		return fmt.Errorf("get webhook info: %w", err)
	}

	if info.URL != webhook.url ||
		info.MaxConnections != webhook.maxConnections ||
		(len(info.AllowedUpdates) > 0 && !slices.Equal(info.AllowedUpdates, webhook.allowedUpdates)) ||
		(webhook.ip != "" && info.IPAddress != webhook.ip) ||
		(info.PendingUpdateCount > 0 && webhook.dropPendingUpdates) {

		webhook.log("current webhook config is outdated, updating...")

		setWebhookCall := webhook.client.SetWebhook(webhook.url)

		if webhook.maxConnections > 0 {
			setWebhookCall = setWebhookCall.MaxConnections(webhook.maxConnections)
		}

		if webhook.ip != "" {
			setWebhookCall = setWebhookCall.IPAddress(webhook.ip)
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

		return setWebhookCall.DoVoid(ctx)
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
		webhook.log("http error: method not allowed: %s", r.Method)
		return
	}

	// check IP
	ip, err := netip.ParseAddr(realip.FromRequest(r))
	if err != nil {
		http.Error(w, "invalid ip", http.StatusBadRequest)
		webhook.log("http error: invalid ip: %v", err)
		return
	}

	if !webhook.isAllowedIP(ip) {
		http.Error(w, "security check failed", http.StatusUnauthorized)
		webhook.log("http error: security check failed: ip '%s' is not allowed", ip)
		return
	}

	// check token
	if webhook.securityToken != "" {
		if securityToken := r.Header.Get(securityTokenHeader); securityToken != webhook.securityToken {
			http.Error(w, "security check failed", http.StatusUnauthorized)
			webhook.log("http error: security check failed: token '%s' is not allowed", securityToken)
			return
		}
	}

	// check content type
	if r.Header.Get("Content-Type") != "application/json" {
		http.Error(w, "content type not supported", http.StatusUnsupportedMediaType)
		webhook.log("http error: content type not supported: %s", r.Header.Get("Content-Type"))
		return
	}

	// read body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read body", http.StatusBadRequest)
		webhook.log("http error: read body failed: %v", err)
		return
	}

	// parse body
	baseUpdate := &tg.Update{}
	if err := json.Unmarshal(body, &baseUpdate); err != nil {
		http.Error(w, "failed to parse body", http.StatusBadRequest)
		webhook.log("http error: parse body failed: %v", err)
		return
	}

	responseChan := make(chan json.Marshaler)

	update := &Update{
		webhookResponse: responseChan,
		Update:          baseUpdate,
		Client:          webhook.client,
	}

	handlerDoneChan := make(chan struct{})

	var handlerClose context.CancelFunc
	go func() {
		handlerCtx := context.Background()

		handlerCtx, handlerClose = context.WithCancel(handlerCtx)
		defer handlerClose()

		// handle update
		if err := webhook.handler.Handle(handlerCtx, update); err != nil {
			webhook.log("handler error: %v", err)
		}

		handlerDoneChan <- struct{}{}
	}()

	select {
	case <-r.Context().Done():
		webhook.log("shutdown...")
		handlerClose()
		return
	case response := <-responseChan:
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if response != nil {
			body, err := json.Marshal(response)

			if err != nil {
				webhook.log("marshal webhook response error: %v", err)
				return
			}

			_, err = w.Write(body)
			if err != nil {
				webhook.log("write webhook response error: %v", err)
				return
			}
		}
		return
	case <-handlerDoneChan:
		w.WriteHeader(http.StatusOK)
	}
}

// Run starts the webhook server.
func (webhook *Webhook) Run(ctx context.Context, listen string) error {
	if !webhook.isSetup {
		if err := webhook.Setup(ctx); err != nil {
			return fmt.Errorf("setup webhook: %v", err)
		}
	}

	server := &http.Server{
		Addr:    listen,
		Handler: webhook,
	}

	go func() {
		<-ctx.Done()

		webhook.log("shutdown server...")

		closeCtx, close := context.WithTimeout(context.Background(), 10*time.Second)
		defer close()

		if err := server.Shutdown(closeCtx); err != nil {
			webhook.log("server shutdown error: %v", err)
		}
	}()

	webhook.log("starting webhook server on %s", listen)

	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		return fmt.Errorf("server error: %v", err)
	}

	return nil
}
