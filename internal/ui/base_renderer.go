package ui

import "fyne.io/fyne/v2"

// baseRenderer 实现 fyne.WidgetRenderer 里样板化的 Objects 与 Destroy，相当于
// Fyne 内部的 BaseRenderer；本包的自绘控件靠它实现渲染器，无需引入内部包。
type baseRenderer struct {
	objects []fyne.CanvasObject
}

// Destroy 是空实现。
func (r *baseRenderer) Destroy() {}

// Objects 返回需要绘制的子对象。
func (r *baseRenderer) Objects() []fyne.CanvasObject {
	return r.objects
}

// SetObjects 替换子对象。
func (r *baseRenderer) SetObjects(objects []fyne.CanvasObject) {
	r.objects = objects
}
