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

		err := r.UnmarshalJSON([]byte(`{"type": "emoji", "emoji": "😀"}`))
		require.NoError(t, err)

		assert.Equal(t, ReactionTypeTypeEmoji, r.Type())
		require.NotNil(t, r.Emoji)
		assert.Equal(t, "😀", r.Emoji.Emoji)
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

		err := r.UnmarshalJSON([]byte(`{"type": "unknown"}`))
		require.Error(t, err)
	})

	t.Run("InvalidFieldType", func(t *testing.T) {
		var r ReactionType

		err := r.UnmarshalJSON([]byte(`{"type": 123}`))
		require.Error(t, err)
	})
}

func TestReactionType_MarshalJSON(t *testing.T) {
	t.Run("Emoji", func(t *testing.T) {
		r := newEmojiReaction("😀")

		assert.Equal(t, ReactionTypeTypeEmoji, r.Type())

		b, err := json.Marshal(r)
		require.NoError(t, err)

		assert.Equal(t, `{"type":"emoji","emoji":"😀"}`, string(b))
	})

	t.Run("CustomEmoji", func(t *testing.T) {
		r := newCustomEmojiReaction("12345")

		assert.Equal(t, ReactionTypeTypeCustomEmoji, r.Type())

		b, err := json.Marshal(r)
		require.NoError(t, err)

		assert.Equal(t, `{"type":"custom_emoji","custom_emoji_id":"12345"}`, string(b))
	})

	t.Run("Unknown", func(t *testing.T) {
		r := ReactionType{}

		assert.Equal(t, ReactionTypeType(0), r.Type())

		_, err := json.Marshal(r)
		require.Error(t, err)
	})
}
