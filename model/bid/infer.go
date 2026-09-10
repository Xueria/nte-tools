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
	// Truncated 为 true 表示组合数或搜索量超过上限，只返回了总价最低的那部分。
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
	// MaxPerTotal 是每个总价最多保留多少种组合；小于等于 0 表示不限制。
	MaxPerTotal int
	// PriorityQuality 是优先关注的品质：每个总价先用含它的组合填充，再用不含的
	// 补足。为空字符串表示不区分。
	PriorityQuality string
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
// 每个结果中；其余位置从 items 里补足。总件数落在 query 的区间内，总价不超过
// query.MaxTotal，且组合的真实均价与 query.Avg 相差小于 1。游戏把均价显示成整数
// 时，四舍五入、向下取整、向上取整三种算法下的真实均价都落在 (avg-1, avg+1) 内，
// 所以这个口径不会漏掉真实组成。
//
// 枚举顺序是「件数升序 × 总价升序」：先算便宜的，所以一旦触到组合数或搜索量
// 上限，被截掉的只会是更贵的那部分，不会出现"列出来的全是贵的"。
func Infer(items, required []Item, query InferQuery) InferResult {
	if len(items) == 0 || query.MinCount < 1 || query.MaxCount < query.MinCount ||
		len(required) > query.MaxCount {
		return InferResult{}
	}

	sorted := make([]Item, len(items))
	copy(sorted, items)
	sort.SliceStable(sorted, func(i, j int) bool { return sorted[i].Value > sorted[j].Value })

	requiredTotal := 0

	for _, item := range required {
		requiredTotal += item.Value
	}

	if query.MaxTotal > 0 && requiredTotal > query.MaxTotal {
		return InferResult{}
	}

	perTotal := query.MaxPerTotal
	if perTotal <= 0 {
		perTotal = maxCompositions
	}

	state := &searchState{
		items:        sorted,
		required:     required,
		smallest:     sorted[len(sorted)-1].Value,
		query:        query,
		baseCount:    len(required),
		baseSum:      requiredTotal,
		budget:       searchBudget,
		perTotal:     perTotal,
		priority:     query.PriorityQuality,
		kept:         make(map[int]int),
		priorityKept: make(map[int]int),
	}
	state.search()

	result := InferResult{Compositions: state.found, Truncated: state.truncated}

	// 总价从低到高；同一总价里优先品质在前，再按件数。
	sort.SliceStable(result.Compositions, func(i, j int) bool {
		left, right := result.Compositions[i], result.Compositions[j]

		if left.Total != right.Total {
			return left.Total < right.Total
		}

		leftPriority := compositionHasQuality(left, query.PriorityQuality)
		rightPriority := compositionHasQuality(right, query.PriorityQuality)

		if leftPriority != rightPriority {
			return leftPriority
		}

		return left.Count < right.Count
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

	baseCount int
	baseSum   int

	budget       int
	perTotal     int
	priority     string
	kept         map[int]int
	priorityKept map[int]int
	passLimit    int

	picked    []Item
	found     []Composition
	truncated bool
}

// search 按件数升序、总价升序逐个目标枚举。
func (s *searchState) search() {
	start := s.query.MinCount
	if start < s.baseCount {
		start = s.baseCount
	}

	basePriority := hasQuality(s.required, s.priority)

	for count := start; count <= s.query.MaxCount; count++ {
		target := windowLow(count, s.query.Avg)

		if target < s.baseSum {
			target = s.baseSum
		}

		high := windowHigh(count, s.query.Avg)

		if s.query.MaxTotal > 0 && high > s.query.MaxTotal {
			high = s.query.MaxTotal
		}

		for ; target <= high; target++ {
			s.collectTotal(count, target, basePriority)

			if s.truncated {
				return
			}
		}
	}
}

// collectTotal 为某个总价收集组合：先用含优先品质的组合填充，再用不含的补足，
// 每个总价最多保留 perTotal 种。
func (s *searchState) collectTotal(count, target int, basePriority bool) {
	if s.kept[target] >= s.perTotal {
		return
	}

	if s.priority != "" && s.priorityKept[target] < s.perTotal {
		s.passLimit = s.perTotal - s.priorityKept[target]
		s.collect(0, count-s.baseCount, target-s.baseSum, target, true, basePriority)

		if s.truncated || s.kept[target] >= s.perTotal {
			return
		}
	}

	s.passLimit = s.perTotal - s.kept[target]
	s.collect(0, count-s.baseCount, target-s.baseSum, target, false, basePriority)
}

// collect 从下标 start 起挑选 remainingCount 件物品，让它们正好凑出 remainingSum。
// 物品按价格不增排列，且只能往后选，因此每个组合只被走到一次。wantPriority 为
// true 时只收含优先品质的组合，false 时只收不含的，两个 pass 互不重叠。
func (s *searchState) collect(start, remainingCount, remainingSum, target int, wantPriority, hasPriority bool) {
	if s.truncated || s.passLimit <= 0 {
		return
	}

	if s.budget <= 0 || len(s.found) >= maxCompositions {
		s.truncated = true
		return
	}

	s.budget--

	if remainingCount == 0 {
		if remainingSum == 0 && hasPriority == wantPriority {
			s.record(target, hasPriority)
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

		nextPriority := hasPriority || (s.priority != "" && item.Quality == s.priority)

		s.picked = append(s.picked, item)
		s.collect(i, remainingCount-1, remainingSum-item.Value, target, wantPriority, nextPriority)
		s.picked = s.picked[:len(s.picked)-1]

		if s.truncated || s.passLimit <= 0 {
			return
		}
	}
}

// record 记下一条组合，并更新该总价的计数。
func (s *searchState) record(target int, hasPriority bool) {
	s.found = append(s.found, buildComposition(s.required, s.picked, target))
	s.kept[target]++
	s.passLimit--

	if hasPriority {
		s.priorityKept[target]++
	}
}

// hasQuality 判断物品里是否含指定品质。
func hasQuality(items []Item, quality string) bool {
	if quality == "" {
		return false
	}

	for _, item := range items {
		if item.Quality == quality {
			return true
		}
	}

	return false
}

// compositionHasQuality 判断组合里是否含指定品质。
func compositionHasQuality(composition Composition, quality string) bool {
	if quality == "" {
		return false
	}

	for _, group := range composition.Items {
		if group.Item.Quality == quality {
			return true
		}
	}

	return false
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
