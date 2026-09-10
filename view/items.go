package view

import (
	"fmt"
	"image/color"

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
	// cardMinWidth 卡片的最小宽度：列数由可用宽度能放下多少个它决定。
	cardMinWidth float32 = 110
	// itemCardHeight 卡片高度；卡片内三行自上而下按固定行高排布。
	itemCardHeight float32 = 62
	// cardMinGap、cardMaxGap 卡片间距的下限与上限：列间剩余宽度会动态
	// 分到间距上，超过上限的部分改为把卡片均分撑满，右边缘始终贴齐。
	cardMinGap float32 = 6
	cardMaxGap float32 = 12
	// cardInset 卡片内容与卡片边缘的距离。
	cardInset float32 = 6
	// cardQualityRowHeight、cardRowHeight 卡片内品质行与文字行的高度。
	cardQualityRowHeight float32 = 14
	cardRowHeight        float32 = 16
	// cardNameTextSize、cardQualityTextSize 卡片内的字号。
	cardNameTextSize    float32 = 12
	cardQualityTextSize float32 = 11
)

// Items 是「拍品清单」页：左侧按占格类型筛选，右侧画出该类型的占格并列出拍品卡片。
type Items struct {
	root fyne.CanvasObject

	grids []bid.Grid

	typeList     *widget.List
	statusLabel  *widget.Label
	previewLabel *widget.Label
	previewBox   *fyne.Container
	cardWall     *cardWall
	cardScroll   *container.Scroll
}

// NewItems 构建拍品清单页。
func NewItems() *Items {
	v := &Items{}
	v.root = v.build()
	return v
}

// NewTab 返回该页的页签标题、图标与内容。
func (v *Items) NewTab() Tab {
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
		v.cardWall.setCards(nil)
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

	v.cardWall = newCardWall()
	// 滚动条浮在滚动内容之上，所以把它的宽度留在内容右侧：滚动条仍停在
	// 面板右边缘，卡片区则缩短，最右一列不会被压住。
	cardContent := container.New(layout.NewCustomPaddedLayout(0, 0, 0, scrollBarInset()), v.cardWall)
	v.cardScroll = container.NewVScroll(cardContent)

	preview := container.NewVBox(v.previewLabel, v.previewBox, widget.NewSeparator())

	// 内侧留空隙，让左右两块面板读起来是独立表面。
	left = container.New(layout.NewCustomPaddedLayout(0, 0, 0, 8), left)
	right := container.New(layout.NewCustomPaddedLayout(0, 0, 8, 0),
		container.NewBorder(preview, nil, nil, nil, v.cardScroll))

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

	cards := make([]fyne.CanvasObject, 0, len(grid.Items))

	for _, item := range grid.Items {
		card := newItemCard()
		card.set(item)
		cards = append(cards, card)
	}

	v.cardWall.setCards(cards)

	// 换类型时回到顶部。
	v.cardScroll.Offset = fyne.NewPos(0, 0)
	v.cardScroll.Refresh()
}

