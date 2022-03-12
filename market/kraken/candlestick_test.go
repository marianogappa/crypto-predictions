package kraken

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
		`{"error":[],"result":{"XBTUSDT":[
			[1626826200,"29804.5","29804.5","29804.5","29804.5","0.0","0.00000000",0],
			[1626826260,"29809.4","29809.4","29809.4","29809.4","29809.4","0.00024675",1]
		], "last":1626826260}}`,
		`{"error":[],"result":{"XBTUSDT":[
			[1626826320,"29801.5","29801.7","29801.4","29801.6","29801.5","0.10940000",4],
			[1626826380,"29818.0","29818.1","29818.0","29818.1","29818.0","0.00215415",4]
		], "last":1626826380}}`,
		`{"error":[],"result":{"XBTUSDT":[], "last":1626826380}}`,
	}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, replies[i%len(replies)])
		i++
	}))
	defer ts.Close()

	b := NewKraken()
	b.overrideAPIURL(ts.URL + "/")
	ci := b.BuildCandlestickIterator("BTC", "USDT", "2021-07-04T14:14:18+00:00")

	expectedResults := []expected{
		{
			candlestick: types.Candlestick{
				Timestamp:      1626826200,
				OpenPrice:      29804.5,
				ClosePrice:     29804.5,
				LowestPrice:    29804.5,
				HighestPrice:   29804.5,
				Volume:         0.00000000,
				NumberOfTrades: 0,
			},
			err: nil,
		},
		{
			candlestick: types.Candlestick{
				Timestamp:      1626826260,
				OpenPrice:      29809.4,
				ClosePrice:     29809.4,
				LowestPrice:    29809.4,
				HighestPrice:   29809.4,
				Volume:         0.00024675,
				NumberOfTrades: 1,
			},
			err: nil,
		},
		{
			candlestick: types.Candlestick{
				Timestamp:      1626826320,
				OpenPrice:      29801.5,
				ClosePrice:     29801.6,
				LowestPrice:    29801.4,
				HighestPrice:   29801.7,
				Volume:         0.10940000,
				NumberOfTrades: 4,
			},
			err: nil,
		},
		{
			candlestick: types.Candlestick{
				Timestamp:      1626826380,
				OpenPrice:      29818.0,
				ClosePrice:     29818.1,
				LowestPrice:    29818.0,
				HighestPrice:   29818.1,
				Volume:         0.00215415,
				NumberOfTrades: 4,
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
