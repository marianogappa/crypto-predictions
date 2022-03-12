package main

import (
	"embed"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/marianogappa/predictions/api"
	"github.com/marianogappa/predictions/backoffice"
	"github.com/marianogappa/predictions/daemon"
	"github.com/marianogappa/predictions/market"
	"github.com/marianogappa/predictions/metadatafetcher"
	"github.com/marianogappa/predictions/statestorage"
)

var (
	//go:embed public/*
	files embed.FS

	flagAPI        = flag.Bool("api", false, "only run API")
	flagBackOffice = flag.Bool("backoffice", false, "only run Back Office")
	flagDaemon     = flag.Bool("daemon", false, "only run Daemon")
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	flag.Parse()

	var (
		runAll        = (!*flagAPI && !*flagBackOffice && !*flagDaemon) || (*flagAPI && *flagBackOffice && *flagDaemon)
		runAPI        = runAll || *flagAPI
		runBackOffice = runAll || *flagBackOffice
		runDaemon     = runAll || *flagDaemon
	)

	if os.Getenv("PREDICTIONS_TWITTER_BEARER_TOKEN") == "" || os.Getenv("PREDICTIONS_YOUTUBE_API_KEY") == "" {
		log.Fatal(
			`Please set the PREDICTIONS_TWITTER_BEARER_TOKEN && PREDICTIONS_YOUTUBE_API_KEY envs; I can't ` +
				`create predictions properly otherwise. If you're not planning to create predictions, just set them ` +
				`to any value.`,
		)
	}

	if runBackOffice && !runAPI && os.Getenv("PREDICTIONS_API_PORT") == "" && os.Getenv("PREDICTIONS_API_URL") == "" {
		log.Fatal(`If you want to run the Back Office but not the API, please specify either PREDICTIONS_API_URL or ` +
			"PREDICTIONS_API_PORT. Otherwise I don't know how to reach the API.")
	}

	var (
		postgresDBStorage = statestorage.MustNewPostgresDBStateStorage()
		market            = market.NewMarket()
		metadataFetcher   = metadatafetcher.NewMetadataFetcher()
		api               = api.NewAPI(market, postgresDBStorage, *metadataFetcher)
		daemon            = daemon.NewDaemon(market, postgresDBStorage)
		backOffice        = backoffice.NewBackOfficeUI(files)

		apiPort        = envOrInt("PREDICTIONS_API_PORT", 2345)
		apiUrl         = envOrStr("PREDICTIONS_API_URL", fmt.Sprintf("http://localhost:%v", apiPort))
		backOfficePort = envOrInt("PREDICTIONS_BACKOFFICE_PORT", 1234)
		daemonDuration = envOrDur("PREDICTIONS_DAEMON_DURATION", 10*time.Second)
	)

	if runAPI {
		go api.MustBlockinglyListenAndServe(apiUrl)
	}

	if runBackOffice {
		go backOffice.MustBlockinglyServe(backOfficePort, apiUrl)
	}

	if runDaemon {
		go daemon.BlockinglyRunEvery(daemonDuration)
	}

	select {}
}

func envOrStr(env, or string) string {
	s := os.Getenv(env)
	if s == "" {
		return or
	}
	return s
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
