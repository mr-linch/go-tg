package tgb

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/mr-linch/go-tg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewWebhook(t *testing.T) {
	t.Run("Default", func(t *testing.T) {
		webhook := NewWebhook(
			HandlerFunc(func(ctx context.Context, update *Update) error { return nil }),
			&tg.Client{},
			"https://example.com/webhook",
		)

		assert.Equal(t, "https://example.com/webhook", webhook.url)
		assert.NotNil(t, webhook.handler)
		assert.NotNil(t, webhook.securityToken)
		assert.Len(t, webhook.securitySubnets, 2)
	})
	t.Run("Custom", func(t *testing.T) {
		webhook := NewWebhook(
			HandlerFunc(func(ctx context.Context, update *Update) error { return nil }),
			&tg.Client{},
			"https://example.com/webhook",
			WithWebhookIP("1.1.1.1"),
			WithWebhookSecuritySubnets(netip.MustParsePrefix("1.1.1.1/24")),
			WithWebhookSecurityToken("12345"),
			WithWebhookMaxConnections(10),
		)

		assert.Equal(t, "https://example.com/webhook", webhook.url)
		assert.NotNil(t, webhook.handler)
		assert.Equal(t, "12345", webhook.securityToken)
		assert.Len(t, webhook.securitySubnets, 1)
		assert.Equal(t, 10, webhook.maxConnections)
		assert.Equal(t, "1.1.1.1", webhook.ip)
	})
}

