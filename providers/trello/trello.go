/*
 * Ferret
 * Copyright (c) 2016 Yieldbot, Inc.
 * For the full copyright and license information, please view the LICENSE.txt file.
 */

// Package trello implements Trello provider
package trello

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/net/context"
	"golang.org/x/net/context/ctxhttp"
)

// Register registers the provider
func Register(config map[string]interface{}, f func(interface{}) error) {

	name, ok := config["Name"].(string)
	if name == "" || !ok {
		name = "trello"
	}
	title, ok := config["Title"].(string)
	if title == "" || !ok {
		title = "Trello"
	}
	priority, ok := config["Priority"].(int64)
	if priority == 0 || !ok {
		priority = 800
	}
	key, _ := config["Key"].(string)
	token, _ := config["Token"].(string)
	querySuffix, _ := config["QuerySuffix"].(string)

	p := Provider{
		provider:    "trello",
		name:        name,
		title:       title,
		priority:    priority,
		url:         "https://api.trello.com/1",
		key:         key,
		token:       token,
		querySuffix: querySuffix,
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
	provider    string
	enabled     bool
	name        string
	title       string
	priority    int64
	url         string
	key         string
	token       string
	querySuffix string
}

// Search makes a search
func (provider *Provider) Search(ctx context.Context, args map[string]interface{}) ([]map[string]interface{}, error) {

	results := []map[string]interface{}{}
	page, ok := args["page"].(int)
	if page < 1 || !ok {
		page = 1
	}
	limit, ok := args["limit"].(int)
	if limit < 1 || !ok {
		limit = 10
	}
	keyword, ok := args["keyword"].(string)

	u := fmt.Sprintf("%s/search?key=%s&token=%s&partial=true&modelTypes=cards&card_fields=name,shortUrl,desc,dateLastActivity&cards_page=%d&cards_limit=%d&query=%s", provider.url, provider.key, provider.token, (page - 1), limit, url.QueryEscape(keyword))
	if provider.querySuffix != "" {
		u += fmt.Sprintf("%s", provider.querySuffix)
	}
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
	for _, v := range sr.Cards {
		d := strings.TrimSpace(v.Description)
		if len(d) > 255 {
			d = d[0:252] + "..."
		}

		var t time.Time
		if ts, err := time.Parse("2006-01-02T15:04:05.000Z", v.DateLastActivity); err == nil {
			t = ts
		}

		ri := map[string]interface{}{
			"Link":        v.URL,
			"Title":       v.Name,
			"Description": d,
			"Date":        t,
		}
		results = append(results, ri)
	}

	return results, err
}

// SearchResult represents the structure of the search result
type SearchResult struct {
	Cards []*SRCards `json:"cards"`
}

// SRCards represents the structure of the search result list
type SRCards struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	URL              string `json:"shortUrl"`
	Description      string `json:"desc"`
	DateLastActivity string `json:"dateLastActivity"`
}
