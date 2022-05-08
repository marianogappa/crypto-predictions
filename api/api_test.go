package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"testing"
	"time"

	"github.com/marianogappa/predictions/compiler"
	"github.com/marianogappa/predictions/daemon"
	"github.com/marianogappa/predictions/metadatafetcher"
	fetcherTypes "github.com/marianogappa/predictions/metadatafetcher/types"
	"github.com/marianogappa/predictions/statestorage"
	"github.com/marianogappa/predictions/types"
	"github.com/stretchr/testify/require"
)

func TestAPI(t *testing.T) {
	tss := []apiTest{
		{
			name: "get base case: get all predictions, when there's nothing",
			test: func(t *testing.T, url string, store statestorage.StateStorage, market *testMarket, daemon *daemon.Daemon, mFetcher *metadatafetcher.MetadataFetcher) {
				apiResp, err := makeGetRequest(map[string][]string{}, url)
				require.Nil(t, err)
				require.Equal(t, 200, apiResp.Status)
				require.Len(t, apiResp.Data.Predictions, 0)
			},
		},
		{
			name: "new base case: invalid json",
			test: func(t *testing.T, url string, store statestorage.StateStorage, market *testMarket, daemon *daemon.Daemon, mFetcher *metadatafetcher.MetadataFetcher) {
				apiResp, err := makeNewRequest(`invalid`, url)
				require.Nil(t, err)
				require.Equal(t, 400, apiResp.Status)
				require.Equal(t, ErrInvalidRequestJSON.Error(), apiResp.ErrorMessage)
			},
		},
		{
			name: "new happy case",
			test: func(t *testing.T, url string, store statestorage.StateStorage, market *testMarket, daemon *daemon.Daemon, mFetcher *metadatafetcher.MetadataFetcher) {
				apiResp, err := makeNewRequest(`
					{
						"prediction": "{\"reporter\": \"admin\", \"postUrl\": \"https://twitter.com/CryptoCapo_/status/1499475622988595206\", \"given\": {\"a\": {\"condition\": \"COIN:BINANCE:BTC-USDT <= 30000\", \"toDuration\": \"3m\", \"errorMarginRatio\": 0.03 } }, \"predict\": {\"predict\": \"a\"} }",
						"store": false
					}
				`, url)
				requireEquals(t, err, nil)
				requireEquals(t, apiResp.Status, 200)
				pred := apiResp.Data.Prediction
				predBs, _ := json.Marshal(pred)
				p, _, err := compiler.NewPredictionCompiler(nil, time.Now).Compile(predBs)
				requireEquals(t, err, nil)
				requireEquals(t, p.Reporter, "admin")
				requireEquals(t, p.PostUrl, "https://twitter.com/CryptoCapo_/status/1499475622988595206")
				requireEquals(t, len(p.Given), 1)
				requireEquals(t, p.Given["a"].Operands[0].Str, "COIN:BINANCE:BTC-USDT")
				requireEquals(t, p.Given["a"].Operands[1].Str, "30000")
				requireEquals(t, p.Given["a"].Operator, "<=")
				requireEquals(t, p.Given["a"].ErrorMarginRatio, 0.03)
				requireEquals(t, p.Given["a"], p.Predict.Predict.Literal)
			},
		},
	}

	for _, ts := range tss {
		t.Run(ts.name, func(t *testing.T) {
			memoryStateStorage := statestorage.NewMemoryStateStorage()
			testMarket := newTestMarket(nil)
			mFetcher := metadatafetcher.NewMetadataFetcher()
			mFetcher.Fetchers = []metadatafetcher.SpecificFetcher{
				testFetcher{isCorrectFetcher: true, postMetadata: fetcherTypes.PostMetadata{
					Author:        types.Account{Handle: "test author"},
					PostCreatedAt: tpToISO("2022-01-02 00:00:00"),
				}, err: nil},
			}
			a := NewAPI(testMarket, memoryStateStorage, *mFetcher)
			l, err := a.Listen("localhost:0")
			if err != nil {
				t.Fatalf(err.Error())
			}
			url := l.Addr().String()
			go a.BlockinglyServe(l)
			daemon := daemon.NewDaemon(testMarket, memoryStateStorage)

			ts.test(t, url, memoryStateStorage, testMarket, daemon, mFetcher)
		})
	}
}

func tpToISO(s string) types.ISO8601 {
	t, _ := time.Parse("2006-01-02 15:04:05", s)
	return types.ISO8601(t.Format(time.RFC3339))
}

type apiTest struct {
	name string
	test func(t *testing.T, url string, store statestorage.StateStorage, market *testMarket, daemon *daemon.Daemon, mFetcher *metadatafetcher.MetadataFetcher)
}

func makeGetRequest(query map[string][]string, url string) (apiResponse[apiResGetPredictions], error) {
	req, _ := http.NewRequest("GET", fmt.Sprintf("http://%v/predictions", url), nil)
	client := &http.Client{Timeout: 10 * time.Second}

	if len(query) > 0 {
		values := req.URL.Query()
		for k, vs := range query {
			for _, v := range vs {
				values.Add(k, v)
			}
		}
		req.URL.RawQuery = values.Encode()
	}

	resp, err := client.Do(req)
	if err != nil {
		return apiResponse[apiResGetPredictions]{}, err
	}
	byts, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return apiResponse[apiResGetPredictions]{}, err
	}

	var res apiResponse[apiResGetPredictions]
	err = json.Unmarshal(byts, &res)
	if err != nil {
		return apiResponse[apiResGetPredictions]{}, err
	}

	return res, nil
}

func makeNewRequest(reqBody string, url string) (apiResponse[apiResPostPrediction], error) {
	req, _ := http.NewRequest("POST", fmt.Sprintf("http://%v/predictions", url), bytes.NewBuffer([]byte(reqBody)))
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return apiResponse[apiResPostPrediction]{}, err
	}
	byts, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return apiResponse[apiResPostPrediction]{}, err
	}

	var res apiResponse[apiResPostPrediction]
	err = json.Unmarshal(byts, &res)
	if err != nil {
		return apiResponse[apiResPostPrediction]{}, err
	}

	return res, nil
}

func requireEquals(t *testing.T, actual, expected interface{}) {
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("expected %+v but got %+v", expected, actual)
		t.FailNow()
	}
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
	ticks map[string][]types.Tick
}

func newTestMarket(ticks map[string]types.Tick) *testMarket {
	return &testMarket{}
}

func (m *testMarket) GetIterator(operand types.Operand, initialISO8601 types.ISO8601, startFromNext bool) (types.Iterator, error) {
	if _, ok := m.ticks[operand.Str]; !ok {
		return nil, types.ErrInvalidMarketPair
	}
	return newTestIterator(m.ticks[operand.Str]), nil
}

type testIterator struct {
	ticks []types.Tick
}

func newTestIterator(ticks []types.Tick) types.Iterator {
	return &testIterator{ticks}
}

func (i *testIterator) NextTick() (types.Tick, error) {
	if len(i.ticks) > 0 {
		tick := i.ticks[0]
		i.ticks = i.ticks[1:]
		return tick, nil
	}
	return types.Tick{}, types.ErrOutOfTicks
}

func (i *testIterator) NextCandlestick() (types.Candlestick, error) {
	tick, err := i.NextTick()
	if err != nil {
		return types.Candlestick{}, err
	}
	return types.Candlestick{OpenPrice: tick.Value, HighestPrice: tick.Value, LowestPrice: tick.Value, ClosePrice: tick.Value}, nil
}

func (i *testIterator) IsOutOfTicks() bool {
	return len(i.ticks) == 0
}
