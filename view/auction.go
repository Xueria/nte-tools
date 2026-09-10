package view

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// auctionView 是「竞拍估价」页。推算算法已移除，这里只保留界面壳。
type auctionView struct {
	root fyne.CanvasObject
}

// newAuctionView 构建竞拍估价页。
func newAuctionView() fyne.CanvasObject {
	v := &auctionView{}

	title := widget.NewLabelWithStyle("竞拍估价", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	placeholder := widget.NewLabel("竞拍估价功能已移除。")
	placeholder.Importance = widget.LowImportance
	placeholder.Wrapping = fyne.TextWrapWord

	content := container.NewVBox(title, widget.NewSeparator(), placeholder)
	v.root = container.NewVScroll(content)
	return v.root
}
