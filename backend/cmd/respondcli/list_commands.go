package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"net/url"
	"strconv"
)

func (c *cli) notifications(args []string) error {
	if len(args) == 0 {
		return errors.New("notifications requires subcommand: list|read|read-all")
	}
	switch args[0] {
	case "list":
		fs := flag.NewFlagSet("notifications list", flag.ContinueOnError)
		fs.SetOutput(io.Discard)
		unreadOnly := fs.Bool("unread-only", true, "only unread notifications")
		page := fs.Int("page", 1, "page")
		perPage := fs.Int("per-page", 20, "per page")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		v := url.Values{"unread_only": {strconv.FormatBool(*unreadOnly)}, "page": {strconv.Itoa(*page)}, "per_page": {strconv.Itoa(*perPage)}}
		return c.callAndPrint("GET", "/users/me/notifications?"+v.Encode(), nil, nil)
	case "read":
		fs := flag.NewFlagSet("notifications read", flag.ContinueOnError)
		fs.SetOutput(io.Discard)
		id := fs.String("id", "", "notification id")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if *id == "" {
			return errors.New("notifications read requires -id")
		}
		return c.callAndPrint("PUT", "/users/me/notifications/"+url.PathEscape(*id)+"/read", nil, nil)
	case "read-all":
		return c.callAndPrint("PUT", "/users/me/notifications/read-all", nil, nil)
	default:
		return fmt.Errorf("unknown notifications subcommand %q", args[0])
	}
}

func (c *cli) tags(args []string) error {
	fs := flag.NewFlagSet("tags", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	q := fs.String("q", "", "search query")
	if err := fs.Parse(args); err != nil {
		return err
	}
	path := "/tags"
	if *q != "" {
		path = "/tags/search?q=" + url.QueryEscape(*q)
	}
	return c.callAndPrint("GET", path, nil, nil)
}
