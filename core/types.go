package core

import (
	"errors"
	"fmt"
	"math"
	"net/url"
	"time"

	"github.com/marianogappa/crypto-candles/candles/common"
	"github.com/marianogappa/crypto-candles/candles/iterator"
)

var (
	// UIUnsupportedPredictionTypes controls which prediction types are not Tweetable/showable/linkable in the UI.
	// If a PredictionType is on this list, it's probably because the overlays or pretty printing are unsupported.
	UIUnsupportedPredictionTypes = map[PredictionType]bool{
		PredictionTypeUnsupported:                         true,
		PredictionTypeCoinWillRange:                       true,
		PredictionTypeCoinWillReachBeforeItReaches:        true,
		PredictionTypeCoinWillReachInvalidatedIfItReaches: true,
	}

	// ErrUnknownOperandType means: unknown value for operandType
	ErrUnknownOperandType = errors.New("unknown value for operandType")

	// ErrUnknownBoolOperator means: unknown value for BoolOperator
	ErrUnknownBoolOperator = errors.New("unknown value for BoolOperator")

	// ErrUnknownConditionStatus means: invalid state: unknown value for ConditionStatus
	ErrUnknownConditionStatus = errors.New("invalid state: unknown value for ConditionStatus")

	// ErrUnknownPredictionStateValue means: invalid state: unknown value for PredictionStateValue
	ErrUnknownPredictionStateValue = errors.New("invalid state: unknown value for PredictionStateValue")

	// ErrInvalidOperand means: invalid operand
	ErrInvalidOperand = errors.New("invalid operand")

	// ErrEmptyQuoteAsset means: quote asset cannot be empty
	ErrEmptyQuoteAsset = errors.New("quote asset cannot be empty")

	// ErrNonEmptyQuoteAssetOnNonCoin means: quote asset must be empty for non-coin operand types
	ErrNonEmptyQuoteAssetOnNonCoin = errors.New("quote asset must be empty for non-coin operand types")

	// ErrEqualBaseQuoteAssets means: base asset cannot be equal to quote asset
	ErrEqualBaseQuoteAssets = errors.New("base asset cannot be equal to quote asset")

	// ErrInvalidDuration means: invalid duration
	ErrInvalidDuration = errors.New("invalid duration")

	// ErrInvalidFromISO8601 means: invalid FromISO8601
	ErrInvalidFromISO8601 = errors.New("invalid FromISO8601")

	// ErrInvalidToISO8601 means: invalid ToISO8601
	ErrInvalidToISO8601 = errors.New("invalid ToISO8601")

	// ErrOneOfToISO8601ToDurationRequired means: one of ToISO8601 or ToDuration is required
	ErrOneOfToISO8601ToDurationRequired = errors.New("one of ToISO8601 or ToDuration is required")

	// ErrInvalidConditionSyntax means: invalid condition syntax
	ErrInvalidConditionSyntax = errors.New("invalid condition syntax")

	// ErrUnknownConditionOperator means: unknown condition operator: supported are [>|<|>=|<=|BETWEEN...AND]
	ErrUnknownConditionOperator = errors.New("unknown condition operator: supported are [>|<|>=|<=|BETWEEN...AND]")

	// ErrErrorMarginRatioAbove30 means: error margin ratio above 30%% is not allowed
	ErrErrorMarginRatioAbove30 = errors.New("error margin ratio above 30%% is not allowed")

	// ErrInvalidJSON means: invalid JSON
	ErrInvalidJSON = errors.New("invalid JSON")

	// ErrEmptyPostURL means: postUrl cannot be empty
	ErrEmptyPostURL = errors.New("postUrl cannot be empty")

	// ErrEmptyReporter means: reporter cannot be empty
	ErrEmptyReporter = errors.New("reporter cannot be empty")

	// ErrEmptyPostAuthor means: postAuthor cannot be empty
	ErrEmptyPostAuthor = errors.New("postAuthor cannot be empty")

	// ErrEmptyPostedAt means: postedAt cannot be empty
	ErrEmptyPostedAt = errors.New("postedAt cannot be empty")

	// ErrInvalidPostedAt means: postedAt must be a valid ISO8601 timestamp
	ErrInvalidPostedAt = errors.New("postedAt must be a valid ISO8601 timestamp")

	// ErrEmptyPredict means: main predict clause cannot be empty
	ErrEmptyPredict = errors.New("main predict clause cannot be empty")

	// ErrMissingRequiredPrePredictPredictIf means: pre-predict clause must have predictIf if it has either wrongIf or annuledIf. Otherwise, add them directly on predict clause
	ErrMissingRequiredPrePredictPredictIf = errors.New("pre-predict clause must have predictIf if it has either wrongIf or annuledIf. Otherwise, add them directly on predict clause")

	// ErrBoolExprSyntaxError means: syntax error in bool expression
	ErrBoolExprSyntaxError = errors.New("syntax error in bool expression")

	// ErrPredictionFinishedAtStartTime means: prediction is finished at start time
	ErrPredictionFinishedAtStartTime = errors.New("prediction is finished at start time")

	// ErrUnknownAPIOrderBy means: unknown API order by
	ErrUnknownAPIOrderBy = errors.New("unknown API order by")

	// ErrInvalidExchange means: the only valid exchanges are 'binance', 'ftx', 'coinbase', 'huobi', 'kraken', 'kucoin' and 'binanceusdmfutures'
	ErrInvalidExchange = errors.New("the only valid exchanges are 'binance', 'ftx', 'coinbase', 'huobi', 'kraken', 'kucoin' and 'binanceusdmfutures'")

	// ErrBaseAssetRequired means: base asset is required (e.g. BTC)
	ErrBaseAssetRequired = errors.New("base asset is required (e.g. BTC)")

	// ErrQuoteAssetRequired means: quote asset is required (e.g. USDT)
	ErrQuoteAssetRequired = errors.New("quote asset is required (e.g. USDT)")

	// From storage

	// ErrStorageErrorRetrievingAccounts means: storage had error retrieving accounts
	ErrStorageErrorRetrievingAccounts = errors.New("storage had error retrieving accounts")
)

