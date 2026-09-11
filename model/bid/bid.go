package bid

// Item 一件拍品：名称、品质与成交价格。
type Item struct {
	Name    string `json:"name"`
	Quality string `json:"quality"`
	Value   int    `json:"value"`
}

// Footprint 一批拍品在仓库里占据的形状：Grid 为 true 时占 Length×Width 格，
// 为 false 表示这批拍品没有占格信息，按普通拍品处理，长宽都取 1。
type Footprint struct {
	// Grid 由长宽推导，不写在数据文件里，缺省为零值，加载时经 NewFootprint 补齐。
	Grid   bool `json:"-"`
	Length int  `json:"length"`
	Width  int  `json:"width"`
}

// NewFootprint 由长宽推导形状：长宽都为正才算占格类型，否则降级为普通拍品。
func NewFootprint(length, width int) Footprint {
	if length < 1 || width < 1 {
		return PlainFootprint()
	}

	return Footprint{Grid: true, Length: length, Width: width}
}

// PlainFootprint 返回普通拍品（不占格）的形状。
func PlainFootprint() Footprint {
	return Footprint{Grid: false, Length: 1, Width: 1}
}

// Cell 判断形状是不是单格（1x1）：普通拍品虽然长宽也是 1，但不算占格单格。
func (f Footprint) Cell() bool {
	return f.Grid && f.Length == 1 && f.Width == 1
}

// Listing 一份拍品清单：名称、拍品，以及可以缺省的占格形状。名称与长宽都能
// 由文件名补齐，因此文件里只写 items 也是一份合法的普通拍品清单。
type Listing struct {
	Name string `json:"name"`
	// Footprint 内嵌，长宽与 Items 一样平铺在文件顶层，方便手写与向后兼容。
	Footprint
	Items []Item `json:"items"`
}
