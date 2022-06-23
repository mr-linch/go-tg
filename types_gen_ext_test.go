package tg

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPeerIDImpl(t *testing.T) {
	for _, test := range []struct {
		PeerID PeerID
		Want   string
	}{
		{UserID(1), "1"},
		{ChatID(1), "1"},
		{&Chat{ID: ChatID(1)}, "1"},
	} {
		assert.Equal(t, test.Want, test.PeerID.PeerID())
	}
}
