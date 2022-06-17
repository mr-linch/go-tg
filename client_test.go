package tg

import (
	"net/http"
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