// Operand represents an operand in a Condition. It can be one of three things:
//
// 1) a COIN, which represents a market e.g. BTC/USDT
// 2) a MARKETCAP, which represents the market capitalization of a crypto asset e.g. the marketcap of BTC
// 3) a NUMBER, which represents a literal number e.g. 1.234
//
// Operands are used in expressions together with an Operator to declare a Condition, e.g. BTC/USDT >= 45000.
type Operand struct {
	Type       OperandType
	Provider   string      // e.g. "BINANCE", "KUCOIN", must be empty if Type == NUMBER
	BaseAsset  string      // e.g. "BTC" in BTC/USDT, must be empty if Type == NUMBER
	QuoteAsset string      // e.g. "USDT" in BTC/USDT, must be empty if Type in {MARKETCAP, NUMBER}
	Number     JSONFloat64 // e.g. "1.234", must be empty if Type != NUMBER
	Str        string      // e.g. "COIN:BINANCE:BTC-USDT", "MARKETCAP:MESSARI:BTC", "1.234"
}

// ToMarketSource translates an Operand to a struct that the market package can work with.
func (o Operand) ToMarketSource() common.MarketSource {
	return common.MarketSource{Type: o.Type.toMarketType(), Provider: o.Provider, BaseAsset: o.BaseAsset, QuoteAsset: o.QuoteAsset}
}

// OperandType is the type of Operand in a condition. Can be NUMBER|COIN|MARKETCAP
type OperandType int

const (
	// NUMBER is e.g. 1.2345
	NUMBER OperandType = iota
	// COIN is e.g. COIN:BINANCE:BTC-USDT
	COIN
	// MARKETCAP is e.g. MARKETCAP:MESSARI:BTC
	MARKETCAP
)

// OperandTypeFromString constructs an OperandType from a string
func OperandTypeFromString(s string) (OperandType, error) {
	switch s {
	case "NUMBER", "":
		return NUMBER, nil
	case "COIN":
		return COIN, nil
	case "MARKETCAP":
		return MARKETCAP, nil
	default:
		return 0, fmt.Errorf("%w: %v", ErrUnknownOperandType, s)
	}
}
func (v OperandType) String() string {
	switch v {
	case NUMBER:
		return "NUMBER"
	case COIN:
		return "COIN"
	case MARKETCAP:
		return "MARKETCAP"
	default:
		return ""
	}
}
func (v OperandType) toMarketType() common.MarketType {
	switch v {
	case COIN:
		return common.COIN
	default:
		return common.UNSUPPORTED
	}
}

