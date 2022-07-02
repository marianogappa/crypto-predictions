package types

import "testing"

func TestConditionStateValue(t *testing.T) {
	tss := []struct {
		name string
		eval bool
	}{
		{
			name: "TRUE.And(TRUE) == TRUE",
			eval: TRUE.and(TRUE) == TRUE,
		},
		{
			name: "TRUE.And(FALSE) == FALSE",
			eval: TRUE.and(FALSE) == FALSE,
		},
		{
			name: "FALSE.And(TRUE) == FALSE",
			eval: FALSE.and(TRUE) == FALSE,
		},
		{
			name: "FALSE.And(FALSE) == FALSE",
			eval: FALSE.and(FALSE) == FALSE,
		},
		{
			name: "TRUE.And(UNDECIDED) == UNDECIDED",
			eval: TRUE.and(UNDECIDED) == UNDECIDED,
		},
		{
			name: "UNDECIDED.And(TRUE) == UNDECIDED",
			eval: UNDECIDED.and(TRUE) == UNDECIDED,
		},
		{
			name: "UNDECIDED.And(FALSE) == FALSE",
			eval: UNDECIDED.and(FALSE) == FALSE,
		},
		{
			name: "FALSE.And(UNDECIDED) == FALSE",
			eval: FALSE.and(UNDECIDED) == FALSE,
		},
		{
			name: "TRUE.Or(TRUE) == TRUE",
			eval: TRUE.or(TRUE) == TRUE,
		},
		{
			name: "TRUE.Or(FALSE) == TRUE",
			eval: TRUE.or(FALSE) == TRUE,
		},
		{
			name: "FALSE.Or(TRUE) == TRUE",
			eval: FALSE.or(TRUE) == TRUE,
		},
		{
			name: "FALSE.Or(FALSE) == FALSE",
			eval: FALSE.or(FALSE) == FALSE,
		},
		{
			name: "TRUE.Or(UNDECIDED) == TRUE",
			eval: TRUE.or(UNDECIDED) == TRUE,
		},
		{
			name: "UNDECIDED.Or(TRUE) == TRUE",
			eval: UNDECIDED.or(TRUE) == TRUE,
		},
		{
			name: "UNDECIDED.Or(FALSE) == UNDECIDED",
			eval: UNDECIDED.or(FALSE) == UNDECIDED,
		},
		{
			name: "FALSE.Or(UNDECIDED) == UNDECIDED",
			eval: FALSE.or(UNDECIDED) == UNDECIDED,
		},
		{
			name: "FALSE.Not() == TRUE",
			eval: FALSE.not() == TRUE,
		},
		{
			name: "TRUE.Not() == FALSE",
			eval: TRUE.not() == FALSE,
		},
		{
			name: "UNDECIDED.Not() == UNDECIDED",
			eval: UNDECIDED.not() == UNDECIDED,
		},
		{
			name: `TRUE.String() == "TRUE"`,
			eval: TRUE.String() == "TRUE",
		},
		{
			name: `FALSE.String() == "FALSE"`,
			eval: FALSE.String() == "FALSE",
		},
		{
			name: `UNDECIDED.String() == "UNDECIDED"`,
			eval: UNDECIDED.String() == "UNDECIDED",
		},
		{
			name: `ConditionStateValue(666).String() == ""`,
			eval: ConditionStateValue(666).String() == "",
		},
	}
	for _, ts := range tss {
		t.Run(ts.name, func(t *testing.T) {
			if !ts.eval {
				t.Log("expression is not true")
				t.FailNow()

			}
		})
	}
}

func TestConditionStateValueFromString(t *testing.T) {
	v, err := ConditionStateValueFromString("TRUE")
	if err != nil {
		t.Error("Error should have been nil")
		t.FailNow()
	}
	if v != TRUE {
		t.Error("Value should have been TRUE")
		t.FailNow()

	}
	v, err = ConditionStateValueFromString("FALSE")
	if err != nil {
		t.Error("Error should have been nil")
		t.FailNow()
	}
	if v != FALSE {
		t.Error("Value should have been FALSE")
		t.FailNow()

	}
	v, err = ConditionStateValueFromString("UNDECIDED")
	if err != nil {
		t.Error("Error should have been nil")
		t.FailNow()
	}
	if v != UNDECIDED {
		t.Error("Value should have been UNDECIDED")
		t.FailNow()

	}
	_, err = ConditionStateValueFromString("???")
	if err == nil {
		t.Error("Should have failed to parse")
		t.FailNow()
	}
}
