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
	itemCardHeight float32 = 96
	// itemRowHeight 卡片里每一行的高度。
	itemRowHeight float32 = 24
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

// itemCard 是拍品清单里的一张拍品卡片：品质色块与品质名、名称、价格。
// 它自己负责绘制，尺寸固定、不随文字内容变化——GridWrap 的 item 是先创建、
// 后由数据填充的，用嵌套容器排版会按空文字把行高算成零，文字就再也显示不出来。
type itemCard struct {
	widget.BaseWidget

	item bid.Item
}

// newItemCard 构建一张空卡片，内容由 set 填充。
func newItemCard() *itemCard {
	card := &itemCard{}
	card.ExtendBaseWidget(card)
	return card
}

// set 用一条拍品数据填充卡片。
func (c *itemCard) set(item bid.Item) {
	c.item = item
	c.Refresh()
}

// MinSize 返回卡片尺寸，GridWrap 按它换行。
func (c *itemCard) MinSize() fyne.Size {
	c.ExtendBaseWidget(c)
	return fyne.NewSize(itemCardWidth, itemCardHeight)
}

// CreateRenderer 创建卡片的绘制对象。
func (c *itemCard) CreateRenderer() fyne.WidgetRenderer {
	background := canvas.NewRectangle(color.Transparent)
	background.CornerRadius = 10

	swatch := canvas.NewRectangle(color.Transparent)
	swatch.CornerRadius = 2

	quality := canvas.NewText("", color.White)
	name := canvas.NewText("", color.White)
	value := canvas.NewText("", color.White)
	value.Alignment = fyne.TextAlignTrailing

	r := &itemCardRenderer{
		baseRenderer: baseRenderer{objects: []fyne.CanvasObject{background, swatch, quality, name, value}},
		card:         c,
		background:   background,
		swatch:       swatch,
		quality:      quality,
		name:         name,
		value:        value,
	}
	r.Refresh()
	return r
}

type itemCardRenderer struct {
	baseRenderer

	card       *itemCard
	background *canvas.Rectangle
	swatch     *canvas.Rectangle
	quality    *canvas.Text
	name       *canvas.Text
	value      *canvas.Text
}

func (r *itemCardRenderer) Refresh() {
	th := r.card.Theme()
	variant := fyne.CurrentApp().Settings().ThemeVariant()

	r.background.FillColor = th.Color(theme.ColorNameOverlayBackground, variant)
	r.swatch.FillColor = qualityColor(r.card.item.Quality)

	r.quality.Text = qualityLabel(r.card.item.Quality)
	r.name.Text = r.card.item.Name
	r.value.Text = strconv.Itoa(r.card.item.Value)

	r.quality.Color = th.Color(theme.ColorNameForeground, variant)
	r.name.Color = th.Color(theme.ColorNameForeground, variant)
	r.value.Color = th.Color(theme.ColorNameForeground, variant)

	canvas.Refresh(r.card)
}

func (r *itemCardRenderer) Layout(size fyne.Size) {
	pad := r.card.Theme().Size(theme.SizeNamePadding)
	left := pad * 2
	textWidth := size.Width - pad*4

	r.background.Resize(size)

	r.swatch.Move(fyne.NewPos(left, left))
	r.swatch.Resize(fyne.NewSize(qualitySwatchSize, qualitySwatchSize))

	r.quality.Move(fyne.NewPos(left+qualitySwatchSize+pad, left))
	r.quality.Resize(fyne.NewSize(textWidth-qualitySwatchSize-pad, itemRowHeight))

	r.name.Move(fyne.NewPos(left, left+itemRowHeight+pad))
	r.name.Resize(fyne.NewSize(textWidth, itemRowHeight))

	r.value.Move(fyne.NewPos(left, left+2*(itemRowHeight+pad)))
	r.value.Resize(fyne.NewSize(textWidth, itemRowHeight))
}

func (r *itemCardRenderer) MinSize() fyne.Size {
	return fyne.NewSize(itemCardWidth, itemCardHeight)
}

// gridTypeLabel 生成类型选择器里的一行：占格尺寸与拍品数。
func gridTypeLabel(grid bid.Grid) string {
	return fmt.Sprintf("%dx%d（%d 件）", grid.Length, grid.Width, len(grid.Items))
}
