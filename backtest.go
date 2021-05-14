package main

import (
	"fmt"
	"time"
)

type Backtest struct {
	Data  map[string]IndexHistory
	Index int
}

func (b *Backtest) Next() (time.Time, bool) {
	var data []IndexDatum
	for _, ih := range b.Data {
		if len(data) == 0 || len(data) > len(ih.Data) {
			data = ih.Data
		}
	}

	if b.Index >= len(data)-1 {
		return time.Time{}, false
	}
	b.Index++

	return data[b.Index].Date, true
}

func (b *Backtest) RelativeValue(name string) (float64, error) {
	ih, ok := b.Data[name]
	if !ok {
		return 0, fmt.Errorf("no such data: %q", name)
	}

	if b.Index >= len(ih.Data) {
		return 0, fmt.Errorf("Index out of bounds: have %d, size %d", b.Index, len(ih.Data))
	}

	return ih.Data[b.Index].Value / ih.Data[b.Index-1].Value, nil
}
