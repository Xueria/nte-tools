package common

import "fyne.io/fyne/v2/theme"

// CardInset 卡片类内容（拍品卡片、组合行）与卡片边缘的距离。
const CardInset float32 = 6

// ScrollBarInset 返回浮动滚动条的宽度加一点余量：Fyne 的滚动条画在内容之上，
// 不预留空间就会遮住最右一列的信息。
func ScrollBarInset() float32 {
	return theme.Current().Size(theme.SizeNameScrollBar) + 6
}
