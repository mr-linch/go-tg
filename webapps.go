package tg

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

// WebAppInitData contains data transferred to the Mini App when it is opened.
// See https://core.telegram.org/bots/webapps#webappinitdata for more information.
type WebAppInitData struct {
	// Optional. A unique identifier for the Mini App session, required for sending messages via the answerWebAppQuery method.
	QueryID string `json:"query_id,omitempty"`

	// Optional. An object containing data about the current user.
	User *WebAppUser `json:"user,omitempty"`

	// Optional. An object containing data about the chat partner of the current user in the chat where the bot was launched via the attachment menu.
	Receiver *WebAppUser `json:"receiver,omitempty"`

	// Optional. An object containing data about the chat where the bot was launched via the attachment menu.
	Chat *WebAppChat `json:"chat,omitempty"`

	// Optional. Type of the chat from which the Mini App was opened.
	ChatType string `json:"chat_type,omitempty"`

	// Optional. Global identifier, uniquely corresponding to the chat from which the Mini App was opened.
	ChatInstance string `json:"chat_instance,omitempty"`

	// Optional. The value of the startattach parameter, passed via link.
	StartParam string `json:"start_param,omitempty"`

	// Optional. Time in seconds, after which a message can be sent via the answerWebAppQuery method.
	CanSendAfter int `json:"can_send_after,omitempty"`

	// Unix time when the form was opened.
	AuthDate int64 `json:"auth_date"`

	// A hash of all passed parameters, which the bot server can use to check their validity.
	Hash string `json:"hash"`

	raw url.Values
}

// AuthDateTime returns time.Time representation of AuthDate field.
func (s *WebAppInitData) AuthDateTime() time.Time {
	return time.Unix(s.AuthDate, 0)
}

// WebAppUser contains the data of the Mini App user.
// See https://core.telegram.org/bots/webapps#webappuser for more information.
type WebAppUser struct {
	// Unique identifier for this user or bot.
	ID UserID `json:"id"`

	// Optional. True, if this user is a bot.
	IsBot bool `json:"is_bot,omitempty"`

	// First name of the user or bot.
	FirstName string `json:"first_name"`

	// Optional. Last name of the user or bot.
	LastName string `json:"last_name,omitempty"`

	// Optional. Username of the user or bot.
	Username string `json:"username,omitempty"`

	// Optional. IETF language tag of the user's language.
	LanguageCode string `json:"language_code,omitempty"`

	// Optional. True, if this user is a Telegram Premium user.
	IsPremium bool `json:"is_premium,omitempty"`

	// Optional. True, if this user added the bot to the attachment menu.
	AddedToAttachmentMenu bool `json:"added_to_attachment_menu,omitempty"`

	// Optional. True, if this user allowed the bot to message them.
	AllowsWriteToPm bool `json:"allows_write_to_pm,omitempty"`

	// Optional. URL of the user's profile photo.
	PhotoURL string `json:"photo_url,omitempty"`
}

// WebAppChat represents a chat in the Mini App context.
// See https://core.telegram.org/bots/webapps#webappchat for more information.
type WebAppChat struct {
	// Unique identifier for this chat.
	ID ChatID `json:"id"`

	// Type of chat.
	Type ChatType `json:"type"`

	// Title of the chat.
	Title string `json:"title"`

	// Optional. Username of the chat.
	Username string `json:"username,omitempty"`

	// Optional. URL of the chat's photo.
	PhotoURL string `json:"photo_url,omitempty"`
}

func getDataCheckString(vs url.Values) string {
	keys := maps.Keys(vs)
	slices.Sort(keys)

	parts := make([]string, len(keys))
	for i, k := range keys {
		parts[i] = k + "=" + vs.Get(k)
	}

	return strings.Join(parts, "\n")
}

// AuthWidget represents Telegram Login Widget data.
//
// See https://core.telegram.org/widgets/login#receiving-authorization-data for more information.
type AuthWidget struct {
	ID        UserID `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name,omitempty"`
	Username  string `json:"username,omitempty"`
	PhotoURL  string `json:"photo_url,omitempty"`
	AuthDate  int64  `json:"auth_date"`
	Hash      string `json:"hash"`
}

