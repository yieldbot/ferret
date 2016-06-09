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
func (provider *Provider) Search(keyword string) ([]search.ResultItem, error) {

	// Prepare the request
	query := fmt.Sprintf("%s/search?key=%s&token=%s&partial=true&modelTypes=cards&card_fields=name,shortUrl&query=%s", provider.url, provider.key, provider.token, url.QueryEscape(keyword))
	req, err := http.NewRequest("GET", query, nil)

	// Make the request
	res, err := provider.do(req)
	if err != nil {
		return nil, errors.New("failed to fetch search result. Error: " + err.Error())
	}

	// Parse and prepare the result
	var sr SearchResult
	if err = json.Unmarshal(res, &sr); err != nil {
		return nil, errors.New("failed to unmarshal JSON data. Error: " + err.Error())
	}
	var result []search.ResultItem
	for _, v := range sr.Cards {
		ri := search.ResultItem{
			Description: v.Name,
			Link:        v.URL,
		}
		result = append(result, ri)
	}

	return result, nil
}

// do makes request
func (provider *Provider) do(req *http.Request) ([]byte, error) {

	// Do request
	var client = &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	// Read data
	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	// Check response
	if res.StatusCode < 200 || res.StatusCode > 299 {
		return data, errors.New("bad response: " + fmt.Sprintf("%d", res.StatusCode))
	}

	return data, nil
}
