package http

import (
	"amartha-recon-service/configuration"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

type reconHandler struct {
	configuration configuration.Configuration
	controller    Controller
}

func NewReconHandler(
	configuration configuration.Configuration,
	controller Controller) *reconHandler {
	return &reconHandler{
		configuration: configuration,
		controller:    controller,
	}
}

func (b *reconHandler) showVersion() {
	version := b.configuration.GetString("app.recon.version")
	log.Println("show-recon-version -> ", version)
}

func (b *reconHandler) BuildHttp(router *mux.Router) http.Handler {
	b.showVersion()
	b.routeRecon(router)
	return router
}
