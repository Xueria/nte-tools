package infer

import (
	"fmt"
	"strings"

	"blind-tools/internal/bid"
	"blind-tools/internal/ui/common"
)

// infer 校验输入，并把推测请求交给装配层。
func (page *Page) infer() {
	query, ok := page.inferQuery()

	if !ok {
		return
	}

	required := page.requiredItems()

	if len(required) > query.MaxCount {
		page.SetStatus(fmt.Sprintf("已确认 %d 件，超过数量上限 %d，请调整数量区间或取消勾选",
			len(required), query.MaxCount))

		return
	}

	page.inferred = true
	page.requiredCount = len(required)
	page.SetStatus("")

	if page.OnInfer != nil {
		page.OnInfer(page.items, required, query)
	}
}

// inferQuery 读取输入并组装推测条件；输入不合法时给出提示并返回 false。
func (page *Page) inferQuery() (bid.InferQuery, bool) {
	minCount, maxCount := page.countRange()

	if page.modeRadio.Selected == priceModeTotal {
		total, ok := common.ParseNonNegativeInt(page.totalEntry.Text)

		if !ok {
			page.SetStatus("请输入组合总价（非负整数）")

			return bid.InferQuery{}, false
		}

		return bid.InferQuery{
			Mode:     bid.TotalMode,
			Price:    total,
			MinCount: minCount,
			MaxCount: maxCount,
		}, true
	}

	avg, ok := common.ParseNonNegativeInt(page.avgEntry.Text)

	if !ok {
		page.SetStatus("请输入均价（非负整数）")

		return bid.InferQuery{}, false
	}

	maxTotal := defaultMaxTotal

	if strings.TrimSpace(page.maxTotalEntry.Text) != "" {
		maxTotal, ok = common.ParseNonNegativeInt(page.maxTotalEntry.Text)

		if !ok {
			page.SetStatus("总价上限要填非负整数")

			return bid.InferQuery{}, false
		}
	}

	return bid.InferQuery{
		Mode:     bid.AverageMode,
		Price:    avg,
		MinCount: minCount,
		MaxCount: maxCount,
		MaxTotal: maxTotal,
	}, true
}

// countRange 返回数量区间滑块给出的件数范围。
func (page *Page) countRange() (minCount, maxCount int) {
	minCount = int(page.countSlider.Lower)
	maxCount = int(page.countSlider.Upper)

	if minCount < minItemCount {
		minCount = minItemCount
	}

	if maxCount < minCount {
		maxCount = minCount
	}

	return minCount, maxCount
}
