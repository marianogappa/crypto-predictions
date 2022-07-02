package types

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

var (
	strVariable = `(COIN|MARKETCAP):([A-Z]+):([A-Z]+)(-([A-Z]+))?`
	rxVariable  = regexp.MustCompile(strVariable)
)

func tp(s string) time.Time {
	t, _ := time.Parse("2006-01-02 15:04:05", s)
	return t
}

func tInt(s string) int {
	return int(tp(s).Unix())
}

func mapOperand(v string) (Operand, error) {
	v = strings.ToUpper(v)
	f, err := strconv.ParseFloat(v, 64)
	if err == nil {
		return Operand{Type: NUMBER, Number: JSONFloat64(f), Str: v}, nil
	}
	matches := rxVariable.FindStringSubmatch(v)
	if len(matches) == 0 {
		return Operand{}, fmt.Errorf("operand %v doesn't parse to float nor match the regex %v", v, strVariable)
	}
	operandType, err := OperandTypeFromString(matches[1])
	if err != nil {
		return Operand{}, fmt.Errorf("invalid operand type %v", matches[1])
	}
	return Operand{
		Type:       operandType,
		Provider:   matches[2],
		BaseAsset:  matches[3],
		QuoteAsset: matches[5],
		Str:        v,
	}, nil
}

func operand(s string) Operand {
	op, _ := mapOperand(s)
	return op
}

