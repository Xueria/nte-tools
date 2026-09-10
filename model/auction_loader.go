package model

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

const (
	// AuctionDirectory 1x1 竞拍数据目录（相对工作目录）。
	AuctionDirectory = "data/auction"
	// AuctionFile 1x1 奖池数据：每个品质下是「名字 -> 价格」。
	AuctionFile = "1x1.json"
)

// auctionFile 是 1x1.json 的结构：
//
//	{
//	  "tiers": [{ "key": "red", "name": "红", "color": "B3261E" }],
//	  "red":    { "永恒之星": 1314520 },
//	  "purple": { "飨目": 3060 }
//	}
//
// 顶层除 tiers 外的每个键都是一个品质，值是该品质下的「名字 -> 价格」。
type auctionFile struct {
	Tiers []AuctionTier `json:"tiers"`
}

// LoadAuctionDefault 读取默认目录下的 1x1 奖池数据。
func LoadAuctionDefault() (AuctionData, error) {
	return LoadAuction(AuctionDirectory)
}

// LoadAuction 读取指定目录下的 1x1 奖池数据并做一次校验。
func LoadAuction(directory string) (AuctionData, error) {
	var data AuctionData

	content, err := os.ReadFile(filepath.Join(directory, AuctionFile))
	if err != nil {
		return data, fmt.Errorf("读取 %s 失败：%w", AuctionFile, err)
	}

	// 先解析出 raw 拿到全部顶层键，再逐品质取「名字 -> 价格」。
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(content, &raw); err != nil {
		return data, fmt.Errorf("解析 %s 失败：%w", AuctionFile, err)
	}

	if tierRaw, ok := raw["tiers"]; ok {
		if err := json.Unmarshal(tierRaw, &data.Tiers); err != nil {
			return data, fmt.Errorf("解析 %s 的 tiers 失败：%w", AuctionFile, err)
		}
	}
	if len(data.Tiers) == 0 {
		data.Tiers = defaultTiers()
	}

	for key, tierRaw := range raw {
		if key == "tiers" {
			continue
		}
		var entries map[string]int
		if err := json.Unmarshal(tierRaw, &entries); err != nil {
			return data, fmt.Errorf("解析品质 %q 失败（应为「名字: 价格」）：%w", key, err)
		}
		for name, price := range entries {
			data.Items = append(data.Items, AuctionItem{Tier: key, Name: name, Price: price})
		}
	}

	// 排序：先按档位声明顺序，档位内按价格降序。
	order := make(map[string]int, len(data.Tiers))
	for i, tier := range data.Tiers {
		order[tier.Key] = i
	}
	sort.SliceStable(data.Items, func(i, j int) bool {
		oi, oj := order[data.Items[i].Tier], order[data.Items[j].Tier]
		if oi != oj {
			return oi < oj
		}
		return data.Items[i].Price > data.Items[j].Price
	})

	if err := ValidateAuction(data); err != nil {
		return data, fmt.Errorf("校验 %s 失败：%w", AuctionFile, err)
	}
	return data, nil
}

// defaultTiers 是数据文件没写 tiers 时的兜底档位表。
func defaultTiers() []AuctionTier {
	return []AuctionTier{
		{Key: "red", Name: "红", Color: "B3261E"},
		{Key: "orange", Name: "橙", Color: "C0602A"},
		{Key: "purple", Name: "紫", Color: "7A3F9D"},
		{Key: "blue", Name: "蓝", Color: "2E5397"},
		{Key: "grey", Name: "灰", Color: "4A494E"},
	}
}

// ValidateAuction 校验品质、名字与价格。
func ValidateAuction(data AuctionData) error {
	if len(data.Items) == 0 {
		return fmt.Errorf("奖池为空")
	}

	known := make(map[string]bool, len(data.Tiers))
	for _, tier := range data.Tiers {
		known[tier.Key] = true
	}

	seen := make(map[string]string, len(data.Items))
	for _, item := range data.Items {
		if item.Price <= 0 {
			return fmt.Errorf("拍品 %q 的价格不合法：%d", item.Name, item.Price)
		}
		if item.Name == "" {
			return fmt.Errorf("存在没有名字的拍品（品质 %q，价格 %d）", item.Tier, item.Price)
		}
		if !known[item.Tier] {
			return fmt.Errorf("拍品 %q 的品质 %q 未在 tiers 里定义", item.Name, item.Tier)
		}
		if previous, ok := seen[item.Name]; ok {
			return fmt.Errorf("拍品名字重复：%q（%s 与 %s）", item.Name, previous, item.Tier)
		}
		seen[item.Name] = item.Tier
	}
	return nil
}

// SortedTiers 返回档位表（数据文件里的声明顺序，即稀有度从高到低）。
func (d AuctionData) SortedTiers() []AuctionTier {
	return append([]AuctionTier(nil), d.Tiers...)
}

// TierOf 返回品质标识对应的档位信息；未知品质返回一个以标识为名的档位。
func (d AuctionData) TierOf(key string) AuctionTier {
	for _, tier := range d.Tiers {
		if tier.Key == key {
			return tier
		}
	}
	return AuctionTier{Key: key, Name: key, Color: "808080"}
}

// ItemsOfTier 返回某品质下的拍品（已按价格降序）。
func (d AuctionData) ItemsOfTier(tier string) []AuctionItem {
	items := make([]AuctionItem, 0, len(d.Items))
	for _, item := range d.Items {
		if item.Tier == tier {
			items = append(items, item)
		}
	}
	return items
}

// TierPriceRange 返回某品质的价格区间；该品质为空时返回 (0, 0, false)。
func (d AuctionData) TierPriceRange(tier string) (int, int, bool) {
	items := d.ItemsOfTier(tier)
	if len(items) == 0 {
		return 0, 0, false
	}
	return items[len(items)-1].Price, items[0].Price, true
}

// PriceRange 返回整个奖池的价格区间。
func (d AuctionData) PriceRange() (int, int, bool) {
	if len(d.Items) == 0 {
		return 0, 0, false
	}
	minPrice, maxPrice := d.Items[0].Price, d.Items[0].Price
	for _, item := range d.Items {
		if item.Price < minPrice {
			minPrice = item.Price
		}
		if item.Price > maxPrice {
			maxPrice = item.Price
		}
	}
	return minPrice, maxPrice, true
}
