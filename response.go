package tg

import "encoding/json"

// Response is Telegram Bot API response structure.
type Response struct {
	// If equals true, the request was successful
	// and the result of the query can be found in the 'result' field
	Ok bool `json:"ok"`

	// A human-readable description of the result.
	// Empty if Ok is true.
	// Containes error message if Ok is false.
	Description string `json:"description"`

	// Optional. The result of the request
	Result json.RawMessage `json:"result"`

	// Optional. ErrorCode is the error code returned by Telegram Bot API.
	ErrorCode int `json:"error_code"`

	// Optional. Parameters describes why a request was unsuccessful in some cases.
	// Parameters *ResponseParameters `json:"parameters"`

	// HTTP response status code.
	StatusCode int
}
