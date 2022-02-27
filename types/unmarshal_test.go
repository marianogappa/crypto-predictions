package types

import (
	"errors"
	"testing"
	"time"
)

func tp(s string) time.Time {
	t, _ := time.Parse("2006-01-02 15:04:05", s)
	return t
}

func TestParseDuration(t *testing.T) {
	var anyError = errors.New("any error for now...")

	tss := []struct {
		name     string
		dur      string
		fromTime time.Time
		err      error
		expected time.Duration
	}{
		{
			name:     "Empty string cannot be parsed",
			dur:      "",
			fromTime: time.Now(),
			err:      anyError,
			expected: 0,
		},
		{
			name:     "EOY works",
			dur:      "eoy",
			fromTime: tp("2020-12-31 23:59:59"),
			err:      nil,
			expected: 1 * time.Second,
		},
		{
			name:     "2d works",
			dur:      "2d",
			fromTime: time.Now(),
			err:      nil,
			expected: 48 * time.Hour,
		},
		{
			name:     "2w works",
			dur:      "2w",
			fromTime: time.Now(),
			err:      nil,
			expected: 24 * 7 * 2 * time.Hour,
		},
		{
			name:     "2m works",
			dur:      "2m",
			fromTime: time.Now(),
			err:      nil,
			expected: 24 * 30 * 2 * time.Hour,
		},
		{
			name:     "2h works",
			dur:      "2h",
			fromTime: time.Now(),
			err:      nil,
			expected: 2 * time.Hour,
		},
	}
	for _, ts := range tss {
		t.Run(ts.name, func(t *testing.T) {
			actual, actualErr := parseDuration(ts.dur, ts.fromTime)

			if actualErr != nil && ts.err == nil {
				t.Logf("expected no error but had '%v'", actualErr)
				t.FailNow()
			}
			if actualErr == nil && ts.err != nil {
				t.Logf("expected error '%v' but had no error", actualErr)
				t.FailNow()
			}
			if actual != ts.expected {
				t.Logf("expected %v but got %v", ts.expected, actual)
				t.FailNow()
			}
		})
	}
}
