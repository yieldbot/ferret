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
)

// Register registers the provider
func Register(f func(name string, provider interface{}) error) {
	// Init the provider
	var p = Provider{
		url:   "https://slack.com/api",
		token: os.Getenv("FERRET_SLACK_TOKEN"),
	}

	// Register the provider
	if err := f("slack", &p); err != nil {
		panic(err)
	}
}

// Provider represents the provider
type Provider struct {
	url   string
	token string
}

// SearchResult represent the structure of the search result
type SearchResult struct {
	Ok       bool                  `json:"ok"`
	Query    string                `json:"query"`
	Messages *SearchResultMessages `json:"messages"`
}

// SearchResultMessages represent the structure of the search result messages
type SearchResultMessages struct {
	Total   int                            `json:"total"`
	Path    string                         `json:"path"`
	Matches []*SearchResultMessagesMatches `json:"matches"`
}

// SearchResultMessagesMatches represent the structure of the search result messages matches
type SearchResultMessagesMatches struct {
	Type      string                              `json:"type"`
	Username  string                              `json:"username"`
	Text      string                              `json:"text"`
	Permalink string                              `json:"permalink"`
	Ts        string                              `json:"ts"`
	Channel   *SearchResultMessagesMatchesChannel `json:"channel"`
}

// SearchResultMessagesMatchesChannel represent the structure of the search result messages matches channel field
type SearchResultMessagesMatchesChannel struct {
	Name string `json:"name"`
}

// Search makes a search
func (provider *Provider) Search(ctx context.Context, args map[string]interface{}) ([]map[string]interface{}, error) {

	var results = []map[string]interface{}{}
	page, ok := args["page"].(int)
	if page < 1 || !ok {
		page = 1
	}
	keyword, ok := args["keyword"].(string)

	var u = fmt.Sprintf("%s/search.all?page=%d&count=10&query=%s&token=%s", provider.url, page, url.QueryEscape(keyword), provider.token)
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, errors.New("failed to prepare request. Error: " + err.Error())
	}

	err = DoWithContext(ctx, nil, req, func(res *http.Response, err error) error {

		if err != nil {
			return errors.New("failed to fetch data. Error: " + err.Error())
		} else if res.StatusCode < 200 || res.StatusCode > 299 {
			return errors.New("bad response: " + fmt.Sprintf("%d", res.StatusCode))
		}
		defer res.Body.Close()
		data, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return err
		}
		var sr SearchResult
		if err = json.Unmarshal(data, &sr); err != nil {
			return errors.New("failed to unmarshal JSON data. Error: " + err.Error())
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

		return nil
	})

	return results, err
}

// DoWithContext makes a HTTP request with the given context
func DoWithContext(ctx context.Context, client *http.Client, req *http.Request, f func(*http.Response, error) error) error {
	tr := &http.Transport{}
	if client == nil {
		client = &http.Client{Transport: tr}
	}
	c := make(chan error, 1)
	go func() {
		c <- f(client.Do(req))
	}()
	select {
	case <-ctx.Done():
		tr.CancelRequest(req)
		<-c
		return ctx.Err()
	case err := <-c:
		return err
	}
}
