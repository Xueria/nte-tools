// Package data 提供数据来源：本地数据根目录与远程 HTTP 根，两者可以组合成
// 「远程优先、失败落回本地」的来源。索引与索引里列出的 JSON 数据都经来源读取，
// 数据始终是外部文件，不随程序打包。
package data

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"path"
	"strings"
	"sync"
)

// Source 是一份数据的来源：按索引里写的相对路径（统一用 / 分隔）读出文件内容。
type Source interface {
	ReadFile(name string) ([]byte, error)
}

// Fallback 把两个来源串成一个：先用 primary 读，读不到再读 secondary 的同名文件。
// 只有 primary 报出「这一份数据整体不可用」（连不上、超时、非 404 的状态码）时才
// 停用它，之后直接走 secondary，免得离线时每个文件都白等一次超时；单独一个文件
// 404 只影响那一个文件，其余仍优先用 primary。
func Fallback(primary, secondary Source) Source {
	return &fallbackSource{primary: primary, secondary: secondary}
}

// fallbackSource 按上面的策略组合两个来源。
type fallbackSource struct {
	primary   Source
	secondary Source

	mu          sync.Mutex
	primaryDown bool
}

// ReadFile 优先读 primary，必要时退回 secondary。
func (s *fallbackSource) ReadFile(name string) ([]byte, error) {
	if s.primaryUsable() {
		content, err := s.primary.ReadFile(name)

		if err == nil {
			return content, nil
		}

		if !errors.Is(err, fs.ErrNotExist) {
			s.disablePrimary()
		}

		log.Printf("read %s from primary source failed, fallback to secondary: %v", name, err)
	}

	return s.secondary.ReadFile(name)
}

// primaryUsable 返回 primary 是否还可以尝试。
func (s *fallbackSource) primaryUsable() bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	return !s.primaryDown
}

// disablePrimary 记住 primary 整体不可用。
func (s *fallbackSource) disablePrimary() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.primaryDown = true
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
