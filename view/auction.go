package view

import (
	"blind-tools/model"
	"fmt"
	"image/color"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// auctionView 是「竞拍估价」页：只输入单格平均价，反推这一格可能的件数与拍品构成。
type auctionView struct {
	data model.AuctionData

	averEntry  *widget.Entry
	depthEntry *widget.Entry
	searchBtn  *widget.Button
	poolHint   *widget.Label
	loadError  string

	summaryLabel *widget.Label
	resultBox    *fyne.Container

	root fyne.CanvasObject
}

// newAuctionView 构建竞拍估价页。
func newAuctionView() fyne.CanvasObject {
	v := &auctionView{}

	data, err := model.LoadAuctionDefault()
	if err != nil {
		v.loadError = err.Error()
	}
	v.data = data

	content := container.NewVBox(
		v.buildHeader(),
		widget.NewSeparator(),
		v.buildForm(),
		widget.NewSeparator(),
		v.buildSummary(),
		v.resultBox,
	)
	v.root = container.NewVScroll(content)

	if v.loadError != "" {
		v.updatePoolHint("数据加载失败：" + v.loadError)
		v.summaryLabel.SetText("没能读取 data/auction/items.json，请检查文件后点「重新加载数据」。")
		return v.root
	}

	v.updatePoolHint("")
	return v.root
}

// buildHeader 构建标题与奖池概况。
func (v *auctionView) buildHeader() fyne.CanvasObject {
	v.poolHint = widget.NewLabel("")
	v.poolHint.Wrapping = fyne.TextWrapWord

	title := widget.NewLabelWithStyle("竞拍估价", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	hint := widget.NewLabel("只填单格平均价：件数未知，程序推算「可能是几件、分别是哪些拍品」。")
	hint.Wrapping = fyne.TextWrapWord
	hint.Importance = widget.LowImportance

	return container.NewVBox(title, hint, v.poolHint)
}

// buildForm 构建输入区：平均价是必填项，件数只是搜索上限。
func (v *auctionView) buildForm() fyne.CanvasObject {
	v.averEntry = widget.NewEntry()
	v.averEntry.SetPlaceHolder("例如 2100")
	v.averEntry.Validator = numericValidator
	v.averEntry.Scroll = fyne.ScrollNone

	v.depthEntry = widget.NewEntry()
	v.depthEntry.SetText(strconv.Itoa(DefaultMaxDepth))
	v.depthEntry.Validator = numericValidator
	v.depthEntry.Scroll = fyne.ScrollNone

	form := widget.NewForm(
		widget.NewFormItem("单格平均价", v.averEntry),
		widget.NewFormItem("最多几件", v.depthEntry),
	)

	v.searchBtn = widget.NewButton("推算构成", v.search)
	v.searchBtn.Importance = widget.HighImportance

	reloadBtn := widget.NewButton("重新加载数据", v.reload)
	reloadBtn.Importance = widget.MediumImportance

	v.averEntry.OnSubmitted = func(string) { v.search() }
	v.depthEntry.OnSubmitted = func(string) { v.search() }

	hint := widget.NewLabel("「最多几件」向上搜索的上限：件数本身也是推算结果。")
	hint.Importance = widget.LowImportance
	hint.Wrapping = fyne.TextWrapWord

	return container.NewVBox(form, container.NewHBox(v.searchBtn, reloadBtn), hint)
}

// buildSummary 构建结果上方的统计行。
func (v *auctionView) buildSummary() fyne.CanvasObject {
	v.summaryLabel = widget.NewLabel("")
	v.summaryLabel.Wrapping = fyne.TextWrapWord

	v.resultBox = container.NewVBox()

	title := widget.NewLabelWithStyle("推测构成", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	return container.NewVBox(v.summaryLabel, title)
}

// updatePoolHint 刷新奖池概况。
func (v *auctionView) updatePoolHint(extra string) {
	if len(v.data.Items) == 0 {
		v.poolHint.SetText(extra)
		return
	}

	parts := make([]string, 0, len(v.data.Tiers))
	minPrice, maxPrice := 0, 0
	for _, tier := range v.data.SortedTiers() {
		parts = append(parts, fmt.Sprintf("%s%d", tier.Name, len(v.data.ItemsOfTier(tier.Key))))

		lo, hi, ok := v.data.TierPriceRange(tier.Key)
		if !ok {
			continue
		}
		if minPrice == 0 || lo < minPrice {
			minPrice = lo
		}
		if hi > maxPrice {
			maxPrice = hi
		}
	}

	text := fmt.Sprintf("奖池 %d 种拍品（%s），单价 %s ～ %s",
		len(v.data.Items), strings.Join(parts, "/"), formatNumber(minPrice), formatNumber(maxPrice))
	if extra != "" {
		text += "｜" + extra
	}
	v.poolHint.SetText(text)
}

// reload 重新读取竞拍数据。
func (v *auctionView) reload() {
	data, err := model.LoadAuctionDefault()
	if err != nil {
		v.loadError = err.Error()
		v.updatePoolHint("数据加载失败：" + v.loadError)
		v.summaryLabel.SetText("仍未读取到 data/auction/items.json。")
		return
	}

	v.loadError = ""
	v.data = data
	v.updatePoolHint("")
	v.search()
}

// search 读取输入并异步推算。
func (v *auctionView) search() {
	if len(v.data.Items) == 0 {
		return
	}

	average, err := parsePositiveInt(v.averEntry.Text)
	if err != nil || average <= 0 {
		v.summaryLabel.SetText("请填写有效的单格平均价（正整数）。")
		return
	}

	depth := DefaultMaxDepth
	if text := strings.TrimSpace(v.depthEntry.Text); text != "" {
		parsed, err := strconv.Atoi(text)
		if err != nil || parsed < 1 {
			v.summaryLabel.SetText("请填写有效的件数上限（正整数，默认 30）。")
			return
		}
		depth = parsed
	}

	v.searchBtn.Disable()
	v.summaryLabel.SetText("正在推算…")
	v.resultBox.RemoveAll()

	go func() {
		guess := searchSlotComposition(v.data, average, depth, 200)
		fyne.Do(func() {
			v.searchBtn.Enable()
			v.showGuess(guess)
		})
	}()
}

// showGuess 渲染推算结果。
func (v *auctionView) showGuess(guess SlotGuess) {
	v.resultBox.RemoveAll()

	base := fmt.Sprintf("平均价 %s（最多 %d 件）", formatNumber(guess.Average), guess.MaxDepth)

	if len(guess.Depths) == 0 {
		v.summaryLabel.SetText(fmt.Sprintf("%s｜奖池单价 %s ～ %s，这个平均价凑不出精确解",
			base, formatNumber(guess.MinPrice), formatNumber(guess.MaxPrice)))

		text := "没有任何件数能精确凑出这个平均价。"
		if len(guess.Nearest) > 0 {
			parts := make([]string, 0, len(guess.Nearest))
			for _, near := range guess.Nearest {
				parts = append(parts, formatNumber(near))
			}
			text += "\n附近能凑出的均价：" + strings.Join(parts, "、")
		}

		notice := widget.NewLabel(text)
		notice.Wrapping = fyne.TextWrapWord
		notice.Importance = widget.WarningImportance
		v.resultBox.Add(notice)
		v.resultBox.Refresh()
		return
	}

	depths := make([]string, 0, len(guess.Depths))
	for _, d := range guess.Depths {
		depths = append(depths, strconv.Itoa(d))
	}
	v.summaryLabel.SetText(fmt.Sprintf("%s｜可能的件数：%s（共 %d 种）｜下面按件数从少到多列出",
		base, strings.Join(depths, "、"), len(guess.Depths)))

	for i, combo := range guess.Combos {
		v.resultBox.Add(v.buildComboCard(i+1, combo))
	}
	v.resultBox.Refresh()
}

// buildComboCard 构建一条构成卡片。
func (v *auctionView) buildComboCard(index int, combo SlotCombo) fyne.CanvasObject {
	title := widget.NewLabelWithStyle(
		fmt.Sprintf("%d. %d 件：%d 种拍品", index, combo.Count, combo.Distinct),
		fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	rows := []fyne.CanvasObject{title}
	for _, entry := range combo.Items {
		rows = append(rows, v.buildComboRow(entry))
	}

	meta := widget.NewLabel(fmt.Sprintf("合计 %s（均价 %s）",
		formatNumber(combo.Total), formatNumber(combo.Total/combo.Count)))
	meta.Wrapping = fyne.TextWrapWord
	meta.Importance = widget.LowImportance
	rows = append(rows, meta)

	return widget.NewCard("", "", container.NewVBox(rows...))
}

// buildComboRow 构建卡片内单件拍品一行。
func (v *auctionView) buildComboRow(entry ComboItem) fyne.CanvasObject {
	bar := canvas.NewRectangle(v.tierColor(entry.Item.Tier))
	bar.SetMinSize(fyne.NewSize(6, 30))
	bar.CornerRadius = 3

	name := entry.Item.Name
	if entry.Count > 1 {
		name = fmt.Sprintf("%s ×%d", name, entry.Count)
	}

	line1 := widget.NewLabelWithStyle(name, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	line2 := widget.NewLabel(fmt.Sprintf("%s档｜单价 %s｜小计 %s",
		v.tierName(entry.Item.Tier), formatNumber(entry.Item.Price), formatNumber(entry.Item.Price*entry.Count)))
	line2.Importance = widget.LowImportance

	return container.NewHBox(bar, container.NewVBox(line1, line2))
}

// tierName 取档位显示名。
func (v *auctionView) tierName(key string) string {
	for _, tier := range v.data.Tiers {
		if tier.Key == key {
			return tier.Name
		}
	}
	return key
}

// tierColor 解析档位色条颜色。
func (v *auctionView) tierColor(key string) color.Color {
	for _, tier := range v.data.Tiers {
		if tier.Key == key {
			return parseHexColor(tier.Color)
		}
	}
	return color.Transparent
}

// parseHexColor 解析 "RRGGBB" 或 "#RRGGBB"。
func parseHexColor(text string) color.Color {
	text = strings.TrimPrefix(text, "#")
	if len(text) != 6 {
		return color.Transparent
	}
	value, err := strconv.ParseUint(text, 16, 32)
	if err != nil {
		return color.Transparent
	}
	return color.NRGBA{R: uint8(value >> 16), G: uint8(value >> 8), B: uint8(value), A: 0xFF}
}

// parsePositiveInt 解析正整数文本。
func parsePositiveInt(text string) (int, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return 0, fmt.Errorf("empty")
	}
	return strconv.Atoi(text)
}

// formatNumber 给整数加千位分隔符。
func formatNumber(n int) string {
	sign := ""
	if n < 0 {
		sign = "-"
		n = -n
	}
	digits := strconv.Itoa(n)
	var parts []string
	for len(digits) > 3 {
		parts = append([]string{digits[len(digits)-3:]}, parts...)
		digits = digits[:len(digits)-3]
	}
	parts = append([]string{digits}, parts...)
	return sign + strings.Join(parts, ",")
}
