package api

import (
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"os/user"
	"strings"
	"testing"
	"time"

	"github.com/marianogappa/predictions/market/common"
	fetcherTypes "github.com/marianogappa/predictions/metadatafetcher/types"
	"github.com/marianogappa/predictions/statestorage"
	"github.com/rs/zerolog/log"
)

var (
	store         statestorage.StateStorage
	mainTestStore = "test_predictions"
	apiTestStore  = "test_predictions_api"
)

func setupTestDB(t *testing.T) statestorage.StateStorage {
	if store == nil {
		mainStore := connectToTestStore(t, mainTestStore)
		createDatabase(t, mainStore.DB(), apiTestStore)
		store = connectToTestStore(t, apiTestStore)
	}
	mustTruncateTables(t, store.DB())
	return store
}

func connectToTestStore(t *testing.T, databaseName string) statestorage.StateStorage {
	if !strings.Contains(databaseName, "test_") {
		log.Error().Msgf("I'm not gonna let you connect to a non-test database!")
		return nil
	}
	osCurUser, err := user.Current()
	if err != nil {
		log.Error().Err(err).Msgf("Failed to get current user (ignoring).")
		return nil
	}
	postgresConf := statestorage.PostgresConf{User: osCurUser.Username, Pass: "", Port: "5432", Database: osCurUser.Username, Host: "localhost"}
	postgresConf.User = envOrStr("PREDICTIONS_POSTGRES_USER", postgresConf.User)
	postgresConf.Pass = envOrStr("PREDICTIONS_POSTGRES_PASS", postgresConf.Pass)
	postgresConf.Port = envOrStr("PREDICTIONS_POSTGRES_PORT", postgresConf.Port)
	postgresConf.Database = databaseName
	postgresConf.Host = envOrStr("PREDICTIONS_POSTGRES_HOST", postgresConf.Host)

	_store, err := statestorage.NewPostgresDBStateStorage(postgresConf)
	if err != nil {
		// Try a second time with root/test. This is because CI on Github is set up with this. Check ci.yml
		postgresConf.User = "root"
		postgresConf.Pass = "test"
		_store, err = statestorage.NewPostgresDBStateStorage(postgresConf)
		if err != nil {
			t.Fatal(err)
			return nil
		}
	}
	return castStore(&_store)
}

func castStore(s *statestorage.PostgresDBStateStorage) statestorage.StateStorage {
	return s
}

func mustTruncateTables(t *testing.T, db *sql.DB) {
	tables := []string{
		"predictions",
		"prediction_state_value_change",
		"accounts",
		"prediction_interactions",
	}
	for _, table := range tables {
		if _, err := db.Exec(fmt.Sprintf("TRUNCATE TABLE %v", table)); err != nil {
			t.Fatal(err)
		}
	}
}

func createDatabase(t *testing.T, db *sql.DB, databaseName string) {
	_, _ = db.Exec(fmt.Sprintf("CREATE DATABASE %v", databaseName))
}

type testFetcher struct {
	isCorrectFetcher bool
	postMetadata     fetcherTypes.PostMetadata
	err              error
}

func (t testFetcher) IsCorrectFetcher(url *url.URL) bool { return t.isCorrectFetcher }
func (t testFetcher) Fetch(url *url.URL) (fetcherTypes.PostMetadata, error) {
	return t.postMetadata, t.err
}

type testMarket struct {
	ticks map[string][]common.Tick
}

func newTestMarket(ticks map[string][]common.Tick) *testMarket {
	return &testMarket{ticks}
}

func (m *testMarket) GetIterator(marketSource common.MarketSource, startTime time.Time, startFromNext bool, intervalMinutes int) (common.Iterator, error) {
	if _, ok := m.ticks[marketSource.String()]; !ok {
		return nil, common.ErrInvalidMarketPair
	}
	return newTestIterator(m.ticks[marketSource.String()]), nil
}

type testIterator struct {
	ticks []common.Tick
}

func newTestIterator(ticks []common.Tick) common.Iterator {
	return &testIterator{ticks}
}

func (i *testIterator) NextTick() (common.Tick, error) {
	if len(i.ticks) > 0 {
		tick := i.ticks[0]
		i.ticks = i.ticks[1:]
		return tick, nil
	}
	return common.Tick{}, common.ErrOutOfTicks
}

func (i *testIterator) NextCandlestick() (common.Candlestick, error) {
	tick, err := i.NextTick()
	if err != nil {
		return common.Candlestick{}, err
	}
	return common.Candlestick{OpenPrice: tick.Value, HighestPrice: tick.Value, LowestPrice: tick.Value, ClosePrice: tick.Value}, nil
}

func (i *testIterator) IsOutOfTicks() bool {
	return len(i.ticks) == 0
}

func envOrStr(env, or string) string {
	s := os.Getenv(env)
	if s == "" {
		return or
	}
	return s
}
