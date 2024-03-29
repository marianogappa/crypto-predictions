package compiler

import (
	"errors"
	"net/url"
	"reflect"
	"testing"
	"time"

	"github.com/marianogappa/predictions/core"
	"github.com/marianogappa/predictions/metadatafetcher"
	mfTypes "github.com/marianogappa/predictions/metadatafetcher/types"
	"github.com/stretchr/testify/require"
)

func tp(s string) time.Time {
	t, _ := time.Parse("2006-01-02 15:04:05", s)
	return t
}

func tpToISO(s string) core.ISO8601 {
	t, _ := time.Parse("2006-01-02 15:04:05", s)
	return core.ISO8601(t.Format(time.RFC3339))
}

func TestParseDuration(t *testing.T) {
	var anyError = errors.New("any error for now... ")

	tss := []struct {
		name     string
		dur      string
		fromTime time.Time
		err      error
		expected time.Duration
	}{
		{
			name:     "Empty string cannot be parsed",
			dur:      "",
			fromTime: time.Now(),
			err:      anyError,
			expected: 0,
		},
		{
			name:     "EOY works",
			dur:      "eoy",
			fromTime: tp("2020-12-31 23:59:59"),
			err:      nil,
			expected: 1 * time.Second,
		},
		{
			name:     "EOD works",
			dur:      "eod",
			fromTime: tp("2020-12-01 23:59:59"),
			err:      nil,
			expected: 1 * time.Second,
		},
		{
			name:     "EOND works",
			dur:      "eond",
			fromTime: tp("2020-12-01 23:59:59"),
			err:      nil,
			expected: 1*time.Second + 24*time.Hour,
		},
		{
			name:     "EOM works",
			dur:      "eom",
			fromTime: tp("2020-12-11 00:00:00"),
			err:      nil,
			expected: 21 * 24 * time.Hour,
		},
		{
			name:     "EONY works",
			dur:      "eony",
			fromTime: tp("2020-12-31 23:59:59"),
			err:      nil,
			expected: 1*time.Second + /* All next year */ 365*24*time.Hour,
		},
		{
			name:     "EONM works",
			dur:      "eonm",
			fromTime: tp("2020-12-11 00:00:00"),
			err:      nil,
			expected:/* January */ 31*24*time.Hour + /* Rest of December */ 21*24*time.Hour,
		},
		{
			name:     "2d works",
			dur:      "2d",
			fromTime: time.Now(),
			err:      nil,
			expected: 48 * time.Hour,
		},
		{
			name:     "2w works",
			dur:      "2w",
			fromTime: time.Now(),
			err:      nil,
			expected: 24 * 7 * 2 * time.Hour,
		},
		{
			name:     "2m works",
			dur:      "2m",
			fromTime: time.Now(),
			err:      nil,
			expected: 24 * 30 * 2 * time.Hour,
		},
		{
			name:     "2h works",
			dur:      "2h",
			fromTime: time.Now(),
			err:      nil,
			expected: 2 * time.Hour,
		},
	}
	for _, ts := range tss {
		t.Run(ts.name, func(t *testing.T) {
			actual, actualErr := parseDuration(ts.dur, ts.fromTime)

			if actualErr != nil && ts.err == nil {
				t.Logf("expected no error but had '%v'", actualErr)
				t.FailNow()
			}
			if actualErr == nil && ts.err != nil {
				t.Logf("expected error '%v' but had no error", actualErr)
				t.FailNow()
			}
			if actual != ts.expected {
				t.Logf("expected %v but got %v", ts.expected, actual)
				t.FailNow()
			}
		})
	}
}

