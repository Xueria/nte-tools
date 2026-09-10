// Package store 是数据访问层：确定数据从哪来（本地目录与远程仓库），
// 再把具体读取工作交给各业务域自己的加载器。
// 数据始终是外部文件，不随程序打包。
package store

import (
	"os"
	"path/filepath"

	"blind-tools/model/bid"
	"blind-tools/model/blindbox"
)

const (
	// DataRootEnv 环境变量：显式指定数据根目录，优先级最高。
	DataRootEnv = "BLIND_TOOLS_DATA"
	// LocalDirectory 数据根目录名，相对可执行文件或当前工作目录。
	LocalDirectory = "data"
	// GridDirectory 竞拍占格数据在数据根下的子目录名。
	GridDirectory = "bid"
	// DefaultRemoteBaseURL 远程数据目录：必须是 raw.githubusercontent.com
	// 路径，且结构与本地数据目录一致。
	DefaultRemoteBaseURL = "https://raw.githubusercontent.com/Xueria/blind-tools/refs/heads/master/data"
)

// Store 数据访问层，持有已解析的数据来源。
type Store struct {
	localRoot  string
	remoteBase string
}

// New 按默认策略构造数据访问层：数据根依次取环境变量、可执行文件同级的
// data/、当前工作目录下的 data/；远程地址取 DefaultRemoteBaseURL。
func New() *Store {
	return &Store{
		localRoot:  resolveLocalRoot(),
		remoteBase: DefaultRemoteBaseURL,
	}
}

// LocalBoxes 读取本地盲盒数据。
func (s *Store) LocalBoxes() ([]blindbox.BlindBox, error) {
	return blindbox.LoadLocalBoxes(s.localRoot)
}

// RemoteBoxes 读取远程盲盒数据。
func (s *Store) RemoteBoxes() ([]blindbox.BlindBox, error) {
	return blindbox.LoadRemoteBoxes(s.remoteBase)
}

// BidGrids 读取本地竞拍占格数据。
func (s *Store) BidGrids() ([]bid.Grid, error) {
	return bid.LoadGrids(filepath.Join(s.localRoot, GridDirectory))
}

// resolveLocalRoot 定位数据根目录：显式环境变量优先，其次可执行文件同级，
// 最后回退到当前工作目录，保证从别处启动 exe 时仍能找到随包分发的数据。
func resolveLocalRoot() string {
	if root := os.Getenv(DataRootEnv); root != "" {
		return root
	}

	if exe, err := os.Executable(); err == nil {
		candidate := filepath.Join(filepath.Dir(exe), LocalDirectory)
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate
		}
	}

	return LocalDirectory
}
