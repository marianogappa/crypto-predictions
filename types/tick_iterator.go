package types

import (
	"errors"
)

type Tick struct {
	Timestamp int         `json:"t"`
	Value     JsonFloat64 `json:"v"`
}

var ErrOutOfTicks = errors.New("out of ticks")

type TickIterator interface {
	Next() (Tick, error)
	IsOutOfTicks() bool
}
