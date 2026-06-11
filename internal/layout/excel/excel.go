package excel

import (
	"fmt"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"tools-thinker/internal/excel/feature"
	"tools-thinker/internal/excel/feature/merge"
	"tools-thinker/internal/layout/right"
)

// @Author dx
// @Date 2025-07-22 22:18:00
// @Desc Excel 功能界面（二级菜单 + 功能注册驱动）

// features 已注册的 Excel 功能列表
var features []feature.Feature

// RegisterFeature 注册一个 Excel 功能
func RegisterFeature(f feature.Feature) {
	features = append(features, f)
}

func init() {
	// 注册合并功能
	RegisterFeature(feature.Feature{
		Name:        "合并 Excel",
		Description: "将多个 Excel 文件合并为一个",
		BuildUI:     buildMergeUI,
	})
}

// Content 显示 Excel 功能列表（二级菜单首页）
func Content(window fyne.Window) {
	var objects []fyne.CanvasObject

	objects = append(objects, widget.NewLabelWithStyle("Excel 工具", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}))
	objects = append(objects, widget.NewSeparator())

	for _, f := range features {
		f := f // 捕获循环变量
		btn := widget.NewButton(f.Name, func() {
			showFeatureUI(f, window)
		})
		objects = append(objects, btn)
	}

	objects = append(objects, layout.NewSpacer())

	back := widget.NewButton("返回", func() {
		if right.RefreshContent != nil {
			right.RefreshContent.Objects = []fyne.CanvasObject{
				widget.NewLabel("你点击了返回，当前是初始界面"),
			}
			right.RefreshContent.Refresh()
		}
	})
	objects = append(objects, back)

	if right.RefreshContent != nil {
		right.RefreshContent.Objects = objects
		right.RefreshContent.Refresh()
	}
}

// showFeatureUI 显示某个功能的具体操作界面
func showFeatureUI(f feature.Feature, window fyne.Window) {
	objects := f.BuildUI(window, right.PrintLog)

	// 附加返回按钮
	backBtn := widget.NewButton("返回 Excel 菜单", func() {
		Content(window)
	})
	objects = append(objects, layout.NewSpacer())
	objects = append(objects, backBtn)

	if right.RefreshContent != nil {
		right.RefreshContent.Objects = objects
		right.RefreshContent.Refresh()
	}
}

// ---- 合并功能 UI ----

func buildMergeUI(window fyne.Window, logFunc func(string)) []fyne.CanvasObject {
	// 输入目录
	var inputDir string
	inputDirLabel := widget.NewLabel("未选择")
	inputDirLabel.Wrapping = fyne.TextWrapWord
	inputDirBtn := widget.NewButton("选择 Excel 文件夹", func() {
		dialog.NewFolderOpen(func(uri fyne.ListableURI, err error) {
			if err != nil || uri == nil {
				return
			}
			inputDir = uri.Path()
			inputDirLabel.SetText(inputDir)
		}, window).Show()
	})

	// 输出文件名
	outputNameEntry := widget.NewEntry()
	outputNameEntry.SetPlaceHolder("merged.xlsx")
	outputNameEntry.SetText("merged.xlsx")

	// 输出目录
	var outputDir string
	outputDirLabel := widget.NewLabel("未选择（默认为输入目录）")
	outputDirLabel.Wrapping = fyne.TextWrapWord
	outputDirBtn := widget.NewButton("选择输出目录", func() {
		dialog.NewFolderOpen(func(uri fyne.ListableURI, err error) {
			if err != nil || uri == nil {
				return
			}
			outputDir = uri.Path()
			outputDirLabel.SetText(outputDir)
		}, window).Show()
	})

	// 合并按钮
	mergeBtn := widget.NewButton("开始合并", func() {
		if inputDir == "" {
			dialog.ShowInformation("提示", "请先选择 Excel 文件夹", window)
			return
		}

		outputName := outputNameEntry.Text
		if outputName == "" {
			outputName = "merged.xlsx"
		}

		opts := merge.MergeOptions{
			InputDir:   inputDir,
			OutputDir:  outputDir,
			OutputName: outputName,
			
			ProgressFunc: func(current, total int, fileName string) {
				logFunc(fmt.Sprintf("[%d/%d] 处理文件：%s", current, total, fileName))
			},
		}

		go func() {
			logFunc("开始合并 Excel 文件...")
			err := merge.HandleWithOptions(opts)
			if err != nil {
				logFunc(fmt.Sprintf("合并失败：%s", err.Error()))
			} else {
				outPath := filepath.Join(opts.OutputDir, opts.OutputName)
				logFunc(fmt.Sprintf("合并完成！输出文件：%s", outPath))
			}
		}()
	})

	// 快速合并
	quickMergeBtn := widget.NewButton("快速合并（选择文件夹后直接合并）", func() {
		dialog.NewFolderOpen(func(uri fyne.ListableURI, err error) {
			if err != nil || uri == nil {
				return
			}
			selectedDir := uri.Path()
			logFunc(fmt.Sprintf("开始快速合并：%s", selectedDir))

			go func() {
				err := merge.Handle(selectedDir)
				if err != nil {
					logFunc(fmt.Sprintf("合并失败：%s", err.Error()))
				} else {
					logFunc(fmt.Sprintf("快速合并完成！输出文件：%s/merged.xlsx", selectedDir))
				}
			}()
		}, window).Show()
	})

	form := widget.NewForm(
		widget.NewFormItem("输入目录", container.NewHBox(inputDirBtn, inputDirLabel)),
		widget.NewFormItem("输出文件名", outputNameEntry),
		widget.NewFormItem("输出目录", container.NewHBox(outputDirBtn, outputDirLabel)),
	)

	return []fyne.CanvasObject{
		widget.NewLabelWithStyle("合并 Excel", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
		widget.NewCard("", "将多个 Excel 文件合并为一个文件", form),
		mergeBtn,
		widget.NewSeparator(),
		quickMergeBtn,
	}
}
