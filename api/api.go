package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/juju/errors"
	htr "github.com/julienschmidt/httprouter"
)

// API defines how a new API will be attached to the router.
// api packages should export a function such as:
//
// func Auth(d db.DB) API {
//	return func(r *htr.Router) error {
// 		r.GET("/user/auth", handleAuth(d))
//	}
// }
type API func(*htr.Router) error

func Routes(apis ...API) (*htr.Router, error) {
	r := htr.New()
	for _, a := range apis {
		if err := a(r); err != nil {
			return nil, err
		}
	}
	return r, nil
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
		log.Printf("failed to write response: %s", err.Error())
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
