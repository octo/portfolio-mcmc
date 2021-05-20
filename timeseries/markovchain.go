package timeseries

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"sort"
	"strings"
	"time"
)

// MarkovChain implements a bootstrapping method. The monthly returns are the
// states. The probability of the next states is determines from an input
// sequence.
// Implements the QuoteProvider interface.
type MarkovChain struct {
	data map[int]edges

	returnsPermille int
	date            time.Time
}

func NewMarkovChain(data Data) *MarkovChain {
	year, month, _ := time.Now().Date()
	mc := &MarkovChain{
		data: make(map[int]edges),
		date: time.Date(year, month, 1, 0, 0, 0, 0, time.Local),
	}

	months := len(data.Data)
	for i := 1; i < months-1; i++ {
		v0 := data.Data[i-1].Value
		v1 := data.Data[i].Value
		v2 := data.Data[i+1].Value

		prevReturns := calcReturnsPermille(v0, v1)
		nextReturns := calcReturnsPermille(v1, v2)

		mc.data[prevReturns] = mc.data[prevReturns].add(nextReturns)
	}

	for _, e := range mc.data {
		e.normalize()
	}

	var keys []int
	for k := range mc.data {
		keys = append(keys, k)
	}
	mc.returnsPermille = keys[rand.Intn(len(keys))]

	/*
		sort.Ints(keys)
		for _, k := range keys {
			fmt.Println(k, "->", mc.data[k])
		}
	*/

	return mc
}

func calcReturnsPermille(t0, t1 float64) int {
	returns := t1/t0 - 1
	return int(math.Round(returns * 1000))
}

// Next advances the time and transitions to the next state.
func (m *MarkovChain) Next() (time.Time, bool) {
	m.date = m.date.AddDate(0, 1, 0)
	if m.date.After(time.Now().AddDate(30, 0, 0)) {
		return time.Time{}, false
	}

	m.returnsPermille = m.data[m.returnsPermille].next()
	return m.date, true
}

// RelativeValue returns the relative change for the position name.  Returns
// 1.0 if there is no change.
func (m *MarkovChain) RelativeValue(_ string) (float64, error) {
	return 1 + (float64(m.returnsPermille) / 1000), nil
}

type edge struct {
	returnsPermille int
	weight          float64
}

type edges []edge

func (e edges) add(returnsPermille int) edges {
	// e may be nil!
	for _, ee := range e {
		if ee.returnsPermille == returnsPermille {
			ee.weight += 1.0
			return e
		}
	}

	return append(e, edge{
		returnsPermille: returnsPermille,
		weight:          1.0,
	})
}

// normalize normalizes the weight of each edge to [0...1). During
// construction, the "weight" field will contain the number of outgoing edges
// with those returns, normalize converts that count into a probability.
func (e edges) normalize() {
	var sum float64
	for _, ee := range e {
		sum += ee.weight
	}
	if sum == 1 {
		return
	}

	for i := range e {
		e[i].weight = e[i].weight / sum
	}
}

func (e edges) next() int {
	r := rand.Float64()
	for _, ee := range e {
		// this assumes the weight has been normalized.
		if r < ee.weight {
			return ee.returnsPermille
		}
		r -= ee.weight
	}
	log.Fatal("this should be unreachable. r = %g, e = %#v", r, e)
	return 0
}

func (e edges) String() string {
	sort.Sort(e)

	var b strings.Builder
	for i, ee := range e {
		if i != 0 {
			fmt.Fprint(&b, ", ")
		}
		fmt.Fprintf(&b, "%.1f%% (p %.0f%%)", float64(ee.returnsPermille)/10, 100*ee.weight)
	}

	return "[" + b.String() + "]"
}

func (e edges) Len() int           { return len(e) }
func (e edges) Less(i, j int) bool { return e[i].returnsPermille < e[j].returnsPermille }
func (e edges) Swap(i, j int)      { e[i], e[j] = e[j], e[i] }
