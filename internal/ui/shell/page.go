package shell

import "fyne.io/fyne/v2"

// Page 是外壳需要的一页：导航栏里的名称与图标，以及页面内容。
// 接口定义在消费方（外壳），各页面只要实现这三个方法就能挂上来。
type Page interface {
	// Title 是该页在导航栏里的名称。
	Title() string
	// Icon 是该页在导航栏里的图标。
	Icon() fyne.Resource
	// Content 是该页的内容。
	Content() fyne.CanvasObject
}

// pinnedPage 是包了一层的「固定项」：外壳据此把它排到导航栏底部，
// 和功能页分开，例如设置。其余方法由内嵌的 Page 提供。
type pinnedPage struct {
	Page
}

// Pinned 把一页标记为固定在导航栏底部。
func Pinned(page Page) Page {
	return pinnedPage{Page: page}
}

// isPinned 判断一页是否被标记为固定项。
func isPinned(page Page) bool {
	marker, ok := page.(interface{ pinned() bool })

	return ok && marker.pinned()
}

// pinned 只为外壳内部的类型断言服务。
func (pinnedPage) pinned() bool {
	return true
}
