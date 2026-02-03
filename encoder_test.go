package tg

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestURLEncodedEncoder_WriteString(t *testing.T) {
	buf := bytes.Buffer{}

	encoder := newURLEncodedEncoder(&buf)

	if assert.NoError(t, encoder.WriteString("a", "1")) {
		assert.Equal(t, "a=1", buf.String())
	}

	if assert.NoError(t, encoder.WriteString("b", "2")) {
		assert.Equal(t, "a=1&b=2", buf.String())
	}

	if assert.NoError(t, encoder.WriteString("c", "3")) {
		assert.Equal(t, "a=1&b=2&c=3", buf.String())
	}
}

func TestURLEncodedEncoder_WriteFile(t *testing.T) {
	encoder := newURLEncodedEncoder(nil)

	assert.Error(t, encoder.WriteFile("a", InputFile{}))
}

func TestURLEncodedEncoder_ContentType(t *testing.T) {
	encoder := newURLEncodedEncoder(nil)

	assert.Equal(t, "application/x-www-form-urlencoded", encoder.ContentType())
}

func TestURLEncodedEncoder_Close(t *testing.T) {
	encoder := newURLEncodedEncoder(nil)

	assert.NoError(t, encoder.Close())
}

func TestMultipartEncoder(t *testing.T) {
	buf := bytes.Buffer{}

	encoder := newMultipartEncoder(&buf)

	assert.NoError(t, encoder.WriteString("a", "1"))
	assert.NoError(t, encoder.WriteString("b", "2"))
	assert.NoError(t, encoder.WriteString("c", "3"))

	assert.NoError(t, encoder.Close())

	assert.NotEmpty(t, buf.String())
}

func TestMultipartEncoder_WriteFile(t *testing.T) {
	buf := bytes.Buffer{}

	encoder := newMultipartEncoder(&buf)

	file := NewInputFileBytes("test.txt", []byte("bla-bla-bla"))

	assert.NoError(t, encoder.WriteFile("document", file))

	assert.NoError(t, encoder.Close())

	assert.NotEmpty(t, buf.String())
}

func TestMultipartEncoder_ContentType(t *testing.T) {
	encoder := newMultipartEncoder(nil)

	assert.True(t, strings.HasPrefix(encoder.ContentType(), "multipart/form-data; boundary"))
}

func TestMultipartEncoder_Close(t *testing.T) {
	buf := bytes.Buffer{}

	encoder := newMultipartEncoder(&buf)

	assert.NoError(t, encoder.Close())
}
