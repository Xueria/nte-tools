package shell

import (
	"image/color"

	"nte-tools/internal/ui/common"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

const (
	// railWidth 导航栏宽度：紧凑排版，只够放图标与短名称。
	railWidth float32 = 68
	// navItemHeight、navItemGap 单个导航项的高度与项间距。
	navItemHeight float32 = 50
	navItemGap    float32 = 0
	// navPillWidth、navPillHeight 选中/悬停时图标底下的胶囊尺寸。
	navPillWidth  float32 = 44
	navPillHeight float32 = 26
	// navIconSize、navLabelSize 图标边长与名称字号。
	navIconSize  float32 = 18
	navLabelSize float32 = 10
	// navLabelGap 胶囊与名称之间的距离。
	navLabelGap float32 = 3
	// navDividerInset 分组分隔线左右两侧的留白。
	navDividerInset float32 = 12
)

// navRail 是左侧导航栏：功能页自上而下排，固定项（例如设置，见 Pinned）贴底排，
// 两者之间留一条细分隔线。当前页高亮，点一下切页。
type navRail struct {
	widget.BaseWidget

	features []*navItem
	pinned   []*navItem
	buttons  []*navItem // 与页列表同序，便于按页下标定位
	selected int

	// OnSelect 由外壳赋值：点某一项时触发。
	OnSelect func(index int)
}

// newNavRail 按页列表构建导航栏，默认选中第一页。
func newNavRail(pages []Page) *navRail {
	rail := &navRail{}
	rail.ExtendBaseWidget(rail)

	for index, page := range pages {
		button := newNavItem(index, page, rail)
		rail.buttons = append(rail.buttons, button)

		if isPinned(page) {
			rail.pinned = append(rail.pinned, button)

			continue
		}

		rail.features = append(rail.features, button)
	}

	return rail
}

// selectItem 选中第 index 项，并通知外壳。
func (rail *navRail) selectItem(index int) {
	if index < 0 || index >= len(rail.buttons) {
		return
	}

	rail.setSelected(index)

	if rail.OnSelect != nil {
		rail.OnSelect(index)
	}
}

// setSelected 只更新高亮，不通知外壳。
func (rail *navRail) setSelected(index int) {
	if rail.selected == index {
		return
	}

	rail.selected = index

	for _, button := range rail.buttons {
		button.Refresh()
	}
}

// MinSize 返回导航栏的固定宽度与所需高度。
func (rail *navRail) MinSize() fyne.Size {
	rail.ExtendBaseWidget(rail)

	height := navItemHeight*float32(len(rail.buttons)) + navItemGap

	return fyne.NewSize(railWidth, height)
}

// CreateRenderer 创建导航栏的绘制对象。
func (rail *navRail) CreateRenderer() fyne.WidgetRenderer {
	return &navRailRenderer{
		BaseRenderer: common.NewBaseRenderer(),
		rail:         rail,
		divider:      canvas.NewRectangle(color.Transparent),
	}
}

type navRailRenderer struct {
	common.BaseRenderer

	rail    *navRail
	divider *canvas.Rectangle
}

func (r *navRailRenderer) Layout(size fyne.Size) {
	objects := make([]fyne.CanvasObject, 0, len(r.rail.buttons)+1)

	y := navItemGap
	for _, button := range r.rail.features {
		button.Move(fyne.NewPos(0, y))
		button.Resize(fyne.NewSize(size.Width, navItemHeight))
		objects = append(objects, button)

		y += navItemHeight + navItemGap
	}

	// 固定项从底部往上排。
	bottom := size.Height - navItemGap
	for i := len(r.rail.pinned) - 1; i >= 0; i-- {
		button := r.rail.pinned[i]
		bottom -= navItemHeight
		button.Move(fyne.NewPos(0, bottom))
		button.Resize(fyne.NewSize(size.Width, navItemHeight))
		objects = append(objects, button)

		bottom -= navItemGap
	}

	if len(r.rail.pinned) > 0 {
		r.divider.Move(fyne.NewPos(navDividerInset, bottom+navItemGap-1))
		r.divider.Resize(fyne.NewSize(size.Width-2*navDividerInset, 1))
		objects = append(objects, r.divider)
	}

	r.SetObjects(objects)
}

func (r *navRailRenderer) Refresh() {
	th, variant := themeColors()
	r.divider.FillColor = th.Color(theme.ColorNameSeparator, variant)

	if size := r.rail.Size(); size.Width > 0 {
		r.Layout(size)
	}

	canvas.Refresh(r.rail)
}

func (r *navRailRenderer) MinSize() fyne.Size {
	return r.rail.MinSize()
}

// navItem 导航栏里的一项：胶囊底 + 图标 + 名称。
type navItem struct {
	widget.BaseWidget

	index   int
	page    Page
	rail    *navRail
	hovered bool
}

// newNavItem 构建一项。
func newNavItem(index int, page Page, rail *navRail) *navItem {
	item := &navItem{index: index, page: page, rail: rail}
	item.ExtendBaseWidget(item)

	return item
}

// Tapped 点一下切到该页。
func (item *navItem) Tapped(*fyne.PointEvent) {
	if item.rail != nil {
		item.rail.selectItem(item.index)
	}
}

// MouseIn 记录悬停。
func (item *navItem) MouseIn(*desktop.MouseEvent) {
	item.hovered = true
	item.Refresh()
}

// MouseMoved 满足悬停接口，无需额外处理。
func (item *navItem) MouseMoved(*desktop.MouseEvent) {}

// MouseOut 取消悬停。
func (item *navItem) MouseOut() {
	item.hovered = false
	item.Refresh()
}

// MinSize 返回单项尺寸。
func (item *navItem) MinSize() fyne.Size {
	item.ExtendBaseWidget(item)

	return fyne.NewSize(railWidth, navItemHeight)
}

// CreateRenderer 创建该项的绘制对象。
func (item *navItem) CreateRenderer() fyne.WidgetRenderer {
	background := canvas.NewRectangle(color.Transparent)
	background.CornerRadius = navPillHeight / 2

	icon := canvas.NewImageFromResource(item.page.Icon())
	icon.FillMode = canvas.ImageFillContain

	label := canvas.NewText(item.page.Title(), color.White)
	label.TextSize = navLabelSize
	label.Alignment = fyne.TextAlignCenter

	renderer := &navItemRenderer{
		BaseRenderer: common.NewBaseRenderer(background, icon, label),
		item:         item,
		background:   background,
		icon:         icon,
		label:        label,
	}
	renderer.Refresh()

	return renderer
}

type navItemRenderer struct {
	common.BaseRenderer

	item       *navItem
	background *canvas.Rectangle
	icon       *canvas.Image
	label      *canvas.Text
}

func (r *navItemRenderer) Refresh() {
	th, variant := themeColors()
	selected := r.item.rail != nil && r.item.rail.selected == r.item.index

	switch {
	case selected:
		r.background.FillColor = th.Color(theme.ColorNameSelection, variant)
		r.label.Color = th.Color(theme.ColorNamePrimary, variant)
	case r.item.hovered:
		r.background.FillColor = th.Color(theme.ColorNameHover, variant)
		r.label.Color = th.Color(theme.ColorNameForeground, variant)
	default:
		r.background.FillColor = color.Transparent
		r.label.Color = th.Color(theme.ColorNameForeground, variant)
	}

	r.label.Text = r.item.page.Title()
	r.icon.Resource = r.item.page.Icon()
	r.icon.Refresh()

	canvas.Refresh(r.item)
}

func (r *navItemRenderer) Layout(size fyne.Size) {
	contentHeight := navPillHeight + navLabelGap + navLabelSize
	pillTop := (size.Height - contentHeight) / 2
	pillLeft := (size.Width - navPillWidth) / 2

	r.background.Move(fyne.NewPos(pillLeft, pillTop))
	r.background.Resize(fyne.NewSize(navPillWidth, navPillHeight))

	iconTop := pillTop + (navPillHeight-navIconSize)/2
	r.icon.Move(fyne.NewPos((size.Width-navIconSize)/2, iconTop))
	r.icon.Resize(fyne.NewSize(navIconSize, navIconSize))

	r.label.Move(fyne.NewPos(0, pillTop+navPillHeight+navLabelGap))
	r.label.Resize(fyne.NewSize(size.Width, navLabelSize+4))
}

func (r *navItemRenderer) MinSize() fyne.Size {
	return r.item.MinSize()
}
