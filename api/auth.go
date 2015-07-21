package api

import (
	"net/http"

	htr "github.com/julienschmidt/httprouter"
	"github.com/segmentio/go-log"
	"github.com/synapse-garden/mf-proto/auth"
	"github.com/synapse-garden/mf-proto/db"
	"github.com/synapse-garden/mf-proto/util"
)

func Auth(d db.DB) API {
	return func(r *htr.Router) error {
		err := db.SetupBuckets(d, auth.Buckets())
		if err != nil {
			return err
		}
		r.GET("/user/create", handleCreate(d))
		r.GET("/user/delete", handleDelete(d))
		r.GET("/user/valid", handleValid(d))
		r.GET("/user/login", handleLogin(d))
		r.GET("/user/logout", handleLogout(d))
		return nil
	}
}

func handleCreate(d db.DB) htr.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps htr.Params) {
		err := r.ParseForm()
		if err != nil {
			WriteResponse(w, newApiError(err.Error(), err))
			log.Error("bad request: %#v", r)
			return
		}
		email := r.Form.Get("email")
		pwhash := r.Form.Get("pwhash")
		// Check whether the user exists
		// If it does, fail
		// If it doesn't, create it
		err = auth.CreateUser(d, email, pwhash)
		if err != nil {
			WriteResponse(w, newApiError(err.Error(), err))
			log.Error("error authenticating user %q, pwhash %q: %s", email, pwhash, err.Error())
			return
		}

		log.Info("user %q created", email)
		WriteResponse(w, &auth.User{
			Email: email,
		})
	}
}

func handleDelete(d db.DB) htr.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps htr.Params) {
		err := r.ParseForm()
		if err != nil {
			WriteResponse(w, newApiError(err.Error(), err))
			log.Error("bad request: %#v", r)
			return
		}
		email := r.Form.Get("email")

		err = auth.DeleteUser(d, email)
		if err != nil {
			WriteResponse(w, newApiError(err.Error(), err))
			log.Error("error deleting user %q: %s", email, err.Error())
			return
		}

		log.Info("user %q deleted", email)
		WriteResponse(w, &auth.User{
			Email: email,
		})
	}
}

func handleValid(d db.DB) htr.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps htr.Params) {
		err := r.ParseForm()
		if err != nil {
			WriteResponse(w, newApiError(err.Error(), err))
			log.Error("bad request: %#v", r)
			return
		}
		email := r.Form.Get("email")
		key := util.Key(r.Form.Get("key"))
		err = auth.Valid(d, email, key)
		if err != nil {
			WriteResponse(w, newApiError(err.Error(), err))
			log.Error("error authenticating user %q, key %q: %s", email, key, err.Error())
			return
		}

		log.Info("user %q validated", email)
		WriteResponse(w, &auth.User{
			Email: email,
		})
	}
}

func handleLogin(d db.DB) htr.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps htr.Params) {
		err := r.ParseForm()
		if err != nil {
			WriteResponse(w, newApiError("bad request: "+err.Error(), err))
			log.Error("bad request: %#v", r)
			return
		}
		email := r.Form.Get("email")
		pwhash := r.Form.Get("pwhash")
		key, err := auth.LoginUser(d, email, pwhash)
		if err != nil {
			WriteResponse(w, newApiError(err.Error(), err))
			log.Error("error logging in user %q, pwhash %q: %s", email, pwhash, err.Error())
			return
		}
		log.Info("user %q logged in", email)
		WriteResponse(w, &auth.User{
			Email: email,
			Key:   key,
		})
	}
}

func handleLogout(d db.DB) htr.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps htr.Params) {
		err := r.ParseForm()
		if err != nil {
			WriteResponse(w, newApiError("bad request: "+err.Error(), err))
			log.Error("bad request: %#v", r)
			return
		}
		email := r.Form.Get("email")
		key := util.Key(r.Form.Get("key"))
		err = auth.LogoutUser(d, email, key)
		if err != nil {
			WriteResponse(w, newApiError(err.Error(), err))
			log.Error("user %q logout for key %q failed: %s", email, key, err.Error())
			return
		}
		WriteResponse(w, &auth.User{
			Email: email,
		})
	}
}
