package api

import (
	htr "github.com/julienschmidt/httprouter"
	"github.com/synapse-garden/mf-proto/db"
)

// TODO: Fixme!
func User(d db.DB) API {
	return func(r *htr.Router) error {
		return nil
	}
}
