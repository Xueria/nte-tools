// Package ui 是界面的全部内容：页签外壳与三个页面（盲盒规划、拍品推测、拍品清单），
// 以及页面共用的主题、自绘控件与排版工具。页面不读数据，数据由装配层推进来。
package ui

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
	// 全应用使用 Material Design 3 主题。
	if a := fyne.CurrentApp(); a != nil {
		a.Settings().SetTheme(newMD3Theme())
	}

	items := make([]*container.TabItem, 0, len(tabs))

	for _, tab := range tabs {
		items = append(items, container.NewTabItemWithIcon(tab.Title, tab.Icon, tab.Content))
	}

	shell := container.NewAppTabs(items...)
	// 页签栏保持紧凑，把窗口高度尽量留给页面内容。
	shell.SetTabLocation(container.TabLocationTop)

	return shell
}
