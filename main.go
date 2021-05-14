package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"sort"
	"strings"
	"time"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	f, err := os.Open("history.csv")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	hist, err := loadHistory(f)
	if err != nil {
		log.Fatal(err)
	}

	if err := evolve(hist); err != nil {
		log.Fatal("evolve: ", err)
	}
	return

	p := Portfolio{
		Positions: []Position{
			{"WORLD VALUE", 40000.0},
			{"USA SMALL CAP VALUE WEIGHTED", 40000.0},
			{"WORLD QUALITY", 20000.0},
			{"WORLD MOMENTUM", 0},
		},
	}

	res, err := p.Eval(&Backtest{
		Data: hist,
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("=== Backtest ===")
	fmt.Printf("data %q (returns: %.1f%%; volatility: %.1f%%; sharpe ratio: %.2f)\n",
		res, res.Returns(), res.Volatility(), res.SharpeRatio())

	fmt.Println("=== Markov Chains ===")
	var results []IndexHistory
	for i := 0; i < 50; i++ {
		res, err := p.Eval(&MarkovChain{
			Data: hist,
		})
		if err != nil {
			log.Fatal("Eval: ", err)
		}

		results = append(results, res)
	}

	sort.Sort(BySharpeRatio(results))
	for _, res := range results {
		fmt.Printf("data %q (returns: %.1f%%; volatility: %.1f%%; sharpe ratio: %.2f)\n",
			res, res.Returns(), res.Volatility(), res.SharpeRatio())
	}

}

type Individual struct {
	Portfolio
	Returns, Volatility, SharpeRatio float64
}

func (i Individual) String() string {
	return fmt.Sprintf("%s\n    Returns: %4.1f, Volatility: %4.1f, Sharpe Ratio: %4.2f",
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

func evolve(hist map[string]IndexHistory) error {
	var names []string
	for name := range hist {
		names = append(names, name)
	}
	sort.Strings(names)
	fmt.Println(strings.Join(names, ","))

	pop := &Population{}
	for i := 0; i < 100; i++ {
		pop.Individuals = append(pop.Individuals, &Individual{
			Portfolio: randomPortfolio(names),
		})
	}

	for k := 0; k < 2000; k++ {
		genHist, err := generateHistory(names, &MarkovChain{
			Data: hist,
		})
		if err != nil {
			return fmt.Errorf("generateHistory: %w", err)
		}

		for _, ind := range pop.Individuals {
			h, err := ind.Portfolio.Eval(&Backtest{
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

			pop.Individuals[i].Portfolio = Recombine(
				pop.Individuals[parent0].Portfolio, pop.Individuals[parent1].Portfolio)
		}
	}

	return nil
}
