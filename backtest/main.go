package main

import (
	"os"
	"log"
	"fmt"

	"github.com/octo/portfolio-mcmc/timeseries"
	"github.com/octo/portfolio-mcmc/portfolio"
)

const inputFile = "history.csv"

func main() {
	f, err := os.Open(inputFile)
	if err != nil {
		log.Fatalf("os.Open(%q): %v", inputFile, err)
	}
	defer f.Close()

	hist, err := timeseries.Load(f)
	if err != nil {
		log.Fatalf("timeseries.Load(): %v", err)
	}

	p := portfolio.Portfolio{
		Positions: []portfolio.Position{
			{"WORLD VALUE", 40000.0},
			{"USA SMALL CAP VALUE WEIGHTED", 40000.0},
			{"WORLD QUALITY", 20000.0},
			{"WORLD MOMENTUM", 0},
		},
	}

	res, err := p.Eval(&timeseries.Backtest{
		Data: hist,
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("=== Backtest ===")
	fmt.Printf("data %q (returns: %.1f%%; volatility: %.1f%%; sharpe ratio: %.2f)\n",
		res, res.Returns(), res.Volatility(), res.SharpeRatio())
}
