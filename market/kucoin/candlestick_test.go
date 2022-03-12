package kucoin

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/marianogappa/predictions/types"
)

type expected struct {
	candlestick types.Candlestick
	err         error
}

func TestCandlesticks(t *testing.T) {
	i := 0
	replies := []string{
		`{"code":"200000","data":[
			["1566789600","10408.6","10416.0","10416.0","10405.4","12.45584973","129666.51508559"],
			["1566789540","10404.7","10408.7","10415.5","10398.2","9.8995635","103014.048407233"]
		]}`,
		`{"code":"200000","data":[
			["1566789720","10411.5","10401.9","10411.5","10396.3","29.11357276","302889.301529914"],
			["1566789660","10416.0","10411.5","10422.3","10411.5","15.61781842","162703.708997029"]
		]}`,
		`{"code":"200000","data":[]}`,
	}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, replies[i%len(replies)])
		i++
	}))
	defer ts.Close()

	b := NewKucoin()
	b.overrideAPIURL(ts.URL + "/")
	ci := b.BuildCandlestickIterator("BTC", "USDT", "2019-07-04T14:14:18+00:00")

	expectedResults := []expected{
		{
			candlestick: types.Candlestick{
				Timestamp:      1566789540,
				OpenPrice:      10404.7,
				ClosePrice:     10408.7,
				LowestPrice:    10398.2,
				HighestPrice:   10415.5,
				Volume:         9.8995635,
				NumberOfTrades: 0,
			},
			err: nil,
		},
		{
			candlestick: types.Candlestick{
				Timestamp:      1566789600,
				OpenPrice:      10408.6,
				ClosePrice:     10416.0,
				LowestPrice:    10405.4,
				HighestPrice:   10416.0,
				Volume:         12.45584973,
				NumberOfTrades: 0,
			},
			err: nil,
		},
		{
			candlestick: types.Candlestick{
				Timestamp:      1566789660,
				OpenPrice:      10416.0,
				ClosePrice:     10411.5,
				LowestPrice:    10411.5,
				HighestPrice:   10422.3,
				Volume:         15.61781842,
				NumberOfTrades: 0,
			},
			err: nil,
		},
		{
			candlestick: types.Candlestick{
				Timestamp:      1566789720,
				OpenPrice:      10411.5,
				ClosePrice:     10401.9,
				LowestPrice:    10396.3,
				HighestPrice:   10411.5,
				Volume:         29.11357276,
				NumberOfTrades: 0,
			},
			err: nil,
		},
		{
			candlestick: types.Candlestick{},
			err:         types.ErrOutOfCandlesticks,
		},
	}
	for i, expectedResult := range expectedResults {
		actualCandlestick, actualErr := ci.Next()
		if actualCandlestick != expectedResult.candlestick {
			t.Errorf("on candlestick %v expected %v but got %v", i, expectedResult.candlestick, actualCandlestick)
			t.FailNow()
		}
		if actualErr != expectedResult.err {
			t.Errorf("on candlestick %v expected no errors but this error happened %v", i, actualErr)
			t.FailNow()
		}
	}
}