// ParseAuthWidgetQuery parses a query string and returns an AuthWidget.
func ParseAuthWidgetQuery(vs url.Values) (*AuthWidget, error) {
	result := &AuthWidget{}

	id, err := strconv.ParseInt(vs.Get("id"), 10, 64)
	if err != nil {
		return nil, fmt.Errorf("parse id %s: %w", vs.Get("id"), err)
	}
	result.ID = UserID(id)

	result.FirstName = vs.Get("first_name")

	authDate, err := strconv.ParseInt(vs.Get("auth_date"), 10, 64)
	if err != nil {
		return nil, fmt.Errorf("parse auth_date %s: %w", vs.Get("auth_date"), err)
	}
	result.AuthDate = authDate

	result.Hash = vs.Get("hash")

	result.LastName = vs.Get("last_name")
	result.Username = vs.Get("username")
	result.PhotoURL = vs.Get("photo_url")

	return result, nil
}

// Query returns a query values for the widget.
func (w AuthWidget) Query() url.Values {
	q := url.Values{}
	q.Set("id", strconv.FormatInt(int64(w.ID), 10))
	q.Set("first_name", w.FirstName)
	q.Set("auth_date", strconv.FormatInt(w.AuthDate, 10))
	q.Set("hash", w.Hash)

	if w.LastName != "" {
		q.Set("last_name", w.LastName)
	}

	if w.Username != "" {
		q.Set("username", w.Username)
	}

	if w.PhotoURL != "" {
		q.Set("photo_url", w.PhotoURL)
	}

	return q
}

// Valid returns true if the signature is valid.
func (w AuthWidget) Valid(token string) bool {
	return subtle.ConstantTimeCompare(
		[]byte(w.Signature(token)),
		[]byte(w.Hash),
	) == 1
}

// Signature returns the signature of the widget data.
func (w AuthWidget) Signature(token string) string {
	vs := w.Query()

	vs.Del("hash")

	data := getDataCheckString(vs)

	key := sha256.Sum256([]byte(token))

	return hex.EncodeToString(getHMAC(data, key[:]))
}

func getHMAC(data string, key []byte) []byte {
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(data))
	return mac.Sum(nil)
}

// AuthDateTime returns the AuthDate as a time.Time.
func (w AuthWidget) AuthDateTime() time.Time {
	return time.Unix(w.AuthDate, 0)
}

// ParseWebAppInitData parses a WebAppInitData from query string.
func ParseWebAppInitData(vs url.Values) (*WebAppInitData, error) {
	result := &WebAppInitData{}

	result.QueryID = vs.Get("query_id")
	if result.QueryID == "" {
		return nil, fmt.Errorf("query_id is empty")
	}

	if vs.Has("user") {
		var user *WebAppUser
		if err := json.Unmarshal([]byte(vs.Get("user")), &user); err != nil {
			return nil, fmt.Errorf("parse user: %w", err)
		}
		result.User = user
	}

	if vs.Has("receiver") {
		var receiver *WebAppUser
		if err := json.Unmarshal([]byte(vs.Get("receiver")), &receiver); err != nil {
			return nil, fmt.Errorf("parse receiver: %w", err)
		}
		result.Receiver = receiver
	}

	if vs.Has("chat") {
		var chat *WebAppChat
		if err := json.Unmarshal([]byte(vs.Get("chat")), &chat); err != nil {
			return nil, fmt.Errorf("parse chat: %w", err)
		}
		result.Chat = chat
	}

	result.StartParam = vs.Get("start_param")

	if vs.Has("can_send_after") {
		canSendAfter, err := strconv.Atoi(vs.Get("can_send_after"))
		if err != nil {
			return nil, fmt.Errorf("parse can_send_after: %w", err)
		}
		result.CanSendAfter = canSendAfter
	}

	authDate, err := strconv.ParseInt(vs.Get("auth_date"), 10, 64)
	if err != nil {
		return nil, fmt.Errorf("parse auth_date %s: %w", vs.Get("auth_date"), err)
	}

	result.AuthDate = authDate

	result.Hash = vs.Get("hash")
	if result.Hash == "" {
		return nil, fmt.Errorf("hash is empty")
	}

	result.raw = vs

	return result, nil
}

// Signature returns the signature of the WebAppInitData.
func (w WebAppInitData) Signature(token string) string {
	vs := w.Query()

	vs.Del("hash")

	data := getDataCheckString(vs)

	key := getHMAC(token, []byte("WebAppData"))

	return hex.EncodeToString(getHMAC(data, key))
}

// Query returns a query values for the WebAppInitData.
func (w WebAppInitData) Query() url.Values {
	vs := make(url.Values, len(w.raw))
	maps.Copy(vs, w.raw)

	vs.Del("hash")

	return vs
}

func (w WebAppInitData) Valid(token string) bool {
	return subtle.ConstantTimeCompare(
		[]byte(w.Signature(token)),
		[]byte(w.Hash),
	) == 1
}
