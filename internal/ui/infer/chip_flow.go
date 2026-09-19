package infer

import (
	"blind-tools/internal/ui/common"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

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
	return &chipFlowRenderer{BaseRenderer: common.NewBaseRenderer(), flow: f}
}

type chipFlowRenderer struct {
	common.BaseRenderer

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

// flowScroll 把标签流放进滚动容器里，右侧留出滚动条的宽度。
func flowScroll(flow *chipFlow) fyne.CanvasObject {
	content := container.New(layout.NewCustomPaddedLayout(0, 0, 0, common.ScrollBarInset()), flow)

	return container.NewVScroll(content)
}
