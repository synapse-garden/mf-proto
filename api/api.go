package api

import htr "github.com/julienschmidt/httprouter"

// API defines how a new API will be attached to the router.
// api packages should export a function such as:
//
// func Auth(d db.DB) API {
//	return func(r *htr.Router) {
// 		r.GET("/user/auth", handleAuth(d))
//	}
// }
type API func(*htr.Router) error

func SetupRoutes(r *htr.Router, apis ...API) error {
	for _, a := range apis {
		err := a(r)
		if err != nil {
			return err
		}
	}
	return nil
}
