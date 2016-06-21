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

	"github.com/yieldbot/gocli"
	"golang.org/x/net/context"
)

var (
	// Logger is the logger
	Logger *log.Logger

	goCommand     = "open"
	searchTimeout = "5000ms"
	providers     = make(map[string]Searcher)
)

func init() {
	Logger = log.New(os.Stderr, "", log.LstdFlags)

	if e := os.Getenv("FERRET_GOTO_CMD"); e != "" {
		goCommand = e
	}

	if e := os.Getenv("FERRET_SEARCH_TIMEOUT"); e != "" {
		searchTimeout = e
	}
}

// Searcher is the interface that must be implemented by a search provider
type Searcher interface {
	// Search makes a search
	Search(ctx context.Context, keyword string, page int) (Results, error)
}

// Register registers a search provider
func Register(name string, provider Searcher) error {
	if _, ok := providers[name]; ok {
		return errors.New("provider " + name + " is already registered")
	}
	providers[name] = provider
	return nil
}

// Providers returns a sorted list of the names of the registered providers
func Providers() []string {
	var list = []string{}
	for name := range providers {
		list = append(list, name)
	}
	sort.Strings(list)
	return list
}

// ByKeyword make a search by the given provider and keyword
func ByKeyword(ctx context.Context) (context.Context, error) {

	// Init query
	var query = Query{}
	if ctx.Value("searchQuery") != nil {
		query = ctx.Value("searchQuery").(Query)
	}

	// Provider
	s, ok := providers[query.Provider]
	if !ok {
		return ctx, fmt.Errorf("invalid provider. Possible providers are %s", Providers())
	}

	// Keyword
	if query.Keyword == "" {
		return ctx, errors.New("missing keyword")
	}

	// Page
	page := 1
	if query.Page != "" {
		i, err := strconv.Atoi(query.Page)
		if err != nil || i <= 0 {
			return ctx, errors.New("invalid page #. It should be greater than 0")
		}
		page = i
	}

	// Timeout
	var timeout = 5000 * time.Millisecond
	if query.Timeout != "" {
		d, err := time.ParseDuration(query.Timeout)
		if err != nil {
			return ctx, errors.New("invalid timeout. It should be a duration (i.e. 5000ms)")
		}
		timeout = d
	} else {
		d, err := time.ParseDuration(searchTimeout)
		if err == nil {
			timeout = d
		}
	}

	// Search
	ctx = context.WithValue(ctx, "timeStart", time.Now())
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	results, err := s.Search(ctx, query.Keyword, page)
	if err != nil {
		return ctx, errors.New("failed to search due to " + err.Error())
	}
	ctx = context.WithValue(ctx, "results", results)
	ctx = context.WithValue(ctx, "timeElapsed", time.Since(ctx.Value("timeStart").(time.Time)))

	// Goto
	if query.Goto != "" {
		i, err := strconv.Atoi(query.Goto)
		if err != nil || (i <= 0 || len(results) < i) {
			return ctx, fmt.Errorf("invalid result # to go. It should be between 1 and %d", len(results))
		}
		link := results[i-1].Link
		if _, err = exec.Command(goCommand, link).Output(); err != nil {
			return ctx, fmt.Errorf("failed to go to %s due to %s. Check FERRET_GOTO_CMD environment variable", link, err.Error())
		}
		return ctx, nil
	}

	return ctx, nil
}

// DoRequest makes a HTTP request with contex
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
func PrintResults(ctx context.Context, err error) {
	if err != nil {
		Logger.Fatal(err)
	}

	if results := ctx.Value("results"); results != nil {
		results := ctx.Value("results").(Results)

		// Init query
		var query = Query{}
		if ctx.Value("searchQuery") != nil {
			query = ctx.Value("searchQuery").(Query)
		}

		if query.Goto == "" {
			var t = gocli.Table{}
			t.AddRow(1, "#", "DESCRIPTION")
			for i, v := range results {
				t.AddRow(i+2, fmt.Sprintf("%d", i+1), v.Description)
			}
			t.PrintData()
			elapsed := ctx.Value("timeElapsed").(time.Duration)
			fmt.Printf("\n%dms\n", int64(elapsed/time.Millisecond))
		}
	}
}
