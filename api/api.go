package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/juju/errors"
	htr "github.com/julienschmidt/httprouter"
	"github.com/segmentio/go-log"
)

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

type apiError struct {
	Code    int    `json:"code,omitempty"`
	Message string `json:"msg,omitempty"`
}

func newApiError(msg string, err error) apiError {
	e := apiError{
		Message: msg,
		Code:    http.StatusInternalServerError,
	}
	if errors.IsNotFound(err) {
		e.Message = fmt.Sprintf("not found: %s", msg)
		e.Code = http.StatusNotFound
	}

	return e
}

func WriteResponse(w http.ResponseWriter, values ...interface{}) {
	e := json.NewEncoder(w)
	response := newResponse(values...)
	if err := e.Encode(response); err != nil {
		code := http.StatusInternalServerError
		log.Error("failed to write response", err)
		http.Error(
			w,
			newFatalResponse("failed to write response", code),
			code,
		)
	}
}

type response struct {
	Code    int           `json:"code,omitempty"`
	Message string        `json:"msg,omitempty"`
	Values  []interface{} `json:"values,omitempty"`
}

func newResponse(v ...interface{}) response {
	return response{
		Values: v,
	}
}

func newFatalResponse(msg string, code int) string {
	r, _ := json.Marshal(response{
		Code:    code,
		Message: msg,
	})

	return string(r)
}
