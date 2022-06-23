package tg

import (
	"encoding/json"
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

func TestChatType_String(t *testing.T) {
	for _, test := range []struct {
		ChatType ChatType
		Want     string
	}{
		{ChatTypePrivate, "private"},
		{ChatTypeGroup, "group"},
		{ChatTypeSupergroup, "supergroup"},
		{ChatTypeChannel, "channel"},
		{ChatType(-1), "unknown"},
	} {
		assert.Equal(t, test.Want, test.ChatType.String())
	}
}

func TestChatType_MarshalJSON(t *testing.T) {
	type sample struct {
		Type ChatType `json:"type"`
	}
	for _, test := range []struct {
		Sample sample
		Want   string
	}{
		{sample{ChatTypePrivate}, `{"type":"private"}`},
		{sample{ChatTypeGroup}, `{"type":"group"}`},
		{sample{ChatTypeSupergroup}, `{"type":"supergroup"}`},
		{sample{ChatTypeChannel}, `{"type":"channel"}`},
		{sample{ChatType(-1)}, `"{"type":"unknown"}"`},
	} {
		actual, err := json.Marshal(test.Sample)
		assert.NoError(t, err)

		assert.Equal(t, test.Want, string(actual))
	}
}

func TestChatType_UnmarshalJSON(t *testing.T) {
	type sample struct {
		Type ChatType `json:"type"`
	}
	for _, test := range []struct {
		Input  string
		Sample sample
		Want   ChatType
		Err    bool
	}{
		{`{"type": "private"}`, sample{}, ChatTypePrivate, false},
		{`{"type": "group"}`, sample{}, ChatTypeGroup, false},
		{`{"type": "supergroup"}`, sample{}, ChatTypeSupergroup, false},
		{`{"type": "channel"}`, sample{}, ChatTypeChannel, false},
		{`{"type": "test"}`, sample{}, ChatType(-1), true},
		{`{"type": {}}`, sample{}, ChatType(-1), true},
	} {
		if test.Err {
			assert.Error(t, json.Unmarshal([]byte(test.Input), &test.Sample))
		} else {
			assert.NoError(t, json.Unmarshal([]byte(test.Input), &test.Sample))
			assert.Equal(t, test.Want, test.Sample.Type)
		}
	}
}
