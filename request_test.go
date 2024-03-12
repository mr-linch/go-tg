package tg

import (
	"encoding/json"
	"errors"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewRequest(t *testing.T) {
	r := NewRequest("getMe")

	assert.Equal(t, "getMe", r.Method)
}

func TestRequest_Setters(t *testing.T) {
	r := NewRequest("getMe")

	r.JSON("foo", "bar")
	r.InputFile("file", InputFile{})
	r.PeerID("peer", ChatID(1))
	r.UserID("user", UserID(1))
	r.String("str", "bar")
	r.Bool("bool", true)
	r.Int("int", 1)
	r.Int64("int64", 1)
	r.Float64("float64", 1)
	r.ChatID("chat", ChatID(1))
	r.Stringer("parse_mode", MD2)
	r.File("file_by_id", FileArg{
		FileID: "file_id",
	})
	r.File("file_by_url", FileArg{
		URL: "file_url",
	})
	r.File("file_input", FileArg{
		Upload: NewInputFileBytes("file_name", []byte("file_content")),
	})

	r.FileID("file_id", FileID("file_id"))

	r.InputMediaSlice("media", []InputMedia{
		&InputMediaDocument{
			Media: FileArg{
				Upload: NewInputFileBytes("file_name", []byte("file_content")),
			},

			Thumbnail: NewInputFileBytes("thumb.jpg", []byte("")).Ptr(),
		},
	})

	encoder := &testEncoder{}

	err := r.Encode(encoder)
	assert.NoError(t, err)

	sort.StringSlice(encoder.stringKeys).Sort()
	sort.StringSlice(encoder.fileKeys).Sort()

	assert.Equal(t, []string{
		"bool",
		"chat",
		"file_by_id",
		"file_by_url",
		"file_id",
		"float64",
		"foo",
		"int",
		"int64",
		"media",
		"parse_mode",
		"peer",
		"str",
		"user",
	}, encoder.stringKeys)

	assert.Equal(t, []string{
		"attachment_0",
		"attachment_0_thumb",
		"file",
		"file_input",
	}, encoder.fileKeys)
}

type testEncoder struct {
	stringKeys []string
	fileKeys   []string
}

func (encoder *testEncoder) WriteString(key, value string) error {
	encoder.stringKeys = append(encoder.stringKeys, key)
	return nil
}

func (encoder *testEncoder) WriteFile(key string, file InputFile) error {
	encoder.fileKeys = append(encoder.fileKeys, key)
	return nil
}

func TestRequest_Encode(t *testing.T) {
	t.Run("JSONToArgsError", func(t *testing.T) {
		r := NewRequest("sendFile")

		obj := &JSONMarshalerMock{}

		obj.On("MarshalJSON").Return(nil, errors.New("error"))

		r.String("chat_id", "1")
		r.JSON("object", obj)

		encoder := &MockEncoder{}

		err := r.Encode(encoder)
		assert.Error(t, err)
	})

	t.Run("WriteFileError", func(t *testing.T) {
		r := NewRequest("sendFile")

		r.String("chat_id", "1")
		r.File("file", NewFileArgUpload(NewInputFileBytes("file_name", []byte("file_content"))))

		encoder := &MockEncoder{}

		encoder.On("WriteFile", "file", mock.Anything).
			Return(errors.New("error"))

		err := r.Encode(encoder)

		assert.Error(t, err)

		encoder.AssertExpectations(t)
	})

	t.Run("WriteStringError", func(t *testing.T) {
		r := NewRequest("sendFile")

		r.String("chat_id", "1")

		encoder := &MockEncoder{}

		encoder.On("WriteString", "chat_id", "1").
			Return(errors.New("error"))

		err := r.Encode(encoder)

		assert.Error(t, err)

		encoder.AssertExpectations(t)
	})
}

type MockEncoder struct {
	mock.Mock
}

func (m *MockEncoder) WriteString(key, value string) error {
	args := m.Called(key, value)
	return args.Error(0)
}

func (m *MockEncoder) WriteFile(key string, file InputFile) error {
	args := m.Called(key, file)
	return args.Error(0)
}

type JSONMarshalerMock struct {
	mock.Mock
}

func (m *JSONMarshalerMock) MarshalJSON() ([]byte, error) {
	args := m.Called()

	v := args.Get(0)
	if v == nil {
		return nil, args.Error(1)
	}

	return v.([]byte), args.Error(1)
}

func TestRequest_MarshalJSON(t *testing.T) {
	t.Run("OK", func(t *testing.T) {

		r := NewRequest("sendMessage")

		r.String("chat_id", "1")
		r.JSON("object", struct {
			Key string
		}{
			Key: "value",
		})

		v, err := json.Marshal(r)
		assert.NoError(t, err)
		assert.Equal(t, `{"chat_id":"1","method":"sendMessage","object":"{\"Key\":\"value\"}"}`, string(v))
	})

	t.Run("Error", func(t *testing.T) {
		r := NewRequest("sendFile")

		r.String("chat_id", "1")
		r.File("file", NewFileArgUpload(NewInputFileBytes("file_name", []byte("file_content"))))

		_, err := json.Marshal(r)
		assert.Error(t, err)
	})

	t.Run("SubMarshalerError", func(t *testing.T) {
		r := NewRequest("sendFile")

		obj := &JSONMarshalerMock{}

		obj.On("MarshalJSON").Return(nil, errors.New("error"))

		r.String("chat_id", "1")
		r.JSON("object", obj)

		_, err := json.Marshal(r)
		assert.Error(t, err)
	})
}

func TestRequest_Has(t *testing.T) {
	r := NewRequest("sendMessage")

	r.String("chat_id", "1")

	assert.True(t, r.Has("chat_id"))
	assert.False(t, r.Has("text"))
}
