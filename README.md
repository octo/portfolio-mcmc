This code is a quick and dirty spare time project to evaluate portfolios of
ETFs using bootstrapping and optimize asset allocation based on these
bootstraps.

## Tools

### Backtest

Tool for backtesting a portfolio.

**Example usage:**

```sh
./backtest -input=history.csv \
  -pos='WORLD:52200' \
  -pos='USA SMALL CAP VALUE WEIGHTED:16000' \
  -pos='WORLD VALUE:16000' \
  -pos='WORLD QUALITY:8000' \
  -pos='EMERGING MARKETS:7800'
```

### Forecast

Tool for forecasting a portfolio.

**Example usage:**

```sh
./forecast -input=history.csv \
  -pos='WORLD:52200' \
  -pos='USA SMALL CAP VALUE WEIGHTED:16000' \
  -pos='WORLD VALUE:16000' \
  -pos='WORLD QUALITY:8000' \
  -pos='EMERGING MARKETS:7800'
```

### Optimize allocation

Tool for optimizing the asset allocation of a portfolio.

**Example usage:**

```sh
./forecast -input=history.csv -size=100 -iterations=2000
```

## Background

### Data

The algorithm is using historic index performance from `history.csv`. The file
is based on data downloaded from the MSCI website. It contains data for the
timespan from January 1999 to April 2021.

### Bootstrapping

This implementation uses a Monte Carlo Markov Chain (MCMC) method. That is a
very fancy way to say that with probability 1/12 a *random* month in the
available data is used, and with probability 11/12 the *following* month is
used. The hope is that this translates some of the inter-month dependence into
the generated data set.

### Optimizing

To optimize asset allocation, the code implements an evolutionary algorithm.
First, 100 completely random portfolios are generated. In each round, a random
30 year sample is generated with the method described above. All portfolios are
tested against this sample and sorted by their Sharpe ratio. The worse half of
portfolios are then replaced by combining two random portfolios of the better
half. Recombination is done by iterating over the positions, randomly picking
the weight of one of the parents. Mutation is done by multiplying each position
with a random number between 95% and 105%.

## Anticipated questions

*   How were the indices chosen?

    Indices were chosen based on whether ETFs based on these indices are
    available in Germany.
*   Why use 12 as the expected length of sequential months?

    Many periodic effects happen yearly, so it felt like not the worst choice
    🤷.
*   You're ignoring the TER, how unrealistic.

    Given the uncertainty of bootstrapping, the difference in TER between fonds
    is negligible.

## Author

Florian Forster (@octo)

## License

ISC License
