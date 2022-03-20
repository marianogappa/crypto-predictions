package messari

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/marianogappa/predictions/types"
)

func (m *Messari) overrideAPIURL(url string) {
	m.apiURL = url
}

func TestHappyToTicks(t *testing.T) {
	testTicks := `{
		"status": {
		  "elapsed": 69,
		  "timestamp": "2022-02-19T13:30:47.726761767Z"
		},
		"data": {
		  "parameters": {
			"asset_key": "btc",
			"asset_id": "1e31218a-e44e-4285-820c-8282ee222035",
			"start": "2022-02-18T00:00:00Z",
			"end": "2022-02-19T13:30:47.659408521Z",
			"interval": "1d",
			"order": "ascending",
			"format": "json",
			"timestamp_format": "unix-milliseconds",
			"columns": [
			  "timestamp",
			  "marketcap_oustanding"
			]
		  },
		  "schema": {
			"metric_id": "mcap.out",
			"name": "Outstanding Marketcap",
			"description": "The sum USD value of the current supply. Also referred to as network value or market capitalization.",
			"values_schema": {
			  "timestamp": "Timestamp in milliseconds since the epoch (1 January 1970 00:00:00 UTC)",
			  "marketcap_oustanding": "The sum USD value of the current supply. Also referred to as network value or market capitalization."
			},
			"minimum_interval": "1d",
			"source_attribution": [
			  {
				"name": "Coinmetrics",
				"url": "https://coinmetrics.io"
			  }
			]
		  },
		  "values": [
			[
				1599782400000,
				192017369942.6865
			],
			[
				1599868800000,
				192951979972.30695
			],
			[
				1599955200000,
				190846477252.03787
			]
		  ]
		}
	  }`

	sr := response{}
	err := json.Unmarshal([]byte(testTicks), &sr)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	ts, err := sr.toTicks()
	if err != nil {
		t.Fatalf("toTicks() should have converted successfully but returned: %v", err)
	}
	if len(ts) != 3 {
		t.Fatalf("Should have converted 3 ticks but converted: %v", len(ts))
	}
	expected := []types.Tick{
		{
			Timestamp: 1599782400,
			Value:     192017369942.6865,
		},
		{
			Timestamp: 1599868800,
			Value:     192951979972.30695,
		},
		{
			Timestamp: 1599955200,
			Value:     190846477252.03787,
		},
	}
	if !reflect.DeepEqual(ts, expected) {
		t.Fatalf("Ticks should have been %v but were %v", expected, ts)
	}
}

func TestUnhappyToCandlesticks(t *testing.T) {
	tests := []string{
		// tick %v has len != 3! Invalid syntax from Messari
		`{"data":{"values":[
			[
				1499040000000
			]
		]}}`,
		// tick %v has non-int timestamp! Invalid syntax from Messari
		`{"data":{"values":[
			[
				"1499040000000",
				0.01634790
			]
		]}}`,
		// tick %v has non-float open! Invalid syntax from Messari
		`{"data":{"values":[
			[
				1499040000000,
				"0.01634790"
			]
		]}}`,
	}

	for i, ts := range tests {
		t.Run(fmt.Sprintf("Unhappy toTicks %v", i), func(t *testing.T) {
			sr := response{}
			err := json.Unmarshal([]byte(ts), &sr)
			if err != nil {
				t.Fatalf("Unmarshal failed: %v", err)
			}

			cs, err := sr.toTicks()
			if err == nil {
				t.Fatalf("Tick should have failed to convert but converted successfully to: %v", cs)
			}
		})
	}
}

func TestTicksInvalidUrl(t *testing.T) {
	i := 0
	replies := []string{
		`{"data":{"values": [
				[
					1599782400000,
					192017369942.6865
				]
			  ]}}`,
	}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, replies[i%len(replies)])
		i++
	}))
	defer ts.Close()

	b := NewMessari()
	b.overrideAPIURL("invalid url")
	_, err := b.RequestTicks(opBTC, tInt("2021-07-04T14:14:18+00:00"))
	if err == nil {
		t.Fatalf("should have failed due to invalid url")
	}
}

func TestTicksErrReadingResponseBody(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "1")
	}))
	defer ts.Close()

	b := NewMessari()
	b.overrideAPIURL(ts.URL + "/")
	b.SetDebug(true)
	_, err := b.RequestTicks(opBTC, tInt("2021-07-04T14:14:18+00:00"))
	if err == nil {
		t.Fatalf("should have failed due to invalid response body")
	}
}

func TestTicksErrorResponse(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, `{"status": {"elapsed": 10, "timestamp": "2022-02-19T13:30:24.796422841Z", "error_code": 400, "error_message": "\"unix-second\" is not a valid format. Valid formats are \"json\" or \"csv\""} }`)
	}))
	defer ts.Close()

	b := NewMessari()
	b.overrideAPIURL(ts.URL + "/")
	_, err := b.RequestTicks(opBTC, tInt("2021-07-04T14:14:18+00:00"))
	if err == nil {
		t.Fatalf("should have failed due to error response")
	}
}
func TestTicksInvalidJSONResponse(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, `invalid json`)
	}))
	defer ts.Close()

	b := NewMessari()
	b.overrideAPIURL(ts.URL + "/")
	_, err := b.RequestTicks(opBTC, tInt("2021-07-04T14:14:18+00:00"))
	if err == nil {
		t.Fatalf("should have failed due to invalid json")
	}
}

func TestTicksInvalidFloatsInJSONResponse(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, `{"data":{"values":[
			[
			1499040000000,
			"invalid"
			]
		]}}`)
	}))
	defer ts.Close()

	b := NewMessari()
	b.overrideAPIURL(ts.URL + "/")
	_, err := b.RequestTicks(opBTC, tInt("2021-07-04T14:14:18+00:00"))
	if err == nil {
		t.Fatalf("should have failed due to invalid floats in json")
	}
}

func tp(s string) time.Time {
	t, _ := time.Parse(time.RFC3339, s)
	return t
}

func tInt(s string) int {
	return int(tp(s).Unix())
}

var (
	opBTC = types.Operand{
		Type:      types.COIN,
		Provider:  "MESSARI",
		BaseAsset: "BTC",
		Str:       "MESSARI:BTC:",
	}
)
