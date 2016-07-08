/*
 * Ferret
 * Copyright (c) 2016 Yieldbot, Inc.
 * For the full copyright and license information, please view the LICENSE.txt file.
 */

// Package providers wraps the provider packages
package providers

import (
	"github.com/yieldbot/ferret/providers/answerhub"
	"github.com/yieldbot/ferret/providers/consul"
	"github.com/yieldbot/ferret/providers/github"
	"github.com/yieldbot/ferret/providers/slack"
	"github.com/yieldbot/ferret/providers/trello"
)

// Register registers the providers
func Register(f func(provider interface{}) error) {
	answerhub.Register(f)
	consul.Register(f)
	github.Register(f)
	slack.Register(f)
	trello.Register(f)
}
