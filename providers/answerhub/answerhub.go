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
	"time"

	"golang.org/x/net/context"
)

// Register registers the provider
func Register(f func(provider interface{}) error) {
	// Init the provider
	var p = Provider{
		name:     "answerhub",
		title:    "AnswerHub",
		url:      strings.TrimSuffix(os.Getenv("FERRET_ANSWERHUB_URL"), "/"),
		username: os.Getenv("FERRET_ANSWERHUB_USERNAME"),
		password: os.Getenv("FERRET_ANSWERHUB_PASSWORD"),
	}

	// Register the provider
	if err := f(&p); err != nil {
		panic(err)
	}
}

// Provider represents the provider
type Provider struct {
	name     string
	title    string
	url      string
	username string
	password string
}

// Info returns information
func (provider *Provider) Info() map[string]interface{} {
	return map[string]interface{}{
		"name":  provider.name,
		"title": provider.title,
	}
}

// SearchResult represent the structure of the search result
type SearchResult struct {
	List []*SearchResultList `json:"list"`
}

// SearchResultList represent the structure of the search result list
type SearchResultList struct {
	ID           int                     `json:"id"`
	Title        string                  `json:"title"`
	Body         string                  `json:"body"`
	Author       *SearchResultListAuthor `json:"author"`
	CreationDate int64                   `json:"creationDate"`
}

// SearchResultListAuthor represent the structure of the search result list author field
type SearchResultListAuthor struct {
	Username string `json:"username"`
	Realname string `json:"realname"`
}

// Search makes a search
func (provider *Provider) Search(ctx context.Context, args map[string]interface{}) ([]map[string]interface{}, error) {

	results := []map[string]interface{}{}
	page, ok := args["page"].(int)
	if page < 1 || !ok {
		page = 1
	}
	keyword, ok := args["keyword"].(string)

	u := fmt.Sprintf("%s/services/v2/node.json?page=%d&pageSize=10&q=%s*", provider.url, page, url.QueryEscape(keyword))
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, errors.New("failed to prepare request. Error: " + err.Error())
	}
	if provider.username != "" || provider.password != "" {
		req.SetBasicAuth(provider.username, provider.password)
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
		for _, v := range sr.List {
			d := strings.TrimSpace(v.Body)
			if len(d) > 255 {
				d = d[0:252] + "..."
			} else if len(d) == 0 {
				if v.Author.Realname != "" {
					d = "Asked by " + v.Author.Realname
				} else {
					d = "Asked by " + v.Author.Username
				}
			}
			ri := map[string]interface{}{
				"Link":        fmt.Sprintf("%s/questions/%d/", provider.url, v.ID),
				"Title":       v.Title,
				"Description": d,
				"Date":        time.Unix(0, v.CreationDate*int64(time.Millisecond)),
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
