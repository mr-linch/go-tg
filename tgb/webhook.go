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

// Webhook is a Telegram bot webhook handler.
// It handles incoming webhook requests and calls the handler function.
// It implements the http.Handler interface, but can be adapted to any other handlers, see [Webhook.ServeRequest].
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

	ipFromRequestFunc func(r *http.Request) string

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

// WithWebhookRequestIP sets function to get the IP address from the request.
// By default the IP address is resolved through the X-Real-Ip and X-Forwarded-For headers.
func WithWebhookRequestIP(ip func(r *http.Request) string) WebhookOption {
	return func(webhook *Webhook) {
		webhook.ipFromRequestFunc = ip
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

		ipFromRequestFunc: realip.FromRequest,
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

// WebhookRequest is the request received from Telegram.
type WebhookRequest struct {
	Method      string
	ContentType string

	IP            netip.Addr
	SecurityToken string

	Body io.Reader
}

// WebhookResponse is the response to be sent to WebhookRequest.
type WebhookResponse struct {
	Status      int
	ContentType string
	Body        []byte
}

func (webhook *Webhook) checkRequest(r *WebhookRequest) *WebhookResponse {
	if r.Method != http.MethodPost {
		webhook.log("request with method '%s' was refused", r.Method)
		return &WebhookResponse{
			Status:      http.StatusMethodNotAllowed,
			ContentType: "text/plain",
			Body:        []byte("method not allowed"),
		}
	}

	if r.ContentType != "application/json" {
		webhook.log("invalid content type '%s'", r.ContentType)
		return &WebhookResponse{
			Status:      http.StatusUnsupportedMediaType,
			ContentType: "text/plain",
			Body:        []byte("unsupported media type"),
		}
	}

	if !webhook.isAllowedIP(r.IP) {
		webhook.log("request from '%s' was refused", r.IP)
		return &WebhookResponse{
			Status:      http.StatusForbidden,
			ContentType: "text/plain",
			Body:        []byte("security check failed"),
		}
	}

	if webhook.securityToken != "" && r.SecurityToken != webhook.securityToken {
		webhook.log("request with token '%s' was refused", r.SecurityToken)
		return &WebhookResponse{
			Status:      http.StatusForbidden,
			ContentType: "text/plain",
			Body:        []byte("security check failed"),
		}
	}

	return nil
}

// ServeRequest is the generic for webhook requests from Telegram.
func (webhook *Webhook) ServeRequest(ctx context.Context, r *WebhookRequest) *WebhookResponse {
	if response := webhook.checkRequest(r); response != nil {
		return response
	}

	// parse body
	baseUpdate := &tg.Update{}
	if err := json.NewDecoder(r.Body).Decode(baseUpdate); err != nil {
		webhook.log("invalid body: %s", err)
		return &WebhookResponse{
			Status:      http.StatusBadRequest,
			ContentType: "text/plain",
			Body:        []byte("failed to parse body"),
		}
	}

	update := newUpdateWebhook(baseUpdate, webhook.client)
	defer update.disableWebhookReply()

	done := make(chan struct{})

	go func() {
		handlerCtx, handlerCtxClose := context.WithCancel(context.Background())
		defer handlerCtxClose()

		// handle update
		if err := webhook.handler.Handle(handlerCtx, update); err != nil {
			webhook.log("handler error: %v", err)
		}

		close(done)
	}()

	select {
	case <-ctx.Done():
		return &WebhookResponse{
			Status: http.StatusOK,
		}
	case reply := <-update.webhookReply:
		response := &WebhookResponse{
			Status: http.StatusOK,
		}

		if reply != nil {
			body, err := json.Marshal(reply)

			if err != nil {
				webhook.log("marshal webhook response error: %v", err)
				return response
			}

			response.ContentType = "application/json"
			response.Body = body
		}

		return response
	case <-done:
		return &WebhookResponse{
			Status: http.StatusOK,
		}
	}
}

// ServeHTTP is the HTTP handler for webhook requests.
// Implementation of http.Handler.
func (webhook *Webhook) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ip, err := netip.ParseAddr(webhook.ipFromRequestFunc(r))
	if err != nil {
		webhook.log("failed to parse ip: %s", err)
		http.Error(w, "failed to parse ip", http.StatusBadRequest)
		return
	}

	request := &WebhookRequest{
		Method:        r.Method,
		ContentType:   r.Header.Get("Content-Type"),
		IP:            ip,
		SecurityToken: r.Header.Get(securityTokenHeader),
		Body:          r.Body,
	}

	response := webhook.ServeRequest(r.Context(), request)

	if response.ContentType != "" {
		w.Header().Set("Content-Type", response.ContentType)
	}

	if response.Status != 0 {
		w.WriteHeader(response.Status)
	}

	if response.Body != nil {
		_, _ = w.Write(response.Body)
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
