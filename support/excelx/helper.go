package excelx

// @Author dx
// @Date 2025-07-20 22:26:00
// @Desc Excel 操作帮助包，封装 excelize 提供统一的操作接口

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/xuri/excelize/v2"
	"tools-thinker/support/file"
	"tools-thinker/support/logger"
)

// Helper 封装 excelize.File 提供高层 Excel 操作接口
type Helper struct {
	File      *excelize.File
	logPrefix logger.LogPrefix
	filePath  string
}

// NewHelper 创建新的空白 Excel 文件
func NewHelper() *Helper {
	return &Helper{
		File:      excelize.NewFile(),
		logPrefix: logger.LogPrefix("[excelx]"),
	}
}

// OpenFile 打开已有的 Excel 文件
func OpenFile(path string) (*Helper, error) {
	f, err := excelize.OpenFile(path)
	if err != nil {
		return nil, fmt.Errorf("打开文件失败 %s: %w", path, err)
	}
	return &Helper{
		File:      f,
		logPrefix: logger.LogPrefix("[excelx]"),
		filePath:  path,
	}, nil
}

// SaveAs 保存文件到指定路径
func (h *Helper) SaveAs(path string) error {
	if err := file.PreCheckDir(path); err != nil {
		return fmt.Errorf("创建输出目录失败: %w", err)
	}
	if err := h.File.SaveAs(path); err != nil {
		return fmt.Errorf("保存文件失败 %s: %w", path, err)
	}
	h.logPrefix.Info("已保存文件：%s", path)
	return nil
}

// Close 关闭文件释放资源
func (h *Helper) Close() error {
	if h.File != nil {
		return h.File.Close()
	}
	return nil
}

// FilePath 返回当前文件路径
func (h *Helper) FilePath() string {
	return h.filePath
}

// SheetNames 返回所有工作表名称
func (h *Helper) SheetNames() []string {
	return h.File.GetSheetList()
}

// GetRows 读取指定 Sheet 的所有行
func (h *Helper) GetRows(sheet string) ([][]string, error) {
	rows, err := h.File.GetRows(sheet)
	if err != nil {
		return nil, fmt.Errorf("读取 Sheet %s 失败: %w", sheet, err)
	}
	return rows, nil
}

// SetRow 向指定位置写入一行数据
func (h *Helper) SetRow(sheet string, row int, values *[]string) error {
	cell, err := excelize.CoordinatesToCellName(1, row)
	if err != nil {
		return fmt.Errorf("生成单元格坐标失败: %w", err)
	}
	if err := h.File.SetSheetRow(sheet, cell, values); err != nil {
		return fmt.Errorf("写入行失败 sheet=%s row=%d: %w", sheet, row, err)
	}
	return nil
}

// TotalRows 返回指定 Sheet 的总行数
func (h *Helper) TotalRows(sheet string) (int, error) {
	rows, err := h.File.Rows(sheet)
	if err != nil {
		return 0, fmt.Errorf("获取行数失败 sheet=%s: %w", sheet, err)
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		count++
	}
	return count, nil
}

// TotalCols 返回指定 Sheet 的总列数（基于第一行）
func (h *Helper) TotalCols(sheet string) (int, error) {
	rows, err := h.File.GetRows(sheet)
	if err != nil {
		return 0, fmt.Errorf("获取列数失败 sheet=%s: %w", sheet, err)
	}
	if len(rows) == 0 {
		return 0, nil
	}
	return len(rows[0]), nil
}

// GetHeaders 读取指定 Sheet 的表头（第一行）
func (h *Helper) GetHeaders(sheet string) ([]string, error) {
	rows, err := h.File.GetRows(sheet)
	if err != nil {
		return nil, fmt.Errorf("读取表头失败 sheet=%s: %w", sheet, err)
	}
	if len(rows) == 0 {
		return []string{}, nil
	}
	return rows[0], nil
}

// NewSheet 创建新的工作表
func (h *Helper) NewSheet(name string) (int, error) {
	idx, err := h.File.NewSheet(name)
	if err != nil {
		return 0, fmt.Errorf("创建工作表失败 %s: %w", name, err)
	}
	return idx, nil
}

// CopySheet 深拷贝一个工作表
func (h *Helper) CopySheet(srcIdx, dstIdx int) error {
	if err := h.File.CopySheet(srcIdx, dstIdx); err != nil {
		return fmt.Errorf("复制工作表失败 %d->%d: %w", srcIdx, dstIdx, err)
	}
	return nil
}

// DeleteSheet 删除指定工作表
func (h *Helper) DeleteSheet(name string) error {
	if err := h.File.DeleteSheet(name); err != nil {
		return fmt.Errorf("删除工作表失败 %s: %w", name, err)
	}
	return nil
}

// SetSheetName 重命名工作表
func (h *Helper) SetSheetName(oldName, newName string) error {
	if err := h.File.SetSheetName(oldName, newName); err != nil {
		return fmt.Errorf("重命名工作表失败 %s->%s: %w", oldName, newName, err)
	}
	return nil
}

// --- CSV 相关操作 ---

// OpenCSV 从 CSV 文件创建 Helper
func OpenCSV(path string, comma rune) (*Helper, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("打开 CSV 文件失败 %s: %w", path, err)
	}
	defer f.Close()

	reader := csv.NewReader(f)
	reader.Comma = comma
	reader.LazyQuotes = true

	h := NewHelper()
	sheet := "Sheet1"
	rowIndex := 1

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			h.Close()
			return nil, fmt.Errorf("读取 CSV 行失败: %w", err)
		}
		if err := h.SetRow(sheet, rowIndex, &record); err != nil {
			h.Close()
			return nil, err
		}
		rowIndex++
	}

	h.logPrefix.Info("已导入 CSV：%s（%d 行）", path, rowIndex-1)
	return h, nil
}

// SaveAsCSV 将指定 Sheet 导出为 CSV 文件
func (h *Helper) SaveAsCSV(sheet string, path string, comma rune) error {
	if err := file.PreCheckDir(path); err != nil {
		return fmt.Errorf("创建输出目录失败: %w", err)
	}

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("创建 CSV 文件失败 %s: %w", path, err)
	}
	defer f.Close()

	writer := csv.NewWriter(f)
	writer.Comma = comma

	rows, err := h.GetRows(sheet)
	if err != nil {
		return err
	}

	for _, row := range rows {
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("写入 CSV 行失败: %w", err)
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return fmt.Errorf("刷新 CSV 写入失败: %w", err)
	}

	h.logPrefix.Info("已导出 CSV：%s（%d 行）", path, len(rows))
	return nil
}

// --- 批量文件操作 ---

// ScanExcelFiles 扫描目录下所有 .xlsx 文件
func ScanExcelFiles(dir string) ([]string, error) {
	return file.GetFiles(dir, file.XLSX)
}

// ScanCSVFiles 扫描目录下所有 .csv 文件
func ScanCSVFiles(dir string) ([]string, error) {
	return file.GetFiles(dir, ".csv")
}

// FileNameWithoutExt 获取不含扩展名的文件名
func FileNameWithoutExt(path string) string {
	base := filepath.Base(path)
	ext := filepath.Ext(base)
	return base[:len(base)-len(ext)]
}
