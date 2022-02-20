package types

import (
	"errors"

	"github.com/marianogappa/signal-checker/common"
)

type Tick struct {
	Timestamp int                `json:"t"`
	Value     common.JsonFloat64 `json:"v"`
}

var ErrOutOfTicks = errors.New("out of ticks")

type TickIterator interface {
	Next() (Tick, error)
	IsOutOfTicks() bool
}
