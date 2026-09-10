package main

import (
	"fmt"

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
	planner.OnRefresh = func() {
		loadLocalBoxes(data, planner)
		loadRemoteBoxes(data, planner)
	}

	window.SetContent(view.NewShell(
		planner.Tab(),
		view.NewAuctionTab(),
		view.NewItemsTab(),
	))

	loadLocalBoxes(data, planner)
	loadRemoteBoxes(data, planner)

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
