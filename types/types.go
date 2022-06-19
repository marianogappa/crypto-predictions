package types

import (
	"errors"
	"fmt"
	"math"
	"net/url"
	"time"

	"github.com/google/uuid"
)

var (
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

	// ErrOutOfTicks means: out of ticks
	ErrOutOfTicks = errors.New("out of ticks")

	// ErrOutOfCandlesticks means: exchange ran out of candlesticks
	ErrOutOfCandlesticks = errors.New("exchange ran out of candlesticks")

	// ErrOutOfTrades means: exchange ran out of trades
	ErrOutOfTrades = errors.New("exchange ran out of trades")

	// ErrInvalidMarketPair means: market pair or asset does not exist on exchange
	ErrInvalidMarketPair = errors.New("market pair or asset does not exist on exchange")

	// ErrRateLimit means: exchange asked us to enhance our calm
	ErrRateLimit = errors.New("exchange asked us to enhance our calm")

	// ErrInvalidExchange means: the only valid exchanges are 'binance', 'ftx', 'coinbase', 'huobi', 'kraken', 'kucoin' and 'binanceusdmfutures'
	ErrInvalidExchange = errors.New("the only valid exchanges are 'binance', 'ftx', 'coinbase', 'huobi', 'kraken', 'kucoin' and 'binanceusdmfutures'")

	// ErrBaseAssetRequired means: base asset is required (e.g. BTC)
	ErrBaseAssetRequired = errors.New("base asset is required (e.g. BTC)")

	// ErrQuoteAssetRequired means: quote asset is required (e.g. USDT)
	ErrQuoteAssetRequired = errors.New("quote asset is required (e.g. USDT)")

	// From TickIterator

	// ErrNoNewTicksYet means: no new ticks yet
	ErrNoNewTicksYet = errors.New("no new ticks yet")

	// ErrExchangeReturnedNoTicks means: exchange returned no ticks
	ErrExchangeReturnedNoTicks = errors.New("exchange returned no ticks")

	// ErrExchangeReturnedOutOfSyncTick means: exchange returned out of sync tick
	ErrExchangeReturnedOutOfSyncTick = errors.New("exchange returned out of sync tick")

	// From PatchTickHoles

	// ErrOutOfSyncTimestampPatchingHoles means: out of sync timestamp found patching holes
	ErrOutOfSyncTimestampPatchingHoles = errors.New("out of sync timestamp found patching holes")

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
	Number     JsonFloat64 // e.g. "1.234", must be empty if Type != NUMBER
	Str        string      // e.g. "COIN:BINANCE:BTC-USDT", "MARKETCAP:MESSARI:BTC", "1.234"
}

type OperandType int

const (
	NUMBER OperandType = iota
	COIN
	MARKETCAP
)

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

type BoolOperator int

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

type ConditionStatus int

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

type PredictionStateValue int

func PredictionStateValueFromString(s string) (PredictionStateValue, error) {
	switch s {
	case "ONGOING_PRE_PREDICTION", "":
		return ONGOING_PRE_PREDICTION, nil
	case "ONGOING_PREDICTION":
		return ONGOING_PREDICTION, nil
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
	case ONGOING_PRE_PREDICTION:
		return "ONGOING_PRE_PREDICTION"
	case ONGOING_PREDICTION:
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

func (v PredictionStateValue) IsFinal() bool {
	return v != ONGOING_PRE_PREDICTION && v != ONGOING_PREDICTION
}

const (
	LITERAL BoolOperator = iota
	AND
	OR
	NOT
)

const (
	UNSTARTED ConditionStatus = iota
	STARTED
	FINISHED
)

const (
	ONGOING_PRE_PREDICTION PredictionStateValue = iota
	ONGOING_PREDICTION
	CORRECT
	INCORRECT
	ANNULLED
)

type PredictionState struct {
	Status ConditionStatus
	LastTs int
	Value  PredictionStateValue
	// add state to provide evidence of alleged condition result
}

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
}

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

	return m
}

type APIAccountFilters struct {
	Handles []string `json:"handles"`
	URLs    []string `json:"urls"`
}

