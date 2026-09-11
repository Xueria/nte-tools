package view

import (
	"fmt"
	"strconv"
	"strings"

	"blind-tools/model/bid"

	"fyne.io/fyne/v2"
)

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

// formatValue 给价格加千分位，便于读七位数。
func formatValue(value int) string {
	return groupDigits(strconv.Itoa(value))
}

// formatAverage 均价保留两位小数，同样带千分位。
func formatAverage(average float64) string {
	whole, fraction, _ := strings.Cut(strconv.FormatFloat(average, 'f', 2, 64), ".")

	return groupDigits(whole) + "." + fraction
}

// groupDigits 给非负整数的数字串加千分位。
func groupDigits(digits string) string {
	if len(digits) <= 3 {
		return digits
	}

	lead := len(digits) % 3
	var text strings.Builder

	if lead > 0 {
		text.WriteString(digits[:lead])
	}

	for i := lead; i < len(digits); i += 3 {
		if text.Len() > 0 {
			text.WriteByte(',')
		}

		text.WriteString(digits[i : i+3])
	}

	return text.String()
}

// describeComposition 把一种组成写成「名称×件数」的形式。
func describeComposition(composition bid.Composition) string {
	parts := make([]string, 0, len(composition.Items))

	for _, group := range composition.Items {
		if group.Count > 1 {
			parts = append(parts, fmt.Sprintf("%s×%d", group.Item.Name, group.Count))
			continue
		}

		parts = append(parts, group.Item.Name)
	}

	return strings.Join(parts, "  ")
}
