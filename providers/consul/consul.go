/*
 * Ferret
 * Copyright (c) 2016 Yieldbot, Inc.
 * For the full copyright and license information, please view the LICENSE.txt file.
 */

// Package consul implements Consul provider
package consul

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"

	"github.com/yieldbot/ferret/search"
)

func init() {
	// Init the provider
	var p = Provider{
		url: strings.TrimSuffix(os.Getenv("FERRET_CONSUL_URL"), "/"),
	}

	// Register the provider
	if err := search.Register("consul", &p); err != nil {
		panic(err)
	}
}

// Provider represents the provider
type Provider struct {
	url string
}

// SearchResult represent the structure of the search result
type SearchResult map[string][]string

// Search makes a search
func (provider *Provider) Search(keyword string, page int) (search.ResultItems, error) {

	var result search.ResultItems

	dcs, err := provider.datacenter()
	if err != nil {
		return nil, errors.New("failed to fetch data. Error: " + err.Error())
	}
	for _, dc := range dcs {

		// Prepare the request
		query := fmt.Sprintf("%s/v1/catalog/services?dc=%s", provider.url, url.QueryEscape(dc))
		req, err := http.NewRequest("GET", query, nil)

		// Make the request
		var client = &http.Client{}
		res, err := client.Do(req)
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

		// Parse and prepare the result
		var sr SearchResult
		if err = json.Unmarshal(data, &sr); err != nil {
			return nil, errors.New("failed to unmarshal JSON data. Error: " + err.Error())
		}
		for k, v := range sr {
			if len(v) > 0 {
				for _, vv := range v {
					if strings.Contains(vv, keyword) || strings.Contains(k, keyword) {
						ri := search.ResultItem{
							Description: fmt.Sprintf("%s.%s.service.%s.consul", vv, k, dc),
							Link:        fmt.Sprintf("%s/ui/#/%s/services/%s", provider.url, dc, k),
						}
						result = append(result, ri)
					}
				}
			} else {
				if strings.Contains(k, keyword) {
					ri := search.ResultItem{
						Description: fmt.Sprintf("%s.service.%s.consul", k, dc),
						Link:        fmt.Sprintf("%s/ui/#/%s/services/%s", provider.url, dc, k),
					}
					result = append(result, ri)
				}
			}
		}
	}

	sort.Sort(result)
	var l, h = 0, 10
	if page > 1 {
		h = (page * 10)
		l = h - 10
	}
	if h > len(result) {
		h = len(result)
	}
	result = result[l:h]

	return result, nil
}

// datacenter gets the list of the datacenters
func (provider *Provider) datacenter() ([]string, error) {

	// Prepare the request
	query := fmt.Sprintf("%s/v1/catalog/datacenters", provider.url)
	req, err := http.NewRequest("GET", query, nil)

	// Make the request
	var client = &http.Client{}
	res, err := client.Do(req)
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

	// Parse and prepare the result
	var result []string
	if err = json.Unmarshal(data, &result); err != nil {
		return nil, errors.New("failed to unmarshal JSON data. Error: " + err.Error())
	}

	return result, nil
}
