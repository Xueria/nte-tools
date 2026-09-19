package common

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// md3Theme 是以 Material Design 3 基准紫为种子色的主题：颜色改用 MD3 的色调
// 角色，圆角也一并放宽，整体是圆润的色调风格。
type md3Theme struct {
	base fyne.Theme
}

// NewMD3Theme 返回 Material Design 3 风格的主题：字体与图标交给 Fyne 默认主题，
// 颜色与尺寸按 MD3 调整。
func NewMD3Theme() fyne.Theme {
	return &md3Theme{base: theme.DefaultTheme()}
}

// Color 返回指定角色在指定明暗变体下的颜色。
func (t *md3Theme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	if variant == theme.VariantLight {
		return lightMD3Color(name)
	}

	return darkMD3Color(name)
}

// Font 把字体解析交给默认主题。
func (t *md3Theme) Font(style fyne.TextStyle) fyne.Resource {
	return t.base.Font(style)
}

// Icon 把图标解析交给默认主题。
func (t *md3Theme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return t.base.Icon(name)
}

// Size 返回 MD3 风格的尺寸，没列出的角色回退到默认主题。
func (t *md3Theme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNamePadding:
		return 4
	case theme.SizeNameInnerPadding:
		return 6
	case theme.SizeNameSplitThickness:
		return 10
	case theme.SizeNameText:
		return 14
	case theme.SizeNameHeadingText:
		return 22
	case theme.SizeNameSubHeadingText:
		return 16
	case theme.SizeNameCaptionText:
		return 12
	case theme.SizeNameButtonRadius:
		return 12
	case theme.SizeNameCardRadius:
		return 14
	case theme.SizeNameInputRadius:
		return 8
	case theme.SizeNameSelectionRadius:
		return 8
	case theme.SizeNameScrollBarRadius:
		return 4
	case theme.SizeNameSeparatorThickness:
		return 1
	default:
		return t.base.Size(name)
	}
}

// nrgba 用十六进制分量拼出一个不透明色。
func nrgba(r, g, b, a uint8) color.NRGBA {
	return color.NRGBA{R: r, G: g, B: b, A: a}
}

// lightMD3Color 把 Fyne 的颜色角色映射到 MD3 浅色基准配色。
func lightMD3Color(name fyne.ThemeColorName) color.Color {
	switch name {
	case theme.ColorNameBackground:
		return nrgba(0xFE, 0xF7, 0xFF, 0xFF)
	case theme.ColorNameForeground:
		return nrgba(0x1D, 0x1B, 0x20, 0xFF)
	case theme.ColorNamePrimary, theme.ColorNameHyperlink:
		return nrgba(0x67, 0x50, 0xA4, 0xFF)
	case theme.ColorNameForegroundOnPrimary:
		return nrgba(0xFF, 0xFF, 0xFF, 0xFF)
	case theme.ColorNameButton:
		return nrgba(0xE8, 0xDE, 0xF8, 0xFF) // secondaryContainer（填充色调按钮）
	case theme.ColorNameDisabledButton:
		return nrgba(0x1D, 0x1B, 0x20, 0x1F)
	case theme.ColorNameDisabled:
		return nrgba(0x1D, 0x1B, 0x20, 0x61)
	case theme.ColorNameInputBackground:
		return nrgba(0xE7, 0xE0, 0xEC, 0xFF) // surfaceVariant
	case theme.ColorNameInputBorder:
		return nrgba(0x79, 0x74, 0x7E, 0xFF) // outline
	case theme.ColorNamePlaceHolder:
		return nrgba(0x49, 0x45, 0x4F, 0xFF) // onSurfaceVariant
	case theme.ColorNameHover:
		return nrgba(0x1D, 0x1B, 0x20, 0x14) // onSurface 8%
	case theme.ColorNamePressed:
		return nrgba(0x1D, 0x1B, 0x20, 0x1F) // onSurface 12%
	case theme.ColorNameFocus:
		return nrgba(0x67, 0x50, 0xA4, 0x40)
	case theme.ColorNameSelection:
		return nrgba(0xEA, 0xDD, 0xFF, 0xFF) // primaryContainer
	case theme.ColorNameSeparator:
		return nrgba(0xCA, 0xC4, 0xD0, 0xFF) // outlineVariant
	case theme.ColorNameOverlayBackground:
		return nrgba(0xEC, 0xE6, 0xF0, 0xFF) // surfaceContainerHigh
	case theme.ColorNameMenuBackground:
		return nrgba(0xF3, 0xED, 0xF7, 0xFF) // surfaceContainer
	case theme.ColorNameHeaderBackground:
		return nrgba(0xF7, 0xF2, 0xFA, 0xFF) // surfaceContainerLow
	case theme.ColorNameScrollBar:
		return nrgba(0x49, 0x45, 0x4F, 0xFF) // onSurfaceVariant
	case theme.ColorNameScrollBarBackground:
		return color.Transparent
	case theme.ColorNameShadow:
		return nrgba(0x00, 0x00, 0x00, 0x28)
	case theme.ColorNameInnerWindowBorder:
		return nrgba(0xCA, 0xC4, 0xD0, 0xFF)
	case theme.ColorNameInnerWindowBorderInactive:
		return nrgba(0xF3, 0xED, 0xF7, 0xFF)
	case theme.ColorNameError:
		return nrgba(0xB3, 0x26, 0x1E, 0xFF)
	case theme.ColorNameForegroundOnError:
		return nrgba(0xFF, 0xFF, 0xFF, 0xFF)
	case theme.ColorNameSuccess:
		return nrgba(0x2E, 0x7D, 0x32, 0xFF)
	case theme.ColorNameForegroundOnSuccess:
		return nrgba(0xFF, 0xFF, 0xFF, 0xFF)
	case theme.ColorNameWarning:
		return nrgba(0xED, 0x6C, 0x02, 0xFF)
	case theme.ColorNameForegroundOnWarning:
		return nrgba(0xFF, 0xFF, 0xFF, 0xFF)
	default:
		return color.Transparent
	}
}

