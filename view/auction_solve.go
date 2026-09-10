package view

import (
	"blind-tools/model"
	"fmt"
	"sort"
	"strings"
)

// DefaultMaxDepth 是「单格最大可能件数」的默认值。
const DefaultMaxDepth = 30

const (
	maxDepth    = 60 // 件数上限，防误输入把搜索撑爆
	maxDistinct = 4  // 一个构成里最多允许几种不同拍品
)

// ComboItem 是构成里的一件拍品及其出现次数。
type ComboItem struct {
	Item  model.AuctionItem
	Count int
}

// SlotCombo 是一条「这一格由哪些拍品组成」的猜测。
type SlotCombo struct {
	Items    []ComboItem
	Count    int // 件数 k
	Total    int // 合计价格 = k × 平均价
	Distinct int // 用到了几种不同拍品
}

// SlotGuess 是一次推测的完整结果。
type SlotGuess struct {
	Average   int // 输入的单格平均价
	MaxDepth  int // 输入的最大可能件数
	PoolSize  int // 奖池拍品数
	MinPrice  int
	MaxPrice  int
	Depths    []int // 哪些件数能凑出这个平均价（升序）
	Combos    []SlotCombo
	Nearest   []int // 一个解都没有时，附近最接近的可行均价
	Truncated bool  // 解太多，只展示了一部分
}

// searchSlotComposition 反推「平均价 average」的一格可能是由多少件、哪些拍品组成。
//
// 件数 k 是未知的（只知道不超过 maxDepth），所以对每个 k = 1..maxDepth 分别求解：
// 需要 n1 + n2 = k 且 n1·p1 + n2·p2 = k·average，即「p1 比均价低多少、
// p2 比均价高多少」要恰好抵消。于是对每对跨越均价的拍品做一次整数判定即可，
// 只有极少几种拍品就能精确命中。
func searchSlotComposition(data model.AuctionData, average, maxDepth, maxResults int) SlotGuess {
	if maxDepth <= 0 {
		maxDepth = DefaultMaxDepth
	}
	if maxDepth > maxDepthLimit() {
		maxDepth = maxDepthLimit()
	}
	if maxResults <= 0 || maxResults > 200 {
		maxResults = 200
	}

	guess := SlotGuess{Average: average, MaxDepth: maxDepth, PoolSize: len(data.Items)}
	if average <= 0 || len(data.Items) == 0 {
		return guess
	}

	items := append([]model.AuctionItem(nil), data.Items...)
	sort.SliceStable(items, func(i, j int) bool { return items[i].Price > items[j].Price })

	guess.MinPrice = items[len(items)-1].Price
	guess.MaxPrice = items[0].Price
	if average < guess.MinPrice || average > guess.MaxPrice {
		// 均价超出奖池单价范围，任何件数都凑不出。
		return guess
	}

	seen := map[string]bool{}
	for k := 1; k <= maxDepth; k++ {
		solutions := solveForCount(data, items, average, k, maxDistinct)
		if len(solutions) == 0 {
			continue
		}
		guess.Depths = append(guess.Depths, k)

		// 每个件数只保留最有代表性的几条，避免同一个 k 刷屏。
		// maxResults 很小时（命令行探针）就别把每个 k 都铺开了。
		perDepth := 2
		if maxResults <= 8 {
			perDepth = 1
		}
		kept := 0
		for _, combo := range solutions {
			key := comboKey(combo)
			if seen[key] {
				continue
			}
			seen[key] = true
			guess.Combos = append(guess.Combos, combo)
			kept++
			if kept >= perDepth {
				break
			}
		}
		if len(guess.Combos) >= maxResults {
			if k < maxDepth {
				guess.Truncated = true
			}
			break
		}
	}

	sortSlotCombos(guess.Combos)
	if len(guess.Combos) > maxResults {
		guess.Combos = guess.Combos[:maxResults]
		guess.Truncated = true
	}

	// 一个解都没有时，给出附近最接近的可行均价，避免只回一句「没结果」。
	if len(guess.Depths) == 0 {
		guess.Nearest = nearestAverages(data, average, maxDepth, 40, 5)
	}
	return guess
}

// nearestAverages 在 target 附近找出真正能凑出的均价，用于「没有精确解」时的提示。
func nearestAverages(data model.AuctionData, target, maxDepth, span, limit int) []int {
	items := append([]model.AuctionItem(nil), data.Items...)
	sort.SliceStable(items, func(i, j int) bool { return items[i].Price > items[j].Price })

	var out []int
	for delta := 1; delta <= span && len(out) < limit; delta++ {
		for _, candidate := range []int{target - delta, target + delta} {
			if candidate <= 0 {
				continue
			}
			if hasSolutionForAverage(data, items, candidate, maxDepth) {
				out = append(out, candidate)
				if len(out) >= limit {
					break
				}
			}
		}
	}
	sort.Ints(out)
	return out
}

// hasSolutionForAverage 判断某个均价是否存在任何件数下的精确构成。
// 只用两件拍品的判定，够快；三件构成是少数情况，这里不做以免拖慢提示。
func hasSolutionForAverage(data model.AuctionData, items []model.AuctionItem, average, maxDepth int) bool {
	if len(items) == 0 || average < items[len(items)-1].Price || average > items[0].Price {
		return false
	}
	for k := 1; k <= maxDepth; k++ {
		if len(solveWithPairs(data, items, average, k)) > 0 {
			return true
		}
	}
	return false
}

// maxDepthLimit 返回件数上限。
func maxDepthLimit() int { return maxDepth }

