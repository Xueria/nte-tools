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

// CellItems 取出单格（1x1）物品，供推测使用。
func CellItems(grids []Grid) []Item {
	for _, grid := range grids {
		if grid.Length == 1 && grid.Width == 1 {
			return grid.Items
		}
	}

	return nil
}

// Infer 反推可能的物品组成：件数落在 [minCount, maxCount] 内，且组合的真实均价
// 与 avg 相差小于 1。游戏把均价显示成整数时，四舍五入、向下取整、向上取整三种
// 算法下的真实均价都落在 (avg-1, avg+1) 内，所以这个口径不会漏掉真实组成。
func Infer(items []Item, avg, minCount, maxCount int) InferResult {
	if len(items) == 0 || minCount < 1 || maxCount < minCount {
		return InferResult{}
	}

	sorted := make([]Item, len(items))
	copy(sorted, items)
	sort.SliceStable(sorted, func(i, j int) bool { return sorted[i].Value > sorted[j].Value })

	state := &searchState{
		items:    sorted,
		smallest: sorted[len(sorted)-1].Value,
		avg:      avg,
		minCount: minCount,
		maxCount: maxCount,
		budget:   searchBudget,
	}
	state.walk(0, 0)

	result := InferResult{Compositions: state.found, Truncated: state.truncated}

	// 先按件数、再按总价排序，让结果稳定可读。
	sort.SliceStable(result.Compositions, func(i, j int) bool {
		if result.Compositions[i].Count != result.Compositions[j].Count {
			return result.Compositions[i].Count < result.Compositions[j].Count
		}

		return result.Compositions[i].Total < result.Compositions[j].Total
	})

	return result
}

// searchState 保存一次推测的搜索状态。
type searchState struct {
	items    []Item
	smallest int
	avg      int
	minCount int
	maxCount int

	budget    int
	picked    []Item
	found     []Composition
	truncated bool
}

// walk 从下标 start 起继续挑选。物品按价格不增排列，且每次只能从当前下标往后
// 选，因此每种组合恰好被走到一次。
func (s *searchState) walk(start, total int) {
	if s.truncated {
		return
	}

	if s.budget <= 0 || len(s.found) >= maxCompositions {
		s.truncated = true
		return
	}

	s.budget--

	count := len(s.picked)

	if count >= s.minCount && withinAverage(total, count, s.avg) {
		s.found = append(s.found, buildComposition(s.picked, total))
	}

	if count == s.maxCount {
		return
	}

	for i := start; i < len(s.items); i++ {
		value := s.items[i].Value

		if !s.reachable(total+value, value) {
			continue
		}

		s.picked = append(s.picked, s.items[i])
		s.walk(i, total+value)
		s.picked = s.picked[:len(s.picked)-1]

		if s.truncated {
			return
		}
	}
}

// reachable 判断在 total 上再加若干件（每件不超过 maxValue、且不小于最小物品价）
// 之后，是否存在件数不超过 maxCount、均价又落进窗口的组合。
func (s *searchState) reachable(total, maxValue int) bool {
	count := len(s.picked) + 1

	for n := count; n <= s.maxCount; n++ {
		extra := n - count
		low := total + extra*s.smallest
		high := total + extra*maxValue

		if high < windowLow(n, s.avg) || low > windowHigh(n, s.avg) {
			continue
		}

		return true
	}

	return false
}

// withinAverage 判断件数为 count、总价为 total 的组合均价是否与 avg 相差小于 1。
func withinAverage(total, count, avg int) bool {
	return total >= windowLow(count, avg) && total <= windowHigh(count, avg)
}

// windowLow、windowHigh 是「均价与 avg 相差小于 1」对应的总价闭区间。
func windowLow(count, avg int) int  { return count*(avg-1) + 1 }
func windowHigh(count, avg int) int { return count*(avg+1) - 1 }

// buildComposition 把按价格降序排列的挑选结果压成「物品 + 件数」。
func buildComposition(picked []Item, total int) Composition {
	groups := make([]CompositionItem, 0, len(picked))

	for _, item := range picked {
		last := len(groups) - 1

		if last >= 0 && groups[last].Item.Name == item.Name && groups[last].Item.Quality == item.Quality {
			groups[last].Count++
			continue
		}

		groups = append(groups, CompositionItem{Item: item, Count: 1})
	}

	return Composition{Items: groups, Count: len(picked), Total: total}
}
