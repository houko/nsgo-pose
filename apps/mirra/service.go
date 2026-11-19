package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// serviceKind is where pictures are made: a cloud API that needs a key, or a server the user runs, which may not.
type serviceKind string

const (
	serviceCloud serviceKind = "cloud"
	serviceLocal serviceKind = "local"
)

const defaultLocalURL = "http://127.0.0.1:8080/v1"

// Connection test states.
type connState int

const (
	connUntested connState = iota
	connTesting
	connOK
	connFailed
)

// CheckConnection asks cfg's API for its models, which every OpenAI-compatible server answers, to check the address and the key before generating.
func CheckConnection(ctx context.Context, client *http.Client, cfg Config) error {
	endpoint := strings.TrimRight(or(cfg.BaseURL, defaultBaseURL), "/") + "/models"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}
	if key := strings.TrimSpace(cfg.APIKey); key != "" {
		req.Header.Set("Authorization", "Bearer "+key)
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 == 2 {
		return nil
	}
	var out struct {
		Error *struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 64<<10))
	if json.Unmarshal(body, &out) == nil && out.Error != nil && out.Error.Message != "" {
		return fmt.Errorf("%s: %s", resp.Status, out.Error.Message)
	}
	return fmt.Errorf("%s", resp.Status)
}
