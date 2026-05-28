package media

import (
	"strconv"
	"strings"
	"time"
	"tools-thinker/support/logger"
)

type stringIterator = func() (item string, hasNext bool)

type M3u8Parser struct {
	content  []byte
	lines    []string
	tsList   []TsInfo
	duration int64
}

// NewM3u8Parser
//
//	@Description: 生成m3u8的解析器
//	@param content
//	@param getStartTimeFunc 根据ts文件名解析播放开始时间，可以为nil； 为nil时不支持解析开始时间
//	@return *M3u8Parser
func NewM3u8Parser(content []byte, getStartTimeFunc func(tsPathLine string) int64) *M3u8Parser {
	res := &M3u8Parser{
		content: content,
	}
	res.parseLine()
	res.parseTsInfoList(getStartTimeFunc)
	return res
}

func (m *M3u8Parser) parseLine() {
	lines := strings.Split(string(m.content), "\n")
	m.lines = lines
}

func (m *M3u8Parser) GetLineIterator() (it stringIterator, hasNext bool) {
	index := 0
	return func() (item string, hasNext bool) {
		item = m.lines[index]
		index++
		return item, index < len(m.lines)
	}, len(m.lines) > 0
}

type TsInfo struct {
	TsFileName string  // ts文件名称
	Path       string  //ts文件路径
	Duration   float64 // ts播放时长,单位毫秒
	StartTime  int64   // 开始播放时间戳,单位毫秒
	Dir        string  //文件所在目录
}

// #EXTM3U
// #EXT-X-VERSION:3
// #EXT-X-MEDIA-SEQUENCE:0
// #EXT-X-ALLOW-CACHE:YES
// #EXT-X-TARGETDURATION:17
// #EXTINF:16.000000,
// s1/9be3441078466b4be7d339b0b1a94372_R_plaso_04967163-879a-48df-ba95-124539d3a5ca_20230726085204295.ts
// #EXT-X-ENDLIST
func (m *M3u8Parser) GetTsInfoList() []TsInfo {
	return m.tsList
}

func (m *M3u8Parser) parseTsInfoList(getStartTimeFunc func(tsPathLine string) int64) {
	res := make([]TsInfo, 0, 100)
	var extinfLine string
	var tsPathLine string
	for _, line := range m.lines {
		line = strings.Trim(line, " ")
		// 防止CRLF的影响
		line = strings.Trim(line, "\r")
		if strings.HasPrefix(line, "#") {
			if strings.HasPrefix(line, "#EXTINF") {
				extinfLine = line
			}
			continue
		} else if strings.HasSuffix(line, "ts") {
			tsPathLine = line
			ts := newTsInfo(tsPathLine, extinfLine, getStartTimeFunc)
			res = append(res, ts)
			extinfLine = ""
			tsPathLine = ""
		} else {
			// 不处理
		}
	}
	m.tsList = res
}

func (m *M3u8Parser) GetDuration() int64 {
	if m.duration == 0 {
		var res int64 // 单位毫秒
		for _, ts := range m.tsList {
			res += int64(ts.Duration)
		}
		m.duration = res
	}
	return m.duration
}

func newTsInfo(
	tsPathLine string,
	extInfoLine string,
	getStartTimeFunc func(tsPathLine string) int64,
) TsInfo {
	lastIndex := strings.LastIndex(tsPathLine, "/")
	var dir string
	var fileName string
	if lastIndex == -1 {
		fileName = tsPathLine
	} else {
		dir = tsPathLine[0:lastIndex]
		fileName = tsPathLine[lastIndex+1:]
	}
	res := TsInfo{
		TsFileName: fileName,
		Path:       tsPathLine,
		Duration:   parseExtinf(extInfoLine),
		Dir:        dir,
	}
	if getStartTimeFunc != nil {
		res.StartTime = getStartTimeFunc(tsPathLine) //不同m3u8中ts的命名规则不同，所以获取开始时间的支持通过函数传入
	}
	return res
}

// 解析lineExtInfo，获取播放时长
// 格式形如 #EXTINF:16.000000,
func parseExtinf(lineExtInfo string) float64 {
	durationArray := strings.Split(lineExtInfo, ":")
	floatValue, err := strconv.ParseFloat(strings.TrimRight(durationArray[1], ","), 32)
	if err != nil {
		logger.Error(
			"deal ts duration,parsefloat failed,origin lineExtInfo is %s,it should not occur ",
			lineExtInfo,
		)
		return 0
	}
	lastDuration := floatValue * 1000
	return lastDuration
}

// 从ts路径中解析出startTime
// 不同的ts规则不同，这里支持的是声网对课堂录制ts，
// 格式形如 s1/9be3441078466b4be7d339b0b1a94372_R_plaso_04967163-879a-48df-ba95-124539d3a5ca_20230726085236294.ts
func ParseStartTimeFromTsPath4AGORA(tspath string) int64 {
	index := strings.LastIndex(tspath, "_")
	timeStr := tspath[index+1:]
	lineTime := getTime(timeStr)
	return lineTime
}

// getTime
//
//	@Description: 把零时区的时间，转为时间戳,
//	@param timeStr  零时区的yyymmddhhmmssxxx格式 20230726085236294
//	@return int64 时间戳
func getTime(timeStr string) int64 {
	year, _ := strconv.Atoi(timeStr[0:4])
	month, _ := strconv.Atoi(timeStr[4:6])
	day, _ := strconv.Atoi(timeStr[6:8])
	hour, _ := strconv.Atoi(timeStr[8:10])
	min, _ := strconv.Atoi(timeStr[10:12])
	sec, _ := strconv.Atoi(timeStr[12:14])
	ms, _ := strconv.Atoi(timeStr[14:17])
	parseTime := time.Date(year, time.Month(month), day, hour, min, sec, ms*1e6, time.UTC)
	return parseTime.UnixNano() / 1e6
}
