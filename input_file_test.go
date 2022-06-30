package tg

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInputFile_WithName(t *testing.T) {
	file := NewInputFile("test.txt", nil)

	newFile := file.WithName("new.txt")

	assert.Equal(t, "new.txt", newFile.Name)
	assert.Equal(t, "test.txt", file.Name)
}

func TestNewInputFile(t *testing.T) {
	body := strings.NewReader("test")

	file := NewInputFile("test.txt", body)

	assert.Equal(t, "test.txt", file.Name)
	assert.Equal(t, body, file.Body)
}

func TestNewInputFileBytes(t *testing.T) {
	body := []byte("test")

	file := NewInputFileBytes("test.txt", body)

	assert.Equal(t, "test.txt", file.Name)
	assert.NotNil(t, file.Body)
}

func TestNewInputFileLocal(t *testing.T) {
	{
		file, close, err := NewInputFileLocal("examples/echo-bot/resources/gopher.png")

		if assert.NoError(t, err) {
			assert.Equal(t, "gopher.png", file.Name)
			assert.NotNil(t, file.Body)
			assert.NoError(t, close())
		}
	}

	{
		file, close, err := NewInputFileLocal("./testdata/not-exist.png")

		assert.Error(t, err)
		assert.Zero(t, file)
		assert.Nil(t, close)
	}
}

func TestInputFile_MarshalJSON(t *testing.T) {
	t.Run("WithoutAddr", func(t *testing.T) {
		file := NewInputFile("test.txt", nil)

		data, err := json.Marshal(&file)

		assert.Error(t, err)
		assert.Nil(t, data)
	})
	t.Run("WithAddr", func(t *testing.T) {
		file := NewInputFile("test.txt", nil)
		file.addr = "attach://test"

		data, err := json.Marshal(&file)

		assert.NoError(t, err)
		assert.Equal(t, `"attach://test"`, string(data))
	})
}

func TestInputFile_Ptr(t *testing.T) {
	file := NewInputFile("test.txt", nil)

	assert.Equal(t, &file, file.Ptr())
}