func TestMapOperand(t *testing.T) {
	tss := []struct {
		raw      string
		err      error
		expected core.Operand
	}{
		{
			raw:      "",
			err:      core.ErrInvalidOperand,
			expected: core.Operand{},
		},
		{
			raw: "1.1",
			err: nil,
			expected: core.Operand{
				Type:       core.NUMBER,
				Provider:   "",
				QuoteAsset: "",
				BaseAsset:  "",
				Number:     1.1,
				Str:        "1.1",
			},
		},
		{
			raw:      "COIN:BINANCE:BTC:USDT",
			err:      core.ErrInvalidOperand,
			expected: core.Operand{},
		},
		{
			raw:      "COIN:BINANCE:BTC-BTC",
			err:      core.ErrEqualBaseQuoteAssets,
			expected: core.Operand{},
		},
		{
			raw: "COIN:BINANCE:BTC-USDT",
			err: nil,
			expected: core.Operand{
				Type:       core.COIN,
				Provider:   "BINANCE",
				BaseAsset:  "BTC",
				QuoteAsset: "USDT",
				Number:     0,
				Str:        "COIN:BINANCE:BTC-USDT",
			},
		},
		{
			raw: "COIN:BINANCE:BTC",
			err: core.ErrEmptyQuoteAsset,
		},
		{
			raw: "MARKETCAP:MESSARI:BTC-USDT",
			err: core.ErrNonEmptyQuoteAssetOnNonCoin,
		},
		{
			raw: "MARKETCAP:MESSARI:BTC",
			err: nil,
			expected: core.Operand{
				Type:      core.MARKETCAP,
				Provider:  "MESSARI",
				BaseAsset: "BTC",
				Number:    0,
				Str:       "MARKETCAP:MESSARI:BTC",
			},
		},
	}
	for _, ts := range tss {
		t.Run(ts.raw, func(t *testing.T) {
			actual, actualErr := mapOperandForTests(ts.raw)

			if actualErr != nil && ts.err == nil {
				t.Logf("expected no error but had '%v'", actualErr)
				t.FailNow()
			}
			if actualErr == nil && ts.err != nil {
				t.Logf("expected error '%v' but had no error", actualErr)
				t.FailNow()
			}
			if ts.err != nil && actualErr != nil && !errors.Is(actualErr, ts.err) {
				t.Logf("expected error '%v' but had error '%v'", ts.err, actualErr)
				t.FailNow()
			}
			if actual != ts.expected {
				t.Logf("expected %v but got %v", ts.expected, actual)
				t.FailNow()
			}
		})
	}
}

func TestMapOperands(t *testing.T) {
	ops, err := mapOperands([]string{"COIN:BINANCE:BTC-USDT", "1.1"})
	if err != nil {
		t.Errorf("should have succeeded but error happened: %v", err)
		t.FailNow()
	}

	expected := []core.Operand{
		{
			Type:       core.COIN,
			Provider:   "BINANCE",
			BaseAsset:  "BTC",
			QuoteAsset: "USDT",
			Number:     0,
			Str:        "COIN:BINANCE:BTC-USDT",
		},
		{
			Type:   core.NUMBER,
			Number: 1.1,
			Str:    "1.1",
		},
	}
	if !reflect.DeepEqual(ops, expected) {
		t.Errorf("expected %v but got: %v", expected, ops)
		t.FailNow()
	}

	_, err = mapOperands([]string{"", "1.1"})
	if !errors.Is(err, core.ErrInvalidOperand) {
		t.Errorf("expected %v but got: %v", core.ErrInvalidOperand, err)
		t.FailNow()
	}
}

func TestMapFromTs(t *testing.T) {
	tss := []struct {
		name     string
		cond     Condition
		postedAt core.ISO8601
		err      error
		expected int
	}{
		{
			name:     "Uses postedAt when empty",
			cond:     Condition{FromISO8601: ""},
			postedAt: tpToISO("2020-01-02 00:00:00"),
			err:      nil,
			expected: int(tp("2020-01-02 00:00:00").Unix()),
		},
		{
			name:     "Invalid dates fail",
			cond:     Condition{FromISO8601: "invalid date"},
			postedAt: tpToISO("2020-01-02 00:00:00"),
			err:      core.ErrInvalidFromISO8601,
			expected: 0,
		},
		{
			name:     "Valid dates take precedence over postedAt",
			cond:     Condition{FromISO8601: tpToISO("2022-01-02 00:00:00")},
			postedAt: tpToISO("2020-01-02 00:00:00"),
			err:      nil,
			expected: int(tp("2022-01-02 00:00:00").Unix()),
		},
	}
	for _, ts := range tss {
		t.Run(ts.name, func(t *testing.T) {
			actual, actualErr := mapFromTs(ts.cond, ts.postedAt)

			if actualErr != nil && ts.err == nil {
				t.Logf("expected no error but had '%v'", actualErr)
				t.FailNow()
			}
			if actualErr == nil && ts.err != nil {
				t.Logf("expected error '%v' but had no error", actualErr)
				t.FailNow()
			}
			if ts.err != nil && actualErr != nil && !errors.Is(actualErr, ts.err) {
				t.Logf("expected error '%v' but had error '%v'", ts.err, actualErr)
				t.FailNow()
			}
			if actual != ts.expected {
				t.Logf("expected %v but got %v", ts.expected, actual)
				t.FailNow()
			}
		})
	}
}

