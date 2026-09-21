// Command nte-tools 是异环盲盒工具的入口：登记应用元数据、创建主窗口，
// 再把界面启动交给 internal/app。
package main

import (
	"nte-tools/internal/app"
	"nte-tools/res"

	"fyne.io/fyne/v2"
	fyneapp "fyne.io/fyne/v2/app"
)

const (
	appID      = "dev.xueria.tools.nte"
	appName    = "nte-tools"
	appVersion = "0.3"

	windowTitle = "异环工具箱"
)

func main() {
	setMetadata()
	app.Run(newWindow())
}

// newWindow 创建应用与主窗口，窗口标题在这里定；尺寸、装配与显示由 internal/app
// 负责。
func newWindow() fyne.Window {
	application := fyneapp.NewWithID(appID)

	return application.NewWindow(windowTitle)
}

// setMetadata 登记应用标识、版本与图标，须在创建应用之前调用。
func setMetadata() {
	fyneapp.SetMetadata(fyne.AppMetadata{
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
