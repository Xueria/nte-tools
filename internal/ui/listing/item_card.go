package listing

import (
	"image/color"

	"blind-tools/internal/bid"
	"blind-tools/internal/ui/common"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

const (
	// itemCardHeight 卡片高度；卡片内三行自上而下按固定行高排布。
	itemCardHeight float32 = 62
	// cardQualityRowHeight、cardRowHeight 卡片内品质行与文字行的高度。
	cardQualityRowHeight float32 = 14
	cardRowHeight        float32 = 16
	// cardNameTextSize、cardQualityTextSize 卡片内的字号。
	cardNameTextSize    float32 = 12
	cardQualityTextSize float32 = 11
)

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

	renderer := &itemCardRenderer{
		BaseRenderer: common.NewBaseRenderer(background, swatch, quality, name, value),
		card:         c,
		background:   background,
		swatch:       swatch,
		quality:      quality,
		name:         name,
		value:        value,
	}
	renderer.Refresh()

	return renderer
}

type itemCardRenderer struct {
	common.BaseRenderer

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

	r.swatch.FillColor = common.QualityColor(r.card.item.Quality)

	r.quality.Text = common.QualityLabel(r.card.item.Quality)
	r.quality.Color = common.QualityColor(r.card.item.Quality)

	r.name.Color = th.Color(theme.ColorNameForeground, variant)

	r.value.Text = common.FormatValue(r.card.item.Value)
	r.value.Color = th.Color(theme.ColorNamePrimary, variant)

	r.fitName(r.card.Size().Width - common.CardInset*2)

	canvas.Refresh(r.card)
}

func (r *itemCardRenderer) Layout(size fyne.Size) {
	textWidth := size.Width - common.CardInset*2

	r.background.Resize(size)

	r.swatch.Move(fyne.NewPos(common.CardInset, common.CardInset+(cardQualityRowHeight-common.QualitySwatchSize)/2))
	r.swatch.Resize(fyne.NewSize(common.QualitySwatchSize, common.QualitySwatchSize))

	qualityLeft := common.CardInset + common.QualitySwatchSize + 4
	r.quality.Move(fyne.NewPos(qualityLeft, common.CardInset))
	r.quality.Resize(fyne.NewSize(textWidth-common.QualitySwatchSize-4, cardQualityRowHeight))

	nameTop := common.CardInset + cardQualityRowHeight + 2
	r.name.Move(fyne.NewPos(common.CardInset, nameTop))
	r.name.Resize(fyne.NewSize(textWidth, cardRowHeight))

	r.value.Move(fyne.NewPos(common.CardInset, nameTop+cardRowHeight+2))
	r.value.Resize(fyne.NewSize(textWidth, cardRowHeight))

	r.fitName(textWidth)
}

func (r *itemCardRenderer) MinSize() fyne.Size {
	return fyne.NewSize(cardMinWidth, itemCardHeight)
}

// fitName 按可用宽度截断名称，超出部分用省略号。
func (r *itemCardRenderer) fitName(width float32) {
	r.name.Text = common.FitText(r.card.item.Name, width, cardNameTextSize, fyne.TextStyle{})
}
