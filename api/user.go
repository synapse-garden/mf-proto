package api

import (
	"log"
	"net/http"

	htr "github.com/julienschmidt/httprouter"
	"github.com/synapse-garden/mf-proto/admin"
	"github.com/synapse-garden/mf-proto/db"
	"github.com/synapse-garden/mf-proto/user"
	"github.com/synapse-garden/mf-proto/util"
)

func User(d db.DB) API {
	return func(r *htr.Router) error {
		if err := db.SetupBuckets(d, user.Buckets()); err != nil {
			return err
		}

		r.GET("/user/create", handleUserCreate(d))
		r.GET("/user/delete", handleUserDelete(d))
		r.GET("/user/valid", handleUserValid(d))
		r.GET("/user/login", handleUserLogin(d))
		r.GET("/user/logout", handleUserLogout(d))
		return nil
	}
}

func handleUserCreate(d db.DB) htr.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps htr.Params) {
		if err := r.ParseForm(); err != nil {
			WriteResponse(w, newApiError(err.Error(), err))
			log.Printf("bad admin request: %#v", r)
			return
		}

		key := util.Key(r.Form.Get("key"))
		if err := admin.IsAdmin(d, key); err != nil {
			WriteResponse(w, newApiError(err.Error(), err))
			log.Printf("bad admin request: %#v", r)
			return
		}

		email := r.Form.Get("email")
		pwhash := r.Form.Get("pwhash")

		if err := user.Create(d, email, pwhash); err != nil {
			WriteResponse(w, newApiError(err.Error(), err))
			log.Printf("error creating user %s: %s", email, err.Error())
			return
		}

		log.Printf("user %s created", email)
		WriteResponse(w, &user.User{
			Email: email,
			Hash:  util.Hash(pwhash),
		})
	}
}

func handleUserDelete(d db.DB) htr.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps htr.Params) {
		if err := r.ParseForm(); err != nil {
			WriteResponse(w, newApiError(err.Error(), err))
			log.Printf("bad admin request: %#v", r)
			return
		}

		key := util.Key(r.Form.Get("key"))
		email := r.Form.Get("email")
		pwhash := r.Form.Get("pwhash")

		switch {
		case key != "":
			// An admin can delete any user.
			if err := admin.IsAdmin(d, key); err != nil {
				WriteResponse(w, newApiError(err.Error(), err))
				log.Printf("bad admin request: %#v", r)
				return
			}

		case email != "", pwhash != "":
			if err := user.CheckUser(d, email, pwhash); err != nil {
				WriteResponse(w, newApiError(err.Error(), err))
				log.Printf("invalid user %s: %s", email, err.Error())
				return
			}
		default:
			// No key, no email, no pwhash -- no delete.
			WriteResponse(w, newApiError("must pass pwhash and email, or API key", nil))
			log.Printf("invalid user delete request: no values")
			return
		}

		if err := user.Delete(d, email); err != nil {
			WriteResponse(w, newApiError(err.Error(), err))
			log.Printf("error deleting user %q: %s", email, err.Error())
			return
		}

		log.Printf("user %q deleted", email)
		WriteResponse(w, &user.User{
			Email: email,
		})
	}
}

func handleUserValid(d db.DB) htr.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps htr.Params) {
		if err := r.ParseForm(); err != nil {
			WriteResponse(w, newApiError(err.Error(), err))
			log.Printf("bad admin request: %#v", r)
			return
		}

		email := r.Form.Get("email")
		key := util.Key(r.Form.Get("key"))

		if err := user.ValidLogin(d, email, key); err != nil {
			WriteResponse(w, newApiError(err.Error(), err))
			log.Printf("error authenticating user %q, key %q: %s", email, key, err.Error())
			return
		}

		log.Printf("user %q validated", email)
		WriteResponse(w, &user.User{
			Email: email,
		})
	}
}

func handleUserLogin(d db.DB) htr.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps htr.Params) {
		if err := r.ParseForm(); err != nil {
			WriteResponse(w, newApiError(err.Error(), err))
			log.Printf("bad admin request: %#v", r)
			return
		}

		email := r.Form.Get("email")
		pwhash := r.Form.Get("pwhash")
		key, err := user.LoginUser(d, email, pwhash)
		if err != nil {
			WriteResponse(w, newApiError(err.Error(), err))
			log.Printf("error logging in user %q, pwhash %q: %s", email, pwhash, err.Error())
			return
		}

		log.Printf("user %q logged in", email)
		WriteResponse(w, &user.User{
			Email: email,
			Key:   key,
		})
	}
}

func handleUserLogout(d db.DB) htr.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps htr.Params) {
		if err := r.ParseForm(); err != nil {
			WriteResponse(w, newApiError("bad request: "+err.Error(), err))
			log.Printf("bad request: %#v", r)
			return
		}

		email := r.Form.Get("email")
		key := util.Key(r.Form.Get("key"))

		if err := user.LogoutUser(d, email, key); err != nil {
			WriteResponse(w, newApiError(err.Error(), err))
			log.Printf("user %q logout for key %q failed: %s", email, key, err.Error())
			return
		}

		WriteResponse(w, &user.User{
			Email: email,
		})
	}
}
