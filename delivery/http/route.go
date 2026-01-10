package http

import (
	"net/http"

	"github.com/gorilla/mux"
)

func (b *reconHandler) routeRecon(r *mux.Router) {
	r.HandleFunc("/v1/internal/recon", b.controller.Proceed).Methods(http.MethodPost)
}
