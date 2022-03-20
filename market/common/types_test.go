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

func TestPatchTickHoles(t *testing.T) {
	tss := []struct {
		name     string
		ticks    []types.Tick
		startTs  int
		durSecs  int
		expected []types.Tick
	}{
		{
			name:     "Base case",
			ticks:    []types.Tick{},
			startTs:  120,
			durSecs:  60,
			expected: []types.Tick{},
		},
		{
			name:     "Does not need to do anything",
			ticks:    []types.Tick{{Timestamp: 120, Value: 1}, {Timestamp: 180, Value: 2}, {Timestamp: 240, Value: 3}},
			startTs:  120,
			durSecs:  60,
			expected: []types.Tick{{Timestamp: 120, Value: 1}, {Timestamp: 180, Value: 2}, {Timestamp: 240, Value: 3}},
		},
		{
			name:     "Removes older entries returned",
			ticks:    []types.Tick{{Timestamp: 60, Value: 2}, {Timestamp: 120, Value: 1}, {Timestamp: 180, Value: 2}, {Timestamp: 240, Value: 3}},
			startTs:  120,
			durSecs:  60,
			expected: []types.Tick{{Timestamp: 120, Value: 1}, {Timestamp: 180, Value: 2}, {Timestamp: 240, Value: 3}},
		},
		{
			name:     "Removes older entries returned, leaving nothing",
			ticks:    []types.Tick{{Timestamp: 60, Value: 2}},
			startTs:  120,
			durSecs:  60,
			expected: []types.Tick{},
		},
		{
			name:     "Needs to add an initial tick",
			ticks:    []types.Tick{{Timestamp: 180, Value: 2}, {Timestamp: 240, Value: 3}},
			startTs:  120,
			durSecs:  60,
			expected: []types.Tick{{Timestamp: 120, Value: 2}, {Timestamp: 180, Value: 2}, {Timestamp: 240, Value: 3}},
		},
		{
			name:     "Needs to add an initial tick, as well as in the middle",
			ticks:    []types.Tick{{Timestamp: 180, Value: 2}, {Timestamp: 360, Value: 3}},
			startTs:  120,
			durSecs:  60,
			expected: []types.Tick{{Timestamp: 120, Value: 2}, {Timestamp: 180, Value: 2}, {Timestamp: 240, Value: 3}, {Timestamp: 300, Value: 3}, {Timestamp: 360, Value: 3}},
		},
		{
			name:     "Adjusts start time to zero seconds",
			ticks:    []types.Tick{{Timestamp: tInt("2020-01-02 00:03:00"), Value: 1}, {Timestamp: tInt("2020-01-02 00:04:00"), Value: 2}, {Timestamp: tInt("2020-01-02 00:05:00"), Value: 3}},
			startTs:  tInt("2020-01-02 00:02:58"),
			durSecs:  60,
			expected: []types.Tick{{Timestamp: tInt("2020-01-02 00:03:00"), Value: 1}, {Timestamp: tInt("2020-01-02 00:04:00"), Value: 2}, {Timestamp: tInt("2020-01-02 00:05:00"), Value: 3}},
		},
		{
			name:     "Adjusts start time to zero seconds rounding up",
			ticks:    []types.Tick{{Timestamp: tInt("2020-01-02 00:03:00"), Value: 1}, {Timestamp: tInt("2020-01-02 00:04:00"), Value: 2}, {Timestamp: tInt("2020-01-02 00:05:00"), Value: 3}},
			startTs:  tInt("2020-01-02 00:03:02"),
			durSecs:  60,
			expected: []types.Tick{{Timestamp: tInt("2020-01-02 00:04:00"), Value: 2}, {Timestamp: tInt("2020-01-02 00:05:00"), Value: 3}},
		},
	}
	for _, ts := range tss {
		t.Run(ts.name, func(t *testing.T) {
			actual := PatchTickHoles(ts.ticks, ts.startTs, ts.durSecs)
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
