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
	windowHeight = 520
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

	// cells 是单格（1x1）物品，单格推测基于它反推组成，随数据刷新一起更新。
	var cells []bid.Item

	reload := func() {
		loadLocalBoxes(data, planner)
		loadRemoteBoxes(data, planner)

		cells = loadBidGrids(data, items)
		bidPage.SetCellItemCount(len(cells))
	}
	planner.OnRefresh = reload

	bidPage.OnInfer = func(avg, minCount, maxCount int) {
		if len(cells) == 0 {
			bidPage.SetStatus("还没有可用的单格物品数据")
			return
		}

		bidPage.SetCompositions(bid.Infer(cells, avg, minCount, maxCount))
	}

	window.SetContent(view.NewShell(
		planner.NewTab(),
		bidPage.NewTab(),
		items.NewTab(),
	))

	reload()

	window.ShowAndRun()
}

// loadLocalBoxes 读取本地盲盒并推给盲盒规划页。
func loadLocalBoxes(data *store.Store, planner *view.Planner) {
	boxes, err := data.LocalBoxes()
	if err != nil {
		planner.SetStatus(fmt.Sprintf("本地加载失败：%v", err))
		return
	}

	planner.SetStatus("")
	planner.SetLocalBoxes(boxes)
}

// loadRemoteBoxes 在后台读取远程盲盒，完成后回到 UI 线程推给盲盒规划页。
func loadRemoteBoxes(data *store.Store, planner *view.Planner) {
	planner.SetRemoteLoading(true)

	go func() {
		boxes, err := data.RemoteBoxes()

		fyne.Do(func() {
			planner.SetRemoteLoading(false)

			if err != nil {
				planner.SetStatus(fmt.Sprintf("远程加载失败：%v", err))
				return
			}

			planner.SetRemoteBoxes(boxes)
		})
	}()
}

// loadBidGrids 读取本地竞拍占格数据：推给拍品清单页，并返回其中的单格物品。
func loadBidGrids(data *store.Store, items *view.Items) []bid.Item {
	grids, err := data.BidGrids()
	if err != nil {
		items.SetStatus(fmt.Sprintf("竞拍数据加载失败：%v", err))
		return nil
	}

	items.SetStatus("")
	items.SetBidGrids(grids)

	return bid.CellItems(grids)
}
