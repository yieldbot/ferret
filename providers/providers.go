/*
 * Ferret
 * Copyright (c) 2016 Yieldbot, Inc.
 * For the full copyright and license information, please view the LICENSE.txt file.
 */

// Package providers wraps the provider packages
package providers

import (
	// For triggering init()
	_ "github.com/yieldbot/ferret/providers/answerhub"
	_ "github.com/yieldbot/ferret/providers/consul"
	_ "github.com/yieldbot/ferret/providers/github"
	_ "github.com/yieldbot/ferret/providers/slack"
	_ "github.com/yieldbot/ferret/providers/trello"
)
