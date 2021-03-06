package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"sort"

	"github.com/octo/portfolio-mcmc/portfolio"
	"github.com/octo/portfolio-mcmc/timeseries"
)

var (
	input = flag.String("input", "history.csv", "file containing historic returns")

	pf = portfolio.Portfolio{}
)

func main() {
	flag.Func("pos", `position as "name:weight"`, pf.FlagFunc())
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

	res, err := pf.Eval(&timeseries.Backtest{
		Data: hist,
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("=== Backtest ===")
	fmt.Printf("data %q (returns: %.1f%%; volatility: %.1f%%; sharpe ratio: %.2f)\n",
		res, res.Returns(), res.Volatility(), res.SharpeRatio())
}
