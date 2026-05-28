package passwd

import (
	"fmt"
	"math/rand"
	"time"
	"tools-thinker/support/logger"
)

const (
	numStr  = "0123456789"
	charStr = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	specStr = "+=-@#~,.[]()!%^*$"
)

const (
	num     = "num"
	char    = "char"
	mix     = "mix"
	advance = "advance"
)

func generatePasswd(length int, charset string) string {
	if length <= 0 {
		logger.Error("generate passwd length set default 8, length: %d", length)
		length = 8
	}
	// 初始化密码切片
	var passwd []byte = make([]byte, length, length)
	// 源字符串
	var sourceStr string
	// 判断字符类型,如果是数字
	switch charset {
	case num:
		sourceStr = numStr
	case char:
		sourceStr = charset
	case mix:
		sourceStr = fmt.Sprintf("%s%s", numStr, charStr)
	case advance:
		sourceStr = fmt.Sprintf("%s%s%s", numStr, charStr, specStr)
	default:
		sourceStr = numStr
	}
	fmt.Println("source str:", sourceStr)

	// 遍历，生成一个随机index索引,
	for i := 0; i < length; i++ {
		index := rand.Intn(len(sourceStr))
		passwd[i] = sourceStr[index]
	}
	return string(passwd)
}

func GenPasswdNum(l int) string {
	rand.Seed(time.Now().UnixNano())
	return generatePasswd(l, num)
}

func GenPasswdChar(l int) string {
	rand.Seed(time.Now().UnixNano())
	return generatePasswd(l, char)
}

func GenPasswdMix(l int) string {
	rand.Seed(time.Now().UnixNano())
	return generatePasswd(l, mix)
}

func GenPasswdAdvance(l int) string {
	rand.Seed(time.Now().UnixNano())
	return generatePasswd(l, advance)
}