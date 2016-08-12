/*
 * Ferret
 * Copyright (c) 2016 Yieldbot, Inc.
 * For the full copyright and license information, please view the LICENSE.txt file.
 */

// Package search provides search interface and functionality
package search

import (
	"errors"
	"os"
	"reflect"
	"sort"
	"strconv"
	"time"

	prov "github.com/yieldbot/ferret/providers"
	"golang.org/x/net/context"
)

var (
	goCommand     = "open"
	searchTimeout = "5000ms"
	providers     = make(map[string]Provider)
)

func init() {
	if e := os.Getenv("FERRET_GOTO_CMD"); e != "" {
		goCommand = e
	}
	if e := os.Getenv("FERRET_SEARCH_TIMEOUT"); e != "" {
		searchTimeout = e
	}

	prov.Register(ProviderRegister)
}

// Provider represents a provider
type Provider struct {
	Name     string
	Title    string
	Enabled  bool
	Noui     bool
	Priority int64
	Searcher
}

// Providers returns a sorted list of the names of the providers
func Providers() []string {
	l := []string{}
	for n := range providers {
		l = append(l, n)
	}
	sort.Strings(l)
	return l
}

// ProviderByName returns a provider by the given name
func ProviderByName(name string) (Provider, error) {
	p, ok := providers[name]
	if !ok {
		return p, errors.New("provider " + name + " couldn't be found")
	}
	return p, nil
}

// ProviderRegister registers a search provider
func ProviderRegister(provider interface{}) error {

	// Init provider
	p, ok := provider.(Searcher)
	if !ok {
		return errors.New("invalid provider")
	}

	var name, title string
	var enabled, noui bool
	var priority int64

	// Get the value of the provider
	v := reflect.Indirect(reflect.ValueOf(p))
	// Iterate the provider fields
	for i := 0; i < v.NumField(); i++ {
		fn := v.Type().Field(i).Name
		ft := v.Field(i).Type().Name()

		if fn == "name" && ft == "string" {
			name = v.Field(i).String()
		} else if fn == "title" && ft == "string" {
			title = v.Field(i).String()
		} else if fn == "enabled" && ft == "bool" {
			enabled = v.Field(i).Bool()
		} else if fn == "noui" && ft == "bool" {
			noui = v.Field(i).Bool()
		} else if fn == "priority" && ft == "int64" {
			priority = v.Field(i).Int()
		}
	}
	if name == "" {
		return errors.New("invalid provider name")
	}
	if title == "" {
		title = name
	}

	// Init provider
	if _, ok := providers[name]; ok {
		return errors.New("search provider " + name + " is already registered")
	}
	np := Provider{
		Name:     name,
		Title:    title,
		Enabled:  enabled,
		Noui:     noui,
		Priority: priority,
		Searcher: p,
	}
	providers[name] = np

	return nil
}

// Searcher is the interface that must be implemented by a search provider
type Searcher interface {
	// Search makes a search
	Search(ctx context.Context, args map[string]interface{}) ([]map[string]interface{}, error)
}

// ParsePage parses page from a given string
func ParsePage(page string) int {
	p := 1
	if page != "" {
		i, err := strconv.Atoi(page)
		if err == nil && i > 0 {
			p = i
		}
	}
	return p
}

// ParseGoto parses goto from a given string
func ParseGoto(gt string) int {
	g := 0
	if gt != "" {
		i, err := strconv.Atoi(gt)
		if err == nil && i > 0 {
			g = i
		}
	}
	return g
}

// ParseTimeout parses timeout from a given string
func ParseTimeout(timeout string) time.Duration {
	t := 5000 * time.Millisecond
	if timeout != "" {
		d, err := time.ParseDuration(timeout)
		if err == nil {
			t = d
		}
	} else {
		d, err := time.ParseDuration(searchTimeout)
		if err == nil {
			t = d
		}
	}
	return t
}

// ParseLimit parses limit from a given string
func ParseLimit(limit string) int {
	l := 10
	if limit != "" {
		i, err := strconv.Atoi(limit)
		if err == nil && i > 0 {
			l = i
		}
	}
	return l
}
