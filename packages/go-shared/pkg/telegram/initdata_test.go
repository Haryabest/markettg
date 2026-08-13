package telegram_test

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/markettg/markettg/packages/go-shared/pkg/telegram"
)

func signInitData(data url.Values, botToken string) string {
	hash := data.Get("hash")
	data.Del("hash")
	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var parts []string
	for _, k := range keys {
		parts = append(parts, k+"="+data.Get(k))
	}
	secretKey := hmac.New(sha256.New, []byte("WebAppData"))
	secretKey.Write([]byte(botToken))
	secret := secretKey.Sum(nil)
	h := hmac.New(sha256.New, secret)
	h.Write([]byte(strings.Join(parts, "\n")))
	return hex.EncodeToString(h.Sum(nil))
}

func TestValidateInitData(t *testing.T) {
	botToken := "123456:ABC-DEF"
	userJSON := `{"id":12345,"first_name":"Test","username":"testuser"}`
	values := url.Values{}
	values.Set("user", userJSON)
	values.Set("auth_date", fmt.Sprintf("%d", time.Now().Unix()))
	values.Set("hash", signInitData(values, botToken))

	initData := values.Encode()
	result, err := telegram.ValidateInitData(initData, botToken, 24*time.Hour)
	if err != nil {
		t.Fatalf("expected valid init data: %v", err)
	}
	if result.User == nil || result.User.ID != 12345 {
		t.Fatalf("unexpected user: %+v", result.User)
	}
}

func TestValidateInitDataInvalidHash(t *testing.T) {
	_, err := telegram.ValidateInitData("user=test&auth_date=1&hash=invalid", "token", time.Hour)
	if err == nil {
		t.Fatal("expected error for invalid hash")
	}
}
