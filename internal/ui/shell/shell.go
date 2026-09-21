// Package shell 是窗口外壳：左边一列导航切页，右边是当前页内容，窗口底部一条
// 全局状态栏。页面不读数据，数据由装配层推进来。
package shell

import (
	"image/color"

	"nte-tools/internal/ui/common"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
)

// App 是装配好的外壳：导航栏、内容区与全局状态栏。
type App struct {
	root    fyne.CanvasObject
	rail    *navRail
	content *fyne.Container
	pages   []Page
	status  *StatusBar

	// OnPageChange 由装配层赋值：用户切页时触发，可用于记住上次停留的页面。
	OnPageChange func(index int)
}

// NewApp 把各页装进导航外壳，并统一应用主题。
func NewApp(pages ...Page) *App {
	// 全应用使用 Material Design 3 主题。
	if application := fyne.CurrentApp(); application != nil {
		application.Settings().SetTheme(common.NewMD3Theme())
	}

	app := &App{pages: pages, status: newStatusBar()}

	contents := make([]fyne.CanvasObject, 0, len(pages))

	for _, page := range pages {
		contents = append(contents, page.Content())
	}

	app.content = container.NewStack(contents...)

	app.rail = newNavRail(pages)
	app.rail.OnSelect = app.show

	// 内容区包一层双向滚动：窗口最小尺寸不再由页面的最小尺寸决定，窗口可以缩得很小，
	// 缩过头时出现滚动条而不是卡住。
	contentArea := container.NewScroll(app.content)

	railArea := container.NewStack(canvas.NewRectangle(railBackgroundColor()), app.rail)
	app.root = container.NewBorder(nil, app.status.content, railArea, divider(), contentArea)

	app.show(0)

	return app
}

// Content 返回可以交给窗口的根对象。
func (app *App) Content() fyne.CanvasObject {
	return app.root
}

// Status 返回全局状态栏，装配层用它报告数据来源与加载情况。
func (app *App) Status() *StatusBar {
	return app.status
}

// Show 切到第 index 页；越界时回到第一页。
func (app *App) Show(index int) {
	if index < 0 || index >= len(app.pages) {
		index = 0
	}

	app.rail.selectItem(index)
}

// show 只让第 index 页可见，其余页留着但不显示（切换时不丢页面状态）。
func (app *App) show(index int) {
	if index < 0 || index >= len(app.pages) {
		return
	}

	for i, page := range app.pages {
		if i == index {
			page.Content().Show()
			continue
		}

		page.Content().Hide()
	}

	app.content.Refresh()

	if app.OnPageChange != nil {
		app.OnPageChange(index)
	}
}

// divider 是导航栏与内容之间的竖线。
func divider() fyne.CanvasObject {
	th, variant := themeColors()

	line := canvas.NewRectangle(th.Color(theme.ColorNameSeparator, variant))
	line.SetMinSize(fyne.NewSize(1, 0))

	return line
}

// railBackgroundColor 返回导航栏的底色，让导航与内容分成两块表面。
func railBackgroundColor() color.Color {
	th, variant := themeColors()

	return th.Color(theme.ColorNameHeaderBackground, variant)
}

// themeColors 返回当前主题与明暗变体；应用还没建好时用主题默认值。
func themeColors() (fyne.Theme, fyne.ThemeVariant) {
	if application := fyne.CurrentApp(); application != nil {
		appSettings := application.Settings()

		return appSettings.Theme(), appSettings.ThemeVariant()
	}

	return theme.Current(), theme.VariantDark
}
