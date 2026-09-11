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
	// compositionRowHeight、compositionRowMinWidth 组合行的高度与建议最小宽度。
	compositionRowHeight   float32 = 42
	compositionRowMinWidth float32 = 240
	// compositionTextSize 列表行（含件数标签）的字号。
	compositionTextSize float32 = 12
	// filterPanelRatio 左侧筛选栏在左右分栏中的初始占比，左栏只放标签，取窄一些。
	filterPanelRatio = 0.4
	// minItemCount、defaultMaxItemCount 数量区间滑块的起点与默认上限。
	minItemCount        = 1
	defaultMaxItemCount = 10
	// defaultMaxTotal 组合总价上限的默认值：1000 万。
	defaultMaxTotal = 10_000_000
	// priorityQuality 关注的品质：含它的组合在结果里标星。
	priorityQuality = "red"
	// availablePanelRatio 左栏上下的初始占比：上面是可选择物品，下面是已选择。
	availablePanelRatio = 0.65
)

// Bid 是「单格推测」页：勾选已确认在组合里的单格物品，输入单格均价、数量区间与
// 总价上限；先用一条横向标签列出有组合的件数，选中某个件数后再看它的组合。
type Bid struct {
	root fyne.CanvasObject

	// 输入
	avgEntry      *widget.Entry
	maxTotalEntry *widget.Entry
	countSlider   *RangeSlider
	countLabel    *widget.Label

	// 已确认物品：上面是可选物品标签，下面是已选择标签
	searchEntry  *widget.Entry
	chipFlow     *chipFlow
	selectedFlow *chipFlow
	items        []bid.Item
	confirmed    []int
	chips        []*chip

	// 件数档位
	countBar    *fyne.Container
	countTabs   []*countTab
	counts      []int
	selectedTab int

	// 组合
	compositionLabel *widget.Label
	resultList       *widget.List
	results          []bid.Composition
	truncated        bool
	inferred         bool
	// requiredCount 是上一次推测里已确认的物品数量。
	requiredCount int

	// 提示与摘要
	statusLabel *widget.Label
	headerLabel *widget.Label

	// OnInfer 由装配层赋值：用户点「推测」时触发，页面本身不做推测。
	OnInfer func(required []bid.Item, query bid.InferQuery)
	// OnSelectCount 由装配层赋值：用户选中某个件数时触发，用于取该件数的组合。
	OnSelectCount func(count int)
}

// NewBid 构建单格推测页。
func NewBid() *Bid {
	v := &Bid{selectedTab: -1}
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
	v.confirmed = make([]int, len(items))
	v.chips = make([]*chip, 0, len(items))

	for range items {
		v.chips = append(v.chips, newChip(v))
	}

	v.configureCountSlider()
	v.applyFilter(v.searchEntry.Text)
	v.refreshSelection()
}

// SetCounts 展示新的件数档位，并默认选中最小的那个。
func (v *Bid) SetCounts(counts []int) {
	v.counts = counts
	v.selectedTab = -1

	v.countTabs = make([]*countTab, 0, len(counts))
	tabs := make([]fyne.CanvasObject, 0, len(counts))

	for i, count := range counts {
		tab := newCountTab(v)
		tab.set(i, count, false)
		v.countTabs = append(v.countTabs, tab)
		tabs = append(tabs, tab)
	}

	v.countBar.Objects = tabs
	v.countBar.Refresh()

	v.results = nil
	v.compositionLabel.SetText("")
	v.resultList.Refresh()
	v.updateHeader()

	if len(counts) == 0 {
		return
	}

	v.selectCount(0)
}

