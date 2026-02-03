package tg

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/url"
	"strings"
)

// Encoder represents request encoder.
type Encoder interface {
	// Writes string argument k to encoder.
	WriteString(k string, v string) error

	// Write files argument k to encoder.
	WriteFile(k string, file InputFile) error
}

type httpEncoder interface {
	Encoder
	io.Closer
	ContentType() string
}

type urlEncodedEncoder struct {
	dst   io.Writer
	pairs int
}

var _ httpEncoder = (*urlEncodedEncoder)(nil)

func newURLEncodedEncoder(dst io.Writer) *urlEncodedEncoder {
	return &urlEncodedEncoder{dst: dst}
}

func (encoder *urlEncodedEncoder) WriteString(k, v string) error {
	buf := strings.Builder{}

	if encoder.pairs > 0 {
		buf.WriteByte('&')
	}

	buf.WriteString(url.QueryEscape(k))
	buf.WriteByte('=')
	buf.WriteString(url.QueryEscape(v))

	_, err := io.WriteString(encoder.dst, buf.String())
	if err != nil {
		return err
	}

	encoder.pairs++

	return nil
}

func (encoder *urlEncodedEncoder) WriteFile(k string, file InputFile) error {
	return errors.New("urlEncodedEncoder doesn't support files")
}

func (encoder *urlEncodedEncoder) ContentType() string {
	return "application/x-www-form-urlencoded"
}

func (encoder *urlEncodedEncoder) Close() error {
	return nil
}

// multipartEncoder encodes the request using multipart encoding.
type multipartEncoder struct {
	w *multipart.Writer
}

// newMultipartEncoder creates multipart encoder.
func newMultipartEncoder(writer io.Writer) *multipartEncoder {
	return &multipartEncoder{
		w: multipart.NewWriter(writer),
	}
}

// AddString encodes string value
func (enc *multipartEncoder) WriteString(k, v string) error {
	return enc.w.WriteField(k, v)
}

// AddFile encodes file value.
func (enc *multipartEncoder) WriteFile(k string, file InputFile) error {
	writer, err := enc.w.CreateFormFile(k, file.Name)
	if err != nil {
		return fmt.Errorf("create form file '%s': %w", k, err)
	}

	if _, err := io.Copy(writer, file.Body); err != nil {
		return fmt.Errorf("copy to form file '%s': %w", k, err)
	}

	return nil
}

// ContentType returns HTTP request content type.
func (enc *multipartEncoder) ContentType() string {
	return enc.w.FormDataContentType()
}

// Close multipart encoder.
func (enc *multipartEncoder) Close() error {
	return enc.w.Close()
}
