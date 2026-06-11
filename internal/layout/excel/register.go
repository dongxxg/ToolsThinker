package excel

// @Author dx
// @Date 2025-07-20 22:26:00
// @Desc Excel 功能注册入口

import (
	"fmt"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"

	"tools-thinker/internal/excel/feature"
	"tools-thinker/internal/excel/feature/split"
)

func init() {
	// 注册拆分功能
	RegisterFeature(feature.Feature{
		Name:        "拆分工作簿",
		Description: "将一个 Excel 文件拆分为多个文件",
		BuildUI:     buildSplitUI,
	})
}

// ---- 拆分功能 UI ----

func buildSplitUI(window fyne.Window, logFunc func(string)) []fyne.CanvasObject {
	// 输入文件
	var inputPath string
	inputLabel := widget.NewLabel("未选择")
	inputLabel.Wrapping = fyne.TextWrapWord
	inputBtn := widget.NewButton("选择 Excel 文件", func() {
		fd := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err != nil || reader == nil {
				return
			}
			inputPath = reader.URI().Path()
			inputLabel.SetText(inputPath)
			reader.Close()
		}, window)
		fd.SetFilter(filterXLSX())
		fd.Show()
	})

	// 输出目录
	var outputDir string
	outputLabel := widget.NewLabel("未选择（默认为输入文件目录）")
	outputLabel.Wrapping = fyne.TextWrapWord
	outputBtn := widget.NewButton("选择输出目录", func() {
		dialog.NewFolderOpen(func(uri fyne.ListableURI, err error) {
			if err != nil || uri == nil {
				return
			}
			outputDir = uri.Path()
			outputLabel.SetText(outputDir)
		}, window).Show()
	})

	// 拆分模式选择
	modeSelect := widget.NewSelect([]string{"按 Sheet 拆分", "按行数拆分"}, func(value string) {})
	modeSelect.SetSelectedIndex(0)

	// 行数输入（仅按行数拆分时有效）
	rowsEntry := widget.NewEntry()
	rowsEntry.SetPlaceHolder("1000")
	rowsEntry.SetText("1000")
	rowsEntry.Disable()

	// 模式切换时启用/禁用行数输入
	modeSelect.OnChanged = func(value string) {
		if value == "按行数拆分" {
			rowsEntry.Enable()
		} else {
			rowsEntry.Disable()
		}
	}

	// 拆分按钮
	splitBtn := widget.NewButton("开始拆分", func() {
		if inputPath == "" {
			dialog.ShowInformation("提示", "请先选择 Excel 文件", window)
			return
		}

		var mode split.SplitMode
		switch modeSelect.Selected {
		case "按 Sheet 拆分":
			mode = split.SplitBySheet
		case "按行数拆分":
			mode = split.SplitByRowCount
		default:
			mode = split.SplitBySheet
		}

		opts := split.SplitOptions{
			InputFile:   inputPath,
			OutputDir:   outputDir,
			Mode:        mode,
			RowsPerFile: 1000,
			ProgressFunc: func(current, total int, detail string) {
				logFunc(fmt.Sprintf("[拆分] [%d/%d] %s", current, total, detail))
			},
		}

		// 如果是按行数模式，解析行数
		if mode == split.SplitByRowCount {
			rows := 1000
			_, err := fmt.Sscanf(rowsEntry.Text, "%d", &rows)
			if err != nil || rows <= 0 {
				dialog.ShowError(fmt.Errorf("请输入有效的行数"), window)
				return
			}
			opts.RowsPerFile = rows
		}

		go func() {
			logFunc("开始拆分工作簿...")
			err := split.Handle(opts)
			if err != nil {
				logFunc(fmt.Sprintf("[拆分] 失败：%s", err.Error()))
			} else {
				displayDir := opts.OutputDir
				if displayDir == "" {
					displayDir = filepath.Dir(opts.InputFile)
				}
				logFunc(fmt.Sprintf("[拆分] 完成！输出目录：%s", displayDir))
			}
		}()
	})

	form := widget.NewForm(
		widget.NewFormItem("输入文件", container.NewHBox(inputBtn, inputLabel)),
		widget.NewFormItem("输出目录", container.NewHBox(outputBtn, outputLabel)),
		widget.NewFormItem("拆分模式", modeSelect),
		widget.NewFormItem("每文件行数", rowsEntry),
	)

	return []fyne.CanvasObject{
		widget.NewLabelWithStyle("拆分工作簿", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
		widget.NewCard("", "将一个 Excel 文件拆分为多个文件", form),
		splitBtn,
	}
}

// filterXLSX 返回 .xlsx 文件过滤器
func filterXLSX() storage.FileFilter {
	return storage.NewExtensionFileFilter([]string{".xlsx"})
}
