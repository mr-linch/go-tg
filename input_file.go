package tg

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
)

// InputFile represents the file that should be uploaded to the telegram.
type InputFile struct {
	// Name of file
	Name string

	// Body of file
	Body io.Reader
}

// WithName creates new InputFile with overridden name.
func (file InputFile) WithName(name string) InputFile {
	file.Name = name
	return file
}

// NewInputFile creates new InputFile with given name and body.
func NewInputFile(name string, body io.Reader) InputFile {
	return InputFile{
		Name: name,
		Body: body,
	}
}

// NewInputFileFromBytes creates new InputFile with given name and bytes slice.
//
// Example:
//   file := NewInputFileBytes("test.txt", []byte("test, test, test..."))
func NewInputFileBytes(name string, body []byte) InputFile {
	return NewInputFile(name, bytes.NewReader(body))
}

// NewInputFileLocal creates the InputFile from provided local file.
// This method just open file by provided path.
// So, you should close it AFTER send.
//
// Example:
//
//   file, close, err := NewInputFileLocal("test.png")
//   if err != nil {
//       return err
//   }
//   defer close()
//
func NewInputFileLocal(path string) (InputFile, func() error, error) {
	file, err := os.Open(path)
	if err != nil {
		return InputFile{}, nil, err
	}

	return NewInputFile(
		filepath.Base(path),
		file,
	), file.Close, nil
}
