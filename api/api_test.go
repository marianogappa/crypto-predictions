package api

import (
	"encoding/json"
	"fmt"
	"net/url"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/marianogappa/predictions/compiler"
	"github.com/marianogappa/predictions/daemon"
	"github.com/marianogappa/predictions/imagebuilder"
	"github.com/marianogappa/predictions/metadatafetcher"
	fetcherTypes "github.com/marianogappa/predictions/metadatafetcher/types"
	"github.com/marianogappa/predictions/serializer"
	"github.com/marianogappa/predictions/statestorage"
	"github.com/marianogappa/predictions/types"
	"github.com/stretchr/testify/require"
)

func TestAPI(t *testing.T) {
	tss := []apiTest{
		{
			name: "get base case: get all predictions, when there's nothing",
			test: func(t *testing.T, a *API, ctx testContext) {
				apiResp := a.getPredictions(apiReqGetPredictions{})

				require.Equal(t, 200, apiResp.Status)
				require.Equal(t, "", apiResp.ErrorCode)
				require.Len(t, apiResp.Data.Predictions, 0)
			},
		},
		{
			name: "new base case: invalid json",
			test: func(t *testing.T, a *API, ctx testContext) {
				apiResp := a.postPrediction(apiReqPostPrediction{Prediction: "invalid"})

				require.Equal(t, 400, apiResp.Status)
				require.Equal(t, "ErrInvalidRequestJSON", apiResp.ErrorCode)
				require.Equal(t, ErrInvalidRequestJSON.Error(), apiResp.ErrorMessage)
			},
		},
		{
			name: "new happy case",
			test: func(t *testing.T, a *API, ctx testContext) {
				rawPrediction := `
					{
						"reporter": "admin",
						"postUrl": "https://twitter.com/CryptoCapo_/status/1499475622988595206",
						"given":
						{
							"a":
							{
								"condition": "COIN:BINANCE:BTC-USDT <= 30000",
								"toDuration": "3m",
								"errorMarginRatio": 0.03
							}
						},
						"predict":
						{
							"predict": "a"
						}
					}
				`
				apiResp := a.postPrediction(apiReqPostPrediction{Prediction: rawPrediction})

				require.Equal(t, apiResp.Status, 200)
				pred := apiResp.Data.Prediction
				predBs, _ := json.Marshal(pred)
				p, _, err := compiler.NewPredictionCompiler(nil, time.Now).Compile(predBs)
				require.Equal(t, err, nil)
				require.Equal(t, p.Reporter, "admin")
				require.Equal(t, p.PostURL, "https://twitter.com/CryptoCapo_/status/1499475622988595206")
				require.Equal(t, len(p.Given), 1)
				require.Equal(t, p.Given["a"].Operands[0].Str, "COIN:BINANCE:BTC-USDT")
				require.Equal(t, p.Given["a"].Operands[1].Str, "30000")
				require.Equal(t, p.Given["a"].Operator, "<=")
				require.Equal(t, p.Given["a"].ErrorMarginRatio, 0.03)
				require.Equal(t, p.Given["a"], p.Predict.Predict.Literal)
			},
		},
		{
			name: "hide prediction",
			test: func(t *testing.T, a *API, ctx testContext) {
				apiResp := a.postPrediction(apiReqPostPrediction{Prediction: string(sampleRawPrediction), Store: true})
				require.Equal(t, apiResp.Status, 200, apiResp.InternalErrorMessage)
				require.Equal(t, apiResp.Data.Stored, true)

				uuid := apiResp.Data.Prediction.UUID

				hideResp := a.predictionStorageActionWithUUID(uuid, ctx.store.HidePrediction)
				require.Equal(t, hideResp.Status, 200, hideResp.InternalErrorMessage)
				require.Equal(t, hideResp.Data.Stored, true)

				getResp := a.getPredictions(apiReqGetPredictions{UUIDs: []string{uuid}, Hidden: pBool(true)})
				require.Equal(t, getResp.Status, 200, getResp.InternalErrorMessage)
				require.Len(t, getResp.Data.Predictions, 1)

				// When filtering by not hidden, it shouldn't show up
				getResp2 := a.getPredictions(apiReqGetPredictions{UUIDs: []string{uuid}, Hidden: pBool(false)})
				require.Equal(t, getResp2.Status, 200, getResp2.InternalErrorMessage)
				require.Len(t, getResp2.Data.Predictions, 0)

				// When posting a new prediction and not hiding it, only the hidden one should show up
				secondPrediction, _ := compile(t, sampleRawPrediction)
				secondPrediction.PostURL = "https://twitter.com/differentUser/status/1499475622988595206"
				apiResp = a.postPrediction(apiReqPostPrediction{Prediction: serialize(t, secondPrediction), Store: true})
				require.Equal(t, apiResp.Status, 200, apiResp.InternalErrorMessage)

				getResp = a.getPredictions(apiReqGetPredictions{Hidden: pBool(true)})
				require.Equal(t, getResp.Status, 200, getResp.InternalErrorMessage)
				require.Len(t, getResp.Data.Predictions, 1)
				require.Equal(t, getResp.Data.Predictions[0].UUID, uuid)
			},
		},
		{
			name: "delete prediction",
			test: func(t *testing.T, a *API, ctx testContext) {
				apiResp := a.postPrediction(apiReqPostPrediction{Prediction: string(sampleRawPrediction), Store: true})
				require.Equal(t, apiResp.Status, 200, apiResp.InternalErrorMessage)
				require.Equal(t, apiResp.Data.Stored, true)

				uuid := apiResp.Data.Prediction.UUID

				deleteResp := a.predictionStorageActionWithUUID(uuid, ctx.store.DeletePrediction)
				require.Equal(t, deleteResp.Status, 200, deleteResp.InternalErrorMessage)
				require.Equal(t, deleteResp.Data.Stored, true)

				getResp := a.getPredictions(apiReqGetPredictions{UUIDs: []string{uuid}, Deleted: pBool(true)})
				require.Equal(t, getResp.Status, 200, getResp.InternalErrorMessage)
				require.Len(t, getResp.Data.Predictions, 1)

				// When filtering by not deleted, it shouldn't show up
				getResp2 := a.getPredictions(apiReqGetPredictions{UUIDs: []string{uuid}, Deleted: pBool(false)})
				require.Equal(t, getResp2.Status, 200, getResp2.InternalErrorMessage)
				require.Len(t, getResp2.Data.Predictions, 0)

				// When posting a new prediction and not deleting it, only the deleted one should show up
				secondPrediction, _ := compile(t, sampleRawPrediction)
				secondPrediction.PostURL = "https://twitter.com/differentUser/status/1499475622988595206"
				apiResp = a.postPrediction(apiReqPostPrediction{Prediction: serialize(t, secondPrediction), Store: true})
				require.Equal(t, apiResp.Status, 200, apiResp.InternalErrorMessage)

				getResp = a.getPredictions(apiReqGetPredictions{Deleted: pBool(true)})
				require.Equal(t, getResp.Status, 200, getResp.InternalErrorMessage)
				require.Len(t, getResp.Data.Predictions, 1)
				require.Equal(t, getResp.Data.Predictions[0].UUID, uuid)
			},
		},
		{
			name: "pause prediction",
			test: func(t *testing.T, a *API, ctx testContext) {
				apiResp := a.postPrediction(apiReqPostPrediction{Prediction: string(sampleRawPrediction), Store: true})
				require.Equal(t, apiResp.Status, 200, apiResp.InternalErrorMessage)
				require.Equal(t, apiResp.Data.Stored, true)

				uuid := apiResp.Data.Prediction.UUID

				pauseResp := a.predictionStorageActionWithUUID(uuid, ctx.store.PausePrediction)
				require.Equal(t, pauseResp.Status, 200, pauseResp.InternalErrorMessage)
				require.Equal(t, pauseResp.Data.Stored, true)

				getResp := a.getPredictions(apiReqGetPredictions{UUIDs: []string{uuid}, Paused: pBool(true)})
				require.Equal(t, getResp.Status, 200, getResp.InternalErrorMessage)
				require.Len(t, getResp.Data.Predictions, 1)

				// When filtering by not paused, it shouldn't show up
				getResp2 := a.getPredictions(apiReqGetPredictions{UUIDs: []string{uuid}, Paused: pBool(false)})
				require.Equal(t, getResp2.Status, 200, getResp2.InternalErrorMessage)
				require.Len(t, getResp2.Data.Predictions, 0)

				// When posting a new prediction and not pausing it, only the paused one should show up
				secondPrediction, _ := compile(t, sampleRawPrediction)
				secondPrediction.PostURL = "https://twitter.com/differentUser/status/1499475622988595206"
				apiResp = a.postPrediction(apiReqPostPrediction{Prediction: serialize(t, secondPrediction), Store: true})
				require.Equal(t, apiResp.Status, 200, apiResp.InternalErrorMessage)

				getResp = a.getPredictions(apiReqGetPredictions{Paused: pBool(true)})
				require.Equal(t, getResp.Status, 200, getResp.InternalErrorMessage)
				require.Len(t, getResp.Data.Predictions, 1)
				require.Equal(t, getResp.Data.Predictions[0].UUID, uuid)
			},
		},
		{
			name: "unhide prediction",
			test: func(t *testing.T, a *API, ctx testContext) {
				// Create sample prediction
				apiResp := a.postPrediction(apiReqPostPrediction{Prediction: string(sampleRawPrediction), Store: true})
				uuid := apiResp.Data.Prediction.UUID

				// Hide it
				a.predictionStorageActionWithUUID(uuid, ctx.store.HidePrediction)

				// Unhide it
				a.predictionStorageActionWithUUID(uuid, ctx.store.UnhidePrediction)

				// When filtering by not hidden, it should show up
				getResp := a.getPredictions(apiReqGetPredictions{UUIDs: []string{uuid}, Hidden: pBool(false)})
				require.Len(t, getResp.Data.Predictions, 1)

				// When filtering by hidden, it shouldn't show up
				getResp2 := a.getPredictions(apiReqGetPredictions{UUIDs: []string{uuid}, Hidden: pBool(true)})
				require.Len(t, getResp2.Data.Predictions, 0)

				// When posting a new prediction and hiding it, only the hidden one should show up
				secondPrediction, _ := compile(t, sampleRawPrediction)
				secondPrediction.PostURL = "https://twitter.com/differentUser/status/1499475622988595206"
				apiResp = a.postPrediction(apiReqPostPrediction{Prediction: serialize(t, secondPrediction), Store: true})
				secondUUID := apiResp.Data.Prediction.UUID

				a.predictionStorageActionWithUUID(secondUUID, ctx.store.HidePrediction)

				getResp = a.getPredictions(apiReqGetPredictions{Hidden: pBool(true)})
				require.Len(t, getResp.Data.Predictions, 1)
				require.Equal(t, secondUUID, getResp.Data.Predictions[0].UUID)
			},
		},
		{
			name: "undelete prediction",
			test: func(t *testing.T, a *API, ctx testContext) {
				// Create sample prediction
				apiResp := a.postPrediction(apiReqPostPrediction{Prediction: string(sampleRawPrediction), Store: true})
				uuid := apiResp.Data.Prediction.UUID

				// Delete it
				a.predictionStorageActionWithUUID(uuid, ctx.store.DeletePrediction)

				// Undelete it
				a.predictionStorageActionWithUUID(uuid, ctx.store.UndeletePrediction)

				// When filtering by not deleted, it should show up
				getResp := a.getPredictions(apiReqGetPredictions{UUIDs: []string{uuid}, Deleted: pBool(false)})
				require.Len(t, getResp.Data.Predictions, 1)

				// When filtering by deleted, it shouldn't show up
				getResp2 := a.getPredictions(apiReqGetPredictions{UUIDs: []string{uuid}, Deleted: pBool(true)})
				require.Len(t, getResp2.Data.Predictions, 0)

				// When posting a new prediction and deleting it, only the deleted one should show up
				secondPrediction, _ := compile(t, sampleRawPrediction)
				secondPrediction.PostURL = "https://twitter.com/differentUser/status/1499475622988595206"
				apiResp = a.postPrediction(apiReqPostPrediction{Prediction: serialize(t, secondPrediction), Store: true})
				secondUUID := apiResp.Data.Prediction.UUID

				a.predictionStorageActionWithUUID(secondUUID, ctx.store.DeletePrediction)

				getResp = a.getPredictions(apiReqGetPredictions{Deleted: pBool(true)})
				require.Len(t, getResp.Data.Predictions, 1)
				require.Equal(t, secondUUID, getResp.Data.Predictions[0].UUID)
			},
		},
		{
			name: "unpause prediction",
			test: func(t *testing.T, a *API, ctx testContext) {
				// Create sample prediction
				apiResp := a.postPrediction(apiReqPostPrediction{Prediction: string(sampleRawPrediction), Store: true})
				uuid := apiResp.Data.Prediction.UUID

				// Pause it
				a.predictionStorageActionWithUUID(uuid, ctx.store.PausePrediction)

				// Unpause it
				a.predictionStorageActionWithUUID(uuid, ctx.store.UnpausePrediction)

				// When filtering by not paused, it should show up
				getResp := a.getPredictions(apiReqGetPredictions{UUIDs: []string{uuid}, Paused: pBool(false)})
				require.Len(t, getResp.Data.Predictions, 1)

				// When filtering by paused, it shouldn't show up
				getResp2 := a.getPredictions(apiReqGetPredictions{UUIDs: []string{uuid}, Paused: pBool(true)})
				require.Len(t, getResp2.Data.Predictions, 0)

				// When posting a new prediction and pausing it, only the paused one should show up
				secondPrediction, _ := compile(t, sampleRawPrediction)
				secondPrediction.PostURL = "https://twitter.com/differentUser/status/1499475622988595206"
				apiResp = a.postPrediction(apiReqPostPrediction{Prediction: serialize(t, secondPrediction), Store: true})
				secondUUID := apiResp.Data.Prediction.UUID

				a.predictionStorageActionWithUUID(secondUUID, ctx.store.PausePrediction)

				getResp = a.getPredictions(apiReqGetPredictions{Paused: pBool(true)})
				require.Len(t, getResp.Data.Predictions, 1)
				require.Equal(t, secondUUID, getResp.Data.Predictions[0].UUID)
			},
		},
		{
			name: "getPagesPrediction with url",
			test: func(t *testing.T, a *API, ctx testContext) {
				// Create sample predictions with different attributes
				samplePreds := []types.Prediction{}
				sampleAccounts := []*types.Account{}
				for i := 0; i < 10; i++ {
					samplePred, sampleAccount := compile(t, sampleRawPrediction)
					samplePred.UUID = uuid.NewString()
					samplePred.PostURL = fmt.Sprintf("https://twitter.com/differentUser/status/%v", i)
					samplePred.PostedAt = types.ISO8601(fmt.Sprintf("2020-06-26T00:00:0%vZ", i))
					samplePred.PostAuthor = fmt.Sprintf("User%v", i)
					samplePred.PostAuthorURL = fmt.Sprintf("https://twitter.com/user%v", i)

					sampleAccount.FollowerCount = 10 - i
					sampleAccount.Name = samplePred.PostAuthor
					sampleAccount.Handle = samplePred.PostAuthor
					sampleAccount.URL, _ = url.Parse(samplePred.PostAuthorURL)

					// The 5 first predictions have ETH instead of BTC
					if i < 5 {
						samplePred.Predict.Predict.Literal.Operands[0].BaseAsset = "ETH"
						samplePred.Predict.Predict.Literal.Operands[0].Str = "COIN:BINANCE:ETH-USDT"
					}
					// 4 predictions have same account as first
					if i >= 4 && i <= 7 {
						samplePred.PostAuthor = "User0"
						samplePred.PostAuthorURL = "https://twitter.com/user0"

						sampleAccount.FollowerCount = 10
						sampleAccount.Name = samplePred.PostAuthor
						sampleAccount.Handle = samplePred.PostAuthor
						sampleAccount.URL, _ = url.Parse(samplePred.PostAuthorURL)
					} else {
						sampleAccounts = append(sampleAccounts, sampleAccount)
					}

					samplePreds = append(samplePreds, samplePred)

					apiResp := a.postPrediction(apiReqPostPrediction{Prediction: serialize(t, samplePred), Store: true})
					require.Equal(t, 200, apiResp.Status)
				}
				_, err := ctx.store.UpsertAccounts(sampleAccounts)
				require.Nil(t, err)

				// Populate the test market with Ticks so that the summary can be created
				curTime := tp("2020-06-24T00:00:00Z")
				for i := 0; i <= 60; i++ {
					for _, coin := range []string{"COIN:BINANCE:BTC-USDT", "COIN:BINANCE:ETH-USDT"} {
						tick := types.Tick{Timestamp: int(curTime.Unix()), Value: 29000}
						ctx.market.ticks[coin] = append(ctx.market.ticks[coin], tick)
					}
					curTime.Add(1 * time.Hour)
				}

				apiResp := a.getPagesPrediction(samplePreds[0].PostURL)
				require.Equal(t, 200, apiResp.Status)

				// The main Prediction's UUID must match the first samplePred
				require.Equal(t, samplePreds[0].UUID, string(apiResp.Data.Prediction))

				// Latest10Predictions must have samplePreds in reverse order (minus main), due to increasing PostedAt
				require.Len(t, apiResp.Data.Latest10Predictions, 9)
				for i := 1; i <= 9; i++ {
					require.Equal(t, samplePreds[10-i].UUID, string(apiResp.Data.Latest10Predictions[i-1]))
				}

				// Top10AccountsByFollowerCount must have 6 sampleAccounts in ascending order due to inc. FollowerCount
				require.Len(t, apiResp.Data.Top10AccountsByFollowerCount, 6)
				for i := 0; i < 6; i++ {
					require.Equal(t, sampleAccounts[i].URL.String(), string(apiResp.Data.Top10AccountsByFollowerCount[i]))
				}

				// Latest5PredictionsSameAccount must have samplePreds 4, 5, 6 & 7 in reverse order
				require.Len(t, apiResp.Data.Latest5PredictionsSameAccount, 4)
				for dataI, sampleI := range []int{7, 6, 5, 4} {
					require.Equal(t, samplePreds[sampleI].UUID, string(apiResp.Data.Latest5PredictionsSameAccount[dataI]))
				}

				// Latest5PredictionsSameCoin must have samplePreds 1, 2, 3 & 4 in reverse order
				require.Len(t, apiResp.Data.Latest5PredictionsSameCoin, 4)
				for dataI, sampleI := range []int{4, 3, 2, 1} {
					require.Equal(t, samplePreds[sampleI].UUID, string(apiResp.Data.Latest5PredictionsSameCoin[dataI]))
				}

				// AccountsByURL must have all sampleAccounts
				for _, sampleAccount := range sampleAccounts {
					require.Equal(t, sampleAccount.Handle, apiResp.Data.AccountsByURL[URL(sampleAccount.URL.String())].Handle)
				}

				// PredictionsByUUID must have all samplePreds
				for _, samplePred := range samplePreds {
					require.Equal(t, samplePred.PostURL, apiResp.Data.PredictionsByUUID[UUID(samplePred.UUID)].PostURL)
				}
			},
		},
	}

	for _, ts := range tss {
		t.Run(ts.name, func(t *testing.T) {
			var (
				store      = setupTestDB(t)
				testMarket = newTestMarket(map[string][]types.Tick{})
				mFetcher   = metadatafetcher.NewMetadataFetcher()
				daemon     = daemon.NewDaemon(testMarket, store, imagebuilder.PredictionImageBuilder{}, false, false, "")
			)
			addTestFetcher(mFetcher)

			a := NewAPI(testMarket, store, *mFetcher, imagebuilder.PredictionImageBuilder{}, "admin", "admin")
			go a.MustBlockinglyListenAndServe("localhost:0")

			ts.test(t, a, testContext{store, testMarket, daemon, mFetcher})
		})
	}
}

