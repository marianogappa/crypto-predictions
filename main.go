package main

import (
	"embed"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/marianogappa/predictions/api"
	"github.com/marianogappa/predictions/backoffice"
	"github.com/marianogappa/predictions/market"
	"github.com/marianogappa/predictions/smrunner"
	"github.com/marianogappa/predictions/statestorage"
)

var (
	//go:embed public/*
	files embed.FS
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	var (
		postgresDBStorage = statestorage.MustNewPostgresDBStateStorage()
		market            = market.NewMarket()
		api               = api.NewAPI(market, postgresDBStorage)
		backgroundRunner  = smrunner.NewSMRunner(market, postgresDBStorage)
		backOffice        = backoffice.NewBackOfficeUI(files)

		apiPort                  = envOrInt("PREDICTIONS_API_PORT", 2345)
		apiUrl                   = fmt.Sprintf("http://localhost:%v", apiPort)
		backOfficePort           = envOrInt("PREDICTIONS_BACKOFFICE_PORT", 1234)
		backgroundRunnerDuration = envOrDur("PREDICTIONS_BACKGROUND_RUNNER_DURATION", 10*time.Second)
	)

	go api.MustBlockinglyServe(apiPort)
	go backOffice.MustBlockinglyServe(backOfficePort, apiUrl)
	go backgroundRunner.BlockinglyRunEvery(backgroundRunnerDuration)

	select {}
}

func envOrInt(s string, or int) int {
	i, err := strconv.Atoi(os.Getenv(s))
	if err != nil {
		return or
	}
	return i
}

func envOrDur(s string, or time.Duration) time.Duration {
	d, err := time.ParseDuration(os.Getenv(s))
	if err != nil {
		return or
	}
	return d
}
