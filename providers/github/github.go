/*
 * Ferret
 * Copyright (c) 2016 Yieldbot, Inc.
 * For the full copyright and license information, please view the LICENSE.txt file.
 */

// Package github implements AnswerHub provider
package github

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"

	"golang.org/x/net/context"
)

// Register registers the provider
func Register(f func(name string, provider interface{}) error) {
	// Init the provider
	var p = Provider{
		url:        strings.TrimSuffix(os.Getenv("FERRET_GITHUB_URL"), "/"),
		token:      os.Getenv("FERRET_GITHUB_TOKEN"),
		searchUser: os.Getenv("FERRET_GITHUB_SEARCH_USER"),
	}

	// Register the provider
	if err := f("github", &p); err != nil {
		panic(err)
	}
}

// Provider represents the provider
type Provider struct {
	url        string
	token      string
	searchUser string
}

// SearchResult represent the structure of the search result
type SearchResult struct {
	TotalCount        int  `json:"total_count"`
	IncompleteResults bool `json:"incomplete_results"`
	Items             []*SearchResultItems
}

// SearchResultItems represent the structure of the search result items
type SearchResultItems struct {
	Name       string                       `json:"name"`
	Path       string                       `json:"path"`
	HTMLUrl    string                       `json:"html_url"`
	Repository *SearchResultItemsRepository `json:"repository"`
}

// SearchResultItemsRepository represent the structure of the search result items repository
type SearchResultItemsRepository struct {
	Fullname    string `json:"full_name"`
	Description string `json:"description"`
}

// Search makes a search
func (provider *Provider) Search(ctx context.Context, args map[string]interface{}) ([]map[string]interface{}, error) {

	var results = []map[string]interface{}{}
	page, ok := args["page"].(int)
	if page < 1 || !ok {
		page = 1
	}
	keyword, ok := args["keyword"].(string)

	var u = fmt.Sprintf("%s/search/code?page=%d&per_page=10&q=%s", provider.url, page, url.QueryEscape(keyword))
	if provider.searchUser != "" {
		u += fmt.Sprintf("+user:%s", url.QueryEscape(provider.searchUser))
	}
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, errors.New("failed to prepare request. Error: " + err.Error())
	}
	if provider.token != "" {
		req.Header.Set("Authorization", "token "+provider.token)
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
		for _, v := range sr.Items {
			// TODO: Improve alignment (repo: path)
			ri := map[string]interface{}{
				"Description": fmt.Sprintf("%s: %s", v.Repository.Fullname, v.Path),
				"Link":        v.HTMLUrl,
			}
			results = append(results, ri)
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
