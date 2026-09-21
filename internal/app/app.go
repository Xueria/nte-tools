// Package app 负责界面启动：把页面与数据来源接起来，装进入口建好的窗口。数据在
// 后台读，读完回到界面线程再更新页面，所以切换数据来源不会卡住界面。
package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"strings"
	"sync/atomic"

	"nte-tools/internal/bid"
	"nte-tools/internal/config"
	"nte-tools/internal/pool"
	"nte-tools/internal/ui/infer"
	"nte-tools/internal/ui/listing"
	"nte-tools/internal/ui/planner"
	"nte-tools/internal/ui/settings"
	"nte-tools/internal/ui/shell"

	"fyne.io/fyne/v2"
)

const (
	windowWidth  float32 = 960
	windowHeight float32 = 640

	// loadingText 加载期间显示在状态栏里的提示。
	loadingText = "正在加载数据…"
)

// Run 装配页面与数据，把内容装进入口建好的窗口，然后进入事件循环。
func Run(window fyne.Window) {
	plannerPage := planner.NewPage()
	inferPage := infer.NewPage()
	listingPage := listing.NewPage()
	settingsPage := settings.NewPage()

	// 设置是应用级入口，固定在导航栏底部，与其他功能页分开。
	pages := []shell.Page{plannerPage, inferPage, listingPage, shell.Pinned(settingsPage)}

	ui := shell.NewApp(pages...)
	status := ui.Status()

	dataLoader := newLoader(plannerPage, listingPage, inferPage, status)

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

	plannerPage.OnRefresh = dataLoader.reload
	settingsPage.OnSourceChange = dataLoader.setSource

	ui.OnPageChange = rememberPage

	// F5 按当前来源重新加载。
	window.Canvas().SetOnTypedKey(func(event *fyne.KeyEvent) {
		if event.Name == fyne.KeyF5 {
			dataLoader.reload()
		}
	})

	window.Resize(fyne.NewSize(windowWidth, windowHeight))
	// 居中显示，免得窗口跑到屏幕边角之外。
	window.CenterOnScreen()
	window.SetContent(ui.Content())
	// 先把窗口显示出来，数据在后台读，读到多少填多少。
	window.Show()

	// 回到上次停留的页面，并按设置页恢复出来的选择做第一次加载。
	ui.Show(rememberedPage())
	dataLoader.setSource(settingsPage.Choice())

	// 应用由入口创建，事件循环交给当前应用跑。
	fyne.CurrentApp().Run()
}

// loader 按当前数据来源读数据，并只认最新一次的结果：切换来源够快时，
// 先发起的加载不会盖掉后发起的结果。
type loader struct {
	plannerPage *planner.Page
	listingPage *listing.Page
	inferPage   *infer.Page
	status      *shell.StatusBar

	choice     config.Choice
	generation atomic.Int64
}

// newLoader 构造加载器，初始用配置里的默认数据来源。
func newLoader(plannerPage *planner.Page, listingPage *listing.Page, inferPage *infer.Page,
	status *shell.StatusBar) *loader {
	return &loader{
		plannerPage: plannerPage,
		listingPage: listingPage,
		inferPage:   inferPage,
		status:      status,
		choice:      config.Default(),
	}
}

// setSource 换数据来源并重新加载。
func (l *loader) setSource(choice config.Choice) {
	l.choice = choice
	l.reload()
}

// reload 在后台读数据，读完回到界面线程推给页面与状态栏。
func (l *loader) reload() {
	token := l.generation.Add(1)
	source := l.choice.Open()

	l.status.SetSource(l.choice.Label())
	l.status.SetMessage(loadingText)
	l.status.SetStats("")

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

// apply 把读到的数据推给页面。读失败的那部分只在状态栏报一句简短原因，
// 完整错误仍由 log 输出到标准错误，从控制台启动时可见。
func (l *loader) apply(pools []pool.Pool, poolsErr error, listings []bid.Listing, listingsErr error) {
	var reasons []string

	if poolsErr != nil {
		reasons = append(reasons, "盲盒数据"+shortReason(poolsErr))
	} else {
		l.plannerPage.SetPools(pools)
	}

	if listingsErr != nil {
		reasons = append(reasons, "竞拍数据"+shortReason(listingsErr))
	} else {
		l.listingPage.SetListings(listings)
		l.inferPage.SetListings(listings)
	}

	if len(reasons) > 0 {
		l.status.SetMessage("加载失败：" + strings.Join(reasons, "；"))

		return
	}

	l.status.SetMessage("")
	l.status.SetStats(fmt.Sprintf("%d 份清单 · %d 个盲盒池", len(listings), len(pools)))
}

// shortReason 把加载错误压成一句简短原因，供状态栏显示。
func shortReason(err error) string {
	switch {
	case errors.Is(err, fs.ErrNotExist):
		return "没有数据文件"
	case isDataError(err):
		return "数据文件内容有误"
	default:
		return "读取失败"
	}
}

// isDataError 判断是不是数据文件本身解析不了（语法错、字段类型不对等）。
func isDataError(err error) bool {
	var syntaxErr *json.SyntaxError
	var typeErr *json.UnmarshalTypeError

	return errors.As(err, &syntaxErr) || errors.As(err, &typeErr)
}
