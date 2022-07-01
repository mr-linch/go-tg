package tgb

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	tg "github.com/mr-linch/go-tg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestPoller(t *testing.T) {
	t.Run("Default", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			var (
				isGetUpdatesCalled     bool
				isGetWebhookInfoCalled bool
				isDeleteWebhookCalled  bool
			)

			switch r.URL.Path {
			case "/bot1234:secret/getWebhookInfo":
				isGetWebhookInfoCalled = true
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"ok":true,"result":{"url":"https://google.com","has_custom_certificate":false,"pending_update_count":3,"last_error_date":1656177074,"last_error_message":"Wrong response from the webhook: 405 Method Not Allowed","max_connections":40,"ip_address":"216.58.208.110"}}`))
			case "/bot1234:secret/deleteWebhook":
				isDeleteWebhookCalled = true
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"ok":true, "result":true}`))
			case "/bot1234:secret/getUpdates":
				isGetUpdatesCalled = true
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"ok":true,"result": [{}]}`))

				body, err := io.ReadAll(r.Body)
				assert.NoError(t, err)

				vs, err := url.ParseQuery(string(body))
				assert.NoError(t, err)

				assert.True(t, vs.Get("offset") == "0" || vs.Get("offset") == "1")
				assert.Equal(t, "5", vs.Get("timeout"))
				assert.Equal(t, "[]", vs.Get("allowed_updates"))

			default:
				t.Fatalf("unexcepted call '%s'", r.URL.Path)
			}

			assert.True(t, isGetUpdatesCalled || isGetWebhookInfoCalled || isDeleteWebhookCalled, "expected call one of getUpdates, getWebhookInfo or deleteWebhook")

		}))

		defer server.Close()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		err := NewPoller(
			HandlerFunc(func(ctx context.Context, update *Update) error {
				cancel()
				return nil
			}),
			tg.New("1234:secret", tg.WithClient(server.URL), tg.WithClientDoer(server.Client())),
		).Run(ctx)

		assert.NoError(t, err)
	})

	t.Run("Custom", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			var (
				isGetUpdatesCalled     bool
				isGetWebhookInfoCalled bool
				isDeleteWebhookCalled  bool
			)

			switch r.URL.Path {
			case "/bot1234:secret/getWebhookInfo":
				isGetWebhookInfoCalled = true
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"ok":true,"result":{"url":"https://google.com","has_custom_certificate":false,"pending_update_count":3,"last_error_date":1656177074,"last_error_message":"Wrong response from the webhook: 405 Method Not Allowed","max_connections":40,"ip_address":"216.58.208.110"}}`))
			case "/bot1234:secret/deleteWebhook":
				isDeleteWebhookCalled = true
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"ok":true, "result":true}`))
			case "/bot1234:secret/getUpdates":
				isGetUpdatesCalled = true
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"ok":true,"result": [{}]}`))

				body, err := io.ReadAll(r.Body)
				assert.NoError(t, err)

				vs, err := url.ParseQuery(string(body))
				assert.NoError(t, err)

				assert.True(t, vs.Get("offset") == "0" || vs.Get("offset") == "1")
				assert.Equal(t, "2", vs.Get("timeout"))
				assert.Equal(t, `["callback_query"]`, vs.Get("allowed_updates"))

			default:
				t.Fatalf("unexcepted call '%s'", r.URL.Path)
			}

			assert.True(t, isGetUpdatesCalled || isGetWebhookInfoCalled || isDeleteWebhookCalled, "expected call one of getUpdates, getWebhookInfo or deleteWebhook")

		}))

		defer server.Close()

		ctx, cancel := context.WithCancel(context.Background())

		err := NewPoller(
			HandlerFunc(func(ctx context.Context, update *Update) error {
				cancel()
				return nil
			}),
			tg.New("1234:secret", tg.WithClient(server.URL), tg.WithClientDoer(server.Client())),
			WithPollerRetryAfter(time.Millisecond),
			WithPollerHandlerTimeout(time.Millisecond),
			WithPollerAllowedUpdates([]string{"callback_query"}),
			WithPollerLimit(50),
			WithPollerTimeout(time.Second*2),
		).Run(ctx)
		assert.NoError(t, err)
	})
}

func TestPolling_log(t *testing.T) {
	t.Run("Default", func(t *testing.T) {
		poller := NewPoller(
			HandlerFunc(func(ctx context.Context, update *Update) error { return nil }),
			&tg.Client{},
		)

		poller.log("test")
	})

	t.Run("WithLogger", func(t *testing.T) {
		logger := &loggerMock{}

		poller := NewPoller(
			HandlerFunc(func(ctx context.Context, update *Update) error { return nil }),
			&tg.Client{},
			WithPollerLogger(logger),
		)

		logger.On("Printf", "tgb.Poller: test", mock.Anything).Return()

		poller.log("test")

		logger.AssertExpectations(t)
	})
}
