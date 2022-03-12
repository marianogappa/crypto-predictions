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
	"github.com/marianogappa/predictions/metadatafetcher"
	"github.com/marianogappa/predictions/smrunner"
	"github.com/marianogappa/predictions/statestorage"
)

var (
	//go:embed public/*
	files embed.FS
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	if os.Getenv("PREDICTIONS_TWITTER_BEARER_TOKEN") == "" || os.Getenv("PREDICTIONS_YOUTUBE_API_KEY") == "" {
		log.Fatal(
			`Please set the PREDICTIONS_TWITTER_BEARER_TOKEN && PREDICTIONS_YOUTUBE_API_KEY envs; I can't ` +
				`create predictions properly otherwise. If you're not planning to create predictions, just set them ` +
				`to any value.`,
		)
	}

	var (
		postgresDBStorage = statestorage.MustNewPostgresDBStateStorage()
		market            = market.NewMarket()
		metadataFetcher   = metadatafetcher.NewMetadataFetcher()
		api               = api.NewAPI(market, postgresDBStorage, *metadataFetcher)
		backgroundRunner  = smrunner.NewSMRunner(market, postgresDBStorage)
		backOffice        = backoffice.NewBackOfficeUI(files)

		apiPort                  = envOrInt("PREDICTIONS_API_PORT", 2345)
		apiUrl                   = fmt.Sprintf("http://localhost:%v", apiPort)
		backOfficePort           = envOrInt("PREDICTIONS_BACKOFFICE_PORT", 1234)
		backgroundRunnerDuration = envOrDur("PREDICTIONS_BACKGROUND_RUNNER_DURATION", 10*time.Second)
	)

	go api.MustBlockinglyListenAndServe(apiUrl)
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