// ConditionState maintains the state of a condition as it is evolved by calling Condition.Run with Ticks.
//
// - LastTs is the last timestamp for the last Ticks supplied in the last Condition.Run invocation.
//
// - LastTicks are the last Ticks supplied in the last Condition.Run invocation.
//
// - Status determines if the Condition has finished evolving or not, and Value determines its result. When Status
//   is not FINISHED, Value must be UNDECIDED.
type ConditionState struct {
	Status    ConditionStatus
	LastTs    int
	LastTicks map[string]Tick
	Value     ConditionStateValue
}

// Clone returns a deep copy of ConditionState that does not share any memory with the original struct.
func (s ConditionState) Clone() ConditionState {
	clonedLastTicks := make(map[string]Tick, len(s.LastTicks))
	for k, v := range s.LastTicks {
		clonedLastTicks[k] = v
	}

	return ConditionState{
		Status:    s.Status,
		LastTs:    s.LastTs,
		LastTicks: clonedLastTicks,
		Value:     s.Value,
	}
}

// BoolOperator is the boolean operation to resolve an expression between conditions in a prediction step.
type BoolOperator int

// BoolOperatorFromString constructs a BoolOperator from a string
func BoolOperatorFromString(s string) (BoolOperator, error) {
	switch s {
	case "LITERAL":
		return LITERAL, nil
	case "AND":
		return AND, nil
	case "OR":
		return OR, nil
	case "NOT":
		return NOT, nil
	default:
		return 0, fmt.Errorf("%w: %v", ErrUnknownBoolOperator, s)
	}
}
func (v BoolOperator) String() string {
	switch v {
	case LITERAL:
		return "LITERAL"
	case AND:
		return "AND"
	case OR:
		return "OR"
	case NOT:
		return "NOT"
	default:
		return "ERROR"
	}
}

// ConditionStatus represents the status of a condition within a prediction step, i.e.: UNSTARTED|STARTED|FINISHED.
type ConditionStatus int

// ConditionStatusFromString constructs a ConditionStatus from a string.
func ConditionStatusFromString(s string) (ConditionStatus, error) {
	switch s {
	case "UNSTARTED", "":
		return UNSTARTED, nil
	case "STARTED":
		return STARTED, nil
	case "FINISHED":
		return FINISHED, nil
	default:
		return 0, fmt.Errorf("%w: %v", ErrUnknownConditionStatus, s)
	}
}
func (v ConditionStatus) String() string {
	switch v {
	case UNSTARTED:
		return "UNSTARTED"
	case STARTED:
		return "STARTED"
	case FINISHED:
		return "FINISHED"
	default:
		return ""
	}
}

// PredictionStateValue represents the value of a prediction, that is, if it was correct or not, or not yet...
type PredictionStateValue int

// PredictionStateValueFromString constructs a PredictionStateValue from a string.
func PredictionStateValueFromString(s string) (PredictionStateValue, error) {
	switch s {
	case "ONGOING_PRE_PREDICTION", "":
		return ONGOINGPREPREDICTION, nil
	case "ONGOING_PREDICTION":
		return ONGOINGPREDICTION, nil
	case "CORRECT":
		return CORRECT, nil
	case "INCORRECT":
		return INCORRECT, nil
	case "ANNULLED":
		return ANNULLED, nil
	default:
		return 0, fmt.Errorf("%w: %v", ErrUnknownPredictionStateValue, s)
	}
}
func (v PredictionStateValue) String() string {
	switch v {
	case ONGOINGPREPREDICTION:
		return "ONGOING_PRE_PREDICTION"
	case ONGOINGPREDICTION:
		return "ONGOING_PREDICTION"
	case CORRECT:
		return "CORRECT"
	case INCORRECT:
		return "INCORRECT"
	case ANNULLED:
		return "ANNULLED"
	default:
		return ""
	}
}

// IsFinal answers if a prediction is either CORRECT, INCORRECT or ANNULLED, but is no longer ongoing.
func (v PredictionStateValue) IsFinal() bool {
	return v != ONGOINGPREPREDICTION && v != ONGOINGPREDICTION
}

