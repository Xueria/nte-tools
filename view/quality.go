package view

import "image/color"

// qualitySwatchSize 品质色块的边长。
const qualitySwatchSize float32 = 8

// qualityLabels 品质标识的展示名，键与数据里的 quality 一致。
var qualityLabels = map[string]string{
	"red":    "红",
	"orange": "橙",
	"purple": "紫",
	"blue":   "蓝",
	"green":  "绿",
	"gray":   "灰",
}

// qualityColors 品质标识的展示色，与游戏内的品质颜色对应。
var qualityColors = map[string]color.Color{
	"red":    nrgba(0xE0, 0x4B, 0x4B, 0xFF),
	"orange": nrgba(0xF2, 0x8C, 0x1E, 0xFF),
	"purple": nrgba(0x9B, 0x5B, 0xE0, 0xFF),
	"blue":   nrgba(0x3D, 0x8B, 0xE0, 0xFF),
	"green":  nrgba(0x3F, 0xA8, 0x5A, 0xFF),
	"gray":   nrgba(0x8A, 0x8A, 0x8A, 0xFF),
}

// qualityLabel 返回品质的展示名，未知品质原样返回标识。
func qualityLabel(quality string) string {
	if label, ok := qualityLabels[quality]; ok {
		return label
	}
	return quality
}

// qualityColor 返回品质的展示色，未知品质用中性灰。
func qualityColor(quality string) color.Color {
	if c, ok := qualityColors[quality]; ok {
		return c
	}
	return nrgba(0x8A, 0x8A, 0x8A, 0xFF)
}
