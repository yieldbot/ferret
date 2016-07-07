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
	"log"
	"net/http"

	"github.com/rakyll/statik/fs"
	// For assets
	_ "github.com/yieldbot/ferret/assets/statik"
)

var (
	statikFS http.FileSystem
)

func init() {
	var err error
	statikFS, err = fs.New()
	if err != nil {
		log.Fatal(err)
	}
}

// PublicHandler is the handler for public assets
func PublicHandler() http.Handler {
	return http.FileServer(statikFS)
}
