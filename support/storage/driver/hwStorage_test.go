package driver

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	storageConfig "tools-thinker/support/storage/config"

	"github.com/go-playground/assert/v2"
	"github.com/north-team/huawei-obs-sdk-go/obs"
)

var hst *hwStorage
var obsConf *storageConfig.StorageConfig

func initHWClient() {

	cMap := make(map[string]string)
	data, err := os.ReadFile(
		"C:/Users/76782/Documents/rongke/code/manage_service/test/obs_conf.json",
	)
	if err != nil {
		panic(err)
	}
	json.Unmarshal(data, &cMap)

	config := &storageConfig.StorageConfig{
		Provider:        cMap["Provider"],
		AccessKeyId:     cMap["AccessKeyId"],
		AccessKeySecret: cMap["AccessKeySecret"],
		Endpoint:        cMap["Endpoint"],
		Bucket:          cMap["Bucket"],
		Region:          cMap["Region"],
	}
	client, err := obs.New(config.AccessKeyId, config.AccessKeySecret, config.Endpoint)
	if err != nil {
		panic(err)
	}
	hst = &hwStorage{config: config, client: client}
	obsConf = config
}

func TestHW_SignFile(t *testing.T) {
	initHWClient()
	_, getSignedUrl := hst.SignFile("TestGetDirTokenWithAction", int64(1200))
	resp := execGetAction(getSignedUrl)
	// 使用assert来验证下载的内容是否与预期匹配
	assert.Equal(t, "TestGetDirTokenWithAction", resp)
}

func TestHW_GetDirToken2(t *testing.T) {
	initHWClient()
	// res := hst.GetDirToken2("TestGetDirToken2")
	// // 上传
	// responsePut := execPutAction(res.PutSignedUrl, "TestGetDirToken2")
	// // 使用assert来验证上传的状态码是否符合预期
	// assert.Equal(t, "200 OK", responsePut.Status)

	// // 下载
	// resp := execGetAction(res.GetSignedUrl)
	// // 使用assert来验证下载的内容是否与预期匹配
	// assert.Equal(t, "TestGetDirToken2", resp)

}

func TestHW_GetDirTokenWithAction(t *testing.T) {
	initHWClient()
	// panic("GetDirTokenWithAction not implemented")
	// ok, res := hst.GetDirTokenWithAction(
	// 	"TestGetDirTokenWithAction",
	// 	storage.PutObjectAction,
	// )
	// assert.Equal(t, ok, true)
	// // 上传
	// responseGet := execPutAction(res.PutSignedUrl, "TestGetDirTokenWithAction")
	// // 使用assert来验证上传的状态码是否符合预期
	// assert.Equal(t, "200 OK", responseGet.Status)
	// ok, res = hst.GetDirTokenWithAction(
	// 	"TestGetDirTokenWithAction",
	// 	storage.GetObjectAction,
	// )
	// assert.Equal(t, ok, true)
	// // 下载
	// resp := execGetAction(res.GetSignedUrl)
	// // 使用assert来验证下载的内容是否与预期匹配
	// assert.Equal(t, "TestGetDirTokenWithAction", resp)

	// // 未知动作
	// okUnKnown, resUnKnown := hst.GetDirTokenWithAction(
	// 	"TestGetDirTokenWithAction",
	// 	storage.PutObjectAclAction,
	// )
	// assert.Equal(t, false, okUnKnown)
	// assert.Equal(t, nil, resUnKnown)
}

func TestHW_GetObjectMeta(t *testing.T) {
	initHWClient()
	content, err := hst.GetObjectMeta(obsConf.Bucket, "TestGetDirTokenWithAction")
	if err != nil {
		panic(err)
	}
	t.Log(content)
}

