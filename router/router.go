package router

import htr "github.com/julienschmidt/httprouter"

type API interface {
	AddRoutes(*htr.Router)
}

func SetupRoutes(router *htr.Router, apis ...API) {
	for _, api := range apis {
		api.AddRoutes(router)
	}
}