const (
	// LITERAL is a BoolOperator
	LITERAL BoolOperator = iota
	// AND is a BoolOperator
	AND
	// OR is a BoolOperator
	OR
	// NOT is a BoolOperator
	NOT
)

const (
	// UNSTARTED is a ConditionStatus
	UNSTARTED ConditionStatus = iota
	// STARTED is a ConditionStatus
	STARTED
	// FINISHED is a ConditionStatus
	FINISHED
)

const (
	// ONGOINGPREPREDICTION is a PredictionStateValue
	ONGOINGPREPREDICTION PredictionStateValue = iota
	// ONGOINGPREDICTION is a PredictionStateValue
	ONGOINGPREDICTION
	// CORRECT is a PredictionStateValue
	CORRECT
	// INCORRECT is a PredictionStateValue
	INCORRECT
	// ANNULLED is a PredictionStateValue
	ANNULLED
)

// PredictionState contains the state of evolving a prediction up to a given date's worth of market data.
// It's not the complete state, as each individual Condition also contains some state.
type PredictionState struct {
	Status ConditionStatus
	LastTs int
	Value  PredictionStateValue
}

// APIFilters is the set of filters for requesting Predictions at API-level and storage-level.
type APIFilters struct {
	Tags                  []string `json:"tags"`
	AuthorHandles         []string `json:"authorHandles"`
	AuthorURLs            []string `json:"authorURLs"`
	UUIDs                 []string `json:"uuids"`
	GreaterThanUUID       string   `json:"greaterThanUUID"`
	URLs                  []string `json:"urls"`
	PredictionStateValues []string `json:"predictionStateValues"`
	PredictionStateStatus []string `json:"predictionStateStatus"`
	Deleted               *bool    `json:"deleted"`
	Paused                *bool    `json:"paused"`
	Hidden                *bool    `json:"hidden"`
	IncludeUIUnsupported  bool     `json:"showUIUnsupported"`
}

// ToQueryStringWithOrderBy converts the struct to a QueryString, including the orderBy parameter
func (f APIFilters) ToQueryStringWithOrderBy(orderBy []string) map[string][]string {
	m := map[string][]string{
		"tags":                  f.Tags,
		"authorHandles":         f.AuthorHandles,
		"authorURLs":            f.AuthorURLs,
		"uuids":                 f.UUIDs,
		"urls":                  f.URLs,
		"predictionStateValues": f.PredictionStateValues,
		"predictionStateStatus": f.PredictionStateStatus,
		"orderBys":              orderBy,
	}
	if f.Deleted != nil {
		m["deleted"] = []string{"false"}
		if *f.Deleted {
			m["deleted"] = []string{"true"}
		}
	}
	if f.Paused != nil {
		m["paused"] = []string{"false"}
		if *f.Paused {
			m["paused"] = []string{"true"}
		}
	}
	if f.Hidden != nil {
		m["hidden"] = []string{"false"}
		if *f.Hidden {
			m["hidden"] = []string{"true"}
		}
	}
	if f.IncludeUIUnsupported {
		m["includeUIUnsupported"] = []string{"true"}
	}

	return m
}

// APIAccountFilters are the available filters for requesting Accounts, at both API & database levels.
type APIAccountFilters struct {
	Handles []string `json:"handles"`
	URLs    []string `json:"urls"`
}

// APIAccountOrderBy are the possible ways to order a list of accounts at API & database levels.
type APIAccountOrderBy int

const (
	// AccountCreatedAtDesc is used to sort a list of Accounts by created_at desc
	AccountCreatedAtDesc APIAccountOrderBy = iota
	// AccountCreatedAtAsc is used to sort a list of Accounts by created_at asc
	AccountCreatedAtAsc
	// AccountFollowerCountDesc is used to sort a list of Accounts by follower_count desc
	AccountFollowerCountDesc
)

