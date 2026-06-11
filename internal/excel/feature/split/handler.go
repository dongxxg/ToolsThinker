package split

// @Author dx
// @Date 2025-07-20 22:26:00
// @Desc Excel 工作簿拆分功能

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/xuri/excelize/v2"
	"tools-thinker/support/excelx"
	"tools-thinker/support/file"
)

// SplitMode 拆分模式
type SplitMode int

const (
	// SplitBySheet 按 Sheet 拆分，每个 Sheet 输出为独立文件
	SplitBySheet SplitMode = iota
	// SplitByRowCount 按行数拆分，每 N 行输出为一个文件
	SplitByRowCount
)

// SplitOptions 拆分功能的配置选项
type SplitOptions struct {
	InputFile    string    // 输入文件路径（必填）
	OutputDir    string    // 输出目录（默认为输入文件所在目录）
	Mode         SplitMode // 拆分模式
	RowsPerFile  int       // 按行数拆分时每文件的行数（不含表头）
	SheetName    string    // 按行数拆分时操作的 Sheet（默认第一个）
	ProgressFunc func(current, total int, detail string)
}

// Handle 执行工作簿拆分
func Handle(opts SplitOptions) error {
	if opts.InputFile == "" {
		return fmt.Errorf("输入文件不能为空")
	}

	h, err := excelx.OpenFile(opts.InputFile)
	if err != nil {
		return err
	}
	defer h.Close()

	outputDir := opts.OutputDir
	if outputDir == "" {
		outputDir = filepath.Dir(opts.InputFile)
	}
	if err := file.PreCheckDir(filepath.Join(outputDir, "dummy")); err != nil {
		return fmt.Errorf("创建输出目录失败: %w", err)
	}

	baseName := excelx.FileNameWithoutExt(opts.InputFile)

	switch opts.Mode {
	case SplitBySheet:
		return splitBySheet(h, outputDir, baseName, opts)
	case SplitByRowCount:
		return splitByRowCount(h, outputDir, baseName, opts)
	default:
		return fmt.Errorf("不支持的拆分模式")
	}
}

// splitBySheet 按 Sheet 拆分
func splitBySheet(h *excelx.Helper, outputDir, baseName string, opts SplitOptions) error {
	sheets := h.SheetNames()
	total := len(sheets)

	for i, sheetName := range sheets {
		if opts.ProgressFunc != nil {
			opts.ProgressFunc(i+1, total, fmt.Sprintf("拆分 Sheet：%s", sheetName))
		}

		rows, err := h.GetRows(sheetName)
		if err != nil {
			return fmt.Errorf("读取 Sheet %s 失败: %w", sheetName, err)
		}
		if len(rows) == 0 {
			continue
		}

		// 创建新文件
		out := excelx.NewHelper()
		defer out.Close()
		outH := out.File

		// 确保有一个工作表
		sheets := outH.GetSheetList()
		targetSheet := "Sheet1"
		if len(sheets) > 0 {
			targetSheet = sheets[0]
		}
		if targetSheet != sheetName {
			outH.SetSheetName(targetSheet, sheetName)
		}

		// 写入数据
		for rowIdx, row := range rows {
			cell, _ := excelCellName(1, rowIdx+1)
			if err := outH.SetSheetRow(sheetName, cell, &row); err != nil {
				return fmt.Errorf("写入行失败: %w", err)
			}
		}

		// 安全的文件名（替换非法字符）
		safeName := sanitizeFileName(sheetName)
		outputPath := filepath.Join(outputDir, fmt.Sprintf("%s_%s.xlsx", baseName, safeName))
		if err := out.SaveAs(outputPath); err != nil {
			return err
		}
		out.Close()
	}

	return nil
}

// splitByRowCount 按行数拆分
func splitByRowCount(h *excelx.Helper, outputDir, baseName string, opts SplitOptions) error {
	if opts.RowsPerFile <= 0 {
		return fmt.Errorf("每文件行数必须大于 0")
	}

	// 确定操作的 Sheet
	sheet := opts.SheetName
	if sheet == "" {
		sheets := h.SheetNames()
		if len(sheets) == 0 {
			return fmt.Errorf("文件中没有工作表")
		}
		sheet = sheets[0]
	}

	rows, err := h.GetRows(sheet)
	if err != nil {
		return fmt.Errorf("读取 Sheet %s 失败: %w", sheet, err)
	}
	if len(rows) == 0 {
		return fmt.Errorf("Sheet %s 没有数据", sheet)
	}

	header := rows[0]
	dataRows := rows[1:]

	// 计算需要拆分的文件数
	totalFiles := (len(dataRows) + opts.RowsPerFile - 1) / opts.RowsPerFile

	for fileIdx := 0; fileIdx < totalFiles; fileIdx++ {
		if opts.ProgressFunc != nil {
			opts.ProgressFunc(fileIdx+1, totalFiles, fmt.Sprintf("生成第 %d 个文件", fileIdx+1))
		}

		start := fileIdx * opts.RowsPerFile
		end := start + opts.RowsPerFile
		if end > len(dataRows) {
			end = len(dataRows)
		}

		out := excelx.NewHelper()
		defer out.Close()
		outH := out.File

		// 确保目标 Sheet 名
		sheets := outH.GetSheetList()
		targetSheet := "Sheet1"
		if len(sheets) > 0 {
			targetSheet = sheets[0]
		}
		if targetSheet != sheet {
			outH.SetSheetName(targetSheet, sheet)
		}

		rowIndex := 1

		// 写入表头
		cell, _ := excelCellName(1, rowIndex)
		outH.SetSheetRow(sheet, cell, &header)
		rowIndex++

		// 写入数据行
		for _, row := range dataRows[start:end] {
			cell, _ := excelCellName(1, rowIndex)
			outH.SetSheetRow(sheet, cell, &row)
			rowIndex++
		}

		outputPath := filepath.Join(outputDir, fmt.Sprintf("%s_%d.xlsx", baseName, fileIdx+1))
		if err := out.SaveAs(outputPath); err != nil {
			return err
		}
		out.Close()
	}

	return nil
}

// excelCellName 生成单元格坐标名
func excelCellName(col, row int) (string, error) {
	return excelize.CoordinatesToCellName(col, row)
}

// sanitizeFileName 替换文件名中的非法字符
func sanitizeFileName(name string) string {
	illegal := []string{"\\", "/", ":", "*", "?", "\"", "<", ">", "|"}
	result := name
	for _, ch := range illegal {
		result = strings.ReplaceAll(result, ch, "_")
	}
	return result
}
