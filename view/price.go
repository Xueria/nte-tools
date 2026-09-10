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
	// priceTabHeight 价位标签的高度。
	priceTabHeight float32 = 28
	// priceTabPadding 价位标签内文字两侧的留白。
	priceTabPadding float32 = 12
)

// priceTab 是价位条上的一个价位标签，点一下就切换查看它的组合。
type priceTab struct {
	widget.BaseWidget

	page     *Bid
	index    int
	option   bid.TotalOption
	selected bool
}

// newPriceTab 构建一个空价位标签，内容由 set 填充。
func newPriceTab(page *Bid) *priceTab {
	tab := &priceTab{page: page}
	tab.ExtendBaseWidget(tab)

	return tab
}

// set 设置标签对应的价位与选中状态。
func (t *priceTab) set(index int, option bid.TotalOption, selected bool) {
	t.index = index
	t.option = option
	t.selected = selected
	t.Refresh()
}

// setSelected 只更新选中状态。
func (t *priceTab) setSelected(selected bool) {
	if t.selected == selected {
		return
	}

	t.selected = selected
	t.Refresh()
}

// Tapped 点击标签即切换价位。
func (t *priceTab) Tapped(*fyne.PointEvent) {
	if t.page != nil {
		t.page.selectPrice(t.index)
	}
}

// MinSize 返回标签尺寸：宽度随总价位数变化。
func (t *priceTab) MinSize() fyne.Size {
	t.ExtendBaseWidget(t)

	return fyne.NewSize(t.textWidth()+priceTabPadding*2, priceTabHeight)
}

// textWidth 返回标签文字需要的宽度。
func (t *priceTab) textWidth() float32 {
	return fyne.MeasureText(priceTabTitle(t.option), compositionTextSize, fyne.TextStyle{}).Width
}

// CreateRenderer 创建标签的绘制对象。
func (t *priceTab) CreateRenderer() fyne.WidgetRenderer {
	background := canvas.NewRectangle(color.Transparent)
	background.CornerRadius = priceTabHeight / 2

	label := canvas.NewText("", color.White)
	label.TextSize = compositionTextSize
	label.Alignment = fyne.TextAlignCenter

	renderer := &priceTabRenderer{
		baseRenderer: baseRenderer{objects: []fyne.CanvasObject{background, label}},
		tab:          t,
		background:   background,
		label:        label,
	}
	renderer.Refresh()

	return renderer
}

type priceTabRenderer struct {
	baseRenderer

	tab        *priceTab
	background *canvas.Rectangle
	label      *canvas.Text
}

func (r *priceTabRenderer) Refresh() {
	th := r.tab.Theme()
	variant := fyne.CurrentApp().Settings().ThemeVariant()

	if r.tab.selected {
		r.background.FillColor = th.Color(theme.ColorNamePrimary, variant)
		r.background.StrokeColor = r.background.FillColor
		r.label.Color = th.Color(theme.ColorNameForegroundOnPrimary, variant)
	} else {
		r.background.FillColor = th.Color(theme.ColorNameInputBackground, variant)
		r.background.StrokeColor = th.Color(theme.ColorNameInputBorder, variant)
		r.label.Color = th.Color(theme.ColorNameForeground, variant)
	}

	r.label.Text = priceTabTitle(r.tab.option)

	canvas.Refresh(r.tab)
}

func (r *priceTabRenderer) Layout(size fyne.Size) {
	lineHeight := compositionTextSize + 4

	r.background.Resize(size)

	r.label.Move(fyne.NewPos(0, (size.Height-lineHeight)/2))
	r.label.Resize(fyne.NewSize(size.Width, lineHeight))
}

func (r *priceTabRenderer) MinSize() fyne.Size {
	return r.tab.MinSize()
}

// priceTabTitle 价位标签上的文字：只显示总价，尽量紧凑。
func priceTabTitle(option bid.TotalOption) string {
	return formatValue(option.Total)
}

// priceOptionLabel 价位的完整说明：总价与能凑出它的件数范围。
func priceOptionLabel(option bid.TotalOption) string {
	if option.MinCount == option.MaxCount {
		return fmt.Sprintf("总价 %s（%d 件）", formatValue(option.Total), option.MinCount)
	}

	return fmt.Sprintf("总价 %s（%d～%d 件）",
		formatValue(option.Total), option.MinCount, option.MaxCount)
}
