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
	"os"
	"strings"
	"time"

	"golang.org/x/net/context"
	"golang.org/x/net/context/ctxhttp"
)

// Provider represents the provider
type Provider struct {
	enabled bool
	name    string
	title   string
	url     string
	key     string
	token   string
}

// Register registers the provider
func Register(f func(provider interface{}) error) {
	var p = Provider{
		name:  "trello",
		title: "Trello",
		url:   "https://api.trello.com/1",
		key:   os.Getenv("FERRET_TRELLO_KEY"),
		token: os.Getenv("FERRET_TRELLO_TOKEN"),
	}
	if p.token != "" {
		p.enabled = true
	}

	if err := f(&p); err != nil {
		panic(err)
	}
}

// SearchResult represent the structure of the search result
type SearchResult struct {
	Cards []*SRCards `json:"cards"`
}

// SRCards represent the structure of the search result list
type SRCards struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	URL              string `json:"shortUrl"`
	Description      string `json:"desc"`
	DateLastActivity string `json:"dateLastActivity"`
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

	var u = fmt.Sprintf("%s/search?key=%s&token=%s&partial=true&modelTypes=cards&card_fields=name,shortUrl,desc,dateLastActivity&cards_page=%d&cards_limit=%d&query=%s", provider.url, provider.key, provider.token, (page - 1), limit, url.QueryEscape(keyword))
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, errors.New("failed to prepare request. Error: " + err.Error())
	}

	res, err := ctxhttp.Do(ctx, nil, req)
	if err != nil {
		return nil, errors.New("failed to fetch data. Error: " + err.Error())
	} else if res.StatusCode < 200 || res.StatusCode > 299 {
		return nil, errors.New("bad response: " + fmt.Sprintf("%d", res.StatusCode))
	}
	defer res.Body.Close()
	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	var sr SearchResult
	if err = json.Unmarshal(data, &sr); err != nil {
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
