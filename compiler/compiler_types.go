package compiler

import (
	"github.com/marianogappa/predictions/types"
)

type ConditionState struct {
	Status    string                `json:"status" enum:"UNSTARTED,STARTED,FINISHED" example:"STARTED"`
	LastTs    int                   `json:"lastTs" example:"1649594376"`
	LastTicks map[string]types.Tick `json:"lastTicks"`
	Value     string                `json:"value" enum:"UNDECIDED,TRUE,FALSE" example:"UNDECIDED"`
}

type PredictionState struct {
	Status string `json:"status" enum:"UNSTARTED,STARTED,FINISHED" example:"FINISHED"`
	LastTs int    `json:"lastTs" example:"1649594376"`
	Value  string `json:"value" enum:"ONGOING_PRE_PREDICTION,ONGOING_PREDICTION,CORRECT,INCORRECT,ANNULLED" example:"CORRECT"`
}

type Condition struct {
	Condition        string         `json:"condition" required:"true" example:"COIN:BINANCE:BTC-USDT <= 29000"`
	FromISO8601      types.ISO8601  `json:"fromISO8601" format:"date-time" example:"2022-01-26T22:35:43Z"`
	ToISO8601        types.ISO8601  `json:"toISO8601" format:"date-time" example:"2023-01-01T00:00:00Z"`
	ToDuration       string         `json:"toDuration" example:"eoy"`
	Assumed          []string       `json:"assumed" example:"[\"toDuration\"]"`
	State            ConditionState `json:"state"`
	ErrorMarginRatio float64        `json:"errorMarginRatio" example:"0.03"`
}

type PrePredict struct {
	WrongIf                           *string `json:"wrongIf,omitempty" example:"a or (b and c)"`
	AnnulledIf                        *string `json:"annulledIf,omitempty" example:"not d"`
	Predict                           *string `json:"predict,omitempty" example:"a and b and c"`
	AnnulledIfPredictIsFalse          bool    `json:"annulledIfPredictIsFalse,omitempty" example:"false"`
	IgnoreUndecidedIfPredictIsDefined bool    `json:"ignoreUndecidedIfPredictIsDefined,omitempty" example:"true"`
}

type Predict struct {
	WrongIf                           *string `json:"wrongIf,omitempty" example:"a or (b and c)"`
	AnnulledIf                        *string `json:"annulledIf,omitempty" example:"not d"`
	Predict                           string  `json:"predict,omitempty" required:"true" example:"a and b and c"`
	IgnoreUndecidedIfPredictIsDefined bool    `json:"ignoreUndecidedIfPredictIsDefined,omitempty" example:"false"`
}

type Prediction struct {
	UUID            string               `json:"uuid" format:"uuid" description:"Prediction's primary key." example:"3a7bc95e-480d-4232-8e8b-f848d5389806"`
	Version         string               `json:"version" description:"Set to 1.0.0." example:"1.0.0"`
	CreatedAt       types.ISO8601        `json:"createdAt" description:"This is automatically calculated when created." format:"date-time" example:"2022-01-26T22:35:43Z"`
	Reporter        string               `json:"reporter" description:"Set to admin for now." required:"true" minLength:"1" example:"admin"`
	PostAuthor      string               `json:"postAuthor" description:"This is automatically calculated based on the tweet/youtube video in postUrl." example:"CryptoCapo_"`
	PostAuthorURL   string               `json:"postAuthorURL,omitempty" description:"This is automatically calculated based on the tweet/youtube video in postUrl." example:"https://twitter.com/CryptoCapo_"`
	PostedAt        types.ISO8601        `json:"postedAt" description:"This is automatically calculated based on the tweet/youtube video in postUrl." format:"date-time" example:"2022-01-26T22:35:43Z"`
	PostUrl         string               `json:"postUrl" description:"The tweet/youtube video's URL." required:"true" format:"uri" example:"https://twitter.com/CryptoCapo_/status/1486467919064322051"`
	Given           map[string]Condition `json:"given,omitempty" description:"The conditions that form this prediction." required:"true"`
	PrePredict      *PrePredict          `json:"prePredict,omitempty"`
	Predict         Predict              `json:"predict,omitempty" required:"true"`
	PredictionState PredictionState      `json:"state"`
	Type            string               `json:"type" description:"This is automatically calculated based on the prediction's structure." enum:"PREDICTION_TYPE_UNSUPPORTED,PREDICTION_TYPE_COIN_OPERATOR_FLOAT_DEADLINE,PREDICTION_TYPE_COIN_WILL_RANGE,PREDICTION_TYPE_COIN_WILL_REACH_BEFORE_IT_REACHES,PREDICTION_TYPE_THE_FLIPPENING" example:"PREDICTION_TYPE_COIN_OPERATOR_FLOAT_DEADLINE"`

	// extra fields for API, but not for Postgres
	PredictionText string            `json:"predictionText,omitempty" example:"COIN:BINANCE:BTC-USDT will hit 29000 by end of year"`
	Summary        PredictionSummary `json:"summary,omitempty"`
}
