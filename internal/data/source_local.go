package data

import (
	"os"
	"path/filepath"
)

// Local 返回读取本地数据根目录下文件的来源，directory 是数据根目录。
func Local(directory string) Source {
	return localSource{root: directory}
}

// localSource 从本地数据根目录读文件。
type localSource struct {
	root string
}

// ReadFile 读取数据根目录下的相对路径文件。
func (s localSource) ReadFile(name string) ([]byte, error) {
	cleaned, err := cleanName(name)

	if err != nil {
		return nil, err
	}

	return os.ReadFile(filepath.Join(s.root, filepath.FromSlash(cleaned)))
}
