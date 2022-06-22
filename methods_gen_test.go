package tg

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	testTokenEnv = "GO_TG_TEST_TOKEN"
	// testChatEnv  = "GO_TG_TEST_CHAT"
)

func getEnv(t *testing.T, k string) string {
	t.Helper()

	v := os.Getenv(k)

	if v == "" {
		t.Fatal("env var " + k + " is not set")
	}

	return v
}

// func getInt64Env(t *testing.T, k string) int64 {
// 	t.Helper()

// 	v := getEnv(t, k)

// 	i, err := strconv.ParseInt(v, 10, 64)
// 	if err != nil {
// 		t.Fatal("env var " + k + " is not an integer")
// 	}

// 	return i
// }

// func testWithClientLocal(
// 	t *testing.T,
// 	do func(t *testing.T, ctx context.Context, client *Client),
// 	handler http.HandlerFunc,
// ) {
// 	t.Helper()

// 	if testing.Short() {
// 		t.Skip("skipping test in short mode")
// 	}

// 	server := httptest.NewServer(handler)
// 	defer server.Close()

// 	client := New(getEnv(t, testTokenEnv),
// 		WithServer(server.URL),
// 		WithDoer(http.DefaultClient),
// 	)

// 	ctx := context.Background()

// 	do(t, ctx, client)
// }

func testWithClient(t *testing.T, do func(t *testing.T, ctx context.Context, client *Client)) {
	t.Helper()

	if testing.Short() {
		t.Log("skipping integration test")
		return
	}

	client := New(getEnv(t, testTokenEnv))

	ctx := context.Background()

	do(t, ctx, client)
}

func TestClient_GetMe(t *testing.T) {
	testWithClient(t, func(t *testing.T, ctx context.Context, client *Client) {
		user, err := client.GetMe().Do(ctx)

		if assert.NoError(t, err) {
			assert.Equal(t, User{
				ID:                      5433024556,
				IsBot:                   true,
				FirstName:               "go-tg: test bot",
				LastName:                "",
				Username:                "go_tg_test_bot",
				LanguageCode:            "",
				IsPremium:               false,
				AddedToAttachmentMenu:   false,
				CanJoinGroups:           true,
				CanReadAllGroupMessages: true,
				SupportsInlineQueries:   true,
			}, user)
		}
	})
}

func TestClient_Updates(t *testing.T) {
	testWithClient(t, func(t *testing.T, ctx context.Context, client *Client) {
		updates, err := client.GetUpdates().
			Offset(0).
			Limit(1).
			Timeout(1).
			AllowedUpdates([]string{"message"}).
			Do(ctx)

		if assert.NoError(t, err) {
			assert.GreaterOrEqual(t, len(updates), 1)
		}
	})
}
