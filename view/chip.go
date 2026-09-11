package view

import (
	"fmt"
	"image/color"

	"blind-tools/model/bid"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

const (
	// chipHeight、chipGap 标签的高度与横竖间距。
	chipHeight float32 = 28
	chipGap    float32 = 6
	// chipPadding 标签内文字两侧的留白。
	chipPadding float32 = 10
	// chipDotSize、chipDotGap 品质色点的尺寸与它到名称的距离。
	chipDotSize float32 = 8
	chipDotGap  float32 = 6
	// chipTextSize 标签字号。
	chipTextSize float32 = 12
	// chipMinWidth 标签的最小宽度，用于撑起标签流的宽度。
	chipMinWidth float32 = 40
	// selectedCrossSize、selectedCrossPadding 已选择标签里叉的尺寸与右侧留白。
	selectedCrossSize    float32 = 12
	selectedCrossPadding float32 = 8
)

// chip 是可选物品区里的一个标签：左键点一下加入一件，右键减少一件。同一物品可以
// 加多件（标签上显示 ×N），加过的标签高亮。
type chip struct {
	widget.BaseWidget

	page  *Bid
	index int
	item  bid.Item
	count int
}

// newChip 构建一个空标签，内容由 set 填充。
func newChip(page *Bid) *chip {
	item := &chip{page: page}
	item.ExtendBaseWidget(item)

	return item
}

// set 设置标签对应的物品与已确认件数。
func (c *chip) set(index int, item bid.Item, count int) {
	c.index = index
	c.item = item
	c.count = count
	c.Refresh()
}

// Tapped 左键点击即加入一件。
func (c *chip) Tapped(*fyne.PointEvent) {
	if c.page != nil {
		c.page.addItem(c.index)
	}
}

// TappedSecondary 右键（或长按）点击即减少一件。
func (c *chip) TappedSecondary(*fyne.PointEvent) {
	if c.page != nil {
		c.page.removeItem(c.index)
	}
}

// MinSize 返回标签尺寸：宽度随文字（带 ×N 时更宽）变化，一行能放多少就放多少。
func (c *chip) MinSize() fyne.Size {
	c.ExtendBaseWidget(c)

	return fyne.NewSize(c.labelWidth()+chipPadding*2+chipDotSize+chipDotGap, chipHeight)
}

// labelWidth 返回标签文字需要的宽度。
func (c *chip) labelWidth() float32 {
	return fyne.MeasureText(chipTitle(c.item.Name, c.count), chipTextSize, fyne.TextStyle{}).Width
}

// CreateRenderer 创建标签的绘制对象。
func (c *chip) CreateRenderer() fyne.WidgetRenderer {
	background := canvas.NewRectangle(color.Transparent)
	background.CornerRadius = chipHeight / 2

	dot := canvas.NewRectangle(color.Transparent)
	dot.CornerRadius = chipDotSize / 2

	name := canvas.NewText("", color.White)
	name.TextSize = chipTextSize

	renderer := &chipRenderer{
		baseRenderer: baseRenderer{objects: []fyne.CanvasObject{background, dot, name}},
		chip:         c,
		background:   background,
		dot:          dot,
		name:         name,
	}
	renderer.Refresh()

	return renderer
}

type chipRenderer struct {
	baseRenderer

	chip       *chip
	background *canvas.Rectangle
	dot        *canvas.Rectangle
	name       *canvas.Text
}

func (r *chipRenderer) Refresh() {
	th := r.chip.Theme()
	variant := fyne.CurrentApp().Settings().ThemeVariant()

	if r.chip.count > 0 {
		r.background.FillColor = th.Color(theme.ColorNamePrimary, variant)
		r.background.StrokeColor = r.background.FillColor
		r.name.Color = th.Color(theme.ColorNameForegroundOnPrimary, variant)
	} else {
		// 未加过的标签也要与背景分开：用 surfaceVariant 填充 + outline 描边。
		r.background.FillColor = th.Color(theme.ColorNameInputBackground, variant)
		r.background.StrokeColor = th.Color(theme.ColorNameInputBorder, variant)
		r.name.Color = th.Color(theme.ColorNameForeground, variant)
	}

	r.dot.FillColor = qualityColor(r.chip.item.Quality)
	r.name.Text = chipTitle(r.chip.item.Name, r.chip.count)

	canvas.Refresh(r.chip)
}

func (r *chipRenderer) Layout(size fyne.Size) {
	lineHeight := chipTextSize + 4

	r.background.Resize(size)

	r.dot.Move(fyne.NewPos(chipPadding, (size.Height-chipDotSize)/2))
	r.dot.Resize(fyne.NewSize(chipDotSize, chipDotSize))

	nameLeft := chipPadding + chipDotSize + chipDotGap
	r.name.Move(fyne.NewPos(nameLeft, (size.Height-lineHeight)/2))
	r.name.Resize(fyne.NewSize(size.Width-nameLeft-chipPadding, lineHeight))
}

func (r *chipRenderer) MinSize() fyne.Size {
	return r.chip.MinSize()
}

// chipTitle 标签文字：加过多件时带上件数。
func chipTitle(name string, count int) string {
	if count < 2 {
		return name
	}

	return fmt.Sprintf("%s ×%d", name, count)
}

// selectedChip 是已选择区里的标签：名称（×N）后面带一个叉。点叉移除该物品，
// 点标签其余部分再加一件。
type selectedChip struct {
	widget.BaseWidget

	page  *Bid
	index int
	item  bid.Item
	count int
}

// newSelectedChip 构建一个空标签，内容由 set 填充。
func newSelectedChip(page *Bid) *selectedChip {
	chip := &selectedChip{page: page}
	chip.ExtendBaseWidget(chip)

	return chip
}

// set 设置标签对应的物品与件数。
func (c *selectedChip) set(index int, item bid.Item, count int) {
	c.index = index
	c.item = item
	c.count = count
	c.Refresh()
}

// Tapped 点叉移除该物品，点其余部分再加一件。
func (c *selectedChip) Tapped(event *fyne.PointEvent) {
	if c.page == nil {
		return
	}

	if event.Position.X >= c.Size().Width-selectedCrossSize-selectedCrossPadding {
		c.page.removeAll(c.index)
		return
	}

	c.page.addItem(c.index)
}

// MinSize 返回标签尺寸：文字加上右侧的叉。
func (c *selectedChip) MinSize() fyne.Size {
	c.ExtendBaseWidget(c)

	width := c.labelWidth() + chipPadding*2 + chipDotSize + chipDotGap + selectedCrossSize + chipGap

	return fyne.NewSize(width, chipHeight)
}

// labelWidth 返回标签文字需要的宽度。
func (c *selectedChip) labelWidth() float32 {
	return fyne.MeasureText(chipTitle(c.item.Name, c.count), chipTextSize, fyne.TextStyle{}).Width
}

// CreateRenderer 创建标签的绘制对象。叉用两条对角线画，避开字体里有没有 ✕ 的问题。
func (c *selectedChip) CreateRenderer() fyne.WidgetRenderer {
	background := canvas.NewRectangle(color.Transparent)
	background.CornerRadius = chipHeight / 2

	dot := canvas.NewRectangle(color.Transparent)
	dot.CornerRadius = chipDotSize / 2

	name := canvas.NewText("", color.White)
	name.TextSize = chipTextSize

	crossA := canvas.NewLine(color.White)
	crossB := canvas.NewLine(color.White)
	crossA.StrokeWidth = 1.5
	crossB.StrokeWidth = 1.5

	renderer := &selectedChipRenderer{
		baseRenderer: baseRenderer{objects: []fyne.CanvasObject{background, dot, name, crossA, crossB}},
		chip:         c,
		background:   background,
		dot:          dot,
		name:         name,
		crossA:       crossA,
		crossB:       crossB,
	}
	renderer.Refresh()

	return renderer
}

type selectedChipRenderer struct {
	baseRenderer

	chip       *selectedChip
	background *canvas.Rectangle
	dot        *canvas.Rectangle
	name       *canvas.Text
	crossA     *canvas.Line
	crossB     *canvas.Line
}

func (r *selectedChipRenderer) Refresh() {
	th := r.chip.Theme()
	variant := fyne.CurrentApp().Settings().ThemeVariant()

	fill := th.Color(theme.ColorNamePrimary, variant)
	onFill := th.Color(theme.ColorNameForegroundOnPrimary, variant)

	r.background.FillColor = fill
	r.background.StrokeColor = fill

	r.dot.FillColor = qualityColor(r.chip.item.Quality)

	r.name.Color = onFill
	r.name.Text = chipTitle(r.chip.item.Name, r.chip.count)

	r.crossA.StrokeColor = onFill
	r.crossB.StrokeColor = onFill

	canvas.Refresh(r.chip)
}

func (r *selectedChipRenderer) Layout(size fyne.Size) {
	lineHeight := chipTextSize + 4

	r.background.Resize(size)

	r.dot.Move(fyne.NewPos(chipPadding, (size.Height-chipDotSize)/2))
	r.dot.Resize(fyne.NewSize(chipDotSize, chipDotSize))

	nameLeft := chipPadding + chipDotSize + chipDotGap
	crossLeft := size.Width - selectedCrossPadding - selectedCrossSize
	nameWidth := crossLeft - nameLeft - chipGap

	r.name.Move(fyne.NewPos(nameLeft, (size.Height-lineHeight)/2))
	r.name.Resize(fyne.NewSize(nameWidth, lineHeight))
	r.name.Text = fitText(chipTitle(r.chip.item.Name, r.chip.count), nameWidth, chipTextSize, fyne.TextStyle{})

	top := (size.Height - selectedCrossSize) / 2
	bottom := top + selectedCrossSize

	r.crossA.Position1 = fyne.NewPos(crossLeft, top)
	r.crossA.Position2 = fyne.NewPos(crossLeft+selectedCrossSize, bottom)
	r.crossB.Position1 = fyne.NewPos(crossLeft, bottom)
	r.crossB.Position2 = fyne.NewPos(crossLeft+selectedCrossSize, top)
}

func (r *selectedChipRenderer) MinSize() fyne.Size {
	return r.chip.MinSize()
}

// chipFlow 把标签按内容宽度横向排列，一行放不下就换到下一行。
type chipFlow struct {
	widget.BaseWidget

	chips []fyne.CanvasObject
	rows  int
}

// newChipFlow 构建一个空的标签区。
func newChipFlow() *chipFlow {
	flow := &chipFlow{rows: 1}
	flow.ExtendBaseWidget(flow)

	return flow
}

// setChips 用新的标签替换内容。
func (f *chipFlow) setChips(chips []fyne.CanvasObject) {
	f.chips = chips
	f.rows = 1
	f.Refresh()
}

// MinSize 返回当前行数下的整体尺寸。
func (f *chipFlow) MinSize() fyne.Size {
	f.ExtendBaseWidget(f)

	rows := f.rows
	if rows < 1 {
		rows = 1
	}

	width := chipMinWidth

	for _, child := range f.chips {
		if childWidth := child.MinSize().Width; childWidth > width {
			width = childWidth
		}
	}

	return fyne.NewSize(width, float32(rows)*chipHeight+float32(rows-1)*chipGap)
}

// CreateRenderer 创建标签区的绘制对象。
func (f *chipFlow) CreateRenderer() fyne.WidgetRenderer {
	return &chipFlowRenderer{baseRenderer: baseRenderer{}, flow: f}
}

type chipFlowRenderer struct {
	baseRenderer

	flow *chipFlow
}

func (r *chipFlowRenderer) Layout(size fyne.Size) {
	chips := r.flow.chips
	r.SetObjects(chips)

	x, y := float32(0), float32(0)
	rows := 1

	for _, child := range chips {
		width := child.MinSize().Width

		if x > 0 && x+chipGap+width > size.Width {
			x = 0
			y += chipHeight + chipGap
			rows++
		}

		child.Move(fyne.NewPos(x, y))
		child.Resize(fyne.NewSize(width, chipHeight))

		x += width + chipGap
	}

	if rows < 1 {
		rows = 1
	}

	if rows != r.flow.rows {
		r.flow.rows = rows
		// 行数变了，MinSize 跟着变，请父级（滚动容器）重新布局。
		canvas.Refresh(r.flow)
	}
}

func (r *chipFlowRenderer) Refresh() {
	r.SetObjects(r.flow.chips)

	// 标签宽度会随「×N」变化，尺寸已知时立刻重排一次。
	if size := r.flow.Size(); size.Width > 0 {
		r.Layout(size)
	}

	canvas.Refresh(r.flow)
}

func (r *chipFlowRenderer) MinSize() fyne.Size {
	return r.flow.MinSize()
}
