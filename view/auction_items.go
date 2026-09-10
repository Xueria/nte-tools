package view

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// itemListView 是「拍品清单」页。数据加载与展示逻辑已移除，这里只保留界面壳。
type itemListView struct{}

// newItemListView 构建拍品清单页。
func newItemListView() fyne.CanvasObject {
	header := widget.NewLabelWithStyle("拍品清单", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	placeholder := widget.NewLabel("拍品清单功能已移除。")
	placeholder.Importance = widget.LowImportance
	placeholder.Wrapping = fyne.TextWrapWord

	return container.NewVBox(header, widget.NewSeparator(), placeholder)
}
