package view

import (
	"blind-tools/model"
	"os"
	"path/filepath"
	"testing"
)

// auctionDir 定位仓库里的 data/auction 目录。
// go test 的工作目录是包目录（view/），所以根目录和上一级都要试。
func auctionDir(t *testing.T) string {
	t.Helper()

	candidates := []string{
		filepath.Join("data", "auction"),
		filepath.Join("..", "data", "auction"),
	}
	for _, dir := range candidates {
		if _, err := os.Stat(filepath.Join(dir, model.AuctionFile)); err == nil {
			return dir
		}
	}

	wd, _ := os.Getwd()
	t.Skipf("找不到竞拍数据（当前工作目录 %s，试过 %v）", wd, candidates)
	return ""
}

// loadTestAuction 读取仓库内的竞拍数据。
func loadTestAuction(t *testing.T) model.AuctionData {
	t.Helper()

	data, err := model.LoadAuction(auctionDir(t))
	if err != nil {
		t.Fatalf("加载竞拍数据失败：%v", err)
	}
	return data
}

// checkCombo 校验一条构成自洽：件数对得上、明细合计 = 件数 × 均价。
func checkCombo(t *testing.T, combo SlotCombo, average int) {
	t.Helper()

	if combo.Count <= 0 {
		t.Fatalf("件数非法：%d", combo.Count)
	}

	count, total := 0, 0
	for _, entry := range combo.Items {
		if entry.Count <= 0 {
			t.Fatalf("拍品 %s 的次数非法：%d", entry.Item.Name, entry.Count)
		}
		count += entry.Count
		total += entry.Item.Price * entry.Count
	}
	if count != combo.Count {
		t.Fatalf("明细件数合计 = %d，件数标注 = %d", count, combo.Count)
	}
	if total != combo.Total || total != combo.Count*average {
		t.Fatalf("合计 %d（明细 %d），期望 %d 件 × 均价 %d = %d",
			combo.Total, total, combo.Count, average, combo.Count*average)
	}
	if combo.Distinct != len(combo.Items) {
		t.Fatalf("Distinct = %d，明细却有 %d 条", combo.Distinct, len(combo.Items))
	}
}

// TestAuctionDataLoads 检查 items.json 能加载并通过校验。
func TestAuctionDataLoads(t *testing.T) {
	data := loadTestAuction(t)

	if len(data.Items) == 0 {
		t.Fatal("拍品列表为空")
	}

	counts := map[string]int{}
	for _, item := range data.Items {
		counts[item.Tier]++
	}
	t.Logf("拍品 %d 种，档位分布 %v", len(data.Items), counts)

	for _, tier := range data.SortedTiers() {
		if counts[tier.Key] == 0 {
			t.Errorf("档位 %s(%s) 没有任何拍品", tier.Key, tier.Name)
		}
	}
}

// TestGuessFindsComposition 用一个真实拍品单价当平均价，必须能推出该单件构成。
func TestGuessFindsComposition(t *testing.T) {
	data := loadTestAuction(t)

	item := data.Items[0]
	guess := searchSlotComposition(data, item.Price, DefaultMaxDepth, 50)

	if len(guess.Depths) == 0 {
		t.Fatalf("平均价 %d 应当至少能推出 1 件的构成", item.Price)
	}
	if guess.Depths[0] != 1 {
		t.Fatalf("单件单价当平均价时，最小件数应为 1，实际 %v", guess.Depths)
	}
	if len(guess.Combos) == 0 {
		t.Fatal("没有返回任何构成")
	}

	found := false
	for _, combo := range guess.Combos {
		checkCombo(t, combo, item.Price)
		if combo.Count == 1 {
			found = true
		}
	}
	if !found {
		t.Fatal("应该给出「1 件」的那条构成")
	}
}

// TestGuessDepthsAreExact 每个被判定可行的件数，都必须真的存在能命中的构成。
func TestGuessDepthsAreExact(t *testing.T) {
	data := loadTestAuction(t)

	averages := []int{500, 1000, 2036, 5000}
	for _, average := range averages {
		guess := searchSlotComposition(data, average, DefaultMaxDepth, 200)

		byCount := map[int][]SlotCombo{}
		for _, combo := range guess.Combos {
			checkCombo(t, combo, average)
			byCount[combo.Count] = append(byCount[combo.Count], combo)
		}

		// 被列出的件数都必须有构成；反之亦然。
		for _, depth := range guess.Depths {
			if len(byCount[depth]) == 0 {
				t.Fatalf("均价 %d：件数 %d 被判定可行，却没有对应构成", average, depth)
			}
		}
		for count := range byCount {
			listed := false
			for _, depth := range guess.Depths {
				if depth == count {
					listed = true
					break
				}
			}
			if !listed {
				t.Fatalf("均价 %d：件数 %d 有构成，却没列进 Depths", average, count)
			}
		}

		t.Logf("均价 %d：件数 %v，共 %d 条构成", average, guess.Depths, len(guess.Combos))
	}
}

// TestGuessImpossibleAverage 奖池凑不出的平均价要返回空结果。
func TestGuessImpossibleAverage(t *testing.T) {
	data := loadTestAuction(t)

	// 低于最低单价、或高于最高单价，都不可能有解。
	guess := searchSlotComposition(data, 1, DefaultMaxDepth, 50)
	if len(guess.Depths) != 0 || len(guess.Combos) != 0 {
		t.Fatalf("平均价 1 不可能有解，实际 %v", guess.Depths)
	}
	if guess.MinPrice <= 0 || guess.MaxPrice <= guess.MinPrice {
		t.Fatalf("单价范围非法：%d ~ %d", guess.MinPrice, guess.MaxPrice)
	}
}

