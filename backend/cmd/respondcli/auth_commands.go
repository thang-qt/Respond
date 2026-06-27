package main

import (
	"encoding/json"
	"errors"
	"flag"
	"io"
	"os"
)

func (c *cli) signup(args []string) error {
	fs := flag.NewFlagSet("signup", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	email := fs.String("email", "", "email")
	username := fs.String("username", "", "username")
	password := fs.String("password", "", "password")
	invite := fs.String("invite-token", "", "invite token")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *email == "" || *username == "" || *password == "" {
		return errors.New("signup requires -email, -username, and -password")
	}
	body := map[string]any{"email": *email, "username": *username, "password": *password}
	if *invite != "" {
		body["invite_token"] = *invite
	}
	return c.authAndPrint("POST", "/auth/signup", body)
}

func (c *cli) login(args []string) error {
	fs := flag.NewFlagSet("login", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	identifier := fs.String("identifier", "", "email or username")
	password := fs.String("password", "", "password")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *identifier == "" || *password == "" {
		return errors.New("login requires -identifier and -password")
	}
	return c.authAndPrint("POST", "/auth/login", map[string]any{"identifier": *identifier, "password": *password})
}

func (c *cli) refresh() error {
	var raw json.RawMessage
	if err := c.do("POST", "/auth/refresh", nil, nil, &raw); err != nil {
		return err
	}
	var envelope struct {
		AccessToken string `json:"access_token"`
	}
	_ = json.Unmarshal(raw, &envelope)
	var wrapped struct {
		Data struct {
			AccessToken string `json:"access_token"`
		} `json:"data"`
	}
	_ = json.Unmarshal(raw, &wrapped)
	if envelope.AccessToken == "" {
		envelope.AccessToken = wrapped.Data.AccessToken
	}
	if envelope.AccessToken != "" {
		cfg, _ := loadConfig()
		cfg.BaseURL = c.baseURL
		cfg.AccessToken = envelope.AccessToken
		_ = saveConfig(cfg)
	}
	writeJSON(os.Stdout, map[string]any{"ok": true, "data": jsonRaw(raw)})
	return nil
}
