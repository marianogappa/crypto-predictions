package ftx

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
		`{"success":true,"result":[
			{"startTime":"2021-07-20T11:06:00+00:00","time":1626779160000.0,"open":29704.0,"high":29729.0,"low":29702.0,"close":29702.0,"volume":16542.6909},
			{"startTime":"2021-07-20T11:07:00+00:00","time":1626779220000.0,"open":29702.0,"high":29704.0,"low":29691.0,"close":29694.0,"volume":528.6186}
		]}`,
		`{"success":true,"result":[
			{"startTime":"2021-07-20T11:08:00+00:00","time":1626779280000.0,"open":29695.0,"high":29695.0,"low":29663.0,"close":29667.0,"volume":11909.3972},
			{"startTime":"2021-07-20T11:09:00+00:00","time":1626779340000.0,"open":29667.0,"high":29677.0,"low":29662.0,"close":29663.0,"volume":2207.2532}
		]}`,
		`{"success":true,"result":[]}`,
	}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, replies[i%len(replies)])
		i++
	}))
	defer ts.Close()

	b := NewFTX()
	b.overrideAPIURL(ts.URL + "/")
	ci := b.BuildCandlestickIterator("BTC", "USDT", "2021-07-04T14:14:18+00:00")

	expectedResults := []expected{
		{
			candlestick: types.Candlestick{
				Timestamp:      1626779160,
				OpenPrice:      29704.0,
				ClosePrice:     29702.0,
				LowestPrice:    29702.0,
				HighestPrice:   29729.0,
				Volume:         16542.6909,
				NumberOfTrades: 0,
			},
			err: nil,
		},
		{
			candlestick: types.Candlestick{
				Timestamp:      1626779220,
				OpenPrice:      29702.0,
				ClosePrice:     29694.0,
				LowestPrice:    29691.0,
				HighestPrice:   29704.0,
				Volume:         528.6186,
				NumberOfTrades: 0,
			},
			err: nil,
		},
		{
			candlestick: types.Candlestick{
				Timestamp:      1626779280,
				OpenPrice:      29695.0,
				ClosePrice:     29667.0,
				LowestPrice:    29663.0,
				HighestPrice:   29695.0,
				Volume:         11909.3972,
				NumberOfTrades: 0,
			},
			err: nil,
		},
		{
			candlestick: types.Candlestick{
				Timestamp:      1626779340,
				OpenPrice:      29667.0,
				ClosePrice:     29663.0,
				LowestPrice:    29662.0,
				HighestPrice:   29677.0,
				Volume:         2207.2532,
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
