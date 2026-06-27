package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
)

func (c *cli) authAndPrint(method, path string, body any) error {
	var raw json.RawMessage
	if err := c.do(method, path, body, nil, &raw); err != nil {
		return err
	}
	access, refresh := extractAccessToken(raw), c.extractRefreshToken()
	cfg, _ := loadConfig()
	cfg.BaseURL = c.baseURL
	if access != "" {
		cfg.AccessToken = access
	}
	if refresh != "" {
		cfg.RefreshToken = refresh
	}
	_ = saveConfig(cfg)
	writeJSON(os.Stdout, map[string]any{"ok": true, "data": jsonRaw(raw), "saved": map[string]any{"config_path": configPath(), "access_token": access != "", "refresh_token": refresh != ""}})
	return nil
}

func (c *cli) callAndPrint(method, path string, body any, headers map[string]string) error {
	var raw json.RawMessage
	if err := c.do(method, path, body, headers, &raw); err != nil {
		return err
	}
	writeJSON(os.Stdout, map[string]any{"ok": true, "data": jsonRaw(raw)})
	return nil
}

func (c *cli) do(method, path string, body any, headers map[string]string, out *json.RawMessage) error {
	var r io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return err
		}
		r = bytes.NewReader(b)
	}
	req, err := http.NewRequest(method, c.baseURL+path, r)
	if err != nil {
		return err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "respondcli/0.1 agent-friendly")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if len(bytes.TrimSpace(b)) == 0 {
		b = []byte("null")
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var ae apiError
		if json.Unmarshal(b, &ae) == nil && ae.Error.Code != "" {
			return fmt.Errorf("api %s: %s", ae.Error.Code, ae.Error.Message)
		}
		return fmt.Errorf("api http %d: %s", resp.StatusCode, string(b))
	}
	*out = json.RawMessage(b)
	return nil
}

func (c *cli) extractRefreshToken() string {
	u, err := url.Parse(c.baseURL)
	if err != nil {
		return ""
	}
	for _, ck := range c.jar.Cookies(u) {
		if ck.Name == "refresh_token" {
			return ck.Value
		}
	}
	return ""
}
