package main

import (
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"time"
)

type Portfolio struct {
	Positions []Position
}

type Position struct {
	Name  string
	Value float64
}

type QuoteProvider interface {
	Next() (time.Time, bool)
	RelativeValue(string) (float64, error)
}

func (p Portfolio) String() string {
	var sum float64
	for _, pos := range p.Positions {
		sum += pos.Value
	}

	var b strings.Builder
	for i, pos := range p.Positions {
		if i != 0 {
			fmt.Fprint(&b, ", ")
		}
		fmt.Fprintf(&b, "%.0f%% %s", 100*pos.Value/sum, pos.Name)
	}

	return b.String()
}

func (p Portfolio) CSV() string {
	var fields []string

	var sum float64
	for _, pos := range p.Positions {
		sum += pos.Value
	}
	for _, pos := range p.Positions {
		fields = append(fields, fmt.Sprintf("%.1f", 100 * pos.Value / sum))
	}

	return strings.Join(fields, ",")
}

func (p Portfolio) Eval(qp QuoteProvider) (IndexHistory, error) {
	positions := make([]Position, len(p.Positions))
	copy(positions, p.Positions)

	ret := IndexHistory{
		Name: "Simulated Portfolio",
	}

	for {
		date, ok := qp.Next()
		if !ok {
			break
		}

		var sum float64
		for i := 0; i < len(positions); i++ {
			rv, err := qp.RelativeValue(positions[i].Name)
			if err != nil {
				return IndexHistory{}, err
			}

			positions[i].Value *= rv
			sum += positions[i].Value
			// fmt.Printf("[%v] %q %.0f (%5.1f%%)\n", date, positions[i].Name, positions[i].Value, 100*(rv-1))
		}

		ret.Data = append(ret.Data, IndexDatum{
			Date:  date,
			Value: sum,
		})
	}

	return ret, nil
}

// Recombine combines two portfolios, p0 and p1, to create a "child" portfolio.
// Recombination is done by calculating the average for each position. Mutation
// is done by multiplying each position with a random number between 95% and
// 105%.
func Recombine(p0, p1 Portfolio) Portfolio {
	positions := map[string]float64{}

	for _, p := range p0.Positions {
		positions[p.Name] += p.Value
	}
	for _, p := range p1.Positions {
		positions[p.Name] += p.Value
	}

	// mutate
	for name, value := range positions {
		positions[name] = (0.95 * value) + (0.1 * rand.Float64() * value)
	}

	// renormalize
	var sum float64
	for _, value := range positions {
		sum += value
	}
	var names []string
	for name, value := range positions {
		names = append(names, name)
		positions[name] = 100000 * value / sum
	}
	sort.Strings(names)

	var ret Portfolio
	for _, name := range names {
		ret.Positions = append(ret.Positions, Position{
			Name:  name,
			Value: positions[name],
		})
	}
	return ret
}