func TestCondition(t *testing.T) {
	var (
		anyError = errors.New("any error for now... ")
		times    = []int{tInt("2022-01-01 00:00:00"), tInt("2022-01-02 00:00:00"), tInt("2022-01-03 00:00:00")}
	)

	tss := []struct {
		name     string
		cond     *Condition
		ticks    map[string]Tick
		err      error
		expected ConditionState
	}{
		{
			name: "errors if no ticks supplied",
			cond: &Condition{
				Name:             "main",
				Operator:         ">",
				Operands:         []Operand{operand("COIN:BINANCE:BTC-USDT"), operand("60000")},
				FromTs:           times[0],
				ToTs:             times[1],
				ErrorMarginRatio: 0.0,
			},
			ticks: nil,
			err:   errAtLeastOneTickRequired,
		},
		{
			name: "errors if invalid tick supplied",
			cond: &Condition{
				Name:     "main",
				Operator: ">",
				Operands: []Operand{operand("COIN:BINANCE:BTC-USDT"), operand("60000")},
				FromTs:   times[0],
				ToTs:     times[1],
			},
			ticks: map[string]Tick{"COIN:BINANCE:BTC-USDT": {}}, // invalid tick!
			err:   errInvalidTickSupplied,
		},
		{
			name: "errors if specific tick not supplied",
			cond: &Condition{
				Name:     "main",
				Operator: ">",
				Operands: []Operand{operand("COIN:BINANCE:BTC-USDT"), operand("60000")},
				FromTs:   times[0],
				ToTs:     times[1],
				State: ConditionState{
					Status:    UNSTARTED,
					LastTs:    0,
					LastTicks: nil,
					Value:     UNDECIDED,
				},
				ErrorMarginRatio: 0.0,
			},
			ticks: map[string]Tick{"COIN:BINANCE:ETH-USDT": {Timestamp: times[0], Value: 61000}}, // not BTC-USDT!
			err:   anyError,
			expected: ConditionState{
				Status:    UNSTARTED,
				LastTs:    0,
				LastTicks: nil,
				Value:     UNDECIDED,
			},
		},
		{
			name: "errors if two ticks are supplied with different timestamps",
			cond: &Condition{
				Name:     "main",
				Operator: ">",
				Operands: []Operand{operand("COIN:BINANCE:BTC-USDT"), operand("COIN:BINANCE:ETH-USDT")},
				FromTs:   times[0],
				ToTs:     times[1],
			},
			ticks: map[string]Tick{
				"COIN:BINANCE:BTC-USDT": {Timestamp: times[0], Value: 61000},
				"COIN:BINANCE:ETH-USDT": {Timestamp: times[1], Value: 4000}}, // Different timestamp!
			err: errMismatchingTickTimestampsSupplied,
		},
		{
			name: "ignores older timestamp ticks",
			cond: &Condition{
				Name:     "main",
				Operator: ">",
				Operands: []Operand{operand("COIN:BINANCE:BTC-USDT"), operand("60000")},
				FromTs:   times[1],
				ToTs:     times[2],
				State: ConditionState{
					Status:    UNSTARTED,
					LastTs:    0,
					LastTicks: nil,
					Value:     UNDECIDED,
				},
				ErrorMarginRatio: 0.0,
			},
			ticks: map[string]Tick{"COIN:BINANCE:BTC-USDT": {Timestamp: times[0], Value: 61000}}, // would be TRUE!
			err:   nil,
			expected: ConditionState{
				Status:    UNSTARTED,
				LastTs:    0,
				LastTicks: nil,
				Value:     UNDECIDED,
			},
		},
		{
			name: "errors if received older timestamp than last tick's",
			cond: &Condition{
				Name:     "main",
				Operator: ">",
				Operands: []Operand{operand("COIN:BINANCE:BTC-USDT"), operand("60000")},
				FromTs:   times[1],
				ToTs:     times[2],
				State: ConditionState{
					Status:    STARTED,
					LastTs:    times[1],
					LastTicks: map[string]Tick{"COIN:BINANCE:BTC-USDT": {Timestamp: times[1], Value: 59000}},
					Value:     UNDECIDED,
				},
			},
			ticks: map[string]Tick{"COIN:BINANCE:BTC-USDT": {Timestamp: times[0], Value: 61000}}, // would be TRUE, but older!
			err:   errOlderTickTimestampSupplied,
			expected: ConditionState{
				Status:    STARTED,
				LastTs:    times[1],
				LastTicks: map[string]Tick{"COIN:BINANCE:BTC-USDT": {Timestamp: times[1], Value: 59000}},
				Value:     UNDECIDED,
			},
		},
		{
			name: "ignores ticks if already decided",
			cond: &Condition{
				Name:     "main",
				Operator: ">",
				Operands: []Operand{operand("COIN:BINANCE:BTC-USDT"), operand("60000")},
				FromTs:   times[0],
				ToTs:     times[2],
				State: ConditionState{
					Status:    FINISHED,
					LastTs:    times[0],
					LastTicks: map[string]Tick{"COIN:BINANCE:BTC-USDT": {Timestamp: times[0], Value: 61000}},
					Value:     TRUE,
				},
				ErrorMarginRatio: 0.0,
			},
			ticks: map[string]Tick{"COIN:BINANCE:BTC-USDT": {Timestamp: times[1], Value: 59000}},
			err:   nil,
			expected: ConditionState{
				Status:    FINISHED,
				LastTs:    times[0],
				LastTicks: map[string]Tick{"COIN:BINANCE:BTC-USDT": {Timestamp: times[0], Value: 61000}},
				Value:     TRUE,
			},
		},
		{
			name: "sets to false if timestamp exceeded (looks like exchange lost a tick for times[1])",
			cond: &Condition{
				Name:     "main",
				Operator: ">",
				Operands: []Operand{operand("COIN:BINANCE:BTC-USDT"), operand("60000")},
				FromTs:   times[0],
				ToTs:     times[1],
				State: ConditionState{
					Status:    STARTED,
					LastTs:    times[0],
					LastTicks: map[string]Tick{"COIN:BINANCE:BTC-USDT": {Timestamp: times[0], Value: 59000}},
					Value:     UNDECIDED,
				},
				ErrorMarginRatio: 0.0,
			},
			ticks: map[string]Tick{"COIN:BINANCE:BTC-USDT": {Timestamp: times[2], Value: 59500}},
			err:   nil,
			expected: ConditionState{
				Status:    FINISHED,
				LastTs:    times[0],
				LastTicks: map[string]Tick{"COIN:BINANCE:BTC-USDT": {Timestamp: times[0], Value: 59000}},
				Value:     FALSE,
			},
		},
		{
			name: "> with two coins",
			cond: &Condition{
				Name:     "main",
				Operator: ">",
				Operands: []Operand{operand("COIN:BINANCE:BTC-USDT"), operand("COIN:BINANCE:ETH-USDT")},
				FromTs:   times[0],
				ToTs:     times[1],
				State: ConditionState{
					Status:    UNSTARTED,
					LastTs:    0,
					LastTicks: nil,
					Value:     UNDECIDED,
				},
				ErrorMarginRatio: 0.0,
			},
			ticks: map[string]Tick{"COIN:BINANCE:BTC-USDT": {Timestamp: times[0], Value: 60000}, "COIN:BINANCE:ETH-USDT": {Timestamp: times[0], Value: 4000}},
			err:   nil,
			expected: ConditionState{
				Status:    FINISHED,
				LastTs:    times[0],
				LastTicks: map[string]Tick{"COIN:BINANCE:BTC-USDT": {Timestamp: times[0], Value: 60000}, "COIN:BINANCE:ETH-USDT": {Timestamp: times[0], Value: 4000}},
				Value:     TRUE,
			},
		},
		{
			name: "> with coin and number works for undecided",
			cond: &Condition{
				Name:     "main",
				Operator: ">",
				Operands: []Operand{operand("COIN:BINANCE:BTC-USDT"), operand("60000")},
				FromTs:   times[0],
				ToTs:     times[1],
				State: ConditionState{
					Status:    UNSTARTED,
					LastTs:    0,
					LastTicks: nil,
					Value:     UNDECIDED,
				},
				ErrorMarginRatio: 0.0,
			},
			ticks: map[string]Tick{"COIN:BINANCE:BTC-USDT": {Timestamp: times[0], Value: 60000}},
			err:   nil,
			expected: ConditionState{
				Status:    STARTED,
				LastTs:    times[0],
				LastTicks: map[string]Tick{"COIN:BINANCE:BTC-USDT": {Timestamp: times[0], Value: 60000}},
				Value:     UNDECIDED,
			},
		},
		{
			name: "> with coin and number works for true",
			cond: &Condition{
				Name:     "main",
				Operator: ">",
				Operands: []Operand{operand("COIN:BINANCE:BTC-USDT"), operand("60000")},
				FromTs:   times[0],
				ToTs:     times[1],
				State: ConditionState{
					Status:    UNSTARTED,
					LastTs:    0,
					LastTicks: nil,
					Value:     UNDECIDED,
				},
				ErrorMarginRatio: 0.0,
			},
			ticks: map[string]Tick{"COIN:BINANCE:BTC-USDT": {Timestamp: times[0], Value: 61000}},
			err:   nil,
			expected: ConditionState{
				Status:    FINISHED,
				LastTs:    times[0],
				LastTicks: map[string]Tick{"COIN:BINANCE:BTC-USDT": {Timestamp: times[0], Value: 61000}},
				Value:     TRUE,
			},
		},
		{
			name: "> with coin and number works for false",
			cond: &Condition{
				Name:     "main",
				Operator: ">",
				Operands: []Operand{operand("COIN:BINANCE:BTC-USDT"), operand("60000")},
				FromTs:   times[0],
				ToTs:     times[1],
				State: ConditionState{
					Status:    UNSTARTED,
					LastTs:    0,
					LastTicks: nil,
					Value:     UNDECIDED,
				},
				ErrorMarginRatio: 0.0,
			},
			ticks: map[string]Tick{"COIN:BINANCE:BTC-USDT": {Timestamp: times[1], Value: 59000}},
			err:   nil,
			expected: ConditionState{
				Status:    FINISHED,
				LastTs:    times[1],
				LastTicks: map[string]Tick{"COIN:BINANCE:BTC-USDT": {Timestamp: times[1], Value: 59000}},
				Value:     FALSE,
			},
		},
		{
			name: "> with error margin",
			cond: &Condition{
				Name:     "main",
				Operator: ">",
				Operands: []Operand{operand("COIN:BINANCE:FTM-USDT"), operand("1000")},
				FromTs:   times[0],
				ToTs:     times[1],
				State: ConditionState{
					Status:    UNSTARTED,
					LastTs:    0,
					LastTicks: nil,
					Value:     UNDECIDED,
				},
				ErrorMarginRatio: 0.1, // 10% error margin so actually FTM-USDT > 900
			},
			ticks: map[string]Tick{"COIN:BINANCE:FTM-USDT": {Timestamp: times[1], Value: 950}},
			err:   nil,
			expected: ConditionState{
				Status:    FINISHED,
				LastTs:    times[1],
				LastTicks: map[string]Tick{"COIN:BINANCE:FTM-USDT": {Timestamp: times[1], Value: 950}},
				Value:     TRUE,
			},
		},
		{
			name: ">= with coin and number works for undecided",
			cond: &Condition{
				Name:     "main",
				Operator: ">=",
				Operands: []Operand{operand("COIN:BINANCE:BTC-USDT"), operand("60000")},
				FromTs:   times[0],
				ToTs:     times[1],
				State: ConditionState{
					Status:    UNSTARTED,
					LastTs:    0,
					LastTicks: nil,
					Value:     UNDECIDED,
				},
				ErrorMarginRatio: 0.0,
			},
			ticks: map[string]Tick{"COIN:BINANCE:BTC-USDT": {Timestamp: times[0], Value: 59000}},
			err:   nil,
			expected: ConditionState{
				Status:    STARTED,
				LastTs:    times[0],
				LastTicks: map[string]Tick{"COIN:BINANCE:BTC-USDT": {Timestamp: times[0], Value: 59000}},
				Value:     UNDECIDED,
			},
		},
		{
			name: ">= with coin and number works for true",
			cond: &Condition{
				Name:     "main",
				Operator: ">=",
				Operands: []Operand{operand("COIN:BINANCE:BTC-USDT"), operand("60000")},
				FromTs:   times[0],
				ToTs:     times[1],
				State: ConditionState{
					Status:    UNSTARTED,
					LastTs:    0,
					LastTicks: nil,
					Value:     UNDECIDED,
				},
				ErrorMarginRatio: 0.0,
			},
			ticks: map[string]Tick{"COIN:BINANCE:BTC-USDT": {Timestamp: times[0], Value: 61000}},
			err:   nil,
			expected: ConditionState{
				Status:    FINISHED,
				LastTs:    times[0],
				LastTicks: map[string]Tick{"COIN:BINANCE:BTC-USDT": {Timestamp: times[0], Value: 61000}},
				Value:     TRUE,
			},
		},
		{
			name: ">= with coin and number works for false",
			cond: &Condition{
				Name:     "main",
				Operator: ">=",
				Operands: []Operand{operand("COIN:BINANCE:BTC-USDT"), operand("60000")},
				FromTs:   times[0],
				ToTs:     times[1],
				State: ConditionState{
					Status:    UNSTARTED,
					LastTs:    0,
					LastTicks: nil,
					Value:     UNDECIDED,
				},
				ErrorMarginRatio: 0.0,
			},
			ticks: map[string]Tick{"COIN:BINANCE:BTC-USDT": {Timestamp: times[1], Value: 59000}},
			err:   nil,
			expected: ConditionState{
				Status:    FINISHED,
				LastTs:    times[1],
				LastTicks: map[string]Tick{"COIN:BINANCE:BTC-USDT": {Timestamp: times[1], Value: 59000}},
				Value:     FALSE,
			},
		},
		{
			name: ">= with error margin",
			cond: &Condition{
				Name:     "main",
				Operator: ">=",
				Operands: []Operand{operand("COIN:BINANCE:FTM-USDT"), operand("1000")},
				FromTs:   times[0],
				ToTs:     times[1],
				State: ConditionState{
					Status:    UNSTARTED,
					LastTs:    0,
					LastTicks: nil,
					Value:     UNDECIDED,
				},
				ErrorMarginRatio: 0.1, // 10% error margin so actually FTM-USDT >= 900
			},
			ticks: map[string]Tick{"COIN:BINANCE:FTM-USDT": {Timestamp: times[1], Value: 900}},
			err:   nil,
			expected: ConditionState{
				Status:    FINISHED,
				LastTs:    times[1],
				LastTicks: map[string]Tick{"COIN:BINANCE:FTM-USDT": {Timestamp: times[1], Value: 900}},
				Value:     TRUE,
			},
		},
		{
			name: "< with coin and number works for undecided",
			cond: &Condition{
				Name:     "main",
				Operator: "<",
				Operands: []Operand{operand("COIN:BINANCE:BTC-USDT"), operand("60000")},
				FromTs:   times[0],
				ToTs:     times[1],
				State: ConditionState{
					Status:    UNSTARTED,
					LastTs:    0,
					LastTicks: nil,
					Value:     UNDECIDED,
				},
				ErrorMarginRatio: 0.0,
			},
			ticks: map[string]Tick{"COIN:BINANCE:BTC-USDT": {Timestamp: times[0], Value: 60000}},
			err:   nil,
			expected: ConditionState{
				Status:    STARTED,
				LastTs:    times[0],
				LastTicks: map[string]Tick{"COIN:BINANCE:BTC-USDT": {Timestamp: times[0], Value: 60000}},
				Value:     UNDECIDED,
			},
		},
		{
			name: "< with coin and number works for true",
			cond: &Condition{
				Name:     "main",
				Operator: "<",
				Operands: []Operand{operand("COIN:BINANCE:BTC-USDT"), operand("60000")},
				FromTs:   times[0],
				ToTs:     times[1],
				State: ConditionState{
					Status:    UNSTARTED,
					LastTs:    0,
					LastTicks: nil,
					Value:     UNDECIDED,
				},
				ErrorMarginRatio: 0.0,
			},
			ticks: map[string]Tick{"COIN:BINANCE:BTC-USDT": {Timestamp: times[0], Value: 59000}},
			err:   nil,
			expected: ConditionState{
				Status:    FINISHED,
				LastTs:    times[0],
				LastTicks: map[string]Tick{"COIN:BINANCE:BTC-USDT": {Timestamp: times[0], Value: 59000}},
				Value:     TRUE,
			},
		},
		{
			name: "< with coin and number works for false",
			cond: &Condition{
				Name:     "main",
				Operator: "<",
				Operands: []Operand{operand("COIN:BINANCE:BTC-USDT"), operand("60000")},
				FromTs:   times[0],
				ToTs:     times[1],
				State: ConditionState{
					Status:    UNSTARTED,
					LastTs:    0,
					LastTicks: nil,
					Value:     UNDECIDED,
				},
				ErrorMarginRatio: 0.0,
			},
			ticks: map[string]Tick{"COIN:BINANCE:BTC-USDT": {Timestamp: times[1], Value: 61000}},
			err:   nil,
			expected: ConditionState{
				Status:    FINISHED,
				LastTs:    times[1],
				LastTicks: map[string]Tick{"COIN:BINANCE:BTC-USDT": {Timestamp: times[1], Value: 61000}},
				Value:     FALSE,
			},
		},
		{
			name: "< with error margin",
			cond: &Condition{
				Name:     "main",
				Operator: "<",
				Operands: []Operand{operand("COIN:BINANCE:FTM-USDT"), operand("1000")},
				FromTs:   times[0],
				ToTs:     times[1],
				State: ConditionState{
					Status:    UNSTARTED,
					LastTs:    0,
					LastTicks: nil,
					Value:     UNDECIDED,
				},
				ErrorMarginRatio: 0.1, // 10% error margin, so actually FTM-USDT < 1100
			},
			ticks: map[string]Tick{"COIN:BINANCE:FTM-USDT": {Timestamp: times[1], Value: 1050}},
			err:   nil,
			expected: ConditionState{
				Status:    FINISHED,
				LastTs:    times[1],
				LastTicks: map[string]Tick{"COIN:BINANCE:FTM-USDT": {Timestamp: times[1], Value: 1050}},
				Value:     TRUE,
			},
		},
		{
			name: "<= with coin and number works for undecided",
			cond: &Condition{
				Name:     "main",
				Operator: "<=",
				Operands: []Operand{operand("COIN:BINANCE:BTC-USDT"), operand("60000")},
				FromTs:   times[0],
				ToTs:     times[1],
				State: ConditionState{
					Status:    UNSTARTED,
					LastTs:    0,
					LastTicks: nil,
					Value:     UNDECIDED,
				},
				ErrorMarginRatio: 0.0,
			},
			ticks: map[string]Tick{"COIN:BINANCE:BTC-USDT": {Timestamp: times[0], Value: 61000}},
			err:   nil,
			expected: ConditionState{
				Status:    STARTED,
				LastTs:    times[0],
				LastTicks: map[string]Tick{"COIN:BINANCE:BTC-USDT": {Timestamp: times[0], Value: 61000}},
				Value:     UNDECIDED,
			},
		},
		{
			name: "<= with coin and number works for true",
			cond: &Condition{
				Name:     "main",
				Operator: "<=",
				Operands: []Operand{operand("COIN:BINANCE:BTC-USDT"), operand("60000")},
				FromTs:   times[0],
				ToTs:     times[1],
				State: ConditionState{
					Status:    UNSTARTED,
					LastTs:    0,
					LastTicks: nil,
					Value:     UNDECIDED,
				},
				ErrorMarginRatio: 0.0,
			},
			ticks: map[string]Tick{"COIN:BINANCE:BTC-USDT": {Timestamp: times[0], Value: 60000}},
			err:   nil,
			expected: ConditionState{
				Status:    FINISHED,
				LastTs:    times[0],
				LastTicks: map[string]Tick{"COIN:BINANCE:BTC-USDT": {Timestamp: times[0], Value: 60000}},
				Value:     TRUE,
			},
		},
		{
			name: "<= with coin and number works for false",
			cond: &Condition{
				Name:     "main",
				Operator: "<=",
				Operands: []Operand{operand("COIN:BINANCE:BTC-USDT"), operand("60000")},
				FromTs:   times[0],
				ToTs:     times[1],
				State: ConditionState{
					Status:    UNSTARTED,
					LastTs:    0,
					LastTicks: nil,
					Value:     UNDECIDED,
				},
				ErrorMarginRatio: 0.0,
			},
			ticks: map[string]Tick{"COIN:BINANCE:BTC-USDT": {Timestamp: times[1], Value: 61000}},
			err:   nil,
			expected: ConditionState{
				Status:    FINISHED,
				LastTs:    times[1],
				LastTicks: map[string]Tick{"COIN:BINANCE:BTC-USDT": {Timestamp: times[1], Value: 61000}},
				Value:     FALSE,
			},
		},
		{
			name: "<= with error margin",
			cond: &Condition{
				Name:     "main",
				Operator: "<=",
				Operands: []Operand{operand("COIN:BINANCE:FTM-USDT"), operand("1000")},
				FromTs:   times[0],
				ToTs:     times[1],
				State: ConditionState{
					Status:    UNSTARTED,
					LastTs:    0,
					LastTicks: nil,
					Value:     UNDECIDED,
				},
				ErrorMarginRatio: 0.1, // 10% error margin, so actually FTM-USDT <= 1100
			},
			ticks: map[string]Tick{"COIN:BINANCE:FTM-USDT": {Timestamp: times[1], Value: 1100}},
			err:   nil,
			expected: ConditionState{
				Status:    FINISHED,
				LastTs:    times[1],
				LastTicks: map[string]Tick{"COIN:BINANCE:FTM-USDT": {Timestamp: times[1], Value: 1100}},
				Value:     TRUE,
			},
		},
		{
			name: "BETWEEN with coin and number works for undecided",
			cond: &Condition{
				Name:     "main",
				Operator: "BETWEEN",
				Operands: []Operand{operand("COIN:BINANCE:BTC-USDT"), operand("60000"), operand("61000")},
				FromTs:   times[0],
				ToTs:     times[1],
				State: ConditionState{
					Status:    UNSTARTED,
					LastTs:    0,
					LastTicks: nil,
					Value:     UNDECIDED,
				},
				ErrorMarginRatio: 0.0,
			},
			ticks: map[string]Tick{"COIN:BINANCE:BTC-USDT": {Timestamp: times[0], Value: 59000}},
			err:   nil,
			expected: ConditionState{
				Status:    STARTED,
				LastTs:    times[0],
				LastTicks: map[string]Tick{"COIN:BINANCE:BTC-USDT": {Timestamp: times[0], Value: 59000}},
				Value:     UNDECIDED,
			},
		},
		{
			name: "BETWEEN with coin and number works for true, inclusive on lower side",
			cond: &Condition{
				Name:     "main",
				Operator: "BETWEEN",
				Operands: []Operand{operand("COIN:BINANCE:BTC-USDT"), operand("60000"), operand("61000")},
				FromTs:   times[0],
				ToTs:     times[1],
				State: ConditionState{
					Status:    UNSTARTED,
					LastTs:    0,
					LastTicks: nil,
					Value:     UNDECIDED,
				},
				ErrorMarginRatio: 0.0,
			},
			ticks: map[string]Tick{"COIN:BINANCE:BTC-USDT": {Timestamp: times[0], Value: 60000}},
			err:   nil,
			expected: ConditionState{
				Status:    FINISHED,
				LastTs:    times[0],
				LastTicks: map[string]Tick{"COIN:BINANCE:BTC-USDT": {Timestamp: times[0], Value: 60000}},
				Value:     TRUE,
			},
		},
		{
			name: "BETWEEN with coin and number works for true, inclusive on middle",
			cond: &Condition{
				Name:     "main",
				Operator: "BETWEEN",
				Operands: []Operand{operand("COIN:BINANCE:BTC-USDT"), operand("60000"), operand("61000")},
				FromTs:   times[0],
				ToTs:     times[1],
				State: ConditionState{
					Status:    UNSTARTED,
					LastTs:    0,
					LastTicks: nil,
					Value:     UNDECIDED,
				},
				ErrorMarginRatio: 0.0,
			},
			ticks: map[string]Tick{"COIN:BINANCE:BTC-USDT": {Timestamp: times[0], Value: 60500}},
			err:   nil,
			expected: ConditionState{
				Status:    FINISHED,
				LastTs:    times[0],
				LastTicks: map[string]Tick{"COIN:BINANCE:BTC-USDT": {Timestamp: times[0], Value: 60500}},
				Value:     TRUE,
			},
		},
		{
			name: "BETWEEN with coin and number works for true, inclusive on higher side",
			cond: &Condition{
				Name:     "main",
				Operator: "BETWEEN",
				Operands: []Operand{operand("COIN:BINANCE:BTC-USDT"), operand("60000"), operand("61000")},
				FromTs:   times[0],
				ToTs:     times[1],
				State: ConditionState{
					Status:    UNSTARTED,
					LastTs:    0,
					LastTicks: nil,
					Value:     UNDECIDED,
				},
				ErrorMarginRatio: 0.0,
			},
			ticks: map[string]Tick{"COIN:BINANCE:BTC-USDT": {Timestamp: times[0], Value: 61000}},
			err:   nil,
			expected: ConditionState{
				Status:    FINISHED,
				LastTs:    times[0],
				LastTicks: map[string]Tick{"COIN:BINANCE:BTC-USDT": {Timestamp: times[0], Value: 61000}},
				Value:     TRUE,
			},
		},
		{
			name: "BETWEEN with coin and number works for false",
			cond: &Condition{
				Name:     "main",
				Operator: "BETWEEN",
				Operands: []Operand{operand("COIN:BINANCE:BTC-USDT"), operand("60000"), operand("61000")},
				FromTs:   times[0],
				ToTs:     times[1],
				State: ConditionState{
					Status:    UNSTARTED,
					LastTs:    0,
					LastTicks: nil,
					Value:     UNDECIDED,
				},
				ErrorMarginRatio: 0.0,
			},
			ticks: map[string]Tick{"COIN:BINANCE:BTC-USDT": {Timestamp: times[1], Value: 59000}},
			err:   nil,
			expected: ConditionState{
				Status:    FINISHED,
				LastTs:    times[1],
				LastTicks: map[string]Tick{"COIN:BINANCE:BTC-USDT": {Timestamp: times[1], Value: 59000}},
				Value:     FALSE,
			},
		},
	}
	for _, ts := range tss {
		t.Run(ts.name, func(t *testing.T) {
			actualErr := ts.cond.Run(ts.ticks)
			if actualErr != nil && ts.err == nil {
				t.Logf("expected no error but had '%v'", actualErr)
				t.FailNow()
			}
			if actualErr == nil && ts.err != nil {
				t.Logf("expected error '%v' but had no error", actualErr)
				t.FailNow()
			}
			if !reflect.DeepEqual(ts.cond.State, ts.expected) {
				t.Logf("expected %v but got %v", ts.expected, ts.cond.State)
				t.FailNow()
			}
		})
	}
}

