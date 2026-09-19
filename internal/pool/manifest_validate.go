package pool

import (
	"fmt"
	"sort"
)

// validateManifestPrices 校验价格表覆盖了该池声明的所有资源。
func validateManifestPrices(p Pool) error {
	declared := make(map[string]struct{}, len(p.Resources))

	for _, resource := range p.Resources {
		declared[resource.ID] = struct{}{}
	}

	if len(declared) == 0 {
		return nil
	}

	covered := make(map[string]struct{})

	for _, price := range p.Manifest.Prices {
		for id := range price.Cost {
			covered[id] = struct{}{}
		}
	}

	var missing []string

	for id := range declared {
		if _, ok := covered[id]; !ok {
			missing = append(missing, id)
		}
	}

	if len(missing) > 0 {
		sort.Strings(missing)

		return fmt.Errorf("清单 %s（%s）的价格表缺少资源：%v", p.Manifest.ID, p.Manifest.Name, missing)
	}

	return nil
}

// validateManifestDraws 校验价格表条目数严格等于 draws。
func validateManifestDraws(p Pool) error {
	actual, expected := len(p.Manifest.Prices), p.Manifest.Draws

	if actual != expected {
		return fmt.Errorf("清单 %s（%s）的价格表有 %d 条，应为 %d 条（draws）",
			p.Manifest.ID, p.Manifest.Name, actual, expected)
	}

	return nil
}
