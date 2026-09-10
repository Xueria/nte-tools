package view

import (
	"fmt"
	"strconv"

	"blind-tools/model/bid"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// previewCellSize 占格预览里单个格子的边长。
const previewCellSize float32 = 34

// Items 是「拍品清单」页：左侧按占格类型筛选，右侧画出该类型的占格并列出拍品。
type Items struct {
	root fyne.CanvasObject

	grids []bid.Grid

	typeList     *widget.List
	statusLabel  *widget.Label
	previewLabel *widget.Label
	previewBox   *fyne.Container
	tableBox     *fyne.Container
}

// NewItems 构建拍品清单页。
func NewItems() *Items {
	v := &Items{}
	v.root = v.build()
	return v
}

// Tab 返回该页的页签标题、图标与内容。
func (v *Items) Tab() Tab {
	return Tab{Title: "拍品清单", Icon: theme.ListIcon(), Content: v.root}
}

// SetBidGrids 用新的竞拍占格数据替换页面内容，并默认选中第一个类型。
func (v *Items) SetBidGrids(grids []bid.Grid) {
	v.grids = grids
	v.typeList.Refresh()

	if len(grids) == 0 {
		v.previewLabel.SetText("暂无竞拍数据")
		v.previewBox.Objects = nil
		v.previewBox.Refresh()
		v.tableBox.Objects = nil
		v.tableBox.Refresh()
		return
	}

	// 先清空选中态再选中，保证 OnSelected 一定触发。
	v.typeList.UnselectAll()
	v.typeList.Select(0)
}

// SetStatus 显示加载状态；传空字符串即隐藏。
func (v *Items) SetStatus(text string) {
	if text == "" {
		v.statusLabel.SetText("")
		v.statusLabel.Hide()
	} else {
		v.statusLabel.SetText(text)
		v.statusLabel.Show()
	}
}

// build 组装左侧类型选择器与右侧占格预览、拍品表格。
func (v *Items) build() fyne.CanvasObject {
	title := widget.NewLabelWithStyle("拍品清单", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	v.typeList = widget.NewList(
		func() int { return len(v.grids) },
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			obj.(*widget.Label).SetText(gridTypeLabel(v.grids[id]))
		},
	)
	v.typeList.OnSelected = func(id widget.ListItemID) { v.selectGrid(int(id)) }

	v.statusLabel = widget.NewLabel("")
	v.statusLabel.Wrapping = fyne.TextWrapWord
	v.statusLabel.Hide()

	typeHeader := widget.NewLabelWithStyle("占格类型", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	header := container.NewVBox(typeHeader, widget.NewSeparator())
	left := container.NewBorder(header, v.statusLabel, nil, nil, v.typeList)

	v.previewLabel = widget.NewLabelWithStyle("", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	v.previewBox = container.NewVBox()
	v.tableBox = container.NewVBox()

	detail := container.NewVBox(v.previewLabel, v.previewBox, widget.NewSeparator(), v.tableBox)

	// 内侧留空隙，让左右两块面板读起来是独立表面。
	left = container.New(layout.NewCustomPaddedLayout(0, 0, 0, 8), left)
	right := container.New(layout.NewCustomPaddedLayout(0, 0, 8, 0), container.NewVScroll(detail))

	split := container.NewHSplit(left, right)
	split.Offset = 0.22

	return container.NewBorder(container.NewVBox(title, widget.NewSeparator()), nil, nil, nil, split)
}

// selectGrid 渲染第 index 个占格类型。
func (v *Items) selectGrid(index int) {
	if index < 0 || index >= len(v.grids) {
		return
	}

	grid := v.grids[index]

	v.previewLabel.SetText(fmt.Sprintf("%dx%d 占格 · %d 件拍品", grid.Length, grid.Width, len(grid.Items)))
	v.previewBox.Objects = []fyne.CanvasObject{footprint(grid.Length, grid.Width)}
	v.previewBox.Refresh()
	v.rebuildTable(grid)
}

// footprint 画出 length 列、width 行的占格形状。
func footprint(length, width int) fyne.CanvasObject {
	cells := make([]fyne.CanvasObject, 0, length*width)

	for i := 0; i < length*width; i++ {
		cells = append(cells, previewCell())
	}

	// 居中放置：格子保持正方，不被布局拉伸。
	return container.NewCenter(container.NewGridWithColumns(length, cells...))
}

// previewCell 生成占格预览里的一个格子。
func previewCell() fyne.CanvasObject {
	settings := fyne.CurrentApp().Settings()
	variant := settings.ThemeVariant()

	cell := canvas.NewRectangle(settings.Theme().Color(theme.ColorNameInputBackground, variant))
	cell.StrokeColor = settings.Theme().Color(theme.ColorNameInputBorder, variant)
	cell.StrokeWidth = 1
	cell.CornerRadius = 4
	cell.SetMinSize(fyne.NewSize(previewCellSize, previewCellSize))

	return cell
}

// rebuildTable 重建拍品表格：名称、品质、价格三列。
func (v *Items) rebuildTable(grid bid.Grid) {
	rows := []fyne.CanvasObject{
		container.NewGridWithColumns(3,
			tableCell("名称", true, fyne.TextAlignLeading),
			tableCell("品质", true, fyne.TextAlignLeading),
			tableCell("价格", true, fyne.TextAlignTrailing),
		),
		widget.NewSeparator(),
	}

	for _, item := range grid.Items {
		rows = append(rows, container.NewGridWithColumns(3,
			widget.NewLabel(item.Name),
			container.NewHBox(qualitySwatch(item.Quality), widget.NewLabel(qualityLabel(item.Quality))),
			tableCell(strconv.Itoa(item.Value), false, fyne.TextAlignTrailing),
		))
	}

	v.tableBox.Objects = rows
	v.tableBox.Refresh()
}

// tableCell 生成一个表格单元。
func tableCell(text string, bold bool, align fyne.TextAlign) fyne.CanvasObject {
	return widget.NewLabelWithStyle(text, align, fyne.TextStyle{Bold: bold})
}

// gridTypeLabel 生成类型选择器里的一行：占格尺寸与拍品数。
func gridTypeLabel(grid bid.Grid) string {
	return fmt.Sprintf("%dx%d（%d 件）", grid.Length, grid.Width, len(grid.Items))
}
