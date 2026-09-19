package main

import (
	"fmt"

	"blind-tools/model/bid"
	"blind-tools/res"
	"blind-tools/store"
	"blind-tools/view"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
)

const (
	appID      = "dev.xueria.tools.blind"
	appName    = "blind-tools"
	appVersion = "0.2"

	windowTitle  = "Blind Tools"
	windowWidth  = 840
	windowHeight = 640
)

func main() {
	setupMetadata()
	setupMainWindowAndRun()
}

func setupMetadata() {
	app.SetMetadata(fyne.AppMetadata{
		ID:      appID,
		Name:    appName,
		Version: appVersion,
		Build:   1,
		Icon:    res.Icon,
		Release: false,
		Custom:  nil,
		Migrations: map[string]bool{
			"fyneDo": true,
		},
	})
}

func setupMainWindowAndRun() {
	application := app.NewWithID(appID)
	window := application.NewWindow(windowTitle)

	window.Resize(fyne.NewSize(windowWidth, windowHeight))

	data := store.New()
	planner := view.NewPlanner()
	bidPage := view.NewBid()
	items := view.NewItems()

	// pool、required 与 query 记住上一次推测用的候选池与条件，
	// 用于按需展开某个件数的组合。
	var (
		pool     []bid.Item
		required []bid.Item
		query    bid.InferQuery
	)

	reload := func() {
		loadLocalPools(data, planner)
		loadRemotePools(data, planner)

		loadBidListings(data, items, bidPage)
	}
	planner.OnRefresh = reload

	// 推测先算出「确实有组合」的件数档位；具体组合等用户选中某个件数时再枚举。
	bidPage.OnInfer = func(items, req []bid.Item, q bid.InferQuery) {
		pool, required, query = items, req, q
		bidPage.SetCounts(bid.InferCounts(pool, req, q))
	}

	bidPage.OnSelectCount = func(count int) {
		bidPage.SetCompositions(bid.InferCount(pool, required, query, count))
	}

	window.SetContent(view.NewShell(
		planner.NewTab(),
		bidPage.NewTab(),
		items.NewTab(),
	))

	reload()

	window.ShowAndRun()
}

// loadLocalPools 读取本地盲盒池并推给盲盒规划页。
func loadLocalPools(data *store.Store, planner *view.Planner) {
	pools, err := data.LocalPools()
	if err != nil {
		planner.SetStatus(fmt.Sprintf("本地加载失败：%v", err))
		return
	}

	planner.SetStatus("")
	planner.SetLocalPools(pools)
}

// loadRemotePools 在后台读取远程盲盒池，完成后回到 UI 线程推给盲盒规划页。
func loadRemotePools(data *store.Store, planner *view.Planner) {
	planner.SetRemoteLoading(true)

	go func() {
		pools, err := data.RemotePools()

		fyne.Do(func() {
			planner.SetRemoteLoading(false)

			if err != nil {
				planner.SetStatus(fmt.Sprintf("远程加载失败：%v", err))
				return
			}

			planner.SetRemotePools(pools)
		})
	}()
}

// loadBidListings 读取本地竞拍清单数据：推给拍品清单页与拍品推测页。
func loadBidListings(data *store.Store, items *view.Items, bidPage *view.Bid) {
	listings, err := data.BidListings()
	if err != nil {
		items.SetStatus(fmt.Sprintf("竞拍数据加载失败：%v", err))
		return
	}

	items.SetStatus("")
	items.SetBidListings(listings)
	bidPage.SetListings(listings)
}
