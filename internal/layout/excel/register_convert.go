package excel

// @Author dx
// @Date 2025-07-20 22:26:00
// @Desc CSV↔XLSX 格式转换功能注册

import (
	"fmt"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"

	"tools-thinker/internal/excel/feature"
	"tools-thinker/internal/excel/feature/convert"
)

func init() {
	RegisterFeature(feature.Feature{
		Name:        "格式转换",
		Description: "CSV ↔ XLSX 双向格式转换",
		BuildUI:     buildConvertUI,
	})
}

// ---- 格式转换 UI ----

func buildConvertUI(window fyne.Window, logFunc func(string)) []fyne.CanvasObject {
	// 转换方向（自动检测）— 提前声明，闭包中引用
	dirLabel := widget.NewLabel("自动检测")

	// 输入文件
	var inputPath string
	inputLabel := widget.NewLabel("未选择")
	inputLabel.Wrapping = fyne.TextWrapWord
	inputBtn := widget.NewButton("选择文件", func() {
		fd := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err != nil || reader == nil {
				return
			}
			inputPath = reader.URI().Path()
			inputLabel.SetText(inputPath)

			// 自动检测方向并更新显示
			dir := convert.DetectDirection(inputPath)
			if dir == convert.CSVToXLSX {
				dirLabel.SetText("CSV → XLSX")
			} else {
				dirLabel.SetText("XLSX → CSV")
			}
			reader.Close()
		}, window)
		fd.SetFilter(storage.NewExtensionFileFilter([]string{".xlsx", ".csv"}))
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

	// 分隔符选择
	commaSelect := widget.NewSelect([]string{"逗号 (,)", "分号 (;)", "Tab (\\t)"}, func(value string) {})
	commaSelect.SetSelectedIndex(0)

	// Sheet 选择（XLSX→CSV 时有效）
	sheetEntry := widget.NewEntry()
	sheetEntry.SetPlaceHolder("留空使用第一个 Sheet")

	// 单文件转换按钮
	convertBtn := widget.NewButton("开始转换", func() {
		if inputPath == "" {
			dialog.ShowInformation("提示", "请先选择文件", window)
			return
		}

		direction := convert.DetectDirection(inputPath)
		comma := getComma(commaSelect.Selected)

		opts := convert.ConvertOptions{
			InputFile: inputPath,
			OutputDir: outputDir,
			Direction: direction,
			SheetName: sheetEntry.Text,
			Comma:     comma,
			ProgressFunc: func(current, total int, detail string) {
				logFunc(fmt.Sprintf("[转换] [%d/%d] %s", current, total, detail))
			},
		}

		go func() {
			logFunc(fmt.Sprintf("开始转换：%s", filepath.Base(inputPath)))
			err := convert.Handle(opts)
			if err != nil {
				logFunc(fmt.Sprintf("[转换] 失败：%s", err.Error()))
			} else {
				logFunc(fmt.Sprintf("[转换] 完成！"))
			}
		}()
	})

	// ---- 批量转换 ----
	var batchDir string
	batchLabel := widget.NewLabel("未选择")
	batchLabel.Wrapping = fyne.TextWrapWord
	batchBtn := widget.NewButton("选择批量转换目录", func() {
		dialog.NewFolderOpen(func(uri fyne.ListableURI, err error) {
			if err != nil || uri == nil {
				return
			}
			batchDir = uri.Path()
			batchLabel.SetText(batchDir)
		}, window).Show()
	})

	batchDirSelect := widget.NewSelect([]string{"CSV → XLSX", "XLSX → CSV"}, func(value string) {})
	batchDirSelect.SetSelectedIndex(0)

	var batchOutputDir string
	batchOutputLabel := widget.NewLabel("未选择（默认为输入目录）")
	batchOutputLabel.Wrapping = fyne.TextWrapWord
	batchOutputBtn := widget.NewButton("选择输出目录", func() {
		dialog.NewFolderOpen(func(uri fyne.ListableURI, err error) {
			if err != nil || uri == nil {
				return
			}
			batchOutputDir = uri.Path()
			batchOutputLabel.SetText(batchOutputDir)
		}, window).Show()
	})

	batchConvertBtn := widget.NewButton("批量转换", func() {
		if batchDir == "" {
			dialog.ShowInformation("提示", "请先选择批量转换目录", window)
			return
		}

		var direction convert.Direction
		switch batchDirSelect.Selected {
		case "CSV → XLSX":
			direction = convert.CSVToXLSX
		case "XLSX → CSV":
			direction = convert.XLSXToCSV
		}

		comma := getComma(commaSelect.Selected)

		opts := convert.BatchOptions{
			InputDir:  batchDir,
			OutputDir: batchOutputDir,
			Direction: direction,
			Comma:     comma,
			ProgressFunc: func(current, total int, detail string) {
				logFunc(fmt.Sprintf("[批量转换] [%d/%d] %s", current, total, detail))
			},
		}

		go func() {
			logFunc(fmt.Sprintf("开始批量转换：方向=%s, 目录=%s", batchDirSelect.Selected, batchDir))
			err := convert.HandleBatch(opts)
			if err != nil {
				logFunc(fmt.Sprintf("[批量转换] 失败：%s", err.Error()))
			} else {
				logFunc("[批量转换] 完成！")
			}
		}()
	})

	// 单文件转换表单
	singleForm := widget.NewForm(
		widget.NewFormItem("输入文件", container.NewHBox(inputBtn, inputLabel)),
		widget.NewFormItem("转换方向", dirLabel),
		widget.NewFormItem("Sheet 名", sheetEntry),
		widget.NewFormItem("分隔符", commaSelect),
		widget.NewFormItem("输出目录", container.NewHBox(outputBtn, outputLabel)),
	)

	// 批量转换表单
	batchForm := widget.NewForm(
		widget.NewFormItem("输入目录", container.NewHBox(batchBtn, batchLabel)),
		widget.NewFormItem("转换方向", batchDirSelect),
		widget.NewFormItem("输出目录", container.NewHBox(batchOutputBtn, batchOutputLabel)),
	)

	return []fyne.CanvasObject{
		widget.NewLabelWithStyle("格式转换", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
		widget.NewCard("", "CSV ↔ XLSX 双向格式转换", singleForm),
		convertBtn,
		widget.NewSeparator(),
		widget.NewCard("", "批量转换目录下所有文件", batchForm),
		batchConvertBtn,
	}
}

// getComma 从选择器文字解析分隔符
func getComma(selected string) rune {
	switch {
	case strings.Contains(selected, "分号"):
		return ';'
	case strings.Contains(selected, "Tab"):
		return '\t'
	default:
		return ','
	}
}
