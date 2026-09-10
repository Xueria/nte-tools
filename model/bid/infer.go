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

// Infer 反推可能的物品组成：required 里的物品是玩家已确认在里面的，必须出现在
// 每个结果中；其余位置从 items 里补足。总件数落在 [minCount, maxCount] 内，且
// 组合的真实均价与 avg 相差小于 1。游戏把均价显示成整数时，四舍五入、向下取整、
// 向上取整三种算法下的真实均价都落在 (avg-1, avg+1) 内，所以这个口径不会漏掉
// 真实组成。
func Infer(items, required []Item, avg, minCount, maxCount int) InferResult {
	if len(items) == 0 || minCount < 1 || maxCount < minCount || len(required) > maxCount {
		return InferResult{}
	}

	sorted := make([]Item, len(items))
	copy(sorted, items)
	sort.SliceStable(sorted, func(i, j int) bool { return sorted[i].Value > sorted[j].Value })

	requiredTotal := 0

	for _, item := range required {
		requiredTotal += item.Value
	}

	state := &searchState{
		items:     sorted,
		required:  required,
		smallest:  sorted[len(sorted)-1].Value,
		avg:       avg,
		minCount:  minCount,
		maxCount:  maxCount,
		baseCount: len(required),
		baseSum:   requiredTotal,
		budget:    searchBudget,
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

// searchState 保存一次推测的搜索状态。walk 遍历的是「补足部分」，真实件数与
// 真实总价要再加上已确认物品的 baseCount 与 baseSum。
type searchState struct {
	items     []Item
	required  []Item
	smallest  int
	avg       int
	minCount  int
	maxCount  int
	baseCount int
	baseSum   int

	budget    int
	picked    []Item
	found     []Composition
	truncated bool
}

// walk 从下标 start 起继续挑选补足物品。物品按价格不增排列，且每次只能从当前
// 下标往后选，因此每种组合恰好被走到一次。
func (s *searchState) walk(start, total int) {
	if s.truncated {
		return
	}

	if s.budget <= 0 || len(s.found) >= maxCompositions {
		s.truncated = true
		return
	}

	s.budget--

	count := s.baseCount + len(s.picked)
	sum := s.baseSum + total

	if count >= s.minCount && withinAverage(sum, count, s.avg) {
		s.found = append(s.found, buildComposition(s.required, s.picked, sum))
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

// reachable 判断在补足部分 total 上再加若干件（每件不超过 maxValue、且不小于
// 最小物品价）之后，是否存在件数不超过 maxCount、均价又落进窗口的组合。
func (s *searchState) reachable(total, maxValue int) bool {
	count := s.baseCount + len(s.picked) + 1

	for n := count; n <= s.maxCount; n++ {
		extra := n - count
		low := s.baseSum + total + extra*s.smallest
		high := s.baseSum + total + extra*maxValue

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
