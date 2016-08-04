/*
 * Ferret
 * Copyright (c) 2016 Yieldbot, Inc.
 * For the full copyright and license information, please view the LICENSE.txt file.
 */

package search

import (
	"time"
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
