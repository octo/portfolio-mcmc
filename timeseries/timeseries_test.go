package timeseries

import (
	"math"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestLoad(t *testing.T) {
	input := `Date,FONDS 0,FONDS 1
1999-01-29,"5,648","4,161"
1999-02-26,"0,686","2,190"
`

	want := map[string]Data{
		"FONDS 0": Data{
			Name: "FONDS 0",
			Data: []Datum{
				{time.Date(1999, time.January, 29, 0, 0, 0, 0, time.UTC), 0.05648},
				{time.Date(1999, time.February, 26, 0, 0, 0, 0, time.UTC), 0.00686},
			},
		},
		"FONDS 1": Data{
			Name: "FONDS 1",
			Data: []Datum{
				{time.Date(1999, time.January, 29, 0, 0, 0, 0, time.UTC), 0.04161},
				{time.Date(1999, time.February, 26, 0, 0, 0, 0, time.UTC), 0.02190},
			},
		},
	}

	got, err := Load(strings.NewReader(input))
	if err != nil {
		t.Fatal("Load(): ", err)
	}

	if diff := cmp.Diff(want, got, cmpopts.EquateApprox(0, 0.00001)); diff != "" {
		t.Errorf("Load(): result differs (-want/+got):\n%s", diff)
	}
}

func TestReturns(t *testing.T) {
	cases := []struct {
		name   string
		values []float64
		want   float64
	}{
		{
			name:   "zero",
			values: []float64{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			want:   0,
		},
		{
			name:   "no variance",
			values: []float64{.05, .05, .05, .05, .05, .05, .05, .05, .05, .05, .05, .05},
			want:   100.0 * (math.Pow(1.05, 12) - 1.0),
		},
		{
			name:   "alternate",
			values: []float64{.05, -.05, .05, -.05, .05, -.05, .05, -.05, .05, -.05, .05, -.05},
			want:   -1.490656191,
		},
		{
			name:   "annualized",
			values: []float64{.03, .03, .03, .03, .03, .03},
			want:   100.0 * (math.Pow(1.03, 12) - 1.0),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := newTestData(tc.name, tc.values).Returns()
			if !cmp.Equal(got, tc.want, cmpopts.EquateApprox(0, 0.00001)) {
				t.Errorf("Returns(%v) = %.5f, want %.5f", tc.values, got, tc.want)
			}
		})
	}
}

func TestVolatility(t *testing.T) {
	cases := []struct {
		name     string
		values   []float64
		wantAvg  float64
		wantVar  float64
		wantVola float64
	}{
		{
			name:     "zero",
			values:   []float64{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			wantAvg:  0,
			wantVar:  0,
			wantVola: 0,
		},
		{
			name:     "no variance",
			values:   []float64{.05, .05, .05, .05, .05, .05, .05, .05, .05, .05, .05, .05},
			wantAvg:  0.05,
			wantVar:  0,
			wantVola: 0,
		},
		{
			name:     "alternate",
			values:   []float64{.05, -.05, .05, -.05, .05, -.05, .05, -.05, .05, -.05, .05, -.05},
			wantAvg:  0,
			wantVar:  .0025, // = 12 * 0.05^2 / 12 = 0.05^2
			wantVola: 100.0 * 0.05 * math.Sqrt(12),
		},
		{
			name:     "annualized",
			values:   []float64{.01, .02, .03, .04, .05},
			wantAvg:  .03,
			wantVar:  0.0002, // = (2*0.02^2 + 2*0.01^2) / 5
			wantVola: 100.0 * math.Sqrt(.0002) * math.Sqrt(12),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ts := newTestData(tc.name, tc.values)

			if got := ts.average(); !cmp.Equal(got, tc.wantAvg, cmpopts.EquateApprox(0, 0.00001)) {
				t.Errorf("average(%v) = %.5f, want %.5f", tc.values, got, tc.wantAvg)
			}
			if got := ts.variance(); !cmp.Equal(got, tc.wantVar, cmpopts.EquateApprox(0, 0.00001)) {
				t.Errorf("variance(%v) = %.5f, want %.5f", tc.values, got, tc.wantVar)
			}
			if got := ts.Volatility(); !cmp.Equal(got, tc.wantVola, cmpopts.EquateApprox(0, 0.00001)) {
				t.Errorf("Volatility(%v) = %.5f, want %.5f", tc.values, got, tc.wantVola)
			}
		})
	}
}

func newTestData(name string, values []float64) Data {
	tm := time.Date(1999, time.January, 31, 0, 0, 0, 0, time.UTC)

	d := Data{
		Name: name,
	}
	for _, v := range values {
		d.Data = append(d.Data, Datum{
			Date:  tm,
			Value: v,
		})
		tm = tm.AddDate(0, 1, 0)
	}

	return d
}
