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
=== Backtest ===
data "Simulated Portfolio" (returns: 7.7%; volatility: 15.7%; sharpe ratio: 0.49)
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
52% WORLD, 16% USA SMALL CAP VALUE WEIGHTED, 16% WORLD VALUE,  8% WORLD QUALITY,  8% EMERGING MARKETS

=== Monte Carlo ===
[P50] returns: 7.8%; volatility: 15.8%; sharpe ratio: 0.49
[P80] returns: 4.3%; volatility: 14.4%; sharpe ratio: 0.30
[P90] returns: 3.2%; volatility: 15.8%; sharpe ratio: 0.20
[P95] returns: 2.0%; volatility: 15.5%; sharpe ratio: 0.13
[P99] returns: -0.1%; volatility: 19.1%; sharpe ratio: -0.01

=== Markov Chain ===
[P50] returns: 6.4%; volatility: 14.1%; sharpe ratio: 0.46
[P80] returns: 4.1%; volatility: 15.3%; sharpe ratio: 0.27
[P90] returns: 3.0%; volatility: 16.6%; sharpe ratio: 0.18
[P95] returns: 1.9%; volatility: 17.0%; sharpe ratio: 0.11
[P99] returns: -0.4%; volatility: 19.0%; sharpe ratio: -0.02
```

### Optimize allocation

Tool for generating portfolios that perform well with the available data.
The generated portfolios are typically overfitted to the available data and
tend to have only two positions, as this typically maximizes the sharpe ratio.

**Example usage:**

```sh
./optimize-allocation -input=history.csv -size=100 -iterations=2000
```

## Background

### Data

The algorithm uses historic index performance from `history.csv`. The file is
based on data downloaded from the MSCI website. It contains data for the
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
