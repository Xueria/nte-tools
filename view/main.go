package view

import (
	"blind-tools/model/blindbox"
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

// BoxSource 提供视图需要的盲盒数据来源：本地目录与远程仓库。
// 具体实现由 main 注入，视图不关心数据放在哪里。
type BoxSource interface {
	LocalBoxes() ([]blindbox.BlindBox, error)
	RemoteBoxes() ([]blindbox.BlindBox, error)
}

// MainView assembles the application UI as a tabbed shell: the blind box
// planner (list on the left, currency planning on the right) plus two
// placeholder tabs (竞拍估价 / 拍品清单) whose auction logic was removed.
func MainView(window fyne.Window, source BoxSource) fyne.CanvasObject {
	// Apply the Material Design 3 theme for the whole app.
	if a := fyne.CurrentApp(); a != nil {
		a.Settings().SetTheme(NewMD3Theme())
	}

	v := &plannerView{source: source}

	root := v.build()

	// Load the local section first so the list is populated, then the remote
	// section in the background so startup is not blocked.
	v.loadLocal()
	v.startRemoteLoad()

	tabs := container.NewAppTabs(
		container.NewTabItemWithIcon("盲盒规划", theme.HomeIcon(), root),
		container.NewTabItemWithIcon("竞拍估价", theme.SearchIcon(), newAuctionView()),
		container.NewTabItemWithIcon("拍品清单", theme.ListIcon(), newItemListView()),
	)
	// Keep the tab bar compact so both pages get the full window height.
	tabs.SetTabLocation(container.TabLocationTop)

	return tabs
}

// plannerView holds the mutable UI state.
type plannerView struct {
	source         BoxSource
	remoteAll      []blindbox.BlindBox
	localAll       []blindbox.BlindBox
	remoteFiltered []blindbox.BlindBox
	localFiltered  []blindbox.BlindBox
	selected       *blindbox.BlindBox
	selectedNodeID string
	remoteLoading  bool

	tree        *widget.Tree
	searchEntry *widget.Entry
	statusLabel *widget.Label
	leftPanel   *fyne.Container

	formCard           *widget.Card
	currencyEntries    []*widget.Entry
	keepCurrencySelect *widget.Select
	rangeSlider        *RangeSlider
	rangeLabel         *widget.Label
	calculateBtn       *widget.Button

	resultBox    *fyne.Container
	summaryLabel *widget.Label

	plan             []blindbox.PlanStep
	insufficient     bool
	insufficientDraw int
}

// build constructs the full split layout once.
func (v *plannerView) build() fyne.CanvasObject {
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

// buildLeft creates the search bar, refresh button and blind box list.
func (v *plannerView) buildLeft() fyne.CanvasObject {
	header := widget.NewLabelWithStyle("盲盒列表", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	v.searchEntry = widget.NewEntry()
	v.searchEntry.SetPlaceHolder("搜索盲盒名称…")
	v.searchEntry.OnChanged = v.applyFilter

	refreshBtn := widget.NewButtonWithIcon("刷新", theme.ViewRefreshIcon(), v.refresh)
	refreshBtn.Importance = widget.MediumImportance

	searchRow := container.NewBorder(nil, nil, nil, refreshBtn, v.searchEntry)

	// The list is split into two sections: remote (GitHub folder) and local
	// (the data folder).
	v.tree = widget.NewTree(
		func(uid widget.TreeNodeID) []widget.TreeNodeID {
			switch uid {
			case "":
				return []widget.TreeNodeID{"remote", "local"}
			case "remote":
				return v.sectionLeafIDs("remote")
			case "local":
				return v.sectionLeafIDs("local")
			}
			return nil
		},
		func(uid widget.TreeNodeID) bool {
			return uid == "" || uid == "remote" || uid == "local"
		},
		func(branch bool) fyne.CanvasObject {
			if branch {
				return widget.NewLabelWithStyle("", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
			}
			return newBlindBoxItem()
		},
		func(uid widget.TreeNodeID, branch bool, obj fyne.CanvasObject) {
			if branch {
				label := obj.(*widget.Label)
				switch uid {
				case "remote":
					switch {
					case v.remoteLoading:
						label.SetText("远程（加载中…）")
					default:
						label.SetText(fmt.Sprintf("远程（%d）", len(v.remoteFiltered)))
					}
				case "local":
					label.SetText(fmt.Sprintf("本地（%d）", len(v.localFiltered)))
				}
				return
			}
			if c, ok := v.boxFor(uid); ok {
				obj.(*blindBoxItem).set(*c)
			}
		},
	)
	v.tree.HideSeparators = true
	v.tree.OnSelected = func(uid widget.TreeNodeID) {
		if uid == "remote" || uid == "local" {
			v.tree.Unselect(uid)
			return
		}
		v.selectBox(uid)
	}
	v.tree.OpenAllBranches()

	v.statusLabel = widget.NewLabel("")
	v.statusLabel.Wrapping = fyne.TextWrapWord
	v.statusLabel.Hide()

	top := container.NewVBox(header, searchRow)
	v.leftPanel = container.NewBorder(top, v.statusLabel, nil, nil, v.tree)
	return v.leftPanel
}

// setStatus shows a transient message at the bottom of the left panel, or
// hides it (and reclaims the space) when the message is empty.
func (v *plannerView) setStatus(text string) {
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

// buildRight creates the currency form and the result table.
func (v *plannerView) buildRight() fyne.CanvasObject {
	v.formCard = widget.NewCard("", "", nil)

	v.keepCurrencySelect = widget.NewSelect(nil, nil)

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

// applyFilter filters both sections by name or id and keeps the selection in sync.
func (v *plannerView) applyFilter(text string) {
	query := strings.ToLower(strings.TrimSpace(text))

	v.remoteFiltered = filterBoxes(v.remoteAll, query)
	v.localFiltered = filterBoxes(v.localAll, query)

	// Re-resolve the current selection against the filtered lists.
	keep := (*blindbox.BlindBox)(nil)
	keepID := ""
	if v.selected != nil {
		if c, ok := v.boxFor(v.selectedNodeID); ok {
			keep = c
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

// filterBoxes returns the boxes matching the query by name or id.
func filterBoxes(boxes []blindbox.BlindBox, query string) []blindbox.BlindBox {
	if query == "" {
		return append([]blindbox.BlindBox(nil), boxes...)
	}

	filtered := make([]blindbox.BlindBox, 0, len(boxes))
	for _, box := range boxes {
		if strings.Contains(strings.ToLower(box.Manifest.Name), query) ||
			strings.Contains(strings.ToLower(box.Manifest.ID), query) {
			filtered = append(filtered, box)
		}
	}
	return filtered
}

// sectionLeafIDs returns the tree leaf node ids of a section ("remote"/"local").
func (v *plannerView) sectionLeafIDs(section string) []widget.TreeNodeID {
	list := v.remoteFiltered
	if section == "local" {
		list = v.localFiltered
	}

	ids := make([]widget.TreeNodeID, len(list))
	for i, box := range list {
		ids[i] = widget.TreeNodeID(section + ":" + box.Manifest.ID)
	}
	return ids
}

// boxFor resolves a tree leaf node id to its blind box.
func (v *plannerView) boxFor(nodeID string) (*blindbox.BlindBox, bool) {
	section, id, ok := strings.Cut(nodeID, ":")
	if !ok {
		return nil, false
	}

	var list []blindbox.BlindBox
	switch section {
	case "remote":
		list = v.remoteFiltered
	case "local":
		list = v.localFiltered
	default:
		return nil, false
	}

	for i := range list {
		if list[i].Manifest.ID == id {
			return &list[i], true
		}
	}
	return nil, false
}

// refresh reloads local and remote data so newly added blind boxes show up.
func (v *plannerView) refresh() {
	v.loadLocal()
	v.startRemoteLoad()
}

// loadLocal reloads the local section and reports the failure, if any.
// The list is refiltered either way so the tree keeps working.
func (v *plannerView) loadLocal() {
	local, err := v.source.LocalBoxes()
	if err != nil {
		v.setStatus(fmt.Sprintf("本地加载失败：%v", err))
	} else {
		v.localAll = local
		v.setStatus("")
	}

	v.applyFilter(v.searchEntry.Text)
}

// startRemoteLoad begins fetching the remote section in the background.
func (v *plannerView) startRemoteLoad() {
	v.remoteLoading = true
	v.tree.Refresh()

	go func() {
		boxes, err := v.source.RemoteBoxes()

		fyne.Do(func() {
			v.remoteLoading = false
			if err != nil {
				v.setStatus(fmt.Sprintf("远程加载失败：%v", err))
				return
			}
			v.remoteAll = boxes
			v.applyFilter(v.searchEntry.Text)
		})
	}()
}

// selectBox stores the chosen blind box and rebuilds the right panel.
func (v *plannerView) selectBox(nodeID widget.TreeNodeID) {
	c, ok := v.boxFor(nodeID)
	if !ok {
		return
	}
	v.selected = c
	v.selectedNodeID = nodeID
	v.applySelection()
}

// applySelection rebuilds the right panel to match the current selection.
func (v *plannerView) applySelection() {
	v.plan = nil
	v.insufficient = false
	v.insufficientDraw = 0
	v.currencyEntries = nil
	v.rebuildResultTable()

	if v.selected == nil {
		v.formCard.SetTitle("货币规划")
		v.formCard.SetSubTitle("请选择一个盲盒")
		v.formCard.SetContent(container.NewCenter(widget.NewLabel("从左侧列表选择一个盲盒开始规划")))
		v.keepCurrencySelect.Disable()
		v.rangeSlider.Disable()
		v.calculateBtn.Disable()
		v.summaryLabel.SetText("")
		return
	}

	c := v.selected
	v.formCard.SetTitle(c.Manifest.Name)
	v.formCard.SetSubTitle(fmt.Sprintf("共 %d 抽 · %d 种货币", c.Manifest.Draws, len(c.Currencies)))

	// Currency quantity inputs.
	form := widget.NewForm()
	for _, currency := range c.Currencies {
		entry := widget.NewEntry()
		entry.SetPlaceHolder("0")
		entry.Validator = numericValidator
		// Don't capture the mouse wheel: keep the panel scrollable over inputs.
		entry.Wrapping = fyne.TextWrapOff
		entry.Scroll = fyne.ScrollNone
		form.Append(currency.Name, entry)
		v.currencyEntries = append(v.currencyEntries, entry)
	}

	currencyTitle := widget.NewLabelWithStyle("货币数量", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	rangeTitle := widget.NewLabelWithStyle("抽数范围", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	keepTitle := widget.NewLabelWithStyle("优先保留的货币", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	// Draw range slider.
	draws := c.Manifest.Draws
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

	// Preferred currency selector. The first option "无" means no preference
	// (just maximise the number of draws).
	names := make([]string, 0, len(c.Currencies)+1)
	names = append(names, "无")
	for _, currency := range c.Currencies {
		names = append(names, currency.Name)
	}
	v.keepCurrencySelect.SetOptions(names)
	v.keepCurrencySelect.SetSelectedIndex(0)
	v.keepCurrencySelect.Enable()

	v.calculateBtn.Enable()
	v.summaryLabel.SetText("填写货币数量后点击「计算方案」")

	content := container.NewVBox(
		currencyTitle,
		form,
		rangeTitle,
		v.rangeSlider,
		v.rangeLabel,
		keepTitle,
		v.keepCurrencySelect,
		v.calculateBtn,
	)
	v.formCard.SetContent(content)
}

// calculate parses inputs, runs the plan and updates the table.
func (v *plannerView) calculate() {
	if v.selected == nil {
		return
	}

	balances := make(map[string]int, len(v.selected.Currencies))
	for i, currency := range v.selected.Currencies {
		text := strings.TrimSpace(v.currencyEntries[i].Text)
		if text == "" {
			text = "0"
		}
		amount, err := strconv.Atoi(text)
		if err != nil || amount < 0 {
			v.summaryLabel.SetText(fmt.Sprintf("请输入有效的「%s」数量", currency.Name))
			return
		}
		balances[currency.ID] = amount
	}

	start := int(v.rangeSlider.Lower)
	end := int(v.rangeSlider.Upper)
	if start < 1 {
		start = 1
	}
	if end < start {
		end = start
	}

	result := blindbox.CalculatePlan(*v.selected, start, end, balances, v.preferredCurrencyID())

	v.plan = result.Steps
	v.insufficient = result.Insufficient
	v.insufficientDraw = result.FailAtDraw
	v.rebuildResultTable()
	v.updateSummary(result)
}

// preferredCurrencyID resolves the selected "keep more" currency id.
func (v *plannerView) preferredCurrencyID() string {
	idx := v.keepCurrencySelect.SelectedIndex()
	if idx <= 0 { // "无" (or no selection) means no preference
		return ""
	}
	i := idx - 1
	if i < len(v.selected.Currencies) {
		return v.selected.Currencies[i].ID
	}
	return ""
}

// updateSummary renders final balances and any insufficiency notice.
func (v *plannerView) updateSummary(result blindbox.PlanResult) {
	parts := make([]string, 0, len(result.Final))
	for _, id := range blindbox.SortedCurrencyIDs(*v.selected) {
		parts = append(parts, fmt.Sprintf("%s %d", blindbox.CurrencyName(*v.selected, id), result.Final[id]))
	}
	summary := "剩余货币：" + strings.Join(parts, "，")
	if result.Insufficient {
		summary = fmt.Sprintf("第 %d 抽货币不足，无法继续｜", result.FailAtDraw) + summary
	}
	v.summaryLabel.SetText(summary)
}

// rebuildResultTable redraws the result as a simple grid that grows to its full
// height. Unlike a widget.Table (which scrolls internally and collapses to a
// single visible row in a small window), this grid lets the outer scroll reveal
// every row at once.
func (v *plannerView) rebuildResultTable() {
	rows := []fyne.CanvasObject{
		container.NewGridWithColumns(4,
			resultCell("抽数", true),
			resultCell("使用货币", true),
			resultCell("花费", true),
			resultCell("剩余", true),
		),
		widget.NewSeparator(),
	}

	for _, step := range v.plan {
		rows = append(rows, container.NewGridWithColumns(4,
			resultCell(strconv.Itoa(step.Draw), false),
			resultCell(step.CurrencyName, false),
			resultCell(strconv.Itoa(step.Cost), false),
			resultCell(strconv.Itoa(step.Remaining), false),
		))
	}

	if v.insufficient {
		rows = append(rows, container.NewGridWithColumns(4,
			resultCell(strconv.Itoa(v.insufficientDraw), false),
			resultCell("货币不足", false),
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

// blindBoxItem is a single row in the blind box list.
type blindBoxItem struct {
	widget.BaseWidget

	title    string
	subtitle string
}

func newBlindBoxItem() fyne.CanvasObject {
	item := &blindBoxItem{}
	item.ExtendBaseWidget(item)
	return item
}

func (b *blindBoxItem) set(box blindbox.BlindBox) {
	b.title = box.Manifest.Name
	b.subtitle = fmt.Sprintf("%d 抽 · %d 种货币", box.Manifest.Draws, len(box.Currencies))
	b.Refresh()
}

// MinSize returns the minimum size of the item.
func (b *blindBoxItem) MinSize() fyne.Size {
	b.ExtendBaseWidget(b)
	return b.BaseWidget.MinSize()
}

// CreateRenderer creates the canvas objects for the item.
func (b *blindBoxItem) CreateRenderer() fyne.WidgetRenderer {
	title := canvas.NewText(b.title, color.White)
	title.TextStyle = fyne.TextStyle{Bold: true}
	subtitle := canvas.NewText(b.subtitle, color.White)

	r := &blindBoxItemRenderer{
		baseRenderer: baseRenderer{objects: []fyne.CanvasObject{title, subtitle}},
		title:        title,
		subtitle:     subtitle,
		item:         b,
	}
	r.Refresh()
	return r
}

type blindBoxItemRenderer struct {
	baseRenderer

	title    *canvas.Text
	subtitle *canvas.Text
	item     *blindBoxItem
}

func (r *blindBoxItemRenderer) Refresh() {
	th := r.item.Theme()
	v := fyne.CurrentApp().Settings().ThemeVariant()

	r.title.Text = r.item.title
	r.title.Color = th.Color(theme.ColorNameForeground, v)
	r.subtitle.Text = r.item.subtitle
	r.subtitle.Color = th.Color(theme.ColorNamePlaceHolder, v)

	canvas.Refresh(r.item)
}

func (r *blindBoxItemRenderer) Layout(size fyne.Size) {
	pad := r.item.Theme().Size(theme.SizeNamePadding)
	titleHeight := r.title.MinSize().Height

	r.title.Move(fyne.NewPos(pad, pad))
	r.title.Resize(fyne.NewSize(size.Width-pad*2, titleHeight))
	r.subtitle.Move(fyne.NewPos(pad, pad+titleHeight))
	r.subtitle.Resize(fyne.NewSize(size.Width-pad*2, r.subtitle.MinSize().Height))
}

func (r *blindBoxItemRenderer) MinSize() fyne.Size {
	pad := r.item.Theme().Size(theme.SizeNamePadding)
	height := r.title.MinSize().Height + r.subtitle.MinSize().Height + pad*3
	return fyne.NewSize(r.title.MinSize().Width+pad*2, height)
}
