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
