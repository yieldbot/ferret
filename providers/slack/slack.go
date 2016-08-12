/*
 * Ferret
 * Copyright (c) 2016 Yieldbot, Inc.
 * For the full copyright and license information, please view the LICENSE.txt file.
 */

// Package slack implements Slack provider
package slack

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/context"
	"golang.org/x/net/context/ctxhttp"
)

// Register registers the provider
func Register(f func(provider interface{}) error) {
	var p = Provider{
		name:     "slack",
		title:    "Slack",
		priority: 500,
		url:      "https://slack.com/api",
		token:    os.Getenv("FERRET_SLACK_TOKEN"),
	}
	if p.token != "" {
		p.enabled = true
	}

	if err := f(&p); err != nil {
		panic(err)
	}
}

// Provider represents the provider
type Provider struct {
	enabled  bool
	name     string
	title    string
	priority int64
	url      string
	token    string
}

// Search makes a search
func (provider *Provider) Search(ctx context.Context, args map[string]interface{}) ([]map[string]interface{}, error) {

	var results = []map[string]interface{}{}
	page, ok := args["page"].(int)
	if page < 1 || !ok {
		page = 1
	}
	limit, ok := args["limit"].(int)
	if limit < 1 || !ok {
		limit = 10
	}
	keyword, ok := args["keyword"].(string)

	var u = fmt.Sprintf("%s/search.all?page=%d&count=%d&query=%s&token=%s", provider.url, page, limit, url.QueryEscape(keyword), provider.token)
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, errors.New("failed to prepare request. Error: " + err.Error())
	}

	res, err := ctxhttp.Do(ctx, nil, req)
	if err != nil {
		return nil, err
	} else if res.StatusCode < 200 || res.StatusCode > 299 {
		return nil, errors.New("bad response: " + fmt.Sprintf("%d", res.StatusCode))
	}
	defer res.Body.Close()
	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	var sr SearchResult
	if err := json.Unmarshal(data, &sr); err != nil {
		return nil, errors.New("failed to unmarshal JSON data. Error: " + err.Error())
	}
	if sr.Messages != nil {
		for _, v := range sr.Messages.Matches {
			// TODO: Improve partial text (i.e. ... keyword ...)
			d := strings.TrimSpace(v.Text)
			if len(d) > 255 {
				d = d[0:252] + "..."
			}

			var t time.Time
			if ts, err := strconv.ParseFloat(v.Ts, 64); err == nil {
				t = time.Unix(int64(ts), 0)
			}

			ri := map[string]interface{}{
				"Link":        v.Permalink,
				"Title":       fmt.Sprintf("@%s in #%s", v.Username, v.Channel.Name),
				"Description": d,
				"Date":        t,
			}
			results = append(results, ri)
		}
	}

	return results, err
}

// SearchResult represents the structure of the search result
type SearchResult struct {
	Ok       bool        `json:"ok"`
	Query    string      `json:"query"`
	Messages *SRMessages `json:"messages"`
}

// SRMessages represents the structure of the search result messages
type SRMessages struct {
	Total   int           `json:"total"`
	Path    string        `json:"path"`
	Matches []*SRMMatches `json:"matches"`
}

// SRMMatches represents the structure of the search result messages matches
type SRMMatches struct {
	Type      string       `json:"type"`
	Username  string       `json:"username"`
	Text      string       `json:"text"`
	Permalink string       `json:"permalink"`
	Ts        string       `json:"ts"`
	Channel   *SRMMChannel `json:"channel"`
}

// SRMMChannel represents the structure of the search result messages matches channel field
type SRMMChannel struct {
	Name string `json:"name"`
}
