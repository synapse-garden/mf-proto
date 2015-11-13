package main

import (
	"io"
	"log"
	"net/http"

	htr "github.com/julienschmidt/httprouter"
	"github.com/synapse-garden/mf-proto/api"
	"github.com/synapse-garden/mf-proto/db"
)

const sourceDoc = `---     S Y N A P S E G A R D E N     ---

            MF-Proto v0.2.0  
         © SynapseGarden 2015

 Licensed under Affero GNU Public License
                version 3

https://github.com/synapse-garden/mf-proto

---                                    ---
`

func source(r *htr.Router) error {
	r.GET("/source",
		func(w http.ResponseWriter, r *http.Request, ps htr.Params) {
			if _, err := io.WriteString(w, sourceDoc); err != nil {
				log.Printf("failed to write response: %s", err.Error())
			}
		},
	)

	return nil
}

func runHTTPListeners(d db.DB) {
	httpMux, err := api.Routes(source)
	if err != nil {
		log.Fatalf("router setup failed: %s\n", err.Error())
	}

	httpsMux, err := api.Routes(
		api.Admin(d),
		api.User(d),
		api.Object(d),
		api.Task(d),
	)
	if err != nil {
		log.Fatalf("router setup failed: %s\n", err.Error())
	}

	var (
		httpErr  = make(chan error)
		httpsErr = make(chan error)
	)

	log.Printf("mf-proto hosting source on HTTP 25000")
	log.Printf("mf-proto listening on HTTPS 25001")

	go func() { httpsErr <- http.ListenAndServeTLS(":25001", "cert.pem", "key.key", httpsMux) }()
	go func() { httpErr <- http.ListenAndServe(":25000", httpMux) }()

	go func() {
		var e error
		select {
		case e = <-httpErr:
		case e = <-httpsErr:
		}
		log.Fatalf("error serving http(s): %s", e.Error())
	}()
}
