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
	// filterRowHeight、filterRowMinWidth 筛选行的高度与建议最小宽度。
	filterRowHeight   float32 = 28
	filterRowMinWidth float32 = 200
	// filterBoxSize、filterDotSize、filterPriceWidth 筛选行里勾选框、品质色块
	// 的尺寸以及价格列的宽度。
	filterBoxSize    float32 = 14
	filterDotSize    float32 = 8
	filterPriceWidth float32 = 74
)

// Bid 是「单格推测」页：左侧勾选可能出现的单格物品，右侧输入单格均价与数量
// 区间，列出可能的物品组成。
type Bid struct {
	root fyne.CanvasObject

	// 左上：参与推测的物品筛选
	searchEntry    *widget.Entry
	selectionLabel *widget.Label
	filterBox      *fyne.Container
	items          []bid.Item
	selected       []bool
	rows           []*filterRow
	visible        []int

	// 右上：输入与结果
	avgEntry    *widget.Entry
	minEntry    *widget.Entry
	maxEntry    *widget.Entry
	statusLabel *widget.Label
	headerLabel *widget.Label
	resultList  *widget.List

	results   []bid.Composition
	truncated bool
	inferred  bool

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

// SetCellItems 用可参与推测的单格物品重建筛选列表，并默认全部勾选。
func (v *Bid) SetCellItems(items []bid.Item) {
	v.items = items
	v.selected = make([]bool, len(items))
	v.rows = make([]*filterRow, 0, len(items))

	for i := range v.selected {
		v.selected[i] = true
		v.rows = append(v.rows, newFilterRow(v))
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
	split.Offset = 0.32

	return split
}

// buildFilter 构建左侧的物品筛选栏。
func (v *Bid) buildFilter() fyne.CanvasObject {
	header := widget.NewLabelWithStyle("参与推测的物品", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	v.selectionLabel = widget.NewLabel("")

	v.searchEntry = widget.NewEntry()
	v.searchEntry.SetPlaceHolder("搜索物品…")
	v.searchEntry.OnChanged = v.applyFilter

	v.filterBox = container.NewVBox()
	// 滚动条浮在内容之上，把它的宽度留在内容右侧，勾选框与价格才不会被压住。
	content := container.New(layout.NewCustomPaddedLayout(0, 0, 0, scrollBarInset()), v.filterBox)
	scroll := container.NewVScroll(content)

	buttons := container.NewHBox(
		widget.NewButton("全选", func() { v.setVisibleChecked(true) }),
		widget.NewButton("反选", v.invertVisible),
		widget.NewButton("清空", func() { v.setVisibleChecked(false) }),
	)

	top := container.NewVBox(header, v.selectionLabel, v.searchEntry, widget.NewSeparator())

	return container.NewBorder(top, buttons, nil, nil, scroll)
}

// applyFilter 按关键字过滤筛选列表；只影响显示，不改动勾选状态。
func (v *Bid) applyFilter(query string) {
	key := strings.ToLower(strings.TrimSpace(query))
	v.visible = v.visible[:0]
	objects := make([]fyne.CanvasObject, 0, len(v.items))

	for i, item := range v.items {
		if key != "" && !matchesItem(item, key) {
			continue
		}

		v.visible = append(v.visible, i)
		v.rows[i].set(i, item, v.selected[i])
		objects = append(objects, v.rows[i])
	}

	v.filterBox.Objects = objects
	v.filterBox.Refresh()
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

// toggle 切换某一行的勾选状态。
func (v *Bid) toggle(index int) {
	if index < 0 || index >= len(v.selected) {
		return
	}

	v.selected[index] = !v.selected[index]
	v.rows[index].set(index, v.items[index], v.selected[index])

	v.updateSelectionLabel()
}

// setVisibleChecked 把当前可见的行整体勾选或取消。
func (v *Bid) setVisibleChecked(checked bool) {
	for _, index := range v.visible {
		v.selected[index] = checked
		v.rows[index].set(index, v.items[index], checked)
	}

	v.updateSelectionLabel()
}

// invertVisible 反转当前可见行的勾选状态。
func (v *Bid) invertVisible() {
	for _, index := range v.visible {
		v.selected[index] = !v.selected[index]
		v.rows[index].set(index, v.items[index], v.selected[index])
	}

	v.updateSelectionLabel()
}

// updateSelectionLabel 刷新已勾选数量。
func (v *Bid) updateSelectionLabel() {
	selected := 0

	for _, checked := range v.selected {
		if checked {
			selected++
		}
	}

	v.selectionLabel.SetText(fmt.Sprintf("已勾选 %d / %d 件", selected, len(v.items)))
}

// selectedItems 返回当前勾选的物品。
func (v *Bid) selectedItems() []bid.Item {
	items := make([]bid.Item, 0, len(v.items))

	for i, item := range v.items {
		if v.selected[i] {
			items = append(items, item)
		}
	}

	return items
}

// updateHeader 刷新顶部摘要：可用物品数与本次推测的可能数。
func (v *Bid) updateHeader() {
	parts := []string{fmt.Sprintf("单格（1x1）物品 %d 件", len(v.items))}

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
	items := v.selectedItems()

	if len(items) == 0 {
		v.SetStatus("请先勾选至少一个可能出现的物品")
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

// filterRow 筛选列表里的一行：勾选框、品质色块、名称与价格，点整行即切换勾选。
type filterRow struct {
	widget.BaseWidget

	page    *Bid
	index   int
	item    bid.Item
	checked bool
}

// newFilterRow 构建一行空筛选行，内容由 set 填充。
func newFilterRow(page *Bid) *filterRow {
	row := &filterRow{page: page}
	row.ExtendBaseWidget(row)

	return row
}

// set 设置该行对应的物品与勾选状态。
func (r *filterRow) set(index int, item bid.Item, checked bool) {
	r.index = index
	r.item = item
	r.checked = checked
	r.Refresh()
}

// Tapped 点击整行即切换勾选。
func (r *filterRow) Tapped(*fyne.PointEvent) {
	if r.page != nil {
		r.page.toggle(r.index)
	}
}

// MinSize 返回该行的建议最小尺寸。
func (r *filterRow) MinSize() fyne.Size {
	r.ExtendBaseWidget(r)

	return fyne.NewSize(filterRowMinWidth, filterRowHeight)
}

// CreateRenderer 创建该行的绘制对象。
func (r *filterRow) CreateRenderer() fyne.WidgetRenderer {
	box := canvas.NewRectangle(color.Transparent)
	box.CornerRadius = 3
	box.StrokeWidth = 1

	tick := canvas.NewText("✓", color.White)
	tick.TextSize = 10

	dot := canvas.NewRectangle(color.Transparent)
	dot.CornerRadius = 4

	name := canvas.NewText("", color.White)
	name.TextSize = compositionTextSize

	price := canvas.NewText("", color.White)
	price.TextSize = compositionTextSize
	price.Alignment = fyne.TextAlignTrailing

	renderer := &filterRowRenderer{
		baseRenderer: baseRenderer{objects: []fyne.CanvasObject{box, tick, dot, name, price}},
		row:          r,
		box:          box,
		tick:         tick,
		dot:          dot,
		name:         name,
		price:        price,
	}
	renderer.Refresh()

	return renderer
}

type filterRowRenderer struct {
	baseRenderer

	row   *filterRow
	box   *canvas.Rectangle
	tick  *canvas.Text
	dot   *canvas.Rectangle
	name  *canvas.Text
	price *canvas.Text
}

func (r *filterRowRenderer) Refresh() {
	th := r.row.Theme()
	variant := fyne.CurrentApp().Settings().ThemeVariant()

	if r.row.checked {
		r.box.FillColor = th.Color(theme.ColorNamePrimary, variant)
		r.box.StrokeColor = r.box.FillColor
		r.tick.Color = th.Color(theme.ColorNameForegroundOnPrimary, variant)
		r.tick.Show()
	} else {
		r.box.FillColor = color.Transparent
		r.box.StrokeColor = th.Color(theme.ColorNameInputBorder, variant)
		r.tick.Hide()
	}

	r.dot.FillColor = qualityColor(r.row.item.Quality)

	r.name.Text = r.row.item.Name
	r.name.Color = th.Color(theme.ColorNameForeground, variant)

	r.price.Text = formatValue(r.row.item.Value)
	r.price.Color = th.Color(theme.ColorNamePrimary, variant)

	canvas.Refresh(r.row)
}

func (r *filterRowRenderer) Layout(size fyne.Size) {
	pad := float32(4)
	lineHeight := float32(16)

	r.box.Move(fyne.NewPos(pad, (size.Height-filterBoxSize)/2))
	r.box.Resize(fyne.NewSize(filterBoxSize, filterBoxSize))

	r.tick.Move(fyne.NewPos(pad+3, (size.Height-filterBoxSize)/2-1))
	r.tick.Resize(fyne.NewSize(filterBoxSize, filterBoxSize))

	dotLeft := pad + filterBoxSize + 6
	r.dot.Move(fyne.NewPos(dotLeft, (size.Height-filterDotSize)/2))
	r.dot.Resize(fyne.NewSize(filterDotSize, filterDotSize))

	priceLeft := size.Width - filterPriceWidth
	r.price.Move(fyne.NewPos(priceLeft, (size.Height-lineHeight)/2))
	r.price.Resize(fyne.NewSize(filterPriceWidth, lineHeight))

	nameLeft := dotLeft + filterDotSize + 6
	nameWidth := priceLeft - nameLeft - pad
	r.name.Move(fyne.NewPos(nameLeft, (size.Height-lineHeight)/2))
	r.name.Resize(fyne.NewSize(nameWidth, lineHeight))
	r.name.Text = fitText(r.row.item.Name, nameWidth, compositionTextSize, fyne.TextStyle{})
}

func (r *filterRowRenderer) MinSize() fyne.Size {
	return fyne.NewSize(filterRowMinWidth, filterRowHeight)
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
	r.items.Text = fitText(describeComposition(r.row.composition), r.row.Size().Width-r.textWidth(),
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
