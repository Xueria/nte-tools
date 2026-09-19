package bid

// Listing 一份拍品清单：名称、属性与拍品。名称与属性都能由文件名补齐，
// 因此文件里只写 items 也是一份合法的普通拍品清单。
type Listing struct {
	Name      string    `json:"name"`
	Attribute Attribute `json:"attribute"`
	Items     []Item    `json:"items"`
}
