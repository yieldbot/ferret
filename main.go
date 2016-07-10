/*
 * Ferret
 * Copyright (c) 2016 Yieldbot, Inc.
 * For the full copyright and license information, please view the LICENSE.txt file.
 */

// Ferret is a search engine
package main

import (
	"flag"

	"github.com/yieldbot/ferret/api"
	_ "github.com/yieldbot/ferret/providers"
	"github.com/yieldbot/ferret/search"
	"github.com/yieldbot/gocli"
	"golang.org/x/net/context"
)

func init() {
	// Init flags
	flag.BoolVar(&usageFlag, "h", false, "Display usage")
	flag.BoolVar(&usageFlag, "help", false, "Display usage")
	flag.BoolVar(&versionFlag, "v", false, "Display version information")
	flag.BoolVar(&versionExtFlag, "vv", false, "Display extended version information")
}

var (
	version        = "latest"
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
			"listen": "Listen for the UI and REST API requests (Usage: ferret listen)",
			"search": "Search by the given provider (Usage: ferret search PROVIDER KEYWORD)",
		},
	}
	cli.Init()
	search.Logger = cli.LogErr

	if versionFlag || versionExtFlag {
		// Version
		cli.PrintVersion(versionExtFlag)
	} else if cli.SubCommand == "search" {
		// Search
		var q = search.Query{
			Page:    search.ParsePage(cli.SubCommandArgsMap["page"]),
			Goto:    search.ParseGoto(cli.SubCommandArgsMap["goto"]),
			Timeout: search.ParseTimeout(cli.SubCommandArgsMap["timeout"]),
		}
		if len(cli.SubCommandArgs) > 0 {
			q.Provider = cli.SubCommandArgs[0]
			if len(cli.SubCommandArgs) > 1 {
				q.Keyword = cli.SubCommandArgs[1]
			}
		}
		search.PrintResults(search.Do(context.Background(), q))
	} else if cli.SubCommand == "listen" {
		// Listen
		api.Listen()
	} else {
		// Default
		cli.PrintUsage()
	}
}