func TestHW_RestoreArchive(t *testing.T) {
	initHWClient()
	// // 先上传
	// res := hst.GetDirToken2("TestRestoreArchive")
	// // 上传
	// responsePut := execPutAction(res.PutSignedUrl, "TestRestoreArchive")
	// // 使用assert来验证上传的状态码是否符合预期
	// assert.Equal(t, "200 OK", responsePut.Status)
	// // 设置归档
	// storageClassType := setArchive("TestRestoreArchive")
	// assert.Equal(t, obs.StorageClassCold, storageClassType)
	// // 判断是不是归档
	// is, err := hst.IsArchive("TestRestoreArchive")
	// assert.Equal(t, true, is)
	// assert.Equal(t, nil, err)
	// // 解除归档
	// r, err := hst.RestoreArchive("TestRestoreArchive")
	// assert.Equal(t, true, r)
	// assert.Equal(t, nil, err)
	// // 判断是不是归档
	// is, err = hst.IsArchive("dev_test")
	// assert.Equal(t, false, is)
	// assert.Equal(t, nil, err)
}

func TestHW_IsArchive(t *testing.T) {
	initHWClient()
	res, err := hst.IsArchive(obsConf.Bucket, "dev_123_test")
	assert.Equal(t, false, res)
	assert.Equal(t, nil, err)
}

// 执行上传
func execPutAction(putSignedUrl string, data string) *http.Response {
	// 调用授权URl进行上传
	payloadPut := strings.NewReader(data)
	reqPut, err := http.NewRequest("PUT", putSignedUrl, payloadPut)
	if err != nil {
		panic(err)
	}
	// 设置归档存储类别
	resp, err := http.DefaultClient.Do(reqPut)
	if err != nil {
		panic(err)
	}
	return resp
}

// 执行下载
func execGetAction(getSignedUrl string) string {
	// 调用授权URl进行下载
	reqGet, err := http.NewRequest("GET", getSignedUrl, nil)
	if err != nil {
		panic(err)
	}
	responseGet, err := http.DefaultClient.Do(reqGet)
	if err != nil {
		panic(err)
	}
	var downloadedContent strings.Builder
	p := make([]byte, 1024)
	var readErr error
	var readCount int
	// 读取对象内容
	for {
		readCount, readErr = responseGet.Body.Read(p)
		if readCount > 0 {
			downloadedContent.Write(p[:readCount])
		}
		if readErr != nil {
			if readErr == io.EOF {
				break
			}
			panic(readErr)
		}
	}
	return downloadedContent.String()
}

func updateArchive(key string) string {
	ak := obsConf.AccessKeyId
	sk := obsConf.AccessKeySecret
	endPoint := obsConf.Endpoint
	obsClient, err := obs.New(ak, sk, endPoint /*, obs.WithSecurityToken(securityToken)*/)
	if err != nil {
		panic(err)
	}
	putObjectInput := &obs.CreateSignedUrlInput{}
	putObjectInput.Method = obs.HttpMethodPut
	putObjectInput.Bucket = obsConf.Bucket
	putObjectInput.Key = key
	putObjectInput.Expires = 3600
	// 生成上传对象的带授权信息的URL
	putObjectOutput, err := obsClient.CreateSignedUrl(putObjectInput)
	if err != nil {
		panic(err)
	}
	return putObjectOutput.SignedUrl
}

// 设置归档属性
func setArchive(key string) obs.StorageClassType {
	ak := obsConf.AccessKeyId
	sk := obsConf.AccessKeySecret
	endPoint := obsConf.Endpoint
	obsClient, err := obs.New(ak, sk, endPoint /*, obs.WithSecurityToken(securityToken)*/)
	if err != nil {
		panic(err)
	}
	input := &obs.SetObjectMetadataInput{}
	// 指定存储桶名称
	input.Bucket = obsConf.Bucket
	// 指定对象，此处以 example/objectname 为例。
	input.Key = key
	// 指定对象存储类型，这里以obs.StorageClassCold为例
	input.StorageClass = obs.StorageClassCold
	// 设置对象元数据
	output, err := obsClient.SetObjectMetadata(input)
	if err != nil {
		panic(err)
	}
	return output.StorageClass
}