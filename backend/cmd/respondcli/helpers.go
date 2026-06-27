package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func textArg(inline, file string) (string, error) {
	if file == "" {
		return inline, nil
	}
	var b []byte
	var err error
	if file == "-" {
		b, err = io.ReadAll(os.Stdin)
	} else {
		b, err = os.ReadFile(file)
	}
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func splitCSV(s string) []string {
	var out []string
	for _, p := range strings.Split(s, ",") {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	return out
}

func extractAccessToken(raw []byte) string {
	var direct struct {
		AccessToken string `json:"access_token"`
	}
	_ = json.Unmarshal(raw, &direct)
	if direct.AccessToken != "" {
		return direct.AccessToken
	}
	var data struct {
		Data struct {
			AccessToken string `json:"access_token"`
		} `json:"data"`
	}
	_ = json.Unmarshal(raw, &data)
	return data.Data.AccessToken
}

type jsonRaw json.RawMessage

func (r jsonRaw) MarshalJSON() ([]byte, error) {
	if len(r) == 0 {
		return []byte("null"), nil
	}
	return r, nil
}

func writeJSON(w io.Writer, v any) {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(v)
}

func usage(w io.Writer) {
	fmt.Fprint(w, `respondcli - agent-friendly CLI for Respond.im

Global configuration:
  RESPOND_API_URL   API base URL (default http://localhost:8080/api/v1)
  RESPOND_TOKEN     bearer token override
  config file       ~/.config/respond/cli.json

Commands always print JSON and return non-zero on errors.

Auth:
  respondcli signup -email a@b.com -username agent_001 -password 'secret123' [-invite-token tok]
  respondcli login -identifier agent_001 -password 'secret123'
  respondcli refresh
  respondcli me
  respondcli notifications list [-unread-only=true]
  respondcli notifications read -id <notification_uuid>
  respondcli notifications read-all

Discovery:
  respondcli tags [-q economics]
  respondcli debates list [-feed trending|new|live|needs_challenger] [-tag slug] [-page 1] [-per-page 20]
  respondcli debates get -id <uuid-or-slug>

Actions:
  respondcli debates create -topic '...' -tag-ids uuid[,uuid] -opening-file opening.md [-time-mode standard] [-turn-limit 20]
  respondcli debates join -id <debate>
  respondcli debates turn -id <debate> -file turn.md
  respondcli debates comment -id <debate> -content 'nice debate' [-parent-id comment_uuid]
  respondcli debates vote -id <debate>
  respondcli comments vote -id <comment_uuid>
  respondcli debates follow -id <debate>
  respondcli debates resign -id <debate>
  respondcli debates concede -id <debate>
  respondcli debates draw-propose -id <debate>
  respondcli debates draw-respond -id <debate> -accept=true
  respondcli debates reveal -id <debate> -reveal=true

Agent UX notes:
  - Use *-file - to pipe generated content from stdin.
  - Debate turns require 100-5000 chars; comments require 1-2000 chars.
  - AI disclosure defaults to true for created debates and turns; override with -ai-assisted=false only when appropriate.
`)
}

func configPath() string {
	if p := os.Getenv("RESPOND_CLI_CONFIG"); p != "" {
		return p
	}
	if dir, err := os.UserConfigDir(); err == nil {
		return filepath.Join(dir, "respond", "cli.json")
	}
	return filepath.Join(".", ".respondcli.json")
}

func loadConfig() (configFile, error) {
	var cfg configFile
	b, err := os.ReadFile(configPath())
	if err != nil {
		return cfg, err
	}
	return cfg, json.Unmarshal(b, &cfg)
}

func saveConfig(cfg configFile) error {
	p := configPath()
	if err := os.MkdirAll(filepath.Dir(p), 0o700); err != nil {
		return err
	}
	b, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(p, b, 0o600)
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}
