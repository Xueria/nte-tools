package ui

import (
	"errors"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

// newNumericEntry 生成一个只接受非负整数的输入框。不捕获滚轮，面板在输入框上
// 也能继续滚动。
func newNumericEntry(placeholder string) *widget.Entry {
	entry := widget.NewEntry()
	entry.SetPlaceHolder(placeholder)
	entry.Validator = numericValidator
	entry.Wrapping = fyne.TextWrapOff
	entry.Scroll = fyne.ScrollNone

	return entry
}

// parseNonNegativeInt 解析用户输入的非负整数。
func parseNonNegativeInt(text string) (int, bool) {
	value, err := strconv.Atoi(strings.TrimSpace(text))

	if err != nil || value < 0 {
		return 0, false
	}

	return value, true
}

// numericValidator 接受空串与非负整数。
func numericValidator(text string) error {
	if text == "" {
		return nil
	}

	for _, r := range text {
		if r < '0' || r > '9' {
			return errors.New("只能输入数字")
		}
	}

	return nil
}
