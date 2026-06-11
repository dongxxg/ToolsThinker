package merge

// @Author dx
// @Date 2025-07-20 22:26:00
// @Desc Excel 文件合并功能

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/xuri/excelize/v2"
	"tools-thinker/support/file"
	"tools-thinker/support/logger"
)

// MergeOptions 合并功能的配置选项
type MergeOptions struct {
	InputDir     string // 输入目录（必填）
	OutputDir    string // 输出目录（默认为 InputDir）
	OutputName   string // 输出文件名（默认 merged.xlsx）
	KeepHeader   bool   // 是否在每个文件中保留表头（默认 false，即非首个文件跳过表头）
	ProgressFunc func(current, total int, fileName string)
}

// Handle 使用默认选项合并 Excel 文件（向后兼容）
func Handle(inputDir string) error {
	return HandleWithOptions(MergeOptions{
		InputDir: inputDir,
	})
}

// HandleWithOptions 使用完整选项合并 Excel 文件
func HandleWithOptions(opts MergeOptions) error {
	if opts.InputDir == "" {
		return errors.New("输入目录不能为空")
	}

	sourceDir := opts.InputDir
	outputDir := opts.OutputDir
	if outputDir == "" {
		outputDir = sourceDir
	}
	outputName := opts.OutputName
	if outputName == "" {
		outputName = "merged.xlsx"
	}
	// 默认行为：非首个文件跳过表头；设置 KeepHeader=true 可保留每个文件的表头
	skipHeader := !opts.KeepHeader

	log := logger.LogPrefix("[merge]")

	excelFiles, err := file.GetFiles(sourceDir, file.XLSX)
	if err != nil {
		log.Error("扫描文件失败: %v", err)
		return fmt.Errorf("扫描文件失败: %w", err)
	}

	if len(excelFiles) == 0 {
		log.Warn("没有找到 Excel 文件")
		return errors.New("没有找到 Excel 文件")
	}

	log.Info("找到 %d 个 Excel 文件", len(excelFiles))

	mergedFile := excelize.NewFile()
	sheet := "Sheet1"
	rowIndex := 1

	for i, filePath := range excelFiles {
		if opts.ProgressFunc != nil {
			opts.ProgressFunc(i+1, len(excelFiles), filepath.Base(filePath))
		}

		log.Info("读取文件：%s", filePath)
		f, err := excelize.OpenFile(filePath)
		if err != nil {
			log.Warn("无法打开文件 %s: %v", filePath, err)
			continue
		}

		rows, err := f.GetRows(sheet)
		if err != nil {
			log.Error("读取数据失败 %s: %v", filePath, err)
			f.Close()
			continue
		}

		for j, row := range rows {
			// 除了第一个文件，其他文件跳过标题行
			if i != 0 && j == 0 && skipHeader {
				continue
			}
			cell, _ := excelize.CoordinatesToCellName(1, rowIndex)
			if err := mergedFile.SetSheetRow(sheet, cell, &row); err != nil {
				log.Error("写入失败: %v", err)
			}
			rowIndex++
		}
		f.Close()
	}

	outputPath := filepath.Join(outputDir, outputName)
	err = mergedFile.SaveAs(outputPath)
	if err != nil {
		log.Error("保存合并文件失败：%v", err)
		return fmt.Errorf("保存合并文件失败: %w", err)
	}

	log.Info("成功保存文件：%s（共 %d 行）", outputPath, rowIndex-1)
	return nil
}
