package kraken

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/marianogappa/predictions/types"
)

func TestHappyToCandlesticks(t *testing.T) {
	testCandlestick := `{"error":[],"result":{"XBTUSDT":[[1625623260,"34221.6","34221.6","34215.7","34215.7","34215.7","0.25998804",7]],"last":1626869340}}`

	sr := response{}
	err := json.Unmarshal([]byte(testCandlestick), &sr)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	cs, err := sr.toCandlesticks()
	if err != nil {
		t.Fatalf("Candlestick should have converted successfully but returned: %v", err)
	}
	if len(cs) != 1 {
		t.Fatalf("Should have converted 1 candlesticks but converted: %v", len(cs))
	}
	expected := types.Candlestick{
		Timestamp:      1625623260,
		OpenPrice:      f(34221.6),
		ClosePrice:     f(34215.7),
		LowestPrice:    f(34215.7),
		HighestPrice:   f(34221.6),
		Volume:         f(0.25998804),
		NumberOfTrades: 7,
	}
	if cs[0] != expected {
		t.Fatalf("Candlestick should have been %v but was %v", expected, cs[0])
	}
}

func TestUnhappyToCandlesticks(t *testing.T) {
	tests := []string{
		// data key [%v] did not contain an array of datapoints
		`{"error":[],"result":{"XBTUSDT":"INVALID","last":1626869340}}`,
		// candlestick [%v] did not contain an array of data fields, instead: [%v]
		`{"error":[],"result":{"XBTUSDT":["INVALID"],"last":1626869340}}`,
		// candlestick %v has non-int open time! Invalid syntax from Kraken
		`{"error":[],"result":{"XBTUSDT":[["INVALID","34221.6","34221.6","34215.7","34215.7","34215.7","0.25998804",7]],"last":1626869340}}`,
		// candlestick %v has non-string open! Invalid syntax from Kraken
		`{"error":[],"result":{"XBTUSDT":[[1625623260,34221.6,"34221.6","34215.7","34215.7","34215.7","0.25998804",7]],"last":1626869340}}`,
		// candlestick %v had open = %v! Invalid syntax from Kraken
		`{"error":[],"result":{"XBTUSDT":[[1625623260,"INVALID","34221.6","34215.7","34215.7","34215.7","0.25998804",7]],"last":1626869340}}`,
		// candlestick %v has non-string high! Invalid syntax from Kraken
		`{"error":[],"result":{"XBTUSDT":[[1625623260,"34221.6",34221.6,"34215.7","34215.7","34215.7","0.25998804",7]],"last":1626869340}}`,
		// candlestick %v had high = %v! Invalid syntax from Kraken
		`{"error":[],"result":{"XBTUSDT":[[1625623260,"34221.6","INVALID","34215.7","34215.7","34215.7","0.25998804",7]],"last":1626869340}}`,
		// candlestick %v has non-string low! Invalid syntax from Kraken
		`{"error":[],"result":{"XBTUSDT":[[1625623260,"34221.6","34221.6",34215.7,"34215.7","34215.7","0.25998804",7]],"last":1626869340}}`,
		// candlestick %v had low = %v! Invalid syntax from Kraken
		`{"error":[],"result":{"XBTUSDT":[[1625623260,"34221.6","34221.6","INVALID","34215.7","34215.7","0.25998804",7]],"last":1626869340}}`,
		// candlestick %v has non-string close! Invalid syntax from Kraken
		`{"error":[],"result":{"XBTUSDT":[[1625623260,"34221.6","34221.6","34215.7",34215.7,"34215.7","0.25998804",7]],"last":1626869340}}`,
		// candlestick %v had close = %v! Invalid syntax from Kraken
		`{"error":[],"result":{"XBTUSDT":[[1625623260,"34221.6","34221.6","34215.7","INVALID","34215.7","0.25998804",7]],"last":1626869340}}`,
		// candlestick %v has non-string vwap! Invalid syntax from Kraken
		`{"error":[],"result":{"XBTUSDT":[[1625623260,"34221.6","34221.6","34215.7","34215.7",34215.7,"0.25998804",7]],"last":1626869340}}`,
		// candlestick %v had vwap = %v! Invalid syntax from Kraken
		`{"error":[],"result":{"XBTUSDT":[[1625623260,"34221.6","34221.6","34215.7","34215.7","INVALID","0.25998804",7]],"last":1626869340}}`,
		// candlestick %v has non-string volume! Invalid syntax from Kraken
		`{"error":[],"result":{"XBTUSDT":[[1625623260,"34221.6","34221.6","34215.7","34215.7","34215.7",0.25998804,7]],"last":1626869340}}`,
		// candlestick %v had volume = %v! Invalid syntax from Kraken
		`{"error":[],"result":{"XBTUSDT":[[1625623260,"34221.6","34221.6","34215.7","34215.7","34215.7","INVALID",7]],"last":1626869340}}`,
		// candlestick %v has non-int trade count! Invalid syntax from Kraken
		`{"error":[],"result":{"XBTUSDT":[[1625623260,"34221.6","34221.6","34215.7","34215.7","34215.7","0.25998804","7"]],"last":1626869340}}`,
	}

	for i, ts := range tests {
		t.Run(fmt.Sprintf("Unhappy toCandlesticks %v", i), func(t *testing.T) {
			sr := response{}
			err := json.Unmarshal([]byte(ts), &sr)
			if err != nil {
				t.Fatalf("Unmarshal failed: %v", err)
			}

			cs, err := sr.toCandlesticks()
			if err == nil {
				t.Fatalf("Candlestick should have failed to convert but converted successfully to: %v", cs)
			}
		})
	}
}

