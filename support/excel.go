package support

import (
	"github.com/xuri/excelize/v2"
	"strconv"
	"tools-thinker/support/logger"
)

// 初始化创建excel多个sheet
func InitCreateExcel(
	sheetName []string,
) (f *excelize.File, err error, close func(f *excelize.File)) {
	f = excelize.NewFile()
	for _, name := range sheetName {
		// 创建一个工作表
		_, err = f.NewSheet(name)
		if err != nil {
			return nil, err, nil
		}
	}
	// 默认会创建一个名为Sheet1的工作表，需要删除
	err = f.DeleteSheet("Sheet1")
	if err != nil {
		return nil, err, nil
	}
	return f, nil, func(f *excelize.File) {
		if err := f.Close(); err != nil {
			logger.Error("关闭文件失败", err)
		}
	}
}

// 将数字转化成字母
func NumberToLetters(n int) string {
	if n <= 0 {
		return ""
	}

	const base = 26
	mod := (n - 1) % base
	return NumberToLetters((n-1)/base) + string(rune('A'+mod))
}

// 从某个单元格开始按行写入数据
func SetCellValueRow(f *excelize.File, sheet string, row, col int, value interface{}) {
	err := f.SetSheetRow(sheet, NumberToLetters(col)+strconv.Itoa(row), value)
	if err != nil {
		logger.Error("sheet:%s 第%d行 第%d列按行写入失败", sheet, row, col)
		return
	}
}
