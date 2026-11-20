package main

import (
	"context"
	"net/http"
	"strings"
)

// Connection test states.
type connState int

const (
	connUntested connState = iota
	connTesting
	connOK
	connFailed
)

// CheckConnection checks the address and the key before generating, with a request the provider answers cheaply: its list of models where it has one, OpenRouter's key, whose models anyone may list, and otherwise an empty generation, which a valid key gets a 400 for and an invalid one a 401 or 403.
func CheckConnection(ctx context.Context, client *http.Client, cfg Config) error {
	var req *http.Request
	var err error
	emptyPost := false
	switch cfg.Protocol {
	case protoSeedream:
		req, err = http.NewRequestWithContext(ctx, http.MethodPost, endpoint(cfg, "/images/generations"), strings.NewReader("{}"))
		emptyPost = true
	case protoDashScope:
		req, err = http.NewRequestWithContext(ctx, http.MethodPost, endpoint(cfg, dashScopePath), strings.NewReader("{}"))
		emptyPost = true
	case protoOpenRouter:
		req, err = http.NewRequestWithContext(ctx, http.MethodGet, endpoint(cfg, "/key"), nil)
	default:
		req, err = http.NewRequestWithContext(ctx, http.MethodGet, endpoint(cfg, "/models"), nil)
	}
	if err != nil {
		return err
	}
	if emptyPost {
		req.Header.Set("Content-Type", "application/json")
	}
	key := strings.TrimSpace(cfg.APIKey)
	switch {
	case key == "":
	case cfg.Protocol == protoGemini:
		req.Header.Set("x-goog-api-key", key)
	case cfg.Protocol == protoAzure:
		req.Header.Set("api-key", key)
	default:
		req.Header.Set("Authorization", "Bearer "+key)
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 == 2 || emptyPost && resp.StatusCode == http.StatusBadRequest {
		return nil
	}
	return apiError(resp)
}
