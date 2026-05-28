package driver

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	storageConfig "tools-thinker/support/storage/config"
)

var (
	minioHelper *StorageHelper
	// {"provider":"minio","accessKeyId":"xxxxxxxxxxxx","accessKeySecret":"xxxxxxxxxxxxxxxxx","username":"xxxx","password":"xxxxxxxx","endpoint":"x.x.x.x:x x x x","endPointInternal":"x.x.x.x:x x x x","stsEndPoint":"","bucket":"test","roleArn":"","region":"","root":"test/","tmpRoot":"tmp/","internal":false,"host":"http://x.x.x.x:x x x x","cdnDomain":"","cdnProtocol":"http","path":"","tmpPath":""}
	configMinioTestPath = "your_path/minio_conf.json"
	// 本地测试文件
	sourceMinioLocalTestFile = "./testdata/tmp.txt"
	configMinio              = &storageConfig.StorageConfig{}
)

// 初始化客户端配置
func initMinioClient() {
	data, err := os.ReadFile(
		configMinioTestPath,
	)
	if err != nil {
		panic(err)
	}
	json.Unmarshal(data, &configMinio)
	fmt.Printf("config: %+v\n", configMinio)
	minioHelper, err = New2(configMinio)
	if err != nil {
		panic(err)
	}
}

func TestGetObject(t *testing.T) {
	initMinioClient()
	err := minioHelper.PutObject("TestGetObject", []byte("这是什么东西"), nil)
	if err != nil {
		t.Fatal(err)
	}
	data, err := minioHelper.GetObject("TestGetObject")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(data))
}

func TestGetFile(t *testing.T) {
	initMinioClient()
	err := minioHelper.GetFile("12344.png", "./test.png")
	if err != nil {
		t.Fatal(err)
	}
}

func TestPutObject(t *testing.T) {
	initMinioClient()
	err := minioHelper.PutObject("TestPutObject", []byte("这是谁的部下"), nil)
	if err != nil {
		t.Fatal(err)
	}
	data, err := minioHelper.GetObject("TestPutObject")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(data))
}

func TestPutObjectWithMeta(t *testing.T) {
	initMinioClient()
	metadata := &Metadata{
		Mime:               "text/plain",
		ContentDisposition: "TestPutObjectWithMeta",
	}
	err := minioHelper.PutObjectWithMeta("PutObjectWithMeta", []byte("吾乃长山赵子龙"), metadata)
	if err != nil {
		t.Fatal(err)
	}
	data, err := minioHelper.GetObject("PutObjectWithMeta")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(data))
}

func TestPutFile(t *testing.T) {
	initMinioClient()
	err := minioHelper.PutFile("TestPutFile", sourceMinioLocalTestFile, nil)
	if err != nil {
		t.Fatal(err)
	}
	data, err := minioHelper.GetObject("TestPutFile")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(data))
}

func TestPutFileWithMeta(t *testing.T) {
	initMinioClient()
	metadata := &Metadata{
		Mime:               "text/plain",
		ContentDisposition: "TestPutFileWithMeta",
	}
	minioHelper.PutFileWithMeta("TestPutFileWithMeta", sourceMinioLocalTestFile, metadata)
	data, err := minioHelper.GetObject("TestPutFileWithMeta")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(data))
}

func TestPutFileWithPart(t *testing.T) {
	initMinioClient()
	metadata := &Metadata{
		ContentDisposition: "iso file",
	}
	// 定义 MiB 的字节数
	const MiB int64 = 1 << 20 // 1,048,576 字节

	// 计算 128 MiB 的字节数
	partSize := int64(128 * MiB)
	// 测试文件大小为 5.3G
	err := minioHelper.PutFileWithPart(
		"test_vedio.iso",
		"/home/fangyuan/下载/bak/Win10_22H2_China_GGK_Chinese_Simplified_x64.iso",
		metadata,
		partSize,
	)
	if err != nil {
		t.Fatal(err)
	}
}

func TestListObjects(t *testing.T) {
	initMinioClient()
	datas, err := minioHelper.ListObjects("zppt10401")
	if err != nil {
		t.Fatal(err)
	}
	for i := range datas {
		t.Logf("%v", datas[i])
	}
}

func TestDeleteObject(t *testing.T) {
	initMinioClient()
	err := minioHelper.PutObject("TestDeleteObject", []byte("这是什么东西"), nil)
	if err != nil {
		t.Fatal(err)
	}
	data, err := minioHelper.GetObject("TestDeleteObject")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(data))
	err = minioHelper.DeleteObject("TestDeleteObject")
	if err != nil {
		t.Fatal(err)
	}
	_, err = minioHelper.GetObject("TestDeleteObject")
	if err != nil {
		fmt.Println(err.Error())
	}
}

