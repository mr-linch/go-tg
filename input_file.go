package tg

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

// InputFile represents the file that should be uploaded to the telegram.
type InputFile struct {
	// Name of file
	Name string

	// Body of file
	Body io.Reader

	addr string
}

func (file *InputFile) MarshalJSON() ([]byte, error) {
	if file.addr != "" {
		return json.Marshal(file.addr)
	}

	return nil, fmt.Errorf("can't marshal InputFile without address")
}

// WithName creates new InputFile with overridden name.
func (file InputFile) WithName(name string) InputFile {
	file.Name = name
	return file
}

// Ptr returns pointer to InputFile. Helper method.
func (file InputFile) Ptr() *InputFile {
	return &file
}

// Close closes body, if body impliments io.Closer.
func (file InputFile) Close() error {
	closer, ok := file.Body.(io.Closer)
	if ok {
		return closer.Close()
	}
	return nil
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

// NewInputFileLocal creates the InputFile from provided local file path.
// This method just open file by provided path.
// So, you should close it AFTER send.
//
// Example:
//
//   file, err := NewInputFileLocal("test.png")
//   if err != nil {
//       return err
//   }
//   defer  close()
//
func NewInputFileLocal(path string) (InputFile, error) {
	file, err := os.Open(path)
	if err != nil {
		return InputFile{}, err
	}

	return NewInputFile(
		filepath.Base(path),
		file,
	), nil
}

// NewInputFileFS creates the InputFile from provided FS and file path.
//
// Usage:
//  //go:embed assets/*
//  var assets embed.FS
//  file, err := NewInputFileFS(assets, "images/test.png")
//  if err != nil {
//    return err
//  }
//  defer file.Close()
func NewInputFileFS(fs fs.FS, path string) (InputFile, error) {
	file, err := fs.Open(path)
	if err != nil {
		return InputFile{}, fmt.Errorf("open file: %w", err)
	}

	return NewInputFile(
		filepath.Base(path),
		file,
	), nil
}
