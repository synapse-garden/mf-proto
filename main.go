package main

import (
	"fmt"
	"net/http"

	htr "github.com/julienschmidt/httprouter"

	"github.com/synapse-garden/mf-proto/auth"
	"github.com/synapse-garden/mf-proto/router"
	"github.com/synapse-garden/mf-proto/task"
	"github.com/synapse-garden/mf-proto/user"
)

func main() {
	httpMux := htr.New()
	httpsMux := htr.New()

	router.SetupRoutes(httpMux, &user.UserAPI{}, &task.TaskAPI{})
	router.SetupRoutes(httpsMux, &auth.AuthAPI{})

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
	panic(e)
}

func home(w http.ResponseWriter, req *http.Request) {
	// The "/" pattern matches everything, so we need to check
	// that we're at the root here.
	if req.URL.Path != "/" {
		http.NotFound(w, req)
		return
	}
	fmt.Fprintf(w, "There is nothing here.")
}
