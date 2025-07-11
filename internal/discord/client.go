package discord

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"

	"enhanced-gateway-scraper/pkg/types"
	"log"
)

type Client struct {
	WebhookURL string
}

func NewClient(config *types.NotificationConfig) *Client {
	if !config.Enabled || config.WebhookURL == "" {
		return nil
	}
	return &Client{
		WebhookURL: config.WebhookURL,
	}
}

func (c *Client) SendNotification(content string) error {
	if c == nil || c.WebhookURL == "" {
		return nil
	}

	payload := map[string]string{"content": content}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", c.WebhookURL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		log.Printf("Failed to send notification, status: %d", resp.StatusCode)
	}

	return nil
}

