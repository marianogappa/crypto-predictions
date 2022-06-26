package statestorage

import (
	"database/sql"
	"fmt"
	"os"
	"os/user"
	"strings"
	"testing"

	"github.com/rs/zerolog/log"
)

var (
	store                 StateStorage
	mainTestStore         = "test_predictions"
	stateStorageTestStore = "test_predictions_statestorage"
)

func setupTestDB(t *testing.T) StateStorage {
	if store == nil {
		mainStore := connectToTestStore(t, mainTestStore)
		createDatabase(t, mainStore.DB(), stateStorageTestStore)
		store = connectToTestStore(t, stateStorageTestStore)
	}
	mustTruncateTables(t, store.DB())
	return store
}

func connectToTestStore(t *testing.T, databaseName string) StateStorage {
	if !strings.Contains(databaseName, "test_") {
		log.Error().Msgf("I'm not gonna let you connect to a non-test database!")
		return nil
	}
	osCurUser, err := user.Current()
	if err != nil {
		log.Error().Err(err).Msgf("Failed to get current user (ignoring).")
		return nil
	}
	postgresConf := PostgresConf{User: osCurUser.Username, Pass: "", Port: "5432", Database: osCurUser.Username, Host: "localhost"}
	postgresConf.User = envOrStr("PREDICTIONS_POSTGRES_USER", postgresConf.User)
	postgresConf.Pass = envOrStr("PREDICTIONS_POSTGRES_PASS", postgresConf.Pass)
	postgresConf.Port = envOrStr("PREDICTIONS_POSTGRES_PORT", postgresConf.Port)
	postgresConf.Database = databaseName
	postgresConf.Host = envOrStr("PREDICTIONS_POSTGRES_HOST", postgresConf.Host)

	_store, err := NewPostgresDBStateStorage(postgresConf)
	if err != nil {
		t.Fatal(err)
		return nil
	}
	return castStore(&_store)
}

func castStore(s *PostgresDBStateStorage) StateStorage {
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

func envOrStr(env, or string) string {
	s := os.Getenv(env)
	if s == "" {
		return or
	}
	return s
}
