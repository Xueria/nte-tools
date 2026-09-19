package ui

import (
	"fmt"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

const (
	// countTabHeight 件数标签的高度。
	countTabHeight float32 = 28
	// countTabPadding 件数标签内文字两侧的留白。
	countTabPadding float32 = 12
)

// countTab 是件数条上的一个标签，点一下就切换查看该件数的组合。
type countTab struct {
	widget.BaseWidget

	page     *InferPage
	index    int
	count    int
	selected bool
}

// newCountTab 构建一个空件数标签，内容由 set 填充。
func newCountTab(page *InferPage) *countTab {
	tab := &countTab{page: page}
	tab.ExtendBaseWidget(tab)

	return tab
}

// set 设置标签对应的件数与选中状态。
func (t *countTab) set(index, count int, selected bool) {
	t.index = index
	t.count = count
	t.selected = selected
	t.Refresh()
}

// setSelected 只更新选中状态。
func (t *countTab) setSelected(selected bool) {
	if t.selected == selected {
		return
	}

	t.selected = selected
	t.Refresh()
}

// Tapped 点击标签即切换件数。
func (t *countTab) Tapped(*fyne.PointEvent) {
	if t.page != nil {
		t.page.selectCount(t.index)
	}
}

// MinSize 返回标签尺寸：宽度随件数位数变化。
func (t *countTab) MinSize() fyne.Size {
	t.ExtendBaseWidget(t)

	return fyne.NewSize(t.textWidth()+countTabPadding*2, countTabHeight)
}

// textWidth 返回标签文字需要的宽度。
func (t *countTab) textWidth() float32 {
	return fyne.MeasureText(countTabTitle(t.count), compositionTextSize, fyne.TextStyle{}).Width
}

// CreateRenderer 创建标签的绘制对象。
func (t *countTab) CreateRenderer() fyne.WidgetRenderer {
	background := canvas.NewRectangle(color.Transparent)
	background.CornerRadius = countTabHeight / 2

	label := canvas.NewText("", color.White)
	label.TextSize = compositionTextSize
	label.Alignment = fyne.TextAlignCenter

	renderer := &countTabRenderer{
		baseRenderer: baseRenderer{objects: []fyne.CanvasObject{background, label}},
		tab:          t,
		background:   background,
		label:        label,
	}
	renderer.Refresh()

	return renderer
}

type countTabRenderer struct {
	baseRenderer

	tab        *countTab
	background *canvas.Rectangle
	label      *canvas.Text
}

func (r *countTabRenderer) Refresh() {
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

	r.label.Text = countTabTitle(r.tab.count)

	canvas.Refresh(r.tab)
}

func (r *countTabRenderer) Layout(size fyne.Size) {
	lineHeight := compositionTextSize + 4

	r.background.Resize(size)

	r.label.Move(fyne.NewPos(0, (size.Height-lineHeight)/2))
	r.label.Resize(fyne.NewSize(size.Width, lineHeight))
}

func (r *countTabRenderer) MinSize() fyne.Size {
	return r.tab.MinSize()
}

// countTabTitle 件数标签上的文字。
func countTabTitle(count int) string {
	return fmt.Sprintf("%d 件", count)
}
