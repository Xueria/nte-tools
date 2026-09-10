package view

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// NewItemsTab 构建「拍品清单」页。数据加载与展示逻辑已移除，这里只保留界面壳。
func NewItemsTab() Tab {
	header := widget.NewLabelWithStyle("拍品清单", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	placeholder := widget.NewLabel("拍品清单功能已移除。")
	placeholder.Importance = widget.LowImportance
	placeholder.Wrapping = fyne.TextWrapWord

	content := container.NewVBox(header, widget.NewSeparator(), placeholder)

	return Tab{
		Title:   "拍品清单",
		Icon:    theme.ListIcon(),
		Content: content,
	}
}
