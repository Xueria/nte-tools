// Package infer 是「拍品推测」页：候选拍品标签、推测条件输入与组合列表。
// 页面不自己做推测，条件与候选池交给装配层。
package infer

import (
	"fmt"
	"strings"

	"nte-tools/internal/bid"
	"nte-tools/internal/ui/common"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

const (
	// filterPanelRatio 左侧筛选栏在左右分栏中的初始占比，左栏只放标签，取窄一些
	// （与盲盒规划页的左栏同宽）。
	filterPanelRatio = 0.32
	// minItemCount、defaultMaxItemCount 数量区间滑块的起点与默认上限。
	minItemCount        = 1
	defaultMaxItemCount = 10
	// defaultMaxTotal 组合总价上限的默认值：1000 万。
	defaultMaxTotal = 10_000_000
	// availablePanelRatio 左栏上下的初始占比：上面是可选择物品，下面是已选择。
	availablePanelRatio = 0.65
	// priceModeAverage、priceModeTotal 是价格方式单选的选项文字。
	priceModeAverage = "均价"
	priceModeTotal   = "总价"
)

// Page 是「拍品推测」页：先选一份拍品清单作为候选池，勾选已确认在组合里的
// 拍品，再填均价或总价与数量区间；先用一条横向标签列出有组合的件数，选中某个
// 件数后再看它的组合。
type Page struct {
	root fyne.CanvasObject

	// 清单与候选池
	listingSelect *widget.Select
	listings      []bid.Listing

	// 输入
	modeRadio     *widget.RadioGroup
	avgEntry      *widget.Entry
	totalEntry    *widget.Entry
	maxTotalEntry *widget.Entry
	countSlider   *common.RangeSlider
	countLabel    *widget.Label

	// 已确认拍品：上面是候选池标签，下面是已选择标签
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

	// OnInfer 由装配层赋值：用户点「推测」时触发，页面本身不做推测，
	// 触发时把当前候选池与已确认拍品一并交给装配层。
	OnInfer func(candidates, required []bid.Item, query bid.InferQuery)
	// OnSelectCount 由装配层赋值：用户选中某个件数时触发，用于取该件数的组合。
	OnSelectCount func(count int)
}

// NewPage 构建拍品推测页。
func NewPage() *Page {
	page := &Page{selectedTab: -1}
	page.root = page.build()

	return page
}

// Title 返回该页在导航栏里的名称。
func (page *Page) Title() string {
	return "拍品推测"
}

// Icon 返回该页在导航栏里的图标。
func (page *Page) Icon() fyne.Resource {
	return theme.SearchIcon()
}

// Content 返回该页的内容。
func (page *Page) Content() fyne.CanvasObject {
	return page.root
}

// SetListings 用新的拍品清单数据重建清单选择器，并默认选中第一份清单，
// 候选池随之换成该清单的全部拍品。
func (page *Page) SetListings(listings []bid.Listing) {
	page.listings = listings

	names := make([]string, 0, len(listings))

	for _, listing := range listings {
		names = append(names, listing.Name)
	}

	page.listingSelect.SetOptions(names)

	if len(names) == 0 {
		page.listingSelect.ClearSelected()
		page.setPool(nil)

		return
	}

	page.listingSelect.SetSelectedIndex(0)
}

// selectListing 把候选池换成选中的那份清单的拍品：换清单后原来的勾选没有意义，
// 因此整池重建。
func (page *Page) selectListing(name string) {
	for _, listing := range page.listings {
		if listing.Name == name {
			page.setPool(listing.Items)
			return
		}
	}
}

// setPool 用新的候选拍品重建标签区，默认没有已确认拍品。
func (page *Page) setPool(items []bid.Item) {
	page.items = items
	page.confirmed = make([]int, len(items))
	page.chips = make([]*chip, 0, len(items))

	for range items {
		page.chips = append(page.chips, newChip(page))
	}

	page.resetResults()
	page.configureCountSlider()
	page.applyFilter(page.searchEntry.Text)
	page.refreshSelection()
}

// resetResults 清空上一次推测的件数档位与组合，换候选池后必须重新推测。
func (page *Page) resetResults() {
	page.inferred = false
	page.requiredCount = 0
	page.counts = nil
	page.countTabs = nil
	page.selectedTab = -1
	page.countBar.Objects = nil
	page.countBar.Refresh()

	page.results = nil
	page.compositionLabel.SetText("")
	page.resultList.Refresh()
}

// SetCounts 展示新的件数档位，并默认选中最小的那个。
func (page *Page) SetCounts(counts []int) {
	page.counts = counts
	page.selectedTab = -1

	page.countTabs = make([]*countTab, 0, len(counts))
	tabs := make([]fyne.CanvasObject, 0, len(counts))

	for i, count := range counts {
		tab := newCountTab(page)
		tab.set(i, count, false)
		page.countTabs = append(page.countTabs, tab)
		tabs = append(tabs, tab)
	}

	page.countBar.Objects = tabs
	page.countBar.Refresh()

	page.results = nil
	page.compositionLabel.SetText("")
	page.resultList.Refresh()
	page.updateHeader()

	if len(counts) == 0 {
		return
	}

	page.selectCount(0)
}

// SetCompositions 展示当前件数下的组合。
func (page *Page) SetCompositions(result bid.InferResult) {
	page.results = result.Compositions
	page.truncated = result.Truncated

	page.compositionLabel.SetText(page.compositionSummary())
	page.resultList.Refresh()
	// 用 ScrollToOffset 而非 ScrollToTop：后者在渲染器尚未创建时会解引用空的 scroller。
	page.resultList.ScrollToOffset(0)
}

// SetStatus 显示提示或错误；传空字符串即隐藏。
func (page *Page) SetStatus(text string) {
	if text == "" {
		page.statusLabel.SetText("")
		page.statusLabel.Hide()
	} else {
		page.statusLabel.SetText(text)
		page.statusLabel.Show()
	}
}

// build 组装左侧筛选栏与右侧输入、件数标签、组合列表。
func (page *Page) build() fyne.CanvasObject {
	left := page.buildFilter()

	page.listingSelect = widget.NewSelect(nil, page.selectListing)

	page.avgEntry = common.NewNumericEntry("例如 5000")
	page.totalEntry = common.NewNumericEntry("例如 50000")
	page.maxTotalEntry = common.NewNumericEntry("默认 10000000")

	page.modeRadio = widget.NewRadioGroup([]string{priceModeAverage, priceModeTotal}, page.applyPriceMode)
	page.modeRadio.Horizontal = true
	page.modeRadio.Required = true
	page.modeRadio.SetSelected(priceModeAverage)

	form := widget.NewForm(
		widget.NewFormItem("清单", page.listingSelect),
		widget.NewFormItem("价格方式", page.modeRadio),
		widget.NewFormItem("均价", page.avgEntry),
		widget.NewFormItem("总价", page.totalEntry),
		widget.NewFormItem("总价上限", page.maxTotalEntry),
	)

	countTitle := widget.NewLabelWithStyle("数量区间", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	page.countLabel = widget.NewLabel("")

	page.countSlider = common.NewRangeSlider(minItemCount, minItemCount)
	page.countSlider.Step = 1
	page.countSlider.OnChanged = func(_, _ float64) { page.updateCountLabel() }
	page.configureCountSlider()

	inferButton := widget.NewButton("推测", page.infer)
	inferButton.Importance = widget.HighImportance

	page.statusLabel = widget.NewLabel("")
	page.statusLabel.Wrapping = fyne.TextWrapWord
	page.statusLabel.Hide()

	page.headerLabel = widget.NewLabelWithStyle("", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	// 件数用一条横向可滚动的标签条展示，省下竖向空间给组合列表。
	page.countBar = container.NewHBox()
	barContent := container.New(layout.NewCustomPaddedLayout(0, common.ScrollBarInset(), 0, 0), page.countBar)
	barScroll := container.NewHScroll(barContent)

	tabTitle := widget.NewLabelWithStyle("件数", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	tabSection := container.NewBorder(nil, nil, tabTitle, nil, barScroll)

	page.compositionLabel = widget.NewLabelWithStyle("", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	page.compositionLabel.Wrapping = fyne.TextWrapWord

	page.resultList = widget.NewList(
		func() int { return len(page.results) },
		func() fyne.CanvasObject { return newCompositionRow() },
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			obj.(*compositionRow).set(page.results[id])
		},
	)

	lists := container.NewBorder(
		container.NewVBox(tabSection, widget.NewSeparator(), page.compositionLabel),
		nil, nil, nil,
		page.resultList,
	)

	page.updateHeader()

	top := container.NewVBox(
		form,
		countTitle,
		page.countSlider,
		page.countLabel,
		inferButton,
		page.statusLabel,
		page.headerLabel,
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

// applyPriceMode 按价格方式启用对应的输入框：均价模式填均价与总价上限，
// 总价模式只填总价。
func (page *Page) applyPriceMode(mode string) {
	if mode == priceModeTotal {
		page.avgEntry.Disable()
		page.maxTotalEntry.Disable()
		page.totalEntry.Enable()

		return
	}

	page.avgEntry.Enable()
	page.maxTotalEntry.Enable()
	page.totalEntry.Disable()
}

// buildFilter 构建左侧的已确认拍品筛选栏。
func (page *Page) buildFilter() fyne.CanvasObject {
	header := widget.NewLabelWithStyle("已确认在里面的拍品", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	page.searchEntry = widget.NewEntry()
	page.searchEntry.SetPlaceHolder("搜索拍品…")
	page.searchEntry.OnChanged = page.applyFilter

	controls := container.NewBorder(nil, nil, nil,
		widget.NewButton("清空勾选", page.clearSelection), page.searchEntry)

	hint := widget.NewLabel("点一下拍品加一件，右键减一件")
	hint.Importance = widget.LowImportance
	hint.Wrapping = fyne.TextWrapWord

	page.chipFlow = newChipFlow()
	available := container.NewBorder(
		container.NewVBox(header, hint, controls, widget.NewSeparator()),
		nil, nil, nil,
		flowScroll(page.chipFlow),
	)

	selectedTitle := widget.NewLabelWithStyle("已选择的拍品", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	page.selectedFlow = newChipFlow()
	selected := container.NewBorder(
		container.NewVBox(selectedTitle, widget.NewSeparator()),
		nil, nil, nil,
		flowScroll(page.selectedFlow),
	)

	split := container.NewVSplit(available, selected)
	split.Offset = availablePanelRatio

	return split
}

// configureCountSlider 按可用物品数设置数量区间的滑块范围。
func (page *Page) configureCountSlider() {
	maxCount := len(page.items)

	if maxCount < minItemCount {
		maxCount = minItemCount
	}

	upper := maxCount

	if upper > defaultMaxItemCount {
		upper = defaultMaxItemCount
	}

	page.countSlider.SetRange(minItemCount, float64(maxCount))
	page.countSlider.SetValues(minItemCount, float64(upper))
	page.updateCountLabel()

	if maxCount <= minItemCount {
		page.countSlider.Disable()

		return
	}

	page.countSlider.Enable()
}

// updateCountLabel 刷新数量区间的说明。
func (page *Page) updateCountLabel() {
	page.countLabel.SetText(fmt.Sprintf("数量 %d ～ %d 件", int(page.countSlider.Lower), int(page.countSlider.Upper)))
}

// selectCount 选中某个件数档位，并向装配层请求它的组合。
func (page *Page) selectCount(index int) {
	if index < 0 || index >= len(page.counts) {
		return
	}

	page.selectedTab = index

	for i, tab := range page.countTabs {
		tab.setSelected(i == index)
	}

	page.compositionLabel.SetText("")

	if page.OnSelectCount != nil {
		page.OnSelectCount(page.counts[index])
	}
}

// compositionSummary 生成组合区标题：当前件数的组合数与总价范围。
func (page *Page) compositionSummary() string {
	if page.selectedTab < 0 || page.selectedTab >= len(page.counts) {
		return ""
	}

	count := page.counts[page.selectedTab]

	if len(page.results) == 0 {
		return fmt.Sprintf("%d 件：没有组合", count)
	}

	parts := []string{fmt.Sprintf("%d 件 的组合：%d 种", count, len(page.results))}
	first, last := page.results[0].Total, page.results[len(page.results)-1].Total

	if first == last {
		parts = append(parts, fmt.Sprintf("总价 %s", common.FormatValue(first)))
	} else {
		parts = append(parts, fmt.Sprintf("总价 %s ～ %s", common.FormatValue(first), common.FormatValue(last)))
	}

	if page.truncated {
		parts = append(parts, "组合过多，只列出前一部分")
	}

	return strings.Join(parts, " · ")
}

// applyFilter 按关键字过滤标签区；只影响显示，不改动勾选状态。
func (page *Page) applyFilter(query string) {
	key := strings.ToLower(strings.TrimSpace(query))
	chips := make([]fyne.CanvasObject, 0, len(page.items))

	for i, item := range page.items {
		if key != "" && !matchesItem(item, key) {
			continue
		}

		page.chips[i].set(i, item, page.confirmed[i])
		chips = append(chips, page.chips[i])
	}

	page.chipFlow.setChips(chips)
}

// matchesItem 判断物品是否匹配搜索关键字（名称、品质标识与品质展示名）。
func matchesItem(item bid.Item, key string) bool {
	fields := []string{item.Name, item.Quality, common.QualityLabel(item.Quality)}

	for _, field := range fields {
		if strings.Contains(strings.ToLower(field), key) {
			return true
		}
	}

	return false
}

// addItem 给某个物品加一件（同一物品可以加多件）。
func (page *Page) addItem(index int) {
	if index < 0 || index >= len(page.confirmed) {
		return
	}

	page.confirmed[index]++

	page.refreshSelection()
}

// removeItem 给某个物品减一件。
func (page *Page) removeItem(index int) {
	if index < 0 || index >= len(page.confirmed) || page.confirmed[index] == 0 {
		return
	}

	page.confirmed[index]--

	page.refreshSelection()
}

// removeAll 把某个物品从已选择里整个移除。
func (page *Page) removeAll(index int) {
	if index < 0 || index >= len(page.confirmed) || page.confirmed[index] == 0 {
		return
	}

	page.confirmed[index] = 0

	page.refreshSelection()
}

// clearSelection 取消所有已确认物品。
func (page *Page) clearSelection() {
	changed := false

	for i := range page.confirmed {
		if page.confirmed[i] == 0 {
			continue
		}

		page.confirmed[i] = 0
		changed = true
	}

	if !changed {
		return
	}

	page.refreshSelection()
}

// refreshSelection 重建可选物品的计数标签、已选择区与摘要。
func (page *Page) refreshSelection() {
	for i, item := range page.items {
		page.chips[i].set(i, item, page.confirmed[i])
	}

	page.chipFlow.Refresh()

	selected := make([]fyne.CanvasObject, 0, len(page.items))

	for i, item := range page.items {
		if page.confirmed[i] == 0 {
			continue
		}

		chip := newSelectedChip(page)
		chip.set(i, item, page.confirmed[i])
		selected = append(selected, chip)
	}

	page.selectedFlow.setChips(selected)

	page.updateHeader()
}

// requiredItems 返回已确认在组合里的物品，按件数展开（同一物品可能出现多次）。
func (page *Page) requiredItems() []bid.Item {
	items := make([]bid.Item, 0)

	for i, item := range page.items {
		for n := 0; n < page.confirmed[i]; n++ {
			items = append(items, item)
		}
	}

	return items
}

// updateHeader 刷新摘要：候选池与本次推测的件数档位。
func (page *Page) updateHeader() {
	parts := []string{fmt.Sprintf("%s：%d 件拍品", page.listingName(), len(page.items))}

	switch {
	case !page.inferred:
		parts = append(parts, "填写后点「推测」")
	case len(page.counts) == 0:
		parts = append(parts, "没有符合条件的件数")
	case len(page.counts) == 1:
		parts = append(parts, fmt.Sprintf("只有 %d 件有组合", page.counts[0]))
	default:
		parts = append(parts, fmt.Sprintf("%d 个件数档位（%d ～ %d 件）",
			len(page.counts), page.counts[0], page.counts[len(page.counts)-1]))
	}

	if page.requiredCount > 0 && page.inferred {
		parts = append(parts, fmt.Sprintf("已确认 %d 件", page.requiredCount))
	}

	page.headerLabel.SetText(strings.Join(parts, " · "))
}

// listingName 返回当前候选池所属的清单名。
func (page *Page) listingName() string {
	if page.listingSelect.Selected == "" {
		return "未选择清单"
	}

	return page.listingSelect.Selected
}