func addTestFetcher(mf *metadatafetcher.MetadataFetcher) {
	postAuthorURL, _ := url.Parse("https://twitter.com/CryptoCapo_")
	mf.Fetchers = []metadatafetcher.SpecificFetcher{
		testFetcher{isCorrectFetcher: true, postMetadata: fetcherTypes.PostMetadata{
			Author:        types.Account{Handle: "test author", URL: postAuthorURL},
			PostCreatedAt: tpToISO("2022-01-02 00:00:00"),
		}, err: nil},
	}
}

func tpToISO(s string) types.ISO8601 {
	t, _ := time.Parse("2006-01-02 15:04:05", s)
	return types.ISO8601(t.Format(time.RFC3339))
}

type testContext struct {
	store    statestorage.StateStorage
	market   *testMarket
	daemon   *daemon.Daemon
	mFetcher *metadatafetcher.MetadataFetcher
}

type apiTest struct {
	name string
	test func(t *testing.T, a *API, ctx testContext)
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

func compile(t *testing.T, rawPrediction []byte) (types.Prediction, *types.Account) {
	mf := metadatafetcher.NewMetadataFetcher()
	addTestFetcher(mf)
	compiledPrediction, account, err := compiler.NewPredictionCompiler(mf, time.Now).Compile(sampleRawPrediction)
	require.Nil(t, err)
	return compiledPrediction, account
}

func serialize(t *testing.T, prediction types.Prediction) string {
	rawPrediction, err := serializer.NewPredictionSerializer(nil).Serialize(&prediction)
	require.Nil(t, err)
	return string(rawPrediction)
}

func pBool(b bool) *bool { return &b }

func tp(s string) time.Time {
	t, _ := time.Parse("2006-01-02 15:04:05", s)
	return t
}
