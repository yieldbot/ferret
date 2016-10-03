/*
 * Ferret
 * Copyright (c) 2016 Yieldbot, Inc.
 * For the full copyright and license information, please view the LICENSE.txt file.
 */

// Package search provides search interface and functionality
package search

import (
	"reflect"

	conf "github.com/yieldbot/ferret/config"
	prov "github.com/yieldbot/ferret/providers"
	"golang.org/x/net/context"
)

var (
	config    conf.Search
	providers = make(map[string]Provider)
)

// Init initializes the search
func Init(c conf.Config) {
	config = c.Search

	// Iterate config and create the config map for providers
	cm := []map[string]interface{}{}
	for _, v := range c.Providers {
		m := map[string]interface{}{}
		vr := reflect.Indirect(reflect.ValueOf(v))
		for i := 0; i < vr.NumField(); i++ {
			f := vr.Type().Field(i)
			if !f.Anonymous {
				m[f.Name] = vr.Field(i).Interface()
			}
		}
		cm = append(cm, m)
	}
	prov.Register(cm, ProviderRegister)
}

// Searcher is the interface that must be implemented by a search provider
type Searcher interface {
	// Search makes a search
	Search(ctx context.Context, args map[string]interface{}) ([]map[string]interface{}, error)
}
