package model

import (
	"fmt"
	"sort"
)

// ValidateManifestPrices 校验盲盒 manifest 的价格表是否覆盖了该盲盒支持的所有货币
func ValidateManifestPrices(box BlindBox) error {
	required := make(map[string]struct{}, len(box.Currencies))
	for _, currency := range box.Currencies {
		required[currency.ID] = struct{}{}
	}

	if len(required) == 0 {
		return nil
	}

	covered := make(map[string]struct{})
	for _, price := range box.Manifest.Prices {
		for currencyID := range price.Cost {
			covered[currencyID] = struct{}{}
		}
	}

	var missing []string
	for currencyID := range required {
		if _, ok := covered[currencyID]; !ok {
			missing = append(missing, currencyID)
		}
	}

	if len(missing) > 0 {
		sort.Strings(missing)
		return fmt.Errorf("manifest %s (%s) price table missing currencies: %v",
			box.Manifest.ID, box.Manifest.Name, missing)
	}

	return nil
}

// ValidateManifestDraws 校验盲盒 manifest 的价格表条目数严格等于 draws
func ValidateManifestDraws(box BlindBox) error {
	actual := len(box.Manifest.Prices)
	expected := box.Manifest.Draws

	if actual != expected {
		return fmt.Errorf("manifest %s (%s) price table has %d entries, want %d (draws)",
			box.Manifest.ID, box.Manifest.Name, actual, expected)
	}

	return nil
}
