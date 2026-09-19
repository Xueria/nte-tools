// Package settings 是「设置」页：应用级设置的入口，目前只有数据来源一项。
// 页面只管展示与上报选择，具体怎么用这个选择由装配层决定。
package settings

import (
	"blind-tools/internal/config"
	"blind-tools/internal/ui/shell"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// Page 是「设置」页。
type Page struct {
	root fyne.CanvasObject

	sourceSelect *widget.Select
	sourceHint   *widget.Label
	statusLabel  *widget.Label

	// OnSourceChange 由装配层赋值：用户切换数据来源时触发，参数是选中的来源。
	OnSourceChange func(source string)
}

// NewPage 构建设置页。
func NewPage() *Page {
	page := &Page{}
	page.root = page.build()

	return page
}

// Tab 返回该页的页签标题、图标与内容。
func (page *Page) Tab() shell.Tab {
	return shell.Tab{Title: "设置", Icon: theme.SettingsIcon(), Content: page.root}
}

// SetStatus 显示加载进度或结果；传空字符串即隐藏。
func (page *Page) SetStatus(text string) {
	if text == "" {
		page.statusLabel.SetText("")
		page.statusLabel.Hide()

		return
	}

	page.statusLabel.SetText(text)
	page.statusLabel.Show()
}

// build 组装数据来源设置。
func (page *Page) build() fyne.CanvasObject {
	title := widget.NewLabelWithStyle("设置", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	sourceTitle := widget.NewLabelWithStyle("数据来源", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	page.sourceSelect = widget.NewSelect(config.Sources(), page.sourceChanged)
	page.sourceSelect.SetSelected(config.DefaultSource())

	page.sourceHint = widget.NewLabel(config.SourceHint(config.DefaultSource()))
	page.sourceHint.Importance = widget.LowImportance
	page.sourceHint.Wrapping = fyne.TextWrapWord

	page.statusLabel = widget.NewLabel("")
	page.statusLabel.Wrapping = fyne.TextWrapWord
	page.statusLabel.Hide()

	content := container.NewVBox(
		title,
		widget.NewSeparator(),
		sourceTitle,
		page.sourceSelect,
		page.sourceHint,
		widget.NewSeparator(),
		page.statusLabel,
	)

	return container.NewPadded(content)
}

// sourceChanged 跟着选择更新说明，并把选择转给装配层。
func (page *Page) sourceChanged(source string) {
	page.sourceHint.SetText(config.SourceHint(source))

	if page.OnSourceChange != nil {
		page.OnSourceChange(source)
	}
}
