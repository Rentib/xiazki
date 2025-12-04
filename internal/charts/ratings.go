package charts

import (
	"strconv"

	"xiazki/internal/model"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
)

func RatingsBar(stats *model.ReviewStats) *charts.Bar {
	var data []opts.BarData
	var axis []string

	for rating := 10; rating >= 1; rating-- {
		count := stats.RatingsSpread[strconv.Itoa(rating)]
		data = append(data, opts.BarData{Value: count})
		axis = append(axis, strconv.Itoa(rating)+" ★")
	}

	bar := charts.NewBar()

	bar.SetGlobalOptions(
		charts.WithLegendOpts(opts.Legend{
			Show: opts.Bool(false),
		}),
		charts.WithXAxisOpts(opts.XAxis{
			Show:      opts.Bool(false),
			SplitLine: &opts.SplitLine{Show: opts.Bool(false)},
			Min:       0,
			Max:       stats.RatingsCount,
		}),
		charts.WithYAxisOpts(opts.YAxis{
			Type:    "category",
			Inverse: opts.Bool(true),
			Data:    axis,
			AxisLabel: &opts.AxisLabel{
				Show:       opts.Bool(true),
				FontSize:   14,
				FontWeight: "bold",
			},
			SplitArea:   &opts.SplitArea{Show: opts.Bool(true)},
			SplitLine:   &opts.SplitLine{Show: opts.Bool(false)},
			AxisLine:    &opts.AxisLine{Show: opts.Bool(false)},
			AxisPointer: &opts.AxisPointer{Show: opts.Bool(false)},
		}),
	)

	bar.AddSeries("ratings", data,
		charts.WithBarChartOpts(opts.BarChart{
			ShowBackground: opts.Bool(true),
			BarWidth:       "50%",
		}),
		charts.WithLabelOpts(opts.Label{Show: opts.Bool(true), Position: "right"}),
	)

	return bar
}
