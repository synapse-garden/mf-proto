package api

import (
	"net/http"

	htr "github.com/julienschmidt/httprouter"
	"github.com/synapse-garden/mf-proto/auth"
	"github.com/synapse-garden/mf-proto/db"
)

func Auth(d db.DB) API {
	return func(r *htr.Router) error {
		err := db.SetupBuckets(d, buckets())
		if err != nil {
			return err
		}
		r.GET("/user/auth", handleAuth(d))
		return nil
	}
}

func buckets() []string {
	return []string{"session-keys"}
}

func handleAuth(d db.DB) htr.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps htr.Params) {
		err := r.ParseForm()
		if err != nil {
			http.Error(w, "parsing form failed: "+err.Error(), 400)
			return
		}
		email := r.Form.Get("email")
		pwhash := r.Form.Get("pwhash")
		key, err := auth.AuthUser(d, email, pwhash)
		if err != nil {
			http.Error(w, "generating key failed: "+err.Error(), 500)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "secure-key",
			Path:     "/",
			Value:    string(key),
			MaxAge:   5 * 60,
			Secure:   true,
			HttpOnly: true,
		})
	}
}
