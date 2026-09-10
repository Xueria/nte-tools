package model

// AuctionItem 是 1x1 奖池里的一件拍品：名字、价格、品质。
type AuctionItem struct {
	Tier  string `json:"tier"`
	Name  string `json:"name"`
	Price int    `json:"price"`
}

// AuctionTier 是品质档位：名字与展示颜色由数据文件给出。
type AuctionTier struct {
	Key   string `json:"key"`
	Name  string `json:"name"`
	Color string `json:"color"` // RGB 十六进制，用于界面上的色条
}

// AuctionData 是 1x1 奖池的全部数据。
type AuctionData struct {
	// Tiers 按数据文件里的声明顺序，也就是稀有度从高到低。
	Tiers []AuctionTier
	// Items 已按档位顺序、档位内按价格降序排列。
	Items []AuctionItem
}
