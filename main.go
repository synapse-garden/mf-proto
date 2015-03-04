package main

import (
	"log"
	"net/http"

	"github.com/boltdb/bolt"
	htr "github.com/julienschmidt/httprouter"

	"github.com/synapse-garden/mf-proto/API"
	"github.com/synapse-garden/mf-proto/router"
)

func main() {
	db, err := bolt.Open("my.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	go runHTTPListeners(db)
	// Handle console input?
	select {}
}

func runHTTPListeners(db *bolt.DB) {
	httpMux := htr.New()
	httpsMux := htr.New()

	router.SetupRoutes(httpMux,
		&API.UserAPI{},
		&API.TaskAPI{},
	)
	router.SetupRoutes(httpsMux,
		&API.AuthAPI{},
	)

	var (
		err    = make(chan error)
		tlsErr = make(chan error)
	)

	println("Listening on HTTP 25000")
	println("Listening on HTTPS 25001")

	go func() { err <- http.ListenAndServeTLS(":25001", "cert.pem", "key.key", httpsMux) }()
	go func() { tlsErr <- http.ListenAndServe(":25000", httpMux) }()

	var e error
	select {
	case e = <-err:
	case e = <-tlsErr:
	}
	log.Fatal(e)
}
