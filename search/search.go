/*
 * Ferret
 * Copyright (c) 2016 Yieldbot, Inc.
 * For the full copyright and license information, please view the LICENSE.txt file.
 */

// Package search provides search interface and functionality
package search

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"time"

	"github.com/yieldbot/ferret/providers"
	"github.com/yieldbot/gocli"
	"golang.org/x/net/context"
)

var (
	// Logger is the logger
	Logger *log.Logger

	goCommand     = "open"
	searchTimeout = "5000ms"
	searchers     = make(map[string]Searcher)
)

func init() {
	Logger = log.New(os.Stderr, "", log.LstdFlags)

	if e := os.Getenv("FERRET_GOTO_CMD"); e != "" {
		goCommand = e
	}

	if e := os.Getenv("FERRET_SEARCH_TIMEOUT"); e != "" {
		searchTimeout = e
	}

	providers.Register(Register)
}

// Searcher is the interface that must be implemented by a search provider
type Searcher interface {
	// Search makes a search
	Search(ctx context.Context, args map[string]interface{}) ([]map[string]interface{}, error)
}

// Register registers a search provider
func Register(name string, provider interface{}) error {
	if _, ok := searchers[name]; ok {
		return errors.New("search provider " + name + " is already registered")
	}
	searchers[name] = provider.(Searcher)
	return nil
}

// Searchers returns a sorted list of the names of the registered searchers
func Searchers() []string {
	var list = []string{}
	for name := range searchers {
		list = append(list, name)
	}
	sort.Strings(list)
	return list
}

// Do makes a query by the given context and query
func Do(ctx context.Context, query Query) (Query, error) {

	// Provider
	s, ok := searchers[query.Provider]
	if !ok {
		return query, fmt.Errorf("invalid search provider. Possible search providers are %s", Searchers())
	}

	// Keyword
	if query.Keyword == "" {
		return query, errors.New("missing keyword")
	}

	// Page
	if query.Page <= 0 {
		return query, errors.New("invalid page #. It should be greater than 0")
	}

	// Search
	query.Start = time.Now()
	ctx, cancel := context.WithTimeout(ctx, query.Timeout)
	defer cancel()
	sq := map[string]interface{}{"page": query.Page, "keyword": query.Keyword}
	sr, err := s.Search(ctx, sq)
	for _, srv := range sr {
		query.Results = append(query.Results, Result{
			Description: srv["Description"].(string),
			Link:        srv["Link"].(string),
		})
	}
	if err != nil {
		return query, errors.New("failed to search due to " + err.Error())
	}
	query.Elapsed = time.Since(query.Start)

	// Goto
	if query.Goto != 0 {
		if query.Goto < 0 || query.Goto > len(query.Results) {
			return query, fmt.Errorf("invalid result # to go. It should be between 1 and %d", len(query.Results))
		}
		link := query.Results[query.Goto-1].Link
		if _, err = exec.Command(goCommand, link).Output(); err != nil {
			return query, fmt.Errorf("failed to go to %s due to %s. Check FERRET_GOTO_CMD environment variable", link, err.Error())
		}
		return query, nil
	}

	return query, nil
}

// DoRequest makes a HTTP request with context
func DoRequest(ctx context.Context, req *http.Request, f func(*http.Response, error) error) error {
	tr := &http.Transport{}
	client := &http.Client{Transport: tr}
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

// PrintResults prints the given search results
func PrintResults(query Query, err error) {
	if err != nil {
		Logger.Fatal(err)
	}

	if query.Goto == 0 {
		var t = gocli.Table{}
		t.AddRow(1, "#", "DESCRIPTION")
		for i, v := range query.Results {
			t.AddRow(i+2, fmt.Sprintf("%d", i+1), v.Description)
		}
		t.PrintData()
		fmt.Printf("\n%dms\n", int64(query.Elapsed/time.Millisecond))
	}
}

// ParsePage parses page from a given string
func ParsePage(p string) int {
	var page = 1
	if p != "" {
		i, err := strconv.Atoi(p)
		if err == nil && i > 0 {
			page = i
		}
	}
	return page
}

// ParseGoto parses goto from a given string
func ParseGoto(g string) int {
	var goo = 0
	if g != "" {
		i, err := strconv.Atoi(g)
		if err == nil && i > 0 {
			goo = i
		}
	}
	return goo
}

// ParseTimeout parses timeout from a given string
func ParseTimeout(t string) time.Duration {
	var timeout = 5000 * time.Millisecond
	if t != "" {
		d, err := time.ParseDuration(t)
		if err == nil {
			timeout = d
		}
	} else {
		d, err := time.ParseDuration(searchTimeout)
		if err == nil {
			timeout = d
		}
	}
	return timeout
}
