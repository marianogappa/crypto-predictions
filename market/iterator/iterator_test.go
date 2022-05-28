package iterator

import (
	"errors"
	"testing"
	"time"

	"github.com/marianogappa/predictions/market/cache"
	"github.com/marianogappa/predictions/types"
	"github.com/stretchr/testify/require"
)

type testSpec struct {
	name                             string
	operand                          types.Operand
	startISO8601                     types.ISO8601
	candlestickProvider              *testCandlestickProvider
	timeNowFunc                      func() time.Time
	startFromNext                    bool
	errCreatingIterator              error
	expectedCandlestickProviderCalls []call
	expectedCallResponses            []response
}

func TestIterator(t *testing.T) {
	opBTCUSDT := types.Operand{
		Type:       types.COIN,
		Provider:   "BINANCE",
		BaseAsset:  "BTC",
		QuoteAsset: "USDT",
	}
	opBTC := types.Operand{
		Type:      types.MARKETCAP,
		Provider:  "MESSARI",
		BaseAsset: "BTC",
	}
	tss := []testSpec{
		// Minutely tests
		{
			name:                             "Minutely: InvalidISO8601",
			operand:                          opBTCUSDT,
			startISO8601:                     types.ISO8601("Invalid!"),
			candlestickProvider:              newTestCandlestickProvider(nil),
			timeNowFunc:                      time.Now,
			startFromNext:                    false,
			errCreatingIterator:              cache.ErrInvalidISO8601,
			expectedCandlestickProviderCalls: nil,
			expectedCallResponses:            nil,
		},
		{
			name:                             "Minutely: ErrNoNewTicksYet because timestamp to request is too early (startFromNext == false)",
			operand:                          opBTCUSDT,
			startISO8601:                     types.ISO8601(tpToISO("2020-01-02 00:01:10")), // 49 secs to now
			candlestickProvider:              newTestCandlestickProvider(nil),
			timeNowFunc:                      func() time.Time { return tp("2020-01-02 00:01:59") },
			startFromNext:                    false,
			errCreatingIterator:              nil,
			expectedCandlestickProviderCalls: nil,
			expectedCallResponses:            []response{{tick: types.Tick{}, err: types.ErrNoNewTicksYet}},
		},
		{
			name:                             "Minutely: ErrNoNewTicksYet because timestamp to request is too early (startFromNext == true)",
			operand:                          opBTCUSDT,
			startISO8601:                     types.ISO8601(tpToISO("2020-01-02 00:01:10")),
			candlestickProvider:              newTestCandlestickProvider(nil),
			timeNowFunc:                      func() time.Time { return tp("2020-01-02 00:02:59") },
			startFromNext:                    true,
			errCreatingIterator:              nil,
			expectedCandlestickProviderCalls: nil,
			expectedCallResponses:            []response{{tick: types.Tick{}, err: types.ErrNoNewTicksYet}},
		},
		{
			name:         "Minutely: ErrOutOfCandlestics because candlestickProvider returned that",
			operand:      opBTCUSDT,
			startISO8601: types.ISO8601(tpToISO("2020-01-02 00:01:10")),
			candlestickProvider: newTestCandlestickProvider([]testCandlestickProviderResponse{
				{candlesticks: []types.Candlestick{}, err: types.ErrOutOfCandlesticks},
			}),
			timeNowFunc:                      func() time.Time { return tp("2020-01-03 00:00:00") },
			startFromNext:                    false,
			errCreatingIterator:              nil,
			expectedCandlestickProviderCalls: []call{{operand: opBTCUSDT, startTimeTs: tInt("2020-01-02 00:02:00")}},
			expectedCallResponses:            []response{{tick: types.Tick{}, err: types.ErrOutOfCandlesticks}},
		},
		{
			name:         "Minutely: ErrExchangeReturnedNoTicks because exchange returned old ticks",
			operand:      opBTCUSDT,
			startISO8601: types.ISO8601(tpToISO("2020-01-02 00:01:10")),
			candlestickProvider: newTestCandlestickProvider([]testCandlestickProviderResponse{
				{candlesticks: []types.Candlestick{
					{Timestamp: tInt("2020-01-02 00:00:00"), OpenPrice: 1234, HighestPrice: 1234, LowestPrice: 1234, ClosePrice: 1234},
					{Timestamp: tInt("2020-01-02 00:01:00"), OpenPrice: 1235, HighestPrice: 1235, LowestPrice: 1235, ClosePrice: 1235}}, err: nil},
			}),
			timeNowFunc:                      func() time.Time { return tp("2020-01-03 00:00:00") },
			startFromNext:                    false,
			errCreatingIterator:              nil,
			expectedCandlestickProviderCalls: []call{{operand: opBTCUSDT, startTimeTs: tInt("2020-01-02 00:02:00")}},
			expectedCallResponses:            []response{{tick: types.Tick{}, err: types.ErrExchangeReturnedNoTicks}},
		},
		{
			name:         "Minutely: ErrExchangeReturnedOutOfSyncTick because exchange returned ticks after requested one",
			operand:      opBTCUSDT,
			startISO8601: types.ISO8601(tpToISO("2020-01-02 00:01:10")),
			candlestickProvider: newTestCandlestickProvider([]testCandlestickProviderResponse{
				{candlesticks: []types.Candlestick{
					{Timestamp: tInt("2020-01-02 00:03:00"), OpenPrice: 1234, HighestPrice: 1234, LowestPrice: 1234, ClosePrice: 1234},
					{Timestamp: tInt("2020-01-02 00:04:00"), OpenPrice: 1235, HighestPrice: 1235, LowestPrice: 1235, ClosePrice: 1235}}, err: nil},
			}),
			timeNowFunc:                      func() time.Time { return tp("2020-01-03 00:00:00") },
			startFromNext:                    false,
			errCreatingIterator:              nil,
			expectedCandlestickProviderCalls: []call{{operand: opBTCUSDT, startTimeTs: tInt("2020-01-02 00:02:00")}},
			expectedCallResponses:            []response{{tick: types.Tick{}, err: types.ErrExchangeReturnedOutOfSyncTick}},
		},
		{
			name:         "Minutely: Succeeds to request one tick",
			operand:      opBTCUSDT,
			startISO8601: types.ISO8601(tpToISO("2020-01-02 00:01:10")),
			candlestickProvider: newTestCandlestickProvider([]testCandlestickProviderResponse{
				{candlesticks: []types.Candlestick{
					{Timestamp: tInt("2020-01-02 00:02:00"), OpenPrice: 1234, HighestPrice: 1234, LowestPrice: 1234, ClosePrice: 1234}}, err: nil},
			}),
			timeNowFunc:                      func() time.Time { return tp("2020-01-03 00:00:00") },
			startFromNext:                    false,
			errCreatingIterator:              nil,
			expectedCandlestickProviderCalls: []call{{operand: opBTCUSDT, startTimeTs: tInt("2020-01-02 00:02:00")}},
			expectedCallResponses:            []response{{tick: types.Tick{Timestamp: tInt("2020-01-02 00:02:00"), Value: 1234}, err: nil}},
		},
		{
			name:         "Minutely: Succeeds to request two ticks, making only one request",
			operand:      opBTCUSDT,
			startISO8601: types.ISO8601(tpToISO("2020-01-02 00:02:00")),
			candlestickProvider: newTestCandlestickProvider([]testCandlestickProviderResponse{
				{candlesticks: []types.Candlestick{
					{Timestamp: tInt("2020-01-02 00:02:00"), OpenPrice: 1234, HighestPrice: 1234, LowestPrice: 1234, ClosePrice: 1234},
					{Timestamp: tInt("2020-01-02 00:03:00"), OpenPrice: 1235, HighestPrice: 1235, LowestPrice: 1235, ClosePrice: 1235}}, err: nil},
			}),
			timeNowFunc:                      func() time.Time { return tp("2020-01-03 00:00:00") },
			startFromNext:                    false,
			errCreatingIterator:              nil,
			expectedCandlestickProviderCalls: []call{{operand: opBTCUSDT, startTimeTs: tInt("2020-01-02 00:02:00")}},
			expectedCallResponses: []response{
				{tick: types.Tick{Timestamp: tInt("2020-01-02 00:02:00"), Value: 1234}, err: nil},
				{tick: types.Tick{Timestamp: tInt("2020-01-02 00:03:00"), Value: 1235}, err: nil}},
		},
		{
			name:         "Minutely: Ignores cache Put error",
			operand:      opBTCUSDT,
			startISO8601: types.ISO8601(tpToISO("2020-01-02 00:02:00")),
			candlestickProvider: newTestCandlestickProvider([]testCandlestickProviderResponse{
				{candlesticks: []types.Candlestick{
					{Timestamp: tInt("2020-01-02 00:02:00"), OpenPrice: 0, HighestPrice: 0, LowestPrice: 0, ClosePrice: 0}}, err: nil},
			}),
			timeNowFunc:                      func() time.Time { return tp("2020-01-03 00:00:00") },
			startFromNext:                    false,
			errCreatingIterator:              nil,
			expectedCandlestickProviderCalls: []call{{operand: opBTCUSDT, startTimeTs: tInt("2020-01-02 00:02:00")}},
			expectedCallResponses: []response{
				{tick: types.Tick{Timestamp: tInt("2020-01-02 00:02:00"), Value: 0}, err: nil},
			},
		},
		// Daily tests
		{
			name:                             "Daily: InvalidISO8601",
			operand:                          opBTC,
			startISO8601:                     types.ISO8601("Invalid!"),
			candlestickProvider:              newTestCandlestickProvider(nil),
			timeNowFunc:                      time.Now,
			startFromNext:                    false,
			errCreatingIterator:              cache.ErrInvalidISO8601,
			expectedCandlestickProviderCalls: nil,
			expectedCallResponses:            nil,
		},
		{
			name:                             "Daily: ErrNoNewTicksYet because timestamp to request is too early (startFromNext == false)",
			operand:                          opBTC,
			startISO8601:                     types.ISO8601(tpToISO("2020-01-02 00:01:10")),
			candlestickProvider:              newTestCandlestickProvider(nil),
			timeNowFunc:                      func() time.Time { return tp("2020-01-02 23:59:59") },
			startFromNext:                    false,
			errCreatingIterator:              nil,
			expectedCandlestickProviderCalls: nil,
			expectedCallResponses:            []response{{tick: types.Tick{}, err: types.ErrNoNewTicksYet}},
		},
		{
			name:                             "Daily: ErrNoNewTicksYet because timestamp to request is too early (startFromNext == true)",
			operand:                          opBTC,
			startISO8601:                     types.ISO8601(tpToISO("2020-01-02 00:01:10")),
			candlestickProvider:              newTestCandlestickProvider(nil),
			timeNowFunc:                      func() time.Time { return tp("2020-01-03 00:02:59") },
			startFromNext:                    true,
			errCreatingIterator:              nil,
			expectedCandlestickProviderCalls: nil,
			expectedCallResponses:            []response{{tick: types.Tick{}, err: types.ErrNoNewTicksYet}},
		},
		{
			name:         "Daily: ErrExchangeReturnedNoTicks because exchange returned old ticks",
			operand:      opBTC,
			startISO8601: types.ISO8601(tpToISO("2020-01-02 00:01:10")),
			candlestickProvider: newTestCandlestickProvider([]testCandlestickProviderResponse{
				{candlesticks: []types.Candlestick{
					{Timestamp: tInt("2020-01-01 00:00:00"), OpenPrice: 1234, HighestPrice: 1234, LowestPrice: 1234, ClosePrice: 1234},
					{Timestamp: tInt("2020-01-02 00:00:00"), OpenPrice: 1235, HighestPrice: 1235, LowestPrice: 1235, ClosePrice: 1235}}, err: nil},
			}),
			timeNowFunc:                      func() time.Time { return tp("2020-01-04 00:00:00") },
			startFromNext:                    false,
			errCreatingIterator:              nil,
			expectedCandlestickProviderCalls: []call{{operand: opBTC, startTimeTs: tInt("2020-01-03 00:00:00")}},
			expectedCallResponses:            []response{{tick: types.Tick{}, err: types.ErrExchangeReturnedNoTicks}},
		},
		{
			name:         "Daily: ErrExchangeReturnedOutOfSyncTick because exchange returned ticks after requested one",
			operand:      opBTC,
			startISO8601: types.ISO8601(tpToISO("2020-01-02 00:01:10")),
			candlestickProvider: newTestCandlestickProvider([]testCandlestickProviderResponse{
				{candlesticks: []types.Candlestick{
					{Timestamp: tInt("2020-01-04 00:00:00"), OpenPrice: 1234, HighestPrice: 1234, LowestPrice: 1234, ClosePrice: 1234},
					{Timestamp: tInt("2020-01-05 00:00:00"), OpenPrice: 1235, HighestPrice: 1235, LowestPrice: 1235, ClosePrice: 1235}}, err: nil},
			}),
			timeNowFunc:                      func() time.Time { return tp("2020-01-05 00:00:00") },
			startFromNext:                    false,
			errCreatingIterator:              nil,
			expectedCandlestickProviderCalls: []call{{operand: opBTC, startTimeTs: tInt("2020-01-03 00:00:00")}},
			expectedCallResponses:            []response{{tick: types.Tick{}, err: types.ErrExchangeReturnedOutOfSyncTick}},
		},
		{
			name:         "Daily: Succeeds to request one tick",
			operand:      opBTC,
			startISO8601: types.ISO8601(tpToISO("2020-01-03 00:00:00")),
			candlestickProvider: newTestCandlestickProvider([]testCandlestickProviderResponse{
				{candlesticks: []types.Candlestick{
					{Timestamp: tInt("2020-01-03 00:00:00"), OpenPrice: 1234, HighestPrice: 1234, LowestPrice: 1234, ClosePrice: 1234}}, err: nil},
			}),
			timeNowFunc:                      func() time.Time { return tp("2020-01-04 00:00:00") },
			startFromNext:                    false,
			errCreatingIterator:              nil,
			expectedCandlestickProviderCalls: []call{{operand: opBTC, startTimeTs: tInt("2020-01-03 00:00:00")}},
			expectedCallResponses: []response{
				{tick: types.Tick{Timestamp: tInt("2020-01-03 00:00:00"), Value: 1234}, err: nil}},
		},
		{
			name:         "Daily: Succeeds to request two ticks, making only one request",
			operand:      opBTC,
			startISO8601: types.ISO8601(tpToISO("2020-01-02 00:01:10")),
			candlestickProvider: newTestCandlestickProvider([]testCandlestickProviderResponse{
				{candlesticks: []types.Candlestick{
					{Timestamp: tInt("2020-01-03 00:00:00"), OpenPrice: 1234, HighestPrice: 1234, LowestPrice: 1234, ClosePrice: 1234},
					{Timestamp: tInt("2020-01-04 00:00:00"), OpenPrice: 1235, HighestPrice: 1235, LowestPrice: 1235, ClosePrice: 1235}}, err: nil},
			}),
			timeNowFunc:                      func() time.Time { return tp("2020-01-05 00:00:00") },
			startFromNext:                    false,
			errCreatingIterator:              nil,
			expectedCandlestickProviderCalls: []call{{operand: opBTC, startTimeTs: tInt("2020-01-03 00:00:00")}},
			expectedCallResponses: []response{
				{tick: types.Tick{Timestamp: tInt("2020-01-03 00:00:00"), Value: 1234}, err: nil},
				{tick: types.Tick{Timestamp: tInt("2020-01-04 00:00:00"), Value: 1235}, err: nil}},
		},
	}
	for _, ts := range tss {
		t.Run(ts.name, func(t *testing.T) {
			cache := cache.NewMemoryCache(map[time.Duration]int{time.Minute: 128, 24 * time.Hour: 128})
			intervalMinutes := 1
			if ts.operand.Type == types.MARKETCAP {
				intervalMinutes = 60 * 24
			}
			iterator, err := NewIterator(ts.operand, ts.startISO8601, cache, ts.candlestickProvider, ts.timeNowFunc, ts.startFromNext, intervalMinutes)
			if err == nil && ts.errCreatingIterator != nil {
				t.Logf("expected error '%v' but had no error", ts.errCreatingIterator)
				t.FailNow()
			}
			if err != nil && ts.errCreatingIterator == nil {
				t.Logf("expected no error but had '%v'", err)
				t.FailNow()
			}
			if err != nil && !errors.Is(err, ts.errCreatingIterator) {
				t.Errorf("expected error %v but got %v", ts.errCreatingIterator, err)
				t.FailNow()
			}

			for _, expectedResp := range ts.expectedCallResponses {
				actualTick, actualErr := iterator.NextTick()
				if actualErr != nil && expectedResp.err == nil {
					t.Logf("expected no error but had '%v'", actualErr)
					t.FailNow()
				}
				if actualErr == nil && expectedResp.err != nil {
					t.Logf("expected error '%v' but had no error", actualErr)
					t.FailNow()
				}
				if expectedResp.err != nil && actualErr != nil && !errors.Is(actualErr, expectedResp.err) {
					t.Logf("expected error '%v' but had error '%v'", expectedResp.err, actualErr)
					t.FailNow()
				}
				if expectedResp.err == nil {
					require.Equal(t, expectedResp.tick, actualTick)
				}
			}

			require.Equal(t, ts.expectedCandlestickProviderCalls, ts.candlestickProvider.calls)
		})
	}
}

