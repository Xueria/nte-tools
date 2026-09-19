// Package data 提供数据来源：本地数据根目录与远程 HTTP 根，二者取其一。
// 索引与索引里列出的 JSON 数据都经来源读取，数据始终是外部文件，不随程序打包。
package data

import (
	"fmt"
	"path"
	"strings"
)

// Source 是一份数据的来源：按索引里写的相对路径（统一用 / 分隔）读出文件内容。
type Source interface {
	ReadFile(name string) ([]byte, error)
}

// cleanName 规范化索引里写的相对路径，并拦掉越出数据根的路径：索引现在也可能
// 来自远程，不能假定里面的路径一定规矩。
func cleanName(name string) (string, error) {
	cleaned := path.Clean(name)

	if cleaned == "." || cleaned == "" || path.IsAbs(cleaned) ||
		cleaned == ".." || strings.HasPrefix(cleaned, "../") {
		return "", fmt.Errorf("invalid data path %q", name)
	}

	return cleaned, nil
}
