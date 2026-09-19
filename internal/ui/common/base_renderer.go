// Package common 收纳界面层可复用的部分：自绘控件的渲染器基座、通用控件、数值
// 格式化、品质配色与 MD3 主题。这里不放业务语义，各页面自己的表现留在各自的包。
package common

import "fyne.io/fyne/v2"

// BaseRenderer 实现 fyne.WidgetRenderer 里样板化的 Objects 与 Destroy，相当于
// Fyne 内部的 BaseRenderer；各页面的自绘控件靠嵌入它实现渲染器。
type BaseRenderer struct {
	objects []fyne.CanvasObject
}

// NewBaseRenderer 用给定的子对象构造渲染器基座。
func NewBaseRenderer(objects ...fyne.CanvasObject) BaseRenderer {
	return BaseRenderer{objects: objects}
}

// Destroy 是空实现。
func (r *BaseRenderer) Destroy() {}

// Objects 返回需要绘制的子对象。
func (r *BaseRenderer) Objects() []fyne.CanvasObject {
	return r.objects
}

// SetObjects 替换子对象。
func (r *BaseRenderer) SetObjects(objects []fyne.CanvasObject) {
	r.objects = objects
}
