package tickiterator

import (
	"errors"
	"testing"
	"time"

	"github.com/marianogappa/predictions/market/cache"
	"github.com/marianogappa/predictions/types"
	"github.com/stretchr/testify/require"
)

type testSpec struct {
	name                      string
	operand                   types.Operand
	startISO8601              types.ISO8601
	tickProvider              *testTickProvider
	timeNowFunc               func() time.Time
	startFromNext             bool
	errCreatingIterator       error
	expectedTickProviderCalls []call
	expectedCallResponses     []response
}

func TestTickIterator(t *testing.T) {
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
			name:                      "Minutely: InvalidISO8601",
			operand:                   opBTCUSDT,
			startISO8601:              types.ISO8601("Invalid!"),
			tickProvider:              newTestTickProvider(nil),
			timeNowFunc:               time.Now,
			startFromNext:             false,
			errCreatingIterator:       cache.ErrInvalidISO8601,
			expectedTickProviderCalls: nil,
			expectedCallResponses:     nil,
		},
		{
			name:                      "Minutely: ErrNoNewTicksYet because timestamp to request is too early (startFromNext == false)",
			operand:                   opBTCUSDT,
			startISO8601:              types.ISO8601(tpToISO("2020-01-02 00:01:10")), // 49 secs to now
			tickProvider:              newTestTickProvider(nil),
			timeNowFunc:               func() time.Time { return tp("2020-01-02 00:01:59") },
			startFromNext:             false,
			errCreatingIterator:       nil,
			expectedTickProviderCalls: nil,
			expectedCallResponses:     []response{{tick: types.Tick{}, err: types.ErrNoNewTicksYet}},
		},
		{
			name:                      "Minutely: ErrNoNewTicksYet because timestamp to request is too early (startFromNext == true)",
			operand:                   opBTCUSDT,
			startISO8601:              types.ISO8601(tpToISO("2020-01-02 00:01:10")),
			tickProvider:              newTestTickProvider(nil),
			timeNowFunc:               func() time.Time { return tp("2020-01-02 00:02:59") },
			startFromNext:             true,
			errCreatingIterator:       nil,
			expectedTickProviderCalls: nil,
			expectedCallResponses:     []response{{tick: types.Tick{}, err: types.ErrNoNewTicksYet}},
		},
		{
			name:         "Minutely: ErrOutOfCandlestics because tickProvider returned that",
			operand:      opBTCUSDT,
			startISO8601: types.ISO8601(tpToISO("2020-01-02 00:01:10")),
			tickProvider: newTestTickProvider([]testTickProviderResponse{
				{ticks: []types.Tick{}, err: types.ErrOutOfCandlesticks},
			}),
			timeNowFunc:               func() time.Time { return tp("2020-01-03 00:00:00") },
			startFromNext:             false,
			errCreatingIterator:       nil,
			expectedTickProviderCalls: []call{{operand: opBTCUSDT, startTimeTs: tInt("2020-01-02 00:02:00")}},
			expectedCallResponses:     []response{{tick: types.Tick{}, err: types.ErrOutOfCandlesticks}},
		},
		{
			name:         "Minutely: ErrExchangeReturnedNoTicks because exchange returned old ticks",
			operand:      opBTCUSDT,
			startISO8601: types.ISO8601(tpToISO("2020-01-02 00:01:10")),
			tickProvider: newTestTickProvider([]testTickProviderResponse{
				{ticks: []types.Tick{{Timestamp: tInt("2020-01-02 00:00:00"), Value: 1234}, {Timestamp: tInt("2020-01-02 00:01:00"), Value: 1235}}, err: nil},
			}),
			timeNowFunc:               func() time.Time { return tp("2020-01-03 00:00:00") },
			startFromNext:             false,
			errCreatingIterator:       nil,
			expectedTickProviderCalls: []call{{operand: opBTCUSDT, startTimeTs: tInt("2020-01-02 00:02:00")}},
			expectedCallResponses:     []response{{tick: types.Tick{}, err: types.ErrExchangeReturnedNoTicks}},
		},
		{
			name:         "Minutely: ErrExchangeReturnedOutOfSyncTick because exchange returned ticks after requested one",
			operand:      opBTCUSDT,
			startISO8601: types.ISO8601(tpToISO("2020-01-02 00:01:10")),
			tickProvider: newTestTickProvider([]testTickProviderResponse{
				{ticks: []types.Tick{{Timestamp: tInt("2020-01-02 00:03:00"), Value: 1234}, {Timestamp: tInt("2020-01-02 00:04:00"), Value: 1235}}, err: nil},
			}),
			timeNowFunc:               func() time.Time { return tp("2020-01-03 00:00:00") },
			startFromNext:             false,
			errCreatingIterator:       nil,
			expectedTickProviderCalls: []call{{operand: opBTCUSDT, startTimeTs: tInt("2020-01-02 00:02:00")}},
			expectedCallResponses:     []response{{tick: types.Tick{}, err: types.ErrExchangeReturnedOutOfSyncTick}},
		},
		{
			name:         "Minutely: Succeeds to request one tick",
			operand:      opBTCUSDT,
			startISO8601: types.ISO8601(tpToISO("2020-01-02 00:01:10")),
			tickProvider: newTestTickProvider([]testTickProviderResponse{
				{ticks: []types.Tick{{Timestamp: tInt("2020-01-02 00:02:00"), Value: 1234}}, err: nil},
			}),
			timeNowFunc:               func() time.Time { return tp("2020-01-03 00:00:00") },
			startFromNext:             false,
			errCreatingIterator:       nil,
			expectedTickProviderCalls: []call{{operand: opBTCUSDT, startTimeTs: tInt("2020-01-02 00:02:00")}},
			expectedCallResponses:     []response{{tick: types.Tick{Timestamp: tInt("2020-01-02 00:02:00"), Value: 1234}, err: nil}},
		},
		{
			name:         "Minutely: Succeeds to request two ticks, making only one request",
			operand:      opBTCUSDT,
			startISO8601: types.ISO8601(tpToISO("2020-01-02 00:02:00")),
			tickProvider: newTestTickProvider([]testTickProviderResponse{
				{ticks: []types.Tick{{Timestamp: tInt("2020-01-02 00:02:00"), Value: 1234}, {Timestamp: tInt("2020-01-02 00:03:00"), Value: 1235}}, err: nil},
			}),
			timeNowFunc:               func() time.Time { return tp("2020-01-03 00:00:00") },
			startFromNext:             false,
			errCreatingIterator:       nil,
			expectedTickProviderCalls: []call{{operand: opBTCUSDT, startTimeTs: tInt("2020-01-02 00:02:00")}},
			expectedCallResponses:     []response{{tick: types.Tick{Timestamp: tInt("2020-01-02 00:02:00"), Value: 1234}, err: nil}, {tick: types.Tick{Timestamp: tInt("2020-01-02 00:03:00"), Value: 1235}, err: nil}},
		},
		{
			name:         "Minutely: Ignores cache Put error",
			operand:      opBTCUSDT,
			startISO8601: types.ISO8601(tpToISO("2020-01-02 00:02:00")),
			tickProvider: newTestTickProvider([]testTickProviderResponse{
				{ticks: []types.Tick{{Timestamp: tInt("2020-01-02 00:02:00"), Value: 0}}, err: nil},
			}),
			timeNowFunc:               func() time.Time { return tp("2020-01-03 00:00:00") },
			startFromNext:             false,
			errCreatingIterator:       nil,
			expectedTickProviderCalls: []call{{operand: opBTCUSDT, startTimeTs: tInt("2020-01-02 00:02:00")}},
			expectedCallResponses:     []response{{tick: types.Tick{Timestamp: tInt("2020-01-02 00:02:00"), Value: 0}, err: nil}},
		},
		// Daily tests
		{
			name:                      "Daily: InvalidISO8601",
			operand:                   opBTC,
			startISO8601:              types.ISO8601("Invalid!"),
			tickProvider:              newTestTickProvider(nil),
			timeNowFunc:               time.Now,
			startFromNext:             false,
			errCreatingIterator:       cache.ErrInvalidISO8601,
			expectedTickProviderCalls: nil,
			expectedCallResponses:     nil,
		},
		{
			name:                      "Daily: ErrNoNewTicksYet because timestamp to request is too early (startFromNext == false)",
			operand:                   opBTC,
			startISO8601:              types.ISO8601(tpToISO("2020-01-02 00:01:10")),
			tickProvider:              newTestTickProvider(nil),
			timeNowFunc:               func() time.Time { return tp("2020-01-02 23:59:59") },
			startFromNext:             false,
			errCreatingIterator:       nil,
			expectedTickProviderCalls: nil,
			expectedCallResponses:     []response{{tick: types.Tick{}, err: types.ErrNoNewTicksYet}},
		},
		{
			name:                      "Daily: ErrNoNewTicksYet because timestamp to request is too early (startFromNext == true)",
			operand:                   opBTC,
			startISO8601:              types.ISO8601(tpToISO("2020-01-02 00:01:10")),
			tickProvider:              newTestTickProvider(nil),
			timeNowFunc:               func() time.Time { return tp("2020-01-03 00:02:59") },
			startFromNext:             true,
			errCreatingIterator:       nil,
			expectedTickProviderCalls: nil,
			expectedCallResponses:     []response{{tick: types.Tick{}, err: types.ErrNoNewTicksYet}},
		},
		{
			name:         "Daily: ErrExchangeReturnedNoTicks because exchange returned old ticks",
			operand:      opBTC,
			startISO8601: types.ISO8601(tpToISO("2020-01-02 00:01:10")),
			tickProvider: newTestTickProvider([]testTickProviderResponse{
				{ticks: []types.Tick{{Timestamp: tInt("2020-01-01 00:00:00"), Value: 1234}, {Timestamp: tInt("2020-01-02 00:00:00"), Value: 1235}}, err: nil},
			}),
			timeNowFunc:               func() time.Time { return tp("2020-01-04 00:00:00") },
			startFromNext:             false,
			errCreatingIterator:       nil,
			expectedTickProviderCalls: []call{{operand: opBTC, startTimeTs: tInt("2020-01-03 00:00:00")}},
			expectedCallResponses:     []response{{tick: types.Tick{}, err: types.ErrExchangeReturnedNoTicks}},
		},
		{
			name:         "Daily: ErrExchangeReturnedOutOfSyncTick because exchange returned ticks after requested one",
			operand:      opBTC,
			startISO8601: types.ISO8601(tpToISO("2020-01-02 00:01:10")),
			tickProvider: newTestTickProvider([]testTickProviderResponse{
				{ticks: []types.Tick{{Timestamp: tInt("2020-01-04 00:00:00"), Value: 1234}, {Timestamp: tInt("2020-01-05 00:00:00"), Value: 1235}}, err: nil},
			}),
			timeNowFunc:               func() time.Time { return tp("2020-01-05 00:00:00") },
			startFromNext:             false,
			errCreatingIterator:       nil,
			expectedTickProviderCalls: []call{{operand: opBTC, startTimeTs: tInt("2020-01-03 00:00:00")}},
			expectedCallResponses:     []response{{tick: types.Tick{}, err: types.ErrExchangeReturnedOutOfSyncTick}},
		},
		{
			name:         "Daily: Succeeds to request one tick",
			operand:      opBTC,
			startISO8601: types.ISO8601(tpToISO("2020-01-03 00:00:00")),
			tickProvider: newTestTickProvider([]testTickProviderResponse{
				{ticks: []types.Tick{{Timestamp: tInt("2020-01-03 00:00:00"), Value: 1234}}, err: nil},
			}),
			timeNowFunc:               func() time.Time { return tp("2020-01-04 00:00:00") },
			startFromNext:             false,
			errCreatingIterator:       nil,
			expectedTickProviderCalls: []call{{operand: opBTC, startTimeTs: tInt("2020-01-03 00:00:00")}},
			expectedCallResponses:     []response{{tick: types.Tick{Timestamp: tInt("2020-01-03 00:00:00"), Value: 1234}, err: nil}},
		},
		{
			name:         "Daily: Succeeds to request two ticks, making only one request",
			operand:      opBTC,
			startISO8601: types.ISO8601(tpToISO("2020-01-02 00:01:10")),
			tickProvider: newTestTickProvider([]testTickProviderResponse{
				{ticks: []types.Tick{{Timestamp: tInt("2020-01-03 00:00:00"), Value: 1234}, {Timestamp: tInt("2020-01-04 00:00:00"), Value: 1235}}, err: nil},
			}),
			timeNowFunc:               func() time.Time { return tp("2020-01-05 00:00:00") },
			startFromNext:             false,
			errCreatingIterator:       nil,
			expectedTickProviderCalls: []call{{operand: opBTC, startTimeTs: tInt("2020-01-03 00:00:00")}},
			expectedCallResponses:     []response{{tick: types.Tick{Timestamp: tInt("2020-01-03 00:00:00"), Value: 1234}, err: nil}, {tick: types.Tick{Timestamp: tInt("2020-01-04 00:00:00"), Value: 1235}, err: nil}},
		},
	}
	for _, ts := range tss {
		t.Run(ts.name, func(t *testing.T) {
			cache := cache.NewMemoryCache(128, 128)
			iterator, err := NewTickIterator(ts.operand, ts.startISO8601, cache, ts.tickProvider, ts.timeNowFunc, ts.startFromNext)
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
				actualTick, actualErr := iterator.Next()
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

			require.Equal(t, ts.expectedTickProviderCalls, ts.tickProvider.calls)
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
	cache := cache.NewMemoryCache(128, 128)
	tick1 := types.Tick{Timestamp: tInt("2020-01-02 00:00:00"), Value: 1234}
	tick2 := types.Tick{Timestamp: tInt("2020-01-02 00:01:00"), Value: 1234}
	tick3 := types.Tick{Timestamp: tInt("2020-01-02 00:02:00"), Value: 1234}

	testTickProvider1 := newTestTickProvider([]testTickProviderResponse{
		{ticks: []types.Tick{tick1, tick2, tick3}, err: nil},
		{ticks: nil, err: types.ErrOutOfCandlesticks},
	})
	it1, _ := NewTickIterator(
		opBTCUSDT,
		tpToISO("2020-01-02 00:00:00"),
		cache,
		testTickProvider1,
		func() time.Time { return tp("2022-01-03 00:00:00") },
		false,
	)
	tick, err := it1.Next()
	require.Nil(t, err)
	require.Equal(t, tick, tick1)
	tick, err = it1.Next()
	require.Nil(t, err)
	require.Equal(t, tick, tick2)
	tick, err = it1.Next()
	require.Nil(t, err)
	require.Equal(t, tick, tick3)
	tick, err = it1.Next()
	require.Equal(t, types.ErrOutOfCandlesticks, err)

	require.Len(t, testTickProvider1.calls, 2)

	testTickProvider2 := newTestTickProvider([]testTickProviderResponse{{ticks: nil, err: types.ErrOutOfCandlesticks}})
	it2, _ := NewTickIterator(
		opBTCUSDT,
		tpToISO("2020-01-02 00:00:00"),
		cache, // Reusing cache, so cache should kick in and prevent testTickProvider2 from being called
		testTickProvider2,
		func() time.Time { return tp("2022-01-03 00:00:00") },
		false,
	)
	tick, err = it2.Next()
	require.Nil(t, err)
	require.Equal(t, tick, tick1)
	tick, err = it2.Next()
	require.Nil(t, err)
	require.Equal(t, tick, tick2)
	tick, err = it2.Next()
	require.Nil(t, err)
	require.Equal(t, tick, tick3)
	_, err = it2.Next()
	require.Equal(t, types.ErrOutOfCandlesticks, err)

	require.Len(t, testTickProvider2.calls, 1) // Cache was used!! Only last call after cache consumed.
}

type response struct {
	tick types.Tick
	err  error
}

type testTickProviderResponse struct {
	ticks []types.Tick
	err   error
}

type call struct {
	operand     types.Operand
	startTimeTs int
}

type testTickProvider struct {
	calls     []call
	responses []testTickProviderResponse
}

func newTestTickProvider(responses []testTickProviderResponse) *testTickProvider {
	return &testTickProvider{responses: responses}
}

func (p *testTickProvider) RequestTicks(operand types.Operand, startTimeTs int) ([]types.Tick, error) {
	resp := p.responses[len(p.calls)]
	p.calls = append(p.calls, call{operand: operand, startTimeTs: startTimeTs})
	return resp.ticks, resp.err
}

func (p *testTickProvider) GetPatience() time.Duration { return 0 * time.Second }

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
