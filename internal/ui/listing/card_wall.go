package listing

import (
	"blind-tools/internal/ui/common"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"
)

const (
	// cardMinWidth 卡片的最小宽度：列数由可用宽度能放下多少个它决定。
	cardMinWidth float32 = 110
	// cardMinGap、cardMaxGap 卡片间距的下限与上限：列间剩余宽度会动态分到间距上，
	// 超过上限的部分改为把卡片均分撑满，右边缘始终贴齐。
	cardMinGap float32 = 6
	cardMaxGap float32 = 12
)

// cardWall 把卡片按可用宽度分列摆放：列数随宽度自适应，列间剩余宽度动态分到
// 间距上，因此右侧不会留下空白。
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
	return &cardWallRenderer{BaseRenderer: common.NewBaseRenderer(), wall: w}
}

type cardWallRenderer struct {
	common.BaseRenderer

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

// cardMetrics 返回卡片宽度与列间距：剩余宽度先分给间距，超过上限后改为把卡片
// 均分撑满，保证最右侧与左边缘一样贴齐。
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
