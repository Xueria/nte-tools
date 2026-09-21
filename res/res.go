package res

import (
	_ "embed"

	"fyne.io/fyne/v2"
)

//go:embed assets/icon.ico
var resourceIconData []byte

var Icon = &fyne.StaticResource{
	StaticName:    "assets/icon.ico",
	StaticContent: resourceIconData,
}
