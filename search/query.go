/*
 * Ferret
 * Copyright (c) 2016 Yieldbot, Inc.
 * For the full copyright and license information, please view the LICENSE.txt file.
 */

package search

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/yieldbot/gocli"
	"golang.org/x/net/context"
)

// Query represents a search query
type Query struct {
	Provider   string
	Keyword    string
	Limit      int
	Page       int
	Goto       int
	Timeout    time.Duration
	Start      time.Time
	Elapsed    time.Duration
	HTTPStatus int
	Results    Results
}

// Do runs the search query
func (query *Query) Do() error {

	// Provider
	provider, ok := providers[query.Provider]
	if !ok {
		query.HTTPStatus = http.StatusBadRequest
		return fmt.Errorf("invalid search provider. Possible search providers are %s", Providers())
	}

	// Keyword
	if query.Keyword == "" {
		query.HTTPStatus = http.StatusBadRequest
		return errors.New("missing keyword")
	}

	// Page
	if query.Page <= 0 {
		query.HTTPStatus = http.StatusBadRequest
		return errors.New("invalid page #. It should be greater than 0")
	}

	// Limit
	if query.Limit <= 0 {
		query.HTTPStatus = http.StatusBadRequest
		return errors.New("invalid limit. It should be greater than 0")
	}

	// Search
	query.Start = time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), query.Timeout)
	defer cancel()
	sq := map[string]interface{}{"page": query.Page, "limit": query.Limit, "keyword": query.Keyword}
	sr, err := provider.Search(ctx, sq)
	if err != nil {
		if err == context.DeadlineExceeded {
			query.HTTPStatus = http.StatusGatewayTimeout
			return errors.New("timeout")
		} else if err == context.Canceled {
			query.HTTPStatus = http.StatusInternalServerError
			return errors.New("canceled")
		}
		query.HTTPStatus = http.StatusInternalServerError
		return errors.New("failed to search due to " + err.Error())
	}
	query.Elapsed = time.Since(query.Start)
	for _, srv := range sr {
		var d string
		if _, ok := srv["Description"]; ok {
			d = srv["Description"].(string)
		}

		var t time.Time
		if _, ok := srv["Date"]; ok {
			t = srv["Date"].(time.Time)
		}

		query.Results = append(query.Results, Result{
			Link:        srv["Link"].(string),
			Title:       srv["Title"].(string),
			Description: d,
			Date:        t,
			From:        provider.Title,
		})
	}

	// Goto
	if query.Goto != 0 {
		if query.Goto < 0 || query.Goto > len(query.Results) {
			return fmt.Errorf("invalid result # to go. It should be between 1 and %d", len(query.Results))
		}
		link := query.Results[query.Goto-1].Link
		if _, err = exec.Command(goCommand, link).Output(); err != nil {
			return fmt.Errorf("failed to go to %s due to %s. Check FERRET_GOTO_CMD environment variable", link, err.Error())
		}
		return nil
	}

	return nil
}

// DoPrint handles terminal output for Do function
func (query *Query) DoPrint(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if query.Goto == 0 {
		t := gocli.Table{}
		t.AddRow(1, "#", "TITLE")
		for i, v := range query.Results {
			ts := ""
			if !v.Date.IsZero() {
				ts = fmt.Sprintf(" (%d-%02d-%02d)", v.Date.Year(), v.Date.Month(), v.Date.Day())
			}
			t.AddRow(i+2, fmt.Sprintf("%d", i+1), fmt.Sprintf("%s%s", v.Title, ts))
		}
		t.PrintData()
		fmt.Printf("\n%d rows in %dms\n", len(query.Results), int64(query.Elapsed/time.Millisecond))
	}
}

// Result represents a search result
type Result struct {
	Link        string    `json:"link"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Date        time.Time `json:"date"`
	From        string    `json:"from"`
}

// Results represents a list of search results
type Results []Result

// Sort implementation
func (r Results) Len() int {
	return len(r)
}
func (r Results) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}
func (r Results) Less(i, j int) bool {
	return r[i].Title < r[j].Title
}
