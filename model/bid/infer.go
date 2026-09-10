package bid

import "sort"

// maxCompositions 单次推测最多返回的组合数，避免组合空间过大时撑爆界面。
const maxCompositions = 20000

// searchBudget 一次推测允许搜索的节点数上限，兜住极端输入的耗时。
const searchBudget = 2_000_000

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

// InferQuery 一次推测的条件。
type InferQuery struct {
	// Avg 是玩家看到的单格均价（整数）。
	Avg int
	// MinCount、MaxCount 是组合总件数的区间。
	MinCount int
	MaxCount int
	// MaxTotal 是组合总价的上限，0 表示不限制。
	MaxTotal int
}

// TotalOption 一个可能的总价档位，以及能凑出它的件数范围。
type TotalOption struct {
	Total    int
	MinCount int
	MaxCount int
}

// CellItems 取出单格（1x1）物品，供推测使用。
func CellItems(grids []Grid) []Item {
	for _, grid := range grids {
		if grid.Length == 1 && grid.Width == 1 {
			return grid.Items
		}
	}

	return nil
}

// InferTotals 推出所有可能的总价档位。它只做算术：件数 n 的总价必然落在
// 「均价与 Avg 相差小于 1」对应的窗口内，且不低于已确认物品的总价、不超过上限，
// 因此不需要搜索就能列出全部价位。
func InferTotals(required []Item, query InferQuery) []TotalOption {
	if query.MinCount < 1 || query.MaxCount < query.MinCount {
		return nil
	}

	requiredTotal := 0

	for _, item := range required {
		requiredTotal += item.Value
	}

	if query.MaxTotal > 0 && requiredTotal > query.MaxTotal {
		return nil
	}

	start := query.MinCount
	if start < len(required) {
		start = len(required)
	}

	totals := make(map[int]TotalOption)

	for count := start; count <= query.MaxCount; count++ {
		low := windowLow(count, query.Avg)

		if low < requiredTotal {
			low = requiredTotal
		}

		high := windowHigh(count, query.Avg)

		if query.MaxTotal > 0 && high > query.MaxTotal {
			high = query.MaxTotal
		}

		for total := low; total <= high; total++ {
			option, ok := totals[total]
			if !ok {
				totals[total] = TotalOption{Total: total, MinCount: count, MaxCount: count}
				continue
			}

			if count < option.MinCount {
				option.MinCount = count
			}

			if count > option.MaxCount {
				option.MaxCount = count
			}

			totals[total] = option
		}
	}

	options := make([]TotalOption, 0, len(totals))

	for _, option := range totals {
		options = append(options, option)
	}

	sort.Slice(options, func(i, j int) bool { return options[i].Total < options[j].Total })

	return options
}

// InferAt 枚举指定总价下的所有组合：required 里的物品必须出现，其余位置从
// items 里补足，件数落在 query 的区间内。枚举按件数升序进行。
func InferAt(items, required []Item, query InferQuery, total int) InferResult {
	if len(items) == 0 || query.MinCount < 1 || query.MaxCount < query.MinCount ||
		len(required) > query.MaxCount {
		return InferResult{}
	}

	requiredTotal := 0

	for _, item := range required {
		requiredTotal += item.Value
	}

	if total < requiredTotal {
		return InferResult{}
	}

	if query.MaxTotal > 0 && total > query.MaxTotal {
		return InferResult{}
	}

	sorted := make([]Item, len(items))
	copy(sorted, items)
	sort.SliceStable(sorted, func(i, j int) bool { return sorted[i].Value > sorted[j].Value })

	state := &searchState{
		items:     sorted,
		required:  required,
		smallest:  sorted[len(sorted)-1].Value,
		query:     query,
		target:    total,
		baseCount: len(required),
		baseSum:   requiredTotal,
		budget:    searchBudget,
	}
	state.search()

	result := InferResult{Compositions: state.found, Truncated: state.truncated}

	// 同一总价下按件数升序。
	sort.SliceStable(result.Compositions, func(i, j int) bool {
		return result.Compositions[i].Count < result.Compositions[j].Count
	})

	return result
}

// searchState 保存一次推测的搜索状态。搜索出来的只是「补足部分」，真实件数与
// 真实总价要再加上已确认物品的 baseCount 与 baseSum。
type searchState struct {
	items    []Item
	required []Item
	smallest int
	query    InferQuery
	target   int

	baseCount int
	baseSum   int

	budget    int
	picked    []Item
	found     []Composition
	truncated bool
}

// search 逐件数枚举目标总价下的组合。
func (s *searchState) search() {
	start := s.query.MinCount
	if start < s.baseCount {
		start = s.baseCount
	}

	for count := start; count <= s.query.MaxCount; count++ {
		if s.target < windowLow(count, s.query.Avg) || s.target > windowHigh(count, s.query.Avg) {
			continue
		}

		s.collect(0, count-s.baseCount, s.target-s.baseSum)

		if s.truncated {
			return
		}
	}
}

// collect 从下标 start 起挑选 remainingCount 件物品，让它们正好凑出 remainingSum。
// 物品按价格不增排列，且只能往后选，因此每个组合只被走到一次。
func (s *searchState) collect(start, remainingCount, remainingSum int) {
	if s.truncated {
		return
	}

	if s.budget <= 0 || len(s.found) >= maxCompositions {
		s.truncated = true
		return
	}

	s.budget--

	if remainingCount == 0 {
		if remainingSum == 0 {
			s.found = append(s.found, buildComposition(s.required, s.picked, s.target))
		}

		return
	}

	if remainingSum < remainingCount*s.smallest {
		return
	}

	for i := start; i < len(s.items); i++ {
		item := s.items[i]

		if item.Value > remainingSum {
			continue
		}

		// 后面的物品只会更小，剩下的位置凑不到目标了。
		if item.Value*remainingCount < remainingSum {
			break
		}

		s.picked = append(s.picked, item)
		s.collect(i, remainingCount-1, remainingSum-item.Value)
		s.picked = s.picked[:len(s.picked)-1]

		if s.truncated {
			return
		}
	}
}

// windowLow、windowHigh 是「均价与 avg 相差小于 1」对应的总价闭区间。
func windowLow(count, avg int) int  { return count*(avg-1) + 1 }
func windowHigh(count, avg int) int { return count*(avg+1) - 1 }

// buildComposition 把已确认物品与补足物品合并后压成「物品 + 件数」，
// 同价同名的条目会并成一条。
func buildComposition(required, picked []Item, total int) Composition {
	all := make([]Item, 0, len(required)+len(picked))
	all = append(all, required...)
	all = append(all, picked...)

	sort.SliceStable(all, func(i, j int) bool {
		if all[i].Value != all[j].Value {
			return all[i].Value > all[j].Value
		}

		return all[i].Name < all[j].Name
	})

	groups := make([]CompositionItem, 0, len(all))

	for _, item := range all {
		last := len(groups) - 1

		if last >= 0 && groups[last].Item.Name == item.Name && groups[last].Item.Quality == item.Quality {
			groups[last].Count++
			continue
		}

		groups = append(groups, CompositionItem{Item: item, Count: 1})
	}

	return Composition{Items: groups, Count: len(all), Total: total}
}