func TestConditionClearState(t *testing.T) {
	expected := ConditionState{
		Status:    UNSTARTED,
		LastTs:    0,
		LastTicks: nil,
		Value:     UNDECIDED,
	}

	c := &Condition{
		State: ConditionState{
			Status:    STARTED,
			LastTs:    tInt("2022-01-01 00:00:00"),
			LastTicks: map[string]Tick{"COIN:BINANCE:BTC-USDT": {Timestamp: tInt("2022-01-01 00:00:00"), Value: 59000}},
			Value:     UNDECIDED,
		},
	}
	c.ClearState()
	if !reflect.DeepEqual(c.State, expected) {
		t.Errorf("expected state to be %v but was %v", expected, c.State)
	}
}

func TestConditionNonNumberOperands(t *testing.T) {
	expected := []Operand{
		{
			Type:       COIN,
			Provider:   "BINANCE",
			BaseAsset:  "BTC",
			QuoteAsset: "USDT",
		},
		{
			Type:      MARKETCAP,
			Provider:  "MESSARI",
			BaseAsset: "BTC",
		},
	}

	c := &Condition{
		Operands: []Operand{
			{
				Type:       COIN,
				Provider:   "BINANCE",
				BaseAsset:  "BTC",
				QuoteAsset: "USDT",
			},
			{
				Type:      MARKETCAP,
				Provider:  "MESSARI",
				BaseAsset: "BTC",
			},
			{
				Type:   NUMBER,
				Number: 3.1415,
			},
		},
	}
	require.Equal(t, expected, c.NonNumberOperands())
}
