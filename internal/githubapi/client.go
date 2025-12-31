package githubapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/dennislee928/github-recap-2025/internal/config"
)

type Client struct {
	cfg config.Config
	http *http.Client
}

func New(cfg config.Config) *Client {
	return &Client{
		cfg: cfg,
		http: &http.Client{Timeout: 45 * time.Second},
	}
}

func (c *Client) doGraphQL(ctx context.Context, query string, variables map[string]any, out any) error {
	payload := map[string]any{
		"query": query,
		"variables": variables,
	}
	b, _ := json.Marshal(payload)

	req, err := http.NewRequestWithContext(ctx, "POST", c.cfg.GraphQLEnd, bytes.NewReader(b))
	if err != nil { return err }
	req.Header.Set("Authorization", "bearer "+c.cfg.Token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil { return err }
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return fmt.Errorf("graphql http %d: %s", resp.StatusCode, string(body))
	}

	var envelope struct{
		Data json.RawMessage `json:"data"`
		Errors []struct{
			Message string `json:"message"`
		} `json:"errors"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		return fmt.Errorf("graphql decode envelope: %w", err)
	}
	if len(envelope.Errors) > 0 {
		return fmt.Errorf("graphql errors: %v", envelope.Errors[0].Message)
	}
	if err := json.Unmarshal(envelope.Data, out); err != nil {
		return fmt.Errorf("graphql decode data: %w", err)
	}
	return nil
}
