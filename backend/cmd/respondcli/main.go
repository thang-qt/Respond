package main

import (
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strings"
	"time"
)

const defaultBaseURL = "http://localhost:8080/api/v1"

type configFile struct {
	BaseURL      string `json:"base_url"`
	AccessToken  string `json:"access_token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
}

type cli struct {
	baseURL string
	token   string
	jar     http.CookieJar
	client  *http.Client
	out     string
}

type apiError struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		writeJSON(os.Stderr, map[string]any{"ok": false, "error": err.Error()})
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 || args[0] == "help" || args[0] == "--help" || args[0] == "-h" {
		usage(os.Stdout)
		return nil
	}

	cfg, _ := loadConfig()
	baseURL := firstNonEmpty(os.Getenv("RESPOND_API_URL"), cfg.BaseURL, defaultBaseURL)
	token := firstNonEmpty(os.Getenv("RESPOND_TOKEN"), cfg.AccessToken)

	jar, _ := cookiejar.New(nil)
	if cfg.RefreshToken != "" {
		if u, err := url.Parse(baseURL); err == nil {
			jar.SetCookies(u, []*http.Cookie{{Name: "refresh_token", Value: cfg.RefreshToken}})
		}
	}

	c := &cli{baseURL: strings.TrimRight(baseURL, "/"), token: token, jar: jar, client: &http.Client{Timeout: 30 * time.Second, Jar: jar}, out: "json"}

	switch args[0] {
	case "signup":
		return c.signup(args[1:])
	case "login":
		return c.login(args[1:])
	case "refresh":
		return c.refresh()
	case "logout":
		return c.callAndPrint("POST", "/auth/logout", nil, nil)
	case "me":
		return c.callAndPrint("GET", "/users/me", nil, nil)
	case "notifications":
		return c.notifications(args[1:])
	case "tags":
		return c.tags(args[1:])
	case "debates":
		return c.debates(args[1:])
	case "agent-tools":
		return c.agentTools()
	case "comments":
		return c.comments(args[1:])
	case "config":
		return c.config(args[1:])
	default:
		return fmt.Errorf("unknown command %q", args[0])
	}
}
