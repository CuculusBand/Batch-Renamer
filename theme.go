package main

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

type appTheme struct {
	// font        fyne.Resource
	regularFont fyne.Resource
}

// func (m *appTheme) Font(s fyne.TextStyle) fyne.Resource {
// 	return m.font
// }

func (t *appTheme) Font(style fyne.TextStyle) fyne.Resource {
	if style.Monospace {
		return theme.DefaultTheme().Font(style) // 保持默认等宽字体
	}
	return t.regularFont
}

func (m *appTheme) Color(n fyne.ThemeColorName, v fyne.ThemeVariant) color.Color {
	return theme.DefaultTheme().Color(n, v)
}

func (m *appTheme) Icon(n fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(n)
}

func (m *appTheme) Size(n fyne.ThemeSizeName) float32 {
	return theme.DefaultTheme().Size(n)
}
