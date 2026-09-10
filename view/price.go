package view

import (
	"fmt"
	"image/color"

	"blind-tools/model/bid"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

const (
	// priceRowHeight、priceRowMinWidth 价位行的高度与建议最小宽度。
	priceRowHeight   float32 = 28
	priceRowMinWidth float32 = 200
)

// priceRow 价位列表里的一行：总价与能凑出它的件数范围。
type priceRow struct {
	widget.BaseWidget

	option bid.TotalOption
}

// newPriceRow 构建一行空价位，内容由 set 填充。
func newPriceRow() *priceRow {
	row := &priceRow{}
	row.ExtendBaseWidget(row)

	return row
}

// set 用一条价位填充该行。
func (r *priceRow) set(option bid.TotalOption) {
	r.option = option
	r.Refresh()
}

// MinSize 返回该行的建议最小尺寸。
func (r *priceRow) MinSize() fyne.Size {
	r.ExtendBaseWidget(r)

	return fyne.NewSize(priceRowMinWidth, priceRowHeight)
}

// CreateRenderer 创建该行的绘制对象。
func (r *priceRow) CreateRenderer() fyne.WidgetRenderer {
	label := canvas.NewText("", color.White)
	label.TextSize = compositionTextSize

	renderer := &priceRowRenderer{
		baseRenderer: baseRenderer{objects: []fyne.CanvasObject{label}},
		row:          r,
		label:        label,
	}
	renderer.Refresh()

	return renderer
}

type priceRowRenderer struct {
	baseRenderer

	row   *priceRow
	label *canvas.Text
}

func (r *priceRowRenderer) Refresh() {
	th := r.row.Theme()
	variant := fyne.CurrentApp().Settings().ThemeVariant()

	r.label.Text = priceRowTitle(r.row.option)
	r.label.Color = th.Color(theme.ColorNameForeground, variant)

	canvas.Refresh(r.row)
}

func (r *priceRowRenderer) Layout(size fyne.Size) {
	lineHeight := compositionTextSize + 4

	r.label.Move(fyne.NewPos(cardInset, (size.Height-lineHeight)/2))
	r.label.Resize(fyne.NewSize(size.Width-cardInset*2, lineHeight))
}

func (r *priceRowRenderer) MinSize() fyne.Size {
	return fyne.NewSize(priceRowMinWidth, priceRowHeight)
}

// priceRowTitle 生成价位行的文字：总价与能凑出它的件数范围。
func priceRowTitle(option bid.TotalOption) string {
	if option.MinCount == option.MaxCount {
		return fmt.Sprintf("总价 %s（%d 件）", formatValue(option.Total), option.MinCount)
	}

	return fmt.Sprintf("总价 %s（%d～%d 件）",
		formatValue(option.Total), option.MinCount, option.MaxCount)
}
