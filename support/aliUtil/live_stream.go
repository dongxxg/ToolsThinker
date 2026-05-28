package aliUtil

import (
	"errors"
	"fmt"
	"tools-thinker/support/util"
)

type PullStreamInfo struct {
	Rtmp string `json:"rtmp"`
	Flv  string `json:"flv"`
	Hls  string `json:"hls"`
	Artc string `json:"artc"`
}

type PushPullStreamInfo struct {
	PushUrl string `json:"pushUrl"`
	*PullStreamInfo
}

// 阿里视频直播转码模板
// 转码的说明文档 https://help.aliyun.com/zh/live/user-guide/configure-transcoding-for-general?spm=a2c4g.11186623.0.i4
// 转码流地址的拼接说明文档 https://help.aliyun.com/zh/live/user-guide/ingest-and-streaming-urls?spm=a2c4g.11186623.0.0.6b043eb73eph1g#concept-2010579
type AliLiveTemplateIdType string

const AliLive_TemplateId_RTS AliLiveTemplateIdType = "RONGKE-RTS"

type AliLiveClient struct {
	AppName     string
	PushSignKey string
	PushDomain  string
	PullSignKey string
	PullDomain  string
}

func NewAliLiveClient(
	AppName string,
	PushSignKey string,
	PushDomain string,
	PullSignKey string,
	PullDomain string,
) (*AliLiveClient, error) {
	if len(AppName) == 0 || len(PushSignKey) == 0 || len(PushDomain) == 0 || len(PullDomain) == 0 ||
		len(PullSignKey) == 0 {
		return &AliLiveClient{}, errors.New("NewAliLiveClient failed,as miss param")
	}
	return &AliLiveClient{
		AppName:     AppName,
		PushSignKey: PushSignKey,
		PushDomain:  PushDomain,
		PullSignKey: PullSignKey,
		PullDomain:  PullDomain,
	}, nil
}

// streamName 规则 teachingActivity@head/desk
// endTime 单位秒
func (a *AliLiveClient) genAuthKey(url string, expire int64, signKey string) string {
	if a == nil {
		return ""
	}
	rand := "0"
	uid := "0"
	signString := fmt.Sprintf("%s-%d-%s-%s-%s", url, expire, rand, uid, signKey)
	hashValue := util.Md5String(signString)
	return fmt.Sprintf("%d-%s-%s-%s", expire, rand, uid, hashValue)
}

func (a *AliLiveClient) genAliPushUrl(streamName string, endTime int64) string {
	if a == nil {
		return ""
	}
	// 生成阿里推流地址
	// streamName 规则 teachingActivity@head/desk
	pushUrl := fmt.Sprintf("/%s/%s", a.AppName, streamName)
	authKey := a.genAuthKey(pushUrl, endTime, a.PushSignKey)
	return fmt.Sprintf(
		"rtmp://%s%s?auth_key=%s",
		a.PushDomain,
		pushUrl,
		authKey,
	)
}

// 生成阿里拉流地址
func (a *AliLiveClient) genAliPullUrl4Artc(streamName string, endTime int64) string {
	if a == nil {
		return ""
	}
	pullUrl := fmt.Sprintf("/%s/%s", a.AppName, streamName)
	authKey := a.genAuthKey(pullUrl, endTime, a.PullSignKey)
	return fmt.Sprintf(
		"artc://%s%s?auth_key=%s",
		a.PullDomain,
		pullUrl,
		authKey,
	)
}

// 生成阿里拉流地址 rtmp
func (a *AliLiveClient) genAliPullUrl4Rtmp(streamName string, endTime int64) string {
	if a == nil {
		return ""
	}
	pullUrl := fmt.Sprintf("/%s/%s", a.AppName, streamName)
	authKey := a.genAuthKey(pullUrl, endTime, a.PullSignKey)
	return fmt.Sprintf(
		"rtmp://%s%s?auth_key=%s",
		a.PullDomain,
		pullUrl,
		authKey,
	)
}

// 生成阿里拉流地址flv
func (a *AliLiveClient) genAliPullUrl4Flv(streamName string, endTime int64) string {
	if a == nil {
		return ""
	}
	pullUrl := fmt.Sprintf("/%s/%s.flv", a.AppName, streamName)
	authKey := a.genAuthKey(pullUrl, endTime, a.PullSignKey)
	return fmt.Sprintf(
		"https://%s%s?auth_key=%s",
		a.PullDomain,
		pullUrl,
		authKey,
	)
}

// 生成阿里拉流地址flv
func (a *AliLiveClient) genAliPullUrl4M3u8(streamName string, endTime int64) string {
	if a == nil {
		return ""
	}
	pullUrl := fmt.Sprintf("/%s/%s.m3u8", a.AppName, streamName)
	authKey := a.genAuthKey(pullUrl, endTime, a.PullSignKey)
	return fmt.Sprintf(
		"https://%s%s?auth_key=%s",
		a.PullDomain,
		pullUrl,
		authKey,
	)
}

// GenAliPullUrlInfo 不同协议的区别 https://help.aliyun.com/zh/live/user-guide/rts-overview?spm=a2c4g.11186623.0.i15
// 生成播放地址参考文档 https://help.aliyun.com/zh/live/developer-reference/ingest-and-streaming-urls?spm=5176.13499635.help.dexternal.79352699zsMOHA
// 生成播放地址鉴权串的参考文档 https://help.aliyun.com/zh/live/developer-reference/url-signing?spm=a2c4g.11186623.0.0.7e4227ceIDkDDO#section-ak5-3ig-mv4
// 生成拉流地址集合
// 支持多个协议的拉流地址，前端可以选择性使用，目前qt原生无法播放artc可播放rtmp，web可以播放artc
func (a *AliLiveClient) GenAliPullUrlInfo(
	streamName string,
	endTime int64,
	tId AliLiveTemplateIdType,
) (PullStreamInfo, error) {
	if a == nil {
		return PullStreamInfo{}, errors.New("AliLiveClient is nil")
	}
	// 如果有转码模块就拼接上
	if len(string(tId)) > 0 {
		streamName = streamName + "_" + string(tId)
	}
	resInfo := PullStreamInfo{
		Rtmp: a.genAliPullUrl4Rtmp(streamName, endTime),
		Flv:  a.genAliPullUrl4Flv(streamName, endTime),
		Hls:  a.genAliPullUrl4M3u8(streamName, endTime),
		Artc: a.genAliPullUrl4Artc(streamName, endTime),
	}
	return resInfo, nil
}

// 生成push和pull的地址
func (a *AliLiveClient) GenAliPushPullUrlInfo(
	streamName string,
	endTime int64,
	tId AliLiveTemplateIdType,
) (info *PushPullStreamInfo, err error) {
	if a == nil {
		return nil, errors.New("AliLiveClient is nil")
	}
	pushUrl := a.genAliPushUrl(streamName, endTime)
	// 如果有转码模块就拼接上
	if len(string(tId)) > 0 {
		streamName = streamName + "_" + string(tId)
	}
	resInfo := &PullStreamInfo{
		Rtmp: a.genAliPullUrl4Rtmp(streamName, endTime),
		Flv:  a.genAliPullUrl4Flv(streamName, endTime),
		Hls:  a.genAliPullUrl4M3u8(streamName, endTime),
		Artc: a.genAliPullUrl4Artc(streamName, endTime),
	}
	return &PushPullStreamInfo{
		PushUrl:        pushUrl,
		PullStreamInfo: resInfo,
	}, nil
}
