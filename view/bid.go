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
	// filterPanelRatio 左侧筛选栏在左右分栏中的初始占比，默认对半分。
	filterPanelRatio = 0.5
	// minItemCount、defaultMaxItemCount 数量区间滑块的起点与默认上限。
	minItemCount        = 1
	defaultMaxItemCount = 10
	// defaultMaxTotal 组合总价上限的默认值：1000 万。
	defaultMaxTotal = 10_000_000
)

// Bid 是「单格推测」页：勾选已确认在组合里的单格物品，输入单格均价、数量区间与
// 总价上限，由算法从全部单格物品里补足其余位置，列出可能的物品组成。
type Bid struct {
	root fyne.CanvasObject

	// 输入
	avgEntry      *widget.Entry
	maxTotalEntry *widget.Entry
	countSlider   *RangeSlider
	countLabel    *widget.Label

	// 已确认物品
	searchEntry    *widget.Entry
	selectionLabel *widget.Label
	chipFlow       *chipFlow
	items          []bid.Item
	selected       []bool
	chips          []*chip

	// 结果
	statusLabel *widget.Label
	headerLabel *widget.Label
	resultList  *widget.List
	results     []bid.Composition
	truncated   bool
	inferred    bool
	// requiredCount 是上一次推测里已确认的物品数量。
	requiredCount int

	// OnInfer 由装配层赋值：用户点「推测」时触发，页面本身不做推测。
	OnInfer func(required []bid.Item, query bid.InferQuery)
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

// SetCellItems 用可参与推测的单格物品重建标签区，默认没有已确认物品。
func (v *Bid) SetCellItems(items []bid.Item) {
	v.items = items
	v.selected = make([]bool, len(items))
	v.chips = make([]*chip, 0, len(items))

	for range items {
		v.chips = append(v.chips, newChip(v))
	}

	v.configureCountSlider()
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
	v.maxTotalEntry = newCountEntry("默认 10000000")

	form := widget.NewForm(
		widget.NewFormItem("单格均价", v.avgEntry),
		widget.NewFormItem("总价上限", v.maxTotalEntry),
	)

	countTitle := widget.NewLabelWithStyle("数量区间", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	v.countLabel = widget.NewLabel("")

	v.countSlider = NewRangeSlider(minItemCount, minItemCount)
	v.countSlider.Step = 1
	v.countSlider.OnChanged = func(_, _ float64) { v.updateCountLabel() }
	v.configureCountSlider()

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

	top := container.NewVBox(
		form,
		countTitle,
		v.countSlider,
		v.countLabel,
		inferButton,
		v.statusLabel,
		v.headerLabel,
		widget.NewSeparator(),
	)
	right := container.NewBorder(top, nil, nil, nil, v.resultList)

	// 内侧留空隙，让左右两块面板读起来是独立表面。
	left = container.New(layout.NewCustomPaddedLayout(0, 0, 0, 8), left)
	right = container.New(layout.NewCustomPaddedLayout(0, 0, 8, 0), right)

	split := container.NewHSplit(left, right)
	split.Offset = filterPanelRatio

	return split
}

// buildFilter 构建左侧的已确认物品筛选栏。
func (v *Bid) buildFilter() fyne.CanvasObject {
	header := widget.NewLabelWithStyle("已确认在里面的物品", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	v.selectionLabel = widget.NewLabel("")
	v.selectionLabel.Wrapping = fyne.TextWrapWord

	v.searchEntry = widget.NewEntry()
	v.searchEntry.SetPlaceHolder("搜索物品…")
	v.searchEntry.OnChanged = v.applyFilter

	controls := container.NewBorder(nil, nil, nil,
		widget.NewButton("清空勾选", v.clearSelection), v.searchEntry)

	v.chipFlow = newChipFlow()
	// 滚动条浮在内容之上，把它的宽度留在内容右侧，标签才不会被压住。
	content := container.New(layout.NewCustomPaddedLayout(0, 0, 0, scrollBarInset()), v.chipFlow)
	scroll := container.NewVScroll(content)

	top := container.NewVBox(header, v.selectionLabel, controls, widget.NewSeparator())

	return container.NewBorder(top, nil, nil, nil, scroll)
}

// configureCountSlider 按可用物品数设置数量区间的滑块范围。
func (v *Bid) configureCountSlider() {
	maxCount := len(v.items)

	if maxCount < minItemCount {
		maxCount = minItemCount
	}

	upper := maxCount
	if upper > defaultMaxItemCount {
		upper = defaultMaxItemCount
	}

	v.countSlider.SetRange(minItemCount, float64(maxCount))
	v.countSlider.SetValues(minItemCount, float64(upper))
	v.updateCountLabel()

	if maxCount <= minItemCount {
		v.countSlider.Disable()
		return
	}

	v.countSlider.Enable()
}

// updateCountLabel 刷新数量区间的说明。
func (v *Bid) updateCountLabel() {
	v.countLabel.SetText(fmt.Sprintf("数量 %d ～ %d 件", int(v.countSlider.Lower), int(v.countSlider.Upper)))
}

// applyFilter 按关键字过滤标签区；只影响显示，不改动勾选状态。
func (v *Bid) applyFilter(query string) {
	key := strings.ToLower(strings.TrimSpace(query))
	chips := make([]fyne.CanvasObject, 0, len(v.items))

	for i, item := range v.items {
		if key != "" && !matchesItem(item, key) {
			continue
		}

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

// clearSelection 取消所有已确认物品。
func (v *Bid) clearSelection() {
	for i := range v.selected {
		if !v.selected[i] {
			continue
		}

		v.selected[i] = false
		v.chips[i].set(i, v.items[i], false)
	}

	v.updateSelectionLabel()
}

// updateSelectionLabel 刷新已确认物品的说明。
func (v *Bid) updateSelectionLabel() {
	confirmed := 0

	for _, checked := range v.selected {
		if checked {
			confirmed++
		}
	}

	if confirmed == 0 {
		v.selectionLabel.SetText(fmt.Sprintf("未勾选＝不限定，将从全部 %d 件里推测", len(v.items)))
		return
	}

	v.selectionLabel.SetText(fmt.Sprintf("已确认 %d 件在里面，其余位置由算法补足", confirmed))
}

// requiredItems 返回已确认在组合里的物品。
func (v *Bid) requiredItems() []bid.Item {
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
		parts = append(parts, fmt.Sprintf("共 %d 种可能", len(v.results)))

		if v.requiredCount > 0 {
			parts = append(parts, fmt.Sprintf("已确认 %d 件", v.requiredCount))
		}
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

	maxTotal := defaultMaxTotal

	if strings.TrimSpace(v.maxTotalEntry.Text) != "" {
		maxTotal, ok = parseCount(v.maxTotalEntry.Text)

		if !ok {
			v.SetStatus("总价上限要填非负整数")
			return
		}
	}

	minCount := int(v.countSlider.Lower)
	maxCount := int(v.countSlider.Upper)

	if minCount < minItemCount {
		minCount = minItemCount
	}

	if maxCount < minCount {
		maxCount = minCount
	}

	required := v.requiredItems()

	if len(required) > maxCount {
		v.SetStatus(fmt.Sprintf("已确认 %d 件，超过数量上限 %d，请调整数量区间或取消勾选",
			len(required), maxCount))
		return
	}

	v.inferred = true
	v.requiredCount = len(required)
	v.SetStatus("")
	v.updateHeader()

	if v.OnInfer != nil {
		v.OnInfer(required, bid.InferQuery{
			Avg:      avg,
			MinCount: minCount,
			MaxCount: maxCount,
			MaxTotal: maxTotal,
		})
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

// compositionTitle 生成结果行的摘要：总价、件数与均价。总价放在最前面，
// 列表是按总价从低到高排的，这样一眼就能核对顺序。
func compositionTitle(composition bid.Composition) string {
	return fmt.Sprintf("总价 %s · %d 件 · 均价 %s",
		formatValue(composition.Total), composition.Count, formatAverage(composition.Average()))
}
