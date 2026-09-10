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
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

const (
	// compositionRowHeight、compositionRowMinWidth 结果行的高度与建议最小宽度。
	compositionRowHeight   float32 = 42
	compositionRowMinWidth float32 = 240
	// compositionTextSize 结果行的字号。
	compositionTextSize float32 = 12
	// filterPanelRatio 左侧筛选栏在左右分栏中的初始占比。
	filterPanelRatio = 0.3
)

// Bid 是「单格推测」页：可用标签筛选可能出现的单格物品，输入单格均价与数量
// 区间，列出可能的物品组成。标签只是筛选条件：一个都不勾选就表示不筛选。
type Bid struct {
	root fyne.CanvasObject

	// 输入
	avgEntry *widget.Entry
	minEntry *widget.Entry
	maxEntry *widget.Entry

	// 物品标签
	searchEntry    *widget.Entry
	selectionLabel *widget.Label
	chipFlow       *chipFlow
	items          []bid.Item
	selected       []bool
	chips          []*chip
	visible        []int

	// 结果
	statusLabel *widget.Label
	headerLabel *widget.Label
	resultList  *widget.List
	results     []bid.Composition
	truncated   bool
	inferred    bool
	// usedCount 是上一次推测实际使用的物品数量。
	usedCount int

	// OnInfer 由装配层赋值：用户点「推测」时触发，页面本身不做推测。
	OnInfer func(items []bid.Item, avg, minCount, maxCount int)
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

// SetCellItems 用可参与推测的单格物品重建标签区，默认不筛选（全部可用）。
func (v *Bid) SetCellItems(items []bid.Item) {
	v.items = items
	v.selected = make([]bool, len(items))
	v.chips = make([]*chip, 0, len(items))

	for range items {
		v.chips = append(v.chips, newChip(v))
	}

	v.applyFilter(v.searchEntry.Text)
	v.updateSelectionLabel()
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

// build 组装左侧筛选栏与右侧输入、结果列表。
func (v *Bid) build() fyne.CanvasObject {
	left := v.buildFilter()

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
	right := container.NewBorder(top, nil, nil, nil, v.resultList)

	// 内侧留空隙，让左右两块面板读起来是独立表面。
	left = container.New(layout.NewCustomPaddedLayout(0, 0, 0, 8), left)
	right = container.New(layout.NewCustomPaddedLayout(0, 0, 8, 0), right)

	split := container.NewHSplit(left, right)
	split.Offset = filterPanelRatio

	return split
}

// buildFilter 构建左侧的物品筛选栏。
func (v *Bid) buildFilter() fyne.CanvasObject {
	header := widget.NewLabelWithStyle("参与推测的物品", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	v.selectionLabel = widget.NewLabel("")
	v.selectionLabel.Wrapping = fyne.TextWrapWord

	v.searchEntry = widget.NewEntry()
	v.searchEntry.SetPlaceHolder("搜索物品…")
	v.searchEntry.OnChanged = v.applyFilter

	buttons := container.NewHBox(
		widget.NewButton("全选", func() { v.setVisibleChecked(true) }),
		widget.NewButton("反选", v.invertVisible),
		widget.NewButton("清空", func() { v.setVisibleChecked(false) }),
	)

	v.chipFlow = newChipFlow()
	// 滚动条浮在内容之上，把它的宽度留在内容右侧，标签才不会被压住。
	content := container.New(layout.NewCustomPaddedLayout(0, 0, 0, scrollBarInset()), v.chipFlow)
	scroll := container.NewVScroll(content)

	top := container.NewVBox(header, v.selectionLabel, v.searchEntry, buttons, widget.NewSeparator())

	return container.NewBorder(top, nil, nil, nil, scroll)
}

// applyFilter 按关键字过滤标签区；只影响显示，不改动勾选状态。
func (v *Bid) applyFilter(query string) {
	key := strings.ToLower(strings.TrimSpace(query))
	v.visible = v.visible[:0]
	chips := make([]fyne.CanvasObject, 0, len(v.items))

	for i, item := range v.items {
		if key != "" && !matchesItem(item, key) {
			continue
		}

		v.visible = append(v.visible, i)
		v.chips[i].set(i, item, v.selected[i])
		chips = append(chips, v.chips[i])
	}

	v.chipFlow.setChips(chips)
}

// matchesItem 判断物品是否匹配搜索关键字（名称、品质标识与品质展示名）。
func matchesItem(item bid.Item, key string) bool {
	fields := []string{item.Name, item.Quality, qualityLabel(item.Quality)}

	for _, field := range fields {
		if strings.Contains(strings.ToLower(field), key) {
			return true
		}
	}

	return false
}

// toggle 切换某个标签的勾选状态。
func (v *Bid) toggle(index int) {
	if index < 0 || index >= len(v.selected) {
		return
	}

	v.selected[index] = !v.selected[index]
	v.chips[index].set(index, v.items[index], v.selected[index])

	v.updateSelectionLabel()
}

// setVisibleChecked 把当前可见的标签整体勾选或取消。
func (v *Bid) setVisibleChecked(checked bool) {
	for _, index := range v.visible {
		v.selected[index] = checked
		v.chips[index].set(index, v.items[index], checked)
	}

	v.updateSelectionLabel()
}

// invertVisible 反转当前可见标签的勾选状态。
func (v *Bid) invertVisible() {
	for _, index := range v.visible {
		v.selected[index] = !v.selected[index]
		v.chips[index].set(index, v.items[index], v.selected[index])
	}

	v.updateSelectionLabel()
}

// updateSelectionLabel 刷新勾选数量；一个都没勾选表示不做筛选。
func (v *Bid) updateSelectionLabel() {
	selected := 0

	for _, checked := range v.selected {
		if checked {
			selected++
		}
	}

	if selected == 0 {
		v.selectionLabel.SetText(fmt.Sprintf("未勾选＝不筛选，将使用全部 %d 件", len(v.items)))
		return
	}

	v.selectionLabel.SetText(fmt.Sprintf("已勾选 %d / %d 件", selected, len(v.items)))
}

// selectedItems 返回当前勾选的物品；一个都没勾选时返回 nil，由调用方决定含义。
func (v *Bid) selectedItems() []bid.Item {
	items := make([]bid.Item, 0, len(v.items))

	for i, item := range v.items {
		if v.selected[i] {
			items = append(items, item)
		}
	}

	return items
}

// updateHeader 刷新摘要：可用物品数与本次推测的可能数。
func (v *Bid) updateHeader() {
	parts := []string{fmt.Sprintf("单格（1x1）物品 %d 件", len(v.items))}

	switch {
	case !v.inferred:
		parts = append(parts, "填写后点「推测」")
	case len(v.results) == 0:
		parts = append(parts, "没有符合条件的组合")
	default:
		parts = append(parts, fmt.Sprintf("共 %d 种可能（基于 %d 件）", len(v.results), v.usedCount))
	}

	if v.truncated {
		parts = append(parts, "组合过多，只列出前一部分")
	}

	v.headerLabel.SetText(strings.Join(parts, " · "))
}

// infer 校验输入，并把推测请求交给装配层。标签只是筛选条件，一个都不勾选时
// 用全部物品参与推测。
func (v *Bid) infer() {
	items := v.selectedItems()

	if len(items) == 0 {
		items = v.items
	}

	if len(items) == 0 {
		v.SetStatus("还没有加载到单格物品数据")
		return
	}

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
	v.usedCount = len(items)
	v.SetStatus("")
	v.updateHeader()

	if v.OnInfer != nil {
		v.OnInfer(items, avg, minCount, maxCount)
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
	r.items.Text = fitText(describeComposition(r.row.composition), r.textWidth(),
		compositionTextSize, fyne.TextStyle{})

	canvas.Refresh(r.row)
}

func (r *compositionRowRenderer) Layout(size fyne.Size) {
	textWidth := r.textWidth()

	r.title.Move(fyne.NewPos(cardInset, 5))
	r.title.Resize(fyne.NewSize(textWidth, 16))

	r.items.Move(fyne.NewPos(cardInset, 22))
	r.items.Resize(fyne.NewSize(textWidth, 16))

	r.items.Text = fitText(describeComposition(r.row.composition), textWidth, compositionTextSize, fyne.TextStyle{})
}

func (r *compositionRowRenderer) MinSize() fyne.Size {
	return fyne.NewSize(compositionRowMinWidth, compositionRowHeight)
}

// textWidth 返回结果行可用于文字的宽度：右侧留出滚动条的位置。
func (r *compositionRowRenderer) textWidth() float32 {
	return r.row.Size().Width - cardInset*2 - scrollBarInset()
}

// compositionTitle 生成结果行的摘要：件数、总价与均价。
func compositionTitle(composition bid.Composition) string {
	return fmt.Sprintf("%d 件 · 总价 %s · 均价 %s",
		composition.Count, formatValue(composition.Total), formatAverage(composition.Average()))
}
