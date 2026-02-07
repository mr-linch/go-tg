package tg

import (
	"encoding/json"
	"errors"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
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
		NewInputMediaDocument(FileArg{
			Upload: NewInputFileBytes("file_name", []byte("file_content")),
		}).
			WithThumbnail(NewInputFileBytes("thumb.jpg", []byte(""))).
			AsInputMedia(),
	})

	encoder := &testEncoder{}

	err := r.Encode(encoder)
	require.NoError(t, err)

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

		require.Error(t, err)

		encoder.AssertExpectations(t)
	})

	t.Run("WriteStringError", func(t *testing.T) {
		r := NewRequest("sendFile")

		r.String("chat_id", "1")

		encoder := &MockEncoder{}

		encoder.On("WriteString", "chat_id", "1").
			Return(errors.New("error"))

		err := r.Encode(encoder)

		require.Error(t, err)

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

func TestRequest_InputMediaUpload(t *testing.T) {
	makeMedia := func() []InputMedia {
		return []InputMedia{
			NewInputMediaPhoto(FileArg{
				Upload: NewInputFileBytes("photo.jpg", []byte("photo_content")),
			}).AsInputMedia(),
		}
	}

	t.Run("InputMediaSlice", func(t *testing.T) {
		r := NewRequest("sendMediaGroup")
		r.InputMediaSlice("media", makeMedia())

		encoder := &testEncoder{}
		err := r.Encode(encoder)
		require.NoError(t, err)

		assert.Contains(t, encoder.stringKeys, "media")
		assert.Contains(t, encoder.fileKeys, "attachment_0")
	})

	t.Run("InputMedia", func(t *testing.T) {
		r := NewRequest("editMessageMedia")
		r.InputMedia("media", makeMedia()[0])

		encoder := &testEncoder{}
		err := r.Encode(encoder)
		require.NoError(t, err)

		assert.Contains(t, encoder.stringKeys, "media")
		assert.Contains(t, encoder.fileKeys, "attachment_0")
	})

	t.Run("MultiItemIndexing", func(t *testing.T) {
		r := NewRequest("sendMediaGroup")

		photo := NewInputMediaPhoto(FileArg{
			Upload: NewInputFileBytes("photo.jpg", []byte("photo")),
		}).AsInputMedia()

		doc := NewInputMediaDocument(FileArg{
			Upload: NewInputFileBytes("doc.pdf", []byte("doc")),
		}).
			WithThumbnail(NewInputFileBytes("thumb.jpg", []byte("thumb"))).
			AsInputMedia()

		r.InputMediaSlice("media", []InputMedia{photo, doc})

		encoder := &testEncoder{}
		err := r.Encode(encoder)
		require.NoError(t, err)

		sort.StringSlice(encoder.fileKeys).Sort()
		assert.Equal(t, []string{
			"attachment_0",
			"attachment_1",
			"attachment_1_thumb",
		}, encoder.fileKeys)
	})

	t.Run("PlainJSON", func(t *testing.T) {
		// This reproduces the bug: using .JSON() without file extraction
		r := NewRequest("sendMediaGroup")
		r.JSON("media", makeMedia())

		encoder := &testEncoder{}
		err := r.Encode(encoder)
		require.Error(t, err, "expected error because FileArg.addr is not set")
		assert.Contains(t, err.Error(), "FileArg is not json serializable")
	})
}

func TestRequest_InputMediaNilVariant(t *testing.T) {
	t.Run("InputMedia", func(t *testing.T) {
		r := NewRequest("editMessageMedia")
		r.InputMedia("media", InputMedia{})

		assert.Empty(t, r.files, "nil variant should not produce file uploads")
	})

	t.Run("InputPaidMediaSlice", func(t *testing.T) {
		r := NewRequest("sendPaidMedia")
		r.InputPaidMediaSlice("media", []InputPaidMedia{{}})

		assert.Empty(t, r.files, "nil variant should not produce file uploads")
	})
}

func TestRequest_InputMediaByRef(t *testing.T) {
	t.Run("InputMediaSlice", func(t *testing.T) {
		r := NewRequest("sendMediaGroup")
		r.InputMediaSlice("media", []InputMedia{
			NewInputMediaPhoto(FileArg{FileID: "existing_file_id"}).AsInputMedia(),
		})

		encoder := &testEncoder{}
		err := r.Encode(encoder)
		require.NoError(t, err)

		assert.Contains(t, encoder.stringKeys, "media")
		assert.Empty(t, encoder.fileKeys, "ref media should not produce file uploads")
	})

	t.Run("InputMedia", func(t *testing.T) {
		r := NewRequest("editMessageMedia")
		r.InputMedia("media", NewInputMediaPhoto(FileArg{URL: "https://example.com/photo.jpg"}).AsInputMedia())

		encoder := &testEncoder{}
		err := r.Encode(encoder)
		require.NoError(t, err)

		assert.Contains(t, encoder.stringKeys, "media")
		assert.Empty(t, encoder.fileKeys, "ref media should not produce file uploads")
	})
}

func TestRequest_InputPaidMediaSlice(t *testing.T) {
	t.Run("Upload", func(t *testing.T) {
		r := NewRequest("sendPaidMedia")
		r.InputPaidMediaSlice("media", []InputPaidMedia{
			NewInputPaidMediaPhoto(FileArg{
				Upload: NewInputFileBytes("photo.jpg", []byte("photo")),
			}).AsInputPaidMedia(),
		})

		encoder := &testEncoder{}
		err := r.Encode(encoder)
		require.NoError(t, err)

		assert.Contains(t, encoder.stringKeys, "media")
		assert.Contains(t, encoder.fileKeys, "attachment_0")
	})

	t.Run("ByRef", func(t *testing.T) {
		r := NewRequest("sendPaidMedia")
		r.InputPaidMediaSlice("media", []InputPaidMedia{
			NewInputPaidMediaPhoto(FileArg{FileID: "existing_file_id"}).AsInputPaidMedia(),
		})

		encoder := &testEncoder{}
		err := r.Encode(encoder)
		require.NoError(t, err)

		assert.Contains(t, encoder.stringKeys, "media")
		assert.Empty(t, encoder.fileKeys)
	})

	t.Run("VideoWithThumbAndCover", func(t *testing.T) {
		media := NewInputPaidMediaVideo(FileArg{
			Upload: NewInputFileBytes("video.mp4", []byte("video")),
		}).
			WithThumbnail(NewInputFileBytes("thumb.jpg", []byte("thumb"))).
			WithCover(FileArg{
				Upload: NewInputFileBytes("cover.jpg", []byte("cover")),
			}).
			AsInputPaidMedia()

		r := NewRequest("sendPaidMedia")
		r.InputPaidMediaSlice("media", []InputPaidMedia{media})

		encoder := &testEncoder{}
		err := r.Encode(encoder)
		require.NoError(t, err)

		sort.StringSlice(encoder.fileKeys).Sort()
		assert.Equal(t, []string{
			"attachment_0",
			"attachment_0_cover",
			"attachment_0_thumb",
		}, encoder.fileKeys)
	})

	t.Run("CoverByRef", func(t *testing.T) {
		media := NewInputPaidMediaVideo(FileArg{
			Upload: NewInputFileBytes("video.mp4", []byte("video")),
		}).
			WithCover(FileArg{FileID: "cover_file_id"}).
			AsInputPaidMedia()

		r := NewRequest("sendPaidMedia")
		r.InputPaidMediaSlice("media", []InputPaidMedia{media})

		encoder := &testEncoder{}
		err := r.Encode(encoder)
		require.NoError(t, err)

		sort.StringSlice(encoder.fileKeys).Sort()
		assert.Equal(t, []string{"attachment_0"}, encoder.fileKeys, "cover by ref should not produce file upload")
	})
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
		require.NoError(t, err)
		assert.JSONEq(t, `{"chat_id":"1","method":"sendMessage","object":"{\"Key\":\"value\"}"}`, string(v))
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
	r.JSON("reply_markup", struct{}{})
	r.InputFile("file", InputFile{})

	assert.True(t, r.Has("chat_id"))
	assert.True(t, r.Has("reply_markup"))
	assert.True(t, r.Has("file"))
	assert.False(t, r.Has("text"))
}

func TestRequest_Get(t *testing.T) {
	r := NewRequest("sendMessage")

	r.String("chat_id", "1")

	v, ok := r.GetArg("chat_id")
	assert.True(t, ok)
	assert.Equal(t, "1", v)

	_, ok = r.GetArg("missing")
	assert.False(t, ok)
}

func TestRequest_GetJSON(t *testing.T) {
	r := NewRequest("sendMessage")

	replyMarkup := InlineKeyboardMarkup{}

	r.JSON("reply_markup", replyMarkup)

	v, ok := r.GetJSON("reply_markup")
	assert.True(t, ok)
	assert.Equal(t, replyMarkup, v)

	_, ok = r.GetJSON("missing")
	assert.False(t, ok)
}
