package cmd

import (
	"amartha-recon-service/configuration"

	"github.com/spf13/cobra"
)

var serveHttp = &cobra.Command{
	Use:   "serveHttp",
	Short: "Turn on amartha billing service HTTP Rest API",
	Long:  "Cobra CLI : turn on Billing service HTTP Rest API",
	Run: func(cmd *cobra.Command, args []string) {
		//init configuration and credential
		_, cre := fetchConfiguration()

		//init database master
		initDB := configuration.NewStoreImpl(cre)
		_, err := initDB.InitDBMaster()

		if err != nil {
			panic(err)
		}

		//loanRepository := repository.NewLoanRepository(masterDB)
		//loanService := loan.NewLoanService(loanRepository)
		//loanController := loan.NewLoanController(loanService)
		//
		//billingHttpServerAddress := cfg.GetString("server.address.http")
		//router := mux.NewRouter()
		//
		//billingHandler := http.NewBillingHandler(cfg, loanController).BuildHttp(router)
		//billingHttpServer := http2.Server{
		//	Addr:    billingHttpServerAddress,
		//	Handler: billingHandler,
		//}
		//
		//go func() {
		//	log.Println(
		//		"[Billing Service HTTP] server started. Listening on port",
		//		billingHttpServerAddress)
		//
		//	if err := billingHttpServer.ListenAndServe(); err != nil &&
		//		!errors.Is(err, http2.ErrServerClosed) {
		//		log.Println("error on close http : " + err.Error())
		//	}
		//}()
		//
		//done := make(chan os.Signal, 1)
		//signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
		//
		//<-done
		//if err := billingHttpServer.Shutdown(context.Background()); err != nil {
		//	log.Println("[Billing Service HTTP], shutdown has error", err)
		//} else {
		//	log.Println("[Billing Service HTTP] server stopped.")
		//}
	},
}
