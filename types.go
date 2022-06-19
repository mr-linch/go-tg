package tg

import (
	"strconv"
)

type ChatID int64

var _ PeerID = (ChatID)(0)

func (id ChatID) PeerID() string {
	return strconv.FormatInt(int64(id), 10)
}

// UserID it's unique identifier for Telegram user or bot.
type UserID int64

var _ PeerID = (UserID)(0)

func (id UserID) PeerID() string {
	return strconv.FormatInt(int64(id), 10)
}

type Username string

func (un Username) PeerID() string {
	return "@" + string(un)
}

type PeerID interface {
	PeerID() string
}

// ParseMode for parsing entities in the message text. See formatting options for more details.
type ParseMode int8

const (
	ParseModeHTML ParseMode = iota + 1
	ParseModeMarkdown
	ParseModeMarkdownV2
)

func (mode ParseMode) String() string {
	switch mode {
	case ParseModeHTML:
		return "HTML"
	case ParseModeMarkdown:
		return "Markdown"
	case ParseModeMarkdownV2:
		return "MarkdownV2"
	default:
		return ""
	}
}

// MessageID it's unique identifier for a message in a chat.
type MessageID int

type FileID string
