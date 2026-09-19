// Command blind-tools 是异环盲盒工具：装配界面与本地数据，然后进入窗口事件循环。
package main

import (
	"fmt"

	"blind-tools/internal/bid"
	"blind-tools/internal/data"
	"blind-tools/internal/pool"
	"blind-tools/internal/ui"
	"blind-tools/res"

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
	setAppMetadata()
	run()
}

// setAppMetadata 登记应用标识、版本与图标，须在创建应用之前调用。
func setAppMetadata() {
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

// run 建窗口、装配页面与数据，然后进入事件循环。
func run() {
	application := app.NewWithID(appID)
	window := application.NewWindow(windowTitle)

	window.Resize(fyne.NewSize(windowWidth, windowHeight))

	root := data.Root()

	planner := ui.NewPlannerPage()
	infer := ui.NewInferPage()
	listing := ui.NewListingPage()

	// candidates、required 与 query 记住上一次推测用的候选池与条件，
	// 用于按需展开某个件数的组合。
	var (
		candidates []bid.Item
		required   []bid.Item
		query      bid.InferQuery
	)

	// 推测先算出「确实有组合」的件数档位；具体组合等用户选中某个件数时再枚举。
	infer.OnInfer = func(items, wanted []bid.Item, q bid.InferQuery) {
		candidates, required, query = items, wanted, q
		infer.SetCounts(bid.InferCounts(candidates, required, query))
	}

	infer.OnSelectCount = func(count int) {
		infer.SetCompositions(bid.InferCount(candidates, required, query, count))
	}

	reload := func() {
		loadPools(root, planner)
		loadListings(root, listing, infer)
	}
	planner.OnRefresh = reload

	window.SetContent(ui.NewShell(planner.Tab(), infer.Tab(), listing.Tab()))

	reload()

	window.ShowAndRun()
}

// loadPools 读取本地盲盒池并推给盲盒规划页。
func loadPools(root string, planner *ui.PlannerPage) {
	pools, err := pool.LoadPools(root)

	if err != nil {
		planner.SetStatus(fmt.Sprintf("本地加载失败：%v", err))

		return
	}

	planner.SetStatus("")
	planner.SetPools(pools)
}

// loadListings 读取本地竞拍清单数据：推给拍品清单页与拍品推测页。
func loadListings(root string, listing *ui.ListingPage, infer *ui.InferPage) {
	listings, err := bid.LoadListings(root)

	if err != nil {
		listing.SetStatus(fmt.Sprintf("竞拍数据加载失败：%v", err))

		return
	}

	listing.SetStatus("")
	listing.SetListings(listings)
	infer.SetListings(listings)
}
