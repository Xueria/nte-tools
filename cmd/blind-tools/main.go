package main

import (
	"blind-tools/model"
	"blind-tools/res"
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

var boxes []model.BlindBox

func main() {
	setupMetadata()
	loadLocalData()
	runMainWindow()
}

func loadLocalData() {
	boxes, _ = model.LoadLocalBoxes(model.DataDirectory)
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

func runMainWindow() {
	application := app.NewWithID(appID)
	window := application.NewWindow(windowTitle)

	window.Resize(fyne.NewSize(windowWidth, windowHeight))
	window.SetContent(view.MainView(window, boxes))
	window.ShowAndRun()
}