// darkMD3Color 把 Fyne 的颜色角色映射到 MD3 深色基准配色。
func darkMD3Color(name fyne.ThemeColorName) color.Color {
	switch name {
	case theme.ColorNameBackground:
		return nrgba(0x14, 0x12, 0x18, 0xFF)
	case theme.ColorNameForeground:
		return nrgba(0xE6, 0xE0, 0xE9, 0xFF)
	case theme.ColorNamePrimary, theme.ColorNameHyperlink:
		return nrgba(0xD0, 0xBC, 0xFF, 0xFF)
	case theme.ColorNameForegroundOnPrimary:
		return nrgba(0x38, 0x1E, 0x72, 0xFF)
	case theme.ColorNameButton:
		return nrgba(0x4A, 0x44, 0x58, 0xFF) // secondaryContainer
	case theme.ColorNameDisabledButton:
		return nrgba(0xE6, 0xE0, 0xE9, 0x1F)
	case theme.ColorNameDisabled:
		return nrgba(0xE6, 0xE0, 0xE9, 0x61)
	case theme.ColorNameInputBackground:
		return nrgba(0x49, 0x45, 0x4F, 0xFF) // surfaceVariant
	case theme.ColorNameInputBorder:
		return nrgba(0x93, 0x8F, 0x99, 0xFF) // outline
	case theme.ColorNamePlaceHolder:
		return nrgba(0xCA, 0xC4, 0xD0, 0xFF) // onSurfaceVariant
	case theme.ColorNameHover:
		return nrgba(0xE6, 0xE0, 0xE9, 0x14) // onSurface 8%
	case theme.ColorNamePressed:
		return nrgba(0xE6, 0xE0, 0xE9, 0x1F) // onSurface 12%
	case theme.ColorNameFocus:
		return nrgba(0xD0, 0xBC, 0xFF, 0x40)
	case theme.ColorNameSelection:
		return nrgba(0x4F, 0x37, 0x8B, 0xFF) // primaryContainer
	case theme.ColorNameSeparator:
		return nrgba(0x49, 0x45, 0x4F, 0xFF) // outlineVariant
	case theme.ColorNameOverlayBackground:
		return nrgba(0x2B, 0x29, 0x30, 0xFF) // surfaceContainerHigh
	case theme.ColorNameMenuBackground:
		return nrgba(0x21, 0x1F, 0x26, 0xFF) // surfaceContainer
	case theme.ColorNameHeaderBackground:
		return nrgba(0x1D, 0x1B, 0x20, 0xFF) // surfaceContainerLow
	case theme.ColorNameScrollBar:
		return nrgba(0xCA, 0xC4, 0xD0, 0xFF) // onSurfaceVariant
	case theme.ColorNameScrollBarBackground:
		return color.Transparent
	case theme.ColorNameShadow:
		return nrgba(0x00, 0x00, 0x00, 0x50)
	case theme.ColorNameInnerWindowBorder:
		return nrgba(0x49, 0x45, 0x4F, 0xFF)
	case theme.ColorNameInnerWindowBorderInactive:
		return nrgba(0x21, 0x1F, 0x26, 0xFF)
	case theme.ColorNameError:
		return nrgba(0xF2, 0xB8, 0xB5, 0xFF)
	case theme.ColorNameForegroundOnError:
		return nrgba(0x60, 0x14, 0x10, 0xFF)
	case theme.ColorNameSuccess:
		return nrgba(0x81, 0xC7, 0x84, 0xFF)
	case theme.ColorNameForegroundOnSuccess:
		return nrgba(0x0B, 0x1F, 0x0D, 0xFF)
	case theme.ColorNameWarning:
		return nrgba(0xFF, 0xB7, 0x4D, 0xFF)
	case theme.ColorNameForegroundOnWarning:
		return nrgba(0x40, 0x26, 0x00, 0xFF)
	default:
		return color.Transparent
	}
}
