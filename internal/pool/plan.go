package pool

import (
	"maps"
	"math"
	"sort"
)

// PlanStep 一次抽数怎么付：用哪种资源、花多少、付完还剩多少。
type PlanStep struct {
	Draw         int
	ResourceID   string
	ResourceName string
	Cost         int
	// Remaining 是这次抽数之后该资源的剩余量。
	Remaining int
}

// PlanResult 一次消耗方案计算的结果。
type PlanResult struct {
	Steps []PlanStep
	// Final 是每种资源的最终余额。
	Final        map[string]int
	Insufficient bool
	FailAtDraw   int
}

// Plan 计算 [start, end] 抽数区间（含两端）的最优消耗方案，目标按字典序依次为：
//
//  1. 完成的抽数最多；
//  2. 优先保留的资源（preferredID 非空时）花得最少；
//  3. 各资源剩余的比例最接近。
//
// 抽数不多，因此枚举每种资源分配（每次抽数各自选一种可支付的资源）后取最优，
// 得到的一定是全局最优解。
func (p Pool) Plan(start, end int, balances map[string]int, preferredID string) PlanResult {
	costs := drawCosts(p)
	order := resourceOrder(p)

	// 至少要有一种支付方式的抽数才算数：没有价格的抽数付不了，
	// 方案在它之前就停住。
	var draws []int

	unpayableDraw := 0

	for draw := start; draw <= end; draw++ {
		if len(costs[draw]) == 0 {
			unpayableDraw = draw
			break
		}

		draws = append(draws, draw)
	}

	result := PlanResult{Final: make(map[string]int, len(balances))}
	maps.Copy(result.Final, balances)

	if len(draws) == 0 {
		result.Insufficient = true
		result.FailAtDraw = start

		return result
	}

	best := candidate{completed: -1}
	bestChoices := make([]string, len(draws))
	choices := make([]string, len(draws))

	var search func(index int)

	search = func(index int) {
		if index == len(draws) {
			current := evaluateCandidate(draws, choices, costs, balances, preferredID)

			if current.better(best) {
				best = current
				copy(bestChoices, choices)
			}

			return
		}

		for _, id := range drawResources(draws[index], costs, order) {
			choices[index] = id
			search(index + 1)
		}
	}
	search(0)

	// 按最优分配重建每一步的结果。
	for i := 0; i < best.completed && i < len(draws); i++ {
		id := bestChoices[i]
		cost := costs[draws[i]][id]
		result.Final[id] -= cost
		result.Steps = append(result.Steps, PlanStep{
			Draw:         draws[i],
			ResourceID:   id,
			ResourceName: p.ResourceName(id),
			Cost:         cost,
			Remaining:    result.Final[id],
		})
	}

	if best.completed < len(draws) {
		result.Insufficient = true
		result.FailAtDraw = draws[best.completed]
	} else if unpayableDraw != 0 {
		result.Insufficient = true
		result.FailAtDraw = unpayableDraw
	}

	return result
}

// candidate 一种资源分配的得分。
type candidate struct {
	completed      int
	preferredSpent int
	imbalance      float64
}

// better 判断 current 是否严格优于 other，比较顺序即方案的目标顺序：
// 抽数更多，其次优先保留的资源花得更少，最后剩余比例更均衡。
func (c candidate) better(other candidate) bool {
	if c.completed != other.completed {
		return c.completed > other.completed
	}

	if c.preferredSpent != other.preferredSpent {
		return c.preferredSpent < other.preferredSpent
	}

	return c.imbalance < other.imbalance
}

// evaluateCandidate 走一遍资源分配（choices[i] 支付 draws[i]），返回它能完成多少抽、
// 以及得分。
func evaluateCandidate(draws []int, choices []string, costs map[int]map[string]int,
	balances map[string]int, preferredID string) candidate {
	spent := make(map[string]int, len(balances))
	completed := 0

	for i, id := range choices {
		cost := costs[draws[i]][id]

		if spent[id]+cost > balances[id] {
			break
		}

		spent[id] += cost
		completed++
	}

	preferredSpent := 0

	if preferredID != "" {
		preferredSpent = spent[preferredID]
	}

	return candidate{
		completed:      completed,
		preferredSpent: preferredSpent,
		imbalance:      imbalanceOf(balances, spent),
	}
}

// imbalanceOf 衡量各资源剩余量的不均衡程度，取值在 [0, 1]：0 表示每种资源
// 剩余的比例都一样。
func imbalanceOf(balances, spent map[string]int) float64 {
	minFraction := math.MaxFloat64
	maxFraction := -math.MaxFloat64

	for id, balance := range balances {
		if balance <= 0 {
			continue
		}

		fraction := float64(balance-spent[id]) / float64(balance)

		if fraction < minFraction {
			minFraction = fraction
		}

		if fraction > maxFraction {
			maxFraction = fraction
		}
	}

	if minFraction == math.MaxFloat64 {
		return 0
	}

	return maxFraction - minFraction
}

// drawCosts 把价格表整理成「抽数 → 资源标识 → 花费」，抽数从 1 开始。
func drawCosts(p Pool) map[int]map[string]int {
	costs := make(map[int]map[string]int, len(p.Manifest.Prices))

	for _, price := range p.Manifest.Prices {
		table := make(map[string]int, len(price.Cost))
		maps.Copy(table, price.Cost)
		costs[price.Draw] = table
	}

	return costs
}

// resourceOrder 返回所有资源标识的固定顺序：先声明过的池资源，
// 再补上只在价格表里出现过的资源。
func resourceOrder(p Pool) []string {
	seen := make(map[string]bool, len(p.Resources)+1)
	order := make([]string, 0, len(p.Resources)+1)

	for _, resource := range p.Resources {
		order = append(order, resource.ID)
		seen[resource.ID] = true
	}

	for _, price := range p.Manifest.Prices {
		for id := range price.Cost {
			if !seen[id] {
				order = append(order, id)
				seen[id] = true
			}
		}
	}

	return order
}

// drawResources 返回某次抽数可用的资源，按 resourceOrder 的顺序。
func drawResources(draw int, costs map[int]map[string]int, order []string) []string {
	table := costs[draw]
	available := make([]string, 0, len(table))

	for _, id := range order {
		if _, ok := table[id]; ok {
			available = append(available, id)
		}
	}

	return available
}

// ResourceName 返回资源标识对应的展示名；池里没声明的标识原样返回。
func (p Pool) ResourceName(id string) string {
	for _, resource := range p.Resources {
		if resource.ID == id {
			return resource.Name
		}
	}

	return id
}

// SortedResourceIDs 返回按标识升序排列的资源列表，用于稳定地展示最终余额。
func (p Pool) SortedResourceIDs() []string {
	ids := make([]string, 0, len(p.Resources))

	for _, resource := range p.Resources {
		ids = append(ids, resource.ID)
	}

	sort.Strings(ids)

	return ids
}