func TestBatchDeleteObject(t *testing.T) {
	initMinioClient()
	fileList := []string{
		"PutObjectWithMeta",
		"TestGetObject",
		"TestPutFile",
		"TestPutFileWithMeta",
		"TestPutObject",
	}
	successFiles, err := minioHelper.BatchDeleteObject(fileList)
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range successFiles {
		t.Log("success delete file: ", file)
	}
}

func TestCopyObject(t *testing.T) {
	initMinioClient()
	err := minioHelper.PutObject("TestCopyObject", []byte("这是什么东西"), nil)
	if err != nil {
		t.Fatal(err)
	}
	err = minioHelper.CopyObject("TestCopyObject", "TestCopyObject_COPY")
	if err != nil {
		t.Fatal(err)
	}
	data, err := minioHelper.GetObject("TestCopyObject_COPY")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(data))
}

func TestIsObjectExist(t *testing.T) {
	initMinioClient()
	err := minioHelper.PutObject("TestIsObjectExist", []byte("这是什么东西"), nil)
	if err != nil {
		t.Fatal(err)
	}
	isObjectExist, err := minioHelper.IsObjectExist("TestIsObjectExist")
	if err != nil {
		t.Fatal(err)
	}
	if isObjectExist {
		t.Log("TestIsObjectExist: 存在")
	}
	err = minioHelper.DeleteObject("TestIsObjectExist")
	if err != nil {
		t.Fatal(err)
	}
	isObjectExist, err = minioHelper.IsObjectExist("TestIsObjectExist")
	if err != nil {
		t.Fatal(err)
	}
	if !isObjectExist {
		t.Log("TestIsObjectExist: 不存在")
	}
}

// 过期时间:秒
func TestSignFileForDownload(t *testing.T) {
	initMinioClient()
	downloadUrl := minioHelper.SignFileForDownload(
		"zppt10401/zppt10401.pptx",
		3600,
		"爱多福多寿.pptx",
	)
	t.Logf("downloadUrl: %s", downloadUrl)
}

func TestGetObjectMeta(t *testing.T) {
	initMinioClient()
	err := minioHelper.PutObject("TestGetObjectMeta", []byte("这是什么东西"), nil)
	if err != nil {
		t.Fatal(err)
	}
	content, err := minioHelper.GetObjectMeta("TestGetObjectMeta")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("content: %v", content)
}

func TestSignFile(t *testing.T) {
	initMinioClient()
	url := minioHelper.SignFile("zppt10401/zppt10401_s.jpg", 3600)
	t.Logf("url: %s", url)
}

func TestSignFile2(t *testing.T) {
	initMinioClient()
	url := minioHelper.SignFile2(configMinio.Bucket, "zppt10401/zppt10401_s.jpg", 3600)
	t.Logf("url: %s", url)
}

// 下述方法采用兼容支持策略

// GetDirToken,GetDirTokenWithAction都是依赖使用GetDirToken2实现,就不统一测试了
func TestGetDirToken2(t *testing.T) {
	initMinioClient()
	token := minioHelper.GetDirToken2("tmp-file")

	// 创建 MinIO 客户端
	client, err := minio.New(minioHelper.storageConfig.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(token.AccessKeyID, token.AccessKeySecret, token.StsToken),
		Secure: true,
	})
	if err != nil {
		t.Errorf("Error creating MinIO client: %v", err)
	}
	data := []byte("TestGetDirToken2\nTestGetDirToken2")
	_, err = client.PutObject(
		context.TODO(),
		configMinio.Bucket,
		"tmp-file/TestCopyObject",
		bytes.NewBuffer(data),
		int64(len(data)),
		minio.PutObjectOptions{},
	)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Printf(
		"Calling list objects on bucket named `%s` with temp creds:\n===\n",
		configMinio.Bucket,
	)
	objCh := client.ListObjects(context.TODO(), configMinio.Bucket, minio.ListObjectsOptions{})
	for obj := range objCh {
		if obj.Err != nil {
			t.Fatalf("Listing error: %v", obj.Err)
		}
		fmt.Printf(
			"Key: %s\nSize: %d\nLast Modified: %s\n===\n",
			obj.Key,
			obj.Size,
			obj.LastModified,
		)
	}
}
