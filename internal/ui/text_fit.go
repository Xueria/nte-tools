package ui

import "fyne.io/fyne/v2"

// fitText 把文本截断到给定宽度内，超出部分用省略号；宽度未知时原样返回。
func fitText(text string, maxWidth, textSize float32, style fyne.TextStyle) string {
	if text == "" || maxWidth <= 0 {
		return text
	}

	if fyne.MeasureText(text, textSize, style).Width <= maxWidth {
		return text
	}

	runes := []rune(text)

	for len(runes) > 1 {
		runes = runes[:len(runes)-1]

		if fyne.MeasureText(string(runes)+"…", textSize, style).Width <= maxWidth {
			return string(runes) + "…"
		}
	}

	return "…"
}
