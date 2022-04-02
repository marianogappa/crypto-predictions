package cache

import (
	"fmt"
	"time"

	"github.com/marianogappa/predictions/types"
)

func (c *MemoryCache) putMinutely(operand types.Operand, ticks []types.Tick) error {
	var (
		key           string
		minsIndex     int
		err           error
		lastTimestamp int
	)
	for _, tick := range ticks {
		if lastTimestamp != 0 && tick.Timestamp-lastTimestamp != 60 {
			lastDateTime := time.Unix(int64(lastTimestamp), 0).UTC().Format(time.Kitchen)
			thisDateTime := time.Unix(int64(tick.Timestamp), 0).UTC().Format(time.Kitchen)
			return fmt.Errorf("%w: last date was %v and this was %v", ErrReceivedNonSubsequentTick, lastDateTime, thisDateTime)
		}
		if tick.Value == 0 {
			return ErrReceivedTickWithZeroValue
		}
		if key == "" || minsIndex >= 1440 {
			key = buildKey(operand, tick.Timestamp)
			minsIndex, err = calculateMinsIndex(tick.Timestamp)
			if err != nil {
				return err
			}
		}
		c.putMinutelyValue(key, minsIndex, tick.Value)

		minsIndex++
		lastTimestamp = tick.Timestamp
	}
	return nil
}

func (c *MemoryCache) putMinutelyValue(key string, index int, value types.JsonFloat64) {
	elem, ok := c.minutelyValueCache.Get(key)
	if !ok {
		elem = [1440]float64{}
	}
	typedElem := elem.([1440]float64)
	typedElem[index] = float64(value)
	c.minutelyValueCache.Add(key, typedElem)
}

func (c *MemoryCache) putDaily(operand types.Operand, ticks []types.Tick) error {
	var (
		key           string
		daysIndex     int
		isLeap        = isLeapYear(ticks[0].Timestamp)
		err           error
		lastTimestamp int
	)
	for _, tick := range ticks {
		if lastTimestamp != 0 && tick.Timestamp-lastTimestamp != 60*60*24 {
			lastDateTime := time.Unix(int64(lastTimestamp), 0).Format("2006-01-02")
			thisDateTime := time.Unix(int64(tick.Timestamp), 0).Format("2006-01-02")
			return fmt.Errorf("%w: last date was %v and this was %v", ErrReceivedNonSubsequentTick, lastDateTime, thisDateTime)
		}
		if tick.Value == 0 {
			return ErrReceivedTickWithZeroValue
		}

		if key == "" || (isLeap && daysIndex >= 366 || !isLeap && daysIndex >= 365) {
			key = buildKey(operand, tick.Timestamp)
			daysIndex, err = calculateDaysIndex(tick.Timestamp)
			isLeap = isLeapYear(ticks[0].Timestamp)
			if err != nil {
				return err
			}
		}
		c.putDailyValue(key, daysIndex, tick.Value)
		daysIndex++
		lastTimestamp = tick.Timestamp
	}
	return nil
}

func (c *MemoryCache) putDailyValue(key string, index int, value types.JsonFloat64) {
	elem, ok := c.dailyValueCache.Get(key)
	if !ok {
		elem = [366]float64{}
	}
	typedElem := elem.([366]float64)
	typedElem[index] = float64(value)
	c.dailyValueCache.Add(key, typedElem)
}

func (c *MemoryCache) getMinutely(operand types.Operand, startingTimestamp int) ([]types.Tick, error) {
	var (
		key      = buildKey(operand, startingTimestamp)
		ticks    = []types.Tick{}
		elem, ok = c.minutelyValueCache.Get(key)
		// This calculation cannot fail because the startingTimestamp always has 00 seconds.
		minsIndex, _ = calculateMinsIndex(startingTimestamp)
	)
	if !ok {
		c.CacheMisses++
		return []types.Tick{}, ErrCacheMiss
	}
	var (
		typedElem    = elem.([1440]float64)
		y, m, d      = time.Unix(int64(startingTimestamp), 0).UTC().Date()
		startOfDayTs = int(time.Date(y, m, d, 0, 0, 0, 0, time.UTC).Unix())
	)
	for i := minsIndex; i <= 1439; i++ {
		if typedElem[i] == 0 {
			break
		}
		ticks = append(ticks, types.Tick{
			Timestamp: startOfDayTs + i*60,
			Value:     types.JsonFloat64(typedElem[i]),
		})
	}

	if len(ticks) == 0 {
		c.CacheMisses++
		return ticks, ErrCacheMiss
	}
	return ticks, nil
}

