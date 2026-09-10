package model

// BidItem 竞拍单格中的一件拍品：标识、名称与该格的成交价格。
type BidItem struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Value int    `json:"value"`
}
