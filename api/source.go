package api

import (
	"fmt"
	"net/http"

	htr "github.com/julienschmidt/httprouter"

	"github.com/synapse-garden/mf-proto/db"
)

const source = `---     S Y N A P S E G A R D E N      ---
            MF-Proto v0.3.0
         © SynapseGarden 2015
 Licensed under Affero GNU Public License
                version 3
https://github.com/synapse-garden/mf-proto
---                                    ---
`

func Source(d db.DB) API {
	return func(r *htr.Router) error {
		r.GET("/source", handleSource)
		return nil
	}
}

func handleSource(w http.ResponseWriter, r *http.Request, ps htr.Params) {
	fmt.Fprint(w, source)
}
