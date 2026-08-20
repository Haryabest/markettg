package telegram

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	botToken   string
	httpClient *http.Client
}

func NewClient(botToken string) *Client {
	return &Client{
		botToken:   botToken,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

type fileResponse struct {
	OK     bool `json:"ok"`
	Result struct {
		FilePath string `json:"file_path"`
	} `json:"result"`
}

func (c *Client) DownloadFile(ctx context.Context, fileID string) ([]byte, string, error) {
	if c.botToken == "" || strings.TrimSpace(fileID) == "" {
		return nil, "", fmt.Errorf("telegram client not configured")
	}

	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/getFile?file_id=%s", c.botToken, fileID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, "", err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	var parsed fileResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return nil, "", err
	}
	if !parsed.OK || parsed.Result.FilePath == "" {
		return nil, "", fmt.Errorf("telegram getFile failed")
	}

	downloadURL := fmt.Sprintf("https://api.telegram.org/file/bot%s/%s", c.botToken, parsed.Result.FilePath)
	dlReq, err := http.NewRequestWithContext(ctx, http.MethodGet, downloadURL, nil)
	if err != nil {
		return nil, "", err
	}
	dlResp, err := c.httpClient.Do(dlReq)
	if err != nil {
		return nil, "", err
	}
	defer dlResp.Body.Close()
	if dlResp.StatusCode >= 400 {
		return nil, "", fmt.Errorf("telegram file download failed: %d", dlResp.StatusCode)
	}

	data, err := io.ReadAll(dlResp.Body)
	if err != nil {
		return nil, "", err
	}

	ext := ".webp"
	if idx := strings.LastIndex(parsed.Result.FilePath, "."); idx >= 0 {
		ext = parsed.Result.FilePath[idx:]
	}
	return data, ext, nil
}
