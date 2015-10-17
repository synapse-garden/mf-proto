package api

import (
	"log"
	"net/http"

	htr "github.com/julienschmidt/httprouter"
	"github.com/synapse-garden/mf-proto/admin"
	"github.com/synapse-garden/mf-proto/db"
	"github.com/synapse-garden/mf-proto/util"
)

func Admin(d db.DB) API {
	return func(r *htr.Router) error {
		if err := db.SetupBuckets(d, admin.Buckets()); err != nil {
			return err
		}
		r.GET("/admin/valid", handleAdminValid(d))
		r.GET("/admin/create", handleAdminCreate(d))
		r.GET("/admin/delete", handleAdminDelete(d))
		return nil
	}
}

func handleAdminValid(d db.DB) htr.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps htr.Params) {
		if err := r.ParseForm(); err != nil {
			WriteResponse(w, newApiError(err.Error(), err))
			log.Printf("bad admin request: %#v", r)
			return
		}

		key := r.Form.Get("key")
		if err := admin.IsAdmin(d, util.Key(key)); err != nil {
			WriteResponse(w, newApiError(err.Error(), err))
			log.Printf("bad admin request: %s", err.Error())
			return
		}

		WriteResponse(w, "ok")
		log.Printf("admin %s verified", key)
	}
}

func handleAdminCreate(d db.DB) htr.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps htr.Params) {
		if err := r.ParseForm(); err != nil {
			WriteResponse(w, newApiError(err.Error(), err))
			log.Printf("bad admin request: %#v", r)
			return
		}

		key := util.Key(r.Form.Get("key"))
		if err := admin.IsAdmin(d, key); err != nil {
			WriteResponse(w, newApiError(err.Error(), err))
			log.Printf("bad admin request: %s", err.Error())
			return
		}

		email := r.Form.Get("email")
		pwhash := r.Form.Get("pwhash")

		key, err := admin.Create(d, email, pwhash)
		if err != nil {
			WriteResponse(w, newApiError(err.Error(), err))
			log.Printf("error creating admin %s: %s", email, err.Error())
			return
		}

		log.Printf("admin %s created with key %s", email, key)
		WriteResponse(w, &admin.Admin{
			Email: email,
			Hash:  util.Hash(pwhash),
			Key:   key,
		})
	}
}

func handleAdminDelete(d db.DB) htr.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps htr.Params) {
		if err := r.ParseForm(); err != nil {
			WriteResponse(w, newApiError(err.Error(), err))
			log.Printf("bad admin request: %#v", r)
			return
		}

		key := util.Key(r.Form.Get("key"))
		if err := admin.IsAdmin(d, key); err != nil {
			WriteResponse(w, newApiError(err.Error(), err))
			log.Printf("bad admin request: %s", err.Error())
			return
		}

		if err := admin.Delete(d, key); err != nil {
			WriteResponse(w, newApiError(err.Error(), err))
			log.Printf("error deleting admin for %s: %s", key, err.Error())
			return
		}

		log.Printf("admin %s deleted", key)
		WriteResponse(w, &admin.Admin{
			Key: key,
		})
	}
}
