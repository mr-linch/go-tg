package tg

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewClient(t *testing.T) {
	doer := &http.Client{}

	client := New("token",
		WithServer("http://example.com"),
		WithDoer(doer),
	)

	assert.Equal(t, "http://example.com", client.server)
}

func TestClientExecute(t *testing.T) {
	t.Run("Simple", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "POST", r.Method)
			assert.Equal(t, "/bot1234:secret/getMe", r.URL.Path)

			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"ok":true,"result":{"id":5556648742,"is_bot":true,"first_name":"go_tg_local_bot","username":"go_tg_local_bot","can_join_groups":true,"can_read_all_group_messages":false,"supports_inline_queries":false}}`))
		}))

		defer ts.Close()

		client := New("1234:secret", WithDoer(ts.Client()), WithServer(ts.URL))
		ctx := context.Background()

		res, err := client.execute(ctx, NewRequest("getMe"))

		if assert.NoError(t, err) {
			assert.Equal(t,
				json.RawMessage(`{"id":5556648742,"is_bot":true,"first_name":"go_tg_local_bot","username":"go_tg_local_bot","can_join_groups":true,"can_read_all_group_messages":false,"supports_inline_queries":false}`),
				res.Result,
			)
		}
	})

	t.Run("Streaming", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "POST", r.Method)
			assert.Equal(t, "/bot1234:secret/sendDocument", r.URL.Path)
			assert.True(t, strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data;"))

			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"ok":true,"result":{"message_id":4,"from":{"id":5556648742,"is_bot":true,"first_name":"go_tg_local_bot","username":"go_tg_local_bot"},"chat":{"id":103980787,"first_name":"Sasha","username":"MrLinch","type":"private"},"date":1655488910,"document":{"file_name":"types.go","file_id":"BQACAgIAAxkDAAMEYqzBjtP0VieRu8CCjHeNxnEetlsAAiIbAALAuWFJgQyZP4JcwDkkBA","file_unique_id":"AgADIhsAAsC5YUk","file_size":30}}}`))
		}))

		defer ts.Close()

		client := New("1234:secret", WithDoer(ts.Client()), WithServer(ts.URL))
		ctx := context.Background()

		file := NewInputFileBytes("types.go", []byte("package tg"))

		res, err := client.execute(ctx,
			NewRequest("sendDocument").
				File("document", file).
				String("chat_id", "1234567"),
		)

		if assert.NoError(t, err) {
			assert.Equal(t,
				json.RawMessage(`{"message_id":4,"from":{"id":5556648742,"is_bot":true,"first_name":"go_tg_local_bot","username":"go_tg_local_bot"},"chat":{"id":103980787,"first_name":"Sasha","username":"MrLinch","type":"private"},"date":1655488910,"document":{"file_name":"types.go","file_id":"BQACAgIAAxkDAAMEYqzBjtP0VieRu8CCjHeNxnEetlsAAiIbAALAuWFJgQyZP4JcwDkkBA","file_unique_id":"AgADIhsAAsC5YUk","file_size":30}}`),
				res.Result,
			)
		}
	})

	t.Run("StreamingError", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "POST", r.Method)
			assert.Equal(t, "/bot1234:secret/sendDocument", r.URL.Path)
			assert.True(t, strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data;"))

			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"ok":false,"error_code":400,"description":"Bad Request: chat not found"}`))
		}))

		defer ts.Close()

		client := New("1234:secret", WithDoer(ts.Client()), WithServer(ts.URL))
		ctx := context.Background()

		file := NewInputFileBytes("types.go", []byte("package tg"))

		res, err := client.execute(ctx,
			NewRequest("sendDocument").
				File("document", file).
				String("chat_id", "1234567"),
		)

		if assert.NoError(t, err) {
			assert.Equal(t, res.Description, "Bad Request: chat not found")
			assert.Equal(t, res.StatusCode, http.StatusBadRequest)
		}
	})

}
