package api

import (
	"log"
	"net/http"

	htr "github.com/julienschmidt/httprouter"
	"github.com/synapse-garden/mf-proto/db"
	"github.com/synapse-garden/mf-proto/object"
	"github.com/synapse-garden/mf-proto/user"
	"github.com/synapse-garden/mf-proto/util"
)

// Object binds the Object database package for the given DB to a Router.
func Object(d db.DB) API {
	return func(r *htr.Router) error {
		if err := db.SetupBuckets(d, object.Buckets()); err != nil {
			return err
		}

		r.PUT("/object/:id", handleObjectPut(d))
		r.DELETE("/object/:id", handleObjectDelete(d))
		r.GET("/object/:id", handleObjectGet(d))
		return nil
	}
}

func handleObjectPut(d db.DB) htr.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps htr.Params) {
		if err := r.ParseForm(); err != nil {
			WriteResponse(w, newApiError(err.Error(), err))
			log.Printf("bad object request: %#v", r)
			return
		}

		email, key := r.Form.Get("email"), util.Key(r.Form.Get("key"))
		if err := user.ValidLogin(d, email, key); err != nil {
			WriteResponse(w, newApiError(err.Error(), err))
			log.Printf("bad login: %#v", r)
			return
		}

		id := util.Key(ps.ByName("id"))
		obj := object.New(r.Form.Get("json"), email)

		if err := object.Put(d, email, id, obj); err != nil {
			WriteResponse(w, newApiError(err.Error(), err))
			log.Printf(
				"error storing object:\n  %s\n  %s",
				id, err.Error(),
			)
			return
		}

		log.Printf("object %s stored with id %s", obj, id)
		WriteResponse(w, obj)
	}
}

func handleObjectGet(d db.DB) htr.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps htr.Params) {
		if err := r.ParseForm(); err != nil {
			WriteResponse(w, newApiError(err.Error(), err))
			log.Printf("bad object request: %#v", r)
			return
		}

		email, key := r.Form.Get("email"), util.Key(r.Form.Get("key"))
		if err := user.ValidLogin(d, email, key); err != nil {
			WriteResponse(w, newApiError(err.Error(), err))
			log.Printf("bad login: %#v", r)
			return
		}

		id := util.Key(ps.ByName("id"))

		obj, err := object.Get(d, email, id)
		if err != nil {
			WriteResponse(w, newApiError(err.Error(), err))
			log.Printf("error fetching object: %#v", r)
			return
		}

		log.Printf("fetched object %s:\n  %s", id, obj)
		WriteResponse(w, obj)
	}
}

func handleObjectDelete(d db.DB) htr.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps htr.Params) {
		if err := r.ParseForm(); err != nil {
			WriteResponse(w, newApiError(err.Error(), err))
			log.Printf("bad request: %#v", r)
			return
		}

		email, key := r.Form.Get("email"), util.Key(r.Form.Get("key"))
		if err := user.ValidLogin(d, email, key); err != nil {
			WriteResponse(w, newApiError(err.Error(), err))
			log.Printf("bad login: %#v", r)
			return
		}

		id := ps.ByName("id")

		if err := object.Delete(d, email, util.Key(id)); err != nil {
			WriteResponse(w, newApiError(err.Error(), err))
			log.Printf("error deleting object %s: %s", id, err.Error())
			return
		}

		log.Printf("object %s deleted", id)
		WriteResponse(w, id)
	}
}
