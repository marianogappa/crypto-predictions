package common

import (
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"testing"
	"time"

	"github.com/marianogappa/predictions/types"
	"github.com/stretchr/testify/require"
)

func TestJsonFloat64(t *testing.T) {
	tss := []struct {
		f        float64
		expected string
	}{
		{f: 1.2, expected: "1.2"},
		{f: 0.0000001234, expected: "0.0000001234"},
		{f: 1.000000, expected: "1"},
		{f: 0.000000, expected: "0"},
		{f: 0.001000, expected: "0.001"},
		{f: 10.0, expected: "10"},
	}
	for _, ts := range tss {
		t.Run(ts.expected, func(t *testing.T) {
			bs, err := json.Marshal(types.JsonFloat64(ts.f))
			if err != nil {
				t.Fatalf("Marshalling failed with %v", err)
			}
			if string(bs) != ts.expected {
				t.Fatalf("Expected marshalling of %f to be exactly '%v' but was '%v'", ts.f, ts.expected, string(bs))
			}
		})
	}
}

func TestJsonFloat64Fails(t *testing.T) {
	tss := []struct {
		f float64
	}{
		{f: math.Inf(1)},
		{f: math.NaN()},
	}
	for _, ts := range tss {
		t.Run(fmt.Sprintf("%f", ts.f), func(t *testing.T) {
			_, err := json.Marshal(types.JsonFloat64(ts.f))
			if err == nil {
				t.Fatal("Expected marshalling to fail")
			}
		})
	}
}

func TestToMillis(t *testing.T) {
	ms, err := types.ISO8601("2021-07-04T14:14:18Z").Millis()
	if err != nil {
		t.Fatalf("should not have errored, but errored with %v", err)
	}
	if ms != 162540805800 {
		t.Fatalf("expected ms to be %v but were %v", 162540805800, ms)
	}

	_, err = types.ISO8601("invalid").Millis()
	if err == nil {
		t.Fatal("should have errored, but didn't")
	}
}

func TestCandlestickToTicks(t *testing.T) {
	ticks := types.Candlestick{
		Timestamp:      1499040000,
		OpenPrice:      f(0.01634790),
		ClosePrice:     f(0.01577100),
		LowestPrice:    f(0.01575800),
		HighestPrice:   f(0.80000000),
		Volume:         f(148976.11427815),
		NumberOfTrades: 308,
	}.ToTicks()

	if len(ticks) != 2 {
		t.Fatalf("expected len(ticks) == 2 but was %v", len(ticks))
	}

	expectedTicks := []types.Tick{
		{
			Timestamp: 1499040000,
			Value:     f(0.01575800),
		},
		{
			Timestamp: 1499040000,
			Value:     f(0.80000000),
		},
	}

	if !reflect.DeepEqual(expectedTicks, ticks) {
		t.Fatalf("expected ticks to be %v but were %v", expectedTicks, ticks)
	}
}

func f(fl float64) types.JsonFloat64 {
	return types.JsonFloat64(fl)
}

