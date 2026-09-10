package view

import (
	"fmt"
	"image/color"
	"strconv"
	"strings"

	"blind-tools/model/bid"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

const (
	// compositionRowHeight、compositionRowMinWidth 结果行的高度与建议最小宽度。
	compositionRowHeight   float32 = 42
	compositionRowMinWidth float32 = 240
	// compositionTextSize 结果行的字号。
	compositionTextSize float32 = 12
)

// Bid 是「单格推测」页：输入单格均价与单格物品数量区间，列出可能的物品组成。
type Bid struct {
	root fyne.CanvasObject

	avgEntry    *widget.Entry
	minEntry    *widget.Entry
	maxEntry    *widget.Entry
	statusLabel *widget.Label
	headerLabel *widget.Label
	resultList  *widget.List

	cellItemCount int
	results       []bid.Composition
	truncated     bool
	inferred      bool

	// OnInfer 由装配层赋值：用户点「推测」时触发，页面本身不做推测。
	OnInfer func(avg, minCount, maxCount int)
}

// NewBid 构建单格推测页。
func NewBid() *Bid {
	v := &Bid{}
	v.root = v.build()
	return v
}

// NewTab 返回该页的页签标题、图标与内容。
func (v *Bid) NewTab() Tab {
	return Tab{Title: "单格推测", Icon: theme.SearchIcon(), Content: v.root}
}

// SetCellItemCount 告诉页面当前可用的单格物品数量。
func (v *Bid) SetCellItemCount(count int) {
	v.cellItemCount = count
	v.updateHeader()
}

// SetCompositions 展示一次推测的结果。
func (v *Bid) SetCompositions(result bid.InferResult) {
	v.results = result.Compositions
	v.truncated = result.Truncated
	v.updateHeader()

	v.resultList.Refresh()
	// 用 ScrollToOffset 而非 ScrollToTop：后者在渲染器尚未创建时会解引用空的 scroller。
	v.resultList.ScrollToOffset(0)
}

// SetStatus 显示提示或错误；传空字符串即隐藏。
func (v *Bid) SetStatus(text string) {
	if text == "" {
		v.statusLabel.SetText("")
		v.statusLabel.Hide()
	} else {
		v.statusLabel.SetText(text)
		v.statusLabel.Show()
	}
}

// build 组装输入表单与结果列表。
func (v *Bid) build() fyne.CanvasObject {
	v.avgEntry = newCountEntry("例如 5000")
	v.minEntry = newCountEntry("1")
	v.maxEntry = newCountEntry("10")

	countRow := container.NewHBox(v.minEntry, widget.NewLabel("～"), v.maxEntry)

	form := widget.NewForm(
		widget.NewFormItem("单格均价", v.avgEntry),
		widget.NewFormItem("数量区间", countRow),
	)

	inferButton := widget.NewButton("推测", v.infer)
	inferButton.Importance = widget.HighImportance

	v.statusLabel = widget.NewLabel("")
	v.statusLabel.Wrapping = fyne.TextWrapWord
	v.statusLabel.Hide()

	v.headerLabel = widget.NewLabelWithStyle("", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	v.resultList = widget.NewList(
		func() int { return len(v.results) },
		func() fyne.CanvasObject { return newCompositionRow() },
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			obj.(*compositionRow).set(v.results[id])
		},
	)

	v.updateHeader()

	top := container.NewVBox(form, inferButton, v.statusLabel, v.headerLabel, widget.NewSeparator())

	return container.NewBorder(top, nil, nil, nil, v.resultList)
}

// updateHeader 刷新顶部摘要：可用物品数与本次推测的可能数。
func (v *Bid) updateHeader() {
	parts := []string{fmt.Sprintf("单格（1x1）物品 %d 件", v.cellItemCount)}

	switch {
	case !v.inferred:
		parts = append(parts, "填写后点「推测」")
	case len(v.results) == 0:
		parts = append(parts, "没有符合条件的组合")
	default:
		parts = append(parts, fmt.Sprintf("共 %d 种可能", len(v.results)))
	}

	if v.truncated {
		parts = append(parts, "组合过多，只列出前一部分")
	}

	v.headerLabel.SetText(strings.Join(parts, " · "))
}

// infer 校验输入，并把推测请求交给装配层。
func (v *Bid) infer() {
	avg, ok := parseCount(v.avgEntry.Text)

	if !ok {
		v.SetStatus("请输入单格均价（非负整数）")
		return
	}

	minCount, ok := parseCount(v.minEntry.Text)

	if !ok || minCount < 1 {
		v.SetStatus("数量下限要填正整数")
		return
	}

	maxCount, ok := parseCount(v.maxEntry.Text)

	if !ok || maxCount < minCount {
		v.SetStatus("数量上限不能小于下限")
		return
	}

	v.inferred = true
	v.SetStatus("")
	v.updateHeader()

	if v.OnInfer != nil {
		v.OnInfer(avg, minCount, maxCount)
	}
}

// newCountEntry 生成一个只接受非负整数的输入框。
func newCountEntry(placeholder string) *widget.Entry {
	entry := widget.NewEntry()
	entry.SetPlaceHolder(placeholder)
	entry.Validator = numericValidator
	entry.Wrapping = fyne.TextWrapOff
	entry.Scroll = fyne.ScrollNone

	return entry
}

// parseCount 解析用户输入的非负整数。
func parseCount(text string) (int, bool) {
	value, err := strconv.Atoi(strings.TrimSpace(text))

	if err != nil || value < 0 {
		return 0, false
	}

	return value, true
}

// compositionRow 结果列表里的一行：件数、总价、均价，以及具体组成。
type compositionRow struct {
	widget.BaseWidget

	composition bid.Composition
}

// newCompositionRow 构建一行空结果，内容由 set 填充。
func newCompositionRow() *compositionRow {
	row := &compositionRow{}
	row.ExtendBaseWidget(row)

	return row
}

// set 用一条推测结果填充该行。
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
		baseRenderer: baseRenderer{objects: []fyne.CanvasObject{title, items}},
		row:          r,
		title:        title,
		items:        items,
	}
	renderer.Refresh()

	return renderer
}

type compositionRowRenderer struct {
	baseRenderer

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
	r.items.Text = fitText(describeComposition(r.row.composition), r.row.Size().Width-cardInset*2,
		compositionTextSize, fyne.TextStyle{})

	canvas.Refresh(r.row)
}

func (r *compositionRowRenderer) Layout(size fyne.Size) {
	textWidth := size.Width - cardInset*2

	r.title.Move(fyne.NewPos(cardInset, 5))
	r.title.Resize(fyne.NewSize(textWidth, 16))

	r.items.Move(fyne.NewPos(cardInset, 22))
	r.items.Resize(fyne.NewSize(textWidth, 16))

	r.items.Text = fitText(describeComposition(r.row.composition), textWidth, compositionTextSize, fyne.TextStyle{})
}

func (r *compositionRowRenderer) MinSize() fyne.Size {
	return fyne.NewSize(compositionRowMinWidth, compositionRowHeight)
}

// compositionTitle 生成结果行的摘要：件数、总价与均价。
func compositionTitle(composition bid.Composition) string {
	return fmt.Sprintf("%d 件 · 总价 %s · 均价 %s",
		composition.Count, formatValue(composition.Total), formatAverage(composition.Average()))
}
