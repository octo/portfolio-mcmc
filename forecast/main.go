package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"sort"
	"time"

	"github.com/octo/portfolio-mcmc/portfolio"
	"github.com/octo/portfolio-mcmc/timeseries"
)

const iterations = 10000

var (
	input = flag.String("input", "history.csv", "file containing historic returns")

	pf = portfolio.Portfolio{}
)

func main() {
	flag.Func("pos", `position as "name:weight"`, pf.FlagFunc())
	flag.Parse()
	rand.Seed(time.Now().UnixNano())

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

	fmt.Println("=== Monte Carlo ===")
	fmt.Println(pf)
	var results []timeseries.Data
	for i := 0; i < iterations; i++ {
		res, err := pf.Eval(&timeseries.MonteCarlo{
			Data: hist,
		})
		if err != nil {
			log.Fatal("Eval: ", err)
		}

		results = append(results, res)
	}

	sort.Sort(timeseries.BySharpeRatio(results))

	printResult(50, results)
	printResult(80, results)
	printResult(90, results)
	printResult(95, results)
	printResult(99, results)

	fmt.Println("=== Markov Chain ===")

	data, err := pf.Eval(&timeseries.Backtest{
		Data: hist,
	})
	if err != nil {
		log.Fatal(err)
	}

	results = nil
	for i := 0; i < iterations; i++ {
		res, err := pf.Eval(timeseries.NewMarkovChain(data))
		if err != nil {
			log.Fatal("Eval: ", err)
		}

		results = append(results, res)
	}

	sort.Sort(timeseries.BySharpeRatio(results))

	printResult(50, results)
	printResult(80, results)
	printResult(90, results)
	printResult(95, results)
	printResult(99, results)
}

func printResult(p int, results []timeseries.Data) {
	idx := len(results) * (100 - p) / 100

	fmt.Printf("[P%d] returns: %.1f%%; volatility: %.1f%%; sharpe ratio: %.2f\n",
		p,
		results[idx].Returns(),
		results[idx].Volatility(),
		results[idx].SharpeRatio())
}
