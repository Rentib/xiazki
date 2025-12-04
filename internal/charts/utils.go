package charts

import (
	"context"
	"io"

	"github.com/a-h/templ"
	"github.com/go-echarts/go-echarts/v2/render"
)

func ToComponent(chart render.Renderer) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		snip := chart.RenderSnippet()
		return templ.Join(
			templ.Raw(snip.Element),
			templ.Raw(snip.Script),
			// templ.Raw(snip.Option),
		).Render(ctx, w)
	})
}
