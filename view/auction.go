package view

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// NewAuctionTab 构建「竞拍估价」页。推算算法已移除，这里只保留界面壳。
func NewAuctionTab() Tab {
	title := widget.NewLabelWithStyle("竞拍估价", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	placeholder := widget.NewLabel("竞拍估价功能已移除。")
	placeholder.Importance = widget.LowImportance
	placeholder.Wrapping = fyne.TextWrapWord

	content := container.NewVBox(title, widget.NewSeparator(), placeholder)

	return Tab{
		Title:   "竞拍估价",
		Icon:    theme.SearchIcon(),
		Content: container.NewVScroll(content),
	}
}
