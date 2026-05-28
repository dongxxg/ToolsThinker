package driver

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"tools-thinker/support/file"
	storageConfig "tools-thinker/support/storage/config"

	"github.com/go-playground/assert/v2"
	"github.com/huaweicloud/huaweicloud-sdk-go-obs/obs"
)

var (
	ossHelper      *StorageHelper
	configTestPath = "C:/Users/76782/Documents/rongke/code/school_internal/test/obs_conf.json"
	// 移动云obs测试目录
	unitTestObsPath = "dev/subdev/unit_test"
	// 本地测试文件
	sourceLocalTestFile = "./testdata/tmp.txt"
	config              = &storageConfig.StorageConfig{}
)

// 初始化客户端配置
func initEcloudClient() {
	data, err := os.ReadFile(
		configTestPath,
	)
	if err != nil {
		panic(err)
	}
	json.Unmarshal(data, &config)
	fmt.Printf("config: %+v\n", config)
	ossHelper, err = New2(config)
	if err != nil {
		panic(err)
	}
}

func TestEcloud_PutGetFile(t *testing.T) {
	initEcloudClient()
	key := unitTestObsPath + "/putgetfile.txt"
	err := ossHelper.PutFile(key, sourceLocalTestFile, nil)
	assert.Equal(t, nil, err)
	res, err := ossHelper.GetObject(key)
	assert.Equal(t, nil, err)
	assert.Equal(t, "package testdata", string(res))
}

func TestEcloud_PutFileWithPart(t *testing.T) {
	initEcloudClient()
	key := unitTestObsPath + "/putgetfile.txt"
	m := &Metadata{
		Mime: "video/mp4",
	}
	err := ossHelper.PutFileWithPart(
		key,
		"C:/Users/76782/Documents/rongke/code/school_internal/test/oceans.mp4",
		m,
		10*1024*1024,
	)
	assert.Equal(t, nil, err)
	res, err := ossHelper.GetObject(key)
	assert.Equal(t, nil, err)
	assert.Equal(t, "package testdata", string(res))
}

func TestEcloud_PutGetObject(t *testing.T) {
	initEcloudClient()
	key := unitTestObsPath + "/testecloud_putgetobject.txt"
	data := "TestEcloud_PutGetObject"
	err := ossHelper.PutObject(key, []byte(data), nil)
	assert.Equal(t, nil, err)
	res, err := ossHelper.GetObject(key)
	assert.Equal(t, nil, err)
	assert.Equal(t, "TestEcloud_PutGetObject", string(res))
}

func TestEcloud_SetObjectAcl(t *testing.T) {
	initEcloudClient()
	key := unitTestObsPath + "/setobjectacl.txt"
	err := ossHelper.PutFile(key, sourceLocalTestFile, nil)
	assert.Equal(t, nil, err)
	err = ossHelper.SetObjectAcl(key, AclPrivate)
	assert.Equal(t, nil, err)
	// 在oss浏览器自行验证acl是否设置成功
	// obs://plaso-school/dev/subdev/unit_test/setobjectacl.txt
}

func TestEcloud_SetObjectMetaData_GetObjectMeta(t *testing.T) {
	initEcloudClient()
	key := unitTestObsPath + "/setobjectmetadata_getobjectmeta.txt"
	err := ossHelper.PutFile(key, sourceLocalTestFile, nil)
	assert.Equal(t, nil, err)
	newDownLoadFilename := "TestEcloud_SetObjectMetaData_GetObjectMeta.txt"
	metadata := &Metadata{
		Acl:                string(AclPublicRead),
		ContentDisposition: file.GetContentDisposition(newDownLoadFilename),
	}
	err = ossHelper.SetObjectMetaData(key, metadata)
	assert.Equal(t, nil, err)
	// 获取对象元数据
	content, err := ossHelper.GetObjectMeta(key)
	assert.Equal(t, nil, err)
	assert.NotEqual(t, nil, content)
	// 在oss浏览器自行验证 MetaData 数据
	// obs://plaso-school/dev/subdev/unit_test/setobjectmetadata_getobjectmeta.txt
	t.Log(ossHelper.ViewUrlWithRemoteDir(key))
}

func TestEcloud_IsObjectExist(t *testing.T) {
	initEcloudClient()
	key := unitTestObsPath + "/isobjectexist.txt"
	// 测试不存在的对象
	isExist, err := ossHelper.IsObjectExist(key)
	assert.Equal(t, nil, err)
	assert.Equal(t, false, isExist)
	// 构建上传对象
	err = ossHelper.PutFile(key, sourceLocalTestFile, nil)
	assert.Equal(t, nil, err)
	// 测试存在的对象
	isExist, err = ossHelper.IsObjectExist(key)
	assert.Equal(t, nil, err)
	assert.Equal(t, true, isExist)
}

