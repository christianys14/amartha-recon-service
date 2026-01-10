package cmd

import (
	"amartha-recon-service/application/recon"
	"amartha-recon-service/configuration"
	"amartha-recon-service/delivery/http"
	"amartha-recon-service/infrastructure/repository/transaction"
	"context"
	"errors"
	"log"
	http2 "net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/spf13/cobra"
)

var serveHttp = &cobra.Command{
	Use:   "serveHttp",
	Short: "Turn on amartha Recon service HTTP Rest API",
	Long:  "Cobra CLI : turn on Recon service HTTP Rest API",
	Run: func(cmd *cobra.Command, args []string) {
		//init configuration and credential
		cfg, cre := fetchConfiguration()

		//init database master
		initDB := configuration.NewStoreImpl(cre)
		dbMaster, err := initDB.InitDBMaster()

		if err != nil {
			panic(err)
		}

		transactionRepository := transaction.NewTransactionRepository(dbMaster)
		transactionService := recon.NewService(cfg, transactionRepository)
		transactionController := http.NewController(transactionService)

		reconHttpServerAddress := cfg.GetString("server.address.http")
		router := mux.NewRouter()

		reconHandler := http.NewReconHandler(cfg, transactionController).BuildHttp(router)
		reconHttpServer := http2.Server{
			Addr:         reconHttpServerAddress,
			Handler:      reconHandler,
			WriteTimeout: 15 * time.Second,
			ReadTimeout:  15 * time.Second,
		}

		go func() {
			log.Println(
				"[Recon Service HTTP] server started. Listening on port",
				reconHttpServerAddress)

			if err := reconHttpServer.ListenAndServe(); err != nil &&
				!errors.Is(err, http2.ErrServerClosed) {
				log.Println("error on close http : " + err.Error())
			}
		}()

		done := make(chan os.Signal, 1)
		signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

		<-done
		if err := reconHttpServer.Shutdown(context.Background()); err != nil {
			log.Println("[Recon Service HTTP], shutdown has error", err)
		} else {
			log.Println("[Recon Service HTTP] server stopped.")
		}
	},
}
