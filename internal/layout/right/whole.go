package right

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// @Author dx
// @Date 2025-07-22 22:36:00
// @Desc

var Content *container.Split
var PrintLog func(string)
var RefreshContent *fyne.Container

func Init(window fyne.Window) {
	// 右侧主内容
	//resource, err := fyne.LoadResourceFromPath("default.jpg")
	//if err != nil {
	//	log.Fatal(err)
	//}
	//img := canvas.NewImageFromResource(resource)
	//img.FillMode = canvas.ImageFillContain // 等比例缩放并居中显示，但不会拉伸填满整个区域。
	//img.SetMinSize(fyne.NewSize(400, 300)) // 可选：设置最小尺寸以撑开容器

	RefreshContent = container.NewVBox(
		widget.NewLabel("Please select a menu from the left."),
	)
	// 创建右下的"日志打印区域"
	logOutput := widget.NewMultiLineEntry()
	logOutput.Wrapping = fyne.TextWrapWord
	logOutput.SetMinRowsVisible(5)
	logOutput.Disable() // 禁止用户编辑

	PrintLog = func(message string) {
		fyne.Do(func() {
			logOutput.SetText(
				logOutput.Text + "\n" + message)
		})
	}
	// 右侧上下分区（页面内容 + 日志输出）
	Content = container.NewVSplit(
		RefreshContent,
		logOutput,
	)
	Content.Offset = 0.7 // 设置上下区域比例 70% : 30%
}
