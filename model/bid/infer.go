package bid

import "sort"

// maxCompositions 单次枚举最多返回的组合数，避免组合空间过大时撑爆界面。
const maxCompositions = 20000

// searchBudget 一次搜索允许遍历的节点数上限，兜住极端输入的耗时。
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

// TotalOption 一个确实存在组合的总价档位，以及能凑出它的件数范围。
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

// InferTotals 推出确实存在组合的总价档位。候选档位用算术就能列出（件数 n 的总价
// 必然落在均价窗口内、不低于已确认物品总价、不超过上限），再逐个用一次「只找一个
// 解」的搜索确认：一个解都没有的档位不会返回，免得列出一堆点了没内容的价位。
func InferTotals(items, required []Item, query InferQuery) []TotalOption {
	candidates := candidateTotals(required, query)

	if len(candidates) == 0 || len(items) == 0 {
		return nil
	}

	sorted := make([]Item, len(items))
	copy(sorted, items)
	sort.SliceStable(sorted, func(i, j int) bool { return sorted[i].Value > sorted[j].Value })

	state := &searchState{
		items:     sorted,
		required:  required,
		smallest:  sorted[len(sorted)-1].Value,
		query:     query,
		baseCount: len(required),
		baseSum:   requiredTotal(required),
		budget:    searchBudget,
		limit:     1,
	}

	options := make([]TotalOption, 0, len(candidates))

	for i, candidate := range candidates {
		state.target = candidate.Total

		minCount, maxCount := 0, 0

		for count := candidate.MinCount; count <= candidate.MaxCount; count++ {
			if !state.probe(count) {
				continue
			}

			if minCount == 0 {
				minCount = count
			}

			maxCount = count
		}

		if minCount == 0 {
			// 预算用尽时无法判定，剩下的候选档位按算术结果原样保留。
			if state.budget <= 0 {
				options = append(options, candidates[i:]...)
				break
			}

			continue
		}

		options = append(options, TotalOption{Total: candidate.Total, MinCount: minCount, MaxCount: maxCount})
	}

	return options
}

// InferAt 枚举指定总价下的所有组合：required 里的物品必须出现，其余位置从
// items 里补足，件数落在 query 的区间内。枚举按件数升序进行。
func InferAt(items, required []Item, query InferQuery, total int) InferResult {
	if len(items) == 0 || query.MinCount < 1 || query.MaxCount < query.MinCount ||
		len(required) > query.MaxCount {
		return InferResult{}
	}

	base := requiredTotal(required)

	if total < base {
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
		baseSum:   base,
		budget:    searchBudget,
		limit:     maxCompositions,
	}
	state.search()

	result := InferResult{Compositions: state.found, Truncated: state.cut}

	// 同一总价下按件数升序。
	sort.SliceStable(result.Compositions, func(i, j int) bool {
		return result.Compositions[i].Count < result.Compositions[j].Count
	})

	return result
}

// candidateTotals 用算术列出候选价位，并记录能凑出该价位的件数范围。
func candidateTotals(required []Item, query InferQuery) []TotalOption {
	if query.MinCount < 1 || query.MaxCount < query.MinCount {
		return nil
	}

	base := requiredTotal(required)

	if query.MaxTotal > 0 && base > query.MaxTotal {
		return nil
	}

	start := query.MinCount
	if start < len(required) {
		start = len(required)
	}

	totals := make(map[int]TotalOption)

	for count := start; count <= query.MaxCount; count++ {
		low := windowLow(count, query.Avg)

		if low < base {
			low = base
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

// requiredTotal 返回已确认物品的总价。
func requiredTotal(items []Item) int {
	total := 0

	for _, item := range items {
		total += item.Value
	}

	return total
}

// searchState 保存一次搜索的状态。搜索出来的只是「补足部分」，真实件数与真实
// 总价要再加上已确认物品的 baseCount 与 baseSum。
type searchState struct {
	items    []Item
	required []Item
	smallest int
	query    InferQuery
	target   int

	baseCount int
	baseSum   int

	budget int
	// limit 是本次运行最多收集多少条；cut 表示因为达到 limit 或预算耗尽提前停止。
	limit int
	cut   bool

	picked []Item
	found  []Composition
}

// search 逐件数枚举目标总价下的组合。
func (s *searchState) search() {
	start := s.query.MinCount
	if start < s.baseCount {
		start = s.baseCount
	}

	for count := start; count <= s.query.MaxCount; count++ {
		s.searchCount(count)

		if s.cut {
			return
		}
	}
}

// searchCount 枚举件数恰好为 count 的目标组合。
func (s *searchState) searchCount(count int) {
	if s.cut || count < s.baseCount || s.target < s.baseSum {
		return
	}

	if s.target < windowLow(count, s.query.Avg) || s.target > windowHigh(count, s.query.Avg) {
		return
	}

	s.collect(0, count-s.baseCount, s.target-s.baseSum)
}

// probe 判断「件数为 count、总价为 target」是否至少存在一个组合。
func (s *searchState) probe(count int) bool {
	if s.budget <= 0 {
		return false
	}

	s.found = s.found[:0]
	s.cut = false
	s.searchCount(count)

	return len(s.found) > 0
}

// collect 从下标 start 起挑选 remainingCount 件物品，让它们正好凑出 remainingSum。
// 物品按价格不增排列，且只能往后选，因此每个组合只被走到一次。
func (s *searchState) collect(start, remainingCount, remainingSum int) {
	if s.cut {
		return
	}

	if len(s.found) >= s.limit || s.budget <= 0 {
		s.cut = true
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

		if s.cut {
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
