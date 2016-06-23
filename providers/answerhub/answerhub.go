/*
 * Ferret
 * Copyright (c) 2016 Yieldbot, Inc.
 * For the full copyright and license information, please view the LICENSE.txt file.
 */

// Package answerhub implements AnswerHub provider
package answerhub

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/yieldbot/ferret/search"
	"golang.org/x/net/context"
)

func init() {
	// Init the provider
	var p = Provider{
		url:      strings.TrimSuffix(os.Getenv("FERRET_ANSWERHUB_URL"), "/"),
		username: os.Getenv("FERRET_ANSWERHUB_USERNAME"),
		password: os.Getenv("FERRET_ANSWERHUB_PASSWORD"),
	}

	// Register the provider
	if err := search.Register("answerhub", &p); err != nil {
		panic(err)
	}
}

// Provider represents the provider
type Provider struct {
	url      string
	username string
	password string
}

// SearchResult represent the structure of the search result
type SearchResult struct {
	Name       string
	Sort       string
	Page       int
	PageSize   int
	PageCount  int
	ListCount  int
	TotalCount int
	Sorts      []string
	List       []*SearchResultList
}

// SearchResultList represent the structure of the search result list
type SearchResultList struct {
	ID    int
	Type  string
	Title string
}

// Search makes a search
func (provider *Provider) Search(ctx context.Context, query search.Query) (search.Query, error) {

	var u = fmt.Sprintf("%s/services/v2/node.json?page=%d&pageSize=10&q=%s*", provider.url, query.Page, url.QueryEscape(query.Keyword))
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return query, errors.New("failed to prepare request. Error: " + err.Error())
	}
	if provider.username != "" || provider.password != "" {
		req.SetBasicAuth(provider.username, provider.password)
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
		for _, v := range sr.List {
			ri := search.Result{
				Description: v.Title,
				Link:        fmt.Sprintf("%s/questions/%d/", provider.url, v.ID),
			}
			query.Results = append(query.Results, ri)
		}

		return nil
	})

	return query, err
}
