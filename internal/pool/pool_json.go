package pool

import (
	"log"
	"path/filepath"

	"blind-tools/internal/data"
)

// LoadPools 按 directory 下的索引加载全部盲盒池，索引里 pool.list 的顺序即返回顺序。
// 单个池文件读取或校验失败只跳过该池，不影响其余数据。
func LoadPools(directory string) ([]Pool, error) {
	index, err := data.LoadIndex(directory)

	if err != nil {
		return nil, err
	}

	var global []Resource

	if file := index.Pool.Global.Resources; file != "" {
		global, err = loadResources(filepath.Join(directory, file))

		if err != nil {
			log.Printf("load global resources %s failed: %v", file, err)
		}
	}

	pools := make([]Pool, 0, len(index.Pool.List))

	for _, relative := range index.Pool.List {
		path := filepath.Join(directory, relative)

		var item Pool

		if err := data.ReadJSON(path, &item); err != nil {
			log.Printf("skip pool %s: %v", path, err)
			continue
		}

		applyGlobalResources(&item, global)

		if len(item.Resources) == 0 {
			log.Printf("skip pool %s: no resources", path)
			continue
		}

		if err := validateManifestPrices(item); err != nil {
			log.Printf("skip pool %s: %v", path, err)
			continue
		}

		if err := validateManifestDraws(item); err != nil {
			log.Printf("skip pool %s: %v", path, err)
			continue
		}

		pools = append(pools, item)
	}

	return pools, nil
}
