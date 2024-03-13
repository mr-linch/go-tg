package tgb

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCallbackDataParser(t *testing.T) {
	parser := NewCallackDataCodec(
		WithCallbackDataCodecDelimiter('$'),
		WithCallbackDataCodecIntBase(16),
		WithCallbackDataCodecFloatFmt('e'),
		WithCallbackDataCodecFloatPrec(3),
		WithCallbackDataCodecDisableLengthCheck(true),
	)

	assert.Equal(t, '$', parser.delimiter)
	assert.Equal(t, 16, parser.intBase)
	assert.Equal(t, byte('e'), parser.floatFmt)
	assert.Equal(t, 3, parser.floatPrec)
}

func TestCallbackDataParserEncode(t *testing.T) {
	t.Run("NotStruct", func(t *testing.T) {
		_, err := EncodeCallbackData(1)
		assert.ErrorContains(t, err, "src should be a struct")
	})

	t.Run("Nil", func(t *testing.T) {
		type test struct {
		}
		var nilStruct *test
		_, err := EncodeCallbackData(nilStruct)
		assert.ErrorContains(t, err, "src is nil")
	})

	t.Run("Empty", func(t *testing.T) {
		type test struct{}
		cbd, err := EncodeCallbackData(test{})
		require.NoError(t, err)

		assert.Equal(t, "", cbd)
	})

	t.Run("AllTypes", func(t *testing.T) {
		type test struct {
			Bool               bool
			BoolFalse          bool
			Int                int
			Uint               uint `tgbase:"10"`
			String             string
			Float32            float32 `tgfmt:"f" tgprec:"2"`
			Float64            float64 `tgprec:"3"`
			Floag64DefaultPrec float64
		}

		cbd, err := EncodeCallbackData(test{
			Bool:               true,
			Int:                -1234567890,
			Uint:               1234567890,
			String:             "xyz",
			Float32:            123.456,
			Float64:            123.4564,
			Floag64DefaultPrec: 123.45,
		})
		require.NoError(t, err)

		assert.Equal(t, "1:0:-kf12oi:1234567890:xyz:123.46:123.456:123.45", cbd)
	})

	t.Run("InvalidInt", func(t *testing.T) {
		type test struct {
			Int int `tgbase:"invalid"`
		}

		_, err := EncodeCallbackData(test{})
		assert.ErrorContains(t, err, "invalid base")
	})

	t.Run("InvalidUint", func(t *testing.T) {
		type test struct {
			Uint uint `tgbase:"invalid"`
		}

		_, err := EncodeCallbackData(test{})
		assert.ErrorContains(t, err, "invalid base")
	})

	t.Run("InvalidFloatFmt", func(t *testing.T) {
		type test struct {
			Float32 float32 `tgfmt:"invalid"`
		}

		_, err := EncodeCallbackData(test{})
		assert.ErrorContains(t, err, "invalid fmt value")
	})

	t.Run("InvalidFloatPrec", func(t *testing.T) {
		type test struct {
			Float32 float32 `tgprec:"invalid"`
		}

		_, err := EncodeCallbackData(test{})
		assert.ErrorContains(t, err, "invalid prec value")
	})

	t.Run("UnsupportedFieldType", func(t *testing.T) {
		type test struct {
			Unsupported chan int
		}

		_, err := EncodeCallbackData(test{})
		assert.ErrorContains(t, err, "unsupported field type: chan")
	})

	t.Run("CallbackDataIsTooLong", func(t *testing.T) {
		type test struct {
			Str string
		}

		_, err := EncodeCallbackData(test{
			Str: "12345678901234567890123456789012345678901234567890123456789012345678901234567890",
		})
		assert.ErrorContains(t, err, "callback data length is too long: 80, max: 64")
	})
}

func TestCallbackDataParserDecode(t *testing.T) {
	t.Run("NotStruct", func(t *testing.T) {
		var v int
		err := DecodeCallbackData("", &v)
		assert.ErrorContains(t, err, "dst should be a pointer to a struct")
	})

	t.Run("Nil", func(t *testing.T) {
		type test struct {
		}
		var nilStruct *test
		err := DecodeCallbackData("", nilStruct)
		assert.ErrorContains(t, err, "dst should be a pointer to a struct")

		var notNilStruct test
		err = DecodeCallbackData("", notNilStruct)
		assert.ErrorContains(t, err, "dst should be a pointer to a struct")
	})

	t.Run("InvalidDataLength", func(t *testing.T) {
		var dst struct {
			A int
			B int
		}

		err := DecodeCallbackData("1", &dst)

		assert.ErrorContains(t, err, "invalid data length")
	})

	t.Run("InvalidBoolValue", func(t *testing.T) {
		type test struct {
			Bool bool
		}

		var dst test
		err := DecodeCallbackData("invalid", &dst)
		assert.ErrorContains(t, err, "invalid bool value")
	})

	t.Run("InvalidInt", func(t *testing.T) {
		var dst struct {
			Int int `tgbase:"invalid"`
		}

		err := DecodeCallbackData("invalid", &dst)
		assert.ErrorContains(t, err, "invalid syntax")

		var dst2 struct {
			Int int `tgbase:"102"`
		}

		err = DecodeCallbackData("invalid", &dst2)
		assert.ErrorContains(t, err, "invalid base 102")
	})

	t.Run("InvalidInt", func(t *testing.T) {
		var dst struct {
			Uint uint `tgbase:"invalid"`
		}

		err := DecodeCallbackData("invalid", &dst)
		assert.ErrorContains(t, err, "invalid syntax")

		var dst2 struct {
			Uint uint `tgbase:"102"`
		}

		err = DecodeCallbackData("invalid", &dst2)
		assert.ErrorContains(t, err, "invalid base 102")
	})

	t.Run("InvalidFloat", func(t *testing.T) {
		var dst struct {
			Float32 float32 `tgfmt:"invalid"`
		}

		err := DecodeCallbackData("invalid", &dst)
		assert.ErrorContains(t, err, "invalid syntax")

		var dst2 struct {
			Float32 float32 `tgfmt:"e"`
		}

		err = DecodeCallbackData("invalid", &dst2)
		assert.ErrorContains(t, err, "invalid syntax")

		var dst3 struct {
			Float32 float32 `tgfmt:"e" tgprec:"invalid"`
		}

		err = DecodeCallbackData("invalid", &dst3)
		assert.ErrorContains(t, err, "invalid syntax")

	})

	t.Run("Empty", func(t *testing.T) {
		type test struct{}
		var dst test
		err := DecodeCallbackData("", &dst)
		require.NoError(t, err)
	})

	t.Run("AllTypes", func(t *testing.T) {
		type test struct {
			Bool      bool
			FalseBool bool
			Int       int
			Uint      uint `tgbase:"10"`
			String    string
			Float32   float32 `tgfmt:"f" tgprec:"1"`
			Float64   float64 `tgprec:"1"`
		}

		var dst test
		err := DecodeCallbackData("1:0:-kf12oi:1234567890:xyz:123.4:123.5", &dst)
		require.NoError(t, err)

		assert.Equal(t, test{
			Bool:    true,
			Int:     -1234567890,
			Uint:    1234567890,
			String:  "xyz",
			Float32: 123.4,
			Float64: 123.5,
		}, dst)
	})

	t.Run("UnsupportedFieldType", func(t *testing.T) {
		type test struct {
			Unsupported chan int
		}

		var dst test
		err := DecodeCallbackData("1", &dst)
		assert.ErrorContains(t, err, "unsupported field type: chan")
	})
}
