package cmd

import (
	"amartha-recon-service/configuration"
	"log"
)

const (
	cfg = "configuration"
	cre = "credential"
)

func fetchConfiguration() (
	configuration.Configuration,
	configuration.Configuration) {
	cfg, err := configuration.FindConfiguration(cfg)
	if err != nil {
		log.Println("[MAIN] error retrieving configuration")
		panic(err)
	}

	cre, err := configuration.FindConfiguration(cre)
	if err != nil {
		log.Println("[MAIN] error retrieving credential")
		panic(err)
	}

	return cfg, cre
}
