package shell

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// StatusBar 是窗口底部的全局状态栏：左边当前数据来源，中间加载进度或最近一次
// 错误，右边上次加载的统计。整条栏用小号字，消息按窗口宽度自行折行；
// 内容放在横向滚动容器里，窄窗口下也只是内容被滚动，不会顶大窗口的最小宽度。
type StatusBar struct {
	content fyne.CanvasObject

	source  *widget.Label
	message *messageText
	stats   *widget.Label
}

// newStatusBar 构建状态栏。
func newStatusBar() *StatusBar {
	bar := &StatusBar{
		source:  statusLabel(fyne.TextAlignLeading),
		message: newMessageText(),
		stats:   statusLabel(fyne.TextAlignTrailing),
	}

	row := container.NewBorder(nil, nil, bar.source, bar.stats, bar.message)
	bar.content = container.NewBorder(widget.NewSeparator(), nil, nil, nil, container.NewHScroll(row))

	return bar
}

// statusLabel 生成状态栏里的一段小字。
func statusLabel(align fyne.TextAlign) *widget.Label {
	label := widget.NewLabel("")
	label.Alignment = align
	label.Importance = widget.LowImportance
	label.SizeName = theme.SizeNameCaptionText

	return label
}

// SetSource 显示当前数据来源。
func (bar *StatusBar) SetSource(text string) {
	bar.source.SetText(text)
}

// SetMessage 显示加载进度或错误；传空字符串即清空。消息按当前宽度折行，
// 行数多了状态栏跟着变高。
func (bar *StatusBar) SetMessage(text string) {
	bar.message.setText(text)
}

// SetStats 显示上次加载的统计。
func (bar *StatusBar) SetStats(text string) {
	bar.stats.SetText(text)
}
