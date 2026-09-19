package infer

import (
	"fmt"
	"image/color"
	"strings"

	"blind-tools/internal/bid"
	"blind-tools/internal/ui/common"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

const (
	// compositionRowHeight、compositionRowMinWidth 组合行的高度与建议最小宽度。
	compositionRowHeight   float32 = 42
	compositionRowMinWidth float32 = 240
	// compositionTextSize 组合行与件数标签共用的字号。
	compositionTextSize float32 = 12
	// priorityQuality 关注的品质：含它的组合在结果里标星。
	priorityQuality = "red"
)

// compositionRow 组合列表里的一行：总价、件数、均价，以及具体组成。
type compositionRow struct {
	widget.BaseWidget

	composition bid.Composition
}

// newCompositionRow 构建一行空组合，内容由 set 填充。
func newCompositionRow() *compositionRow {
	row := &compositionRow{}
	row.ExtendBaseWidget(row)

	return row
}

// set 用一条组合填充该行。
func (r *compositionRow) set(composition bid.Composition) {
	r.composition = composition
	r.Refresh()
}

// MinSize 返回该行的建议最小尺寸。
func (r *compositionRow) MinSize() fyne.Size {
	r.ExtendBaseWidget(r)

	return fyne.NewSize(compositionRowMinWidth, compositionRowHeight)
}

// CreateRenderer 创建该行的绘制对象。
func (r *compositionRow) CreateRenderer() fyne.WidgetRenderer {
	title := canvas.NewText("", color.White)
	title.TextSize = compositionTextSize

	items := canvas.NewText("", color.White)
	items.TextSize = compositionTextSize

	renderer := &compositionRowRenderer{
		BaseRenderer: common.NewBaseRenderer(title, items),
		row:          r,
		title:        title,
		items:        items,
	}
	renderer.Refresh()

	return renderer
}

type compositionRowRenderer struct {
	common.BaseRenderer

	row   *compositionRow
	title *canvas.Text
	items *canvas.Text
}

func (r *compositionRowRenderer) Refresh() {
	th := r.row.Theme()
	variant := fyne.CurrentApp().Settings().ThemeVariant()

	r.title.Text = compositionTitle(r.row.composition)
	r.title.Color = th.Color(theme.ColorNameForeground, variant)

	r.items.Color = th.Color(theme.ColorNamePlaceHolder, variant)
	r.items.Text = common.FitText(describeComposition(r.row.composition), r.textWidth(),
		compositionTextSize, fyne.TextStyle{})

	canvas.Refresh(r.row)
}

func (r *compositionRowRenderer) Layout(size fyne.Size) {
	textWidth := r.textWidth()

	r.title.Move(fyne.NewPos(common.CardInset, 5))
	r.title.Resize(fyne.NewSize(textWidth, 16))

	r.items.Move(fyne.NewPos(common.CardInset, 22))
	r.items.Resize(fyne.NewSize(textWidth, 16))

	r.items.Text = common.FitText(describeComposition(r.row.composition), textWidth, compositionTextSize, fyne.TextStyle{})
}

func (r *compositionRowRenderer) MinSize() fyne.Size {
	return fyne.NewSize(compositionRowMinWidth, compositionRowHeight)
}

// textWidth 返回组合行可用于文字的宽度：右侧留出滚动条的位置。
func (r *compositionRowRenderer) textWidth() float32 {
	return r.row.Size().Width - common.CardInset*2 - common.ScrollBarInset()
}

// compositionTitle 生成组合行的摘要：总价、件数与均价；含关注品质的组合加星标。
func compositionTitle(composition bid.Composition) string {
	title := fmt.Sprintf("总价 %s · %d 件 · 均价 %s",
		common.FormatValue(composition.Total), composition.Count, common.FormatAverage(composition.Average()))

	if containsQuality(composition, priorityQuality) {
		return "★ " + title
	}

	return title
}

// containsQuality 判断组合里是否含指定品质。
func containsQuality(composition bid.Composition, quality string) bool {
	for _, group := range composition.Items {
		if group.Item.Quality == quality {
			return true
		}
	}

	return false
}

// describeComposition 把一种组成写成「名称×件数」的形式。
func describeComposition(composition bid.Composition) string {
	parts := make([]string, 0, len(composition.Items))

	for _, group := range composition.Items {
		if group.Count > 1 {
			parts = append(parts, fmt.Sprintf("%s×%d", group.Item.Name, group.Count))
			continue
		}

		parts = append(parts, group.Item.Name)
	}

	return strings.Join(parts, "  ")
}