// GetDirToken 使用的是 GetDirToken2的函数
func TestEcloud_GetDirToken2(t *testing.T) {
	initEcloudClient()
	res := ossHelper.GetDirToken2(unitTestObsPath)
	assert.NotEqual(t, nil, res)
	// 使用生成的 ak sk token 上传文件
	configer := obs.WithSecurityToken(res.StsToken)
	hwObsClient, err := obs.New(
		res.AccessKeyID,
		res.AccessKeySecret,
		res.EndPoint,
		configer,
	)
	assert.Equal(t, nil, err)
	s := ecloudObsStorage{
		client:    hwObsClient,
		config:    nil,
		iamClient: nil,
	}
	// 上传文件
	key := unitTestObsPath + "/getdirtoken2.txt"
	data := "test_GetDirToken2"
	err = s.PutObject(res.Bucket, key, []byte(data), nil)
	assert.Equal(t, nil, err)
	// 获取文件
	getRaw, err := ossHelper.GetObject(key)
	assert.Equal(t, nil, err)
	assert.Equal(t, data, string(getRaw))
	ree := map[string]any{
		"token":     res,
		"urlPathId": -1,
	}
	d, _ := json.Marshal(ree)
	t.Logf("%v", string(d))
}

func TestEcloud_GetDirTokenWithAction(t *testing.T) {
	initEcloudClient()
	ok, res := ossHelper.GetDirTokenWithAction(unitTestObsPath, PutObjectAction)
	assert.NotEqual(t, nil, res)
	assert.Equal(t, true, ok)
	// 使用生成的 ak sk token 上传文件
	configer := obs.WithSecurityToken(res.StsToken)
	hwObsClient, _ := obs.New(
		res.AccessKeyID,
		res.AccessKeySecret,
		config.Endpoint,
		configer,
	)
	s := ecloudObsStorage{
		client:    hwObsClient,
		config:    nil,
		iamClient: nil,
	}
	// 上传文件
	key := unitTestObsPath + "/getdirtokenwithaction.txt"
	data := "test_GetDirTokenWithAction"
	err := s.PutObject(res.Bucket, key, []byte(data), nil)
	assert.Equal(t, nil, err)
	// 获取文件
	getRaw, err := ossHelper.GetObject(key)
	assert.Equal(t, nil, err)
	assert.Equal(t, data, string(getRaw))
}

// SignFile SignFile2 两个函数的功能底层调用一致
func TestEcloud_SignFileForDownload(t *testing.T) {
	initEcloudClient()
	// 构造上传数据
	key := unitTestObsPath + "/signfilefordownload.txt"
	err := ossHelper.PutFile(key, sourceLocalTestFile, nil)
	assert.Equal(t, nil, err)
	// 生成下载链接
	downloadName := "test_signfilefordownload.txt"
	expiredTime := int64(1000)
	url := ossHelper.SignFileForDownload(key, expiredTime, downloadName)
	assert.NotEqual(t, "", url)
	// 验证下载链接
	// 链接类似这种
	// https://plaso-school.obs.joint.cmecloud.cn:443/dev/subdev/unit_test/signfilefordownload.txt?response-content-disposition=attachment%3B+filename%2A%3Dutf-8%27%27test_signfilefordownload.txt&AWSAccessKeyId=AGRP6TGVXJOWWJ5O7FW5&Expires=1719382468&Signature=UTtRpNTSfcwObihaeTSMK5GwETI%3D
	t.Log("DownloadUrl: ", url)
}

func TestEcloud_DeleteObject(t *testing.T) {
	initEcloudClient()
	key := unitTestObsPath + "/testecloud_deleteobject.txt"
	// 构建上传对象
	err := ossHelper.PutFile(key, sourceLocalTestFile, nil)
	assert.Equal(t, nil, err)
	// 测试存在的对象
	isExist, err := ossHelper.IsObjectExist(key)
	assert.Equal(t, nil, err)
	assert.Equal(t, true, isExist)
	// 删除对象(确保ak sk 有删除权限)
	err = ossHelper.DeleteObject(key)
	assert.Equal(t, nil, err)
	// 测试对象删除应该不存在
	isExist, err = ossHelper.IsObjectExist(key)
	assert.Equal(t, nil, err)
	assert.Equal(t, false, isExist)
}
