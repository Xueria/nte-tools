// Command blind-tools 是异环盲盒工具：装配界面与本地数据，然后进入窗口事件循环。
package main

import (
	"fmt"

	"blind-tools/internal/bid"
	"blind-tools/internal/data"
	"blind-tools/internal/pool"
	"blind-tools/internal/ui/infer"
	"blind-tools/internal/ui/listing"
	"blind-tools/internal/ui/planner"
	"blind-tools/internal/ui/shell"
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

	// remoteBaseURL 远程数据根：索引与索引里列出的数据文件都先从这里取，
	// 取不到再落回本地数据目录。
	remoteBaseURL = "https://nte-data.xueria.workers.dev/nte/data/"

	// 数据来源选项：自动 = 远程优先、失败落回本地。
	sourceAuto   = "自动（远程优先）"
	sourceRemote = "仅远程"
	sourceLocal  = "仅本地"
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

	// 数据来源可以在左栏切换：自动是远程优先、失败落回本地，也可以只用其中一边。
	remote := data.Remote(remoteBaseURL)
	local := data.Local(data.Root())
	mode := sourceAuto

	plannerPage := planner.NewPage()
	inferPage := infer.NewPage()
	listingPage := listing.NewPage()

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

	// sourceFor 按选择的数据来源造来源，并给出说明这次实际用了哪一边的文字。
	sourceFor := func(selected string) (data.Source, func() string) {
		switch selected {
		case sourceRemote:
			return remote, func() string { return "当前来源：远程" }
		case sourceLocal:
			return local, func() string { return "当前来源：本地" }
		default:
			// 组合来源按次构造：远程整体不可用时，这一次加载剩下的文件直接走本地，
			// 不会逐个等超时；下次刷新会重新试一次远程。
			chain := data.Fallback(remote, local)

			return chain, func() string { return describeOrigin(chain.Origin()) }
		}
	}

	reload := func() {
		source, describe := sourceFor(mode)
		loadPools(source, plannerPage)
		loadListings(source, listingPage, inferPage)
		plannerPage.SetSourceStatus(describe())
	}
	plannerPage.OnRefresh = reload

	plannerPage.SetSourceOptions([]string{sourceAuto, sourceRemote, sourceLocal}, sourceAuto)
	plannerPage.OnSourceChange = func(option string) {
		mode = option
		reload()
	}

	window.SetContent(shell.NewShell(plannerPage.Tab(), inferPage.Tab(), listingPage.Tab()))

	reload()

	window.ShowAndRun()
}

// loadPools 读取盲盒池并推给盲盒规划页。
func loadPools(source data.Source, plannerPage *planner.Page) {
	pools, err := pool.LoadPools(source)

	if err != nil {
		plannerPage.SetStatus(fmt.Sprintf("数据加载失败：%v", err))

		return
	}

	plannerPage.SetStatus("")
	plannerPage.SetPools(pools)
}

// loadListings 读取竞拍清单数据：推给拍品清单页与拍品推测页。
func loadListings(source data.Source, listingPage *listing.Page, inferPage *infer.Page) {
	listings, err := bid.LoadListings(source)

	if err != nil {
		listingPage.SetStatus(fmt.Sprintf("竞拍数据加载失败：%v", err))

		return
	}

	listingPage.SetStatus("")
	listingPage.SetListings(listings)
	inferPage.SetListings(listings)
}

// describeOrigin 把数据实际读到哪一边翻成界面上的说明。
func describeOrigin(origin data.Origin) string {
	switch origin {
	case data.OriginPrimary:
		return "当前来源：远程"
	case data.OriginSecondary:
		return "当前来源：本地 · 远程不可用"
	default:
		return "当前来源：—"
	}
}
