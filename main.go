/*
 * Ferret
 * Copyright (c) 2016 Yieldbot, Inc.
 * For the full copyright and license information, please view the LICENSE.txt file.
 */

// Ferret is a search engine
package main

import (
	"flag"

	_ "github.com/yieldbot/ferret/providers"
	"github.com/yieldbot/ferret/search"
	"github.com/yieldbot/gocli"
)

func init() {
	// Init flags
	flag.BoolVar(&usageFlag, "h", false, "Display usage")
	flag.BoolVar(&usageFlag, "help", false, "Display usage")
	flag.BoolVar(&versionFlag, "v", false, "Display version information")
	flag.BoolVar(&versionExtFlag, "vv", false, "Display extended version information")
}

var (
	version        = ""
	gitCommit      = ""
	cli            gocli.Cli
	usageFlag      bool
	versionFlag    bool
	versionExtFlag bool
)

func main() {

	// Init cli
	cli = gocli.Cli{
		Name:        "Ferret",
		Version:     version,
		Description: "Ferret is a search engine",
		Commands: map[string]string{
			"search": "Search by the given provider (Usage: ferret search PROVIDER KEYWORD)",
		},
	}
	cli.Init()

	if versionFlag || versionExtFlag {
		// Version
		cli.PrintVersion(versionExtFlag)
	} else if cli.SubCommand == "search" {
		// Search
		var provider, keyword string
		if len(cli.SubCommandArgs) > 0 {
			provider = cli.SubCommandArgs[0]
			if len(cli.SubCommandArgs) > 1 {
				keyword = cli.SubCommandArgs[0]
			}
		}
		search.ByKeyword(provider, keyword, cli.SubCommandArgsMap)
	} else {
		// Default
		cli.PrintUsage()
	}
}
