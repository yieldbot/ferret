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
func Register(args []map[string]interface{}, f func(interface{}) error) {
	for _, v := range args {
		p, ok := v["Provider"].(string)
		if !ok {
			continue
		}

		switch p {
		case "answerhub":
			answerhub.Register(v, f)
		case "consul":
			consul.Register(v, f)
		case "github":
			github.Register(v, f)
		case "slack":
			slack.Register(v, f)
		case "trello":
			trello.Register(v, f)
		default:
			panic("invalid provider: " + p)
		}
	}
}