// SetCompositions 展示当前件数下的组合。
func (v *Bid) SetCompositions(result bid.InferResult) {
	v.results = result.Compositions
	v.truncated = result.Truncated

	v.compositionLabel.SetText(v.compositionSummary())
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

// build 组装左侧筛选栏与右侧输入、件数标签、组合列表。
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

	// 件数用一条横向可滚动的标签条展示，省下竖向空间给组合列表。
	v.countBar = container.NewHBox()
	barContent := container.New(layout.NewCustomPaddedLayout(0, scrollBarInset(), 0, 0), v.countBar)
	barScroll := container.NewHScroll(barContent)

	tabTitle := widget.NewLabelWithStyle("件数", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	tabSection := container.NewBorder(nil, nil, tabTitle, nil, barScroll)

	v.compositionLabel = widget.NewLabelWithStyle("", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	v.compositionLabel.Wrapping = fyne.TextWrapWord

	v.resultList = widget.NewList(
		func() int { return len(v.results) },
		func() fyne.CanvasObject { return newCompositionRow() },
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			obj.(*compositionRow).set(v.results[id])
		},
	)

	lists := container.NewBorder(
		container.NewVBox(tabSection, widget.NewSeparator(), v.compositionLabel),
		nil, nil, nil,
		v.resultList,
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
	right := container.NewBorder(top, nil, nil, nil, lists)

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

	v.searchEntry = widget.NewEntry()
	v.searchEntry.SetPlaceHolder("搜索物品…")
	v.searchEntry.OnChanged = v.applyFilter

	controls := container.NewBorder(nil, nil, nil,
		widget.NewButton("清空勾选", v.clearSelection), v.searchEntry)

	hint := widget.NewLabel("点一下物品加一件，右键减一件")
	hint.Importance = widget.LowImportance
	hint.Wrapping = fyne.TextWrapWord

	v.chipFlow = newChipFlow()
	available := container.NewBorder(
		container.NewVBox(header, hint, controls, widget.NewSeparator()),
		nil, nil, nil,
		flowScroll(v.chipFlow),
	)

	selectedTitle := widget.NewLabelWithStyle("已选择的物品", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	v.selectedFlow = newChipFlow()
	selected := container.NewBorder(
		container.NewVBox(selectedTitle, widget.NewSeparator()),
		nil, nil, nil,
		flowScroll(v.selectedFlow),
	)

	split := container.NewVSplit(available, selected)
	split.Offset = availablePanelRatio

	return split
}

// flowScroll 把标签流放进滚动容器里，右侧留出滚动条的宽度。
func flowScroll(flow *chipFlow) fyne.CanvasObject {
	content := container.New(layout.NewCustomPaddedLayout(0, 0, 0, scrollBarInset()), flow)

	return container.NewVScroll(content)
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

// selectCount 选中某个件数档位，并向装配层请求它的组合。
func (v *Bid) selectCount(index int) {
	if index < 0 || index >= len(v.counts) {
		return
	}

	v.selectedTab = index

	for i, tab := range v.countTabs {
		tab.setSelected(i == index)
	}

	v.compositionLabel.SetText("")

	if v.OnSelectCount != nil {
		v.OnSelectCount(v.counts[index])
	}
}

// compositionSummary 生成组合区标题：当前件数的组合数与总价范围。
func (v *Bid) compositionSummary() string {
	if v.selectedTab < 0 || v.selectedTab >= len(v.counts) {
		return ""
	}

	count := v.counts[v.selectedTab]

	if len(v.results) == 0 {
		return fmt.Sprintf("%d 件：没有组合", count)
	}

	parts := []string{fmt.Sprintf("%d 件 的组合：%d 种", count, len(v.results))}
	first, last := v.results[0].Total, v.results[len(v.results)-1].Total

	if first == last {
		parts = append(parts, fmt.Sprintf("总价 %s", formatValue(first)))
	} else {
		parts = append(parts, fmt.Sprintf("总价 %s ～ %s", formatValue(first), formatValue(last)))
	}

	if v.truncated {
		parts = append(parts, "组合过多，只列出前一部分")
	}

	return strings.Join(parts, " · ")
}

// applyFilter 按关键字过滤标签区；只影响显示，不改动勾选状态。
func (v *Bid) applyFilter(query string) {
	key := strings.ToLower(strings.TrimSpace(query))
	chips := make([]fyne.CanvasObject, 0, len(v.items))

	for i, item := range v.items {
		if key != "" && !matchesItem(item, key) {
			continue
		}

		v.chips[i].set(i, item, v.confirmed[i])
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

// addItem 给某个物品加一件（同一物品可以加多件）。
func (v *Bid) addItem(index int) {
	if index < 0 || index >= len(v.confirmed) {
		return
	}

	v.confirmed[index]++

	v.refreshSelection()
}

// removeItem 给某个物品减一件。
func (v *Bid) removeItem(index int) {
	if index < 0 || index >= len(v.confirmed) || v.confirmed[index] == 0 {
		return
	}

	v.confirmed[index]--

	v.refreshSelection()
}

// removeAll 把某个物品从已选择里整个移除。
func (v *Bid) removeAll(index int) {
	if index < 0 || index >= len(v.confirmed) || v.confirmed[index] == 0 {
		return
	}

	v.confirmed[index] = 0

	v.refreshSelection()
}

// clearSelection 取消所有已确认物品。
func (v *Bid) clearSelection() {
	changed := false

	for i := range v.confirmed {
		if v.confirmed[i] == 0 {
			continue
		}

		v.confirmed[i] = 0
		changed = true
	}

	if !changed {
		return
	}

	v.refreshSelection()
}

// refreshSelection 重建可选物品的计数标签、已选择区与摘要。
func (v *Bid) refreshSelection() {
	for i, item := range v.items {
		v.chips[i].set(i, item, v.confirmed[i])
	}

	v.chipFlow.Refresh()

	selected := make([]fyne.CanvasObject, 0, len(v.items))

	for i, item := range v.items {
		if v.confirmed[i] == 0 {
			continue
		}

		chip := newSelectedChip(v)
		chip.set(i, item, v.confirmed[i])
		selected = append(selected, chip)
	}

	v.selectedFlow.setChips(selected)

	v.updateHeader()
}

// requiredItems 返回已确认在组合里的物品，按件数展开（同一物品可能出现多次）。
func (v *Bid) requiredItems() []bid.Item {
	items := make([]bid.Item, 0)

	for i, item := range v.items {
		for n := 0; n < v.confirmed[i]; n++ {
			items = append(items, item)
		}
	}

	return items
}

// updateHeader 刷新摘要：可用物品数与本次推测的件数档位。
func (v *Bid) updateHeader() {
	parts := []string{fmt.Sprintf("单格（1x1）物品 %d 件", len(v.items))}

	switch {
	case !v.inferred:
		parts = append(parts, "填写后点「推测」")
	case len(v.counts) == 0:
		parts = append(parts, "没有符合条件的件数")
	case len(v.counts) == 1:
		parts = append(parts, fmt.Sprintf("只有 %d 件有组合", v.counts[0]))
	default:
		parts = append(parts, fmt.Sprintf("%d 个件数档位（%d ～ %d 件）",
			len(v.counts), v.counts[0], v.counts[len(v.counts)-1]))
	}

	if v.requiredCount > 0 && v.inferred {
		parts = append(parts, fmt.Sprintf("已确认 %d 件", v.requiredCount))
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

// textWidth 返回组合行可用于文字的宽度：右侧留出滚动条的位置。
func (r *compositionRowRenderer) textWidth() float32 {
	return r.row.Size().Width - cardInset*2 - scrollBarInset()
}

// compositionTitle 生成组合行的摘要：总价、件数与均价；含关注品质的组合加星标。
func compositionTitle(composition bid.Composition) string {
	title := fmt.Sprintf("总价 %s · %d 件 · 均价 %s",
		formatValue(composition.Total), composition.Count, formatAverage(composition.Average()))

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
