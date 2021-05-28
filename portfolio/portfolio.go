// Package portfolio implements data structures and methods to work with portfolios.
package portfolio

import (
	"fmt"
	"math/rand"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/octo/portfolio-mcmc/timeseries"
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
		fmt.Fprintf(&b, "%2.0f%% %s", 100*pos.Value/sum, pos.Name)
	}

	return b.String()
}

func (p Portfolio) Position(name string) float64 {
	for _, pos := range p.Positions {
		if pos.Name == name {
			return pos.Value
		}
	}
	return 0
}

func (p Portfolio) CSV() string {
	var fields []string

	var sum float64
	for _, pos := range p.Positions {
		sum += pos.Value
	}
	for _, pos := range p.Positions {
		fields = append(fields, fmt.Sprintf("%.1f", 100*pos.Value/sum))
	}

	return strings.Join(fields, ",")
}

func (p Portfolio) Eval(qp QuoteProvider) (timeseries.Data, error) {
	positions := make([]Position, len(p.Positions))
	copy(positions, p.Positions)

	ret := timeseries.Data{
		Name: "Simulated Portfolio",
	}

	var prevValue float64
	for _, p := range positions {
		prevValue += p.Value
	}

	for {
		date, ok := qp.Next()
		if !ok {
			break
		}

		var nextValue float64
		for i := 0; i < len(positions); i++ {
			rv, err := qp.RelativeValue(positions[i].Name)
			if err != nil {
				return timeseries.Data{}, err
			}

			positions[i].Value *= rv
			nextValue += positions[i].Value
		}

		ret.Data = append(ret.Data, timeseries.Datum{
			Date:  date,
			Value: nextValue/prevValue - 1,
		})
		prevValue = nextValue
	}

	return ret, nil
}

// FlagFunc returns a function that can be passed to flag.Func() for flag parsing.
func (p *Portfolio) FlagFunc() func(string) error {
	return func(flagValue string) error {
		fields := strings.Split(flagValue, ":")
		if len(fields) != 2 {
			return fmt.Errorf(`got %q, want "<name>:<weight>"`, flagValue)
		}

		weight, err := strconv.ParseFloat(fields[1], 64)
		if err != nil {
			return fmt.Errorf("ParseFloat(%q): %w", fields[1], err)
		}

		p.Positions = append(p.Positions, Position{
			Name:  fields[0],
			Value: weight,
		})
		return nil
	}
}

// Recombine combines two portfolios, p0 and p1, to create a "child" portfolio.
// Recombination is done by iterating over the positions, randomly picking the
// weight of one of the parents. Mutation is done by multiplying each position
// with a random number between 95% and 105%.
func Recombine(p0, p1 Portfolio) Portfolio {
	nameMap := map[string]bool{}
	for _, p := range p0.Positions {
		nameMap[p.Name] = true
	}
	for _, p := range p1.Positions {
		nameMap[p.Name] = true
	}
	var names []string
	for name := range nameMap {
		names = append(names, name)
	}
	sort.Strings(names)

	var (
		positions = map[string]float64{}
		sum       float64
	)
	for _, name := range names {
		if rand.Float64() < .5 {
			positions[name] = p0.Position(name)
		} else {
			positions[name] = p1.Position(name)
		}

		// mutate
		value := positions[name]
		positions[name] = value * (0.95 + 0.1*rand.Float64())

		sum += positions[name]
	}

	var ret Portfolio
	for _, name := range names {
		ret.Positions = append(ret.Positions, Position{
			Name:  name,
			Value: 100000 * positions[name] / sum,
		})
	}
	return ret
}

// Random generates a random portfolio.
func Random(names []string) Portfolio {
	var p Portfolio
	for _, name := range names {
		p.Positions = append(p.Positions, Position{
			Name: name,
		})
	}

	for i := 0; i < 100; i++ {
		idx := rand.Intn(len(p.Positions))
		p.Positions[idx].Value += 1000
	}

	return p
}
