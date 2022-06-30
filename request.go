package tg

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// Request is Telegram Bot API request structure.
type Request struct {
	Method string

	json  map[string]any
	args  map[string]string
	files map[string]InputFile

	attachmentIdx int
}

func NewRequest(method string) *Request {
	return &Request{
		Method: method,
		json:   make(map[string]any),
		args:   make(map[string]string),
		files:  make(map[string]InputFile),
	}
}

func (r *Request) JSON(name string, v any) *Request {
	r.json[name] = v
	return r
}

func (r *Request) InputFile(name string, file InputFile) *Request {
	r.files[name] = file
	return r
}

func (r *Request) PeerID(name string, v PeerID) *Request {
	return r.String(name, v.PeerID())
}

func (r *Request) UserID(name string, v UserID) *Request {
	return r.PeerID(name, v)
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

func (r *Request) Int64(name string, value int64) *Request {
	return r.String(name, strconv.FormatInt(value, 10))
}

func (r *Request) Float64(name string, value float64) *Request {
	return r.String(name, strconv.FormatFloat(value, 'f', -1, 64))
}

func (r *Request) ChatID(name string, v ChatID) *Request {
	return r.Int64(name, int64(v))
}

func (r *Request) File(name string, arg FileArg) *Request {
	if arg.FileID != "" {
		return r.String(name, string(arg.FileID))
	} else {
		return r.InputFile(name, arg.Upload)
	}
}

func (r *Request) InputMediaSlice(name string, im []InputMedia) *Request {
	for _, v := range im {
		r.InputMedia(v)
	}

	r.JSON(name, im)

	return r
}

func (r *Request) InputMedia(im InputMedia) *Request {
	media, thumb := im.getMedia()

	id := fmt.Sprintf("attachment_%d", r.attachmentIdx)
	addr := fmt.Sprintf("attach://%s", id)

	if media.getString() == "" {
		r.InputFile(id, media.Upload)
		media.addr = addr
		r.attachmentIdx++
	}

	if thumb != nil {
		thumbID := id + "_thumb"
		r.InputFile(thumbID, *thumb)
		thumb.addr = fmt.Sprintf("attach://%s", thumbID)
	}

	return r
}

func (r *Request) Stringer(name string, v fmt.Stringer) *Request {
	return r.String(name, v.String())
}

// Encode request using encoder.
func (r *Request) Encode(encoder Encoder) error {

	for k, jn := range r.json {
		v, err := json.Marshal(jn)
		if err != nil {
			return fmt.Errorf("failed to marshal %s: %w", k, err)
		}
		r.args[k] = string(v)
	}

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
