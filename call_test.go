package tg

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBindClient(t *testing.T) {
	t.Run("Call", func(t *testing.T) {
		call := NewGetMeCall()

		assert.Nil(t, call.client)

		BindClient(call, &Client{})

		assert.NotNil(t, call.client)
	})

	t.Run("CallNoResult", func(t *testing.T) {
		call := NewSetChatTitleCall(ChatID(1), "Hello")

		assert.Nil(t, call.client)

		BindClient(call, &Client{})

		assert.NotNil(t, call.client)
	})
}

func TestCall_MarshalJSON(t *testing.T) {
	call := NewSendMessageCall(ChatID(1), "Hello")

	v, err := call.MarshalJSON()
	assert.NoError(t, err)

	assert.JSONEq(t, `{"chat_id":"1","text":"Hello","method":"sendMessage"}`, string(v))
}

func TestCallNoResult_MarshalJSON(t *testing.T) {
	call := NewSetChatTitleCall(ChatID(1), "Hello")

	v, err := call.MarshalJSON()
	assert.NoError(t, err)

	assert.JSONEq(t, `{"chat_id":"1","title":"Hello","method":"setChatTitle"}`, string(v))
}
