// Package listing 是「拍品清单」页：清单列表、占格预览与拍品卡片。
package listing

import (
	"nte-tools/internal/bid"
	"nte-tools/internal/ui/common"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// listingSplitRatio 清单列表在左右分栏中的初始占比，左栏只放名称，取窄一些。
const listingSplitRatio = 0.22

// Page 是「拍品清单」页：左侧列出所有清单，右侧按清单的呈现方式渲染
// 预览并列出拍品卡片。
type Page struct {
	root fyne.CanvasObject

	listings []bid.Listing

	listingList  *widget.List
	previewLabel *widget.Label
	previewBox   *fyne.Container
	cardWall     *cardWall
	cardScroll   *container.Scroll
}

// NewPage 构建拍品清单页。
func NewPage() *Page {
	page := &Page{}
	page.root = page.build()

	return page
}

// Title 返回该页在导航栏里的名称。
func (page *Page) Title() string {
	return "拍品清单"
}

// Icon 返回该页在导航栏里的图标。
func (page *Page) Icon() fyne.Resource {
	return theme.ListIcon()
}

// Content 返回该页的内容。
func (page *Page) Content() fyne.CanvasObject {
	return page.root
}

// SetListings 用新的竞拍清单数据替换页面内容，并默认选中第一份清单。
func (page *Page) SetListings(listings []bid.Listing) {
	page.listings = listings
	page.listingList.Refresh()

	if len(listings) == 0 {
		page.previewLabel.SetText("暂无竞拍数据")
		page.previewBox.Objects = nil
		page.previewBox.Refresh()
		page.cardWall.setCards(nil)

		return
	}

	// 先清空选中态再选中，保证 OnSelected 一定触发。
	page.listingList.UnselectAll()
	page.listingList.Select(0)
}

// build 组装左侧清单选择器与右侧清单预览、拍品卡片。
func (page *Page) build() fyne.CanvasObject {
	page.listingList = widget.NewList(
		func() int { return len(page.listings) },
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			obj.(*widget.Label).SetText(page.listings[id].Name)
		},
	)
	page.listingList.OnSelected = func(id widget.ListItemID) { page.selectListing(int(id)) }

	listHeader := widget.NewLabelWithStyle("清单", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	header := container.NewVBox(listHeader, widget.NewSeparator())
	left := container.NewBorder(header, nil, nil, nil, page.listingList)

	page.previewLabel = widget.NewLabelWithStyle("", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	page.previewBox = container.NewVBox()

	page.cardWall = newCardWall()
	// 滚动条浮在滚动内容之上，所以把它的宽度留在内容右侧：滚动条仍停在
	// 面板右边缘，卡片区则缩短，最右一列不会被压住。
	cardContent := container.New(layout.NewCustomPaddedLayout(0, 0, 0, common.ScrollBarInset()), page.cardWall)
	page.cardScroll = container.NewVScroll(cardContent)

	preview := container.NewVBox(page.previewLabel, page.previewBox, widget.NewSeparator())

	// 内侧留出空隙，让左右两块面板读起来是独立表面。
	left = container.New(layout.NewCustomPaddedLayout(0, 0, 0, 8), left)
	right := container.New(layout.NewCustomPaddedLayout(0, 0, 8, 0),
		container.NewBorder(preview, nil, nil, nil, page.cardScroll))

	split := container.NewHSplit(left, right)
	split.Offset = listingSplitRatio

	return split
}

// selectListing 渲染第 index 份清单：标题取清单自己的名称，预览内容交给
// 清单自身决定。
func (page *Page) selectListing(index int) {
	if index < 0 || index >= len(page.listings) {
		return
	}

	listing := page.listings[index]

	page.previewLabel.SetText(listing.Name)
	page.previewBox.Objects = previewObjects(listingPreview(listing))
	page.previewBox.Refresh()

	cards := make([]fyne.CanvasObject, 0, len(listing.Items))

	for _, item := range listing.Items {
		card := newItemCard()
		card.set(item)
		cards = append(cards, card)
	}

	page.cardWall.setCards(cards)

	// 换清单时回到顶部。
	page.cardScroll.Offset = fyne.NewPos(0, 0)
	page.cardScroll.Refresh()
}
