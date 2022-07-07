package statestorage

import (
	"net/url"
	"testing"
	"time"

	pq "github.com/lib/pq"
	"github.com/marianogappa/predictions/compiler"
	"github.com/marianogappa/predictions/core"
	"github.com/marianogappa/predictions/metadatafetcher"
	fetcherTypes "github.com/marianogappa/predictions/metadatafetcher/types"
	"github.com/stretchr/testify/require"
)

func TestPostgres(t *testing.T) {
	tss := []storeTest{
		{
			name: "prediction upsert: base case",
			test: func(t *testing.T, store StateStorage) {
				prediction, _ := compile(t, sampleRawPrediction)
				_, err := store.UpsertPredictions([]*core.Prediction{&prediction})
				require.Nil(t, err)

				actualPreds, err := store.GetPredictions(core.APIFilters{UUIDs: []string{prediction.UUID}}, []string{}, 0, 0)
				require.Nil(t, err)
				require.Len(t, actualPreds, 1)
				require.Equal(t, prediction.PostURL, actualPreds[0].PostURL)
			},
		},
		{
			name: "prediction upsert: two",
			test: func(t *testing.T, store StateStorage) {
				prediction1, _ := compile(t, sampleRawPrediction)
				prediction2, _ := compile(t, sampleRawPrediction)
				prediction2.PostURL = "http://different.url"

				_, err := store.UpsertPredictions([]*core.Prediction{&prediction1, &prediction2})
				require.Nil(t, err)

				actualPreds, err := store.GetPredictions(core.APIFilters{}, []string{}, 0, 0)
				require.Nil(t, err)
				require.Len(t, actualPreds, 2)
				require.Equal(t, prediction1.PostURL, actualPreds[0].PostURL)
				require.Equal(t, prediction2.PostURL, actualPreds[1].PostURL)
			},
		},
		{
			name: "prediction upsert: two with same url fails",
			test: func(t *testing.T, store StateStorage) {
				prediction, _ := compile(t, sampleRawPrediction)

				_, err := store.UpsertPredictions([]*core.Prediction{&prediction, &prediction})
				require.NotNil(t, err)
				postgresErr := err.(*pq.Error)
				require.Equal(t, "ON CONFLICT DO UPDATE command cannot affect row a second time", postgresErr.Message)
			},
		},
		{
			name: "prediction hide",
			test: func(t *testing.T, store StateStorage) {
				prediction, _ := compile(t, sampleRawPrediction)
				_, err := store.UpsertPredictions([]*core.Prediction{&prediction})
				require.Nil(t, err)

				err = store.HidePrediction(prediction.UUID)
				require.Nil(t, err)

				actualPreds, err := store.GetPredictions(core.APIFilters{Hidden: pBool(true)}, []string{}, 0, 0)
				require.Nil(t, err)
				require.Len(t, actualPreds, 1)
				require.Equal(t, prediction.PostURL, actualPreds[0].PostURL)

				actualPreds, err = store.GetPredictions(core.APIFilters{Hidden: pBool(false)}, []string{}, 0, 0)
				require.Nil(t, err)
				require.Len(t, actualPreds, 0)
			},
		},
		{
			name: "prediction unhide",
			test: func(t *testing.T, store StateStorage) {
				prediction, _ := compile(t, sampleRawPrediction)
				_, err := store.UpsertPredictions([]*core.Prediction{&prediction})
				require.Nil(t, err)

				err = store.HidePrediction(prediction.UUID)
				require.Nil(t, err)

				err = store.UnhidePrediction(prediction.UUID)
				require.Nil(t, err)

				actualPreds, err := store.GetPredictions(core.APIFilters{Hidden: pBool(false)}, []string{}, 0, 0)
				require.Nil(t, err)
				require.Len(t, actualPreds, 1)
				require.Equal(t, prediction.PostURL, actualPreds[0].PostURL)

				actualPreds, err = store.GetPredictions(core.APIFilters{Hidden: pBool(true)}, []string{}, 0, 0)
				require.Nil(t, err)
				require.Len(t, actualPreds, 0)
			},
		},
		{
			name: "prediction delete",
			test: func(t *testing.T, store StateStorage) {
				prediction, _ := compile(t, sampleRawPrediction)
				_, err := store.UpsertPredictions([]*core.Prediction{&prediction})
				require.Nil(t, err)

				err = store.DeletePrediction(prediction.UUID)
				require.Nil(t, err)

				actualPreds, err := store.GetPredictions(core.APIFilters{Deleted: pBool(true)}, []string{}, 0, 0)
				require.Nil(t, err)
				require.Len(t, actualPreds, 1)
				require.Equal(t, prediction.PostURL, actualPreds[0].PostURL)

				actualPreds, err = store.GetPredictions(core.APIFilters{Deleted: pBool(false)}, []string{}, 0, 0)
				require.Nil(t, err)
				require.Len(t, actualPreds, 0)
			},
		},
		{
			name: "prediction undelete",
			test: func(t *testing.T, store StateStorage) {
				prediction, _ := compile(t, sampleRawPrediction)
				_, err := store.UpsertPredictions([]*core.Prediction{&prediction})
				require.Nil(t, err)

				err = store.DeletePrediction(prediction.UUID)
				require.Nil(t, err)

				err = store.UndeletePrediction(prediction.UUID)
				require.Nil(t, err)

				actualPreds, err := store.GetPredictions(core.APIFilters{Deleted: pBool(false)}, []string{}, 0, 0)
				require.Nil(t, err)
				require.Len(t, actualPreds, 1)
				require.Equal(t, prediction.PostURL, actualPreds[0].PostURL)

				actualPreds, err = store.GetPredictions(core.APIFilters{Deleted: pBool(true)}, []string{}, 0, 0)
				require.Nil(t, err)
				require.Len(t, actualPreds, 0)
			},
		},
		{
			name: "prediction pause",
			test: func(t *testing.T, store StateStorage) {
				prediction, _ := compile(t, sampleRawPrediction)
				_, err := store.UpsertPredictions([]*core.Prediction{&prediction})
				require.Nil(t, err)

				err = store.PausePrediction(prediction.UUID)
				require.Nil(t, err)

				actualPreds, err := store.GetPredictions(core.APIFilters{Paused: pBool(true)}, []string{}, 0, 0)
				require.Nil(t, err)
				require.Len(t, actualPreds, 1)
				require.Equal(t, prediction.PostURL, actualPreds[0].PostURL)

				actualPreds, err = store.GetPredictions(core.APIFilters{Paused: pBool(false)}, []string{}, 0, 0)
				require.Nil(t, err)
				require.Len(t, actualPreds, 0)
			},
		},
		{
			name: "prediction unpause",
			test: func(t *testing.T, store StateStorage) {
				prediction, _ := compile(t, sampleRawPrediction)
				_, err := store.UpsertPredictions([]*core.Prediction{&prediction})
				require.Nil(t, err)

				err = store.PausePrediction(prediction.UUID)
				require.Nil(t, err)

				err = store.UnpausePrediction(prediction.UUID)
				require.Nil(t, err)

				actualPreds, err := store.GetPredictions(core.APIFilters{Paused: pBool(false)}, []string{}, 0, 0)
				require.Nil(t, err)
				require.Len(t, actualPreds, 1)
				require.Equal(t, prediction.PostURL, actualPreds[0].PostURL)

				actualPreds, err = store.GetPredictions(core.APIFilters{Paused: pBool(true)}, []string{}, 0, 0)
				require.Nil(t, err)
				require.Len(t, actualPreds, 0)
			},
		},
		{
			name: "account upsert: base case",
			test: func(t *testing.T, store StateStorage) {
				_, account := compile(t, sampleRawPrediction)
				_, err := store.UpsertAccounts([]*core.Account{account})
				require.Nil(t, err)

				actualAccounts, err := store.GetAccounts(core.APIAccountFilters{URLs: []string{account.URL.String()}}, []string{}, 0, 0)
				require.Nil(t, err)
				require.Len(t, actualAccounts, 1)
				require.Equal(t, account.Handle, actualAccounts[0].Handle)
			},
		},
		{
			name: "account upsert: two",
			test: func(t *testing.T, store StateStorage) {
				_, account1 := compile(t, sampleRawPrediction)
				_, account2 := compile(t, sampleRawPrediction)
				account2.URL, _ = url.Parse("http://twitter.com/different")
				account2.Handle = "different"
				_, err := store.UpsertAccounts([]*core.Account{account1, account2})
				require.Nil(t, err)

				actualAccounts, err := store.GetAccounts(core.APIAccountFilters{}, []string{}, 0, 0)
				require.Nil(t, err)
				require.Len(t, actualAccounts, 2)
				require.Equal(t, account1.Handle, actualAccounts[0].Handle)
				require.Equal(t, account2.Handle, actualAccounts[1].Handle)
			},
		},
		{
			name: "account upsert: two with same URL fails",
			test: func(t *testing.T, store StateStorage) {
				_, account1 := compile(t, sampleRawPrediction)
				_, account2 := compile(t, sampleRawPrediction)
				_, err := store.UpsertAccounts([]*core.Account{account1, account2})
				require.NotNil(t, err)
				postgresErr := err.(*pq.Error)
				require.Equal(t, "ON CONFLICT DO UPDATE command cannot affect row a second time", postgresErr.Message)
			},
		},
	}

	for _, ts := range tss {
		t.Run(ts.name, func(t *testing.T) {
			var store = setupTestDB(t)
			ts.test(t, store)
		})
	}
}

