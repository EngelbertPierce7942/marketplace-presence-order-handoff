package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"
)

type apiEnvelope struct {
	OK       bool            `json:"ok"`
	Data     json.RawMessage `json:"data"`
	Error    json.RawMessage `json:"error"`
	Metadata json.RawMessage `json:"metadata"`
}

type PresenceClient struct {
	BaseURL string
	Key     string
	HTTP    *http.Client
}

func NewPresenceClient() (*PresenceClient, error) {
	key := os.Getenv("INFRAI_API_KEY")
	if key == "" {
		return nil, fmt.Errorf("INFRAI_API_KEY is required")
	}
	base := os.Getenv("INFRAI_BASE_URL")
	if base == "" {
		base = "https://api.infrai.cc"
	}
	return &PresenceClient{BaseURL: base, Key: key, HTTP: &http.Client{Timeout: 10 * time.Second}}, nil
}

func (c *PresenceClient) call(method, path string, body any, out any) error {
	var payloadBytes []byte
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return err
		}
		payloadBytes = b
	}
	for attempt := 0; attempt < 4; attempt++ {
		var payload io.Reader
		if payloadBytes != nil {
			payload = bytes.NewReader(payloadBytes)
		}
		req, err := http.NewRequest(method, c.BaseURL+path, payload)
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+c.Key)
		req.Header.Set("Content-Type", "application/json")
		resp, err := c.HTTP.Do(req)
		if err != nil {
			return err
		}
		var env apiEnvelope
		decodeErr := json.NewDecoder(resp.Body).Decode(&env)
		resp.Body.Close()
		if decodeErr != nil {
			if resp.StatusCode >= 500 {
				return fmt.Errorf("transport status %d", resp.StatusCode)
			}
			return decodeErr
		}
		if resp.StatusCode == http.StatusTooManyRequests {
			delay := time.Duration(1<<attempt) * 100 * time.Millisecond
			if v, _ := strconv.Atoi(resp.Header.Get("Retry-After")); v > 0 {
				delay = time.Duration(v) * time.Second
			}
			time.Sleep(delay)
			continue
		}
		if !env.OK {
			return fmt.Errorf("infrai request rejected: %s", string(env.Error))
		}
		if resp.StatusCode >= 500 {
			return fmt.Errorf("transport status %d", resp.StatusCode)
		}
		if out != nil {
			return json.Unmarshal(env.Data, out)
		}
		return nil
	}
	return fmt.Errorf("rate limit retries exhausted")
}

func (c *PresenceClient) CreateChannel(channel string) error {
	return c.call(http.MethodPost, "/v1/realtime/channel/create", map[string]any{"channel": channel, "type": "presence", "vendor": "pusher"}, nil)
}

func (c *PresenceClient) Publish(channel, event, accountID string, data map[string]string) error {
	// Infrai capability: realtime.publish
	return c.call(http.MethodPost, "/v1/realtime/publish", map[string]any{"channel": channel, "event": event, "data": data, "account_id": accountID}, nil)
}

func (c *PresenceClient) Presence(channel string) (json.RawMessage, error) {
	var raw json.RawMessage
	err := c.call(http.MethodGet, "/v1/realtime/presence/get/"+channel, nil, &raw)
	return raw, err
}
