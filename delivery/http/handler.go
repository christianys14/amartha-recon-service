package http

import (
	"amartha-recon-service/configuration"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

type billingHandler struct {
	configuration configuration.Configuration
}

func NewBillingHandler(
	configuration configuration.Configuration) *billingHandler {
	return &billingHandler{
		configuration: configuration,
	}
}

func (b *billingHandler) showVersion() {
	version := b.configuration.GetString("app.recon.version")
	log.Println("show-recon-version -> ", version)
}

func (b *billingHandler) BuildHttp(router *mux.Router) http.Handler {
	b.showVersion()
	b.routeBilling(router)
	return router
}
