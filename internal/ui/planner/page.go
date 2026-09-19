// Package planner 是「盲盒规划」页及其列表条目：页面只负责展示与收集输入，
// 方案怎么算留在 pool 域。
package planner

import (
	"fmt"
	"strconv"
	"strings"

	"blind-tools/internal/pool"
	"blind-tools/internal/ui/common"
	"blind-tools/internal/ui/shell"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

const (
	// plannerFilterRatio 盲盒列表在左右分栏中的初始占比，左栏只放列表，取窄一些。
	plannerFilterRatio = 0.32
	// localBranchID 本地数据在列表树里的分支节点标识。
	localBranchID = "local"
)

// Page 是「盲盒规划」页：左侧盲盒列表，右侧资源规划。
type Page struct {
	root fyne.CanvasObject

	localAll       []pool.Pool
	localFiltered  []pool.Pool
	selected       *pool.Pool
	selectedNodeID string

	tree        *widget.Tree
	searchEntry *widget.Entry
	statusLabel *widget.Label
	leftPanel   *fyne.Container

	formCard           *widget.Card
	resourceEntries    []*widget.Entry
	keepResourceSelect *widget.Select
	rangeSlider        *common.RangeSlider
	rangeLabel         *widget.Label
	calculateBtn       *widget.Button

	resultBox    *fyne.Container
	summaryLabel *widget.Label

	plan             []pool.PlanStep
	insufficient     bool
	insufficientDraw int

	// OnRefresh 由装配层赋值：用户点「刷新」时触发，页面本身不关心刷新要做什么。
	OnRefresh func()
}

// NewPage 构建盲盒规划页。
func NewPage() *Page {
	page := &Page{}
	page.root = page.build()

	return page
}

// Tab 返回该页的页签标题、图标与内容。
func (page *Page) Tab() shell.Tab {
	return shell.Tab{Title: "盲盒规划", Icon: theme.HomeIcon(), Content: page.root}
}

// build 构建整页的左右分栏。
func (page *Page) build() fyne.CanvasObject {
	left := page.buildLeft()
	right := page.buildRight()

	// 内侧留出空隙，让左右两块面板读起来是独立表面。
	left = container.New(layout.NewCustomPaddedLayout(0, 0, 0, 8), left)
	right = container.New(layout.NewCustomPaddedLayout(0, 0, 8, 0), right)

	split := container.NewHSplit(left, right)
	split.Offset = plannerFilterRatio

	page.applySelection()

	return split
}

// buildLeft 构建搜索栏、刷新按钮与盲盒列表。
func (page *Page) buildLeft() fyne.CanvasObject {
	header := widget.NewLabelWithStyle("盲盒列表", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	page.searchEntry = widget.NewEntry()
	page.searchEntry.SetPlaceHolder("搜索盲盒名称…")
	page.searchEntry.OnChanged = page.applyFilter

	refreshButton := widget.NewButtonWithIcon("刷新", theme.ViewRefreshIcon(), func() {
		if page.OnRefresh != nil {
			page.OnRefresh()
		}
	})
	refreshButton.Importance = widget.MediumImportance

	searchRow := container.NewBorder(nil, nil, nil, refreshButton, page.searchEntry)

	// 列表按数据来源分组，目前只有随包分发的本地 data 目录。
	page.tree = widget.NewTree(
		func(uid widget.TreeNodeID) []widget.TreeNodeID {
			switch uid {
			case "":
				return []widget.TreeNodeID{localBranchID}
			case localBranchID:
				return page.leafIDs()
			}

			return nil
		},
		func(uid widget.TreeNodeID) bool {
			return uid == "" || uid == localBranchID
		},
		func(branch bool) fyne.CanvasObject {
			if branch {
				return widget.NewLabelWithStyle("", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
			}

			return newPoolItem()
		},
		func(uid widget.TreeNodeID, branch bool, obj fyne.CanvasObject) {
			if branch {
				if uid == localBranchID {
					obj.(*widget.Label).SetText(fmt.Sprintf("本地（%d）", len(page.localFiltered)))
				}

				return
			}

			if blindPool, ok := page.poolFor(uid); ok {
				obj.(*poolItem).set(*blindPool)
			}
		},
	)
	page.tree.HideSeparators = true
	page.tree.OnSelected = func(uid widget.TreeNodeID) {
		if uid == localBranchID {
			page.tree.Unselect(uid)
			return
		}

		page.selectPool(uid)
	}
	page.tree.OpenAllBranches()

	page.statusLabel = widget.NewLabel("")
	page.statusLabel.Wrapping = fyne.TextWrapWord
	page.statusLabel.Hide()

	top := container.NewVBox(header, searchRow)
	page.leftPanel = container.NewBorder(top, page.statusLabel, nil, nil, page.tree)

	return page.leftPanel
}

// SetStatus 在左栏底部显示提示；传空字符串则隐藏并让出空间。
func (page *Page) SetStatus(text string) {
	if text == "" {
		page.statusLabel.SetText("")
		page.statusLabel.Hide()
	} else {
		page.statusLabel.SetText(text)
		page.statusLabel.Show()
	}

	if page.leftPanel != nil {
		page.leftPanel.Refresh()
	}
}

// buildRight 构建资源表单与结果表格。
func (page *Page) buildRight() fyne.CanvasObject {
	page.formCard = widget.NewCard("", "", nil)

	page.keepResourceSelect = widget.NewSelect(nil, nil)

	page.rangeSlider = common.NewRangeSlider(1, 1)
	page.rangeSlider.Step = 1
	page.rangeSlider.OnChanged = func(lower, upper float64) {
		page.rangeLabel.SetText(fmt.Sprintf("第 %d 抽 ～ 第 %d 抽", int(lower), int(upper)))
	}

	page.rangeLabel = widget.NewLabel("")

	page.calculateBtn = widget.NewButton("计算方案", page.calculate)
	page.calculateBtn.Importance = widget.HighImportance

	page.summaryLabel = widget.NewLabel("")
	page.summaryLabel.Wrapping = fyne.TextWrapWord

	page.resultBox = container.NewVBox()

	resultTitle := widget.NewLabelWithStyle("计算方案", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	resultArea := container.NewBorder(resultTitle, page.summaryLabel, nil, nil, page.resultBox)

	right := container.NewBorder(page.formCard, nil, nil, nil, resultArea)
	// 整块面板放进滚动容器，最小高度就不会顶住窗口，窗口可自由缩放。
	return container.NewVScroll(right)
}

// applyFilter 按名称或标识过滤列表，并让选中项跟着过滤结果走。
func (page *Page) applyFilter(text string) {
	query := strings.ToLower(strings.TrimSpace(text))

	page.localFiltered = filterPools(page.localAll, query)

	// 重新在当前过滤结果里确认选中项是否还在。
	keep := (*pool.Pool)(nil)
	keepID := ""

	if page.selected != nil {
		if blindPool, ok := page.poolFor(page.selectedNodeID); ok {
			keep = blindPool
			keepID = page.selectedNodeID
		}
	}

	page.selected = keep

	if keep == nil {
		page.selectedNodeID = ""
	}

	// 同步列表树的选中态：否则被过滤掉又回到同一个标识的条目会点不动，
	// 因为 Tree.Select 会因为标识没变而直接返回。
	page.tree.UnselectAll()

	if keepID != "" {
		page.tree.Select(keepID)
	}

	page.tree.Refresh()

	page.applySelection()
}

// filterPools 返回名称或标识匹配 query 的盲盒池。
func filterPools(pools []pool.Pool, query string) []pool.Pool {
	if query == "" {
		return append([]pool.Pool(nil), pools...)
	}

	filtered := make([]pool.Pool, 0, len(pools))

	for _, blindPool := range pools {
		if strings.Contains(strings.ToLower(blindPool.Manifest.Name), query) ||
			strings.Contains(strings.ToLower(blindPool.Manifest.ID), query) {
			filtered = append(filtered, blindPool)
		}
	}

	return filtered
}

// leafIDs 返回当前过滤结果在列表树里的叶子节点标识。
func (page *Page) leafIDs() []widget.TreeNodeID {
	ids := make([]widget.TreeNodeID, len(page.localFiltered))

	for i, blindPool := range page.localFiltered {
		ids[i] = widget.TreeNodeID(blindPool.Manifest.ID)
	}

	return ids
}

// poolFor 把列表树的叶子节点标识解析成盲盒池。
func (page *Page) poolFor(nodeID string) (*pool.Pool, bool) {
	for i := range page.localFiltered {
		if page.localFiltered[i].Manifest.ID == nodeID {
			return &page.localFiltered[i], true
		}
	}

	return nil, false
}

// SetPools 用新的盲盒池替换列表内容并重新过滤。
func (page *Page) SetPools(pools []pool.Pool) {
	page.localAll = pools
	page.applyFilter(page.searchEntry.Text)
}

// selectPool 记住选中的盲盒池并重建右栏。
func (page *Page) selectPool(nodeID widget.TreeNodeID) {
	blindPool, ok := page.poolFor(nodeID)

	if !ok {
		return
	}

	page.selected = blindPool
	page.selectedNodeID = nodeID
	page.applySelection()
}

// applySelection 按当前选中项重建右栏。
func (page *Page) applySelection() {
	page.plan = nil
	page.insufficient = false
	page.insufficientDraw = 0
	page.resourceEntries = nil
	page.rebuildResultTable()

	if page.selected == nil {
		page.formCard.SetTitle("资源规划")
		page.formCard.SetSubTitle("请选择一个盲盒")
		page.formCard.SetContent(container.NewCenter(widget.NewLabel("从左侧列表选择一个盲盒开始规划")))
		page.keepResourceSelect.Disable()
		page.rangeSlider.Disable()
		page.calculateBtn.Disable()
		page.summaryLabel.SetText("")

		return
	}

	selected := page.selected
	page.formCard.SetTitle(selected.Manifest.Name)
	page.formCard.SetSubTitle(fmt.Sprintf("共 %d 抽 · %d 种资源", selected.Manifest.Draws, len(selected.Resources)))

	// 每种资源一个数量输入框。
	form := widget.NewForm()

	for _, resource := range selected.Resources {
		entry := common.NewNumericEntry("0")
		form.Append(resource.Name, entry)
		page.resourceEntries = append(page.resourceEntries, entry)
	}

	resourceTitle := widget.NewLabelWithStyle("资源数量", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	rangeTitle := widget.NewLabelWithStyle("抽数范围", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	keepTitle := widget.NewLabelWithStyle("优先保留的资源", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	// 抽数范围滑块。
	draws := selected.Manifest.Draws

	if draws < 1 {
		draws = 1
	}

	page.rangeSlider.SetRange(1, float64(draws))
	page.rangeSlider.SetValues(1, float64(draws))
	page.rangeLabel.SetText(fmt.Sprintf("第 1 抽 ～ 第 %d 抽", draws))

	if draws <= 1 {
		page.rangeSlider.Disable()
	} else {
		page.rangeSlider.Enable()
	}

	// 优先保留的资源选择器：第一项「无」表示不指定，只求完成的抽数最多。
	names := make([]string, 0, len(selected.Resources)+1)
	names = append(names, "无")

	for _, resource := range selected.Resources {
		names = append(names, resource.Name)
	}

	page.keepResourceSelect.SetOptions(names)
	page.keepResourceSelect.SetSelectedIndex(0)
	page.keepResourceSelect.Enable()

	page.calculateBtn.Enable()
	page.summaryLabel.SetText("填写资源数量后点击「计算方案」")

	content := container.NewVBox(
		resourceTitle,
		form,
		rangeTitle,
		page.rangeSlider,
		page.rangeLabel,
		keepTitle,
		page.keepResourceSelect,
		page.calculateBtn,
	)
	page.formCard.SetContent(content)
}

// calculate 解析输入、计算方案并刷新结果表格。
func (page *Page) calculate() {
	if page.selected == nil {
		return
	}

	balances := make(map[string]int, len(page.selected.Resources))

	for i, resource := range page.selected.Resources {
		text := strings.TrimSpace(page.resourceEntries[i].Text)

		if text == "" {
			text = "0"
		}

		amount, ok := common.ParseNonNegativeInt(text)

		if !ok {
			page.summaryLabel.SetText(fmt.Sprintf("请输入有效的「%s」数量", resource.Name))
			return
		}

		balances[resource.ID] = amount
	}

	start := int(page.rangeSlider.Lower)
	end := int(page.rangeSlider.Upper)

	if start < 1 {
		start = 1
	}

	if end < start {
		end = start
	}

	result := page.selected.Plan(start, end, balances, page.preferredResourceID())

	page.plan = result.Steps
	page.insufficient = result.Insufficient
	page.insufficientDraw = result.FailAtDraw
	page.rebuildResultTable()
	page.updateSummary(result)
}

// preferredResourceID 返回「优先保留」选中的资源标识。
func (page *Page) preferredResourceID() string {
	index := page.keepResourceSelect.SelectedIndex()

	if index <= 0 { // 「无」（或未选中）表示不指定
		return ""
	}

	resourceIndex := index - 1

	if resourceIndex < len(page.selected.Resources) {
		return page.selected.Resources[resourceIndex].ID
	}

	return ""
}

// updateSummary 展示最终余额，以及资源不足时的提示。
func (page *Page) updateSummary(result pool.PlanResult) {
	ids := page.selected.SortedResourceIDs()
	parts := make([]string, 0, len(ids))

	for _, id := range ids {
		parts = append(parts, fmt.Sprintf("%s %d", page.selected.ResourceName(id), result.Final[id]))
	}

	summary := "剩余资源：" + strings.Join(parts, "，")

	if result.Insufficient {
		summary = fmt.Sprintf("第 %d 抽资源不足，无法继续｜", result.FailAtDraw) + summary
	}

	page.summaryLabel.SetText(summary)
}

// rebuildResultTable 把结果重画成一张按内容撑满高度的表格。不用 widget.Table：
// 它自带滚动、窗口小的时候只显示一行，而这里的表格交给外层滚动容器，能一次
// 露出全部行。
func (page *Page) rebuildResultTable() {
	rows := []fyne.CanvasObject{
		container.NewGridWithColumns(4,
			resultCell("抽数", true),
			resultCell("使用资源", true),
			resultCell("花费", true),
			resultCell("剩余", true),
		),
		widget.NewSeparator(),
	}

	for _, step := range page.plan {
		rows = append(rows, container.NewGridWithColumns(4,
			resultCell(strconv.Itoa(step.Draw), false),
			resultCell(step.ResourceName, false),
			resultCell(strconv.Itoa(step.Cost), false),
			resultCell(strconv.Itoa(step.Remaining), false),
		))
	}

	if page.insufficient {
		rows = append(rows, container.NewGridWithColumns(4,
			resultCell(strconv.Itoa(page.insufficientDraw), false),
			resultCell("资源不足", false),
			resultCell("—", false),
			resultCell("—", false),
		))
	}

	page.resultBox.Objects = rows
	page.resultBox.Refresh()
}

// resultCell 生成结果表格里的一个居中单元格。
func resultCell(text string, bold bool) fyne.CanvasObject {
	return widget.NewLabelWithStyle(text, fyne.TextAlignCenter, fyne.TextStyle{Bold: bold})
}
