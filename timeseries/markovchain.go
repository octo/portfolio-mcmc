package timeseries

import (
	"time"
	"fmt"
	"math/rand"
)

// expectedDuration is the expected value for the length of sequential months.
const expectedDuration = 12 // [months]

// MarkovChain implements a bootstrapping method that favor the subsequent
// month over a random month.
// Implements the QuoteProvider interface.
type MarkovChain struct {
	Data map[string]Data

	index int
	date  time.Time
}

// Next advances the time and chooses the next month to return data from.
func (m *MarkovChain) Next() (time.Time, bool) {
	var data []Datum
	for _, ih := range m.Data {
		if len(data) == 0 || len(data) > len(ih.Data) {
			data = ih.Data
			break
		}
	}

	if m.date.IsZero() {
		year, month, _ := time.Now().Date()
		m.date = time.Date(year, month, 1, 0, 0, 0, 0, time.Local)
	}
	m.date = m.date.AddDate(0, 1, 0)

	if m.date.After(time.Now().AddDate(30, 0, 0)) {
		return time.Time{}, false
	}

	if m.index == 0 || m.index >= len(data)-1 || rand.Intn(expectedDuration) == 0 {
		m.index = 1 + rand.Intn(len(data)-1)
	} else {
		m.index++
	}

	return m.date, true
}

// RelativeValue returns the relative change for the position name.  Returns
// 1.0 if there is no change.
func (m *MarkovChain) RelativeValue(name string) (float64, error) {
	ih, ok := m.Data[name]
	if !ok {
		return 0, fmt.Errorf("no such data: %q", name)
	}

	if m.index < 1 || m.index >= len(ih.Data) {
		return 0, fmt.Errorf("index out of bounds: have %d, size %d", m.index, len(ih.Data))
	}

	return ih.Data[m.index].Value / ih.Data[m.index-1].Value, nil
}
