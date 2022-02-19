package types

import (
	"github.com/marianogappa/signal-checker/common"
)

type conditionState struct {
	Status string `json:"status"`
	LastTs int    `json:"lastTs"`
	Value  string `json:"value"`
	// add state to provide evidence of alleged condition result
}

type predictionState struct {
	Status string `json:"status"`
	LastTs int    `json:"lastTs"`
	Value  string `json:"value"`
	// add state to provide evidence of alleged condition result
}

type condition struct {
	Operator string         `json:"operator"`
	Operands []string       `json:"operands"`
	FromTs   common.ISO8601 `json:"fromTs"` // won't work for dynamic
	ToTs     common.ISO8601 `json:"toTs"`   // won't work for dynamic
	Assumed  []string       `json:"assumed"`
	State    conditionState `json:"state"`
}

type prePredict struct {
	WrongIf    *string `json:"wrongIf,omitempty"`
	AnnulledIf *string `json:"annulledIf,omitempty"`
	PredictIf  *string `json:"predict,omitempty"`
}

type predict struct {
	WrongIf    *string `json:"wrongIf,omitempty"`
	AnnulledIf *string `json:"annulledIf,omitempty"`
	Predict    string  `json:"predict,omitempty"`
}

type prediction struct {
	Version         string               `json:"version"`
	CreatedAt       common.ISO8601       `json:"createdAt"`
	AuthorHandle    string               `json:"authorHandle"`
	Post            string               `json:"post"`
	Define          map[string]condition `json:"define"`
	PrePredict      *prePredict          `json:"prePredict,omitempty"`
	Predict         predict              `json:"predict"`
	PredictionState predictionState      `json:"state"`
}