// APIAccountOrderByFromString constructs a APIAccountOrderBy from a string.
func APIAccountOrderByFromString(s string) (APIAccountOrderBy, error) {
	switch s {
	case "ACCOUNT_CREATED_AT_DESC", "":
		return AccountCreatedAtDesc, nil
	case "ACCOUNT_CREATED_AT_ASC":
		return AccountCreatedAtAsc, nil
	case "ACCOUNT_FOLLOWER_COUNT_DESC":
		return AccountFollowerCountDesc, nil
	default:
		return 0, fmt.Errorf("%w: %v", ErrUnknownAPIOrderBy, s)
	}
}
func (v APIAccountOrderBy) String() string {
	switch v {
	case AccountCreatedAtDesc:
		return "ACCOUNT_CREATED_AT_DESC"
	case AccountCreatedAtAsc:
		return "ACCOUNT_CREATED_AT_ASC"
	case AccountFollowerCountDesc:
		return "ACCOUNT_FOLLOWER_COUNT_DESC"
	default:
		return ""
	}
}

// APIPredictionsOrderBy are the possible ways to order a list of predictions at API & database levels.
type APIPredictionsOrderBy int

const (
	// PredictionsCreatedAtDesc sorts a list of predictions by createdAt desc
	PredictionsCreatedAtDesc APIPredictionsOrderBy = iota
	// PredictionsCreatedAtAsc sorts a list of predictions by createdAt desc
	PredictionsCreatedAtAsc
	// PredictionsPostedAtDesc sorts a list of predictions by postedAt desc
	PredictionsPostedAtDesc
	// PredictionsPostedAtAsc sorts a list of predictions by postedAt asc
	PredictionsPostedAtAsc
	// PredictionsUUIDAsc sorts a list of predictions by UUID asc
	PredictionsUUIDAsc
)

// APIPredictionsOrderByFromString constructs a APIPredictionsOrderBy from a string
func APIPredictionsOrderByFromString(s string) (APIPredictionsOrderBy, error) {
	switch s {
	case "CREATED_AT_DESC", "":
		return PredictionsCreatedAtDesc, nil
	case "CREATED_AT_ASC":
		return PredictionsCreatedAtAsc, nil
	case "POSTED_AT_DESC":
		return PredictionsPostedAtDesc, nil
	case "POSTED_AT_ASC":
		return PredictionsPostedAtAsc, nil
	case "UUID_ASC":
		return PredictionsUUIDAsc, nil
	default:
		return 0, fmt.Errorf("%w: %v", ErrUnknownAPIOrderBy, s)
	}
}
func (v APIPredictionsOrderBy) String() string {
	switch v {
	case PredictionsCreatedAtDesc:
		return "CREATED_AT_DESC"
	case PredictionsCreatedAtAsc:
		return "CREATED_AT_ASC"
	case PredictionsPostedAtDesc:
		return "POSTED_AT_DESC"
	case PredictionsPostedAtAsc:
		return "POSTED_AT_ASC"
	case PredictionsUUIDAsc:
		return "UUID_ASC"
	default:
		return ""
	}
}

// JSONFloat64 exists only for the purpose of marshalling floats in a nicer way.
type JSONFloat64 float64

// MarshalJSON overrides the marshalling of floats in a nicer way.
func (jf JSONFloat64) MarshalJSON() ([]byte, error) {
	f := float64(jf)
	if math.IsInf(f, 0) || math.IsNaN(f) {
		return nil, errors.New("unsupported value")
	}
	bs := []byte(fmt.Sprintf("%.12f", f))
	var i int
	for i = len(bs) - 1; i >= 0; i-- {
		if bs[i] == '0' {
			continue
		}
		if bs[i] == '.' {
			return bs[:i], nil
		}
		break
	}
	return bs[:i+1], nil
}

// ISO8601 adds convenience methods for converting ISO8601-formatted date strings.
type ISO8601 string

// Time converts an ISO8601-formatted date string into a time.Time.
func (t ISO8601) Time() (time.Time, error) {
	return time.Parse(time.RFC3339, string(t))
}

// Seconds converts an ISO8601-formatted date string into a Unix timestamp.
func (t ISO8601) Seconds() (int, error) {
	tm, err := t.Time()
	if err != nil {
		return 0, fmt.Errorf("failed to convert %v to seconds because %v", string(t), err.Error())
	}
	return int(tm.Unix()), nil
}

// Millis converts an ISO8601-formatted date string into a Javascript millisecond timestamp.
func (t ISO8601) Millis() (int, error) {
	tm, err := t.Seconds()
	if err != nil {
		return 0, err
	}
	return tm * 100, nil
}

