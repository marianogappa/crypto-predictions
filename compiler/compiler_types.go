package compiler

import (
	"github.com/marianogappa/predictions/market/common"
	"github.com/marianogappa/predictions/types"
	"github.com/swaggest/jsonschema-go"
)

// ConditionState holds the state of evolving a condition using market data.
type ConditionState struct {
	Status    string                 `json:"status" enum:"UNSTARTED,STARTED,FINISHED" example:"STARTED"`
	LastTs    int                    `json:"lastTs" example:"1649594376"`
	LastTicks map[string]common.Tick `json:"lastTicks"`
	Value     string                 `json:"value" enum:"UNDECIDED,TRUE,FALSE" example:"UNDECIDED"`
}

// PredictionState holds the state of evolving a prediction using market data.
type PredictionState struct {
	Status string `json:"status" enum:"UNSTARTED,STARTED,FINISHED" example:"FINISHED"`
	LastTs int    `json:"lastTs" example:"1649594376"`
	Value  string `json:"value" enum:"ONGOING_PRE_PREDICTION,ONGOING_PREDICTION,CORRECT,INCORRECT,ANNULLED" example:"CORRECT"`
}

// Condition is each of the conditions that form a prediction, e.g. "COIN:BINANCE:BTC-USDT <= 29000 within 3 weeks".
type Condition struct {
	Condition        string         `json:"condition" required:"true" example:"COIN:BINANCE:BTC-USDT <= 29000"`
	FromISO8601      types.ISO8601  `json:"fromISO8601" format:"date-time" example:"2022-01-26T22:35:43Z"`
	ToISO8601        types.ISO8601  `json:"toISO8601" format:"date-time" example:"2023-01-01T00:00:00Z"`
	ToDuration       string         `json:"toDuration" example:"eoy"`
	Assumed          []string       `json:"assumed" example:"[\"toDuration\"]"`
	State            ConditionState `json:"state"`
	ErrorMarginRatio float64        `json:"errorMarginRatio" example:"0.03"`
}

// PrePredict is a subpart of a Prediction that represents an initial step that is required for a two-step prediction.
type PrePredict struct {
	WrongIf                           *string `json:"wrongIf,omitempty" example:"a or (b and c)"`
	AnnulledIf                        *string `json:"annulledIf,omitempty" example:"not d"`
	Predict                           *string `json:"predict,omitempty" example:"a and b and c"`
	AnnulledIfPredictIsFalse          bool    `json:"annulledIfPredictIsFalse,omitempty" example:"false"`
	IgnoreUndecidedIfPredictIsDefined bool    `json:"ignoreUndecidedIfPredictIsDefined,omitempty" example:"true"`
}

// Predict is the main prediction (considering it could be a two-step prediction, in which case this is the final one).
// A prediction is represented as a boolean expression involving variable names that represent Conditions.
type Predict struct {
	WrongIf                           *string `json:"wrongIf,omitempty" example:"a or (b and c)"`
	AnnulledIf                        *string `json:"annulledIf,omitempty" example:"not d"`
	Predict                           string  `json:"predict,omitempty" required:"true" example:"a and b and c"`
	IgnoreUndecidedIfPredictIsDefined bool    `json:"ignoreUndecidedIfPredictIsDefined,omitempty" example:"false"`
}

// Prediction is the main struct that represents a Prediction, containing all properties that must be persisted.
type Prediction struct {
	UUID            string               `json:"uuid" format:"uuid" description:"Prediction's primary key." example:"3a7bc95e-480d-4232-8e8b-f848d5389806"`
	Version         string               `json:"version" description:"Set to 1.0.0." example:"1.0.0"`
	CreatedAt       types.ISO8601        `json:"createdAt" description:"This is automatically calculated when created." format:"date-time" example:"2022-01-26T22:35:43Z"`
	Reporter        string               `json:"reporter" description:"Set to admin for now." required:"true" minLength:"1" example:"admin"`
	PostAuthor      string               `json:"postAuthor" description:"This is automatically calculated based on the tweet/youtube video in postUrl." example:"CryptoCapo_"`
	PostAuthorURL   string               `json:"postAuthorURL,omitempty" description:"This is automatically calculated based on the tweet/youtube video in postUrl." example:"https://twitter.com/CryptoCapo_"`
	PostedAt        types.ISO8601        `json:"postedAt" description:"This is automatically calculated based on the tweet/youtube video in postUrl." format:"date-time" example:"2022-01-26T22:35:43Z"`
	PostURL         string               `json:"postUrl" description:"The tweet/youtube video's URL." required:"true" format:"uri" example:"https://twitter.com/CryptoCapo_/status/1486467919064322051"`
	Given           map[string]Condition `json:"given,omitempty" description:"The conditions that form this prediction." required:"true"`
	PrePredict      *PrePredict          `json:"prePredict,omitempty"`
	Predict         Predict              `json:"predict,omitempty" required:"true"`
	PredictionState PredictionState      `json:"state"`
	Type            string               `json:"type" description:"This is automatically calculated based on the prediction's structure." enum:"PREDICTION_TYPE_UNSUPPORTED,PREDICTION_TYPE_COIN_OPERATOR_FLOAT_DEADLINE,PREDICTION_TYPE_COIN_WILL_RANGE,PREDICTION_TYPE_COIN_WILL_REACH_BEFORE_IT_REACHES,PREDICTION_TYPE_THE_FLIPPENING" example:"PREDICTION_TYPE_COIN_OPERATOR_FLOAT_DEADLINE"`

	// extra fields for API, but not for Postgres
	PredictionText string            `json:"predictionText,omitempty" example:"COIN:BINANCE:BTC-USDT will hit 29000 by end of year"`
	Summary        PredictionSummary `json:"summary,omitempty"`
	Deleted        bool              `json:"deleted,omitempty"`
	Paused         bool              `json:"paused,omitempty"`
	Hidden         bool              `json:"hidden,omitempty"`
}

