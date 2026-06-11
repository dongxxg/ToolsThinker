package convert

// @Author dx
// @Date 2025-07-20 22:26:00
// @Desc CSV ↔ XLSX 格式转换功能

import (
	"fmt"
	"path/filepath"
	"strings"

	"tools-thinker/support/excelx"
	"tools-thinker/support/file"
)

// Direction 转换方向
type Direction int

const (
	// CSVToXLSX CSV 转 XLSX
	CSVToXLSX Direction = iota
	// XLSXToCSV XLSX 转 CSV
	XLSXToCSV
)

// ConvertOptions 格式转换的配置选项
type ConvertOptions struct {
	InputFile    string    // 输入文件路径（必填）
	OutputDir    string    // 输出目录（默认为输入文件目录）
	Direction    Direction // 转换方向（自动检测时忽略）
	SheetName    string    // XLSX→CSV 时选择哪个 Sheet（默认第一个）
	Comma        rune      // CSV 分隔符（默认 ','）
	ProgressFunc func(current, total int, detail string)
}

// DetectDirection 根据文件扩展名自动检测转换方向
func DetectDirection(filePath string) Direction {
	ext := strings.ToLower(filepath.Ext(filePath))
	if ext == ".csv" {
		return CSVToXLSX
	}
	return XLSXToCSV
}

// Handle 执行单个文件的格式转换
func Handle(opts ConvertOptions) error {
	if opts.InputFile == "" {
		return fmt.Errorf("输入文件不能为空")
	}

	outputDir := opts.OutputDir
	if outputDir == "" {
		outputDir = filepath.Dir(opts.InputFile)
	}
	if err := file.PreCheckDir(filepath.Join(outputDir, "dummy")); err != nil {
		return fmt.Errorf("创建输出目录失败: %w", err)
	}

	comma := opts.Comma
	if comma == 0 {
		comma = ','
	}

	switch opts.Direction {
	case CSVToXLSX:
		return csvToXLSX(opts, outputDir, comma)
	case XLSXToCSV:
		return xlsxToCSV(opts, outputDir, comma)
	default:
		return fmt.Errorf("不支持的转换方向")
	}
}

// HandleBatch 批量转换目录下的文件
func HandleBatch(opts BatchOptions) error {
	if opts.InputDir == "" {
		return fmt.Errorf("输入目录不能为空")
	}

	outputDir := opts.OutputDir
	if outputDir == "" {
		outputDir = opts.InputDir
	}

	comma := opts.Comma
	if comma == 0 {
		comma = ','
	}

	// 收集需要转换的文件
	var files []string
	direction := opts.Direction

	switch direction {
	case CSVToXLSX:
		csvFiles, err := excelx.ScanCSVFiles(opts.InputDir)
		if err != nil {
			return fmt.Errorf("扫描 CSV 文件失败: %w", err)
		}
		files = csvFiles
	case XLSXToCSV:
		xlsxFiles, err := excelx.ScanExcelFiles(opts.InputDir)
		if err != nil {
			return fmt.Errorf("扫描 XLSX 文件失败: %w", err)
		}
		files = xlsxFiles
	}

	if len(files) == 0 {
		return fmt.Errorf("没有找到可转换的文件")
	}

	total := len(files)
	for i, filePath := range files {
		if opts.ProgressFunc != nil {
			opts.ProgressFunc(i+1, total, filepath.Base(filePath))
		}

		singleOpts := ConvertOptions{
			InputFile:    filePath,
			OutputDir:    outputDir,
			Direction:    direction,
			SheetName:    opts.SheetName,
			Comma:        comma,
			ProgressFunc: nil, // 批量时不传递单个文件进度
		}

		if err := Handle(singleOpts); err != nil {
			return fmt.Errorf("转换 %s 失败: %w", filepath.Base(filePath), err)
		}
	}

	return nil
}

// BatchOptions 批量转换选项
type BatchOptions struct {
	InputDir     string
	OutputDir    string
	Direction    Direction
	SheetName    string
	Comma        rune
	ProgressFunc func(current, total int, detail string)
}

// csvToXLSX 将 CSV 文件转为 XLSX
func csvToXLSX(opts ConvertOptions, outputDir string, comma rune) error {
	h, err := excelx.OpenCSV(opts.InputFile, comma)
	if err != nil {
		return err
	}
	defer h.Close()

	baseName := excelx.FileNameWithoutExt(opts.InputFile)
	outputPath := filepath.Join(outputDir, baseName+".xlsx")

	return h.SaveAs(outputPath)
}

// xlsxToCSV 将 XLSX 文件转为 CSV
func xlsxToCSV(opts ConvertOptions, outputDir string, comma rune) error {
	h, err := excelx.OpenFile(opts.InputFile)
	if err != nil {
		return err
	}
	defer h.Close()

	// 确定要导出的 Sheet
	sheet := opts.SheetName
	if sheet == "" {
		sheets := h.SheetNames()
		if len(sheets) == 0 {
			return fmt.Errorf("文件中没有工作表")
		}
		sheet = sheets[0]
	}

	baseName := excelx.FileNameWithoutExt(opts.InputFile)
	outputPath := filepath.Join(outputDir, baseName+".csv")

	return h.SaveAsCSV(sheet, outputPath, comma)
}
