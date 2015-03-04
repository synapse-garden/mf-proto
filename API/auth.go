package API

import (
	"net/http"

	htr "github.com/julienschmidt/httprouter"

	"github.com/synapse-garden/mf-proto/db"
)

type AuthAPI struct{ db db.DB }

func (t *AuthAPI) AddRoutes(router *htr.Router) {
	router.GET("/user/auth", authUser)
}

func authUser(w http.ResponseWriter, r *http.Request, ps htr.Params) {
	println("Do something")
}
