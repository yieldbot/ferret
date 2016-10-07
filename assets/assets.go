/*
 * Ferret
 * Copyright (c) 2016 Yieldbot, Inc.
 * For the full copyright and license information, please view the LICENSE.txt file.
 */

// For latest assets run `go generate ./assets` from project root folder
//go:generate statik -src=./public

// Package assets provides embedded assets
package assets

import (
	"html/template"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/rakyll/statik/fs"
	conf "github.com/yieldbot/ferret/config"
	// For assets
	_ "github.com/yieldbot/ferret/assets/statik"
)

var (
	config   conf.Assets
	statikFS http.FileSystem
)

// Init initializes the api
func Init(c conf.Config) {
	config = c.Assets

	var err error
	statikFS, err = fs.New()
	if err != nil {
		log.Fatal(err)
	}
}

// IndexHandler is the handler for entry point
func IndexHandler(w http.ResponseWriter, req *http.Request) {
	// Open file
	f, err := statikFS.Open("/index.html")
	if err != nil {
		log.Fatal(err)
	}

	// Read content
	b, err := ioutil.ReadAll(f)
	if err != nil {
		log.Fatal(err)
	}

	// Create and execute template
	t, err := template.New("index").Parse(string(b))
	if err != nil {
		log.Fatal(err)
	}

	data := struct {
		GATrackingCode string
		Menu           conf.AssetsMenu
	}{
		GATrackingCode: config.GATrackingCode,
		Menu:           config.Menu,
	}

	if err := t.Execute(w, data); err != nil {
		log.Fatal(err)
	}
}

// PublicHandler is the handler for the public files
func PublicHandler() http.Handler {
	return http.FileServer(statikFS)
}
