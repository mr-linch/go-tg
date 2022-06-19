package tg

import (
	"strconv"
)

// Request is Telegram Bot API request structure.
type Request struct {
	Method string
	args   map[string]string
	files  map[string]InputFile
}

func NewRequest(method string) *Request {
	return &Request{
		Method: method,
		args:   make(map[string]string),
		files:  make(map[string]InputFile),
	}
}

func (r *Request) File(name string, file InputFile) *Request {
	r.files[name] = file
	return r
}

func (r *Request) PeerID(name string, v PeerID) *Request {
	return r.String(name, v.PeerID())
}

func (r *Request) String(name, value string) *Request {
	r.args[name] = value
	return r
}

func (r *Request) Bool(name string, value bool) *Request {
	return r.String(name, strconv.FormatBool(value))
}

func (r *Request) Int(name string, value int) *Request {
	return r.String(name, strconv.Itoa(value))
}

// Encode request using encoder.
func (r *Request) Encode(encoder Encoder) error {

	// add files
	for k, v := range r.files {
		if err := encoder.WriteFile(k, v); err != nil {
			return err
		}
	}

	// add arguments
	for k, v := range r.args {
		if err := encoder.WriteString(k, v); err != nil {
			return err
		}
	}

	return nil
}
