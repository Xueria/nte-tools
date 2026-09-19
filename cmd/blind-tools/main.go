// Command blind-tools 是异环盲盒工具的入口：登记应用元数据，再把界面启动交给
// internal/app。
package main

import (
	"blind-tools/internal/app"
	"blind-tools/res"

	"fyne.io/fyne/v2"
	fyneapp "fyne.io/fyne/v2/app"
)

const (
	appID      = "dev.xueria.tools.blind"
	appName    = "blind-tools"
	appVersion = "0.2"
)

func main() {
	setMetadata()
	app.Run(appID)
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
