package planner

import (
	"fmt"
	"image/color"

	"blind-tools/internal/pool"
	"blind-tools/internal/ui/common"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// poolItem 是盲盒列表里的一行：名称与「抽数 · 资源种数」。
type poolItem struct {
	widget.BaseWidget

	title    string
	subtitle string
}

// newPoolItem 构建一行空条目，内容由 set 填充。
func newPoolItem() fyne.CanvasObject {
	item := &poolItem{}
	item.ExtendBaseWidget(item)

	return item
}

// set 用一条盲盒池数据填充该行。
func (item *poolItem) set(p pool.Pool) {
	item.title = p.Manifest.Name
	item.subtitle = fmt.Sprintf("%d 抽 · %d 种资源", p.Manifest.Draws, len(p.Resources))
	item.Refresh()
}

// MinSize 返回该行的建议最小尺寸。
func (item *poolItem) MinSize() fyne.Size {
	item.ExtendBaseWidget(item)

	return item.BaseWidget.MinSize()
}

// CreateRenderer 创建该行的绘制对象。
func (item *poolItem) CreateRenderer() fyne.WidgetRenderer {
	title := canvas.NewText(item.title, color.White)
	title.TextStyle = fyne.TextStyle{Bold: true}

	subtitle := canvas.NewText(item.subtitle, color.White)

	renderer := &poolItemRenderer{
		BaseRenderer: common.NewBaseRenderer(title, subtitle),
		title:        title,
		subtitle:     subtitle,
		item:         item,
	}
	renderer.Refresh()

	return renderer
}

type poolItemRenderer struct {
	common.BaseRenderer

	title    *canvas.Text
	subtitle *canvas.Text
	item     *poolItem
}

func (r *poolItemRenderer) Refresh() {
	th := r.item.Theme()
	variant := fyne.CurrentApp().Settings().ThemeVariant()

	r.title.Text = r.item.title
	r.title.Color = th.Color(theme.ColorNameForeground, variant)
	r.subtitle.Text = r.item.subtitle
	r.subtitle.Color = th.Color(theme.ColorNamePlaceHolder, variant)

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
