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

	"github.com/yieldbot/ferret/search"
	"golang.org/x/net/context"
)

func init() {
	// Init the provider
	var p = Provider{
		url:   "https://api.trello.com/1",
		key:   os.Getenv("FERRET_TRELLO_KEY"),
		token: os.Getenv("FERRET_TRELLO_TOKEN"),
	}

	// Register the provider
	if err := search.Register("trello", &p); err != nil {
		panic(err)
	}
}

// Provider represents the provider
type Provider struct {
	url   string
	key   string
	token string
}

// SearchResult represent the structure of the search result
type SearchResult struct {
	Cards []*SearchResultCards
}

// SearchResultCards represent the structure of the search result list
type SearchResultCards struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	URL  string `json:"shortUrl"`
}

// Search makes a search
func (provider *Provider) Search(ctx context.Context, query search.Query) (search.Query, error) {

	var u = fmt.Sprintf("%s/search?key=%s&token=%s&partial=true&modelTypes=cards&card_fields=name,shortUrl&cards_limit=10&cards_page=%d&query=%s", provider.url, provider.key, provider.token, (query.Page - 1), url.QueryEscape(query.Keyword))
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return query, errors.New("failed to prepare request. Error: " + err.Error())
	}

	err = search.DoRequest(ctx, req, func(res *http.Response, err error) error {

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
		for _, v := range sr.Cards {
			ri := search.Result{
				Description: v.Name,
				Link:        v.URL,
			}
			query.Results = append(query.Results, ri)
		}

		return nil
	})

	return query, err
}
