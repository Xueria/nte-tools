// Package app 负责界面启动：建窗口、把页面与数据来源接起来。数据在后台读，
// 读完回到界面线程再更新页面，所以切换数据来源不会卡住界面。
package app

import (
	"fmt"
	"sync/atomic"

	"blind-tools/internal/bid"
	"blind-tools/internal/pool"
	"blind-tools/internal/settings"
	"blind-tools/internal/ui/infer"
	"blind-tools/internal/ui/listing"
	"blind-tools/internal/ui/planner"
	"blind-tools/internal/ui/shell"

	"fyne.io/fyne/v2"
	fyneapp "fyne.io/fyne/v2/app"
)

const (
	windowTitle  = "Blind Tools"
	windowWidth  = 840
	windowHeight = 640

	// loadingText 加载期间显示在页面底部的提示。
	loadingText = "正在加载数据…"
)

// Run 建窗口、装配页面与数据，然后进入事件循环。appID 与入口登记的应用元数据
// 是同一个标识，用于命名应用的存储空间。
func Run(appID string) {
	application := fyneapp.NewWithID(appID)
	window := application.NewWindow(windowTitle)

	window.Resize(fyne.NewSize(windowWidth, windowHeight))

	plannerPage := planner.NewPage()
	inferPage := infer.NewPage()
	listingPage := listing.NewPage()

	dataLoader := newLoader(plannerPage, listingPage, inferPage)

	// candidates、required 与 query 记住上一次推测用的候选池与条件，
	// 用于按需展开某个件数的组合。
	var (
		candidates []bid.Item
		required   []bid.Item
		query      bid.InferQuery
	)

	// 推测先算出「确实有组合」的件数档位；具体组合等用户选中某个件数时再枚举。
	inferPage.OnInfer = func(items, wanted []bid.Item, q bid.InferQuery) {
		candidates, required, query = items, wanted, q
		inferPage.SetCounts(bid.InferCounts(candidates, required, query))
	}

	inferPage.OnSelectCount = func(count int) {
		inferPage.SetCompositions(bid.InferCount(candidates, required, query, count))
	}

	plannerPage.SetSourceOptions(settings.Sources(), settings.DefaultSource())
	plannerPage.OnRefresh = dataLoader.reload
	plannerPage.OnSourceChange = dataLoader.setSource

	window.SetContent(shell.NewShell(plannerPage.Tab(), inferPage.Tab(), listingPage.Tab()))
	// 先把窗口显示出来，数据在后台读，读到多少填多少。
	window.Show()

	dataLoader.reload()

	application.Run()
}

// loader 按当前数据来源读数据，并只认最新一次的结果：切换来源够快时，
// 先发起的加载不会盖掉后发起的结果。
type loader struct {
	plannerPage *planner.Page
	listingPage *listing.Page
	inferPage   *infer.Page

	source     string
	generation atomic.Int64
}

// newLoader 构造加载器，初始用默认数据来源。
func newLoader(plannerPage *planner.Page, listingPage *listing.Page, inferPage *infer.Page) *loader {
	return &loader{
		plannerPage: plannerPage,
		listingPage: listingPage,
		inferPage:   inferPage,
		source:      settings.DefaultSource(),
	}
}

// setSource 换数据来源并重新加载。
func (l *loader) setSource(source string) {
	l.source = source
	l.reload()
}

// reload 在后台读数据，读完回到界面线程推给页面。
func (l *loader) reload() {
	token := l.generation.Add(1)
	source := settings.OpenSource(l.source)

	l.plannerPage.SetStatus(loadingText)
	l.listingPage.SetStatus(loadingText)

	go func() {
		pools, poolsErr := pool.LoadPools(source)
		listings, listingsErr := bid.LoadListings(source)

		fyne.Do(func() {
			if token != l.generation.Load() {
				return
			}

			l.apply(pools, poolsErr, listings, listingsErr)
		})
	}()
}

// apply 把读到的数据推给页面；读失败的那部分给出提示，成功的部分照常显示。
func (l *loader) apply(pools []pool.Pool, poolsErr error, listings []bid.Listing, listingsErr error) {
	if poolsErr != nil {
		l.plannerPage.SetStatus(fmt.Sprintf("盲盒数据加载失败：%v", poolsErr))
	} else {
		l.plannerPage.SetStatus("")
		l.plannerPage.SetPools(pools)
	}

	if listingsErr != nil {
		l.listingPage.SetStatus(fmt.Sprintf("竞拍数据加载失败：%v", listingsErr))

		return
	}

	l.listingPage.SetStatus("")
	l.listingPage.SetListings(listings)
	l.inferPage.SetListings(listings)
}
