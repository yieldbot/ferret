/*
 * Ferret
 * Copyright (c) 2016 Yieldbot, Inc.
 * For the full copyright and license information, please view the LICENSE.txt file.
 */

package search

import (
	"strconv"
	"time"
)

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
