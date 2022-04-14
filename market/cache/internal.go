package cache

import (
	"fmt"
	"time"

	"github.com/marianogappa/predictions/types"
)

func (c *MemoryCache) putMinutely(operand types.Operand, candlesticks []types.Candlestick) error {
	var (
		key           string
		minsIndex     int
		err           error
		lastTimestamp int
	)
	for _, candlestick := range candlesticks {
		if lastTimestamp != 0 && candlestick.Timestamp-lastTimestamp != 60 {
			lastDateTime := time.Unix(int64(lastTimestamp), 0).UTC().Format(time.Kitchen)
			thisDateTime := time.Unix(int64(candlestick.Timestamp), 0).UTC().Format(time.Kitchen)
			return fmt.Errorf("%w: last date was %v and this was %v", ErrReceivedNonSubsequentCandlestick, lastDateTime, thisDateTime)
		}
		if candlestick.OpenPrice == 0 || candlestick.HighestPrice == 0 || candlestick.LowestPrice == 0 || candlestick.ClosePrice == 0 {
			return ErrReceivedCandlestickWithZeroValue
		}
		if key == "" || minsIndex >= 1440 {
			key = buildKey(operand, candlestick.Timestamp)
			minsIndex, err = calculateMinsIndex(candlestick.Timestamp)
			if err != nil {
				return err
			}
		}
		c.putMinutelyValue(key, minsIndex, candlestick)

		minsIndex++
		lastTimestamp = candlestick.Timestamp
	}
	return nil
}

func (c *MemoryCache) putMinutelyValue(key string, index int, candlestick types.Candlestick) {
	elem, ok := c.minutelyValueCache.Get(key)
	if !ok {
		elem = [1440]types.Candlestick{}
	}
	typedElem := elem.([1440]types.Candlestick)
	typedElem[index] = candlestick
	c.minutelyValueCache.Add(key, typedElem)
}

func (c *MemoryCache) putDaily(operand types.Operand, candlesticks []types.Candlestick) error {
	var (
		key           string
		daysIndex     int
		isLeap        = isLeapYear(candlesticks[0].Timestamp)
		err           error
		lastTimestamp int
	)
	for _, candlestick := range candlesticks {
		if lastTimestamp != 0 && candlestick.Timestamp-lastTimestamp != 60*60*24 {
			lastDateTime := time.Unix(int64(lastTimestamp), 0).Format("2006-01-02")
			thisDateTime := time.Unix(int64(candlestick.Timestamp), 0).Format("2006-01-02")
			return fmt.Errorf("%w: last date was %v and this was %v", ErrReceivedNonSubsequentCandlestick, lastDateTime, thisDateTime)
		}
		if candlestick.OpenPrice == 0 || candlestick.HighestPrice == 0 || candlestick.LowestPrice == 0 || candlestick.ClosePrice == 0 {
			return ErrReceivedCandlestickWithZeroValue
		}

		if key == "" || (isLeap && daysIndex >= 366 || !isLeap && daysIndex >= 365) {
			key = buildKey(operand, candlestick.Timestamp)
			daysIndex, err = calculateDaysIndex(candlestick.Timestamp)
			isLeap = isLeapYear(candlesticks[0].Timestamp)
			if err != nil {
				return err
			}
		}
		c.putDailyValue(key, daysIndex, candlestick)
		daysIndex++
		lastTimestamp = candlestick.Timestamp
	}
	return nil
}

func (c *MemoryCache) putDailyValue(key string, index int, candlestick types.Candlestick) {
	elem, ok := c.dailyValueCache.Get(key)
	if !ok {
		elem = [366]types.Candlestick{}
	}
	typedElem := elem.([366]types.Candlestick)
	typedElem[index] = candlestick
	c.dailyValueCache.Add(key, typedElem)
}

func (c *MemoryCache) getMinutely(operand types.Operand, startingTimestamp int) ([]types.Candlestick, error) {
	var (
		key          = buildKey(operand, startingTimestamp)
		candlesticks = []types.Candlestick{}
		elem, ok     = c.minutelyValueCache.Get(key)
		// This calculation cannot fail because the startingTimestamp always has 00 seconds.
		minsIndex, _ = calculateMinsIndex(startingTimestamp)
	)
	if !ok {
		c.CacheMisses++
		return []types.Candlestick{}, ErrCacheMiss
	}
	var (
		typedElem = elem.([1440]types.Candlestick)
	)
	for i := minsIndex; i <= 1439; i++ {
		if typedElem[i] == (types.Candlestick{}) {
			break
		}
		candlesticks = append(candlesticks, typedElem[i])
	}

	if len(candlesticks) == 0 {
		c.CacheMisses++
		return candlesticks, ErrCacheMiss
	}
	return candlesticks, nil
}

func (c *MemoryCache) getDaily(operand types.Operand, startingTimestamp int) ([]types.Candlestick, error) {
	var (
		key          = buildKey(operand, startingTimestamp)
		candlesticks = []types.Candlestick{}
		elem, ok     = c.dailyValueCache.Get(key)
		// This calculation cannot fail because the startingTimestamp always has 00 seconds.
		days, _ = calculateDaysIndex(startingTimestamp)
	)
	if !ok {
		c.CacheMisses++
		return []types.Candlestick{}, ErrCacheMiss
	}
	var (
		typedElem = elem.([366]types.Candlestick)
		isLeap    = isLeapYear(startingTimestamp)
	)
	for i := days; (isLeap && i <= 365) || (!isLeap && i <= 364); i++ {
		if typedElem[i] == (types.Candlestick{}) {
			break
		}
		candlesticks = append(candlesticks, typedElem[i])
	}

	if len(candlesticks) == 0 {
		c.CacheMisses++
		return candlesticks, ErrCacheMiss
	}
	return candlesticks, nil
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
