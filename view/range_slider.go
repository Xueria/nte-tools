package view

import (
	"image/color"
	"math"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// Compile time interface checks.
var (
	_ fyne.Draggable    = (*RangeSlider)(nil)
	_ fyne.Focusable    = (*RangeSlider)(nil)
	_ desktop.Hoverable = (*RangeSlider)(nil)
	_ fyne.Tappable     = (*RangeSlider)(nil)
	_ fyne.Disableable  = (*RangeSlider)(nil)
)

// RangeSlider is a horizontal slider with two draggable thumbs that select a
// range between Lower and Upper (both inclusive).
type RangeSlider struct {
	widget.BaseWidget

	Min   float64
	Max   float64
	Step  float64
	Lower float64
	Upper float64

	// OnChanged is called whenever Lower or Upper changes.
	OnChanged func(lower, upper float64)

	hovered       bool
	focused       bool
	disabled      bool
	keyboardThumb int // thumb controlled by keyboard: 0 = lower, 1 = upper
	dragging      int // thumb being dragged: -1 = none, 0 = lower, 1 = upper

	hoveredThumb int // thumb under the pointer: -1 = none, 0 = lower, 1 = upper
}

// NewRangeSlider creates a RangeSlider with the given bounds.
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

// SetValues updates both thumbs and clamps them into the valid range.
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

// SetRange updates the bounds and re-clamps the current selection.
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

// clamp snaps a value onto Step and keeps it inside [Min, Max].
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

// DragEnd clears the dragging state.
func (s *RangeSlider) DragEnd() {
	if s.disabled {
		return
	}
	s.dragging = -1
}

// Dragged moves the grabbed thumb (chosen on first event by proximity).
func (s *RangeSlider) Dragged(e *fyne.DragEvent) {
	if s.disabled {
		return
	}
	if s.dragging < 0 {
		s.dragging = s.nearestThumb(e.Position.X)
	}
	s.applyToThumb(s.dragging, e.Position.X)
}

// Tapped moves the closest thumb to the tapped position.
func (s *RangeSlider) Tapped(e *fyne.PointEvent) {
	if s.disabled {
		return
	}
	s.keyboardThumb = s.nearestThumb(e.Position.X)
	s.applyToThumb(s.keyboardThumb, e.Position.X)
}

// applyToThumb sets the given thumb to the value under the pointer and clamps
// the other thumb so Lower never exceeds Upper.
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

// valueFromPosition converts a horizontal pointer coordinate into a stepped value.
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

// nearestThumb returns 0 or 1 depending on which thumb is closest to x.
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

// FocusGained marks the widget focused.
func (s *RangeSlider) FocusGained() {
	s.focused = true
	if !s.disabled {
		s.Refresh()
	}
}

// FocusLost clears the focused flag.
func (s *RangeSlider) FocusLost() {
	s.focused = false
	if !s.disabled {
		s.Refresh()
	}
}

// MouseIn marks the widget hovered and picks the thumb under the pointer.
func (s *RangeSlider) MouseIn(e *desktop.MouseEvent) {
	s.hovered = true
	s.hoveredThumb = s.nearestThumb(e.Position.X)
	if !s.disabled {
		s.Refresh()
	}
}

// MouseMoved keeps the hover indicator glued to whichever thumb is closest to
// the pointer.
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

// MouseOut clears the hovered flag.
func (s *RangeSlider) MouseOut() {
	s.hovered = false
	s.hoveredThumb = -1
	if !s.disabled {
		s.Refresh()
	}
}

// TypedKey moves the keyboard thumb with the arrow keys.
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

// nudge moves a thumb by delta while respecting the other thumb's bound.
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

// TypedRune is a no-op hook for the focus interface.
func (s *RangeSlider) TypedRune(_ rune) {}

// Disable disables the widget.
func (s *RangeSlider) Disable() {
	if s.disabled {
		return
	}
	s.disabled = true
	s.Refresh()
}

// Enable enables the widget.
func (s *RangeSlider) Enable() {
	if !s.disabled {
		return
	}
	s.disabled = false
	s.Refresh()
}

// Disabled reports whether the widget is disabled.
func (s *RangeSlider) Disabled() bool {
	return s.disabled
}

// MinSize returns the minimum size of the widget.
func (s *RangeSlider) MinSize() fyne.Size {
	s.ExtendBaseWidget(s)
	return s.BaseWidget.MinSize()
}

// CreateRenderer builds the canvas objects for the widget.
func (s *RangeSlider) CreateRenderer() fyne.WidgetRenderer {
	s.ExtendBaseWidget(s)
	th := s.Theme()
	v := fyne.CurrentApp().Settings().ThemeVariant()

	track := canvas.NewRectangle(th.Color(theme.ColorNameInputBackground, v))
	activeRange := canvas.NewRectangle(th.Color(theme.ColorNamePrimary, v))
	lowerThumb := &canvas.Circle{FillColor: th.Color(theme.ColorNamePrimary, v)}
	upperThumb := &canvas.Circle{FillColor: th.Color(theme.ColorNamePrimary, v)}
	focusIndicator := &canvas.Circle{FillColor: color.Transparent}

	objects := []fyne.CanvasObject{track, activeRange, focusIndicator, lowerThumb, upperThumb}
	r := &rangeSliderRenderer{
		baseRenderer:   baseRenderer{objects: objects},
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

func (s *RangeSlider) almostEqual(a, b float64) bool {
	step := s.Step
	if step <= 0 {
		step = 1
	}
	return math.Abs(a-b) <= step/2
}

// buttonDiameter matches the built-in slider thumb size.
func (s *RangeSlider) buttonDiameter(inlineIconSize float32) float32 {
	return inlineIconSize - 4
}

// endOffset is the horizontal padding before/after the track.
func (s *RangeSlider) endOffset() float32 {
	th := s.Theme()
	return s.buttonDiameter(th.Size(theme.SizeNameInlineIcon))/2 +
		th.Size(theme.SizeNameInnerPadding) - 1.5
}

// positionOf returns the x position of a value within the track.
func (s *RangeSlider) positionOf(value float64, pad float32, size fyne.Size) float32 {
	if s.Max == s.Min {
		return pad
	}
	ratio := float32((value - s.Min) / (s.Max - s.Min))
	return pad + ratio*(size.Width-pad*2)
}

type rangeSliderRenderer struct {
	baseRenderer

	track          *canvas.Rectangle
	activeRange    *canvas.Rectangle
	lowerThumb     *canvas.Circle
	upperThumb     *canvas.Circle
	focusIndicator *canvas.Circle
	slider         *RangeSlider
}

// Refresh updates colours and layout.
func (r *rangeSliderRenderer) Refresh() {
	th := r.slider.Theme()
	v := fyne.CurrentApp().Settings().ThemeVariant()

	r.track.FillColor = th.Color(theme.ColorNameInputBackground, v)

	if r.slider.disabled {
		r.lowerThumb.FillColor = th.Color(theme.ColorNameDisabled, v)
		r.upperThumb.FillColor = th.Color(theme.ColorNameDisabled, v)
		r.activeRange.FillColor = th.Color(theme.ColorNameDisabled, v)
	} else {
		r.lowerThumb.FillColor = th.Color(theme.ColorNamePrimary, v)
		r.upperThumb.FillColor = th.Color(theme.ColorNamePrimary, v)
		r.activeRange.FillColor = th.Color(theme.ColorNamePrimary, v)
	}

	if r.slider.focused && !r.slider.disabled {
		r.focusIndicator.FillColor = th.Color(theme.ColorNameFocus, v)
	} else if r.slider.hovered && !r.slider.disabled {
		r.focusIndicator.FillColor = th.Color(theme.ColorNameHover, v)
	} else {
		r.focusIndicator.FillColor = color.Transparent
	}
	r.focusIndicator.Refresh()

	r.Layout(r.slider.Size())
	canvas.Refresh(r.slider)
}

// Layout positions the track, active range and thumbs.
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

	// Track
	trackPos := fyne.NewPos(pad, size.Height/2-trackWidth/2)
	trackSize := fyne.NewSize(size.Width-pad*2, trackWidth)
	if trackSize.Width < 0 {
		trackSize.Width = 0
	}
	r.track.Move(trackPos)
	r.track.Resize(trackSize)

	// Active range between the two thumbs
	lowerPos := r.slider.positionOf(r.slider.Lower, pad, size)
	upperPos := r.slider.positionOf(r.slider.Upper, pad, size)
	activeRangePos := fyne.NewPos(lowerPos, trackPos.Y)
	r.activeRange.Move(activeRangePos)
	r.activeRange.Resize(fyne.NewSize(upperPos-lowerPos, trackWidth))

	// Thumbs
	thumbY := trackPos.Y - (diameter-trackSize.Height)/2
	r.lowerThumb.Move(fyne.NewPos(lowerPos-diameter/2, thumbY))
	r.lowerThumb.Resize(fyne.NewSize(diameter, diameter))
	r.upperThumb.Move(fyne.NewPos(upperPos-diameter/2, thumbY))
	r.upperThumb.Resize(fyne.NewSize(diameter, diameter))

	// Focus indicator follows the hovered thumb, falling back to the keyboard
	// thumb when not hovering.
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

// MinSize calculates the minimum size of the widget.
func (r *rangeSliderRenderer) MinSize() fyne.Size {
	th := r.slider.Theme()
	pad := th.Size(theme.SizeNameInnerPadding)
	tap := th.Size(theme.SizeNameInlineIcon)
	dia := r.slider.buttonDiameter(tap)

	return fyne.NewSize(minLongSide+dia, tap+pad*2)
}

const minLongSide = float32(34)
