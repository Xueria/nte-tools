package view

import (
	"blind-tools/model"
	"fmt"
	"sort"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// itemListView 是「拍品清单」页：按档位展示奖池里的每一件拍品及其价格。
type itemListView struct {
	data model.AuctionData
	rows []model.AuctionItem
}

// newItemListView 构建拍品清单页。
func newItemListView() fyne.CanvasObject {
	v := &itemListView{}
	v.data, _ = model.LoadAuctionDefault()
	v.rows = v.sortedItems()

	header := widget.NewLabelWithStyle("拍品清单", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	hint := widget.NewLabel(v.summaryText())
	hint.Wrapping = fyne.TextWrapWord
	hint.Importance = widget.LowImportance

	if len(v.rows) == 0 {
		return container.NewVBox(header, hint,
			widget.NewLabel("未能读取 data/auction/items.json。"))
	}

	table := widget.NewTable(
		func() (int, int) { return len(v.rows) + 1, 4 },
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(id widget.TableCellID, cell fyne.CanvasObject) { v.updateCell(id, cell) },
	)
	table.SetColumnWidth(0, 52)  // 档位
	table.SetColumnWidth(1, 300) // 名称
	table.SetColumnWidth(2, 110) // 价格
	table.SetColumnWidth(3, 120) // 千分位

	return container.NewVBox(
		header,
		hint,
		widget.NewSeparator(),
		container.NewGridWrap(fyne.NewSize(620, 460), table),
	)
}

// sortedItems 按档位（稀有度从高到低）再按价格降序返回全部拍品。
func (v *itemListView) sortedItems() []model.AuctionItem {
	// 档位顺序就是数据文件里的声明顺序，用下标当排序键。
	order := make(map[string]int, len(v.data.Tiers))
	for i, tier := range v.data.SortedTiers() {
		order[tier.Key] = i
	}

	rows := append([]model.AuctionItem(nil), v.data.Items...)
	sort.SliceStable(rows, func(i, j int) bool {
		if order[rows[i].Tier] != order[rows[j].Tier] {
			return order[rows[i].Tier] < order[rows[j].Tier]
		}
		return rows[i].Price > rows[j].Price
	})
	return rows
}

// summaryText 生成表头上方的统计说明。
func (v *itemListView) summaryText() string {
	if len(v.data.Items) == 0 {
		return "奖池数据为空。"
	}

	text := fmt.Sprintf("共 %d 件拍品，按档位与价格排列；", len(v.data.Items))
	for i, tier := range v.data.SortedTiers() {
		if i > 0 {
			text += "、"
		}
		text += fmt.Sprintf("%s档 %d 件", tier.Name, len(v.data.ItemsOfTier(tier.Key)))
	}
	text += "。"
	return text
}

// updateCell 渲染单元格：表头加粗、价格右对齐、档位用色块强调。
// 注意 Table 会复用单元格对象，所以每次都要把所有样式重置一遍。
func (v *itemListView) updateCell(id widget.TableCellID, cell fyne.CanvasObject) {
	label := cell.(*widget.Label)
	label.Importance = widget.MediumImportance
	label.Alignment = fyne.TextAlignLeading
	label.TextStyle = fyne.TextStyle{}
	label.SizeName = theme.SizeNameText

	if id.Row == 0 {
		label.TextStyle = fyne.TextStyle{Bold: true}
		switch id.Col {
		case 0:
			label.SetText("档位")
		case 1:
			label.SetText("名称")
		case 2:
			label.Alignment = fyne.TextAlignTrailing
			label.SetText("价格")
		case 3:
			label.SetText("价格（千分位）")
		}
		label.Refresh()
		return
	}

	index := id.Row - 1
	if index < 0 || index >= len(v.rows) {
		label.SetText("")
		label.Refresh()
		return
	}
	item := v.rows[index]

	switch id.Col {
	case 0:
		label.SetText("● " + v.tierName(item.Tier))
		label.Importance = tierImportance(item.Tier)
	case 1:
		label.SetText(item.Name)
	case 2:
		label.Alignment = fyne.TextAlignTrailing
		label.SetText(fmt.Sprintf("%d", item.Price))
	case 3:
		label.SetText(formatNumber(item.Price))
	}
	label.Refresh()
}

// tierName 取档位显示名。
func (v *itemListView) tierName(key string) string {
	for _, tier := range v.data.Tiers {
		if tier.Key == key {
			return tier.Name
		}
	}
	return key
}

// tierImportance 把档位映射到主题语义色：越高档越"警示"，视觉上自然分层。
// widget.Label 没有公开的颜色字段，颜色只能通过 Importance 角色表达。
func tierImportance(key string) widget.Importance {
	switch key {
	case "red":
		return widget.DangerImportance
	case "orange":
		return widget.WarningImportance
	case "purple":
		return widget.HighImportance
	case "blue":
		return widget.SuccessImportance
	default:
		return widget.LowImportance
	}
}
