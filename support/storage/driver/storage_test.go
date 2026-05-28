package driver

import (
	"fmt"
	"os"
	"testing"
	"tools-thinker/support/logger"
	storageConfig "tools-thinker/support/storage/config"
)

func TestAliGetObject(t *testing.T) {

	// Init()
	// config.Init(configName string, serverConfig map[string]string)
	config := &storageConfig.StorageConfig{
		Provider:        "ali",
		AccessKeyId:     "",
		AccessKeySecret: "",
		Endpoint:        "oss-cn-hangzhou.aliyuncs.com",
		Bucket:          "file-plaso",
	}
	ossHelper, err := New2(config)
	fmt.Println("err", err)

	osspath := "dev/exttmp/7/agora/Rplaso07ae6dd7a9de41b7a22ae21d85a04e61/"
	res := ossHelper.FindM3u8(osspath)
	fmt.Println(res)

	ossHelper.SetFolderAcl(
		"dev/exttmp/7/agora/Rplasob61f324812fe422db5e79e5051d6b2f9/s1/",
		"public-read",
	)
	//ossHelper.GetFile("dev-plaso/temp/ali.txt", "/temp2/ali.txt")
	//
	//data, _ := ossHelper.GetObject("dev-plaso/temp/ali.txt")
	//fmt.Println("get file from oss,content:", string(data))
	//
	//_, err3 := ossHelper.IsObjectExist("dev-plaso/temp/ali.txt")
	//
	//if err3 != nil {
	//	logger.Error("dev-plaso/temp/ali.txt not exists ")
	//} else {
	//	logger.Error("dev-plaso/temp/ali.txt exists ")
	//}
	//
	//ossHelper.PutFile("dev-plaso/temp/ali2.txt", "/temp2/ali.txt", nil)
	//ossHelper.PutObject("dev-plaso/temp/ali3.txt", []byte("ali3.txt"), nil)
	//
	//data, _ = ossHelper.GetObject("dev-plaso/temp/ali3.txt")
	//fmt.Println("get file from oss ali3.txt ,content:", string(data))
	//ossHelper.DeleteObject("dev-plaso/temp/ali3.txt")
	//result, _ := ossHelper.ListObjects("dev-plaso/temp/1")
	//fmt.Println("alilist:", result)
	//ossHelper.SetObjectAcl("dev-plaso/temp/ali2.txt", AclPublicRead)
	//ossHelper.SetObjectAcl("dev-plaso/temp/ali.txt", AclPublicReadWrite)
	//_, err := ossHelper.GetFolder("dev-plaso/temp/1", "/temp2/a/")
	//logger.Error("get file error,%s", err)

}

func TestHwGetObject(t *testing.T) {
	// Init()
	config := &storageConfig.StorageConfig{
		Provider:        "huawei",
		AccessKeyId:     "EHQFDNVDANXM3JEDCO",
		AccessKeySecret: "J3jPv9YCcp2ONbklD9qy9wGMOhFpqqrZW0",
		Endpoint:        "obs.cn-east-3.myhuaweicloud.com",
		Bucket:          "file-plaso",
	}

	ossHelper, _ := New2(config)

	ossHelper.GetFile("dev-plaso/temp/hw.txt", "/temp2/hw.txt")

	data, _ := ossHelper.GetObject("dev-plaso/temp/hw.txt")
	fmt.Println("get file from oss,content:", string(data))

	_, err3 := ossHelper.IsObjectExist("dev-plaso/temp/hw.txt")

	if err3 != nil {
		logger.Error("dev-plaso/temp/hw.txt not exists ")
	} else {

		logger.Error("dev-plaso/temp/hw.txt exists ")
	}
	ossHelper.PutFile("dev-plaso/temp/hw2.txt", "/temp2/hw.txt", nil)
	ossHelper.PutObject("dev-plaso/temp/hw3.txt", []byte("hw3.txt"), nil)

	data, _ = ossHelper.GetObject("dev-plaso/temp/hw3.txt")
	fmt.Println("get file from oss hw3.txt ,content:", string(data))
	ossHelper.DeleteObject("dev-plaso/temp/hw3.txt")
	result, _ := ossHelper.ListObjects("dev-plaso/temp/")
	fmt.Println("hwlist:", result)
	ossHelper.SetObjectAcl("dev-plaso/temp/hw2.txt", AclPublicRead)
	// ossHelper.SetObjectAcl("dev-plaso/temp/hw.txt", AclPublicReadWrite)
	_, err := ossHelper.GetFolder("dev-plaso/temp/", "/temp2/b/")
	logger.Error("get file error,%s", err)
}

func TestAliStorage_GetObjectMeta(t *testing.T) {
	ossHelper := initAliOssHelper()
	res, err := ossHelper.GetObjectMeta("dev-plaso/infinite_wb/pdf/diff_wh.pdf")
	if err != nil {
		fmt.Println("error:", err)
	}
	fmt.Println(res)
}

func initAliOssHelper() *StorageHelper {
	var config = &storageConfig.StorageConfig{
		Provider:        "ali",
		AccessKeyId:     "",
		AccessKeySecret: "",
		Endpoint:        "oss-cn-hangzhou.aliyuncs.com",
		Bucket:          "file-plaso",
		Region:          "oss-cn-hangzhou",
		Root:            "dev-plaso/",
		Host:            "file.plaso.cn",
	}
	ossHelper, _ := New2(config)
	return ossHelper
}