// TestGuessSortsByCount 结果按件数从少到多排列。
func TestGuessSortsByCount(t *testing.T) {
	data := loadTestAuction(t)

	guess := searchSlotComposition(data, 2036, DefaultMaxDepth, 200)
	if len(guess.Combos) < 2 {
		t.Skip("构成太少，无法验证排序")
	}

	for i := 1; i < len(guess.Combos); i++ {
		if guess.Combos[i].Count < guess.Combos[i-1].Count {
			t.Fatalf("结果未按件数升序：%d 在 %d 之后",
				guess.Combos[i].Count, guess.Combos[i-1].Count)
		}
	}
}

// TestGuessRespectsMaxDepth 件数上限会被尊重。
func TestGuessRespectsMaxDepth(t *testing.T) {
	data := loadTestAuction(t)

	guess := searchSlotComposition(data, 2036, 4, 200)
	for _, depth := range guess.Depths {
		if depth > 4 {
			t.Fatalf("件数 %d 超过上限 4", depth)
		}
	}
	for _, combo := range guess.Combos {
		if combo.Count > 4 {
			t.Fatalf("构成件数 %d 超过上限 4", combo.Count)
		}
	}
}

// TestPairSolver 两件拍品的整数判定：n1·p1 + n2·p2 = k·average 必须精确成立。
func TestPairSolver(t *testing.T) {
	data := loadTestAuction(t)

	// 2036 与 2040 的均价 2038：k 件里一半 2036、一半 2040 即可（k 为偶数）。
	guess := searchSlotComposition(data, 2038, 10, 200)
	if len(guess.Depths) == 0 {
		t.Fatal("均价 2038 应当有解")
	}
	for _, combo := range guess.Combos {
		checkCombo(t, combo, 2038)
	}
	t.Logf("均价 2038：件数 %v", guess.Depths)
}

// TestGuessExtremeAverages 边界均价（等于最高/最低单价）也要有解。
func TestGuessExtremeAverages(t *testing.T) {
	data := loadTestAuction(t)

	maxPrice, minPrice := 0, 0
	for _, item := range data.Items {
		if item.Price > maxPrice {
			maxPrice = item.Price
		}
		if minPrice == 0 || item.Price < minPrice {
			minPrice = item.Price
		}
	}

	for _, average := range []int{maxPrice, minPrice} {
		guess := searchSlotComposition(data, average, DefaultMaxDepth, 50)
		if len(guess.Depths) == 0 {
			t.Fatalf("均价 %d（奖池单价边界）应当有解：1 件即可", average)
		}
		if guess.Depths[0] != 1 {
			t.Fatalf("均价 %d 的最小件数应为 1，实际 %v", average, guess.Depths)
		}
		foundSingle := false
		for _, combo := range guess.Combos {
			checkCombo(t, combo, average)
			if combo.Count == 1 {
				foundSingle = true
			}
		}
		if !foundSingle {
			// 边界均价可能因为 maxResults 截断而没带上「1 件」那条，这里明确提醒。
			t.Errorf("均价 %d 应当给出「1 件」的构成", average)
		}
		t.Logf("均价 %d：件数 %v，共 %d 条构成", average, guess.Depths, len(guess.Combos))
	}
}

// TestGuessNoSolutionGivesNearest 完全无解时要给出附近可行均价，而不是空手而归。
func TestGuessNoSolutionGivesNearest(t *testing.T) {
	data := loadTestAuction(t)

	// 均价低于最低单价：任何件数都无解。
	guess := searchSlotComposition(data, 1, DefaultMaxDepth, 50)
	if len(guess.Depths) != 0 || len(guess.Combos) != 0 {
		t.Fatalf("均价 1 不应有解，实际 %v", guess.Depths)
	}
	if len(guess.Nearest) == 0 {
		t.Fatal("无解时应给出附近可行均价")
	}
	for _, near := range guess.Nearest {
		if near <= 0 {
			t.Fatalf("附近均价非法：%d", near)
		}
		// 提示的均价必须真的能凑出来（用与提示相同的判定口径）。
		found := false
		for k := 1; k <= DefaultMaxDepth && !found; k++ {
			if len(solveWithPairs(data, data.Items, near, k)) > 0 {
				found = true
			}
			if len(solveWithTriples(data, data.Items, near, k)) > 0 {
				found = true
			}
		}
		if !found {
			t.Fatalf("提示的附近均价 %d 其实凑不出来", near)
		}
	}
	t.Logf("均价 1 无解，附近可行均价：%v", guess.Nearest)
}

// TestFormatNumber 验证千位分隔。
func TestFormatNumber(t *testing.T) {
	cases := map[int]string{
		0:       "0",
		99:      "99",
		1000:    "1,000",
		24975:   "24,975",
		1314520: "1,314,520",
		-1234:   "-1,234",
	}
	for in, want := range cases {
		if got := formatNumber(in); got != want {
			t.Fatalf("formatNumber(%d) = %q，期望 %q", in, got, want)
		}
	}
}

// TestPriceBitset 验证位集合的基本操作。
func TestPriceBitset(t *testing.T) {
	set := newPriceBitset(200)
	set.set(0)
	set.set(65)
	set.set(199)

	if !set.has(0) || !set.has(65) || !set.has(199) {
		t.Fatalf("置位失败：%v", set)
	}
	if set.has(1) || set.has(64) || set.has(200) {
		t.Fatalf("未置位的价格被判定为可达：%v", set)
	}
	if set.count() != 3 {
		t.Fatalf("count = %d，期望 3", set.count())
	}
	if !set.nonEmpty() {
		t.Fatal("非空集合被判定为空")
	}
	if newPriceBitset(10).nonEmpty() {
		t.Fatal("空集合被判定为非空")
	}
}