// scrollBarInset 返回浮动滚动条的宽度加一点余量：Fyne 的滚动条画在内容
// 之上，不预留空间就会遮住最右一列的信息。
func scrollBarInset() float32 {
	return theme.Current().Size(theme.SizeNameScrollBar) + 6
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

// cardWall 把卡片按可用宽度分列摆放：列数随宽度自适应，列间剩余宽度
// 动态分到间距上，因此右侧不会留下空白。
type cardWall struct {
	widget.BaseWidget

	cards   []fyne.CanvasObject
	columns int
	rows    int
}

// newCardWall 构建一个空的卡片墙。
func newCardWall() *cardWall {
	wall := &cardWall{columns: 1, rows: 1}
	wall.ExtendBaseWidget(wall)
	return wall
}

// setCards 用新的卡片替换墙面内容。
func (w *cardWall) setCards(cards []fyne.CanvasObject) {
	w.cards = cards
	w.columns, w.rows = 1, 1
	w.Refresh()
}

// MinSize 返回当前列数下的整体尺寸。
func (w *cardWall) MinSize() fyne.Size {
	w.ExtendBaseWidget(w)

	if len(w.cards) == 0 {
		return fyne.NewSize(cardMinWidth, 0)
	}

	return fyne.NewSize(cardMinWidth, w.contentHeight())
}

// contentHeight 返回所有行叠起来的高度。
func (w *cardWall) contentHeight() float32 {
	rows := w.rows
	if rows < 1 {
		rows = 1
	}

	return float32(rows)*itemCardHeight + float32(rows-1)*cardMinGap
}

// CreateRenderer 创建卡片墙的绘制对象。
func (w *cardWall) CreateRenderer() fyne.WidgetRenderer {
	return &cardWallRenderer{baseRenderer: baseRenderer{}, wall: w}
}

type cardWallRenderer struct {
	baseRenderer

	wall *cardWall
}

func (r *cardWallRenderer) Layout(size fyne.Size) {
	cards := r.wall.cards
	r.SetObjects(cards)

	if len(cards) == 0 {
		return
	}

	columns := cardColumnsFor(size.Width, len(cards))
	cardWidth, gap := cardMetrics(size.Width, columns)

	rows := (len(cards) + columns - 1) / columns
	changed := columns != r.wall.columns || rows != r.wall.rows
	r.wall.columns, r.wall.rows = columns, rows

	for i, card := range cards {
		row, column := i/columns, i%columns

		card.Move(fyne.NewPos(float32(column)*(cardWidth+gap), float32(row)*(itemCardHeight+cardMinGap)))
		card.Resize(fyne.NewSize(cardWidth, itemCardHeight))
	}

	if changed {
		// 列数或行数变了，MinSize 随之变化，请父级（滚动容器）重新布局。
		canvas.Refresh(r.wall)
	}
}

func (r *cardWallRenderer) Refresh() {
	r.SetObjects(r.wall.cards)
	canvas.Refresh(r.wall)
}

func (r *cardWallRenderer) MinSize() fyne.Size {
	return r.wall.MinSize()
}

// cardColumnsFor 返回给定宽度放下多少列卡片。
func cardColumnsFor(width float32, count int) int {
	if width <= 0 || count < 1 {
		return 1
	}

	columns := int((width + cardMinGap) / (cardMinWidth + cardMinGap))

	if columns < 1 {
		columns = 1
	}

	if columns > count {
		columns = count
	}

	return columns
}

// cardMetrics 返回卡片宽度与列间距：剩余宽度先分给间距，超过上限后
// 改为把卡片均分撑满，保证最右侧与左边缘一样贴齐。
func cardMetrics(width float32, columns int) (cardWidth, gap float32) {
	if columns < 2 {
		return width, 0
	}

	gap = (width - cardMinWidth*float32(columns)) / float32(columns-1)

	if gap <= cardMaxGap {
		return cardMinWidth, gap
	}

	gap = cardMaxGap

	return (width - gap*float32(columns-1)) / float32(columns), gap
}

// itemCard 是拍品清单里的一张拍品卡片：品质色块与品质名、名称、价格。
// 它自己负责绘制，尺寸与文字内容无关——卡片可能先创建、后填充数据，
// 用嵌套容器排版会按空文字把行高算成零，文字就再也显示不出来。
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

// MinSize 返回卡片的建议最小尺寸。
func (c *itemCard) MinSize() fyne.Size {
	c.ExtendBaseWidget(c)
	return fyne.NewSize(cardMinWidth, itemCardHeight)
}

// CreateRenderer 创建卡片的绘制对象。
func (c *itemCard) CreateRenderer() fyne.WidgetRenderer {
	background := canvas.NewRectangle(color.Transparent)
	background.CornerRadius = 8
	background.StrokeWidth = 1

	swatch := canvas.NewRectangle(color.Transparent)
	swatch.CornerRadius = 2

	quality := canvas.NewText("", color.White)
	quality.TextSize = cardQualityTextSize

	name := canvas.NewText("", color.White)
	name.TextSize = cardNameTextSize

	value := canvas.NewText("", color.White)
	value.Alignment = fyne.TextAlignTrailing
	value.TextSize = cardNameTextSize
	value.TextStyle = fyne.TextStyle{Bold: true}

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
	r.background.StrokeColor = th.Color(theme.ColorNameSeparator, variant)

	r.swatch.FillColor = qualityColor(r.card.item.Quality)

	r.quality.Text = qualityLabel(r.card.item.Quality)
	r.quality.Color = qualityColor(r.card.item.Quality)

	r.name.Color = th.Color(theme.ColorNameForeground, variant)

	r.value.Text = formatValue(r.card.item.Value)
	r.value.Color = th.Color(theme.ColorNamePrimary, variant)

	r.fitName(r.card.Size().Width - cardInset*2)

	canvas.Refresh(r.card)
}

func (r *itemCardRenderer) Layout(size fyne.Size) {
	textWidth := size.Width - cardInset*2

	r.background.Resize(size)

	r.swatch.Move(fyne.NewPos(cardInset, cardInset+(cardQualityRowHeight-qualitySwatchSize)/2))
	r.swatch.Resize(fyne.NewSize(qualitySwatchSize, qualitySwatchSize))

	qualityLeft := cardInset + qualitySwatchSize + 4
	r.quality.Move(fyne.NewPos(qualityLeft, cardInset))
	r.quality.Resize(fyne.NewSize(textWidth-qualitySwatchSize-4, cardQualityRowHeight))

	nameTop := cardInset + cardQualityRowHeight + 2
	r.name.Move(fyne.NewPos(cardInset, nameTop))
	r.name.Resize(fyne.NewSize(textWidth, cardRowHeight))

	r.value.Move(fyne.NewPos(cardInset, nameTop+cardRowHeight+2))
	r.value.Resize(fyne.NewSize(textWidth, cardRowHeight))

	r.fitName(textWidth)
}

func (r *itemCardRenderer) MinSize() fyne.Size {
	return fyne.NewSize(cardMinWidth, itemCardHeight)
}

// fitName 按可用宽度截断名称，超出部分用省略号。
func (r *itemCardRenderer) fitName(width float32) {
	r.name.Text = fitText(r.card.item.Name, width, cardNameTextSize, fyne.TextStyle{})
}

// gridTypeLabel 生成类型选择器里的一行：占格尺寸与拍品数。
func gridTypeLabel(grid bid.Grid) string {
	return fmt.Sprintf("%dx%d（%d 件）", grid.Length, grid.Width, len(grid.Items))
}
