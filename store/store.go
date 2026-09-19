// Package store 是数据访问层：确定本地数据根目录在哪，
// 再把具体读取工作交给各业务域自己的加载器。
// 数据始终是外部文件，不随程序打包。
package store

import (
	"os"
	"path/filepath"

	"blind-tools/model/bid"
	"blind-tools/model/pool"
)

const (
	// DataRootEnv 环境变量：显式指定数据根目录，优先级最高。
	DataRootEnv = "BLIND_TOOLS_DATA"
	// LocalDirectory 数据根目录名，相对可执行文件或当前工作目录。
	LocalDirectory = "data"
)

// Store 数据访问层，持有已解析的数据来源。
type Store struct {
	localRoot string
}

// New 按默认策略构造数据访问层：数据根依次取环境变量、可执行文件同级的
// data/、当前工作目录下的 data/。
func New() *Store {
	return &Store{
		localRoot: resolveLocalRoot(),
	}
}

// LocalPools 读取本地盲盒池数据。
func (s *Store) LocalPools() ([]pool.Pool, error) {
	return pool.LoadLocalPools(s.localRoot)
}

// BidListings 读取本地竞拍清单数据。
func (s *Store) BidListings() ([]bid.Listing, error) {
	return bid.LoadListings(s.localRoot)
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