func TestMapToTs(t *testing.T) {
	tss := []struct {
		name     string
		cond     Condition
		fromTs   int
		err      error
		expected int
	}{
		{
			name:   "All empty",
			cond:   Condition{ToISO8601: "", ToDuration: ""},
			fromTs: int(tp("2020-01-02 00:00:00").Unix()),
			err:    core.ErrOneOfToISO8601ToDurationRequired,
		},
		{
			name:   "Invalid duration",
			cond:   Condition{ToISO8601: "", ToDuration: "invalid"},
			fromTs: int(tp("2020-01-02 00:00:00").Unix()),
			err:    core.ErrInvalidDuration,
		},
		{
			name:     "Uses FromISO8601+ToDuration when ToISO8601",
			cond:     Condition{ToISO8601: "", ToDuration: "2d"},
			fromTs:   int(tp("2020-01-02 00:00:00").Unix()),
			err:      nil,
			expected: int(tp("2020-01-04 00:00:00").Unix()),
		},
		{
			name:   "Invalid dates fail",
			cond:   Condition{ToISO8601: "invalid date", ToDuration: "2w"},
			fromTs: int(tp("2020-01-02 00:00:00").Unix()),
			err:    core.ErrInvalidToISO8601,
		},
		{
			name:     "Valid dates take precedence over everything",
			cond:     Condition{ToISO8601: tpToISO("2022-01-02 00:00:00"), ToDuration: "2w"},
			fromTs:   int(tp("2020-01-02 00:00:00").Unix()),
			err:      nil,
			expected: int(tp("2022-01-02 00:00:00").Unix()),
		},
	}
	for _, ts := range tss {
		t.Run(ts.name, func(t *testing.T) {
			actual, actualErr := mapToTs(ts.cond, ts.fromTs)

			if actualErr != nil && ts.err == nil {
				t.Logf("expected no error but had '%v'", actualErr)
				t.FailNow()
			}
			if actualErr == nil && ts.err != nil {
				t.Logf("expected error '%v' but had no error", actualErr)
				t.FailNow()
			}
			if ts.err != nil && actualErr != nil && !errors.Is(actualErr, ts.err) {
				t.Logf("expected error '%v' but had error '%v'", ts.err, actualErr)
				t.FailNow()
			}
			if actual != ts.expected {
				t.Logf("expected %v but got %v", ts.expected, actual)
				t.FailNow()
			}
		})
	}
}

