/*
 * Ferret
 * Copyright (c) 2016 Yieldbot, Inc.
 * For the full copyright and license information, please view the LICENSE.txt file.
 */

package search

// Query represents a search query
type Query struct {
	Provider string
	Keyword  string
	Page     string
	Goto     string
	Timeout  string
}
