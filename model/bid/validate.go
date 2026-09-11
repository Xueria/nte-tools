package bid

import (
	"fmt"
	"strconv"
	"strings"
)

// ParseSizeFromName 从清单文件名解析占格长宽（<length>x<width>.json）。
// ok 为 false 表示文件名没有带长宽，该清单按普通拍品处理。
func ParseSizeFromName(name string) (length, width int, ok bool) {
	lengthText, widthText, found := strings.Cut(strings.TrimSuffix(name, ListingFileSuffix), "x")

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

// ValidateListing 校验一份清单：名称非空、占格时长宽为正，且每件拍品字段完整。
func ValidateListing(listing Listing) error {
	if strings.TrimSpace(listing.Name) == "" {
		return fmt.Errorf("清单缺少名称")
	}

	attribute := listing.Attribute

	if attribute.Grid && (attribute.Length < 1 || attribute.Width < 1) {
		return fmt.Errorf("占格尺寸必须为正：%dx%d", attribute.Length, attribute.Width)
	}

	if len(listing.Items) == 0 {
		return fmt.Errorf("清单 %s 没有任何拍品", listing.Name)
	}

	for i, item := range listing.Items {
		if strings.TrimSpace(item.Name) == "" {
			return fmt.Errorf("第 %d 件拍品缺少名称", i+1)
		}

		if strings.TrimSpace(item.Quality) == "" {
			return fmt.Errorf("拍品 %s 缺少品质", item.Name)
		}

		if item.Value < 0 {
			return fmt.Errorf("拍品 %s 的价格为负数：%d", item.Name, item.Value)
		}
	}

	return nil
}
