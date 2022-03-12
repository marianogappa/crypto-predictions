package common

import (
	"errors"
	"testing"
	"time"

	"github.com/marianogappa/predictions/types"
)

func TestCandlestickIterator(t *testing.T) {
	errForTest := errors.New("error for test")
	cs1 := types.Candlestick{Timestamp: 1, OpenPrice: f(1), ClosePrice: f(2), HighestPrice: f(3), LowestPrice: f(1)}
	cs2 := types.Candlestick{Timestamp: 2, OpenPrice: f(2), ClosePrice: f(3), HighestPrice: f(4), LowestPrice: f(2)}

	tci := testCandlestickIterator([]cwe{
		{cs: cs1, err: nil},
		{cs: cs2, err: errForTest},
	})

	ci := NewCandlestickIterator(tci)
	ci.SaveCandlesticks()

	actualCs1, actualErr1 := ci.Next()
	if actualCs1 != cs1 {
		t.Fatalf("expected %v but got %v", cs1, actualCs1)
	}
	if actualErr1 != nil {
		t.Fatalf("expected no error but got %v", actualErr1)
	}

	actualCs2, actualErr2 := ci.Next()
	if actualCs2 != cs2 {
		t.Fatalf("expected %v but got %v", cs2, actualCs2)
	}
	if actualErr2 != errForTest {
		t.Fatalf("expected errForTest but got %v", actualErr2)
	}

	_, actualErr3 := ci.Next()
	if actualErr3 != types.ErrOutOfCandlesticks {
		t.Fatalf("expected ErrOutOfCandlesticks but got %v", actualErr3)
	}

	if len(ci.SavedCandlesticks) != 1 {
		t.Fatalf("expected len of saved candlesticks to be 1 but got %v", len(ci.SavedCandlesticks))
	}
}

func TestGetPriceAtFails(t *testing.T) {
	ci := NewCandlestickIterator(func() (types.Candlestick, error) { return types.Candlestick{}, nil })
	_, err := ci.GetPriceAt("invalid date")
	if err == nil {
		t.Fatalf("expected GetPriceAt to fail")
	}
}
func TestGetPriceAtTimeouts(t *testing.T) {
	cs1 := types.Candlestick{Timestamp: 1, OpenPrice: f(1), ClosePrice: f(2), HighestPrice: f(3), LowestPrice: f(1)}

	tci := testCandlestickIterator([]cwe{
		{cs: cs1, err: types.ErrRateLimit},
		{cs: cs1, err: types.ErrRateLimit},
		{cs: cs1, err: types.ErrRateLimit},
		{cs: cs1, err: types.ErrRateLimit},
		{cs: cs1, err: types.ErrRateLimit},
		{cs: cs1, err: types.ErrRateLimit},
	})

	ci := NewCandlestickIterator(tci)
	ci.calmDuration = 0 * time.Second
	_, err := ci.GetPriceAt(types.ISO8601("2021-07-04T14:14:18Z"))
	if err != types.ErrRateLimit {
		t.Fatalf("expected GetPriceAt to fail with ErrRateLimit")
	}
}

func TestGetPriceAtSucceedsAfterErrTimeLimitAndOldTimestamp(t *testing.T) {
	iso := types.ISO8601("2021-07-04T14:14:18Z")
	isoTs, _ := iso.Seconds()

	cs1 := types.Candlestick{Timestamp: isoTs - 1, OpenPrice: f(1)} // Old timestamp
	cs2 := types.Candlestick{Timestamp: isoTs, OpenPrice: f(100)}   // 100!

	tci := testCandlestickIterator([]cwe{
		{cs: types.Candlestick{}, err: types.ErrRateLimit},
		{cs: types.Candlestick{}, err: types.ErrRateLimit},
		{cs: types.Candlestick{}, err: types.ErrRateLimit},
		{cs: types.Candlestick{}, err: types.ErrRateLimit},
		{cs: cs1, err: nil},
		{cs: cs2, err: nil},
	})

	ci := NewCandlestickIterator(tci)
	ci.calmDuration = 0 * time.Second
	price, err := ci.GetPriceAt(iso)
	if err != nil {
		t.Fatalf("expected GetPriceAt to not fail, but failed with %v", err)
	}
	if price != f(100) {
		t.Fatalf("expected pice to be 100.0 but was %v", price)
	}
}

type cwe struct {
	cs  types.Candlestick
	err error
}

func testCandlestickIterator(cwes []cwe) func() (types.Candlestick, error) {
	i := 0
	last := cwe{}
	return func() (types.Candlestick, error) {
		if i >= len(cwes) {
			return last.cs, types.ErrOutOfCandlesticks
		}
		i++
		last = cwes[i-1]
		return cwes[i-1].cs, cwes[i-1].err
	}
}
