// Package settings 是「设置」页：应用级设置的入口，目前只有数据来源一项。
// 页面负责展示选项、记住上次的选择并上报变化，怎么用这个选择由装配层决定。
package settings

import (
	"strings"

	"blind-tools/internal/config"
	"blind-tools/internal/ui/shell"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// 记住选择用的键。
const (
	prefKind      = "dataSource"
	prefCustomURL = "customRemoteURL"
)

// Page 是「设置」页。
type Page struct {
	root fyne.CanvasObject

	sourceSelect *widget.Select
	sourceHint   *widget.Label
	customEntry  *widget.Entry
	customButton *widget.Button
	statusLabel  *widget.Label

	choice config.Choice

	// OnSourceChange 由装配层赋值：数据来源变化时触发，用于按新来源重新加载。
	OnSourceChange func(choice config.Choice)
}

// NewPage 构建设置页，并恢复上次记住的选择。
func NewPage() *Page {
	page := &Page{}
	page.root = page.build()

	return page
}

// Tab 返回该页的页签标题、图标与内容。
func (page *Page) Tab() shell.Tab {
	return shell.Tab{Title: "设置", Icon: theme.SettingsIcon(), Content: page.root}
}

// Choice 返回当前选中的数据来源。
func (page *Page) Choice() config.Choice {
	return page.choice
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

// build 组装设置项。
func (page *Page) build() fyne.CanvasObject {
	title := widget.NewLabelWithStyle("设置", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	sourceTitle := widget.NewLabelWithStyle("数据来源", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	page.sourceHint = widget.NewLabel("")
	page.sourceHint.Importance = widget.LowImportance
	page.sourceHint.Wrapping = fyne.TextWrapWord

	page.sourceSelect = widget.NewSelect(labels(), page.selectChanged)

	page.customEntry = widget.NewEntry()
	page.customEntry.SetPlaceHolder("https://example.com/nte/data/")
	page.customEntry.OnChanged = func(text string) {
		page.choice.CustomURL = strings.TrimSpace(text)
	}
	page.customEntry.OnSubmitted = func(string) { page.loadCustom() }

	page.customButton = widget.NewButton("用这个地址加载", page.loadCustom)
	page.customButton.Importance = widget.MediumImportance

	customRow := container.NewBorder(nil, nil, nil, page.customButton, page.customEntry)

	page.statusLabel = widget.NewLabel("")
	page.statusLabel.Wrapping = fyne.TextWrapWord
	page.statusLabel.Hide()

	content := container.NewVBox(
		title,
		widget.NewSeparator(),
		sourceTitle,
		page.sourceSelect,
		page.sourceHint,
		customRow,
		widget.NewSeparator(),
		page.statusLabel,
	)

	// 所有控件建好后才能恢复选择：选中会立刻回调 selectChanged。
	page.restore()

	return container.NewPadded(content)
}

// labels 返回下拉里的展示名。
func labels() []string {
	options := config.Options()
	names := make([]string, 0, len(options))

	for _, option := range options {
		names = append(names, option.Label)
	}

	return names
}

// restore 恢复上次记住的选择。
func (page *Page) restore() {
	page.choice = remembered()

	page.customEntry.SetText(page.choice.CustomURL)
	page.sourceSelect.SetSelected(page.choice.Label())
}

// remembered 读取上次记住的选择；存的值不可用时退回默认。
func remembered() config.Choice {
	choice := config.Default()

	prefs := appPreferences()

	if prefs == nil {
		return choice
	}

	if kind, ok := config.KindOfLabel(prefs.String(prefKind)); ok {
		choice.Kind = kind
	}

	choice.CustomURL = strings.TrimSpace(prefs.String(prefCustomURL))

	// 自定义远程还没填地址的话退回默认，免得拿空地址去请求。
	if choice.Kind == config.CustomKind && choice.CustomURL == "" {
		choice.Kind = config.Default().Kind
	}

	return choice
}

// remember 记住当前选择。
func (page *Page) remember() {
	prefs := appPreferences()

	if prefs == nil {
		return
	}

	prefs.SetString(prefKind, page.choice.Label())
	prefs.SetString(prefCustomURL, page.choice.CustomURL)
}

// appPreferences 返回应用的偏好存储；应用还没建好时为 nil。
func appPreferences() fyne.Preferences {
	if application := fyne.CurrentApp(); application != nil {
		return application.Preferences()
	}

	return nil
}

// selectChanged 跟随下拉切换选择；自定义远程还没填地址时只提示，等地址来了再加载。
func (page *Page) selectChanged(label string) {
	kind, ok := config.KindOfLabel(label)

	if !ok {
		return
	}

	page.choice.Kind = kind

	if kind == config.CustomKind && page.choice.CustomURL == "" {
		page.sourceHint.SetText("自定义远程要先填地址，再点「用这个地址加载」。")
		page.remember()

		return
	}

	page.apply()
}

// loadCustom 用输入框里的地址加载，并切到自定义远程。
func (page *Page) loadCustom() {
	customURL := strings.TrimSpace(page.customEntry.Text)

	if customURL == "" {
		page.SetStatus("请先填自定义远程的地址")

		return
	}

	page.choice = config.Choice{Kind: config.CustomKind, CustomURL: customURL}

	if page.sourceSelect.Selected == page.choice.Label() {
		// 下拉已经在这一项上，不会再回调，直接应用。
		page.apply()

		return
	}

	page.sourceSelect.SetSelected(page.choice.Label())
}

// apply 刷新说明、记住选择，并把它交给装配层重新加载。
func (page *Page) apply() {
	page.sourceHint.SetText(page.choice.Hint())
	page.remember()

	if page.OnSourceChange != nil {
		page.OnSourceChange(page.choice)
	}
}
