package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"
)

var (
	ErrTelegramSignature = errors.New("invalid telegram signature")
	ErrTelegramExpired   = errors.New("telegram auth data expired")
)

// TelegramAuthData is the payload sent by the Telegram Login Widget.
// https://core.telegram.org/widgets/login#checking-authorization
type TelegramAuthData struct {
	ID        int64  `json:"id"`
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
	Username  string `json:"username,omitempty"`
	PhotoURL  string `json:"photo_url,omitempty"`
	AuthDate  int64  `json:"auth_date"`
	Hash      string `json:"hash"`
}

// telegramVerifier validates Login Widget payloads against the bot token.
type telegramVerifier struct {
	botToken string
	maxAge   time.Duration
}

func newTelegramVerifier(botToken string, maxAge time.Duration) *telegramVerifier {
	return &telegramVerifier{botToken: botToken, maxAge: maxAge}
}

// verify checks the HMAC signature and freshness of the auth data.
// `now` is injected for deterministic testing.
func (v *telegramVerifier) verify(d TelegramAuthData, now time.Time) error {
	if v.botToken == "" {
		return errors.New("telegram bot token not configured")
	}

	// secret_key = SHA256(bot_token)
	secret := sha256.Sum256([]byte(v.botToken))

	// data_check_string = all fields except `hash`, sorted by key, joined by "\n".
	fields := map[string]string{
		"id":        fmt.Sprintf("%d", d.ID),
		"auth_date": fmt.Sprintf("%d", d.AuthDate),
	}
	if d.FirstName != "" {
		fields["first_name"] = d.FirstName
	}
	if d.LastName != "" {
		fields["last_name"] = d.LastName
	}
	if d.Username != "" {
		fields["username"] = d.Username
	}
	if d.PhotoURL != "" {
		fields["photo_url"] = d.PhotoURL
	}

	keys := make([]string, 0, len(fields))
	for k := range fields {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	pairs := make([]string, 0, len(keys))
	for _, k := range keys {
		pairs = append(pairs, k+"="+fields[k])
	}
	checkString := strings.Join(pairs, "\n")

	mac := hmac.New(sha256.New, secret[:])
	mac.Write([]byte(checkString))
	expected := hex.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(expected), []byte(strings.ToLower(d.Hash))) {
		return ErrTelegramSignature
	}

	// Reject stale payloads to prevent replay.
	if v.maxAge > 0 {
		age := now.Sub(time.Unix(d.AuthDate, 0))
		if age > v.maxAge || age < -v.maxAge {
			return ErrTelegramExpired
		}
	}
	return nil
}
