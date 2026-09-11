package bid

// Item 一件拍品：名称、品质与成交价格。
type Item struct {
	Name    string `json:"name"`
	Quality string `json:"quality"`
	Value   int    `json:"value"`
}

// Attribute 一份清单在仓库里的属性：Grid 表示这批拍品是否占格，占格时长为
// Length、宽为 Width；非占格（普通拍品）时长宽都取 1。
type Attribute struct {
	Grid   bool `json:"grid"`
	Length int  `json:"length"`
	Width  int  `json:"width"`
}

// PlainAttribute 返回普通拍品（不占格）的属性。
func PlainAttribute() Attribute {
	return Attribute{Grid: false, Length: 1, Width: 1}
}

// GridAttribute 返回占格 length×width 的属性。
func GridAttribute(length, width int) Attribute {
	return Attribute{Grid: true, Length: length, Width: width}
}

// IsZero 判断属性是否完全缺省，即数据里没写 attribute。
func (a Attribute) IsZero() bool {
	return !a.Grid && a.Length == 0 && a.Width == 0
}

// Normalize 返回补齐后的属性：非占格即降级为普通拍品，占格时长宽缺省按 1。
func (a Attribute) Normalize() Attribute {
	if !a.Grid {
		return PlainAttribute()
	}

	return GridAttribute(sizeOrOne(a.Length), sizeOrOne(a.Width))
}

// Cell 判断属性是不是单格（1x1）：普通拍品虽然长宽也是 1，但不算占格单格。
func (a Attribute) Cell() bool {
	return a.Grid && a.Length == 1 && a.Width == 1
}

// sizeOrOne 把缺省（非正）的格数按 1 处理。
func sizeOrOne(size int) int {
	if size < 1 {
		return 1
	}

	return size
}

// Listing 一份拍品清单：名称、属性与拍品。名称与属性都能由文件名补齐，
// 因此文件里只写 items 也是一份合法的普通拍品清单。
type Listing struct {
	Name      string    `json:"name"`
	Attribute Attribute `json:"attribute"`
	Items     []Item    `json:"items"`
}
