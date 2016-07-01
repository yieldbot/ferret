/*
 * Ferret
 * Copyright (c) 2016 Yieldbot, Inc.
 * For the full copyright and license information, please view the LICENSE.txt file.
 */

// Package api provides REST API functionality
package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/yieldbot/ferret/assets"
	"github.com/yieldbot/ferret/search"
	"golang.org/x/net/context"
)

var (
	listenPort = "3030"
)

type httpError struct {
	StatusCode int    `json:"statusCode"`
	Error      string `json:"error"`
	Message    string `json:"message"`
}

func init() {
	if e := os.Getenv("FERRET_LISTEN_PORT"); e != "" {
		listenPort = e
	}
}

// Listen initializes HTTP handlers and listens for the requests
func Listen() {
	// Init handlers
	http.Handle("/", assets.PublicHandler())
	http.HandleFunc("/search", SearchHandler)

	// Listen
	log.Printf("listening on %s", listenPort)
	if err := http.ListenAndServe(fmt.Sprintf(":%s", listenPort), nil); err != nil {
		log.Fatal(err)
	}
}

// SearchHandler is the handler for search route
func SearchHandler(w http.ResponseWriter, req *http.Request) {

	// Search
	q := search.Query{
		Provider: req.URL.Query().Get("provider"),
		Keyword:  req.URL.Query().Get("keyword"),
		Page:     search.ParsePage(req.URL.Query().Get("page")),
		Timeout:  search.ParseTimeout(req.URL.Query().Get("timeout")),
	}
	q, err := search.Do(context.Background(), q)
	if err != nil {
		w.WriteHeader(q.HTTPStatus)
		data, _ := json.Marshal(httpError{
			StatusCode: q.HTTPStatus,
			Error:      http.StatusText(q.HTTPStatus),
			Message:    err.Error(),
		})
		HandleResponse(w, req, data)
		return
	}

	// Prepare data
	var data []byte
	if len(q.Results) > 0 {
		if req.URL.Query().Get("output") == "pretty" {
			data, err = json.MarshalIndent(q.Results, "", "  ")
		} else {
			data, err = json.Marshal(q.Results)
		}
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			data, _ := json.Marshal(httpError{
				StatusCode: http.StatusInternalServerError,
				Error:      http.StatusText(http.StatusInternalServerError),
				Message:    err.Error(),
			})
			HandleResponse(w, req, data)
			return
		}
	}

	HandleResponse(w, req, data)
}

// HandleResponse handles HTTP responses
func HandleResponse(w http.ResponseWriter, req *http.Request, data []byte) {
	cb := req.URL.Query().Get("callback")
	if cb != "" {
		w.Header().Set("Content-Type", "application/javascript")
		fmt.Fprintf(w, "%s(%s)", cb, data)
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.Write(data)
	}
}
