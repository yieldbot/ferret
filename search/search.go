/*
 * Ferret
 * Copyright (c) 2016 Yieldbot, Inc.
 * For the full copyright and license information, please view the LICENSE.txt file.
 */

// Package search provides search interface and functionality
package search

import (
	"os"

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

// Searcher is the interface that must be implemented by a search provider
type Searcher interface {
	// Search makes a search
	Search(ctx context.Context, args map[string]interface{}) ([]map[string]interface{}, error)
}
