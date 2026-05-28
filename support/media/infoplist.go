package media

import "tools-thinker/support/logger"

const LastestInfoplistVersion = "2.12"

type InfoPlist struct {
	Media            interface{} `json:"media"`  // 头像,用于flutter播放器
	Layers           interface{} `json:"layers"` // 层级关系,用于flutter播放器
	Duration         int64       `json:"duration"`
	HasLive          bool        `json:"hasLive"`
	Version          string      `json:"version"`          // upime协议版本
	HistoryMedia     string      `json:"historyMedia"`     // 回看历史文件名,用于web播放器
	HistoryMediaList []string    `json:"historyMediaList"` // 存储多路回放, 用于web播放器
}

const HistoryFileName = "history.mp4"
const HistoryDeskFileName = "historyDesk.mp4"
const HistoryHeadFileName = "historyHead.mp4"

// NewInfoPlist 2.3版本的info.plist信息, 用于支持旧的flutter播放器
func NewInfoPlist(duration int64) *InfoPlist {
	// 初始化layer0的空信息
	return &InfoPlist{
		HasLive:  true,
		Version:  "2.3",
		Duration: duration,
		Media:    make([]interface{}, 0),
		Layers: [][][][]interface{}{
			{{{0, 103, 1, 1, 1, "deskShare"}, {duration + 100, 1, 1, [][]int{{7, 0}}}}},
			{{{0, 98, 1, 1, HistoryFileName, duration}, {duration, 98, 1, 2, "", 0}}},
		},
		HistoryMedia: HistoryFileName,
	}
}

// NewInfoPlistV25 2.5版本的info.plist信息, 2.5版本后授课回放变成了双路, 所以需要和旧的一路回放做区分
func NewInfoPlistV25(deskDuration, headDuration int64) *InfoPlist {
	maxDuration := int64(0)
	if deskDuration > headDuration {
		maxDuration = deskDuration
	} else {
		maxDuration = headDuration
	}

	// 初始化layer0的空信息
	return &InfoPlist{
		HasLive:  true,
		Version:  "2.5",
		Duration: maxDuration,
		Media:    [][]any{{0, headDuration, HistoryHeadFileName}}, // 这里是为了兼容老的flutter播放器
		Layers: [][][][]interface{}{
			{{{0, 103, 1, 1, 1, "deskShare"}, {deskDuration + 100, 1, 1, [][]int{{7, 0}}}}},
			{{{0, 98, 1, 1, HistoryDeskFileName, deskDuration}, {deskDuration, 98, 1, 2, "", 0}}},
		},
		HistoryMediaList: []string{HistoryDeskFileName, HistoryHeadFileName},
	}
}

// NewInfoPlistBeyondVersion2V12 2.12之后的版本 初始化方法
func NewInfoPlistBeyondVersion2V12(deskDuration, headDuration int64, version string) *InfoPlist {
	if version != "2.12" {
		// 后续看具体情况再考虑支持
		logger.Warn("需要实现该版本的infolist, 版本号 %s", version)
	}
	// 通用逻辑
	res := NewInfoPlistV25(deskDuration, headDuration)
	res.Version = version
	return res
}
