package pool

import (
	"fmt"
	"sort"
)

// ValidateManifestPrices 校验盲盒池 manifest 的价格表是否覆盖了该池支持的所有资源
func ValidateManifestPrices(p Pool) error {
	required := make(map[string]struct{}, len(p.Resources))
	for _, resource := range p.Resources {
		required[resource.ID] = struct{}{}
	}

	if len(required) == 0 {
		return nil
	}

	covered := make(map[string]struct{})
	for _, price := range p.Manifest.Prices {
		for resourceID := range price.Cost {
			covered[resourceID] = struct{}{}
		}
	}

	var missing []string
	for resourceID := range required {
		if _, ok := covered[resourceID]; !ok {
			missing = append(missing, resourceID)
		}
	}

	if len(missing) > 0 {
		sort.Strings(missing)
		return fmt.Errorf("manifest %s (%s) price table missing resources: %v",
			p.Manifest.ID, p.Manifest.Name, missing)
	}

	return nil
}

// ValidateManifestDraws 校验盲盒池 manifest 的价格表条目数严格等于 draws
func ValidateManifestDraws(p Pool) error {
	actual := len(p.Manifest.Prices)
	expected := p.Manifest.Draws

	if actual != expected {
		return fmt.Errorf("manifest %s (%s) price table has %d entries, want %d (draws)",
			p.Manifest.ID, p.Manifest.Name, actual, expected)
	}

	return nil
}
