package coinbase

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
		`[
			[1626868320,31518.39,31556.85,31518.39,31541.36,1.51839032],
			[1626868260,31508.43,31511.41,31508.43,31511.41,0.12575184]
		]`,
		`[
			[1626868560,31540.72,31584.3,31540.72,31576.13,0.08432516],
			[1626868380,31513.7,31538.7,31538.7,31513.7,0.36927379]
		]`,
		`[]`,
	}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, replies[i%len(replies)])
		i++
	}))
	defer ts.Close()

	b := NewCoinbase()
	b.overrideAPIURL(ts.URL + "/")
	ci := b.BuildCandlestickIterator("BTC", "USDT", "2021-07-20T14:14:18+00:00")

	expectedResults := []expected{
		{
			candlestick: types.Candlestick{
				Timestamp:      1626868260,
				OpenPrice:      31508.43,
				ClosePrice:     31511.41,
				LowestPrice:    31508.43,
				HighestPrice:   31511.41,
				Volume:         0.12575184,
				NumberOfTrades: 0,
			},
			err: nil,
		},
		{
			candlestick: types.Candlestick{
				Timestamp:      1626868320,
				OpenPrice:      31518.39,
				ClosePrice:     31541.36,
				LowestPrice:    31518.39,
				HighestPrice:   31556.85,
				Volume:         1.51839032,
				NumberOfTrades: 0,
			},
			err: nil,
		},
		{
			candlestick: types.Candlestick{
				Timestamp:      1626868380,
				OpenPrice:      31538.7,
				ClosePrice:     31513.7,
				LowestPrice:    31513.7,
				HighestPrice:   31538.7,
				Volume:         0.36927379,
				NumberOfTrades: 0,
			},
			err: nil,
		},
		{
			candlestick: types.Candlestick{
				Timestamp:      1626868560,
				OpenPrice:      31540.72,
				ClosePrice:     31576.13,
				LowestPrice:    31540.72,
				HighestPrice:   31584.3,
				Volume:         0.08432516,
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
