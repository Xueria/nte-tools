// Package metadata 解析数据根目录的索引文件 metadata.json。各业务域的数据
// 文件由索引显式列出，加载器不再扫描目录，因此哪些数据生效由数据自己决定。
package metadata

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// File 数据根目录下的索引文件名。
const File = "metadata.json"

// Metadata 数据根目录的索引：各业务域的数据文件路径一律相对数据根。
type Metadata struct {
	Bid  Bid  `json:"bid"`
	Pool Pool `json:"pool"`
}

// Bid 竞拍清单域：竞拍清单文件列表。
type Bid struct {
	List []string `json:"list"`
}

// Pool 盲盒池域：盲盒池文件列表，以及池未自带资源时回退的全局资源文件。
type Pool struct {
	Global Global   `json:"global"`
	List   []string `json:"list"`
}

// Global 索引里域内共享的数据文件。
type Global struct {
	Resources string `json:"resources"`
}

// Load 读取 directory 下的索引文件。
func Load(directory string) (Metadata, error) {
	path := filepath.Join(directory, File)

	content, err := os.ReadFile(path)

	if err != nil {
		return Metadata{}, fmt.Errorf("read %s failed: %w", path, err)
	}

	return decode(content)
}

// decode 解析索引内容。
func decode(content []byte) (Metadata, error) {
	var meta Metadata

	if err := json.Unmarshal(content, &meta); err != nil {
		return Metadata{}, fmt.Errorf("unmarshal %s failed: %w", File, err)
	}

	return meta, nil
}
