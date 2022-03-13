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
	// Static assets are embedded into the binary, so that the whole engine can be powered by a single file.
	//go:embed public/*
	files embed.FS

	// These flags enable the binary to run any combination of components. Running individual components can help
	// spread resource load over different instances. Don't use them if you want everything to run at once.
	flagAPI        = flag.Bool("api", false, "only run API")
	flagBackOffice = flag.Bool("backoffice", false, "only run Back Office")
	flagDaemon     = flag.Bool("daemon", false, "only run Daemon")
)

func main() {
	// Logging file and line numbers on STDERR is very useful for debugging.
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	flag.Parse()

	// Parse flags and figure out what components need to run.
	var (
		runAll        = (!*flagAPI && !*flagBackOffice && !*flagDaemon) || (*flagAPI && *flagBackOffice && *flagDaemon)
		runAPI        = runAll || *flagAPI
		runBackOffice = runAll || *flagBackOffice
		runDaemon     = runAll || *flagDaemon
	)

	// Unless only the daemon is running, predictions may be created. If so, credentials for Twitter & Youtube APIs are
	// required, so that metadata from URLs can be fetched.
	if (runAPI || runBackOffice) && (os.Getenv("PREDICTIONS_TWITTER_BEARER_TOKEN") == "" || os.Getenv("PREDICTIONS_YOUTUBE_API_KEY") == "") {
		log.Fatal(
			`Please set the PREDICTIONS_TWITTER_BEARER_TOKEN && PREDICTIONS_YOUTUBE_API_KEY envs; I can't ` +
				`create predictions properly otherwise. If you're not planning to create predictions, just set them ` +
				`to any value.`,
		)
	}

	// If Back Office will run, but API won't, then there must be an addressable API in the network, and its URL/port
	// must be supplied.
	if runBackOffice && !runAPI && os.Getenv("PREDICTIONS_API_PORT") == "" && os.Getenv("PREDICTIONS_API_URL") == "" {
		log.Fatal(`If you want to run the Back Office but not the API, please specify either PREDICTIONS_API_URL or ` +
			"PREDICTIONS_API_PORT. Otherwise I don't know how to reach the API.")
	}

	// Resolve & instantiate all components.
	var (
		// The state storage component is responsible for durably storing predictions.
		postgresDBStorage = statestorage.MustNewPostgresDBStateStorage()

		// The market component queries all exchange APIs for market data.
		market = market.NewMarket()

		// The metadataFetcher component queries the Twitter/Youtube APIs for social post metadata, e.g. timestamps.
		metadataFetcher = metadatafetcher.NewMetadataFetcher()

		// The API component is responsible for CRUDing predictions and related entities.
		api = api.NewAPI(market, postgresDBStorage, *metadataFetcher)

		// The Daemon component is responsible for continuously running prediction state machines against market data.
		daemon = daemon.NewDaemon(market, postgresDBStorage)

		// The BackOffice component is a UI for admins to maintain the predictions system.
		backOffice = backoffice.NewBackOfficeUI(files)

		// Resolve all urls, ports & configs from environment variables, with defaults.
		apiPort        = envOrInt("PREDICTIONS_API_PORT", 2345)
		apiUrl         = envOrStr("PREDICTIONS_API_URL", fmt.Sprintf("http://localhost:%v", apiPort))
		backOfficePort = envOrInt("PREDICTIONS_BACKOFFICE_PORT", 1234)
		daemonDuration = envOrDur("PREDICTIONS_DAEMON_DURATION", 10*time.Second)
	)

	// Run all components.
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
