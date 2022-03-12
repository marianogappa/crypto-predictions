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
	"github.com/marianogappa/predictions/metadatafetcher"
	fetcherTypes "github.com/marianogappa/predictions/metadatafetcher/types"
	"github.com/marianogappa/predictions/smrunner"
	"github.com/marianogappa/predictions/statestorage"
	"github.com/marianogappa/predictions/types"
	"github.com/marianogappa/signal-checker/common"
)

func TestAPI(t *testing.T) {
	tss := []apiTest{
		{
			name: "get base case: get all predictions, when there's nothing",
			test: func(t *testing.T, url string, store statestorage.StateStorage, market *testMarket, daemon *smrunner.SMRunner, mFetcher *metadatafetcher.MetadataFetcher) {
				apiResp, err := makeGetRequest(`{}`, url)
				requireEquals(t, err, nil)
				requireEquals(t, apiResp.Status, 200)
				requireEquals(t, len(*apiResp.Predictions), 0)
			},
		},
		{
			name: "get base case: send an invalid json",
			test: func(t *testing.T, url string, store statestorage.StateStorage, market *testMarket, daemon *smrunner.SMRunner, mFetcher *metadatafetcher.MetadataFetcher) {
				apiResp, err := makeGetRequest(`invalid`, url)
				requireEquals(t, err, nil)
				requireEquals(t, apiResp.Status, 400)
				requireEquals(t, apiResp.Message, ErrInvalidRequestJSON.Error())
			},
		},
		{
			name: "new base case: invalid json",
			test: func(t *testing.T, url string, store statestorage.StateStorage, market *testMarket, daemon *smrunner.SMRunner, mFetcher *metadatafetcher.MetadataFetcher) {
				apiResp, err := makeNewRequest(`invalid`, url)
				requireEquals(t, err, nil)
				requireEquals(t, apiResp.Status, 400)
				requireEquals(t, apiResp.Message, ErrInvalidRequestJSON.Error())
			},
		},
		{
			name: "new happy case",
			test: func(t *testing.T, url string, store statestorage.StateStorage, market *testMarket, daemon *smrunner.SMRunner, mFetcher *metadatafetcher.MetadataFetcher) {
				apiResp, err := makeNewRequest(`
					{
						"prediction": {
							"reporter": "admin",
							"postUrl": "https://twitter.com/CryptoCapo_/status/1499475622988595206",
							"given": {
								"a": {
									"condition": "COIN:BINANCE:BTC-USDT <= 30000",
									"toDuration": "3m",
									"errorMarginRatio": 0.03
								}
							},
							"predict": {
								"predict": "a"
							}
						},
						"store": false
					}
				`, url)
				requireEquals(t, err, nil)
				requireEquals(t, apiResp.Status, 200)
				p, err := compiler.NewPredictionCompiler(nil, time.Now).Compile(*apiResp.Prediction)
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
					Author:        "test author",
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
			daemon := smrunner.NewSMRunner(testMarket, memoryStateStorage)

			ts.test(t, url, memoryStateStorage, testMarket, daemon, mFetcher)
		})
	}
}

func tpToISO(s string) common.ISO8601 {
	t, _ := time.Parse("2006-01-02 15:04:05", s)
	return common.ISO8601(t.Format(time.RFC3339))
}

type apiTest struct {
	name string
	test func(t *testing.T, url string, store statestorage.StateStorage, market *testMarket, daemon *smrunner.SMRunner, mFetcher *metadatafetcher.MetadataFetcher)
}

func makeGetRequest(reqBody string, url string) (APIResponse, error) {
	req, _ := http.NewRequest("POST", fmt.Sprintf("http://%v/get", url), bytes.NewBuffer([]byte(reqBody)))
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return APIResponse{}, err
	}
	byts, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return APIResponse{}, err
	}

	var res APIResponse
	err = json.Unmarshal(byts, &res)
	if err != nil {
		return APIResponse{}, err
	}

	return res, nil
}

func makeNewRequest(reqBody string, url string) (APIResponse, error) {
	req, _ := http.NewRequest("POST", fmt.Sprintf("http://%v/new", url), bytes.NewBuffer([]byte(reqBody)))
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return APIResponse{}, err
	}
	byts, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return APIResponse{}, err
	}

	var res APIResponse
	err = json.Unmarshal(byts, &res)
	if err != nil {
		return APIResponse{}, err
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

func (m *testMarket) GetTickIterator(operand types.Operand, initialISO8601 common.ISO8601) (types.TickIterator, error) {
	if _, ok := m.ticks[operand.Str]; !ok {
		return nil, common.ErrInvalidMarketPair
	}
	return newTestTickIterator(m.ticks[operand.Str]), nil
}

type testTickIterator struct {
	ticks []types.Tick
}

func newTestTickIterator(ticks []types.Tick) *testTickIterator {
	return &testTickIterator{ticks}
}

func (i *testTickIterator) Next() (types.Tick, error) {
	if len(i.ticks) > 0 {
		tick := i.ticks[0]
		i.ticks = i.ticks[1:]
		return tick, nil
	}
	return types.Tick{}, types.ErrOutOfTicks
}

func (i *testTickIterator) IsOutOfTicks() bool {
	return len(i.ticks) == 0
}
