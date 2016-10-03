/*
 * Ferret
 * Copyright (c) 2016 Yieldbot, Inc.
 * For the full copyright and license information, please view the LICENSE.txt file.
 */

// Ferret is a search engine
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/yieldbot/ferret/api"
	conf "github.com/yieldbot/ferret/config"
	"github.com/yieldbot/ferret/search"
	"github.com/yieldbot/gocli"
)

func init() {
	// Init flags
	flag.BoolVar(&usageFlag, "h", false, "Display usage")
	flag.BoolVar(&usageFlag, "help", false, "Display usage")
	flag.BoolVar(&versionFlag, "v", false, "Display version information")
	flag.BoolVar(&versionExtFlag, "vv", false, "Display extended version information")
	flag.StringVar(&configFlag, "config", "", "Config file")
}

var (
	version        = "latest"
	commit         = ""
	cli            gocli.Cli
	usageFlag      bool
	versionFlag    bool
	versionExtFlag bool
	configFlag     string
	config         conf.Config
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

	// Config
	config = conf.Config{}
	if configFlag != "" {
		config.File = configFlag
	} else if os.Getenv("FERRET_CONFIG") != "" {
		config.File = os.Getenv("FERRET_CONFIG")
	}
	if config.File != "" {
		if err := config.Load(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	if versionFlag || versionExtFlag {
		// Version
		cli.PrintVersion(versionExtFlag)
	} else if cli.SubCommand == "search" {
		// Search
		search.Init(config)
		q := search.Query{
			Page:    search.ParsePage(cli.SubCommandArgsMap["page"]),
			Goto:    search.ParseGoto(cli.SubCommandArgsMap["goto"]),
			Timeout: search.ParseTimeout(cli.SubCommandArgsMap["timeout"]),
			Limit:   search.ParseLimit(cli.SubCommandArgsMap["limit"]),
		}
		if len(cli.SubCommandArgs) > 0 {
			q.Provider = cli.SubCommandArgs[0]
			if len(cli.SubCommandArgs) > 1 {
				q.Keyword = cli.SubCommandArgs[1]
			}
		}
		q.DoPrint(q.Do())
	} else if cli.SubCommand == "listen" {
		// Listen
		search.Init(config)
		api.Init(config)
		api.Listen()
	} else {
		// Default
		cli.PrintUsage()
	}
}