func (c *MemoryCache) getDaily(operand types.Operand, startingTimestamp int) ([]types.Tick, error) {
	var (
		key      = buildKey(operand, startingTimestamp)
		ticks    = []types.Tick{}
		elem, ok = c.dailyValueCache.Get(key)
		// This calculation cannot fail because the startingTimestamp always has 00 seconds.
		days, _ = calculateDaysIndex(startingTimestamp)
	)
	if !ok {
		c.CacheMisses++
		return []types.Tick{}, ErrCacheMiss
	}
	var (
		typedElem     = elem.([366]float64)
		isLeap        = isLeapYear(startingTimestamp)
		y, _, _       = time.Unix(int64(startingTimestamp), 0).UTC().Date()
		startOfYearTs = int(time.Date(y, 1, 1, 0, 0, 0, 0, time.UTC).Unix())
	)
	for i := days; (isLeap && i <= 365) || (!isLeap && i <= 364); i++ {
		if typedElem[i] == 0 {
			break
		}
		ticks = append(ticks, types.Tick{
			Timestamp: startOfYearTs + i*60*60*24,
			Value:     types.JsonFloat64(typedElem[i]),
		})
	}

	if len(ticks) == 0 {
		c.CacheMisses++
		return ticks, ErrCacheMiss
	}
	return ticks, nil
}

func calculateNormalizedStartingTimestamp(operand types.Operand, tm time.Time) int {
	if isMinutely(operand) {
		if tm.Second() == 0 {
			return int(tm.Unix())
		}
		year, month, day := tm.Date()
		return int(time.Date(year, month, day, tm.Hour(), tm.Minute()+1, 0, 0, time.UTC).Unix())
	}

	if tm.Second() == 0 && tm.Minute() == 0 && tm.Hour() == 0 {
		return int(tm.Unix())
	}
	year, month, day := tm.Date()
	return int(time.Date(year, month, day+1, 0, 0, 0, 0, time.UTC).Unix())
}

func isMinutely(op types.Operand) bool {
	return op.Type == types.COIN
}

func buildKey(op types.Operand, timestamp int) string {
	tm := time.Unix(int64(timestamp), 0).UTC()
	datePart := tm.Format("2006")
	if isMinutely(op) {
		datePart = tm.Format("2006-01-02")
	}
	return fmt.Sprintf("%v-%v-%v-%v", op.Provider, op.BaseAsset, op.QuoteAsset, datePart)
}

func calculateMinsIndex(ts int) (int, error) {
	tm := time.Unix(int64(ts), 0).UTC()
	if tm.Second() != 0 {
		return -1, fmt.Errorf("%w: was %v", ErrTimestampMustHaveZeroInSecondsPart, tm.Second())
	}
	return tm.Hour()*60 + tm.Minute(), nil
}

func calculateDaysIndex(ts int) (int, error) {
	tm := time.Unix(int64(ts), 0).UTC()
	if tm.Second() != 0 || tm.Hour() != 0 || tm.Minute() != 0 {
		return -1, fmt.Errorf("%w: was %v:%v:%v", ErrTimestampMustHaveZeroInTimePart, tm.Hour(), tm.Minute(), tm.Second())
	}
	return tm.YearDay() - 1, nil
}

func isLeapYear(ts int) bool {
	tm := time.Unix(int64(ts), 0).UTC()
	year := tm.Year()
	return year%4 == 0 && (year%100 != 0 || year%400 == 0)
}
