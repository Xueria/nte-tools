package bid

import (
	"fmt"
	"strconv"
	"strings"
)

// ParseGridName 从占格数据文件名解析长宽（<length>x<width>.json）。
// ok 为 false 表示文件名不符合约定。
func ParseGridName(name string) (length, width int, ok bool) {
	lengthText, widthText, found := strings.Cut(strings.TrimSuffix(name, GridFileSuffix), "x")

	if !found {
		return 0, 0, false
	}

	length, err := strconv.Atoi(lengthText)

	if err != nil || length < 1 {
		return 0, 0, false
	}

	width, err = strconv.Atoi(widthText)

	if err != nil || width < 1 {
		return 0, 0, false
	}

	return length, width, true
}

// ValidateGrid 校验一份占格数据：长宽为正、与文件名一致，且每个物品字段完整。
func ValidateGrid(grid Grid, length, width int) error {
	if length < 1 || width < 1 {
		return fmt.Errorf("占格尺寸必须为正：%dx%d", length, width)
	}

	if grid.Length != length || grid.Width != width {
		return fmt.Errorf("文件名是 %dx%d，数据里是 %dx%d", length, width, grid.Length, grid.Width)
	}

	if len(grid.Items) == 0 {
		return fmt.Errorf("占格 %dx%d 没有任何物品", length, width)
	}

	for i, item := range grid.Items {
		if strings.TrimSpace(item.Name) == "" {
			return fmt.Errorf("第 %d 个物品缺少名称", i+1)
		}

		if strings.TrimSpace(item.Quality) == "" {
			return fmt.Errorf("物品 %s 缺少品质", item.Name)
		}

		if item.Value < 0 {
			return fmt.Errorf("物品 %s 的价格为负数：%d", item.Name, item.Value)
		}
	}

	return nil
}
