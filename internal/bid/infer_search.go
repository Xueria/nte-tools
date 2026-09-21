package bid

import "sort"

// searchBudget 一次搜索允许遍历的节点数上限，兜住极端输入的耗时。
const searchBudget = 2_000_000

// newSearchState 按候选物品与查询条件准备一次搜索。
func newSearchState(items, required []Item, query InferQuery) *searchState {
	sorted := make([]Item, len(items))
	copy(sorted, items)
	sort.SliceStable(sorted, func(i, j int) bool { return sorted[i].Value > sorted[j].Value })

	return &searchState{
		items:     sorted,
		required:  required,
		smallest:  sorted[len(sorted)-1].Value,
		query:     query,
		baseCount: len(required),
		baseSum:   requiredTotal(required),
		budget:    searchBudget,
	}
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

// searchWindow 在件数 count 的总价窗口内逐档枚举组合，窗口从小到大。
func (s *searchState) searchWindow(count int) {
	low, high := countWindow(count, s.baseSum, s.query)

	for total := low; total <= high; total++ {
		s.target = total
		s.collect(0, count-s.baseCount, total-s.baseSum)

		if s.cut {
			return
		}
	}
}

// probe 判断件数 count 下是否至少存在一个组合。
func (s *searchState) probe(count int) bool {
	if s.budget <= 0 {
		return false
	}

	low, high := countWindow(count, s.baseSum, s.query)

	for total := low; total <= high; total++ {
		s.found = s.found[:0]
		s.cut = false
		s.target = total
		s.collect(0, count-s.baseCount, total-s.baseSum)

		if len(s.found) > 0 {
			return true
		}

		if s.budget <= 0 {
			return false
		}
	}

	return false
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

// countWindow 返回件数 count 可能的总价区间（已按已确认总价与上限裁剪）。
func countWindow(count, base int, query InferQuery) (low, high int) {
	low, high = priceWindow(count, query)

	if low < base {
		low = base
	}

	if query.MaxTotal > 0 && high > query.MaxTotal {
		high = query.MaxTotal
	}

	return low, high
}

// priceWindow 返回件数 count 对应的总价闭区间：均价模式下按「均价相差小于 1」
// 换算，总价模式下就是填写的那个总价。
func priceWindow(count int, query InferQuery) (low, high int) {
	if query.Mode == TotalMode {
		return query.Price, query.Price
	}

	return windowLow(count, query.Price), windowHigh(count, query.Price)
}

// windowLow、windowHigh 是「均价与 avg 相差小于 1」对应的总价闭区间。
func windowLow(count, avg int) int  { return count*(avg-1) + 1 }
func windowHigh(count, avg int) int { return count*(avg+1) - 1 }

// requiredTotal 返回已确认物品的总价。
func requiredTotal(items []Item) int {
	total := 0

	for _, item := range items {
		total += item.Value
	}

	return total
}

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
