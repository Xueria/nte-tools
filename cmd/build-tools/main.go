package main

import (
	"blind-tools/model"
	"blind-tools/res"
	"blind-tools/view"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
)

const (
	AppID       = "dev.xueria.tools.blind"
	WindowTitle = "Blind Tools"
	Width       = 840
	Height      = 520
)

var data []model.Container

func main() {
	setupMetadata()
	prepareLoad()
	mainWindow()
}

func prepareLoad() {
	data, _ = model.LoadDataDefault()
}

func setupMetadata() {
	//goland:noinspection SpellCheckingInspection
	app.SetMetadata(fyne.AppMetadata{
		ID:      "dev.xueria.tools.blind",
		Name:    "blind-tools",
		Version: "0.2",
		Build:   1,
		Icon:    res.Icon,
		Release: false,
		Custom:  nil,
		Migrations: map[string]bool{
			"fyneDo": true,
		},
	})
}

func mainWindow() {
	application := app.NewWithID(AppID)
	window := application.NewWindow(WindowTitle)

	window.Resize(fyne.NewSize(Width, Height))
	window.SetContent(view.MainView(window, data))
	window.ShowAndRun()
}
