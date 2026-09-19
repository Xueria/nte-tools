package ui

import "fyne.io/fyne/v2/theme"

// 本文件集中放置被多个 UI 组件共用的尺寸，避免它们散落在各自的页面文件里。

const (
	// cardInset 卡片类内容（拍品卡片、组合行）与卡片边缘的距离。
	cardInset float32 = 6
	// compositionTextSize 组合行与件数标签的字号。
	compositionTextSize float32 = 12
)

// scrollBarInset 返回浮动滚动条的宽度加一点余量：Fyne 的滚动条画在内容之上，
// 不预留空间就会遮住最右一列的信息。
func scrollBarInset() float32 {
	return theme.Current().Size(theme.SizeNameScrollBar) + 6
}
