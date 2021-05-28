package timeseries

import (
	"fmt"
	"time"
)

// Backtest iterates over the provided historical data.
// Implements the QuoteProvider interface.
type Backtest struct {
	Data map[string]Data

	init  bool
	index int
}

// Next advances the time to the next month.
func (b *Backtest) Next() (time.Time, bool) {
	var data []Datum
	for _, ih := range b.Data {
		if len(data) == 0 || len(data) > len(ih.Data) {
			data = ih.Data
		}
	}

	if b.init {
		b.index++
		if b.index >= len(data) {
			return time.Time{}, false
		}
	} else {
		b.init = true
	}

	return data[b.index].Date, true
}

// RelativeValue returns the relative change for the position name.  Returns
// 1.0 if there is no change.
func (b *Backtest) RelativeValue(name string) (float64, error) {
	ih, ok := b.Data[name]
	if !ok {
		return 0, fmt.Errorf("no such data: %q", name)
	}

	if b.index >= len(ih.Data) {
		return 0, fmt.Errorf("Index out of bounds: have %d, size %d", b.index, len(ih.Data))
	}

	return 1 + ih.Data[b.index].Value, nil
}
