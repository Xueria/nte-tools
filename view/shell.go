package view

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
)

// Tab 页签栏中的一页：标题、图标与内容。
type Tab struct {
	Title   string
	Icon    fyne.Resource
	Content fyne.CanvasObject
}

// NewShell 把各页装成页签窗口，只负责主题与排版，不涉及任何数据。
func NewShell(tabs ...Tab) fyne.CanvasObject {
	// Apply the Material Design 3 theme for the whole app.
	if a := fyne.CurrentApp(); a != nil {
		a.Settings().SetTheme(NewMD3Theme())
	}

	items := make([]*container.TabItem, 0, len(tabs))
	for _, tab := range tabs {
		items = append(items, container.NewTabItemWithIcon(tab.Title, tab.Icon, tab.Content))
	}

	shell := container.NewAppTabs(items...)
	// Keep the tab bar compact so both pages get the full window height.
	shell.SetTabLocation(container.TabLocationTop)

	return shell
}
