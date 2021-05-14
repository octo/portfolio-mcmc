package main

import (
	"encoding/csv"
	"io"
	"math"
	"strconv"
	"strings"
	"time"
)

type IndexDatum struct {
	Date  time.Time
	Value float64
}

type IndexHistory struct {
	Name string
	Data []IndexDatum
}

func (h IndexHistory) String() string {
	return h.Name
}

func (h IndexHistory) Average() float64 {
	var ret float64

	for _, d := range h.Data {
		ret += d.Value
	}

	return ret / float64(len(h.Data))
}

func (h IndexHistory) Variance() float64 {
	var ret float64
	avg := h.Average()

	for _, d := range h.Data {
		dev := d.Value - avg
		ret += dev * dev
	}

	return ret / float64(len(h.Data))
}

func (h IndexHistory) StdDev() float64 {
	return math.Sqrt(h.Variance())
}

func (h IndexHistory) Volatility() float64 {
	var (
		relChange []float64
		avg       float64
	)
	for i := 1; i < len(h.Data); i++ {
		v0 := h.Data[i-1].Value
		v1 := h.Data[i].Value

		d := 100 * (v1 - v0) / v0
		relChange = append(relChange, d)
		avg += d
	}
	avg = avg / float64(len(h.Data)-1)

	var variance float64
	for _, v := range relChange {
		diff := v - avg
		variance += diff * diff
	}
	variance = variance / float64(len(relChange))

	stdDev := math.Sqrt(variance)

	annuallized := stdDev * math.Sqrt(12)
	return annuallized
}

func (h IndexHistory) Returns() float64 {
	v0 := h.Data[0].Value
	v1 := h.Data[len(h.Data)-1].Value

	years := float64(len(h.Data)) / 12

	chg := math.Pow(v1/v0, 1/years)
	return 100 * (chg - 1)
}

func (h IndexHistory) SharpeRatio() float64 {
	return h.Returns() / h.Volatility()
}

func (h IndexHistory) Min() float64 {
	var ret float64

	for _, d := range h.Data {
		if ret == 0 || ret > d.Value {
			ret = d.Value
		}
	}

	return ret
}

func (h IndexHistory) Max() float64 {
	var ret float64

	for _, d := range h.Data {
		if ret == 0 || ret < d.Value {
			ret = d.Value
		}
	}

	return ret
}

func (h IndexHistory) Last() float64 {
	if len(h.Data) == 0 {
		return math.NaN()
	}

	return h.Data[len(h.Data)-1].Value
}

func loadHistory(r io.Reader) (map[string]IndexHistory, error) {
	data, err := csv.NewReader(r).ReadAll()
	if err != nil {
		return nil, err
	}

	header := data[0]

	ret := make([]IndexHistory, len(header)-1)
	for i := 1; i < len(header); i++ {
		ret[i-1].Name = header[i]
	}

	for row := 1; row < len(data); row++ {
		t, err := time.Parse("2006-01-02", data[row][0])
		if err != nil {
			return nil, err
		}

		for col := 1; col < len(data[row]); col++ {
			v, err := strconv.ParseFloat(strings.Replace(data[row][col], ",", ".", -1), 64)
			if err != nil {
				return nil, err
			}

			ret[col-1].Data = append(ret[col-1].Data, IndexDatum{
				Date:  t,
				Value: v,
			})
		}
	}

	m := make(map[string]IndexHistory)
	for i := 1; i < len(header); i++ {
		m[header[i]] = ret[i-1]
	}

	return m, nil
}

func generateHistory(names []string, qp QuoteProvider) (map[string]IndexHistory, error) {
	histories := map[string]IndexHistory{}
	for _, name := range names {
		histories[name] = IndexHistory{
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
			if len(h.Data) == 0 {
				h.Data = []IndexDatum{{
					Date: date.AddDate(0, -1, 0),
					Value: 100,
				}}
			}
			value := h.Data[len(h.Data)-1].Value * rv
			h.Data = append(h.Data, IndexDatum{
				Date: date,
				Value: value,
			})
			histories[name] = h
		}
	}

	return histories, nil
}

type BySharpeRatio []IndexHistory

func (b BySharpeRatio) Len() int {
	return len(b)
}

func (b BySharpeRatio) Less(i, j int) bool {
	return b[i].SharpeRatio() < b[j].SharpeRatio()
}

func (b BySharpeRatio) Swap(i, j int) {
	b[i], b[j] = b[j], b[i]
}
