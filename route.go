package main

import (
	"log"
	"net/http"

	"github.com/rs/cors"
	"github.com/synapse-garden/mf-proto/api"
	"github.com/synapse-garden/mf-proto/db"
)

func runHTTPListeners(d db.DB) {
	httpMux, err := api.Routes(api.Source(d))
	if err != nil {
		log.Fatalf("router setup failed: %s\n", err.Error())
	}

	httpsMux, err := api.Routes(
		api.Admin(d),
		api.User(d),
		api.Object(d),
		api.Task(d),
		api.Source(d),
	)
	if err != nil {
		log.Fatalf("router setup failed: %s\n", err.Error())
	}

	httpErr := make(chan error)
	httpsErr := make(chan error)

	defaultCORSOptions := cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "PUT", "POST", "DELETE"},
		AllowCredentials: true,
	}

	httpMuxCORS := cors.New(defaultCORSOptions).Handler(httpMux)
	httpsMuxCORS := cors.New(defaultCORSOptions).Handler(httpsMux)

	log.Printf("mf-proto hosting source on HTTP 25000")
	log.Printf("mf-proto listening on HTTPS 25001")

	go func() { httpsErr <- http.ListenAndServeTLS(":25001", "cert.pem", "key.key", httpsMuxCORS) }()
	go func() { httpErr <- http.ListenAndServe(":25000", httpMuxCORS) }()

	go func() {
		var e error
		select {
		case e = <-httpErr:
		case e = <-httpsErr:
		}
		log.Fatalf("error serving http(s): %s", e.Error())
	}()
}