func tpToISO(s string) core.ISO8601 {
	t, _ := time.Parse("2006-01-02 15:04:05", s)
	return core.ISO8601(t.Format(time.RFC3339))
}

type storeTest struct {
	name string
	test func(t *testing.T, store StateStorage)
}

var (
	sampleRawPrediction = []byte(`{
		"reporter": "admin",
		"postUrl": "https://twitter.com/CryptoCapo_/status/1499475622988595206",
		"given":
		{
			"a":
			{
				"condition": "COIN:BINANCE:BTC-USDT <= 30000",
				"toDuration": "1d",
				"errorMarginRatio": 0.03
			}
		},
		"predict":
		{
			"predict": "a"
		}
	}`)
)

func addTestFetcher(mf *metadatafetcher.MetadataFetcher) {
	postAuthorURL, _ := url.Parse("https://twitter.com/CryptoCapo_")
	mf.Fetchers = []metadatafetcher.SpecificFetcher{
		testFetcher{isCorrectFetcher: true, postMetadata: fetcherTypes.PostMetadata{
			Author:        core.Account{Handle: "test author", URL: postAuthorURL},
			PostCreatedAt: tpToISO("2022-01-02 00:00:00"),
		}, err: nil},
	}
}

func compile(t *testing.T, rawPrediction []byte) (core.Prediction, *core.Account) {
	mf := metadatafetcher.NewMetadataFetcher()
	addTestFetcher(mf)
	compiledPrediction, account, err := compiler.NewPredictionCompiler(mf, time.Now).Compile(sampleRawPrediction)
	require.Nil(t, err)
	return compiledPrediction, account
}