func TestWebhook_ServeHTTP(t *testing.T) {
	t.Run("InvalidMethod", func(t *testing.T) {
		w := httptest.NewRecorder()

		req, err := http.NewRequest(http.MethodGet, "/", strings.NewReader(""))
		assert.NoError(t, err)
		req.RemoteAddr = "1.1.1.1"

		webhook := NewWebhook(
			HandlerFunc(func(ctx context.Context, update *Update) error { return nil }),
			&tg.Client{},
			"http://test.io/",
		)

		webhook.ServeHTTP(w, req)

		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})

	t.Run("InvalidIP", func(t *testing.T) {
		w := httptest.NewRecorder()

		req, err := http.NewRequest(http.MethodPost, "/", strings.NewReader(""))
		assert.NoError(t, err)

		req.Header.Set("X-Forwarded-For", "1-1-1-1")

		webhook := NewWebhook(
			HandlerFunc(func(ctx context.Context, update *Update) error { return nil }),
			&tg.Client{},
			"http://test.io/",
		)

		webhook.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("SecurityCheckIP", func(t *testing.T) {
		w := httptest.NewRecorder()

		req, err := http.NewRequest(http.MethodPost, "/", strings.NewReader(""))
		assert.NoError(t, err)

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Forwarded-For", "1.1.1.1")

		webhook := NewWebhook(
			HandlerFunc(func(ctx context.Context, update *Update) error { return nil }),
			&tg.Client{},
			"http://test.io/",
		)

		webhook.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("SecurityCheckToken", func(t *testing.T) {
		w := httptest.NewRecorder()

		req, err := http.NewRequest(http.MethodPost, "/", strings.NewReader(""))
		assert.NoError(t, err)

		req.Header.Set(securityTokenHeader, "secret")
		req.Header.Set("X-Forwarded-For", "1.1.1.1")
		req.Header.Set("Content-Type", "application/json")

		webhook := NewWebhook(
			HandlerFunc(func(ctx context.Context, update *Update) error { return nil }),
			&tg.Client{},
			"http://test.io/",
			WithWebhookSecuritySubnets(), // disable ip check
		)

		webhook.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("InvalidContentType", func(t *testing.T) {
		w := httptest.NewRecorder()

		req, err := http.NewRequest(http.MethodPost, "/", strings.NewReader(""))
		assert.NoError(t, err)

		req.RemoteAddr = "1.1.1.1"
		req.Header.Set("Content-Type", "text/plain")

		webhook := NewWebhook(
			HandlerFunc(func(ctx context.Context, update *Update) error { return nil }),
			&tg.Client{},
			"http://test.io/",
			WithWebhookSecuritySubnets(),
			WithWebhookSecurityToken(""),
		)

		webhook.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnsupportedMediaType, w.Code)
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		w := httptest.NewRecorder()

		req, err := http.NewRequest(http.MethodPost, "/", strings.NewReader("{"))
		assert.NoError(t, err)

		req.RemoteAddr = "1.1.1.1"
		req.Header.Set("Content-Type", "application/json")

		webhook := NewWebhook(
			HandlerFunc(func(ctx context.Context, update *Update) error { return nil }),
			&tg.Client{},
			"http://test.io/",
			WithWebhookSecuritySubnets(),
			WithWebhookSecurityToken(""),
		)

		webhook.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("HandleError", func(t *testing.T) {
		w := httptest.NewRecorder()

		req, err := http.NewRequest(http.MethodPost, "/", strings.NewReader("{}"))
		assert.NoError(t, err)

		req.RemoteAddr = "1.1.1.1"
		req.Header.Set("Content-Type", "application/json")

		webhook := NewWebhook(
			HandlerFunc(func(ctx context.Context, update *Update) error { return errors.New("something goes wrong") }),
			&tg.Client{},
			"http://test.io/",
			WithWebhookSecuritySubnets(),
			WithWebhookSecurityToken(""),
		)

		webhook.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("HandleOKEmpty", func(t *testing.T) {
		w := httptest.NewRecorder()

		req, err := http.NewRequest(http.MethodPost, "/", strings.NewReader("{}"))
		assert.NoError(t, err)

		req.RemoteAddr = "149.154.160.2" // ip from default telegram range
		req.Header.Set("Content-Type", "application/json")

		webhook := NewWebhook(
			HandlerFunc(func(ctx context.Context, update *Update) error { return nil }),
			&tg.Client{},
			"http://test.io/",
			WithWebhookSecurityToken(""),
		)

		webhook.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("HandleOKOneReplyCall", func(t *testing.T) {
		w := httptest.NewRecorder()

		req, err := http.NewRequest(http.MethodPost, "/", strings.NewReader(`{"update_id": 123456, "message": {"chat": {"id": 1234}}}`))
		assert.NoError(t, err)

		req.RemoteAddr = "1.1.1.1"
		req.Header.Set("Content-Type", "application/json")

		isHandlerCalled := false

		webhook := NewWebhook(
			HandlerFunc(func(ctx context.Context, update *Update) error {
				isHandlerCalled = true
				return update.Reply(ctx, tg.NewSendMessageCall(update.Message.Chat, "test"))
			}),
			&tg.Client{},
			"http://test.io/",
			WithWebhookSecuritySubnets(),
			WithWebhookSecurityToken(""),
		)

		webhook.ServeHTTP(w, req)

		assert.True(t, isHandlerCalled, "handler should be called")

		assert.Equal(t, http.StatusOK, w.Code)

		body, err := io.ReadAll(w.Body)
		assert.NoError(t, err)

		assert.Equal(t, `{"chat_id":"1234","method":"sendMessage","text":"test"}`, string(body))
	})

	t.Run("HandleOKTwoReplyCall", func(t *testing.T) {

		isHandlerCalled := false

		handlerDone := make(chan struct{})

		webhook := NewWebhook(
			HandlerFunc(func(ctx context.Context, update *Update) error {
				// first call should be send response to webhook
				err := update.Reply(ctx, tg.NewSendMessageCall(update.Message.Chat, "test"))
				assert.NoError(t, err)

				// second call should call UpdateReply.DoVoid()
				ur := &MockUpdateReply{}

				ur.On("Bind", mock.Anything).Return()
				ur.On("DoVoid", mock.Anything).Return(nil)

				err = update.Reply(ctx, ur)
				assert.NoError(t, err)

				ur.AssertExpectations(t)

				isHandlerCalled = true

				close(handlerDone)

				return nil
			}),
			&tg.Client{},
			"http://test.io/",
			WithWebhookSecuritySubnets(),
			WithWebhookSecurityToken(""),
		)

		req, err := http.NewRequest(http.MethodPost, "/", strings.NewReader(`{"update_id": 123456, "message": {"chat": {"id": 1234}}}`))
		assert.NoError(t, err)
		req.RemoteAddr = "1.1.1.1"
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()

		webhook.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		body, err := io.ReadAll(w.Body)
		assert.NoError(t, err)

		assert.Equal(t, `{"chat_id":"1234","method":"sendMessage","text":"test"}`, string(body))

		<-handlerDone

		assert.True(t, isHandlerCalled, "handler is not called")

	})
}

func TestWebhook_Setup(t *testing.T) {
	t.Run("NoUpdate", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/bot1234:secret/getWebhookInfo":
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"ok":true,"result":{"url":"https://google.com","has_custom_certificate":false,"pending_update_count":3,"last_error_date":1656177074,"last_error_message":"Wrong response from the webhook: 405 Method Not Allowed","max_connections":40,"ip_address":"216.58.208.110"}}`))
			default:
				t.Fatalf("unexcepted call '%s'", r.URL.Path)
			}
		}))

		defer server.Close()

		webhook := NewWebhook(
			HandlerFunc(func(ctx context.Context, update *Update) error { return nil }),
			tg.New("1234:secret", tg.WithClientServerURL(server.URL), tg.WithClientDoer(server.Client())),
			"https://google.com",
		)

		err := webhook.Setup(context.Background())
		assert.NoError(t, err)
	})

	t.Run("ShouldUpdateBecouseDropPendingAndHavePending", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/bot1234:secret/getWebhookInfo":
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"ok":true,"result":{"url":"https://google.com","has_custom_certificate":false,"pending_update_count":3,"last_error_date":1656177074,"last_error_message":"Wrong response from the webhook: 405 Method Not Allowed","max_connections":40,"ip_address":"216.58.208.110"}}`))
			case "/bot1234:secret/setWebhook":
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"ok":true,"result": true}`))

				body, err := io.ReadAll(r.Body)
				assert.NoError(t, err)

				vs, err := url.ParseQuery(string(body))
				assert.NoError(t, err)
				assert.Equal(t, url.Values{
					"drop_pending_updates": []string{"true"},
					"max_connections":      []string{"40"},
					"secret_token":         []string{"973b4c22458364768284928867d93c992e2b2db94e81f7dbca28e171390a0363"},
					"url":                  []string{"https://google.com"},
					"ip_address":           []string{"1.1.1.1"},
					"allowed_updates":      []string{"[]"},
				}, vs)
			default:
				t.Fatalf("unexcepted call '%s'", r.URL.Path)
			}
		}))

		defer server.Close()

		webhook := NewWebhook(
			HandlerFunc(func(ctx context.Context, update *Update) error { return nil }),
			tg.New("1234:secret", tg.WithClientServerURL(server.URL), tg.WithClientDoer(server.Client())),
			"https://google.com",
			WithDropPendingUpdates(true),
			WithWebhookIP("1.1.1.1"),
		)

		err := webhook.Setup(context.Background())
		assert.NoError(t, err)
	})

	t.Run("ShouldUpdateBecouseAllowedUpdatesChanged", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/bot1234:secret/getWebhookInfo":
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"ok":true,"result":{"url":"https://google.com","has_custom_certificate":false,"pending_update_count":3,"last_error_date":1656177074,"last_error_message":"Wrong response from the webhook: 405 Method Not Allowed","max_connections":40,"ip_address":"216.58.208.110"}}`))
			case "/bot1234:secret/setWebhook":
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"ok":true,"result": true}`))

				body, err := io.ReadAll(r.Body)
				assert.NoError(t, err)

				vs, err := url.ParseQuery(string(body))
				assert.NoError(t, err)
				assert.EqualValues(t, url.Values{
					"secret_token":    []string{"973b4c22458364768284928867d93c992e2b2db94e81f7dbca28e171390a0363"},
					"url":             []string{"https://google.com"},
					"ip_address":      []string{"1.1.1.1"},
					"allowed_updates": []string{"[\"message\"]"},
					"max_connections": []string{"40"},
				}, vs)

			default:
				t.Fatalf("unexcepted call '%s'", r.URL.Path)
			}
		}))

		defer server.Close()

		webhook := NewWebhook(
			HandlerFunc(func(ctx context.Context, update *Update) error { return nil }),
			tg.New("1234:secret", tg.WithClientServerURL(server.URL), tg.WithClientDoer(server.Client())),
			"https://google.com",
			WithWebhookAllowedUpdates(tg.UpdateTypeMessage),
			WithWebhookIP("1.1.1.1"),
		)

		err := webhook.Setup(context.Background())
		assert.NoError(t, err)
	})

}

type loggerMock struct {
	mock.Mock
}

func (mock *loggerMock) Printf(format string, v ...interface{}) {
	mock.Called(format, v)
}

func TestWebhook_log(t *testing.T) {
	t.Run("Default", func(t *testing.T) {
		wh := NewWebhook(
			HandlerFunc(func(ctx context.Context, update *Update) error { return nil }),
			&tg.Client{},
			"https://google.com",
		)

		wh.log("test")
	})

	t.Run("WithLogger", func(t *testing.T) {
		logger := &loggerMock{}
		logger.On("Printf", "tgb.Webhook: test", mock.Anything).Return()

		wh := NewWebhook(
			HandlerFunc(func(ctx context.Context, update *Update) error { return nil }),
			&tg.Client{},
			"https://google.com",
			WithWebhookLogger(logger),
		)

		wh.log("test")

		logger.AssertExpectations(t)
	})
}

func TestWebhook_Run(t *testing.T) {
	webhook := NewWebhook(
		HandlerFunc(func(ctx context.Context, update *Update) error { return nil }),
		&tg.Client{},
		"https://google.com",
	)
	webhook.isSetup = true

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		err := webhook.Run(ctx, ":12345")
		assert.NoError(t, err)
	}()

	time.Sleep(time.Millisecond * 100)

	cancel()
}
