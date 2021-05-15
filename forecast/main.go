package main

import (
	"fmt"
	"math/rand"
	"flag"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/octo/portfolio-mcmc/timeseries"
	"github.com/octo/portfolio-mcmc/portfolio"
)

// expectedDuration is the expected value for the length of sequential months.
const expectedDuration = 12 // [months]

var (
	input = flag.String("input", "history.csv", "file containing historic returns")

	pf = portfolio.Portfolio{}
)

func addPosition(flagValue string) error {
	fields := strings.Split(flagValue, ":")
	if len(fields) != 2 {
		return fmt.Errorf(`got %q, want "<name>:<weight>"`, flagValue)
	}

	weight, err := strconv.ParseFloat(fields[1], 64)
	if err != nil {
		return fmt.Errorf("ParseFloat(%q): %w", fields[1], err)
	}

	pf.Positions = append(pf.Positions, portfolio.Position{
		Name:  fields[0],
		Value: weight,
	})
	return nil
}

func main() {
	flag.Func("pos", `position as "name:weight"`, addPosition)
	flag.Parse()

	f, err := os.Open(*input)
	if err != nil {
		log.Fatalf("os.Open(%q): %v", *input, err)
	}
	defer f.Close()

	hist, err := timeseries.Load(f)
	if err != nil {
		log.Fatalf("timeseries.Load(): %v", err)
	}

	if len(pf.Positions) == 0 {
		var names []string
		for name := range hist {
			names = append(names, name)
		}
		sort.Strings(names)

		fmt.Println("ERROR: specify one or more -pos arguments.")
		fmt.Println()
		fmt.Println("Available time series:")
		fmt.Println()
		for _, name := range names {
			fmt.Println("  *", name)
		}

		return
	}

	fmt.Println("=== Markov Chains ===")
	fmt.Println(pf)
	var results []timeseries.Data
	for i := 0; i < 50; i++ {
		res, err := pf.Eval(&MarkovChain{
			Data: hist,
		})
		if err != nil {
			log.Fatal("Eval: ", err)
		}

		results = append(results, res)
	}

	sort.Sort(timeseries.BySharpeRatio(results))
	for _, res := range results {
		fmt.Printf("data %q (returns: %.1f%%; volatility: %.1f%%; sharpe ratio: %.2f)\n",
			res, res.Returns(), res.Volatility(), res.SharpeRatio())
	}
}

// MarkovChain implements a bootstrapping method that favor the subsequent
// month over a random month.
// Implements the QuoteProvider interface.
type MarkovChain struct {
	Data map[string]timeseries.Data

	index int
	date  time.Time
}

// Next advances the time and chooses the next month to return data from.
func (m *MarkovChain) Next() (time.Time, bool) {
	var data []timeseries.Datum
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