func TestMapCondition(t *testing.T) {
	tss := []struct {
		name     string
		cond     Condition
		condName string
		postedAt core.ISO8601
		err      error
		expected core.Condition
	}{
		{
			name:     "Invalid condition syntax",
			cond:     Condition{Condition: "invalid condition syntax!!"},
			condName: "main",
			postedAt: tpToISO("2020-01-02 00:00:00"),
			err:      core.ErrInvalidConditionSyntax,
		},
		{
			name:     "Empty quote asset",
			cond:     Condition{Condition: "COIN:BINANCE:BTC >= 60000"},
			condName: "main",
			postedAt: tpToISO("2020-01-02 00:00:00"),
			err:      core.ErrEmptyQuoteAsset,
		},
		{
			name:     "Unknown condition operator",
			cond:     Condition{Condition: "COIN:BINANCE:BTC-USDT != 60000"},
			condName: "main",
			postedAt: tpToISO("2020-01-02 00:00:00"),
			err:      core.ErrUnknownConditionOperator,
		},
		{
			name:     "Unknown condition operator",
			cond:     Condition{Condition: "COIN:BINANCE:BTC-USDT >= 60000", ErrorMarginRatio: 0.4},
			condName: "main",
			postedAt: tpToISO("2020-01-02 00:00:00"),
			err:      core.ErrErrorMarginRatioAbove30,
		},
		{
			name:     "Unknown condition state value",
			cond:     Condition{Condition: "COIN:BINANCE:BTC-USDT >= 60000", State: ConditionState{Value: "???"}},
			condName: "main",
			postedAt: tpToISO("2020-01-02 00:00:00"),
			err:      core.ErrUnknownConditionStateValue,
		},
		{
			name:     "Unknown condition status",
			cond:     Condition{Condition: "COIN:BINANCE:BTC-USDT >= 60000", State: ConditionState{Value: "UNDECIDED", Status: "???"}},
			condName: "main",
			postedAt: tpToISO("2020-01-02 00:00:00"),
			err:      core.ErrUnknownConditionStatus,
		},
		{
			name: "Invalid FromISO8601",
			cond: Condition{
				Condition:   "COIN:BINANCE:BTC-USDT >= 60000",
				State:       ConditionState{Value: "UNDECIDED", Status: "STARTED"},
				FromISO8601: "invalid",
			},
			condName: "main",
			postedAt: tpToISO("2020-01-02 00:00:00"),
			err:      core.ErrInvalidFromISO8601,
		},
		{
			name: "Invalid ToISO8601",
			cond: Condition{
				Condition: "COIN:BINANCE:BTC-USDT >= 60000",
				State:     ConditionState{Value: "UNDECIDED", Status: "STARTED"},
				ToISO8601: "invalid",
			},
			condName: "main",
			postedAt: tpToISO("2020-01-02 00:00:00"),
			err:      core.ErrInvalidToISO8601,
		},
		{
			name: "Happy case",
			cond: Condition{
				Condition: "COIN:BINANCE:BTC-USDT BETWEEN 60000 AND 70000",
				State:     ConditionState{Value: "UNDECIDED", Status: "STARTED"},
				ToISO8601: tpToISO("2020-01-03 00:00:00"),
			},
			condName: "main",
			postedAt: tpToISO("2020-01-02 00:00:00"),
			err:      nil,
			expected: core.Condition{
				Name:     "main",
				Operator: "BETWEEN",
				Operands: []core.Operand{
					{
						Type:       core.COIN,
						Provider:   "BINANCE",
						QuoteAsset: "USDT",
						BaseAsset:  "BTC",
						Str:        "COIN:BINANCE:BTC-USDT",
					},
					{
						Type:   core.NUMBER,
						Number: 60000,
						Str:    "60000",
					},
					{
						Type:   core.NUMBER,
						Number: 70000,
						Str:    "70000",
					},
				},
				FromTs:           int(tp("2020-01-02 00:00:00").Unix()),
				ToTs:             int(tp("2020-01-03 00:00:00").Unix()),
				ToDuration:       "",
				Assumed:          nil,
				State:            core.ConditionState{Value: core.UNDECIDED, Status: core.STARTED},
				ErrorMarginRatio: 0,
			},
		},
	}
	for _, ts := range tss {
		t.Run(ts.name, func(t *testing.T) {
			actual, actualErr := mapCondition(ts.cond, ts.condName, ts.postedAt)

			if actualErr != nil && ts.err == nil {
				t.Logf("expected no error but had '%v'", actualErr)
				t.FailNow()
			}
			if actualErr == nil && ts.err != nil {
				t.Logf("expected error '%v' but had no error", actualErr)
				t.FailNow()
			}
			if ts.err != nil && actualErr != nil && !errors.Is(actualErr, ts.err) {
				t.Logf("expected error '%v' but had error '%v'", ts.err, actualErr)
				t.FailNow()
			}
			if !reflect.DeepEqual(actual, ts.expected) {
				t.Logf("expected %v but got %v", ts.expected, actual)
				t.FailNow()
			}
		})
	}
}

func TestMapBoolExprNil(t *testing.T) {
	res, err := mapBoolExpr(nil, nil)
	if res != nil {
		t.Errorf("res should have been nil but was %v", res)
	}
	if err != nil {
		t.Errorf("err should have been nil but was %v", err)
	}
}

func TestMapBoolExprInvalid(t *testing.T) {
	invalid := "invalid"
	_, err := mapBoolExpr(&invalid, nil)
	if !errors.Is(err, core.ErrBoolExprSyntaxError) {
		t.Errorf("err should have been ErrBoolExprSyntaxError but was %v", err)
	}
}

func TestMapBoolExprHappyCase(t *testing.T) {
	valid := "main"
	_, err := mapBoolExpr(&valid, map[string]*core.Condition{"main": {}})
	if err != nil {
		t.Errorf("err should have been nil but was %v", err)
	}
}

type testMetadataFetcher struct {
	postMetadata         mfTypes.PostMetadata
	postMetadataFetchErr error
}

func newTestMetadataFetcher(postMetadata mfTypes.PostMetadata, postMetadataFetchErr error) *testMetadataFetcher {
	return &testMetadataFetcher{postMetadata, postMetadataFetchErr}
}

func (f testMetadataFetcher) Fetch(u *url.URL) (mfTypes.PostMetadata, error) {
	return f.postMetadata, f.postMetadataFetchErr
}

func (f testMetadataFetcher) IsCorrectFetcher(u *url.URL) bool {
	return true
}

