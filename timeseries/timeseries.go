package timeseries

import (
	"encoding/csv"
	"io"
	"math"
	"strconv"
	"strings"
	"time"
)

// TODO: create a new type for map[string]Data.

type Datum struct {
	Date  time.Time
	Value float64
}

// Data holds timeseries data.
type Data struct {
	Name string
	Data []Datum
}

func (h Data) String() string {
	return h.Name
}

func (h Data) average() float64 {
	var ret float64

	for _, d := range h.Data {
		ret += d.Value
	}

	return ret / float64(len(h.Data))
}

func (h Data) variance() float64 {
	var ret float64
	avg := h.average()

	for _, d := range h.Data {
		dev := d.Value - avg
		ret += dev * dev
	}

	return ret / float64(len(h.Data))
}

func (h Data) stdDev() float64 {
	return math.Sqrt(h.variance())
}

func (h Data) Volatility() float64 {
	annuallized := h.stdDev() * math.Sqrt(12)
	return 100 * annuallized
}

func (h Data) Returns() float64 {
	var compounded float64 = 1.0

	for _, d := range h.Data {
		compounded *= 1.0 + d.Value
	}

	years := float64(len(h.Data)) / 12

	annualized := math.Pow(compounded, 1/years)
	return 100 * (annualized - 1)
}

func (h Data) SharpeRatio() float64 {
	return h.Returns() / h.Volatility()
}

func (h Data) Min() float64 {
	var ret float64

	for _, d := range h.Data {
		if ret == 0 || ret > d.Value {
			ret = d.Value
		}
	}

	return ret
}

func (h Data) Max() float64 {
	var ret float64

	for _, d := range h.Data {
		if ret == 0 || ret < d.Value {
			ret = d.Value
		}
	}

	return ret
}

func (h Data) Last() float64 {
	if len(h.Data) == 0 {
		return math.NaN()
	}

	return h.Data[len(h.Data)-1].Value
}

// Load loads timeseries data from an io.Reader.
func Load(r io.Reader) (map[string]Data, error) {
	data, err := csv.NewReader(r).ReadAll()
	if err != nil {
		return nil, err
	}

	header := data[0]

	ret := make([]Data, len(header)-1)
	for i := 1; i < len(header); i++ {
		ret[i-1].Name = header[i]
	}

	for row := 1; row < len(data); row++ {
		t, err := time.Parse("2006-01-02", data[row][0])
		if err != nil {
			return nil, err
		}

		for col := 1; col < len(data[row]); col++ {
			// accept comma as decimal separator, too.
			v, err := strconv.ParseFloat(strings.Replace(data[row][col], ",", ".", -1), 64)
			if err != nil {
				return nil, err
			}

			ret[col-1].Data = append(ret[col-1].Data, Datum{
				Date:  t,
				Value: v / 100,
			})
		}
	}

	m := make(map[string]Data)
	for i := 1; i < len(header); i++ {
		m[header[i]] = ret[i-1]
	}

	return m, nil
}

type QuoteProvider interface {
	Next() (time.Time, bool)
	RelativeValue(string) (float64, error)
}

// Generate uses a QuoteProvider to generate a data set.
func Generate(names []string, qp QuoteProvider) (map[string]Data, error) {
	histories := map[string]Data{}
	for _, name := range names {
		histories[name] = Data{
			Name: name,
		}
	}

	for {
		date, ok := qp.Next()
		if !ok {
			break
		}

		for _, name := range names {
			rv, err := qp.RelativeValue(name)
			if err != nil {
				return nil, err
			}

			h := histories[name]
			h.Data = append(h.Data, Datum{
				Date:  date,
				Value: rv - 1.0,
			})
			histories[name] = h
		}
	}

	return histories, nil
}

// BySharpeRatio allows to sort a Data slice by their Sharpe Ratio.
type BySharpeRatio []Data

func (b BySharpeRatio) Len() int {
	return len(b)
}

func (b BySharpeRatio) Less(i, j int) bool {
	return b[i].SharpeRatio() < b[j].SharpeRatio()
}

func (b BySharpeRatio) Swap(i, j int) {
	b[i], b[j] = b[j], b[i]
}
