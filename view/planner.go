package view

import (
	"blind-tools/model/pool"
	"errors"
	"fmt"
	"image/color"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// NewPlanner 构建盲盒规划页：左侧盲盒列表，右侧资源规划。
func NewPlanner() *Planner {
	v := &Planner{}
	v.root = v.build()
	return v
}

// NewTab 返回该页的页签标题、图标与内容。
func (v *Planner) NewTab() Tab {
	return Tab{Title: "盲盒规划", Icon: theme.HomeIcon(), Content: v.root}
}

// Planner holds the mutable UI state of the blind pool planner page.
type Planner struct {
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
	rangeSlider        *RangeSlider
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

// build constructs the full split layout once.
func (v *Planner) build() fyne.CanvasObject {
	left := v.buildLeft()
	right := v.buildRight()

	// Add a gutter on the inner sides so the two panels read as separate
	// surfaces instead of one merged panel.
	left = container.New(layout.NewCustomPaddedLayout(0, 0, 0, 8), left)
	right = container.New(layout.NewCustomPaddedLayout(0, 0, 8, 0), right)

	split := container.NewHSplit(left, right)
	split.Offset = 0.32

	v.applySelection()

	return split
}

// buildLeft creates the search bar, refresh button and blind pool list.
func (v *Planner) buildLeft() fyne.CanvasObject {
	header := widget.NewLabelWithStyle("盲盒列表", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	v.searchEntry = widget.NewEntry()
	v.searchEntry.SetPlaceHolder("搜索盲盒名称…")
	v.searchEntry.OnChanged = v.applyFilter

	refreshBtn := widget.NewButtonWithIcon("刷新", theme.ViewRefreshIcon(), func() {
		if v.OnRefresh != nil {
			v.OnRefresh()
		}
	})
	refreshBtn.Importance = widget.MediumImportance

	searchRow := container.NewBorder(nil, nil, nil, refreshBtn, v.searchEntry)

	// The list is grouped by data source; local is the on-disk data folder.
	v.tree = widget.NewTree(
		func(uid widget.TreeNodeID) []widget.TreeNodeID {
			switch uid {
			case "":
				return []widget.TreeNodeID{"local"}
			case "local":
				return v.leafIDs()
			}
			return nil
		},
		func(uid widget.TreeNodeID) bool {
			return uid == "" || uid == "local"
		},
		func(branch bool) fyne.CanvasObject {
			if branch {
				return widget.NewLabelWithStyle("", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
			}
			return newPoolItem()
		},
		func(uid widget.TreeNodeID, branch bool, obj fyne.CanvasObject) {
			if branch {
				if uid == "local" {
					obj.(*widget.Label).SetText(fmt.Sprintf("本地（%d）", len(v.localFiltered)))
				}
				return
			}
			if p, ok := v.poolFor(uid); ok {
				obj.(*poolItem).set(*p)
			}
		},
	)
	v.tree.HideSeparators = true
	v.tree.OnSelected = func(uid widget.TreeNodeID) {
		if uid == "local" {
			v.tree.Unselect(uid)
			return
		}
		v.selectPool(uid)
	}
	v.tree.OpenAllBranches()

	v.statusLabel = widget.NewLabel("")
	v.statusLabel.Wrapping = fyne.TextWrapWord
	v.statusLabel.Hide()

	top := container.NewVBox(header, searchRow)
	v.leftPanel = container.NewBorder(top, v.statusLabel, nil, nil, v.tree)
	return v.leftPanel
}

// SetStatus shows a transient message at the bottom of the left panel, or
// hides it (and reclaims the space) when the message is empty.
func (v *Planner) SetStatus(text string) {
	if text == "" {
		v.statusLabel.SetText("")
		v.statusLabel.Hide()
	} else {
		v.statusLabel.SetText(text)
		v.statusLabel.Show()
	}
	if v.leftPanel != nil {
		v.leftPanel.Refresh()
	}
}

// buildRight creates the resource form and the result table.
func (v *Planner) buildRight() fyne.CanvasObject {
	v.formCard = widget.NewCard("", "", nil)

	v.keepResourceSelect = widget.NewSelect(nil, nil)

	v.rangeSlider = NewRangeSlider(1, 1)
	v.rangeSlider.Step = 1
	v.rangeSlider.OnChanged = func(lower, upper float64) {
		v.rangeLabel.SetText(fmt.Sprintf("第 %d 抽 ～ 第 %d 抽", int(lower), int(upper)))
	}

	v.rangeLabel = widget.NewLabel("")

	v.calculateBtn = widget.NewButton("计算方案", v.calculate)
	v.calculateBtn.Importance = widget.HighImportance

	v.summaryLabel = widget.NewLabel("")
	v.summaryLabel.Wrapping = fyne.TextWrapWord

	v.resultBox = container.NewVBox()

	resultTitle := widget.NewLabelWithStyle("计算方案", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	resultArea := container.NewBorder(resultTitle, v.summaryLabel, nil, nil, v.resultBox)

	right := container.NewBorder(v.formCard, nil, nil, nil, resultArea)
	// Wrap the whole panel so its minimum height stays small and the window
	// remains freely resizable; content scrolls when space is tight.
	return container.NewVScroll(right)
}

// applyFilter filters the list by name or id and keeps the selection in sync.
func (v *Planner) applyFilter(text string) {
	query := strings.ToLower(strings.TrimSpace(text))

	v.localFiltered = filterPools(v.localAll, query)

	// Re-resolve the current selection against the filtered lists.
	keep := (*pool.Pool)(nil)
	keepID := ""
	if v.selected != nil {
		if p, ok := v.poolFor(v.selectedNodeID); ok {
			keep = p
			keepID = v.selectedNodeID
		}
	}
	v.selected = keep
	if keep == nil {
		v.selectedNodeID = ""
	}

	// Synchronise the tree's selection state. Without this, an item that was
	// filtered out and later returned at the same id could not be selected
	// again, because Tree.Select early-returns on the stale id.
	v.tree.UnselectAll()
	if keepID != "" {
		v.tree.Select(keepID)
	}
	v.tree.Refresh()

	v.applySelection()
}

// filterPools returns the pools matching the query by name or id.
func filterPools(pools []pool.Pool, query string) []pool.Pool {
	if query == "" {
		return append([]pool.Pool(nil), pools...)
	}

	filtered := make([]pool.Pool, 0, len(pools))
	for _, p := range pools {
		if strings.Contains(strings.ToLower(p.Manifest.Name), query) ||
			strings.Contains(strings.ToLower(p.Manifest.ID), query) {
			filtered = append(filtered, p)
		}
	}
	return filtered
}

// leafIDs returns the tree leaf node ids of the current filter result.
func (v *Planner) leafIDs() []widget.TreeNodeID {
	ids := make([]widget.TreeNodeID, len(v.localFiltered))
	for i, p := range v.localFiltered {
		ids[i] = widget.TreeNodeID(p.Manifest.ID)
	}
	return ids
}

// poolFor resolves a tree leaf node id to its blind pool.
func (v *Planner) poolFor(nodeID string) (*pool.Pool, bool) {
	for i := range v.localFiltered {
		if v.localFiltered[i].Manifest.ID == nodeID {
			return &v.localFiltered[i], true
		}
	}
	return nil, false
}

// SetLocalPools 用新的本地盲盒池替换列表内容并重新过滤。
func (v *Planner) SetLocalPools(pools []pool.Pool) {
	v.localAll = pools
	v.applyFilter(v.searchEntry.Text)
}

// selectPool stores the chosen blind pool and rebuilds the right panel.
func (v *Planner) selectPool(nodeID widget.TreeNodeID) {
	p, ok := v.poolFor(nodeID)
	if !ok {
		return
	}
	v.selected = p
	v.selectedNodeID = nodeID
	v.applySelection()
}

// applySelection rebuilds the right panel to match the current selection.
func (v *Planner) applySelection() {
	v.plan = nil
	v.insufficient = false
	v.insufficientDraw = 0
	v.resourceEntries = nil
	v.rebuildResultTable()

	if v.selected == nil {
		v.formCard.SetTitle("资源规划")
		v.formCard.SetSubTitle("请选择一个盲盒")
		v.formCard.SetContent(container.NewCenter(widget.NewLabel("从左侧列表选择一个盲盒开始规划")))
		v.keepResourceSelect.Disable()
		v.rangeSlider.Disable()
		v.calculateBtn.Disable()
		v.summaryLabel.SetText("")
		return
	}

	p := v.selected
	v.formCard.SetTitle(p.Manifest.Name)
	v.formCard.SetSubTitle(fmt.Sprintf("共 %d 抽 · %d 种资源", p.Manifest.Draws, len(p.Resources)))

	// Resource quantity inputs.
	form := widget.NewForm()
	for _, resource := range p.Resources {
		entry := widget.NewEntry()
		entry.SetPlaceHolder("0")
		entry.Validator = numericValidator
		// Don't capture the mouse wheel: keep the panel scrollable over inputs.
		entry.Wrapping = fyne.TextWrapOff
		entry.Scroll = fyne.ScrollNone
		form.Append(resource.Name, entry)
		v.resourceEntries = append(v.resourceEntries, entry)
	}

	resourceTitle := widget.NewLabelWithStyle("资源数量", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	rangeTitle := widget.NewLabelWithStyle("抽数范围", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	keepTitle := widget.NewLabelWithStyle("优先保留的资源", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	// Draw range slider.
	draws := p.Manifest.Draws
	if draws < 1 {
		draws = 1
	}
	v.rangeSlider.SetRange(1, float64(draws))
	v.rangeSlider.SetValues(1, float64(draws))
	v.rangeLabel.SetText(fmt.Sprintf("第 1 抽 ～ 第 %d 抽", draws))
	if draws <= 1 {
		v.rangeSlider.Disable()
	} else {
		v.rangeSlider.Enable()
	}

	// Preferred resource selector. The first option "无" means no preference
	// (just maximise the number of draws).
	names := make([]string, 0, len(p.Resources)+1)
	names = append(names, "无")
	for _, resource := range p.Resources {
		names = append(names, resource.Name)
	}
	v.keepResourceSelect.SetOptions(names)
	v.keepResourceSelect.SetSelectedIndex(0)
	v.keepResourceSelect.Enable()

	v.calculateBtn.Enable()
	v.summaryLabel.SetText("填写资源数量后点击「计算方案」")

	content := container.NewVBox(
		resourceTitle,
		form,
		rangeTitle,
		v.rangeSlider,
		v.rangeLabel,
		keepTitle,
		v.keepResourceSelect,
		v.calculateBtn,
	)
	v.formCard.SetContent(content)
}

// calculate parses inputs, runs the plan and updates the table.
func (v *Planner) calculate() {
	if v.selected == nil {
		return
	}

	balances := make(map[string]int, len(v.selected.Resources))
	for i, resource := range v.selected.Resources {
		text := strings.TrimSpace(v.resourceEntries[i].Text)
		if text == "" {
			text = "0"
		}
		amount, err := strconv.Atoi(text)
		if err != nil || amount < 0 {
			v.summaryLabel.SetText(fmt.Sprintf("请输入有效的「%s」数量", resource.Name))
			return
		}
		balances[resource.ID] = amount
	}

	start := int(v.rangeSlider.Lower)
	end := int(v.rangeSlider.Upper)
	if start < 1 {
		start = 1
	}
	if end < start {
		end = start
	}

	result := pool.CalculatePlan(*v.selected, start, end, balances, v.preferredResourceID())

	v.plan = result.Steps
	v.insufficient = result.Insufficient
	v.insufficientDraw = result.FailAtDraw
	v.rebuildResultTable()
	v.updateSummary(result)
}

// preferredResourceID resolves the selected "keep more" resource id.
func (v *Planner) preferredResourceID() string {
	idx := v.keepResourceSelect.SelectedIndex()
	if idx <= 0 { // "无" (or no selection) means no preference
		return ""
	}
	i := idx - 1
	if i < len(v.selected.Resources) {
		return v.selected.Resources[i].ID
	}
	return ""
}

// updateSummary renders final balances and any insufficiency notice.
func (v *Planner) updateSummary(result pool.PlanResult) {
	parts := make([]string, 0, len(result.Final))
	for _, id := range pool.SortedResourceIDs(*v.selected) {
		parts = append(parts, fmt.Sprintf("%s %d", pool.ResourceName(*v.selected, id), result.Final[id]))
	}
	summary := "剩余资源：" + strings.Join(parts, "，")
	if result.Insufficient {
		summary = fmt.Sprintf("第 %d 抽资源不足，无法继续｜", result.FailAtDraw) + summary
	}
	v.summaryLabel.SetText(summary)
}

// rebuildResultTable redraws the result as a simple grid that grows to its full
// height. Unlike a widget.Table (which scrolls internally and collapses to a
// single visible row in a small window), this grid lets the outer scroll reveal
// every row at once.
func (v *Planner) rebuildResultTable() {
	rows := []fyne.CanvasObject{
		container.NewGridWithColumns(4,
			resultCell("抽数", true),
			resultCell("使用资源", true),
			resultCell("花费", true),
			resultCell("剩余", true),
		),
		widget.NewSeparator(),
	}

	for _, step := range v.plan {
		rows = append(rows, container.NewGridWithColumns(4,
			resultCell(strconv.Itoa(step.Draw), false),
			resultCell(step.ResourceName, false),
			resultCell(strconv.Itoa(step.Cost), false),
			resultCell(strconv.Itoa(step.Remaining), false),
		))
	}

	if v.insufficient {
		rows = append(rows, container.NewGridWithColumns(4,
			resultCell(strconv.Itoa(v.insufficientDraw), false),
			resultCell("资源不足", false),
			resultCell("—", false),
			resultCell("—", false),
		))
	}

	v.resultBox.Objects = rows
	v.resultBox.Refresh()
}

// resultCell builds a centred label cell for the result grid.
func resultCell(text string, bold bool) fyne.CanvasObject {
	return widget.NewLabelWithStyle(text, fyne.TextAlignCenter, fyne.TextStyle{Bold: bold})
}

// numericValidator accepts empty strings and non-negative integers.
func numericValidator(text string) error {
	if text == "" {
		return nil
	}
	for _, r := range text {
		if r < '0' || r > '9' {
			return errors.New("只能输入数字")
		}
	}
	return nil
}

// poolItem is a single row in the blind pool list.
type poolItem struct {
	widget.BaseWidget

	title    string
	subtitle string
}

func newPoolItem() fyne.CanvasObject {
	item := &poolItem{}
	item.ExtendBaseWidget(item)
	return item
}

func (b *poolItem) set(p pool.Pool) {
	b.title = p.Manifest.Name
	b.subtitle = fmt.Sprintf("%d 抽 · %d 种资源", p.Manifest.Draws, len(p.Resources))
	b.Refresh()
}

// MinSize returns the minimum size of the item.
func (b *poolItem) MinSize() fyne.Size {
	b.ExtendBaseWidget(b)
	return b.BaseWidget.MinSize()
}

// CreateRenderer creates the canvas objects for the item.
func (b *poolItem) CreateRenderer() fyne.WidgetRenderer {
	title := canvas.NewText(b.title, color.White)
	title.TextStyle = fyne.TextStyle{Bold: true}
	subtitle := canvas.NewText(b.subtitle, color.White)

	r := &poolItemRenderer{
		baseRenderer: baseRenderer{objects: []fyne.CanvasObject{title, subtitle}},
		title:        title,
		subtitle:     subtitle,
		item:         b,
	}
	r.Refresh()
	return r
}

type poolItemRenderer struct {
	baseRenderer

	title    *canvas.Text
	subtitle *canvas.Text
	item     *poolItem
}

func (r *poolItemRenderer) Refresh() {
	th := r.item.Theme()
	v := fyne.CurrentApp().Settings().ThemeVariant()

	r.title.Text = r.item.title
	r.title.Color = th.Color(theme.ColorNameForeground, v)
	r.subtitle.Text = r.item.subtitle
	r.subtitle.Color = th.Color(theme.ColorNamePlaceHolder, v)

	canvas.Refresh(r.item)
}

func (r *poolItemRenderer) Layout(size fyne.Size) {
	pad := r.item.Theme().Size(theme.SizeNamePadding)
	titleHeight := r.title.MinSize().Height

	r.title.Move(fyne.NewPos(pad, pad))
	r.title.Resize(fyne.NewSize(size.Width-pad*2, titleHeight))
	r.subtitle.Move(fyne.NewPos(pad, pad+titleHeight))
	r.subtitle.Resize(fyne.NewSize(size.Width-pad*2, r.subtitle.MinSize().Height))
}

func (r *poolItemRenderer) MinSize() fyne.Size {
	pad := r.item.Theme().Size(theme.SizeNamePadding)
	height := r.title.MinSize().Height + r.subtitle.MinSize().Height + pad*3
	return fyne.NewSize(r.title.MinSize().Width+pad*2, height)
}
