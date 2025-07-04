package handlers

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (h Handler) Workload() httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {

	}
}
