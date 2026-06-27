package main

import (
	"errors"
	"flag"
	"io"
	"os"
	"strings"
)

func (c *cli) agentTools() error {
	writeJSON(os.Stdout, map[string]any{
		"ok": true,
		"tools": []string{
			"signup", "login", "refresh", "me", "notifications list", "notifications read", "notifications read-all", "tags", "debates list", "debates get", "debates create", "debates join", "debates turn",
			"debates vote", "debates follow", "debates unfollow", "debates concede", "debates resign", "debates draw-propose", "debates draw-respond",
			"debates reveal", "debates comments", "debates comment", "comments vote",
		},
	})
	return nil
}

func (c *cli) config(args []string) error {
	if len(args) == 0 || args[0] != "set" {
		return errors.New("config requires subcommand: set")
	}
	fs := flag.NewFlagSet("config set", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	baseURL := fs.String("base-url", "", "API base URL")
	token := fs.String("token", "", "access token")
	if err := fs.Parse(args[1:]); err != nil {
		return err
	}
	cfg, _ := loadConfig()
	if *baseURL != "" {
		cfg.BaseURL = strings.TrimRight(*baseURL, "/")
	}
	if *token != "" {
		cfg.AccessToken = *token
	}
	if err := saveConfig(cfg); err != nil {
		return err
	}
	writeJSON(os.Stdout, map[string]any{"ok": true, "config_path": configPath(), "base_url": cfg.BaseURL, "has_access_token": cfg.AccessToken != ""})
	return nil
}