type APIAccountOrderBy int

const (
	ACCOUNT_CREATED_AT_DESC APIAccountOrderBy = iota
	ACCOUNT_CREATED_AT_ASC
	ACCOUNT_FOLLOWER_COUNT_DESC
)

func APIAccountOrderByFromString(s string) (APIAccountOrderBy, error) {
	switch s {
	case "ACCOUNT_CREATED_AT_DESC", "":
		return ACCOUNT_CREATED_AT_DESC, nil
	case "ACCOUNT_CREATED_AT_ASC":
		return ACCOUNT_CREATED_AT_ASC, nil
	case "ACCOUNT_FOLLOWER_COUNT_DESC":
		return ACCOUNT_FOLLOWER_COUNT_DESC, nil
	default:
		return 0, fmt.Errorf("%w: %v", ErrUnknownAPIOrderBy, s)
	}
}
func (v APIAccountOrderBy) String() string {
	switch v {
	case ACCOUNT_CREATED_AT_DESC:
		return "ACCOUNT_CREATED_AT_DESC"
	case ACCOUNT_CREATED_AT_ASC:
		return "ACCOUNT_CREATED_AT_ASC"
	case ACCOUNT_FOLLOWER_COUNT_DESC:
		return "ACCOUNT_FOLLOWER_COUNT_DESC"
	default:
		return ""
	}
}

type APIOrderBy int

const (
	CREATED_AT_DESC APIOrderBy = iota
	CREATED_AT_ASC
	POSTED_AT_DESC
	POSTED_AT_ASC
	UUID_ASC
)

func APIOrderByFromString(s string) (APIOrderBy, error) {
	switch s {
	case "CREATED_AT_DESC", "":
		return CREATED_AT_DESC, nil
	case "CREATED_AT_ASC":
		return CREATED_AT_ASC, nil
	case "POSTED_AT_DESC":
		return POSTED_AT_DESC, nil
	case "POSTED_AT_ASC":
		return POSTED_AT_ASC, nil
	case "UUID_ASC":
		return UUID_ASC, nil
	default:
		return 0, fmt.Errorf("%w: %v", ErrUnknownAPIOrderBy, s)
	}
}
func (v APIOrderBy) String() string {
	switch v {
	case CREATED_AT_DESC:
		return "CREATED_AT_DESC"
	case CREATED_AT_ASC:
		return "CREATED_AT_ASC"
	case POSTED_AT_DESC:
		return "POSTED_AT_DESC"
	case POSTED_AT_ASC:
		return "POSTED_AT_ASC"
	case UUID_ASC:
		return "UUID_ASC"
	default:
		return ""
	}
}

type JsonFloat64 float64

