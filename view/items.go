package view

import (
	"fmt"
	"image/color"
	"strconv"

	"blind-tools/model/bid"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

const (
	// previewCellSize 占格预览里单个格子的边长。
	previewCellSize float32 = 34
	// itemCardWidth、itemCardHeight 拍品卡片的尺寸，GridWrap 按它换行。
	itemCardWidth  float32 = 190
	itemCardHeight float32 = 92
	// itemNameWidth、itemNameHeight 卡片里名称的固定显示区，
	// 超出部分交给 Label 的省略号截断，避免长名字把卡片撑宽。
	itemNameWidth  float32 = 170
	itemNameHeight float32 = 22
)

// Items 是「拍品清单」页：左侧按占格类型筛选，右侧画出该类型的占格并列出拍品卡片。
type Items struct {
	root fyne.CanvasObject

	// grids 是全部占格数据，shown 是当前选中类型下的拍品。
	grids []bid.Grid
	shown []bid.Item

	typeList     *widget.List
	statusLabel  *widget.Label
	previewLabel *widget.Label
	previewBox   *fyne.Container
	cardGrid     *widget.GridWrap
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
		v.shown = nil
		v.cardGrid.Refresh()
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

// build 组装左侧类型选择器与右侧占格预览、拍品卡片。
func (v *Items) build() fyne.CanvasObject {
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

	v.cardGrid = widget.NewGridWrap(
		func() int { return len(v.shown) },
		func() fyne.CanvasObject { return newItemCard() },
		func(id widget.GridWrapItemID, obj fyne.CanvasObject) {
			obj.(*itemCard).set(v.shown[id])
		},
	)

	preview := container.NewVBox(v.previewLabel, v.previewBox, widget.NewSeparator())

	// 内侧留空隙，让左右两块面板读起来是独立表面。
	left = container.New(layout.NewCustomPaddedLayout(0, 0, 0, 8), left)
	right := container.New(layout.NewCustomPaddedLayout(0, 0, 8, 0),
		container.NewBorder(preview, nil, nil, nil, v.cardGrid))

	split := container.NewHSplit(left, right)
	split.Offset = 0.22

	return split
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

	v.shown = grid.Items
	// 换类型时回到顶部：用 ScrollToOffset 而非 ScrollToTop，
	// 后者在渲染器尚未创建时会解引用空的 scroller。
	v.cardGrid.ScrollToOffset(0)
	v.cardGrid.Refresh()
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

// itemCard 是拍品清单里的一张拍品卡片：品质色块 + 品质名、名称与价格。
type itemCard struct {
	*fyne.Container

	name    *widget.Label
	quality *widget.Label
	swatch  *canvas.Rectangle
	value   *widget.Label
}

// newItemCard 构建一张空卡片，内容由 set 填充。
func newItemCard() *itemCard {
	background := canvas.NewRectangle(cardBackground())
	background.CornerRadius = 10
	background.SetMinSize(fyne.NewSize(itemCardWidth, itemCardHeight))

	name := widget.NewLabel("")
	name.Truncation = fyne.TextTruncateEllipsis

	quality := widget.NewLabel("")

	swatch := canvas.NewRectangle(color.Transparent)
	swatch.CornerRadius = 2
	swatch.SetMinSize(fyne.NewSize(qualitySwatchSize, qualitySwatchSize))

	value := widget.NewLabelWithStyle("", fyne.TextAlignTrailing, fyne.TextStyle{Bold: true})

	nameBox := container.NewGridWrap(fyne.NewSize(itemNameWidth, itemNameHeight), name)
	content := container.NewVBox(container.NewHBox(swatch, quality), nameBox, value)

	return &itemCard{
		Container: container.NewStack(background, container.NewPadded(content)),
		name:      name,
		quality:   quality,
		swatch:    swatch,
		value:     value,
	}
}

// set 用一条拍品数据填充卡片。
func (c *itemCard) set(item bid.Item) {
	c.name.SetText(item.Name)
	c.quality.SetText(qualityLabel(item.Quality))
	c.swatch.FillColor = qualityColor(item.Quality)
	c.swatch.Refresh()
	c.value.SetText(strconv.Itoa(item.Value))
}

// cardBackground 返回拍品卡片的底色。
func cardBackground() color.Color {
	settings := fyne.CurrentApp().Settings()
	return settings.Theme().Color(theme.ColorNameOverlayBackground, settings.ThemeVariant())
}

// gridTypeLabel 生成类型选择器里的一行：占格尺寸与拍品数。
func gridTypeLabel(grid bid.Grid) string {
	return fmt.Sprintf("%dx%d（%d 件）", grid.Length, grid.Width, len(grid.Items))
}
