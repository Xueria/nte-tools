package bid

import "sort"

// maxCompositions 单次枚举最多返回的组合数，避免组合空间过大时撑爆界面。
const maxCompositions = 20000

// CompositionItem 组成里的一种物品及其件数。
type CompositionItem struct {
	Item  Item
	Count int
}

// Composition 一种可能的物品组成。
type Composition struct {
	Items []CompositionItem
	Count int
	Total int
}

// Average 返回该组合的平均价格。
func (c Composition) Average() float64 {
	if c.Count == 0 {
		return 0
	}

	return float64(c.Total) / float64(c.Count)
}

// InferResult 推测结果。
type InferResult struct {
	Compositions []Composition
	// Truncated 为 true 表示组合数或搜索量超过上限，只返回了能算完的那部分。
	Truncated bool
}

// PriceMode 用户给的价格是均价还是总价。
type PriceMode int

const (
	// AverageMode 按均价推测：组合的单件均价与填写值相差小于 1。
	AverageMode PriceMode = iota
	// TotalMode 按总价推测：组合总价正好等于填写值。
	TotalMode
)

// InferQuery 一次推测的条件。
type InferQuery struct {
	// Mode 与 Price 是价格条件：均价模式下 Price 是单件均价，
	// 总价模式下 Price 是组合总价。
	Mode  PriceMode
	Price int
	// MinCount、MaxCount 是组合总件数的区间。
	MinCount int
	MaxCount int
	// MaxTotal 是组合总价的上限，0 表示不限制；总价模式下总价已由 Price 定死。
	MaxTotal int
}

// InferCounts 返回确实存在组合的件数档位。件数 n 的总价必然落在价格条件给出的
// 窗口内，因此逐个件数做一次「只找一个解」的探测，没有解（或超过总价上限）的
// 件数不会返回，免得列出点了没内容的档位。
func InferCounts(items, required []Item, query InferQuery) []int {
	if len(items) == 0 || query.MinCount < 1 || query.MaxCount < query.MinCount ||
		len(required) > query.MaxCount {
		return nil
	}

	if query.MaxTotal > 0 && requiredTotal(required) > query.MaxTotal {
		return nil
	}

	start := query.MinCount
	if start < len(required) {
		start = len(required)
	}

	state := newSearchState(items, required, query)
	state.limit = 1

	counts := make([]int, 0, query.MaxCount-start+1)

	for count := start; count <= query.MaxCount; count++ {
		if state.probe(count) || state.budget <= 0 {
			// 预算用尽时无法判定，保守起见按有结果处理。
			counts = append(counts, count)
		}
	}

	return counts
}

// InferCount 枚举件数恰好为 count 的所有组合，按总价升序。
func InferCount(items, required []Item, query InferQuery, count int) InferResult {
	if len(items) == 0 || query.MinCount < 1 || query.MaxCount < query.MinCount ||
		len(required) > query.MaxCount || count < len(required) || count > query.MaxCount {
		return InferResult{}
	}

	if query.MaxTotal > 0 && requiredTotal(required) > query.MaxTotal {
		return InferResult{}
	}

	state := newSearchState(items, required, query)
	state.limit = maxCompositions
	state.searchWindow(count)

	result := InferResult{Compositions: state.found, Truncated: state.cut}

	// 同一件数下按总价升序。
	sort.SliceStable(result.Compositions, func(i, j int) bool {
		return result.Compositions[i].Total < result.Compositions[j].Total
	})

	return result
}
