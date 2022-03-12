package common

import (
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"testing"

	"github.com/marianogappa/predictions/types"
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
