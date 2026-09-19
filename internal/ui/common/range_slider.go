package common

import (
	"image/color"
	"math"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// 编译期接口检查。
var (
	_ fyne.Draggable    = (*RangeSlider)(nil)
	_ fyne.Focusable    = (*RangeSlider)(nil)
	_ desktop.Hoverable = (*RangeSlider)(nil)
	_ fyne.Tappable     = (*RangeSlider)(nil)
	_ fyne.Disableable  = (*RangeSlider)(nil)
)

// minLongSide 滑块最小尺寸里长边的下限。
const minLongSide = float32(34)

// RangeSlider 是带两个可拖动滑块的横向滑块，用来选择 [Lower, Upper] 区间（含两端）。
type RangeSlider struct {
	widget.BaseWidget

	Min   float64
	Max   float64
	Step  float64
	Lower float64
	Upper float64

	// OnChanged 在 Lower 或 Upper 变化时调用。
	OnChanged func(lower, upper float64)

	hovered       bool
	focused       bool
	disabled      bool
	keyboardThumb int // 键盘控制的滑块：0 = 下端，1 = 上端
	dragging      int // 正在拖动的滑块：-1 = 无，0 = 下端，1 = 上端

	hoveredThumb int // 指针下的滑块：-1 = 无，0 = 下端，1 = 上端
}

// NewRangeSlider 按给定的上下界创建一个区间滑块。
func NewRangeSlider(min, max float64) *RangeSlider {
	s := &RangeSlider{
		Min:      min,
		Max:      max,
		Step:     1,
		Lower:    min,
		Upper:    max,
		dragging: -1,
	}
	s.ExtendBaseWidget(s)

	return s
}

// SetValues 同时设置两个滑块，并把它们夹回合法区间。
func (s *RangeSlider) SetValues(lower, upper float64) {
	s.Lower = s.clamp(lower)
	s.Upper = s.clamp(upper)

	if s.Lower > s.Upper {
		s.Lower, s.Upper = s.Upper, s.Lower
	}

	s.Refresh()

	if s.OnChanged != nil {
		s.OnChanged(s.Lower, s.Upper)
	}
}

// SetRange 更新上下界，并重新夹取当前选择。
func (s *RangeSlider) SetRange(min, max float64) {
	if min > max {
		min, max = max, min
	}

	s.Min = min
	s.Max = max
	s.Lower = s.clamp(s.Lower)
	s.Upper = s.clamp(s.Upper)
	s.Refresh()
}

// clamp 把值对齐到 Step 并限制在 [Min, Max] 内。
func (s *RangeSlider) clamp(value float64) float64 {
	if value <= s.Min {
		return s.Min
	}

	if value >= s.Max {
		return s.Max
	}

	if s.Step <= 0 {
		return value
	}

	return math.Round(value/s.Step) * s.Step
}

// DragEnd 清除拖动状态。
func (s *RangeSlider) DragEnd() {
	if s.disabled {
		return
	}

	s.dragging = -1
}

// Dragged 移动被抓住的滑块，抓住哪个由第一次事件的位置决定。
func (s *RangeSlider) Dragged(e *fyne.DragEvent) {
	if s.disabled {
		return
	}

	if s.dragging < 0 {
		s.dragging = s.nearestThumb(e.Position.X)
	}

	s.applyToThumb(s.dragging, e.Position.X)
}

// Tapped 把离点击位置最近的滑块移过去。
func (s *RangeSlider) Tapped(e *fyne.PointEvent) {
	if s.disabled {
		return
	}

	s.keyboardThumb = s.nearestThumb(e.Position.X)
	s.applyToThumb(s.keyboardThumb, e.Position.X)
}

// applyToThumb 把指定滑块移到指针处的值，并夹住另一个滑块，保证 Lower 不超过 Upper。
func (s *RangeSlider) applyToThumb(thumb int, x float32) {
	value := s.valueFromPosition(x)
	lower, upper := s.Lower, s.Upper

	if thumb == 0 {
		lower = value

		if lower > upper {
			lower = upper
		}
	} else {
		upper = value

		if upper < lower {
			upper = lower
		}
	}

	if s.almostEqual(lower, s.Lower) && s.almostEqual(upper, s.Upper) {
		return
	}

	s.Lower = lower
	s.Upper = upper
	s.Refresh()

	if s.OnChanged != nil {
		s.OnChanged(s.Lower, s.Upper)
	}
}

// valueFromPosition 把横向坐标换算成对齐过 Step 的值。
func (s *RangeSlider) valueFromPosition(x float32) float64 {
	pad := s.endOffset()
	size := s.Size()

	if x <= pad {
		return s.Min
	}

	if x >= size.Width-pad {
		return s.Max
	}

	if size.Width <= pad*2 {
		return s.Min
	}

	ratio := float64(x-pad) / float64(size.Width-pad*2)

	return s.clamp(s.Min + ratio*(s.Max-s.Min))
}

// nearestThumb 返回离 x 更近的滑块：0 或 1。
func (s *RangeSlider) nearestThumb(x float32) int {
	pad := s.endOffset()
	size := s.Size()
	lowerPos := s.positionOf(s.Lower, pad, size)
	upperPos := s.positionOf(s.Upper, pad, size)

	if math.Abs(float64(x-lowerPos)) <= math.Abs(float64(x-upperPos)) {
		return 0
	}

	return 1
}

// FocusGained 标记控件获得焦点。
func (s *RangeSlider) FocusGained() {
	s.focused = true

	if !s.disabled {
		s.Refresh()
	}
}

// FocusLost 清除焦点标记。
func (s *RangeSlider) FocusLost() {
	s.focused = false

	if !s.disabled {
		s.Refresh()
	}
}

// MouseIn 标记指针进入，并记下指针下的滑块。
func (s *RangeSlider) MouseIn(e *desktop.MouseEvent) {
	s.hovered = true
	s.hoveredThumb = s.nearestThumb(e.Position.X)

	if !s.disabled {
		s.Refresh()
	}
}

// MouseMoved 让悬停指示始终贴住离指针最近的滑块。
func (s *RangeSlider) MouseMoved(e *desktop.MouseEvent) {
	if s.disabled {
		return
	}

	thumb := s.nearestThumb(e.Position.X)

	if thumb != s.hoveredThumb {
		s.hoveredThumb = thumb
		s.Refresh()
	}
}

// MouseOut 清除悬停标记。
func (s *RangeSlider) MouseOut() {
	s.hovered = false
	s.hoveredThumb = -1

	if !s.disabled {
		s.Refresh()
	}
}

// TypedKey 用方向键移动键盘控制的滑块。
func (s *RangeSlider) TypedKey(key *fyne.KeyEvent) {
	if s.disabled {
		return
	}

	switch key.Name {
	case fyne.KeyLeft:
		s.nudge(s.keyboardThumb, -s.Step)
	case fyne.KeyRight:
		s.nudge(s.keyboardThumb, s.Step)
	case fyne.KeyHome:
		s.nudge(s.keyboardThumb, s.Min-s.Lower)
	case fyne.KeyEnd:
		s.nudge(s.keyboardThumb, s.Max-s.Lower)
	}
}

// nudge 移动一个滑块，同时不越过另一个滑块。
func (s *RangeSlider) nudge(thumb int, delta float64) {
	lower, upper := s.Lower, s.Upper

	if thumb == 0 {
		lower = s.clamp(lower + delta)

		if lower > upper {
			lower = upper
		}
	} else {
		upper = s.clamp(upper + delta)

		if upper < lower {
			upper = lower
		}
	}

	if s.almostEqual(lower, s.Lower) && s.almostEqual(upper, s.Upper) {
		return
	}

	s.Lower = lower
	s.Upper = upper
	s.Refresh()

	if s.OnChanged != nil {
		s.OnChanged(s.Lower, s.Upper)
	}
}

// TypedRune 是焦点接口要求的空实现。
func (s *RangeSlider) TypedRune(_ rune) {}

// Disable 禁用控件。
func (s *RangeSlider) Disable() {
	if s.disabled {
		return
	}

	s.disabled = true
	s.Refresh()
}

// Enable 启用控件。
func (s *RangeSlider) Enable() {
	if !s.disabled {
		return
	}

	s.disabled = false
	s.Refresh()
}

// Disabled 返回控件是否已禁用。
func (s *RangeSlider) Disabled() bool {
	return s.disabled
}

// MinSize 返回控件的最小尺寸。
func (s *RangeSlider) MinSize() fyne.Size {
	s.ExtendBaseWidget(s)

	return s.BaseWidget.MinSize()
}

// CreateRenderer 创建控件的绘制对象。
func (s *RangeSlider) CreateRenderer() fyne.WidgetRenderer {
	s.ExtendBaseWidget(s)
	th := s.Theme()
	variant := fyne.CurrentApp().Settings().ThemeVariant()

	track := canvas.NewRectangle(th.Color(theme.ColorNameInputBackground, variant))
	activeRange := canvas.NewRectangle(th.Color(theme.ColorNamePrimary, variant))
	lowerThumb := &canvas.Circle{FillColor: th.Color(theme.ColorNamePrimary, variant)}
	upperThumb := &canvas.Circle{FillColor: th.Color(theme.ColorNamePrimary, variant)}
	focusIndicator := &canvas.Circle{FillColor: color.Transparent}

	objects := []fyne.CanvasObject{track, activeRange, focusIndicator, lowerThumb, upperThumb}
	r := &rangeSliderRenderer{
		BaseRenderer:   NewBaseRenderer(objects...),
		track:          track,
		activeRange:    activeRange,
		lowerThumb:     lowerThumb,
		upperThumb:     upperThumb,
		focusIndicator: focusIndicator,
		slider:         s,
	}
	r.Refresh()

	return r
}

// almostEqual 判断两个值在当前步长下是否可视为相等。
func (s *RangeSlider) almostEqual(a, b float64) bool {
	step := s.Step

	if step <= 0 {
		step = 1
	}

	return math.Abs(a-b) <= step/2
}

// buttonDiameter 与内置滑块的圆点直径保持一致。
func (s *RangeSlider) buttonDiameter(inlineIconSize float32) float32 {
	return inlineIconSize - 4
}

// endOffset 是轨道两端预留的横向留白。
func (s *RangeSlider) endOffset() float32 {
	th := s.Theme()

	return s.buttonDiameter(th.Size(theme.SizeNameInlineIcon))/2 +
		th.Size(theme.SizeNameInnerPadding) - 1.5
}

// positionOf 返回值在轨道上的横坐标。
func (s *RangeSlider) positionOf(value float64, pad float32, size fyne.Size) float32 {
	if s.Max == s.Min {
		return pad
	}

	ratio := float32((value - s.Min) / (s.Max - s.Min))

	return pad + ratio*(size.Width-pad*2)
}

type rangeSliderRenderer struct {
	BaseRenderer

	track          *canvas.Rectangle
	activeRange    *canvas.Rectangle
	lowerThumb     *canvas.Circle
	upperThumb     *canvas.Circle
	focusIndicator *canvas.Circle
	slider         *RangeSlider
}

// Refresh 更新颜色与布局。
func (r *rangeSliderRenderer) Refresh() {
	th := r.slider.Theme()
	variant := fyne.CurrentApp().Settings().ThemeVariant()

	r.track.FillColor = th.Color(theme.ColorNameInputBackground, variant)

	if r.slider.disabled {
		r.lowerThumb.FillColor = th.Color(theme.ColorNameDisabled, variant)
		r.upperThumb.FillColor = th.Color(theme.ColorNameDisabled, variant)
		r.activeRange.FillColor = th.Color(theme.ColorNameDisabled, variant)
	} else {
		r.lowerThumb.FillColor = th.Color(theme.ColorNamePrimary, variant)
		r.upperThumb.FillColor = th.Color(theme.ColorNamePrimary, variant)
		r.activeRange.FillColor = th.Color(theme.ColorNamePrimary, variant)
	}

	if r.slider.focused && !r.slider.disabled {
		r.focusIndicator.FillColor = th.Color(theme.ColorNameFocus, variant)
	} else if r.slider.hovered && !r.slider.disabled {
		r.focusIndicator.FillColor = th.Color(theme.ColorNameHover, variant)
	} else {
		r.focusIndicator.FillColor = color.Transparent
	}

	r.focusIndicator.Refresh()

	r.Layout(r.slider.Size())
	canvas.Refresh(r.slider)
}

// Layout 摆放轨道、已选区间与两个滑块。
func (r *rangeSliderRenderer) Layout(size fyne.Size) {
	th := r.slider.Theme()
	inputBorderSize := th.Size(theme.SizeNameInputBorder)
	trackWidth := inputBorderSize * 2

	if trackWidth < 2 {
		trackWidth = 2
	}

	inlineIconSize := th.Size(theme.SizeNameInlineIcon)
	diameter := r.slider.buttonDiameter(inlineIconSize)
	pad := r.slider.endOffset()

	// 轨道
	trackPos := fyne.NewPos(pad, size.Height/2-trackWidth/2)
	trackSize := fyne.NewSize(size.Width-pad*2, trackWidth)

	if trackSize.Width < 0 {
		trackSize.Width = 0
	}

	r.track.Move(trackPos)
	r.track.Resize(trackSize)

	// 两个滑块之间的已选区间
	lowerPos := r.slider.positionOf(r.slider.Lower, pad, size)
	upperPos := r.slider.positionOf(r.slider.Upper, pad, size)
	r.activeRange.Move(fyne.NewPos(lowerPos, trackPos.Y))
	r.activeRange.Resize(fyne.NewSize(upperPos-lowerPos, trackWidth))

	// 滑块
	thumbY := trackPos.Y - (diameter-trackSize.Height)/2
	r.lowerThumb.Move(fyne.NewPos(lowerPos-diameter/2, thumbY))
	r.lowerThumb.Resize(fyne.NewSize(diameter, diameter))
	r.upperThumb.Move(fyne.NewPos(upperPos-diameter/2, thumbY))
	r.upperThumb.Resize(fyne.NewSize(diameter, diameter))

	// 焦点指示跟随悬停的滑块，没悬停时退回键盘控制的滑块。
	thumb := r.slider.keyboardThumb

	if r.slider.hovered && r.slider.hoveredThumb >= 0 {
		thumb = r.slider.hoveredThumb
	}

	focusX := lowerPos

	if thumb == 1 {
		focusX = upperPos
	}

	focusSize := fyne.NewSquareSize(inlineIconSize + th.Size(theme.SizeNameInnerPadding))
	delta := (focusSize.Width - diameter) / 2
	r.focusIndicator.Resize(focusSize)
	r.focusIndicator.Move(fyne.NewPos(focusX-focusSize.Width/2, thumbY-delta))
}

// MinSize 计算控件的最小尺寸。
func (r *rangeSliderRenderer) MinSize() fyne.Size {
	th := r.slider.Theme()
	pad := th.Size(theme.SizeNameInnerPadding)
	tap := th.Size(theme.SizeNameInlineIcon)
	dia := r.slider.buttonDiameter(tap)

	return fyne.NewSize(minLongSide+dia, tap+pad*2)
}
