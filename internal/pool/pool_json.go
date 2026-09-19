package pool

import (
	"log"

	"blind-tools/internal/data"
)

// LoadPools 按来源里的索引加载全部盲盒池，索引里 pool.list 的顺序即返回顺序。
// 单个池文件读取或校验失败只跳过该池，不影响其余数据。
func LoadPools(source data.Source) ([]Pool, error) {
	index, err := data.LoadIndex(source)

	if err != nil {
		return nil, err
	}

	var global []Resource

	if file := index.Pool.Global.Resources; file != "" {
		global, err = loadResources(source, file)

		if err != nil {
			log.Printf("load global resources %s failed: %v", file, err)
		}
	}

	pools := make([]Pool, 0, len(index.Pool.List))

	for _, relative := range index.Pool.List {
		var item Pool

		if err := data.ReadJSON(source, relative, &item); err != nil {
			log.Printf("skip pool %s: %v", relative, err)
			continue
		}

		applyGlobalResources(&item, global)

		if len(item.Resources) == 0 {
			log.Printf("skip pool %s: no resources", relative)
			continue
		}

		if err := validateManifestPrices(item); err != nil {
			log.Printf("skip pool %s: %v", relative, err)
			continue
		}

		if err := validateManifestDraws(item); err != nil {
			log.Printf("skip pool %s: %v", relative, err)
			continue
		}

		pools = append(pools, item)
	}

	return pools, nil
}