func TestPatchCandlestickHoles(t *testing.T) {
	tss := []struct {
		name         string
		candlesticks []types.Candlestick
		startTs      int
		durSecs      int
		expected     []types.Candlestick
	}{
		{
			name:         "Base case",
			candlesticks: []types.Candlestick{},
			startTs:      120,
			durSecs:      60,
			expected:     []types.Candlestick{},
		},
		{
			name: "Does not need to do anything",
			candlesticks: []types.Candlestick{
				{Timestamp: 120, OpenPrice: 1, HighestPrice: 1, ClosePrice: 1, LowestPrice: 1},
				{Timestamp: 180, OpenPrice: 2, HighestPrice: 2, ClosePrice: 2, LowestPrice: 2},
				{Timestamp: 240, OpenPrice: 3, HighestPrice: 3, ClosePrice: 3, LowestPrice: 3},
			},
			startTs: 120,
			durSecs: 60,
			expected: []types.Candlestick{
				{Timestamp: 120, OpenPrice: 1, HighestPrice: 1, ClosePrice: 1, LowestPrice: 1},
				{Timestamp: 180, OpenPrice: 2, HighestPrice: 2, ClosePrice: 2, LowestPrice: 2},
				{Timestamp: 240, OpenPrice: 3, HighestPrice: 3, ClosePrice: 3, LowestPrice: 3},
			},
		},
		{
			name: "Removes older entries returned",
			candlesticks: []types.Candlestick{
				{Timestamp: 60, OpenPrice: 2, HighestPrice: 2, ClosePrice: 2, LowestPrice: 2},
				{Timestamp: 120, OpenPrice: 1, HighestPrice: 1, ClosePrice: 1, LowestPrice: 1},
				{Timestamp: 180, OpenPrice: 2, HighestPrice: 2, ClosePrice: 2, LowestPrice: 2},
				{Timestamp: 240, OpenPrice: 3, HighestPrice: 3, ClosePrice: 3, LowestPrice: 3},
			},
			startTs: 120,
			durSecs: 60,
			expected: []types.Candlestick{
				{Timestamp: 120, OpenPrice: 1, HighestPrice: 1, ClosePrice: 1, LowestPrice: 1},
				{Timestamp: 180, OpenPrice: 2, HighestPrice: 2, ClosePrice: 2, LowestPrice: 2},
				{Timestamp: 240, OpenPrice: 3, HighestPrice: 3, ClosePrice: 3, LowestPrice: 3},
			},
		},
		{
			name: "Removes older entries returned, leaving nothing",
			candlesticks: []types.Candlestick{
				{Timestamp: 60, OpenPrice: 2, HighestPrice: 2, ClosePrice: 2, LowestPrice: 2},
			},
			startTs:  120,
			durSecs:  60,
			expected: []types.Candlestick{},
		},
		{
			name: "Needs to add an initial tick",
			candlesticks: []types.Candlestick{
				{Timestamp: 180, OpenPrice: 2, HighestPrice: 2, ClosePrice: 2, LowestPrice: 2},
				{Timestamp: 240, OpenPrice: 3, HighestPrice: 3, ClosePrice: 3, LowestPrice: 3},
			},
			startTs: 120,
			durSecs: 60,
			expected: []types.Candlestick{
				{Timestamp: 120, OpenPrice: 2, HighestPrice: 2, ClosePrice: 2, LowestPrice: 2},
				{Timestamp: 180, OpenPrice: 2, HighestPrice: 2, ClosePrice: 2, LowestPrice: 2},
				{Timestamp: 240, OpenPrice: 3, HighestPrice: 3, ClosePrice: 3, LowestPrice: 3},
			},
		},
		{
			name: "Needs to add an initial tick, as well as in the middle",
			candlesticks: []types.Candlestick{
				{Timestamp: 180, OpenPrice: 2, HighestPrice: 2, ClosePrice: 2, LowestPrice: 2},
				{Timestamp: 360, OpenPrice: 3, HighestPrice: 3, ClosePrice: 3, LowestPrice: 3},
			},
			startTs: 120,
			durSecs: 60,
			expected: []types.Candlestick{
				{Timestamp: 120, OpenPrice: 2, HighestPrice: 2, ClosePrice: 2, LowestPrice: 2},
				{Timestamp: 180, OpenPrice: 2, HighestPrice: 2, ClosePrice: 2, LowestPrice: 2},
				{Timestamp: 240, OpenPrice: 3, HighestPrice: 3, ClosePrice: 3, LowestPrice: 3},
				{Timestamp: 300, OpenPrice: 3, HighestPrice: 3, ClosePrice: 3, LowestPrice: 3},
				{Timestamp: 360, OpenPrice: 3, HighestPrice: 3, ClosePrice: 3, LowestPrice: 3},
			},
		},
		{
			name: "Adjusts start time to zero seconds",
			candlesticks: []types.Candlestick{
				{Timestamp: tInt("2020-01-02 00:03:00"), OpenPrice: 1, HighestPrice: 1, ClosePrice: 1, LowestPrice: 1},
				{Timestamp: tInt("2020-01-02 00:04:00"), OpenPrice: 2, HighestPrice: 2, ClosePrice: 2, LowestPrice: 2},
				{Timestamp: tInt("2020-01-02 00:05:00"), OpenPrice: 3, HighestPrice: 3, ClosePrice: 3, LowestPrice: 3},
			},
			startTs: tInt("2020-01-02 00:02:58"),
			durSecs: 60,
			expected: []types.Candlestick{
				{Timestamp: tInt("2020-01-02 00:03:00"), OpenPrice: 1, HighestPrice: 1, ClosePrice: 1, LowestPrice: 1},
				{Timestamp: tInt("2020-01-02 00:04:00"), OpenPrice: 2, HighestPrice: 2, ClosePrice: 2, LowestPrice: 2},
				{Timestamp: tInt("2020-01-02 00:05:00"), OpenPrice: 3, HighestPrice: 3, ClosePrice: 3, LowestPrice: 3},
			},
		},
		{
			name: "Adjusts start time to zero seconds rounding up",
			candlesticks: []types.Candlestick{
				{Timestamp: tInt("2020-01-02 00:03:00"), OpenPrice: 1, HighestPrice: 1, ClosePrice: 1, LowestPrice: 1},
				{Timestamp: tInt("2020-01-02 00:04:00"), OpenPrice: 2, HighestPrice: 2, ClosePrice: 2, LowestPrice: 2},
				{Timestamp: tInt("2020-01-02 00:05:00"), OpenPrice: 3, HighestPrice: 3, ClosePrice: 3, LowestPrice: 3},
			},
			startTs: tInt("2020-01-02 00:03:02"),
			durSecs: 60,
			expected: []types.Candlestick{
				{Timestamp: tInt("2020-01-02 00:04:00"), OpenPrice: 2, HighestPrice: 2, ClosePrice: 2, LowestPrice: 2},
				{Timestamp: tInt("2020-01-02 00:05:00"), OpenPrice: 3, HighestPrice: 3, ClosePrice: 3, LowestPrice: 3},
			},
		},
	}
	for _, ts := range tss {
		t.Run(ts.name, func(t *testing.T) {
			actual := PatchCandlestickHoles(ts.candlesticks, ts.startTs, ts.durSecs)
			require.Equal(t, ts.expected, actual)
		})
	}
}

func tp(s string) time.Time {
	t, _ := time.Parse("2006-01-02 15:04:05", s)
	return t
}

func tInt(s string) int {
	return int(tp(s).Unix())
}