// PredictionSummary contains all necessary information about the prediction to make a candlestick chart of it.
type PredictionSummary struct {
	// Only in "PredictionTheFlippening" type
	OtherCoin string `json:"otherCoin,omitempty"`

	// Only in "PredictionTypeCoinOperatorFloatDeadline" type
	Goal                                    types.JSONFloat64 `json:"goal,omitempty"`
	GoalWithError                           types.JSONFloat64 `json:"goalWithError,omitempty"`
	EndedAtTruncatedDueToResultInvalidation types.ISO8601     `json:"endedAtTruncatedDueToResultInvalidation,omitempty"`

	// Only in "PredictionTypeCoinWillReachInvalidatedIfItReaches"
	InvalidatedIfItReaches types.JSONFloat64 `json:"invalidatedIfItReaches,omitempty"`

	// Only in "PredictionWillRange type"
	RangeLow           types.JSONFloat64 `json:"rangeLow,omitempty"`
	RangeLowWithError  types.JSONFloat64 `json:"rangeLowWithError,omitempty"`
	RangeHigh          types.JSONFloat64 `json:"rangeHigh,omitempty"`
	RangeHighWithError types.JSONFloat64 `json:"rangeHighWithError,omitempty"`

	// Only in "PredictionWillReachBeforeItReaches type"
	WillReach                types.JSONFloat64 `json:"willReach,omitempty"`
	WillReachWithError       types.JSONFloat64 `json:"willReachWithError,omitempty"`
	BeforeItReaches          types.JSONFloat64 `json:"beforeItReaches,omitempty"`
	BeforeItReachesWithError types.JSONFloat64 `json:"beforeItReachesWithError,omitempty"`

	// In all prediction types
	CandlestickMap   map[string][]common.Candlestick `json:"candlestickMap,omitempty"`
	Coin             string                          `json:"coin,omitempty"`
	ErrorMarginRatio types.JSONFloat64               `json:"errorMarginRatio,omitempty"`
	Operator         string                          `json:"operator,omitempty"`
	Deadline         types.ISO8601                   `json:"deadline,omitempty"`
	EndedAt          types.ISO8601                   `json:"endedAt,omitempty"`
	PredictionType   string                          `json:"predictionType,omitempty"`
}

// PrepareJSONSchema provides an example of the structure for Swagger docs
func (PredictionSummary) PrepareJSONSchema(schema *jsonschema.Schema) error {
	schema.WithExamples(PredictionSummary{
		CandlestickMap: map[string][]common.Candlestick{
			"BINANCE:COIN:BTC-USDT": {
				{Timestamp: 1651161957, OpenPrice: 39000, HighestPrice: 39500, LowestPrice: 39000, ClosePrice: 39050},
				{Timestamp: 1651162017, OpenPrice: 39500, HighestPrice: 39550, LowestPrice: 39200, ClosePrice: 39020},
			},
		},
		Coin:                                    "BINANCE:COIN:BTC-USDT",
		Goal:                                    45000,
		GoalWithError:                           43650,
		ErrorMarginRatio:                        0.03,
		Operator:                                ">=",
		Deadline:                                "2022-06-24T07:51:06Z",
		EndedAt:                                 "2022-06-24T07:51:06Z",
		EndedAtTruncatedDueToResultInvalidation: "2022-06-23T00:00:00Z",
		PredictionType:                          "PREDICTION_TYPE_COIN_OPERATOR_FLOAT_DEADLINE",
	})

	return nil
}