func TestCompile(t *testing.T) {
	tss := []struct {
		name                 string
		pred                 string
		postMetadata         mfTypes.PostMetadata
		timeNow              func() time.Time
		postMetadataFetchErr error
		err                  error
		expected             core.Prediction
	}{
		{
			name:     "Invalid JSON",
			pred:     "invalid!!",
			err:      core.ErrInvalidJSON,
			expected: core.Prediction{},
		},
		{
			name:     "Empty reporter",
			pred:     `{"reporter": ""}`,
			err:      core.ErrEmptyReporter,
			expected: core.Prediction{},
		},
		{
			name:     "Empty postUrl",
			pred:     `{"reporter": "admin", "postUrl": ""}`,
			err:      core.ErrEmptyPostURL,
			expected: core.Prediction{},
		},
		{
			name:                 "Metadata fetcher returns error",
			pred:                 `{"reporter": "admin", "postUrl": "https://twitter.com/CryptoCapo_/status/1491357566974054400"}`,
			postMetadataFetchErr: core.ErrEmptyPostAuthor,
			err:                  core.ErrEmptyPostAuthor,
			expected:             core.Prediction{},
		},
		{
			name:                 "Metadata fetcher returns postAuthor but not postedAt",
			pred:                 `{"reporter": "admin", "postUrl": "https://twitter.com/CryptoCapo_/status/1491357566974054400"}`,
			postMetadataFetchErr: nil,
			postMetadata: mfTypes.PostMetadata{
				Author:        core.Account{Handle: "CryptoCapo_"},
				PostCreatedAt: core.ISO8601(""),
			},
			err:      core.ErrEmptyPostedAt,
			expected: core.Prediction{},
		},
		{
			name:                 "Metadata fetcher returns postAuthor and invalid postedAt",
			pred:                 `{"reporter": "admin", "postUrl": "https://twitter.com/CryptoCapo_/status/1491357566974054400"}`,
			postMetadataFetchErr: nil,
			postMetadata: mfTypes.PostMetadata{
				Author:        core.Account{Handle: "CryptoCapo_"},
				PostCreatedAt: core.ISO8601("INVALID!!!"),
			},
			err:      core.ErrInvalidPostedAt,
			expected: core.Prediction{},
		},
		{
			name: "Empty main predict",
			pred: `{"reporter": "admin", "postUrl": "https://twitter.com/CryptoCapo_/status/1491357566974054400"}`,
			postMetadata: mfTypes.PostMetadata{
				Author:        core.Account{Handle: "CryptoCapo_"},
				PostCreatedAt: tpToISO("2020-01-02 00:00:00"),
			},
			err: core.ErrEmptyPredict,
		},
		{
			name: "Error mapping condition: no ToISO8601",
			pred: `{
				"reporter": "admin",
				"postUrl": "https://twitter.com/CryptoCapo_/status/1491357566974054400",
				"given": {
					"main": {
						"condition": "COIN:BINANCE:ADA-USDT <= 0.845"
					}
				},
				"predict": {
					"predict": "main"
				}
			}`,
			postMetadataFetchErr: nil,
			postMetadata: mfTypes.PostMetadata{
				Author:        core.Account{Handle: "CryptoCapo_"},
				PostCreatedAt: tpToISO("2020-01-02 00:00:00"),
			},
			err:      core.ErrOneOfToISO8601ToDurationRequired,
			expected: core.Prediction{},
		},
		{
			name: "Error mapping prePredict.predict: ErrBoolExprSyntaxError",
			pred: `{
				"reporter": "admin",
				"postUrl": "https://twitter.com/CryptoCapo_/status/1491357566974054400",
				"given": {
					"main": {
						"condition": "COIN:BINANCE:ADA-USDT <= 0.845",
						"toDuration": "2w"
					}
				},
				"prePredict": {
					"predict": "???"
				},
				"predict": {
					"predict": "main"
				}
			}`,
			postMetadataFetchErr: nil,
			postMetadata: mfTypes.PostMetadata{
				Author:        core.Account{Handle: "CryptoCapo_"},
				PostCreatedAt: tpToISO("2020-01-02 00:00:00"),
			},
			err:      core.ErrBoolExprSyntaxError,
			expected: core.Prediction{},
		},
		{
			name: "Error mapping prePredict.wrongIf: ErrBoolExprSyntaxError",
			pred: `{
				"reporter": "admin",
				"postUrl": "https://twitter.com/CryptoCapo_/status/1491357566974054400",
				"given": {
					"main": {
						"condition": "COIN:BINANCE:ADA-USDT <= 0.845",
						"toDuration": "2w"
					}
				},
				"prePredict": {
					"wrongIf": "???",
					"predict": "main"
				},
				"predict": {
					"predict": "main"
				}
			}`,
			postMetadataFetchErr: nil,
			postMetadata: mfTypes.PostMetadata{
				Author:        core.Account{Handle: "CryptoCapo_"},
				PostCreatedAt: tpToISO("2020-01-02 00:00:00"),
			},
			err:      core.ErrBoolExprSyntaxError,
			expected: core.Prediction{},
		},
		{
			name: "Error mapping prePredict.annulledIf: ErrBoolExprSyntaxError",
			pred: `{
				"reporter": "admin",
				"postUrl": "https://twitter.com/CryptoCapo_/status/1491357566974054400",
				"given": {
					"main": {
						"condition": "COIN:BINANCE:ADA-USDT <= 0.845",
						"toDuration": "2w"
					}
				},
				"prePredict": {
					"annulledIf": "???",
					"wrongIf": "main",
					"predict": "main"
				},
				"predict": {
					"predict": "main"
				}
			}`,
			postMetadataFetchErr: nil,
			postMetadata: mfTypes.PostMetadata{
				Author:        core.Account{Handle: "CryptoCapo_"},
				PostCreatedAt: tpToISO("2020-01-02 00:00:00"),
			},
			err:      core.ErrBoolExprSyntaxError,
			expected: core.Prediction{},
		},
		{
			name: "Must have prePredict.predict if it has prePredict.wrongIf",
			pred: `{
				"reporter": "admin",
				"postUrl": "https://twitter.com/CryptoCapo_/status/1491357566974054400",
				"given": {
					"main": {
						"condition": "COIN:BINANCE:ADA-USDT <= 0.845",
						"toDuration": "2w"
					}
				},
				"prePredict": {
					"wrongIf": "main"
				}
			}`,
			postMetadataFetchErr: nil,
			postMetadata: mfTypes.PostMetadata{
				Author:        core.Account{Handle: "CryptoCapo_"},
				PostCreatedAt: tpToISO("2020-01-02 00:00:00"),
			},
			err:      core.ErrMissingRequiredPrePredictPredictIf,
			expected: core.Prediction{},
		},
		{
			name: "Must have prePredict.predict if it has prePredict.annulledIf",
			pred: `{
				"reporter": "admin",
				"postUrl": "https://twitter.com/CryptoCapo_/status/1491357566974054400",
				"given": {
					"main": {
						"condition": "COIN:BINANCE:ADA-USDT <= 0.845",
						"toDuration": "2w"
					}
				},
				"prePredict": {
					"annulledIf": "main"
				}
			}`,
			postMetadataFetchErr: nil,
			postMetadata: mfTypes.PostMetadata{
				Author:        core.Account{Handle: "CryptoCapo_"},
				PostCreatedAt: tpToISO("2020-01-02 00:00:00"),
			},
			err:      core.ErrMissingRequiredPrePredictPredictIf,
			expected: core.Prediction{},
		},
		{
			name: "Error mapping predict.wrongIf: ErrBoolExprSyntaxError",
			pred: `{
				"reporter": "admin",
				"postUrl": "https://twitter.com/CryptoCapo_/status/1491357566974054400",
				"given": {
					"main": {
						"condition": "COIN:BINANCE:ADA-USDT <= 0.845",
						"toDuration": "2w"
					}
				},
				"predict": {
					"wrongIf": "???",
					"predict": "main"
				}
			}`,
			postMetadataFetchErr: nil,
			postMetadata: mfTypes.PostMetadata{
				Author:        core.Account{Handle: "CryptoCapo_"},
				PostCreatedAt: tpToISO("2020-01-02 00:00:00"),
			},
			err:      core.ErrBoolExprSyntaxError,
			expected: core.Prediction{},
		},
		{
			name: "Error mapping predict.annulledIf: ErrBoolExprSyntaxError",
			pred: `{
				"reporter": "admin",
				"postUrl": "https://twitter.com/CryptoCapo_/status/1491357566974054400",
				"given": {
					"main": {
						"condition": "COIN:BINANCE:ADA-USDT <= 0.845",
						"toDuration": "2w"
					}
				},
				"predict": {
					"annulledIf": "???",
					"predict": "main"
				}
			}`,
			postMetadataFetchErr: nil,
			postMetadata: mfTypes.PostMetadata{
				Author:        core.Account{Handle: "CryptoCapo_"},
				PostCreatedAt: tpToISO("2020-01-02 00:00:00"),
			},
			err:      core.ErrBoolExprSyntaxError,
			expected: core.Prediction{},
		},
		{
			name: "Error mapping predict.predict: ErrBoolExprSyntaxError",
			pred: `{
				"reporter": "admin",
				"postUrl": "https://twitter.com/CryptoCapo_/status/1491357566974054400",
				"given": {
					"main": {
						"condition": "COIN:BINANCE:ADA-USDT <= 0.845",
						"toDuration": "2w"
					}
				},
				"predict": {
					"predict": "???"
				}
			}`,
			postMetadataFetchErr: nil,
			postMetadata: mfTypes.PostMetadata{
				Author:        core.Account{Handle: "CryptoCapo_"},
				PostCreatedAt: tpToISO("2020-01-02 00:00:00"),
			},
			err:      core.ErrBoolExprSyntaxError,
			expected: core.Prediction{},
		},
		{
			name: "Error mapping prediction state: ErrUnknownConditionStatus",
			pred: `{
				"reporter": "admin",
				"postUrl": "https://twitter.com/CryptoCapo_/status/1491357566974054400",
				"given": {
					"main": {
						"condition": "COIN:BINANCE:ADA-USDT <= 0.845",
						"toDuration": "2w"
					}
				},
				"prePredict": {
					"predict": "main"
				},
				"predict": {
					"annulledIf": "main",
					"wrongIf": "main",
					"predict": "main"
				},
				"state": {
					"status": "???"
				}
			}`,
			postMetadataFetchErr: nil,
			postMetadata: mfTypes.PostMetadata{
				Author:        core.Account{Handle: "CryptoCapo_"},
				PostCreatedAt: tpToISO("2020-01-02 00:00:00"),
			},
			err:      core.ErrUnknownConditionStatus,
			expected: core.Prediction{},
		},
		{
			name: "Error mapping prediction state: ErrUnknownPredictionStateValue",
			pred: `{
				"reporter": "admin",
				"postUrl": "https://twitter.com/CryptoCapo_/status/1491357566974054400",
				"given": {
					"main": {
						"condition": "COIN:BINANCE:ADA-USDT <= 0.845",
						"toDuration": "2w"
					}
				},
				"prePredict": {
					"predict": "main"
				},
				"predict": {
					"annulledIf": "main",
					"wrongIf": "main",
					"predict": "main"
				},
				"state": {
					"status": "STARTED",
					"value": "???"
				}
			}`,
			postMetadataFetchErr: nil,
			postMetadata: mfTypes.PostMetadata{
				Author:        core.Account{Handle: "CryptoCapo_"},
				PostCreatedAt: tpToISO("2020-01-02 00:00:00"),
			},
			err:      core.ErrUnknownPredictionStateValue,
			expected: core.Prediction{},
		},
		{
			name: "Does not overwrite author & created at with metadata, when fields are set",
			pred: `{
				"reporter": "admin",
				"postUrl": "https://twitter.com/CryptoCapo_/status/1491357566974054400",
				"postAuthor": "NOT CryptoCapo!",
				"postedAt": "2022-02-09T10:25:26.000Z",
				"given": {
					"main": {
						"condition": "COIN:BINANCE:ADA-USDT <= 0.845",
						"toDuration": "2d"
					}
				},
				"predict": {
					"predict": "main"
				}
			}`,
			postMetadataFetchErr: nil,
			postMetadata: mfTypes.PostMetadata{
				Author:        core.Account{Handle: "CryptoCapo_"},
				PostCreatedAt: tpToISO("2020-01-02 00:00:00"),
			},
			timeNow: func() time.Time { return tp("2020-01-03 00:00:00") },
			err:     nil,
			expected: core.Prediction{
				Version:    "1.0.0",
				CreatedAt:  tpToISO("2020-01-03 00:00:00"),
				Reporter:   "admin",
				PostAuthor: "NOT CryptoCapo!",
				PostText:   "",
				PostedAt:   core.ISO8601("2022-02-09T10:25:26.000Z"),
				PostURL:    "https://twitter.com/CryptoCapo_/status/1491357566974054400",
				Given: map[string]*core.Condition{
					"main": {
						Name:     "main",
						Operator: "<=",
						Operands: []core.Operand{
							{Type: core.COIN, Provider: "BINANCE", BaseAsset: "ADA", QuoteAsset: "USDT", Str: "COIN:BINANCE:ADA-USDT"},
							{Type: core.NUMBER, Number: core.JSONFloat64(0.845), Str: "0.845"},
						},
						FromTs:     int(tp("2022-02-09 10:25:26").Unix()),
						ToTs:       int(tp("2022-02-11 10:25:26").Unix()),
						ToDuration: "2d",
					},
				},
				Predict: core.Predict{
					Predict: core.BoolExpr{
						Operator: core.LITERAL,
						Operands: nil,
						Literal: &core.Condition{
							Name:     "main",
							Operator: "<=",
							Operands: []core.Operand{
								{Type: core.COIN, Provider: "BINANCE", BaseAsset: "ADA", QuoteAsset: "USDT", Str: "COIN:BINANCE:ADA-USDT"},
								{Type: core.NUMBER, Number: core.JSONFloat64(0.845), Str: "0.845"},
							},
							FromTs:     int(tp("2022-02-09 10:25:26").Unix()),
							ToTs:       int(tp("2022-02-11 10:25:26").Unix()),
							ToDuration: "2d",
						},
					},
				},
				Type: core.PredictionTypeCoinOperatorFloatDeadline,
			},
		},
		{
			name: "Overwrites author but not created at with metadata, when created at is set",
			pred: `{
				"reporter": "admin",
				"postUrl": "https://twitter.com/CryptoCapo_/status/1491357566974054400",
				"postedAt": "2022-02-09T10:25:26.000Z",
				"given": {
					"main": {
						"condition": "COIN:BINANCE:ADA-USDT <= 0.845",
						"toDuration": "2d"
					}
				},
				"predict": {
					"predict": "main"
				}
			}`,
			postMetadataFetchErr: nil,
			postMetadata: mfTypes.PostMetadata{
				Author:        core.Account{Handle: "CryptoCapo_"},
				PostCreatedAt: tpToISO("2020-01-02 00:00:00"),
			},
			timeNow: func() time.Time { return tp("2020-01-03 00:00:00") },
			err:     nil,
			expected: core.Prediction{
				Version:    "1.0.0",
				CreatedAt:  tpToISO("2020-01-03 00:00:00"),
				Reporter:   "admin",
				PostAuthor: "CryptoCapo_",
				PostText:   "",
				PostedAt:   core.ISO8601("2022-02-09T10:25:26.000Z"),
				PostURL:    "https://twitter.com/CryptoCapo_/status/1491357566974054400",
				Given: map[string]*core.Condition{
					"main": {
						Name:     "main",
						Operator: "<=",
						Operands: []core.Operand{
							{Type: core.COIN, Provider: "BINANCE", BaseAsset: "ADA", QuoteAsset: "USDT", Str: "COIN:BINANCE:ADA-USDT"},
							{Type: core.NUMBER, Number: core.JSONFloat64(0.845), Str: "0.845"},
						},
						FromTs:     int(tp("2022-02-09 10:25:26").Unix()),
						ToTs:       int(tp("2022-02-11 10:25:26").Unix()),
						ToDuration: "2d",
					},
				},
				Predict: core.Predict{
					Predict: core.BoolExpr{
						Operator: core.LITERAL,
						Operands: nil,
						Literal: &core.Condition{
							Name:     "main",
							Operator: "<=",
							Operands: []core.Operand{
								{Type: core.COIN, Provider: "BINANCE", BaseAsset: "ADA", QuoteAsset: "USDT", Str: "COIN:BINANCE:ADA-USDT"},
								{Type: core.NUMBER, Number: core.JSONFloat64(0.845), Str: "0.845"},
							},
							FromTs:     int(tp("2022-02-09 10:25:26").Unix()),
							ToTs:       int(tp("2022-02-11 10:25:26").Unix()),
							ToDuration: "2d",
						},
					},
				},
				Type: core.PredictionTypeCoinOperatorFloatDeadline,
			},
		},
	}
	for _, ts := range tss {
		t.Run(ts.name, func(t *testing.T) {
			pc := NewPredictionCompiler(metadatafetcher.NewMetadataFetcher(), nil)
			pc.metadataFetcher.Fetchers = []metadatafetcher.SpecificFetcher{
				newTestMetadataFetcher(ts.postMetadata, ts.postMetadataFetchErr),
			}
			if ts.timeNow != nil {
				pc.timeNow = ts.timeNow
			}
			// TODO test accounts
			actual, _, actualErr := pc.Compile([]byte(ts.pred))

			if actualErr != nil && ts.err == nil {
				t.Logf("expected no error but had '%v'", actualErr)
				t.FailNow()
			}
			if actualErr == nil && ts.err != nil {
				t.Logf("expected error '%v' but had no error", actualErr)
				t.FailNow()
			}
			if ts.err != nil && actualErr != nil && !errors.Is(actualErr, ts.err) {
				t.Logf("expected error '%v' but had error '%v'", ts.err, actualErr)
				t.FailNow()
			}
			if ts.err == nil {
				require.Equal(t, ts.expected, actual)
			}
		})
	}
}
