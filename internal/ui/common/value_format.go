package common

import (
	"strconv"
	"strings"
)

// FormatValue 给价格加千分位，便于读七位数。
func FormatValue(value int) string {
	return groupDigits(strconv.Itoa(value))
}

// FormatAverage 均价保留两位小数，同样带千分位。
func FormatAverage(average float64) string {
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