// PredictionStateValueChange represents a database-row for the event of a prediction changing value.
type PredictionStateValueChange struct {
	PredictionUUID string
	StateValue     string
	CreatedAt      ISO8601
}

// Account represents a post author's social media account, both at the API & database-level.
type Account struct {
	URL           *url.URL   `json:"url"`
	AccountType   string     `json:"account_type"`
	Handle        string     `json:"handle"`
	FollowerCount int        `json:"follower_count"`
	Thumbnails    []*url.URL `json:"thumbnails"`
	Name          string     `json:"name"`
	Description   string     `json:"description"`
	CreatedAt     *time.Time `json:"created_at"`
	IsVerified    bool       `json:"is_verified"`
}

// PredictionType identifies the structure of a prediction. It is used conditionally to decide how to pretty-print it
// and produce overlays for charting it.
type PredictionType int

const (
	// PredictionTypeUnsupported is the type of prediction used when this prediction is unsupported.
	PredictionTypeUnsupported PredictionType = iota

	// PredictionTypeCoinOperatorFloatDeadline is a type of prediction that looks like this:
	// "Bitcoin will be below $17k in 2 weeks"
	PredictionTypeCoinOperatorFloatDeadline

	// PredictionTypeCoinWillRange is a type of prediction that looks like this:
	// "Bitcoin will range between $17k and $20k for 2 weeks"
	PredictionTypeCoinWillRange

	// PredictionTypeCoinWillReachBeforeItReaches is a type of prediction that looks like this:
	// "Bitcoin will reach $17k before it reaches $20k"
	PredictionTypeCoinWillReachBeforeItReaches

	// PredictionTypeCoinWillReachInvalidatedIfItReaches is a type of prediction that looks like this:
	// "If Bitcoin doesn't fall below $10k, it will reach $60k within 3 months"
	PredictionTypeCoinWillReachInvalidatedIfItReaches
)

// PredictionTypeFromString constructs a PredictionType from a string.
func PredictionTypeFromString(s string) PredictionType {
	switch s {
	case "PREDICTION_TYPE_COIN_OPERATOR_FLOAT_DEADLINE", "":
		return PredictionTypeCoinOperatorFloatDeadline
	case "PREDICTION_TYPE_COIN_WILL_RANGE":
		return PredictionTypeCoinWillRange
	case "PREDICTION_TYPE_COIN_WILL_REACH_BEFORE_IT_REACHES":
		return PredictionTypeCoinWillReachBeforeItReaches
	case "PREDICTION_TYPE_COIN_WILL_REACH_INVALIDATED_IF_IT_REACHES":
		return PredictionTypeCoinWillReachInvalidatedIfItReaches
	default:
		return PredictionTypeUnsupported
	}
}
func (v PredictionType) String() string {
	switch v {
	case PredictionTypeCoinOperatorFloatDeadline:
		return "PREDICTION_TYPE_COIN_OPERATOR_FLOAT_DEADLINE"
	case PredictionTypeCoinWillRange:
		return "PREDICTION_TYPE_COIN_WILL_RANGE"
	case PredictionTypeCoinWillReachBeforeItReaches:
		return "PREDICTION_TYPE_COIN_WILL_REACH_BEFORE_IT_REACHES"
	case PredictionTypeCoinWillReachInvalidatedIfItReaches:
		return "PREDICTION_TYPE_COIN_WILL_REACH_INVALIDATED_IF_IT_REACHES"
	default:
		return "PREDICTION_TYPE_UNSUPPORTED"
	}
}

// PredictionInteraction is a social media post posted by this engine upon an event of a prediction.
type PredictionInteraction struct {
	PostURL            string
	ActionType         string
	InteractionPostURL string
	PredictionUUID     string
	Status             string // PENDING, POSTED, ERROR
	Error              string
}

// IMarket is an interface to candles.Market just to be able to mock it for tests
type IMarket interface {
	Iterator(marketSource common.MarketSource, startTime time.Time, candlestickInterval time.Duration) (iterator.Iterator, error)
}

// Tick is a subset of a candlestick for one of the 4 available prices
type Tick struct {
	Timestamp int                `json:"t"`
	Value     common.JSONFloat64 `json:"v"`
}
