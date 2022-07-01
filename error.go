package tg

import (
	"fmt"
	"strings"
)

// Error is Telegram Bot API error structure.
type Error struct {
	Code       int                 `json:"code"`
	Message    string              `json:"message"`
	Parameters *ResponseParameters `json:"parameters"`
}

func (err *Error) Error() string {
	if err.Parameters == nil {
		return fmt.Sprintf("%d: %s", err.Code, err.Message)
	} else {
		return fmt.Sprintf("%d: %s (%+v)", err.Code, err.Message, *err.Parameters)
	}
}

func (err *Error) Contains(v string) bool {
	return strings.Contains(strings.ToLower(err.Message), v)
}
