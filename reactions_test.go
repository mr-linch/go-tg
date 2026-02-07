package tg

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReactionType(t *testing.T) {
	t.Run("Emoji", func(t *testing.T) {
		var r ReactionType

		err := r.UnmarshalJSON([]byte(`{"type": "emoji", "emoji": "👍"}`))
		require.NoError(t, err)

		assert.Equal(t, ReactionTypeTypeEmoji, r.Type())
		require.NotNil(t, r.Emoji)
		assert.Equal(t, ReactionEmojiThumbsUp, r.Emoji.Emoji)
	})

	t.Run("CustomEmoji", func(t *testing.T) {
		var r ReactionType

		err := r.UnmarshalJSON([]byte(`{"type": "custom_emoji", "custom_emoji_id": "12345"}`))
		require.NoError(t, err)

		assert.Equal(t, ReactionTypeTypeCustomEmoji, r.Type())
		require.NotNil(t, r.CustomEmoji)
		assert.Equal(t, "12345", r.CustomEmoji.CustomEmojiID)
	})

	t.Run("Unknown", func(t *testing.T) {
		var r ReactionType

		err := r.UnmarshalJSON([]byte(`{"type": "future_type", "some_field": "value"}`))
		require.NoError(t, err)

		assert.True(t, r.IsUnknown())
		require.NotNil(t, r.Unknown)
		assert.Equal(t, "future_type", r.Unknown.Type)
		assert.Equal(t, ReactionTypeType(0), r.Type())
		assert.Nil(t, r.Emoji)
		assert.Nil(t, r.CustomEmoji)
		assert.Nil(t, r.Paid)
	})

	t.Run("InvalidFieldType", func(t *testing.T) {
		var r ReactionType

		err := r.UnmarshalJSON([]byte(`{"type": 123}`))
		require.Error(t, err)
	})
}

func TestReactionType_MarshalJSON(t *testing.T) {
	t.Run("Emoji", func(t *testing.T) {
		r := NewReactionTypeEmoji(ReactionEmojiThumbsUp).AsReactionType()

		assert.Equal(t, ReactionTypeTypeEmoji, r.Type())

		b, err := json.Marshal(r)
		require.NoError(t, err)

		assert.JSONEq(t, `{"type":"emoji","emoji":"👍"}`, string(b))
	})

	t.Run("CustomEmoji", func(t *testing.T) {
		r := NewReactionTypeCustomEmoji("12345").AsReactionType()

		assert.Equal(t, ReactionTypeTypeCustomEmoji, r.Type())

		b, err := json.Marshal(r)
		require.NoError(t, err)

		assert.JSONEq(t, `{"type":"custom_emoji","custom_emoji_id":"12345"}`, string(b))
	})

	t.Run("Empty", func(t *testing.T) {
		r := ReactionType{}

		assert.Equal(t, ReactionTypeType(0), r.Type())

		_, err := json.Marshal(r)
		require.Error(t, err)
	})

	t.Run("Unknown", func(t *testing.T) {
		input := `{"type":"future_type","some_field":"value"}`
		var r ReactionType
		err := r.UnmarshalJSON([]byte(input))
		require.NoError(t, err)

		assert.True(t, r.IsUnknown())

		// Re-marshal should preserve original JSON
		output, err := json.Marshal(r)
		require.NoError(t, err)
		assert.JSONEq(t, input, string(output))
	})
}
