package view

import "math/bits"

// priceBitset 表示一组价格：第 i 位为 1 表示价格 i 在这个集合里。
//
// 反推组合时要反复判断「剩下的钱能不能正好凑出 c 件」，用位集合比
// map[int]bool 快得多，也少得多的分配。
type priceBitset []uint64

// newPriceBitset 创建至少能容纳 maxValue 的位集合。
func newPriceBitset(maxValue int) priceBitset {
	if maxValue < 0 {
		maxValue = 0
	}
	return make(priceBitset, maxValue/64+1)
}

// set 置位 price。
func (b priceBitset) set(price int) {
	if price < 0 || price>>6 >= len(b) {
		return
	}
	b[price>>6] |= 1 << uint(price&63)
}

// has 判断 price 是否在集合里。
func (b priceBitset) has(price int) bool {
	if price < 0 || price>>6 >= len(b) {
		return false
	}
	return b[price>>6]&(1<<uint(price&63)) != 0
}

// nonEmpty 判断是否有任意一位被置位。
func (b priceBitset) nonEmpty() bool {
	for _, word := range b {
		if word != 0 {
			return true
		}
	}
	return false
}

// count 返回集合里价格的数量。
func (b priceBitset) count() int {
	total := 0
	for _, word := range b {
		total += bits.OnesCount64(word)
	}
	return total
}