func TestKlinesInvalidUrl(t *testing.T) {
	i := 0
	replies := []string{
		`{"error":[],"result":{"XBTUSDT":[[1625623260,"34221.6","34221.6","34215.7","34215.7","34215.7","0.25998804",7]],"last":1626869340}}`,
	}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, replies[i%len(replies)])
		i++
	}))
	defer ts.Close()

	b := NewKraken()
	b.overrideAPIURL("invalid url")
	_, err := b.RequestTicks(opBTCUSDT, tInt("2021-07-04T14:14:18+00:00"))
	if err == nil {
		t.Fatalf("should have failed due to invalid url")
	}
}

func TestKlinesErrReadingResponseBody(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "1")
	}))
	defer ts.Close()

	b := NewKraken()
	b.overrideAPIURL(ts.URL + "/")
	_, err := b.RequestTicks(opBTCUSDT, tInt("2021-07-04T14:14:18+00:00"))
	if err == nil {
		t.Fatalf("should have failed due to invalid response body")
	}
}

func TestKlinesErrorResponse(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, `{"message":"error!"}`)
	}))
	defer ts.Close()

	b := NewKraken()
	b.overrideAPIURL(ts.URL + "/")
	_, err := b.RequestTicks(opBTCUSDT, tInt("2021-07-04T14:14:18+00:00"))
	if err == nil {
		t.Fatalf("should have failed due to error response")
	}
}

func TestKlinesNon200Response(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
	}))
	defer ts.Close()

	b := NewKraken()
	b.overrideAPIURL(ts.URL + "/")
	_, err := b.RequestTicks(opBTCUSDT, tInt("2021-07-04T14:14:18+00:00"))
	if err == nil {
		t.Fatalf("should have failed due to 500 response")
	}
}

func TestKlinesInvalidJSONResponse(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, `invalid json`)
	}))
	defer ts.Close()

	b := NewKraken()
	b.overrideAPIURL(ts.URL + "/")
	_, err := b.RequestTicks(opBTCUSDT, tInt("2021-07-04T14:14:18+00:00"))
	if err == nil {
		t.Fatalf("should have failed due to invalid json")
	}
}

func TestKlinesInvalidFloatsInJSONResponse(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, `{"error":[],"result":{"XBTUSDT":"INVALID","last":1626869340}}`)
	}))
	defer ts.Close()

	b := NewKraken()
	b.overrideAPIURL(ts.URL + "/")
	_, err := b.RequestTicks(opBTCUSDT, tInt("2021-07-04T14:14:18+00:00"))
	if err == nil {
		t.Fatalf("should have failed due to invalid floats in json")
	}
}

func TestKlinesErrorInJSONResponse(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, `{"error":["one error!"],"result":{"XBTUSDT":[[1625623260,"34221.6","34221.6","34215.7","34215.7","34215.7","0.25998804",7]],"last":1626869340}}`)
	}))
	defer ts.Close()

	b := NewKraken()
	b.overrideAPIURL(ts.URL + "/")
	_, err := b.RequestTicks(opBTCUSDT, tInt("2021-07-04T14:14:18+00:00"))
	if err == nil {
		t.Fatalf("should have failed due to error in json response")
	}
}

func TestKlinesErrorInJSONResponseLastField(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, `{"error":[],"result":{"XBTUSDT":[[1625623260,"34221.6","34221.6","34215.7","34215.7","34215.7","0.25998804",7]],"last":"1626869340"}}`)
	}))
	defer ts.Close()

	b := NewKraken()
	b.overrideAPIURL(ts.URL + "/")
	_, err := b.RequestTicks(opBTCUSDT, tInt("2021-07-04T14:14:18+00:00"))
	if err == nil {
		t.Fatalf("should have failed due to error in json response's 'last' field")
	}
}

func f(fl float64) types.JsonFloat64 {
	return types.JsonFloat64(fl)
}

func tp(s string) time.Time {
	t, _ := time.Parse(time.RFC3339, s)
	return t
}

func tInt(s string) int {
	return int(tp(s).Unix())
}

var (
	opBTCUSDT = types.Operand{
		Type:       types.COIN,
		Provider:   "BINANCE",
		BaseAsset:  "BTC",
		QuoteAsset: "USDT",
		Str:        "BINANCE:BTC:USDT",
	}
)
