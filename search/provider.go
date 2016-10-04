/*
 * Ferret
 * Copyright (c) 2016 Yieldbot, Inc.
 * For the full copyright and license information, please view the LICENSE.txt file.
 */

package search

import (
	"errors"
	"reflect"
	"sort"
)

// Provider represents a provider
type Provider struct {
	Name     string
	Title    string
	Enabled  bool
	Noui     bool
	Priority int64
	Rewrite  string
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

	var name, title, rewrite string
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
		} else if fn == "rewrite" && ft == "string" {
			rewrite = v.Field(i).String()
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
		Rewrite:  rewrite,
		Searcher: p,
	}
	providers[name] = np

	return nil
}
