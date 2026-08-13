package telegram

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

type WebAppUser struct {
	ID           int64  `json:"id"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	Username     string `json:"username"`
	LanguageCode string `json:"language_code"`
	PhotoURL     string `json:"photo_url"`
}

type InitData struct {
	User      *WebAppUser `json:"user"`
	AuthDate  int64       `json:"auth_date"`
	Hash      string      `json:"hash"`
	QueryID   string      `json:"query_id"`
	StartParam string     `json:"start_param"`
}

func ValidateInitData(initData, botToken string, maxAge time.Duration) (*InitData, error) {
	if initData == "" {
		return nil, fmt.Errorf("empty init data")
	}

	values, err := url.ParseQuery(initData)
	if err != nil {
		return nil, fmt.Errorf("parse init data: %w", err)
	}

	receivedHash := values.Get("hash")
	if receivedHash == "" {
		return nil, fmt.Errorf("missing hash")
	}
	values.Del("hash")

	keys := make([]string, 0, len(values))
	for k := range values {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var dataCheck []string
	for _, k := range keys {
		dataCheck = append(dataCheck, k+"="+values.Get(k))
	}
	dataCheckString := strings.Join(dataCheck, "\n")

	secretKey := hmac.New(sha256.New, []byte("WebAppData"))
	secretKey.Write([]byte(botToken))
	secret := secretKey.Sum(nil)

	h := hmac.New(sha256.New, secret)
	h.Write([]byte(dataCheckString))
	expectedHash := hex.EncodeToString(h.Sum(nil))

	if !hmac.Equal([]byte(expectedHash), []byte(receivedHash)) {
		return nil, fmt.Errorf("invalid hash")
	}

	authDateStr := values.Get("auth_date")
	authDate, err := strconv.ParseInt(authDateStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid auth_date")
	}

	if maxAge > 0 && time.Since(time.Unix(authDate, 0)) > maxAge {
		return nil, fmt.Errorf("init data expired")
	}

	result := &InitData{
		AuthDate:   authDate,
		Hash:       receivedHash,
		QueryID:    values.Get("query_id"),
		StartParam: values.Get("start_param"),
	}

	if userStr := values.Get("user"); userStr != "" {
		var user WebAppUser
		if err := json.Unmarshal([]byte(userStr), &user); err != nil {
			return nil, fmt.Errorf("invalid user data")
		}
		result.User = &user
	}

	return result, nil
}