// solveForCount 求出「恰好 k 件、均价恰好为 average」的拍品构成。
//
// 先用两件拍品求解（覆盖率已经很高）：
//
//	n1·p1 + n2·p2 = k·average，n1 + n2 = k
//
// 要求 (p1-average) 与 (p2-average) 反号，且解出来的次数是整数。
// 若两件无解，再用三件拍品做一次有界搜索。
func solveForCount(data model.AuctionData, items []model.AuctionItem, average, k, maxDistinct int) []SlotCombo {
	solutions := solveWithPairs(data, items, average, k)

	if len(solutions) == 0 && maxDistinct >= 3 {
		solutions = solveWithTriples(data, items, average, k)
	}
	return solutions
}

// solveWithPairs 用两件拍品求解，返回所有可行构成（按次数分布合理度排序）。
func solveWithPairs(data model.AuctionData, items []model.AuctionItem, average, k int) []SlotCombo {
	var out []SlotCombo

	for i := 0; i < len(items); i++ {
		p1 := items[i].Price
		if p1 > average {
			continue // p1 必须是「比均价便宜」的那件
		}
		d1 := average - p1

		for j := i + 1; j < len(items); j++ {
			p2 := items[j].Price
			if p2 <= average {
				break // items 按价格降序，后面的只会更便宜
			}
			d2 := p2 - average

			// n2 = k·d1 / (d1 + d2)，n1 = k - n2，两者都要是整数且非负。
			num := k * d1
			den := d1 + d2
			if num%den != 0 {
				continue
			}
			n2 := num / den
			n1 := k - n2
			if n1 < 0 || n2 < 0 {
				continue
			}
			if n1 == 0 && n2 == 0 {
				continue
			}

			out = append(out, buildSlotCombo([]ComboItem{
				{Item: items[i], Count: n1},
				{Item: items[j], Count: n2},
			}, average, k))
		}
	}

	sortComboCandidates(out)
	return out
}

// solveWithTriples 用三件拍品求解：固定第一件的次数，剩下的件数用两件拍品的
// 整数判定解决（要求剩下两件的均价恰好落在它们的单价之间）。
func solveWithTriples(data model.AuctionData, items []model.AuctionItem, average, k int) []SlotCombo {
	var out []SlotCombo

	for i := 0; i < len(items) && len(out) < 8; i++ {
		p1 := items[i].Price
		if p1 > average {
			continue
		}
		for n1 := 1; n1 < k && len(out) < 8; n1++ {
			left := k - n1
			if left < 2 {
				break
			}
			need := k*average - n1*p1
			if need <= 0 {
				continue
			}

			// 剩下 left 件由两件拍品组成，要求这两件的单价跨越 need/left。
			for j := 0; j < len(items) && len(out) < 8; j++ {
				if j == i {
					continue
				}
				p2 := items[j].Price
				for m := j + 1; m < len(items) && len(out) < 8; m++ {
					if m == i {
						continue
					}
					p3 := items[m].Price
					den := p2 - p3
					if den <= 0 {
						continue
					}
					num := need - left*p3
					if num%den != 0 {
						continue
					}
					n2 := num / den
					n3 := left - n2
					if n2 < 0 || n3 < 0 {
						continue
					}
					out = append(out, buildSlotCombo([]ComboItem{
						{Item: items[i], Count: n1},
						{Item: items[j], Count: n2},
						{Item: items[m], Count: n3},
					}, average, k))
				}
			}
		}
	}

	sortComboCandidates(out)
	return out
}

// buildSlotCombo 组装一个构成：只保留真正出现的拍品，并计算种类数。
func buildSlotCombo(entries []ComboItem, average, k int) SlotCombo {
	combo := SlotCombo{Count: k, Total: k * average}
	for _, entry := range entries {
		if entry.Count <= 0 {
			continue
		}
		combo.Items = append(combo.Items, entry)
		combo.Distinct++
	}
	return combo
}

// sortComboCandidates 候选内部排序：种类少的优先，其次次数分布更均衡的优先。
func sortComboCandidates(combos []SlotCombo) {
	sort.SliceStable(combos, func(i, j int) bool {
		a, b := combos[i], combos[j]
		if a.Distinct != b.Distinct {
			return a.Distinct < b.Distinct
		}
		return imbalance(a) < imbalance(b)
	})
}

// sortSlotCombos 总排序：件数少的优先（同样均价下件数越少越"确定"），
// 其次拍品种类少、次数分布均衡。
func sortSlotCombos(combos []SlotCombo) {
	sort.SliceStable(combos, func(i, j int) bool {
		a, b := combos[i], combos[j]
		if a.Count != b.Count {
			return a.Count < b.Count
		}
		if a.Distinct != b.Distinct {
			return a.Distinct < b.Distinct
		}
		return imbalance(a) < imbalance(b)
	})
}

// imbalance 衡量一个构成里各拍品次数是否悬殊，越大越像"硬凑"。
func imbalance(combo SlotCombo) float64 {
	if len(combo.Items) == 0 || combo.Count <= 0 {
		return 0
	}
	minCount, maxCount := combo.Items[0].Count, combo.Items[0].Count
	for _, entry := range combo.Items {
		if entry.Count < minCount {
			minCount = entry.Count
		}
		if entry.Count > maxCount {
			maxCount = entry.Count
		}
	}
	return float64(maxCount-minCount) / float64(combo.Count)
}

// comboKey 生成用于去重的键（按价格与品质，因为数据里没有名字）。
func comboKey(combo SlotCombo) string {
	parts := make([]string, 0, len(combo.Items))
	for _, entry := range combo.Items {
		parts = append(parts, fmt.Sprintf("%s/%d*%d", entry.Item.Tier, entry.Item.Price, entry.Count))
	}
	sort.Strings(parts)
	return strings.Join(parts, ",")
}
