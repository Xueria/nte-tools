package bid

// BidItem 竞拍单格中的一件拍品：名称、品质与该格的成交价格。
type BidItem struct {
	Name    string `json:"name"`
	Quality string `json:"quality"`
	Value   int    `json:"value"`
}

// BidGrid 一组占格尺寸相同的拍品：物品在仓库中各占据 length×width 格，
// 与文件名 <length>x<width>.json 对应。
type BidGrid struct {
	Length int       `json:"length"`
	Width  int       `json:"width"`
	Items  []BidItem `json:"items"`
}
