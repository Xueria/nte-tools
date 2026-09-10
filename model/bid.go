package model

// BidItem 竞拍单格中的一件拍品：名称、品质与该格的成交价格。
type BidItem struct {
	Name    string `json:"name"`
	Quality string `json:"quality"`
	Value   int    `json:"value"`
}
