package cache

import (
	"errors"
	"testing"
	"time"

	"github.com/marianogappa/predictions/types"
	"github.com/stretchr/testify/require"
)

type operation struct {
	opType              string
	operand             types.Operand
	candlestickInterval time.Duration
	candlesticks        []types.Candlestick
	initialISO8601      types.ISO8601
	expectedErr         error
	expectedTicks       []types.Candlestick
}

func TestCache(t *testing.T) {
	opBTCUSDT := types.Operand{
		Type:       types.COIN,
		Provider:   "BINANCE",
		BaseAsset:  "BTC",
		QuoteAsset: "USDT",
		Str:        "COIN:BINANCE:BTC-USDT",
	}
	opETHUSDT := types.Operand{
		Type:       types.COIN,
		Provider:   "BINANCE",
		BaseAsset:  "ETH",
		QuoteAsset: "USDT",
		Str:        "COIN:BINANCE:ETH-USDT",
	}
	opBTC := types.Operand{
		Type:      types.MARKETCAP,
		Provider:  "MESSARI",
		BaseAsset: "BTC",
		Str:       "MARKETCAP:MESSARI:BTC",
	}
	opETH := types.Operand{
		Type:      types.MARKETCAP,
		Provider:  "MESSARI",
		BaseAsset: "ETH",
		Str:       "MARKETCAP:MESSARI:ETH",
	}

	tss := []struct {
		name string
		ops  []operation
	}{
		// Minutely tests
		{
			name: "MINUTELY: Get empty returns ErrCacheMiss",
			ops: []operation{
				{
					opType:              "GET",
					operand:             opBTCUSDT,
					candlestickInterval: 1 * time.Minute,
					initialISO8601:      tpToISO("2020-01-02 03:04:00"),
					expectedErr:         ErrCacheMiss,
					expectedTicks:       []types.Candlestick{},
				},
			},
		},
		{
			name: "MINUTELY: Get with an invalid date returns ErrInvalidISO8601",
			ops: []operation{
				{
					opType:              "GET",
					operand:             opBTCUSDT,
					candlestickInterval: 1 * time.Minute,
					initialISO8601:      types.ISO8601("invalid"),
					expectedErr:         ErrInvalidISO8601,
					expectedTicks:       []types.Candlestick{},
				},
			},
		},
		{
			name: "MINUTELY: Put empty returns empty",
			ops: []operation{
				{
					opType:              "PUT",
					operand:             opBTCUSDT,
					candlestickInterval: 1 * time.Minute,
					candlesticks:        []types.Candlestick{},
					expectedErr:         nil,
				},
			},
		},
		{
			name: "MINUTELY: Put with non-zero second fails",
			ops: []operation{
				{
					opType:              "PUT",
					operand:             opBTCUSDT,
					candlestickInterval: 1 * time.Minute,
					candlesticks: []types.Candlestick{
						{Timestamp: tInt("2020-01-02 03:04:01"), OpenPrice: 1234, HighestPrice: 1234, ClosePrice: 1234, LowestPrice: 1234},
					},
					expectedErr: ErrTimestampMustBeMultipleOfCandlestickInterval,
				},
			},
		},
		{
			name: "MINUTELY: Put with zero value fails",
			ops: []operation{
				{
					opType:              "PUT",
					operand:             opBTCUSDT,
					candlestickInterval: 1 * time.Minute,
					candlesticks: []types.Candlestick{
						{Timestamp: tInt("2020-01-02 03:04:00"), OpenPrice: 1234, HighestPrice: 1234, ClosePrice: 1234, LowestPrice: 1234},
						{Timestamp: tInt("2020-01-02 03:05:00"), OpenPrice: 0, HighestPrice: 1234, ClosePrice: 1234, LowestPrice: 1234},
					},
					expectedErr: ErrReceivedCandlestickWithZeroValue,
				},
			},
		},
		{
			name: "MINUTELY: Put with non-subsequent timestamps fails",
			ops: []operation{
				{
					opType:              "PUT",
					operand:             opBTCUSDT,
					candlestickInterval: 1 * time.Minute,
					candlesticks: []types.Candlestick{
						{Timestamp: tInt("2020-01-02 03:04:00"), OpenPrice: 1234, HighestPrice: 1234, ClosePrice: 1234, LowestPrice: 1234},
						{Timestamp: tInt("2020-01-02 03:06:00"), OpenPrice: 1234, HighestPrice: 1234, ClosePrice: 1234, LowestPrice: 1234},
					},
					expectedErr: ErrReceivedNonSubsequentCandlestick,
				},
			},
		},
		{
			name: "MINUTELY: Put with non-zero seconds fails",
			ops: []operation{
				{
					opType:              "PUT",
					operand:             opBTCUSDT,
					candlestickInterval: 1 * time.Minute,
					candlesticks: []types.Candlestick{
						{Timestamp: tInt("2020-01-02 03:04:01"), OpenPrice: 1234, HighestPrice: 1234, ClosePrice: 1234, LowestPrice: 1234},
					},
					expectedErr: ErrTimestampMustBeMultipleOfCandlestickInterval,
				},
			},
		},
		{
			name: "MINUTELY: Valid Put succeeds",
			ops: []operation{
				{
					opType:              "PUT",
					operand:             opBTCUSDT,
					candlestickInterval: 1 * time.Minute,
					candlesticks: []types.Candlestick{
						{Timestamp: tInt("2020-01-02 03:04:00"), OpenPrice: 1234, HighestPrice: 1234, ClosePrice: 1234, LowestPrice: 1234},
						{Timestamp: tInt("2020-01-02 03:05:00"), OpenPrice: 2345, HighestPrice: 2345, ClosePrice: 2345, LowestPrice: 2345},
					},
					expectedErr: nil,
				},
			},
		},
		{
			name: "MINUTELY: Valid Put succeeds, and a get of a different key does not return anything, but a get of same key works",
			ops: []operation{
				{
					opType:              "PUT",
					operand:             opBTCUSDT,
					candlestickInterval: 1 * time.Minute,
					candlesticks: []types.Candlestick{
						{Timestamp: tInt("2020-01-02 03:04:00"), OpenPrice: 1234, HighestPrice: 1234, ClosePrice: 1234, LowestPrice: 1234},
						{Timestamp: tInt("2020-01-02 03:05:00"), OpenPrice: 2345, HighestPrice: 2345, ClosePrice: 2345, LowestPrice: 2345},
					},
					expectedErr: nil,
				},
				{
					opType:              "GET",
					operand:             opETHUSDT,
					candlestickInterval: 1 * time.Minute,
					initialISO8601:      tpToISO("2020-01-02 03:04:00"),
					expectedErr:         ErrCacheMiss,
					expectedTicks:       []types.Candlestick{},
				},
				{
					opType:              "GET",
					operand:             opBTCUSDT,
					candlestickInterval: 1 * time.Minute,
					initialISO8601:      tpToISO("2020-01-02 03:04:00"),
					expectedErr:         nil,
					expectedTicks: []types.Candlestick{
						{Timestamp: tInt("2020-01-02 03:04:00"), OpenPrice: 1234, HighestPrice: 1234, ClosePrice: 1234, LowestPrice: 1234},
						{Timestamp: tInt("2020-01-02 03:05:00"), OpenPrice: 2345, HighestPrice: 2345, ClosePrice: 2345, LowestPrice: 2345},
					},
				},
			},
		},
		{
			name: "MINUTELY: A secondary PUT overrides the first one's values, with full overlap",
			ops: []operation{
				{
					opType:              "PUT",
					operand:             opBTCUSDT,
					candlestickInterval: 1 * time.Minute,
					candlesticks: []types.Candlestick{
						{Timestamp: tInt("2020-01-02 03:04:00"), OpenPrice: 1234, HighestPrice: 1234, ClosePrice: 1234, LowestPrice: 1234},
						{Timestamp: tInt("2020-01-02 03:05:00"), OpenPrice: 2345, HighestPrice: 2345, ClosePrice: 2345, LowestPrice: 2345},
					},
					expectedErr: nil,
				},
				{
					opType:              "PUT",
					operand:             opBTCUSDT,
					candlestickInterval: 1 * time.Minute,
					candlesticks: []types.Candlestick{
						{Timestamp: tInt("2020-01-02 03:04:00"), OpenPrice: 2345, HighestPrice: 2345, ClosePrice: 2345, LowestPrice: 2345},
						{Timestamp: tInt("2020-01-02 03:05:00"), OpenPrice: 3456, HighestPrice: 3456, ClosePrice: 3456, LowestPrice: 3456},
					},
					expectedErr: nil,
				},
				{
					opType:              "GET",
					operand:             opBTCUSDT,
					candlestickInterval: 1 * time.Minute,
					initialISO8601:      tpToISO("2020-01-02 03:04:00"),
					expectedErr:         nil,
					expectedTicks: []types.Candlestick{
						{Timestamp: tInt("2020-01-02 03:04:00"), OpenPrice: 2345, HighestPrice: 2345, ClosePrice: 2345, LowestPrice: 2345},
						{Timestamp: tInt("2020-01-02 03:05:00"), OpenPrice: 3456, HighestPrice: 3456, ClosePrice: 3456, LowestPrice: 3456},
					},
				},
			},
		},
		{
			name: "MINUTELY: A secondary PUT with overlap makes the sequence larger on GET",
			ops: []operation{
				{
					opType:              "PUT",
					operand:             opBTCUSDT,
					candlestickInterval: 1 * time.Minute,
					candlesticks: []types.Candlestick{
						{Timestamp: tInt("2020-01-02 03:04:00"), OpenPrice: 1234, HighestPrice: 1234, ClosePrice: 1234, LowestPrice: 1234},
						{Timestamp: tInt("2020-01-02 03:05:00"), OpenPrice: 2345, HighestPrice: 2345, ClosePrice: 2345, LowestPrice: 2345},
					},
					expectedErr: nil,
				},
				{
					opType:              "PUT",
					operand:             opBTCUSDT,
					candlestickInterval: 1 * time.Minute,
					candlesticks: []types.Candlestick{
						{Timestamp: tInt("2020-01-02 03:05:00"), OpenPrice: 2345, HighestPrice: 2345, ClosePrice: 2345, LowestPrice: 2345},
						{Timestamp: tInt("2020-01-02 03:06:00"), OpenPrice: 3456, HighestPrice: 3456, ClosePrice: 3456, LowestPrice: 3456},
					},
					expectedErr: nil,
				},
				{
					opType:              "GET",
					operand:             opBTCUSDT,
					candlestickInterval: 1 * time.Minute,
					initialISO8601:      tpToISO("2020-01-02 03:04:00"),
					expectedErr:         nil,
					expectedTicks: []types.Candlestick{
						{Timestamp: tInt("2020-01-02 03:04:00"), OpenPrice: 1234, HighestPrice: 1234, ClosePrice: 1234, LowestPrice: 1234},
						{Timestamp: tInt("2020-01-02 03:05:00"), OpenPrice: 2345, HighestPrice: 2345, ClosePrice: 2345, LowestPrice: 2345},
						{Timestamp: tInt("2020-01-02 03:06:00"), OpenPrice: 3456, HighestPrice: 3456, ClosePrice: 3456, LowestPrice: 3456},
					},
				},
			},
		},
		{
			name: "MINUTELY: A secondary PUT without overlap does not make the sequence larger on GET, and a second get gets the other one",
			ops: []operation{
				{
					opType:              "PUT",
					operand:             opBTCUSDT,
					candlestickInterval: 1 * time.Minute,
					candlesticks: []types.Candlestick{
						{Timestamp: tInt("2020-01-02 03:04:00"), OpenPrice: 1234, HighestPrice: 1234, ClosePrice: 1234, LowestPrice: 1234},
						{Timestamp: tInt("2020-01-02 03:05:00"), OpenPrice: 2345, HighestPrice: 2345, ClosePrice: 2345, LowestPrice: 2345},
					},
					expectedErr: nil,
				},
				{
					opType:              "PUT",
					operand:             opBTCUSDT,
					candlestickInterval: 1 * time.Minute,
					candlesticks: []types.Candlestick{
						{Timestamp: tInt("2020-01-02 03:07:00"), OpenPrice: 3456, HighestPrice: 3456, ClosePrice: 3456, LowestPrice: 3456},
						{Timestamp: tInt("2020-01-02 03:08:00"), OpenPrice: 4567, HighestPrice: 4567, ClosePrice: 4567, LowestPrice: 4567},
					},
					expectedErr: nil,
				},
				{
					opType:              "GET",
					operand:             opBTCUSDT,
					candlestickInterval: 1 * time.Minute,
					initialISO8601:      tpToISO("2020-01-02 03:04:00"),
					expectedErr:         nil,
					expectedTicks: []types.Candlestick{
						{Timestamp: tInt("2020-01-02 03:04:00"), OpenPrice: 1234, HighestPrice: 1234, ClosePrice: 1234, LowestPrice: 1234},
						{Timestamp: tInt("2020-01-02 03:05:00"), OpenPrice: 2345, HighestPrice: 2345, ClosePrice: 2345, LowestPrice: 2345},
					},
				},
				{
					opType:              "GET",
					operand:             opBTCUSDT,
					candlestickInterval: 1 * time.Minute,
					initialISO8601:      tpToISO("2020-01-02 03:07:00"),
					expectedErr:         nil,
					expectedTicks: []types.Candlestick{
						{Timestamp: tInt("2020-01-02 03:07:00"), OpenPrice: 3456, HighestPrice: 3456, ClosePrice: 3456, LowestPrice: 3456},
						{Timestamp: tInt("2020-01-02 03:08:00"), OpenPrice: 4567, HighestPrice: 4567, ClosePrice: 4567, LowestPrice: 4567},
					},
				},
			},
		},
		{
			name: "MINUTELY: Two separate series work at the same time",
			ops: []operation{
				{
					opType:              "PUT",
					operand:             opBTCUSDT,
					candlestickInterval: 1 * time.Minute,
					candlesticks: []types.Candlestick{
						{Timestamp: tInt("2020-01-02 03:04:00"), OpenPrice: 1234, HighestPrice: 1234, ClosePrice: 1234, LowestPrice: 1234},
						{Timestamp: tInt("2020-01-02 03:05:00"), OpenPrice: 2345, HighestPrice: 2345, ClosePrice: 2345, LowestPrice: 2345},
					},
					expectedErr: nil,
				},
				{
					opType:              "PUT",
					operand:             opETHUSDT,
					candlestickInterval: 1 * time.Minute,
					candlesticks: []types.Candlestick{
						{Timestamp: tInt("2020-01-02 03:04:00"), OpenPrice: 3456, HighestPrice: 3456, ClosePrice: 3456, LowestPrice: 3456},
						{Timestamp: tInt("2020-01-02 03:05:00"), OpenPrice: 4567, HighestPrice: 4567, ClosePrice: 4567, LowestPrice: 4567},
					},
					expectedErr: nil,
				},
				{
					opType:              "GET",
					operand:             opBTCUSDT,
					candlestickInterval: 1 * time.Minute,
					initialISO8601:      tpToISO("2020-01-02 03:04:00"),
					expectedErr:         nil,
					expectedTicks: []types.Candlestick{
						{Timestamp: tInt("2020-01-02 03:04:00"), OpenPrice: 1234, HighestPrice: 1234, ClosePrice: 1234, LowestPrice: 1234},
						{Timestamp: tInt("2020-01-02 03:05:00"), OpenPrice: 2345, HighestPrice: 2345, ClosePrice: 2345, LowestPrice: 2345},
					},
				},
				{
					opType:              "GET",
					operand:             opETHUSDT,
					candlestickInterval: 1 * time.Minute,
					initialISO8601:      tpToISO("2020-01-02 03:04:00"),
					expectedErr:         nil,
					expectedTicks: []types.Candlestick{
						{Timestamp: tInt("2020-01-02 03:04:00"), OpenPrice: 3456, HighestPrice: 3456, ClosePrice: 3456, LowestPrice: 3456},
						{Timestamp: tInt("2020-01-02 03:05:00"), OpenPrice: 4567, HighestPrice: 4567, ClosePrice: 4567, LowestPrice: 4567},
					},
				},
			},
		},
		{
			name: "MINUTELY: Get of a day on an empty time is a cache miss",
			ops: []operation{
				{
					opType:              "PUT",
					operand:             opBTCUSDT,
					candlestickInterval: 1 * time.Minute,
					candlesticks: []types.Candlestick{
						{Timestamp: tInt("2020-01-02 03:04:00"), OpenPrice: 1234, HighestPrice: 1234, ClosePrice: 1234, LowestPrice: 1234},
						{Timestamp: tInt("2020-01-02 03:05:00"), OpenPrice: 2345, HighestPrice: 2345, ClosePrice: 2345, LowestPrice: 2345},
					},
					expectedErr: nil,
				},
				{
					opType:              "GET",
					operand:             opBTCUSDT,
					candlestickInterval: 1 * time.Minute,
					initialISO8601:      tpToISO("2020-01-02 03:06:00"),
					expectedErr:         ErrCacheMiss,
					expectedTicks:       []types.Candlestick{},
				},
			},
		},
		{
			name: "MINUTELY: Get of a minute before, but with non-zero seconds, returns the tick of the next minute",
			ops: []operation{
				{
					opType:              "PUT",
					operand:             opBTCUSDT,
					candlestickInterval: 1 * time.Minute,
					candlesticks: []types.Candlestick{
						{Timestamp: tInt("2020-01-02 03:04:00"), OpenPrice: 1234, HighestPrice: 1234, ClosePrice: 1234, LowestPrice: 1234},
						{Timestamp: tInt("2020-01-02 03:05:00"), OpenPrice: 2345, HighestPrice: 2345, ClosePrice: 2345, LowestPrice: 2345},
					},
					expectedErr: nil,
				},
				{
					opType:              "GET",
					operand:             opBTCUSDT,
					candlestickInterval: 1 * time.Minute,
					initialISO8601:      tpToISO("2020-01-02 03:03:01"),
					expectedErr:         nil,
					expectedTicks: []types.Candlestick{
						{Timestamp: tInt("2020-01-02 03:04:00"), OpenPrice: 1234, HighestPrice: 1234, ClosePrice: 1234, LowestPrice: 1234},
						{Timestamp: tInt("2020-01-02 03:05:00"), OpenPrice: 2345, HighestPrice: 2345, ClosePrice: 2345, LowestPrice: 2345},
					},
				},
			},
		},
		{
			name: "MINUTELY: Putting ticks that span two truncated intervals works, but requires two gets to get both ticks",
			ops: []operation{
				{
					opType:              "PUT",
					operand:             opBTCUSDT,
					candlestickInterval: 1 * time.Minute,
					candlesticks: []types.Candlestick{
						{Timestamp: tInt("2020-01-02 16:39:00"), OpenPrice: 1234, HighestPrice: 1234, ClosePrice: 1234, LowestPrice: 1234},
						{Timestamp: tInt("2020-01-02 16:40:00"), OpenPrice: 2345, HighestPrice: 2345, ClosePrice: 2345, LowestPrice: 2345},
					},
					expectedErr: nil,
				},
				{
					opType:              "GET",
					operand:             opBTCUSDT,
					candlestickInterval: 1 * time.Minute,
					initialISO8601:      tpToISO("2020-01-02 16:39:00"),
					expectedErr:         nil,
					expectedTicks: []types.Candlestick{
						{Timestamp: tInt("2020-01-02 16:39:00"), OpenPrice: 1234, HighestPrice: 1234, ClosePrice: 1234, LowestPrice: 1234},
					},
				},
				{
					opType:              "GET",
					operand:             opBTCUSDT,
					candlestickInterval: 1 * time.Minute,
					initialISO8601:      tpToISO("2020-01-02 16:40:00"),
					expectedErr:         nil,
					expectedTicks: []types.Candlestick{
						{Timestamp: tInt("2020-01-02 16:40:00"), OpenPrice: 2345, HighestPrice: 2345, ClosePrice: 2345, LowestPrice: 2345},
					},
				},
			},
		},
		// Daily tests
		{
			name: "DAILY: Get empty returns ErrCacheMiss",
			ops: []operation{
				{
					opType:              "GET",
					operand:             opBTC,
					candlestickInterval: 24 * time.Hour,
					initialISO8601:      tpToISO("2020-01-02 03:04:00"),
					expectedErr:         ErrCacheMiss,
					expectedTicks:       []types.Candlestick{},
				},
			},
		},
		{
			name: "DAILY: Get with an invalid date returns ErrInvalidISO8601",
			ops: []operation{
				{
					opType:              "GET",
					operand:             opBTC,
					candlestickInterval: 24 * time.Hour,
					initialISO8601:      types.ISO8601("invalid"),
					expectedErr:         ErrInvalidISO8601,
					expectedTicks:       []types.Candlestick{},
				},
			},
		},
		{
			name: "DAILY: Put empty returns empty",
			ops: []operation{
				{
					opType:              "PUT",
					operand:             opBTC,
					candlestickInterval: 24 * time.Hour,
					candlesticks:        []types.Candlestick{},
					expectedErr:         nil,
				},
			},
		},
		{
			name: "DAILY: Put with non-zero second fails",
			ops: []operation{
				{
					opType:              "PUT",
					operand:             opBTC,
					candlestickInterval: 24 * time.Hour,
					candlesticks: []types.Candlestick{
						{Timestamp: tInt("2020-01-02 03:00:00"), OpenPrice: 1234, HighestPrice: 1234, ClosePrice: 1234, LowestPrice: 1234},
					},
					expectedErr: ErrTimestampMustBeMultipleOfCandlestickInterval,
				},
			},
		},
		{
			name: "DAILY: Put with zero value fails",
			ops: []operation{
				{
					opType:              "PUT",
					operand:             opBTC,
					candlestickInterval: 24 * time.Hour,
					candlesticks: []types.Candlestick{
						{Timestamp: tInt("2020-01-02 00:00:00"), OpenPrice: 1234, HighestPrice: 1234, ClosePrice: 1234, LowestPrice: 1234},
						{Timestamp: tInt("2020-01-03 00:00:00"), OpenPrice: 0, HighestPrice: 1234, ClosePrice: 1234, LowestPrice: 1234},
					},
					expectedErr: ErrReceivedCandlestickWithZeroValue,
				},
			},
		},
		{
			name: "DAILY: Put with non-subsequent timestamps fails",
			ops: []operation{
				{
					opType:              "PUT",
					operand:             opBTC,
					candlestickInterval: 24 * time.Hour,
					candlesticks: []types.Candlestick{
						{Timestamp: tInt("2020-01-02 00:00:00"), OpenPrice: 1234, HighestPrice: 1234, ClosePrice: 1234, LowestPrice: 1234},
						{Timestamp: tInt("2020-01-04 00:00:00"), OpenPrice: 2345, HighestPrice: 2345, ClosePrice: 2345, LowestPrice: 2345},
					},
					expectedErr: ErrReceivedNonSubsequentCandlestick,
				},
			},
		},
		{
			name: "DAILY: Valid Put succeeds",
			ops: []operation{
				{
					opType:              "PUT",
					operand:             opBTC,
					candlestickInterval: 24 * time.Hour,
					candlesticks: []types.Candlestick{
						{Timestamp: tInt("2020-01-02 00:00:00"), OpenPrice: 1234, HighestPrice: 1234, ClosePrice: 1234, LowestPrice: 1234},
						{Timestamp: tInt("2020-01-03 00:00:00"), OpenPrice: 2345, HighestPrice: 2345, ClosePrice: 2345, LowestPrice: 2345},
					},
					expectedErr: nil,
				},
			},
		},
		{
			name: "DAILY: Valid Put succeeds, and a get of a different key does not return anything, but a get of same key works",
			ops: []operation{
				{
					opType:              "PUT",
					operand:             opBTC,
					candlestickInterval: 24 * time.Hour,
					candlesticks: []types.Candlestick{
						{Timestamp: tInt("2021-01-02 00:00:00"), OpenPrice: 1234, HighestPrice: 1234, ClosePrice: 1234, LowestPrice: 1234},
						{Timestamp: tInt("2021-01-03 00:00:00"), OpenPrice: 2345, HighestPrice: 2345, ClosePrice: 2345, LowestPrice: 2345},
					},
					expectedErr: nil,
				},
				{
					opType:              "GET",
					operand:             opETH,
					candlestickInterval: 24 * time.Hour,
					initialISO8601:      tpToISO("2021-01-02 00:00:00"),
					expectedErr:         ErrCacheMiss,
					expectedTicks:       []types.Candlestick{},
				},
				{
					opType:              "GET",
					operand:             opBTC,
					candlestickInterval: 24 * time.Hour,
					initialISO8601:      tpToISO("2021-01-02 00:00:00"),
					expectedErr:         nil,
					expectedTicks: []types.Candlestick{
						{Timestamp: tInt("2021-01-02 00:00:00"), OpenPrice: 1234, HighestPrice: 1234, ClosePrice: 1234, LowestPrice: 1234},
						{Timestamp: tInt("2021-01-03 00:00:00"), OpenPrice: 2345, HighestPrice: 2345, ClosePrice: 2345, LowestPrice: 2345},
					},
				},
			},
		},
		{
			name: "DAILY: A secondary PUT overrides the first one's values, with full overlap",
			ops: []operation{
				{
					opType:              "PUT",
					operand:             opBTC,
					candlestickInterval: 24 * time.Hour,
					candlesticks: []types.Candlestick{
						{Timestamp: tInt("2020-01-02 00:00:00"), OpenPrice: 1234, HighestPrice: 1234, ClosePrice: 1234, LowestPrice: 1234},
						{Timestamp: tInt("2020-01-03 00:00:00"), OpenPrice: 2345, HighestPrice: 2345, ClosePrice: 2345, LowestPrice: 2345},
					},
					expectedErr: nil,
				},
				{
					opType:              "PUT",
					operand:             opBTC,
					candlestickInterval: 24 * time.Hour,
					candlesticks: []types.Candlestick{
						{Timestamp: tInt("2020-01-02 00:00:00"), OpenPrice: 2345, HighestPrice: 2345, ClosePrice: 2345, LowestPrice: 2345},
						{Timestamp: tInt("2020-01-03 00:00:00"), OpenPrice: 3456, HighestPrice: 3456, ClosePrice: 3456, LowestPrice: 3456},
					},
					expectedErr: nil,
				},
				{
					opType:              "GET",
					operand:             opBTC,
					candlestickInterval: 24 * time.Hour,
					initialISO8601:      tpToISO("2020-01-02 00:00:00"),
					expectedErr:         nil,
					expectedTicks: []types.Candlestick{
						{Timestamp: tInt("2020-01-02 00:00:00"), OpenPrice: 2345, HighestPrice: 2345, ClosePrice: 2345, LowestPrice: 2345},
						{Timestamp: tInt("2020-01-03 00:00:00"), OpenPrice: 3456, HighestPrice: 3456, ClosePrice: 3456, LowestPrice: 3456},
					},
				},
			},
		},
		{
			name: "DAILY: A secondary PUT with overlap makes the sequence larger on GET",
			ops: []operation{
				{
					opType:              "PUT",
					operand:             opBTC,
					candlestickInterval: 24 * time.Hour,
					candlesticks: []types.Candlestick{
						{Timestamp: tInt("2020-01-02 00:00:00"), OpenPrice: 1234, HighestPrice: 1234, ClosePrice: 1234, LowestPrice: 1234},
						{Timestamp: tInt("2020-01-03 00:00:00"), OpenPrice: 2345, HighestPrice: 2345, ClosePrice: 2345, LowestPrice: 2345},
					},
					expectedErr: nil,
				},
				{
					opType:              "PUT",
					operand:             opBTC,
					candlestickInterval: 24 * time.Hour,
					candlesticks: []types.Candlestick{
						{Timestamp: tInt("2020-01-03 00:00:00"), OpenPrice: 2345, HighestPrice: 2345, ClosePrice: 2345, LowestPrice: 2345},
						{Timestamp: tInt("2020-01-04 00:00:00"), OpenPrice: 3456, HighestPrice: 3456, ClosePrice: 3456, LowestPrice: 3456},
					},
					expectedErr: nil,
				},
				{
					opType:              "GET",
					operand:             opBTC,
					candlestickInterval: 24 * time.Hour,
					initialISO8601:      tpToISO("2020-01-02 00:00:00"),
					expectedErr:         nil,
					expectedTicks: []types.Candlestick{
						{Timestamp: tInt("2020-01-02 00:00:00"), OpenPrice: 1234, HighestPrice: 1234, ClosePrice: 1234, LowestPrice: 1234},
						{Timestamp: tInt("2020-01-03 00:00:00"), OpenPrice: 2345, HighestPrice: 2345, ClosePrice: 2345, LowestPrice: 2345},
						{Timestamp: tInt("2020-01-04 00:00:00"), OpenPrice: 3456, HighestPrice: 3456, ClosePrice: 3456, LowestPrice: 3456},
					},
				},
			},
		},
		{
			name: "DAILY: A secondary PUT without overlap does not make the sequence larger on GET, and a second get gets the other one",
			ops: []operation{
				{
					opType:              "PUT",
					operand:             opBTC,
					candlestickInterval: 24 * time.Hour,
					candlesticks: []types.Candlestick{
						{Timestamp: tInt("2020-01-02 00:00:00"), OpenPrice: 1234, HighestPrice: 1234, ClosePrice: 1234, LowestPrice: 1234},
						{Timestamp: tInt("2020-01-03 00:00:00"), OpenPrice: 2345, HighestPrice: 2345, ClosePrice: 2345, LowestPrice: 2345},
					},
					expectedErr: nil,
				},
				{
					opType:              "PUT",
					operand:             opBTC,
					candlestickInterval: 24 * time.Hour,
					candlesticks: []types.Candlestick{
						{Timestamp: tInt("2020-01-05 00:00:00"), OpenPrice: 3456, HighestPrice: 3456, ClosePrice: 3456, LowestPrice: 3456},
						{Timestamp: tInt("2020-01-06 00:00:00"), OpenPrice: 4567, HighestPrice: 4567, ClosePrice: 4567, LowestPrice: 4567},
					},
					expectedErr: nil,
				},
				{
					opType:              "GET",
					operand:             opBTC,
					candlestickInterval: 24 * time.Hour,
					initialISO8601:      tpToISO("2020-01-02 00:00:00"),
					expectedErr:         nil,
					expectedTicks: []types.Candlestick{
						{Timestamp: tInt("2020-01-02 00:00:00"), OpenPrice: 1234, HighestPrice: 1234, ClosePrice: 1234, LowestPrice: 1234},
						{Timestamp: tInt("2020-01-03 00:00:00"), OpenPrice: 2345, HighestPrice: 2345, ClosePrice: 2345, LowestPrice: 2345},
					},
				},
				{
					opType:              "GET",
					operand:             opBTC,
					candlestickInterval: 24 * time.Hour,
					initialISO8601:      tpToISO("2020-01-05 00:00:00"),
					expectedErr:         nil,
					expectedTicks: []types.Candlestick{
						{Timestamp: tInt("2020-01-05 00:00:00"), OpenPrice: 3456, HighestPrice: 3456, ClosePrice: 3456, LowestPrice: 3456},
						{Timestamp: tInt("2020-01-06 00:00:00"), OpenPrice: 4567, HighestPrice: 4567, ClosePrice: 4567, LowestPrice: 4567},
					},
				},
			},
		},
		{
			name: "DAILY: Two separate series work at the same time",
			ops: []operation{
				{
					opType:              "PUT",
					operand:             opBTC,
					candlestickInterval: 24 * time.Hour,
					candlesticks: []types.Candlestick{
						{Timestamp: tInt("2020-01-02 00:00:00"), OpenPrice: 1234, HighestPrice: 1234, ClosePrice: 1234, LowestPrice: 1234},
						{Timestamp: tInt("2020-01-03 00:00:00"), OpenPrice: 2345, HighestPrice: 2345, ClosePrice: 2345, LowestPrice: 2345},
					},
					expectedErr: nil,
				},
				{
					opType:              "PUT",
					operand:             opETH,
					candlestickInterval: 24 * time.Hour,
					candlesticks: []types.Candlestick{
						{Timestamp: tInt("2020-01-02 00:00:00"), OpenPrice: 3456, HighestPrice: 3456, ClosePrice: 3456, LowestPrice: 3456},
						{Timestamp: tInt("2020-01-03 00:00:00"), OpenPrice: 4567, HighestPrice: 4567, ClosePrice: 4567, LowestPrice: 4567},
					},
					expectedErr: nil,
				},
				{
					opType:              "GET",
					operand:             opBTC,
					candlestickInterval: 24 * time.Hour,
					initialISO8601:      tpToISO("2020-01-02 00:00:00"),
					expectedErr:         nil,
					expectedTicks: []types.Candlestick{
						{Timestamp: tInt("2020-01-02 00:00:00"), OpenPrice: 1234, HighestPrice: 1234, ClosePrice: 1234, LowestPrice: 1234},
						{Timestamp: tInt("2020-01-03 00:00:00"), OpenPrice: 2345, HighestPrice: 2345, ClosePrice: 2345, LowestPrice: 2345},
					},
				},
				{
					opType:              "GET",
					operand:             opETH,
					candlestickInterval: 24 * time.Hour,
					initialISO8601:      tpToISO("2020-01-02 00:00:00"),
					expectedErr:         nil,
					expectedTicks: []types.Candlestick{
						{Timestamp: tInt("2020-01-02 00:00:00"), OpenPrice: 3456, HighestPrice: 3456, ClosePrice: 3456, LowestPrice: 3456},
						{Timestamp: tInt("2020-01-03 00:00:00"), OpenPrice: 4567, HighestPrice: 4567, ClosePrice: 4567, LowestPrice: 4567},
					},
				},
			},
		},
		{
			name: "DAILY: Get of a day on an empty time is a cache miss",
			ops: []operation{
				{
					opType:              "PUT",
					operand:             opBTC,
					candlestickInterval: 24 * time.Hour,
					candlesticks: []types.Candlestick{
						{Timestamp: tInt("2020-01-02 00:00:00"), OpenPrice: 1234, HighestPrice: 1234, ClosePrice: 1234, LowestPrice: 1234},
						{Timestamp: tInt("2020-01-03 00:00:00"), OpenPrice: 2345, HighestPrice: 2345, ClosePrice: 2345, LowestPrice: 2345},
					},
					expectedErr: nil,
				},
				{
					opType:              "GET",
					operand:             opBTC,
					candlestickInterval: 24 * time.Hour,
					initialISO8601:      tpToISO("2020-01-04 00:00:00"),
					expectedErr:         ErrCacheMiss,
					expectedTicks:       []types.Candlestick{},
				},
			},
		},
		{
			name: "DAILY: Get of a minute before, but with non-zero seconds, returns the tick of the next minute",
			ops: []operation{
				{
					opType:              "PUT",
					operand:             opBTC,
					candlestickInterval: 24 * time.Hour,
					candlesticks: []types.Candlestick{
						{Timestamp: tInt("2020-01-02 00:00:00"), OpenPrice: 1234, HighestPrice: 1234, ClosePrice: 1234, LowestPrice: 1234},
						{Timestamp: tInt("2020-01-03 00:00:00"), OpenPrice: 2345, HighestPrice: 2345, ClosePrice: 2345, LowestPrice: 2345},
					},
					expectedErr: nil,
				},
				{
					opType:              "GET",
					operand:             opBTC,
					candlestickInterval: 24 * time.Hour,
					initialISO8601:      tpToISO("2020-01-01 03:03:01"),
					expectedErr:         nil,
					expectedTicks: []types.Candlestick{
						{Timestamp: tInt("2020-01-02 00:00:00"), OpenPrice: 1234, HighestPrice: 1234, ClosePrice: 1234, LowestPrice: 1234},
						{Timestamp: tInt("2020-01-03 00:00:00"), OpenPrice: 2345, HighestPrice: 2345, ClosePrice: 2345, LowestPrice: 2345},
					},
				},
			},
		},
		{
			name: "DAILY: Putting ticks that span two intervals works, but requires two gets to get both ticks",
			ops: []operation{
				{
					opType:              "PUT",
					operand:             opBTC,
					candlestickInterval: 24 * time.Hour,
					candlesticks: []types.Candlestick{
						{Timestamp: tInt("2020-03-16 00:00:00"), OpenPrice: 1234, HighestPrice: 1234, ClosePrice: 1234, LowestPrice: 1234},
						{Timestamp: tInt("2020-03-17 00:00:00"), OpenPrice: 2345, HighestPrice: 2345, ClosePrice: 2345, LowestPrice: 2345},
					},
					expectedErr: nil,
				},
				{
					opType:              "GET",
					operand:             opBTC,
					candlestickInterval: 24 * time.Hour,
					initialISO8601:      tpToISO("2020-03-16 00:00:00"),
					expectedErr:         nil,
					expectedTicks: []types.Candlestick{
						{Timestamp: tInt("2020-03-16 00:00:00"), OpenPrice: 1234, HighestPrice: 1234, ClosePrice: 1234, LowestPrice: 1234},
					},
				},
				{
					opType:              "GET",
					operand:             opBTC,
					candlestickInterval: 24 * time.Hour,
					initialISO8601:      tpToISO("2020-03-17 00:00:00"),
					expectedErr:         nil,
					expectedTicks: []types.Candlestick{
						{Timestamp: tInt("2020-03-17 00:00:00"), OpenPrice: 2345, HighestPrice: 2345, ClosePrice: 2345, LowestPrice: 2345},
					},
				},
			},
		},
	}
	for _, ts := range tss {
		t.Run(ts.name, func(t *testing.T) {
			cache := NewMemoryCache(map[time.Duration]int{time.Minute: 128, 24 * time.Hour: 128})
			var (
				actualCandlesticks []types.Candlestick
				actualErr          error
			)

			for _, op := range ts.ops {
				metric := Metric{Name: op.operand.Str, CandlestickInterval: op.candlestickInterval}
				if op.opType == "GET" {
					actualCandlesticks, actualErr = cache.Get(metric, op.initialISO8601)
				} else if op.opType == "PUT" {
					actualErr = cache.Put(metric, op.candlesticks)
				}
				if actualErr != nil && op.expectedErr == nil {
					t.Logf("expected no error but had '%v'", actualErr)
					t.FailNow()
				}
				if actualErr == nil && op.expectedErr != nil {
					t.Logf("expected error '%v' but had no error", op.expectedErr)
					t.FailNow()
				}
				if op.expectedErr != nil && actualErr != nil && !errors.Is(actualErr, op.expectedErr) {
					t.Logf("expected error '%v' but had error '%v'", op.expectedErr, actualErr)
					t.FailNow()
				}
				if op.expectedErr == nil && op.opType == "GET" {
					require.Equal(t, op.expectedTicks, actualCandlesticks)
				}
			}
		})
	}
}

func tpToISO(s string) types.ISO8601 {
	t, _ := time.Parse("2006-01-02 15:04:05", s)
	return types.ISO8601(t.Format(time.RFC3339))
}

func tp(s string) time.Time {
	t, _ := time.Parse("2006-01-02 15:04:05", s)
	return t
}

func tInt(s string) int {
	return int(tp(s).Unix())
}

func TestDoesNotFailWhenCreatedWithZeroSize(t *testing.T) {
	NewMemoryCache(map[time.Duration]int{time.Minute: 0, 24 * time.Hour: 0})
}
