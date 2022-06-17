package tg

import (
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
		file, close, err := NewInputFileLocal("./testdata/gopher.png")

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
