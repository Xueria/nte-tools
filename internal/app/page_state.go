package app

import "fyne.io/fyne/v2"

// prefPage 记住上次停留页面的键。
const prefPage = "page"

// rememberedPage 返回上次停留的页面下标；没记住时返回 0。
func rememberedPage() int {
	prefs := appPreferences()

	if prefs == nil {
		return 0
	}

	index := prefs.Int(prefPage)
	if index < 0 {
		return 0
	}

	return index
}

// rememberPage 记住当前页面下标。
func rememberPage(index int) {
	prefs := appPreferences()

	if prefs == nil {
		return
	}

	prefs.SetInt(prefPage, index)
}

// appPreferences 返回应用的偏好存储；应用还没建好时为 nil。
func appPreferences() fyne.Preferences {
	if application := fyne.CurrentApp(); application != nil {
		return application.Preferences()
	}

	return nil
}
