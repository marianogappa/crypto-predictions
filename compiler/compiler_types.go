package compiler

import (
	"github.com/marianogappa/predictions/types"
	"github.com/marianogappa/signal-checker/common"
)

type conditionState struct {
	Status    string                `json:"status"`
	LastTs    int                   `json:"lastTs"`
	LastTicks map[string]types.Tick `json:"lastTicks"`
	Value     string                `json:"value"`
	// add state to provide evidence of alleged condition result
}

type predictionState struct {
	Status string `json:"status"`
	LastTs int    `json:"lastTs"`
	Value  string `json:"value"`
	// add state to provide evidence of alleged condition result
}

type condition struct {
	Condition        string         `json:"condition"`
	FromISO8601      common.ISO8601 `json:"fromISO8601"`
	ToISO8601        common.ISO8601 `json:"toISO8601"`
	ToDuration       string         `json:"toDuration"`
	Assumed          []string       `json:"assumed"`
	State            conditionState `json:"state"`
	ErrorMarginRatio float64        `json:"errorMarginRatio"`
}

type prePredict struct {
	WrongIf    *string `json:"wrongIf,omitempty"`
	AnnulledIf *string `json:"annulledIf,omitempty"`
	PredictIf  *string `json:"predictIf,omitempty"`
}

type predict struct {
	WrongIf    *string `json:"wrongIf,omitempty"`
	AnnulledIf *string `json:"annulledIf,omitempty"`
	Predict    string  `json:"predict,omitempty"`
}

type prediction struct {
	UUID            string               `json:"uuid"`
	Version         string               `json:"version"`
	CreatedAt       common.ISO8601       `json:"createdAt"`
	PostAuthor      string               `json:"postAuthor"`
	PostedAt        common.ISO8601       `json:"postedAt"`
	PostUrl         string               `json:"postUrl"`
	Given           map[string]condition `json:"given"`
	PrePredict      *prePredict          `json:"prePredict,omitempty"`
	Predict         predict              `json:"predict"`
	PredictionState predictionState      `json:"state"`
}