func TestTickIteratorUsesCache(t *testing.T) {
	opBTCUSDT := types.Operand{
		Type:       types.COIN,
		Provider:   "BINANCE",
		BaseAsset:  "BTC",
		QuoteAsset: "USDT",
	}
	cache := cache.NewMemoryCache(map[time.Duration]int{time.Minute: 128, 24 * time.Hour: 128})
	cstick1 := types.Candlestick{Timestamp: tInt("2020-01-02 00:00:00"), OpenPrice: 1234, HighestPrice: 1234, LowestPrice: 1234, ClosePrice: 1234}
	tick1 := types.Tick{Timestamp: tInt("2020-01-02 00:00:00"), Value: 1234}
	cstick2 := types.Candlestick{Timestamp: tInt("2020-01-02 00:01:00"), OpenPrice: 1234, HighestPrice: 1234, LowestPrice: 1234, ClosePrice: 1234}
	tick2 := types.Tick{Timestamp: tInt("2020-01-02 00:01:00"), Value: 1234}
	cstick3 := types.Candlestick{Timestamp: tInt("2020-01-02 00:02:00"), OpenPrice: 1234, HighestPrice: 1234, LowestPrice: 1234, ClosePrice: 1234}
	tick3 := types.Tick{Timestamp: tInt("2020-01-02 00:02:00"), Value: 1234}

	testCandlestickProvider1 := newTestCandlestickProvider([]testCandlestickProviderResponse{
		{candlesticks: []types.Candlestick{cstick1, cstick2, cstick3}, err: nil},
		{candlesticks: nil, err: types.ErrOutOfCandlesticks},
	})
	it1, _ := NewIterator(
		opBTCUSDT,
		tpToISO("2020-01-02 00:00:00"),
		cache,
		testCandlestickProvider1,
		func() time.Time { return tp("2022-01-03 00:00:00") },
		false,
		1,
	)
	tick, err := it1.NextTick()
	require.Nil(t, err)
	require.Equal(t, tick, tick1)
	tick, err = it1.NextTick()
	require.Nil(t, err)
	require.Equal(t, tick, tick2)
	tick, err = it1.NextTick()
	require.Nil(t, err)
	require.Equal(t, tick, tick3)
	_, err = it1.NextTick()
	require.Equal(t, types.ErrOutOfCandlesticks, err)

	require.Len(t, testCandlestickProvider1.calls, 2)

	testCandlestickProvider2 := newTestCandlestickProvider([]testCandlestickProviderResponse{{candlesticks: nil, err: types.ErrOutOfCandlesticks}})
	it2, _ := NewIterator(
		opBTCUSDT,
		tpToISO("2020-01-02 00:00:00"),
		cache, // Reusing cache, so cache should kick in and prevent testCandlestickProvider2 from being called
		testCandlestickProvider2,
		func() time.Time { return tp("2022-01-03 00:00:00") },
		false,
		1,
	)
	tick, err = it2.NextTick()
	require.Nil(t, err)
	require.Equal(t, tick, tick1)
	tick, err = it2.NextTick()
	require.Nil(t, err)
	require.Equal(t, tick, tick2)
	tick, err = it2.NextTick()
	require.Nil(t, err)
	require.Equal(t, tick, tick3)
	_, err = it2.NextTick()
	require.Equal(t, types.ErrOutOfCandlesticks, err)

	require.Len(t, testCandlestickProvider2.calls, 1) // Cache was used!! Only last call after cache consumed.
}

type response struct {
	tick types.Tick
	err  error
}

type testCandlestickProviderResponse struct {
	candlesticks []types.Candlestick
	err          error
}

type call struct {
	operand     types.Operand
	startTimeTs int
}

type testCandlestickProvider struct {
	calls     []call
	responses []testCandlestickProviderResponse
}

func newTestCandlestickProvider(responses []testCandlestickProviderResponse) *testCandlestickProvider {
	return &testCandlestickProvider{responses: responses}
}

func (p *testCandlestickProvider) RequestCandlesticks(operand types.Operand, startTimeTs, intervalMinutes int) ([]types.Candlestick, error) {
	resp := p.responses[len(p.calls)]
	p.calls = append(p.calls, call{operand: operand, startTimeTs: startTimeTs})
	return resp.candlesticks, resp.err
}

func (p *testCandlestickProvider) GetPatience() time.Duration { return 0 * time.Second }

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
