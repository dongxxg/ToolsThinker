package util

import (
	"image/color"
	"strconv"
	logger "tools-thinker/support/logger"
)

var WHITE = color.RGBA{0xF5, 0xF8, 0xFA, 0xff}

/* ConvertHex2Rgb
* 功能 将hex颜色值转为RGBA类型
* @param value 形如#2e3038
* 如果转换失败，就返回白色背景
 */
func ConvertHex2Rgb(value string) color.RGBA {
	if len(value) != 7 {
		//错误的颜色值
		return WHITE
	}
	rgb, err := strconv.ParseUint(value[1:], 16, 24)
	if err != nil {
		logger.Warn("%s is not valid color HEX;eg #2e3038;error :%v", value, err)
		return WHITE
	}
	res := color.RGBA{uint8((rgb >> 16) & 255), uint8((rgb >> 8) & 255), uint8(rgb & 255), 0xff}
	return res
}