func TestEcloudHelper(t *testing.T) {
	var config = &storageConfig.StorageConfig{
		Provider:        "ecloud",
		AccessKeyId:     "HD04UGTM7CKXHTPZAWG3",
		AccessKeySecret: "zDPDtOhr4Rf7JIEDJ14oQspkIYsBvcQCGfG7hQcY",
		Endpoint:        "https://obs.joint.cmecloud.cn",
		Bucket:          "plaso-school",
		Region:          "",
		//Root:            "dev-plaso/",
		//Domain:          "file.plaso.cn",
	}
	ossHelper, e := New2(config)
	if e != nil {
		fmt.Println("failed ", ossHelper, e)
		return
	}
	//list, e := ossHelper.ListObjects("dev")
	ExecTestFunc(ossHelper)
	//fmt.Println(list, e)
}

func Errorf(funcname string, err error, info ...interface{}) {
	if err != nil {
		fmt.Println("ERROR", funcname, err)
		os.Exit(1)
	} else {
		fmt.Println(funcname, info)
	}
}

func ExecTestFunc(helper *StorageHelper) {
	res, err := helper.ListObjects("dev")
	Errorf("ListObjects", err, res)
	res2 := helper.GetRootDir()
	Errorf("GetRootDir", nil, res2)
	res3 := helper.GetBucketName()
	Errorf("GetBucketName", nil, res3)
	res4 := helper.GetEndPoint()
	Errorf("GetEndPoint", nil, res4)
	// helper.GetObject("") ([]byte, error)
	//helper.GetFile(key string, localFile string) error {
	//helper.PutObject(key string, data []byte, metadata map[string]string) error {
	//helper.PutObjectWithMeta(key string, data []byte, metadata *Metadata) error {
	//helper.PutFile(key string, srcFile string, metadata map[string]string) error {
	//helper.PutFileFromFile(
	//helper.PutFileWithMeta(key string, srcFile string, metadata *Metadata) error {
	//helper.ListObjects(prefix string) ([]Content, error) {
	//helper.DeleteObject(key string) error {
	//helper.BatchDeleteObject(fileList []string) (successList []string, err error) {
	//helper.CopyObject(srcKey string, destKey string) error {
	//helper.SetObjectAcl(key string, acl StorageAcl) error {
	//helper.GetFolder(remoteDir string, localFolder string) ([]Content, error) {
	//helper.GetFolderRegex(
	//helper.GetFolderFilter(
	//helper.PutFolder(
	//helper.PutFolderToOss(
	//helper.PutFolder2(
	//helper.GetFolderSize(remoteDir string) (int64, error) {
	//helper.DeleteFolder(remoteDir string) error {
	//helper.CopyFolder(remoteDir string, remoteDistDir string) error {
	//helper.CopyFolderWithAcl(remoteDir string, remoteDistDir string, acl StorageAcl) error {
	//helper.SetFolderAcl(remoteDir string, acl StorageAcl) error {
	//helper.Move(srcKey string, destKey string) error {
	//helper.MoveWithAcl(srcKey string, destKey string, acl StorageAcl) error {
	//helper.MoveFolder(remoteDir string, remoteDistDir string) error {
	//helper.MoveFolderWithAcl(
	//helper.IsObjectExist(key string) (bool, error) {
	//helper.GetDirToken(remoteDir string) map[string]interface{} {
	//helper.GetDirToken2(remoteDir string) *StorageToken {
	//helper.GetDirTokenWithAction(
	//helper.SignFile(remoteDir string, expiredTime int64) string {
	//helper.RemoteDirPathWithRoot(remoteDir string) string {
	//helper.ViewUrlWithRemoteDir(remoteDir string) string {
	//helper.ViewCdnUrlWithRemoteDir(remoteDir string) string {
	//helper.SignFile2(bucket string, remoteDir string, expiredTime int64) string {
	//helper.SignFileWithCdn(
	//helper.GetObjectMeta(key string) (*Content, error) {
	//helper.RestoreArchive(key string) (bool, error) {
	//helper.IsArchive(key string) (bool, error) {
	//helper.PutFileWithPart(
	//helper.SetObjectMetaData(key string, metadata *Metadata) error {
	//helper.FindM3u8(ossPath string) []string {
	//
}

func TestCreateStorage(t *testing.T) {
	sconfig := &storageConfig.StorageConfig{
		Provider:        "tecent",
		AccessKeyId:     "",
		AccessKeySecret: "",
		Endpoint:        "",
		StsEndPoint:     "",
		Bucket:          "trtc-records-1300172876",
		RoleArn:         "",
		Host:            "",
		Region:          "ap-shanghai",
		Root:            "",
	}
	v, e := New2(sconfig)
	if e != nil {
		fmt.Println(v)
		fmt.Println("fail")
		return
	}
	var todoDelFile []string = make([]string, 0, 0)
	for i := 0; i < 2501; i++ {
		todoDelFile = append(todoDelFile, fmt.Sprintf("dev/dev/wyytest/testgenfile/test_%v.txt", i))
	}
	fileList, e := v.BatchDeleteObject(todoDelFile)
	fmt.Println("batch err", e)
	fmt.Println(len(fileList))
}
