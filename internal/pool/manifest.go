package pool

// PriceEntry 单次抽数的价格：每种资源需要花费多少。
type PriceEntry struct {
	Draw int            `json:"draw"`
	Cost map[string]int `json:"cost"`
}

// Manifest 盲盒池的定义：标识、总抽数以及逐抽价格表。
type Manifest struct {
	ID     string       `json:"id"`
	Name   string       `json:"name"`
	Draws  int          `json:"draws"`
	Prices []PriceEntry `json:"prices"`
}
