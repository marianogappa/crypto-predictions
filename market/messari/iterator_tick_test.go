package messari

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/marianogappa/predictions/types"
)

type expected struct {
	tick types.Tick
	err  error
}

func TestTicks(t *testing.T) {
	i := 0
	replies := []string{
		`{"data":{"values": [
			[
				1599782400000,
				192017369942.6865
			],
			[
				1599868800000,
				192951979972.30695
			]
		]}}`,
		`{"data":{"values": [
			[
				1599955200000,
				190846477252.03787
			]
		]}}`,
		`{"data":{"values": null}}`,
	}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, replies[i%len(replies)])
		i++
	}))
	defer ts.Close()

	b := NewMessari()
	b.overrideAPIURL(ts.URL + "/")
	b.SetDebug(true)
	ti := b.BuildTickIterator("BTC", "mcap.out", "2020-09-11T00:00:00+00:00")

	expectedResults := []expected{
		{
			tick: types.Tick{
				Timestamp: 1599782400,
				Value:     192017369942.6865,
			},
			err: nil,
		},
		{
			tick: types.Tick{
				Timestamp: 1599868800,
				Value:     192951979972.30695,
			},
			err: nil,
		},
		{
			tick: types.Tick{
				Timestamp: 1599955200,
				Value:     190846477252.03787,
			},
			err: nil,
		},
		{
			tick: types.Tick{},
			err:  types.ErrOutOfTicks,
		},
	}
	for i, expectedResult := range expectedResults {
		actualTick, actualErr := ti.Next()
		if actualTick != expectedResult.tick {
			t.Errorf("on tick %v expected %v but got %v", i, expectedResult.tick, actualTick)
			t.FailNow()
		}
		if actualErr != expectedResult.err {
			t.Errorf("on tick %v expected no errors but this error happened %v", i, actualErr)
			t.FailNow()
		}
	}
}

func TestOldTicksNotReturned(t *testing.T) {
	i := 0
	replies := []string{
		`{"data":{"values": [
			[
				1599782400000,
				192017369942.6865
			],
			[
				1599868800000,
				192951979972.30695
			]
		]}}`,
		`{"data":{"values": null}}`,
	}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, replies[i%len(replies)])
		i++
	}))
	defer ts.Close()

	b := NewMessari()
	b.overrideAPIURL(ts.URL + "/")
	ti := b.BuildTickIterator("BTC", "mcap.out", "2021-09-11T00:00:00+00:00")

	expectedResults := []expected{
		{
			tick: types.Tick{},
			err:  types.ErrOutOfTicks,
		},
	}
	for i, expectedResult := range expectedResults {
		actualTick, actualErr := ti.Next()
		if actualTick != expectedResult.tick {
			t.Errorf("on tick %v expected %v but got %v", i, expectedResult.tick, actualTick)
			t.FailNow()
		}
		if actualErr != expectedResult.err {
			t.Errorf("on tick %v expected no errors but this error happened %v", i, actualErr)
			t.FailNow()
		}
	}
}
