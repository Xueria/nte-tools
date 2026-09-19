package pool

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"

	"blind-tools/model/metadata"
)

// LoadLocalPools 按 directory 下的 metadata.json 加载全部盲盒池。
// 单个池文件读取或校验失败只跳过该池，不影响其余数据。
func LoadLocalPools(directory string) ([]Pool, error) {
	meta, err := metadata.Load(directory)

	if err != nil {
		return nil, err
	}

	var global []Resource

	if file := meta.Pool.Global.Resources; file != "" {
		global, err = LoadResources(filepath.Join(directory, file))

		if err != nil {
			log.Printf("load global resources %s failed: %v", file, err)
		}
	}

	pools := make([]Pool, 0, len(meta.Pool.List))

	for _, relative := range meta.Pool.List {
		path := filepath.Join(directory, relative)

		content, err := os.ReadFile(path)

		if err != nil {
			log.Printf("skip pool %s: read failed: %v", path, err)
			continue
		}

		var item Pool

		if err := json.Unmarshal(content, &item); err != nil {
			log.Printf("skip pool %s: unmarshal failed: %v", path, err)
			continue
		}

		applyGlobalResources(&item, global)

		if len(item.Resources) == 0 {
			log.Printf("skip pool %s: no resources", path)
			continue
		}

		if err := ValidateManifestPrices(item); err != nil {
			log.Printf("skip pool %s: %v", path, err)
			continue
		}

		if err := ValidateManifestDraws(item); err != nil {
			log.Printf("skip pool %s: %v", path, err)
			continue
		}

		pools = append(pools, item)
	}

	return pools, nil
}