func (jf JsonFloat64) MarshalJSON() ([]byte, error) {
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

type ISO8601 string

func (t ISO8601) Time() (time.Time, error) {
	return time.Parse(time.RFC3339, string(t))
}

func (t ISO8601) Seconds() (int, error) {
	tm, err := t.Time()
	if err != nil {
		return 0, fmt.Errorf("failed to convert %v to seconds because %v", string(t), err.Error())
	}
	return int(tm.Unix()), nil
}

func (t ISO8601) Millis() (int, error) {
	tm, err := t.Seconds()
	if err != nil {
		return 0, err
	}
	return tm * 100, nil
}

// Candlestick is the generic struct for candlestick data for all supported exchanges.
type Candlestick struct {
	// Timestamp is the UNIX timestamp (i.e. seconds since UTC Epoch) at which the candlestick started.
	Timestamp int `json:"t"`

	// OpenPrice is the price at which the candlestick opened.
	OpenPrice JsonFloat64 `json:"o"`

	// ClosePrice is the price at which the candlestick closed.
	ClosePrice JsonFloat64 `json:"c"`

	// LowestPrice is the lowest price reached during the candlestick duration.
	LowestPrice JsonFloat64 `json:"l"`

	// HighestPrice is the highest price reached during the candlestick duration.
	HighestPrice JsonFloat64 `json:"h"`

	// Volume is the traded volume in base asset during this candlestick.
	Volume JsonFloat64 `json:"v"`

	// NumberOfTrades is the total number of filled order book orders in this candlestick.
	NumberOfTrades int `json:"n,omitempty"`
}

// ToTicks converts a Candlestick to two Ticks. Lowest value is put first, because since there's no way to tell
// which one happened first, this library chooses to be pessimistic.
func (c Candlestick) ToTicks() []Tick {
	return []Tick{
		{Timestamp: c.Timestamp, Value: c.LowestPrice},
		{Timestamp: c.Timestamp, Value: c.HighestPrice},
	}
}

type PredictionStateValueChange struct {
	PredictionUUID string
	StateValue     string
	CreatedAt      ISO8601
}

func (p PredictionStateValueChange) PK() string {
	return fmt.Sprintf("%v|%v", p.PredictionUUID, p.StateValue)
}

func (p PredictionStateValueChange) Validate() error {
	if _, err := p.CreatedAt.Time(); err != nil {
		return errors.New("CreatedAt is an invalid ISO8601")
	}
	if _, err := PredictionStateValueFromString(p.StateValue); err != nil {
		return err
	}
	if _, err := uuid.Parse(p.PredictionUUID); err != nil {
		return err
	}
	return nil
}

type Tick struct {
	Timestamp int         `json:"t"`
	Value     JsonFloat64 `json:"v"`
}
type Iterator interface {
	NextTick() (Tick, error)
	NextCandlestick() (Candlestick, error)
}

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

type PredictionType int

const (
	PREDICTION_TYPE_UNSUPPORTED PredictionType = iota
	PREDICTION_TYPE_COIN_OPERATOR_FLOAT_DEADLINE
	PREDICTION_TYPE_COIN_WILL_RANGE
	PREDICTION_TYPE_COIN_WILL_REACH_BEFORE_IT_REACHES
	PREDICTION_TYPE_THE_FLIPPENING
	PREDICTION_TYPE_COIN_WILL_REACH_INVALIDATED_IF_IT_REACHES
)

func PredictionTypeFromString(s string) PredictionType {
	switch s {
	case "PREDICTION_TYPE_COIN_OPERATOR_FLOAT_DEADLINE", "":
		return PREDICTION_TYPE_COIN_OPERATOR_FLOAT_DEADLINE
	case "PREDICTION_TYPE_COIN_WILL_RANGE":
		return PREDICTION_TYPE_COIN_WILL_RANGE
	case "PREDICTION_TYPE_COIN_WILL_REACH_BEFORE_IT_REACHES":
		return PREDICTION_TYPE_COIN_WILL_REACH_BEFORE_IT_REACHES
	case "PREDICTION_TYPE_THE_FLIPPENING":
		return PREDICTION_TYPE_THE_FLIPPENING
	case "PREDICTION_TYPE_COIN_WILL_REACH_INVALIDATED_IF_IT_REACHES":
		return PREDICTION_TYPE_COIN_WILL_REACH_INVALIDATED_IF_IT_REACHES
	default:
		return PREDICTION_TYPE_UNSUPPORTED
	}
}
func (v PredictionType) String() string {
	switch v {
	case PREDICTION_TYPE_COIN_OPERATOR_FLOAT_DEADLINE:
		return "PREDICTION_TYPE_COIN_OPERATOR_FLOAT_DEADLINE"
	case PREDICTION_TYPE_COIN_WILL_RANGE:
		return "PREDICTION_TYPE_COIN_WILL_RANGE"
	case PREDICTION_TYPE_COIN_WILL_REACH_BEFORE_IT_REACHES:
		return "PREDICTION_TYPE_COIN_WILL_REACH_BEFORE_IT_REACHES"
	case PREDICTION_TYPE_THE_FLIPPENING:
		return "PREDICTION_TYPE_THE_FLIPPENING"
	case PREDICTION_TYPE_COIN_WILL_REACH_INVALIDATED_IF_IT_REACHES:
		return "PREDICTION_TYPE_COIN_WILL_REACH_INVALIDATED_IF_IT_REACHES"
	default:
		return "PREDICTION_TYPE_UNSUPPORTED"
	}
}
