package main

import (
	"log"
	"net/http"

	"github.com/boltdb/bolt"
	htr "github.com/julienschmidt/httprouter"
	"github.com/synapse-garden/mf-proto/api"
	"github.com/synapse-garden/mf-proto/db"
)

func main() {
	d, err := bolt.Open("my.db", 0600, nil)
	if err != nil {
		log.Fatalf("setting up db failed: %s", err.Error())
	}
	defer d.Close()

	go runHTTPListeners(d)

	// TODO: Handle console input
	select {}
}

func runHTTPListeners(d db.DB) {
	httpMux := htr.New()
	httpsMux := htr.New()

	// httpMux.SetupRoutes()
	err := api.SetupRoutes(
		httpsMux,
		api.Auth(d),
		api.User(d),
		api.Task(d),
	)

	if err != nil {
		log.Fatalf("router setup failed: %s\n", err.Error())
	}

	var (
		httpErr  = make(chan error)
		httpsErr = make(chan error)
	)

	println("Listening on HTTP 25000")
	println("Listening on HTTPS 25001")

	go func() { httpErr <- http.ListenAndServeTLS(":25001", "cert.pem", "key.key", httpsMux) }()
	go func() { httpsErr <- http.ListenAndServe(":25000", httpMux) }()

	var e error
	select {
	case e = <-httpErr:
	case e = <-httpsErr:
	}
	log.Fatal("error serving http(s): %s", e.Error())
}
