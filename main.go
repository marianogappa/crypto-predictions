package main

import (
	"embed"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"strconv"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/marianogappa/crypto-candles/candles"
	"github.com/marianogappa/predictions/api"
	"github.com/marianogappa/predictions/backoffice"
	"github.com/marianogappa/predictions/daemon"
	"github.com/marianogappa/predictions/imagebuilder"
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
	flagDaemonOnce = flag.Bool("daemononce", false, "only run Daemon once")
)

func main() {
	flag.Parse()
	loadEnvsFromConfigJSON()

	// Parse flags and figure out what components need to run.
	var (
		runAll        = (!*flagAPI && !*flagBackOffice && !*flagDaemon && !*flagDaemonOnce) || (*flagAPI && *flagBackOffice && *flagDaemon)
		runAPI        = runAll || *flagAPI
		runBackOffice = runAll || *flagBackOffice
		runDaemon     = runAll || *flagDaemon
		runDaemonOnce = !runAll && !*flagDaemon && *flagDaemonOnce
	)

	// Unless only the daemon is running, predictions may be created. If so, credentials for Twitter & Youtube APIs are
	// required, so that metadata from URLs can be fetched.
	if (runAPI || runBackOffice) && (os.Getenv("PREDICTIONS_TWITTER_BEARER_TOKEN") == "" || os.Getenv("PREDICTIONS_YOUTUBE_API_KEY") == "") {
		log.Fatal().Msg(
			`Please set the PREDICTIONS_TWITTER_BEARER_TOKEN && PREDICTIONS_YOUTUBE_API_KEY envs; I can't ` +
				`create predictions properly otherwise. If you're not planning to create predictions, just set them ` +
				`to any value.`,
		)
	}

	// If Back Office will run, but API won't, then there must be an addressable API in the network, and its URL/port
	// must be supplied.
	if runBackOffice && !runAPI && os.Getenv("PREDICTIONS_API_PORT") == "" && os.Getenv("PREDICTIONS_API_URL") == "" {
		log.Fatal().Msg(`If you want to run the Back Office but not the API, please specify either PREDICTIONS_API_URL or ` +
			"PREDICTIONS_API_PORT. Otherwise I don't know how to reach the API.")
	}

	osCurUser, err := user.Current()
	if err != nil {
		log.Error().Err(err).Msgf("Failed to get current user (ignoring).")
	}
	postgresConf := statestorage.PostgresConf{User: osCurUser.Username, Pass: "", Port: "5432", Database: osCurUser.Username, Host: "localhost"}
	postgresConf.User = envOrStr("PREDICTIONS_POSTGRES_USER", postgresConf.User)
	postgresConf.Pass = envOrStr("PREDICTIONS_POSTGRES_PASS", postgresConf.Pass)
	postgresConf.Port = envOrStr("PREDICTIONS_POSTGRES_PORT", postgresConf.Port)
	postgresConf.Database = envOrStr("PREDICTIONS_POSTGRES_DATABASE", postgresConf.Database)
	postgresConf.Host = envOrStr("PREDICTIONS_POSTGRES_HOST", postgresConf.Host)

	// Resolve & instantiate all components.
	var (
		// The state storage component is responsible for durably storing predictions.
		postgresDBStorage = statestorage.MustNewPostgresDBStateStorage(postgresConf)

		marketCacheSizes = map[time.Duration]int{
			time.Minute:    envOrInt("PREDICTIONS_MARKET_CACHE_SIZE_1_MINUTE", 10000),
			1 * time.Hour:  envOrInt("PREDICTIONS_MARKET_CACHE_SIZE_1_HOUR", 1000),
			24 * time.Hour: envOrInt("PREDICTIONS_MARKET_CACHE_SIZE_1_DAY", 1000),
		}

		// The market component queries all exchange APIs for market data.
		market = candles.NewMarket(candles.WithCacheSizes(marketCacheSizes))

		// The metadataFetcher component queries the Twitter/Youtube APIs for social post metadata, e.g. timestamps.
		metadataFetcher = metadatafetcher.NewMetadataFetcher()

		// The predictionImageBuilder builds images for actioning prediction event posts, e.g. created, finalised.
		chromePath             = envOrStr("PREDICTIONS_CHROME_PATH", envOrStr("GOOGLE_CHROME_BIN", ""))
		predictionImageBuilder = imagebuilder.NewPredictionImageBuilder(market, files, chromePath)

		// All endpoints across the API and the BackOffice (except for /predictions/pages) are behind BasicAuth.
		basicAuthUser = envOrStr("PREDICTIONS_BASIC_AUTH_USER", "admin")
		basicAuthPass = envOrStr("PREDICTIONS_BASIC_AUTH_PASS", "admin")

		// The API component is responsible for CRUDing predictions and related entities.
		api = api.NewAPI(market, postgresDBStorage, *metadataFetcher, predictionImageBuilder, basicAuthUser, basicAuthPass)

		// The Daemon component is responsible for continuously running prediction state machines against market data.
		enableTweeting = envOrStr("PREDICTIONS_DAEMON_ENABLE_TWEETING", "") != ""
		enableReplying = envOrStr("PREDICTIONS_DAEMON_ENABLE_REPLYING", "") != ""
		websiteURL     = envOrStr("PREDICTIONS_WEBSITE_URL", "")
		daemon         = daemon.NewDaemon(market, postgresDBStorage, predictionImageBuilder, enableTweeting, enableReplying, websiteURL)

		// The BackOffice component is a UI for admins to maintain the predictions system.
		backOffice = backoffice.NewBackOfficeUI(files, basicAuthUser, basicAuthPass)

		// Resolve all urls, ports & configs from environment variables, with defaults.
		apiPort        = envOrInt("PREDICTIONS_API_PORT", 2345)
		apiURL         = envOrStr("PREDICTIONS_API_URL", fmt.Sprintf("http://0.0.0.0:%v", apiPort))
		backOfficePort = envOrInt("PREDICTIONS_BACKOFFICE_PORT", 1234)
		daemonDuration = envOrDur("PREDICTIONS_DAEMON_DURATION", 60*time.Second)
	)

	if os.Getenv("PREDICTIONS_DEBUG") != "" {
		postgresDBStorage.SetDebug(true)
		backOffice.SetDebug(true)
		api.SetDebug(true)
		market.SetDebug(true)
	}

	// Run all components.
	if runAPI {
		go api.MustBlockinglyListenAndServe(apiURL)
	}

	if runBackOffice {
		go backOffice.MustBlockinglyServe(backOfficePort, apiURL)
	}

	if !runDaemonOnce && runDaemon {
		go daemon.BlockinglyRunEvery(daemonDuration)
	}

	if runDaemonOnce {
		daemon.Run(int(time.Now().Unix()))
	} else {
		select {}
	}
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

func loadEnvsFromConfigJSON() {
	file, err := ioutil.ReadFile("config.json")
	if err != nil {
		return
	}
	envMap := map[string]string{}
	err = json.Unmarshal(file, &envMap)
	if err != nil {
		return
	}
	for key, value := range envMap {
		os.Setenv(key, value)
	}
}
