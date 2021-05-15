package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/octo/portfolio-mcmc/portfolio"
	"github.com/octo/portfolio-mcmc/timeseries"
)

var (
	input          = flag.String("input", "history.csv", "file containing historic returns")
	populationSize = flag.Int("size", 100, "population size")
	iterations     = flag.Int("iterations", 2000, "number of iterations")
)

func main() {
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

	if err := evolve(hist); err != nil {
		log.Fatal("evolve: ", err)
	}
}

type Individual struct {
	portfolio.Portfolio
	Returns, Volatility, SharpeRatio float64
}

func (i Individual) String() string {
	return fmt.Sprintf("%s (%4.1f/%4.1f/%4.2f)",
		i.Portfolio, i.Returns, i.Volatility, i.SharpeRatio)
}

type Population struct {
	Individuals []*Individual
}

func (p *Population) Len() int {
	return len(p.Individuals)
}

func (p *Population) Less(i, j int) bool {
	return p.Individuals[i].SharpeRatio < p.Individuals[j].SharpeRatio
}

func (p *Population) Swap(i, j int) {
	p.Individuals[i], p.Individuals[j] = p.Individuals[j], p.Individuals[i]
}

func evolve(hist map[string]timeseries.Data) error {
	var names []string
	for name := range hist {
		names = append(names, name)
	}
	sort.Strings(names)
	fmt.Println(strings.Join(names, ","))

	pop := &Population{}
	for i := 0; i < *populationSize; i++ {
		pop.Individuals = append(pop.Individuals, &Individual{
			Portfolio: portfolio.Random(names),
		})
	}

	for k := 0; k < *iterations; k++ {
		genHist, err := timeseries.Generate(names, &timeseries.MarkovChain{
			Data: hist,
		})
		if err != nil {
			return fmt.Errorf("timeseries.Generate: %w", err)
		}

		for _, ind := range pop.Individuals {
			h, err := ind.Portfolio.Eval(&timeseries.Backtest{
				Data: genHist,
			})
			if err != nil {
				return fmt.Errorf("Portfolio.Eval: %w", err)
			}

			ind.Returns = h.Returns()
			ind.Volatility = h.Volatility()
			ind.SharpeRatio = h.SharpeRatio()
		}

		sort.Sort(pop)

		fmt.Println(pop.Individuals[len(pop.Individuals)-1])
		// fmt.Println(pop.Individuals[len(pop.Individuals)-1].Portfolio.CSV())

		// replace the worse half of the population.
		num := len(pop.Individuals) / 2
		for i := 0; i < num; i++ {
			parent0 := num + rand.Intn(len(pop.Individuals)-num)
			parent1 := num + rand.Intn(len(pop.Individuals)-num)

			pop.Individuals[i].Portfolio = portfolio.Recombine(
				pop.Individuals[parent0].Portfolio, pop.Individuals[parent1].Portfolio)
		}
	}

	return nil
}
