package feature

// @Author dx
// @Date 2025-07-20 22:26:00
// @Desc Excel 功能注册机制

import "fyne.io/fyne/v2"

// Feature 定义一个 Excel 操作功能的 UI 和元信息
type Feature struct {
	Name        string // 按钮显示文字
	Description string // 功能描述
	BuildUI     func(window fyne.Window, logFunc func(string)) []fyne.CanvasObject
}
