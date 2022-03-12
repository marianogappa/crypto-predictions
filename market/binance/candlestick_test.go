package binance

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
			[1626766200000,"29737.59000000","29747.75000000","29737.58000000","29741.31000000","23.43370400",1626766259999,"696965.08012368",468,"11.14470300","331446.42584469","0"],
			[1626766260000,"29741.30000000","29757.01000000","29741.30000000","29751.19000000","23.84928700",1626766319999,"709465.45886498",414,"15.85058900","471521.15833536","0"]
		]`,
		`[
			[1626766320000,"29751.19000000","29754.86000000","29741.31000000","29751.04000000","10.33420300",1626766379999,"307431.19906366",394,"5.88530500","175077.01040434","0"],
			[1626766380000,"29751.03000000","29759.75000000","29735.00000000","29745.00000000","26.49070300",1626766439999,"787968.66501515",416,"10.20193300","303518.55853338","0"]
		]`,
		`[]`,
	}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, replies[i%len(replies)])
		i++
	}))
	defer ts.Close()

	b := NewBinance()
	b.overrideAPIURL(ts.URL + "/")
	ci := b.BuildCandlestickIterator("BTC", "USDT", "2021-07-04T14:14:18+00:00")

	expectedResults := []expected{
		{
			candlestick: types.Candlestick{
				Timestamp:      1626766200,
				OpenPrice:      29737.59,
				ClosePrice:     29741.31,
				LowestPrice:    29737.58,
				HighestPrice:   29747.75,
				Volume:         23.433704,
				NumberOfTrades: 468,
			},
			err: nil,
		},
		{
			candlestick: types.Candlestick{
				Timestamp:      1626766260,
				OpenPrice:      29741.3,
				ClosePrice:     29751.19,
				LowestPrice:    29741.3,
				HighestPrice:   29757.01,
				Volume:         23.849287,
				NumberOfTrades: 414,
			},
			err: nil,
		},
		{
			candlestick: types.Candlestick{
				Timestamp:      1626766320,
				OpenPrice:      29751.19,
				ClosePrice:     29751.04,
				LowestPrice:    29741.31,
				HighestPrice:   29754.86,
				Volume:         10.334203,
				NumberOfTrades: 394,
			},
			err: nil,
		},
		{
			candlestick: types.Candlestick{
				Timestamp:      1626766380,
				OpenPrice:      29751.03,
				ClosePrice:     29745,
				LowestPrice:    29735,
				HighestPrice:   29759.75,
				Volume:         26.490703,
				NumberOfTrades: 416,
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
