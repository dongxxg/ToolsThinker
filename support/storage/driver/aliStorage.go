package driver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"tools-thinker/support/file"
	"tools-thinker/support/logger"
	"tools-thinker/support/storage/config"

	sts20150401 "github.com/alibabacloud-go/sts-20150401/client"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

const MAX_BATCH_DELETE_COUNT = 1000

type AliStorage struct {
	config    *config.StorageConfig
	client    *oss.Client
	stsClient *sts20150401.Client //安全授权服务客户端
}

func (_this *AliStorage) IsPublicAcl(bucket string, key string) (bool, error) {
	panic("implement me")
}
func (_this *AliStorage) GetObject(bucket string, key string) ([]byte, error) {
	bucketObj, _ := _this.client.Bucket(bucket)

	body, err := bucketObj.GetObject(key)
	if err != nil {
		logger.Warn("oss GetObject remotefile %s fail %s", key, err)
		return nil, err
	}
	defer body.Close()
	buf := new(bytes.Buffer)
	io.Copy(buf, body)

	return buf.Bytes(), nil
}

func (_this *AliStorage) GetFile(bucket string, key string, localFile string) error {
	bucketObj, _ := _this.client.Bucket(bucket)
	// GetObjectToFile下载gzip的会报错
	// err := bucketObj.GetObjectToFile(key, localFile)

	body, err := bucketObj.GetObject(key)
	if err != nil {
		logger.Warn("oss GetObject remotefile %s fail %s", key, err)
		return err
	}
	defer body.Close()
	os.MkdirAll(path.Dir(localFile), 0755)
	fd, err := os.OpenFile(localFile, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0755)
	if err != nil {
		logger.Warn("open file %v fail,%s", localFile, err)
		return err
	}
	defer fd.Close()
	io.Copy(fd, body)
	return nil
}

func (_this *AliStorage) PutObject(
	bucket string,
	key string,
	data []byte,
	metadata map[string]string,
) error {
	bucketObj, err := _this.client.Bucket(bucket)
	if err != nil {
		logger.Warn("oss PutObject this.client.Bucket %s fail %s", key, err)
		return err
	}
	options := []oss.Option{}
	if metadata != nil {
		// options := []oss.Option{}
		if mime, ok := metadata["mime"]; ok {
			options = append(options, oss.ContentType(mime))
			delete(metadata, "mime")
		}
		if encoding, ok := metadata["Content-Encoding"]; ok {
			options = append(options, oss.ContentEncoding(encoding))
			delete(metadata, "Content-Encoding")
		}
		if acl, ok := metadata["acl"]; ok {
			options = append(options, oss.ObjectACL(oss.ACLType(acl)))
			delete(metadata, "acl")
		}
		for key, value := range metadata {
			options = append(options, oss.Meta(key, value))
		}
	}

	err = bucketObj.PutObject(key, bytes.NewReader(data), options...)
	return err
}

func (_this *AliStorage) PutObjectFromFile(
	bucket string,
	key string,
	filePath string,
	metadata map[string]string,
) error {
	bucketObj, err := _this.client.Bucket(bucket)
	if err != nil {
		logger.Warn("oss PutObjectFromFile this.client.Bucket %s fail %s", key, err)
		return err
	}
	options := []oss.Option{}
	if metadata != nil {
		// options := []oss.Option{}
		if mime, ok := metadata["mime"]; ok {
			options = append(options, oss.ContentType(mime))
			delete(metadata, "mime")
		}
		if encoding, ok := metadata["Content-Encoding"]; ok {
			options = append(options, oss.ContentEncoding(encoding))
			delete(metadata, "Content-Encoding")
		}
		if acl, ok := metadata["acl"]; ok {
			options = append(options, oss.ObjectACL(oss.ACLType(acl)))
			delete(metadata, "acl")
		}
		for key, value := range metadata {
			options = append(options, oss.Meta(key, value))
		}
	}

	err = bucketObj.PutObjectFromFile(key, filePath, options...)
	return err
}

func (_this *AliStorage) PutObjectWithMeta(
	bucket string,
	key string,
	data []byte,
	metadata *Metadata,
) error {
	bucketObj, err := _this.client.Bucket(bucket)
	if err != nil {
		logger.Warn("oss PutObjectWithMeta this.client.Bucket %s fail %s", key, err)
		return err
	}
	options := []oss.Option{}
	if metadata.Mime != "" {
		options = append(options, oss.ContentType(metadata.Mime))
	}
	if metadata.ContentEncoding != "" {
		options = append(options, oss.ContentEncoding(metadata.ContentEncoding))
	}
	if metadata.Acl != "" {
		options = append(options, oss.ObjectACL(oss.ACLType(metadata.Acl)))
	}
	if metadata.ContentDisposition != "" {
		options = append(options, oss.ContentDisposition(metadata.ContentDisposition))
	}
	err = bucketObj.PutObject(key, bytes.NewReader(data), options...)
	return err
}

func (_this *AliStorage) PutFileWithPart(
	bucketStr string,
	key string,
	localFile string,
	metadata *Metadata,
	partSize int64,
) error {
	if partSize == 0 {
		partSize = 10 * 1024 * 1024
	}
	bucket, err := _this.client.Bucket(bucketStr)
	// 获取存储空间。
	if err != nil {
		logger.Warn("%s Error: %v", key, err)
		return err
	}

	// 指定Object的读写权限，默认为继承Bucket的读写权限。
	var metaOptions []oss.Option

	if metadata != nil {
		if metadata.Mime != "" {
			metaOptions = append(metaOptions, oss.ContentType(metadata.Mime))
		}
		if metadata.ContentEncoding != "" {
			metaOptions = append(metaOptions, oss.ContentEncoding(metadata.ContentEncoding))
		}
		if metadata.Acl != "" {
			metaOptions = append(metaOptions, oss.ObjectACL(oss.ACLType(metadata.Acl)))
		}
		if metadata.ContentDisposition != "" {
			metaOptions = append(metaOptions, oss.ContentDisposition(metadata.ContentDisposition))
		}
	}
	metaOptions = append(metaOptions, oss.Routines(5))
	metaOptions = append(metaOptions, oss.CheckpointDir(true, filepath.Dir(localFile)))

	// 重试配置
	maxRetries := 3
	retryInterval := time.Second * 2
	attempt := 0

	for {
		attempt++
		// 通过UploadFile实现断点续传上传时，限制分片数量不能超过10000。
		// 您需要结合上传文件的大小，合理设置每个分片的大小。每个分片大小的取值范围为100 KB~5 GB。默认值为100 KB（即100*1024）。
		// 通过oss.Routines指定分片上传并发数为3。
		// yourObjectName填写Object完整路径，完整路径中不能包含Bucket名称，例如exampledir/exampleobject.txt。
		// yourLocalFile填写本地文件的完整路径，例如D:\\localpath\\examplefile.txt。如果未指定本地路径，则默认从示例程序所属项目对应本地路径中上传文件。
		err = bucket.UploadFile(
			key,
			localFile,
			partSize,
			metaOptions...,
		)

		if err != nil {
			logger.Info(
				"AliStorage PutFileWithPart failed, will retry %d/%d in %s",
				attempt,
				maxRetries,
				retryInterval,
			)
			if attempt >= maxRetries {
				logger.Info(
					"AliStorage maxRetries failed",
				)
				return fmt.Errorf(
					"AliStorage PutFileWithPart error after %d attempts: %s",
					attempt,
					err,
				)
			}

			time.Sleep(retryInterval)
			continue
		}
		// 如果上传成功，跳出循环
		break
	}
	logger.Debug("AliStorage %s PutFileWithPart success", key)
	return nil
}

func (_this *AliStorage) PutFile(
	bucket string,
	key string,
	localFile string,
	metadata map[string]string,
) error {
	return _this.PutObjectFromFile(bucket, key, localFile, metadata)
}

func (_this *AliStorage) PutFileWithMeta(
	bucket string,
	key string,
	localFile string,
	metadata *Metadata,
) error {
	bucketIns, err := _this.client.Bucket(bucket)
	if err != nil {
		logger.Warn("oss put file this.client.Bucket %s fail %s", localFile, err)
		return err
	}
	options := []oss.Option{}
	if metadata != nil {
		if metadata.Mime != "" {
			options = append(options, oss.ContentType(metadata.Mime))
		}
		if metadata.ContentEncoding != "" {
			options = append(options, oss.ContentEncoding(metadata.ContentEncoding))
		}
		if metadata.Acl != "" {
			options = append(options, oss.ObjectACL(oss.ACLType(metadata.Acl)))
		}
		if metadata.ContentDisposition != "" {
			options = append(options, oss.ContentDisposition(metadata.ContentDisposition))
		}
	}
	if len(options) > 0 {
		err = bucketIns.PutObjectFromFile(key, localFile, options...)
	} else {
		err = bucketIns.PutObjectFromFile(key, localFile)
	}
	if err != nil {
		logger.Warn("oss put file PutObjectFromFile %s fail %s", localFile, err)
		return err
	}
	return nil
}

func (_this *AliStorage) ListObjects(bucket string, prefix string) ([]Content, error) {
	bucketObj, _ := _this.client.Bucket(bucket)

	a := [32]Content{}
	result := a[0:0]

	prefixOption := oss.Prefix(prefix)
	maxKeys := oss.MaxKeys(1000)
	continueToken := ""
	for {

		lsRes, err := bucketObj.ListObjectsV2(
			prefixOption,
			maxKeys,
			oss.ContinuationToken(continueToken),
		)
		if err != nil {
			return result, err
		}
		for _, val := range lsRes.Objects {
			content := Content{
				Key:          val.Key,
				Size:         val.Size,
				ETag:         val.ETag,
				LastModified: val.LastModified,
			}
			result = append(result, content)
		}

		if lsRes.IsTruncated {
			continueToken = lsRes.NextContinuationToken
		} else {
			break
		}
	}
	return result, nil
}

func (_this *AliStorage) DeleteObject(bucket string, key string) error {
	bucketObj, err := _this.client.Bucket(bucket)

	err = bucketObj.DeleteObject(key)
	return err
}

func (_this *AliStorage) BatchDeleteObject(
	bucket string,
	filelist []string,
) (successList []string, e error) {
	bucketObj, e := _this.client.Bucket(bucket)
	if e != nil {
		return nil, e
	}
	successList = make([]string, 0, 10)
	for len(filelist) > 0 {
		var splitPos int
		if len(filelist) > MAX_BATCH_DELETE_COUNT {
			splitPos = MAX_BATCH_DELETE_COUNT
		} else {
			splitPos = len(filelist)
		}
		toDelFileList := filelist[0:splitPos]
		filelist = filelist[splitPos:]
		fmt.Println("del")
		res, e := bucketObj.DeleteObjects(toDelFileList)
		if e != nil {
			return successList, e
		}
		fmt.Println("DEL OK")
		successList = append(successList, res.DeletedObjects...)
	}
	return successList, e
}

func (_this *AliStorage) CopyObject(bucket string, srcKey string, destKey string) error {
	bucketObj, err := _this.client.Bucket(bucket)
	_, err = bucketObj.CopyObject(srcKey, destKey)
	return err
}

func (_this *AliStorage) SetObjectAcl(bucket string, key string, acl StorageAcl) error {
	bucketObj, err := _this.client.Bucket(bucket)
	err = bucketObj.SetObjectACL(key, oss.ACLType(string(acl)))

	return err
}

func (_this *AliStorage) IsObjectExist(bucket string, key string) (bool, error) {
	bucketObj, _ := _this.client.Bucket(bucket)
	isExists, err := bucketObj.IsObjectExist(key)

	return isExists, err
}

/*
*
OBS会把目录自己也作为一条记录返回  所以最少2条
*/
func (_this *AliStorage) IsNotEmptyDirExist(bucket string, prefix string) (bool, error) {
	bucketObj, _ := _this.client.Bucket(bucket)

	prefixOption := oss.Prefix(prefix)
	maxKeys := oss.MaxKeys(2)
	continueToken := ""
	output, err := bucketObj.ListObjectsV2(
		prefixOption,
		maxKeys,
		oss.ContinuationToken(continueToken),
	)
	if err != nil {
		return false, err
	}
	if len(output.Objects) == 2 {
		return true, nil
	}
	return false, nil
}

// 判断是否为归档文件
func (_this *AliStorage) IsArchive(bucket string, key string) (bool, error) {
	bucketObj, _ := _this.client.Bucket(bucket)
	//用于获取某个文件（Object）的元信息
	//参考文档：https://help.aliyun.com/document_detail/31984.html#title-ux9-txy-7cn
	//其中x-oss-storage-class
	//表示Object的存储类型，分别为：标准存储类型（Standard）、低频访问存储类型（IA）、归档存储类型（Archive）和冷归档存储类型（ColdArchive）。
	meta, err := bucketObj.GetObjectDetailedMeta(key)
	if err != nil {
		return false, err
	}
	if meta.Get("X-Oss-Storage-Class") == string(oss.StorageArchive) {
		//如果Bucket类型为Archive，且用户已经提交Restore请求，则响应头中会以x-oss-restore返回该Object的Restore状态，分如下几种情况：
		//如果没有提交Restore或者Restore已经超时，则不返回该字段。
		//如果已经提交Restore，且Restore没有完成，则返回的x-oss-restore值为ongoing-request=”true”。
		//如果已经提交Restore，且Restore已经完成，则返回的x-oss-restore值为ongoing-request=”false”, expiry-date=”Sun, 16 Apr 2017 08:12:33 GMT”，其中expiry-date是Restore完成后Object进入可读状态的过期时间。
		//参考https://help.aliyun.com/document_detail/31984.html?spm=5176.21213303.J_6704733920.12.5ee63eda3OqGiK&scm=20140722.S_help@@ææ¡£@@31984.S_0+os.ID_31984-RL_XDASOssDASRestore-OR_helpmain-V_2-P0_2
		if strings.Contains(meta.Get("x-oss-restore"), "false") {
			return false, nil
		}
		return true, nil
	} else {
		return false, nil
	}
}

// 解冻归档文件，如果没有归档就直接return，否则先解冻再return；
func (_this *AliStorage) RestoreArchive(bucket string, key string) (bool, error) {
	bucketObj, _ := _this.client.Bucket(bucket)
	// 检查是否为归档类型文件。
	meta, err := bucketObj.GetObjectDetailedMeta(key)
	if err != nil {
		return false, err
	}
	if meta.Get("X-Oss-Storage-Class") == string(oss.StorageArchive) {
		if len(meta.Get("x-oss-restore")) > 0 {
			logger.Debug(
				"%s storage class is %s %s",
				key,
				meta.Get("X-Oss-Storage-Class"),
				meta.Get("x-oss-restore"),
			)
			//已经在解冻或者解冻好了
			return true, nil
		}
		// 解冻归档类型文件。
		err = bucketObj.RestoreObject(key)
		if err != nil {
			return false, err
		}
	}
	return true, nil
}

func (_this *AliStorage) SetObjectMetaData(bucket string, key string, metadata *Metadata) error {
	if metadata == nil {
		return nil
	}
	var options = make([]oss.Option, 0)

	if metadata.Mime != "" {
		options = append(options, oss.ContentType(metadata.Mime))
	}
	if metadata.ContentEncoding != "" {
		options = append(options, oss.ContentEncoding(metadata.ContentEncoding))
	}
	if metadata.Acl != "" {
		options = append(options, oss.ObjectACL(oss.ACLType(metadata.Acl)))
	}
	if metadata.ContentDisposition != "" {
		options = append(options, oss.ContentDisposition(metadata.ContentDisposition))
	}

	bucketObj, _ := _this.client.Bucket(bucket)
	err := bucketObj.SetObjectMeta(key, options...)
	if err != nil {
		return err
	}
	return nil
}

type Policy struct {
	Version   string      `json:"Version"`
	Statement []Statement `json:"Statement"`
}
type Statement struct {
	Effect   string   `json:"Effect"`
	Action   []string `json:"Action"`
	Resource string   `json:"Resource"`
}
type TokenInfo struct {
	AccessKeyId     string
	AccessKeySecret string
	SecurityToken   string
	ExpireTimeStamp int64 //毫秒
}

var validTime int64 = 3600 //秒

func (_this *AliStorage) AssumeRole(policy Policy) (TokenInfo, error) {

	policyByte, err := json.Marshal(&policy)
	if err != nil {
		return TokenInfo{}, err
	}
	var assumeRoleFunc = func([]byte) (res *sts20150401.AssumeRoleResponse, err error) {
		request := sts20150401.AssumeRoleRequest{}
		request.SetPolicy(string(policyByte))
		request.SetRoleSessionName("test")
		request.SetRoleArn(_this.config.RoleArn)
		request.SetDurationSeconds(validTime)
		response, err := _this.stsClient.AssumeRole(&request)
		return response, err
	}

	var response *sts20150401.AssumeRoleResponse
	var aErr error
	var retryCount int = 3
	for i := 1; i <= retryCount; i++ {
		response, aErr = assumeRoleFunc(policyByte)
		if aErr != nil {
			respByte, _ := json.Marshal(response)
			logger.Warn(
				"assumeRole %s failed, as %v, resp %v",
				string(policyByte),
				aErr,
				string(respByte),
			)
			// 暂停i秒 ，再去尝试
			if i == retryCount {
				break
			}
			time.Sleep(time.Duration(i) * time.Second)
		} else {
			break
		}
	}

	if aErr != nil {
		return TokenInfo{}, aErr
	}

	expireTimeStamp := time.Now().UnixNano()/1e6 + (validTime-10)*1000
	res := TokenInfo{AccessKeyId: *response.Body.Credentials.AccessKeyId,
		AccessKeySecret: *response.Body.Credentials.AccessKeySecret,
		SecurityToken:   *response.Body.Credentials.SecurityToken,
		ExpireTimeStamp: expireTimeStamp,
	}
	return res, nil

}

// 过期时间:秒
func (_this *AliStorage) SignFile(remoteFilePath string, expiredTime int64) (error, string) {
	return _this.SignFile2(_this.config.Bucket, remoteFilePath, expiredTime)
}

// 过期时间:秒
func (_this *AliStorage) SignFile2(b, remoteFilePath string, expiredTime int64) (error, string) {
	objectName := remoteFilePath
	// 获取存储空间。
	bucket, err := _this.client.Bucket(b)
	if err != nil {
		return err, ""
	}
	// 使用签名URL将OSS文件下载到流。
	signedURL, err := bucket.SignURL(objectName, oss.HTTPGet, expiredTime)
	if err != nil {
		return err, ""
	}
	return nil, _this.replaceInterPoint(signedURL)
}

// 针对可能配置启用的内网域名，提供统一函数讲内网域名转成外网域名
func (_this *AliStorage) replaceInterPoint(url string) string {
	//url本身就是个完整的类似 P_OSS_BUCKET_P: "plaso-school", //oss bucket
	//    P_OSS_REGION_P: "oss-cn-hangzhou", //oss region
	//   https://plaso-school.oss-cn-hangzhou-internal.aliyuncs.com  ${bucket},${endPoint}
	if !_this.config.Internal || len(_this.config.EndpointInternal) == 0 {
		return url
	}
	signUrl := strings.ReplaceAll(
		url,
		_this.config.Bucket+"."+strings.TrimPrefix(_this.config.EndpointInternal, "https://"),
		_this.config.Host,
	)
	return signUrl

}

// SignFileForDownload 过期时间:秒
func (_this *AliStorage) SignFileForDownload(
	remoteFilePath string,
	expiredTime int64,
	downLoadFilename string,
) string {
	objectName := remoteFilePath
	// 获取存储空间。
	bucket, err := _this.client.Bucket(_this.config.Bucket)
	if err != nil {
		return "err"
	}
	//https://help.aliyun.com/zh/oss/user-guide/set-the-file-name-for-downloading-an-oss-file?spm=a2c4g.11186623.0.i18
	contentDispositionOption := oss.ResponseContentDisposition(
		file.GetContentDisposition(downLoadFilename),
	)
	// 使用签名URL将OSS文件下载到流。
	signedURL, err := bucket.SignURL(objectName, oss.HTTPGet, expiredTime, contentDispositionOption)
	return _this.replaceInterPoint(signedURL)
}

func (_this *AliStorage) GetDirToken(remoteDir string) map[string]interface{} {
	resource := fmt.Sprintf("acs:oss:*:*:%s/%s*", _this.config.Bucket, remoteDir)
	policy := Policy{Version: "1", Statement: []Statement{{
		Effect:   "Allow",
		Action:   []string{"oss:PutObject", "oss:GetObject", "oss:PutObjectAcl"},
		Resource: resource,
	}}}
	osstoken, err := _this.AssumeRole(policy)
	if err != nil {
		logger.Warn("get remoteDir ossToken %s fail %s", remoteDir, err)
		return nil
	}
	res := make(map[string]interface{})
	res["accessKeyId"] = osstoken.AccessKeyId
	res["accessKeySecret"] = osstoken.AccessKeySecret
	res["stsToken"] = osstoken.SecurityToken
	res["bucket"] = _this.config.Bucket
	res["region"] = _this.config.Region
	res["provider"] = _this.config.Provider
	res["expire"] = osstoken.ExpireTimeStamp
	res["uploadPath"] = remoteDir
	res["host"] = _this.config.Host
	return res
}

func (_this *AliStorage) GetDirToken2(remoteDir string) *StorageToken {
	resource := fmt.Sprintf("acs:oss:*:*:%s/%s*", _this.config.Bucket, remoteDir)
	policy := Policy{Version: "1", Statement: []Statement{{
		Effect:   "Allow",
		Action:   []string{"oss:PutObject", "oss:GetObject", "oss:PutObjectAcl"},
		Resource: resource,
	}}}
	osstoken, err := _this.AssumeRole(policy)
	if err != nil {
		logger.Warn("get remoteDir ossToken %s fail %s", remoteDir, err)
		return nil
	}
	res := &StorageToken{
		AccessKeyID:     osstoken.AccessKeyId,
		AccessKeySecret: osstoken.AccessKeySecret,
		StsToken:        osstoken.SecurityToken,
		Bucket:          _this.config.Bucket,
		Region:          _this.config.Region,
		Provider:        _this.config.Provider,
		Expire:          osstoken.ExpireTimeStamp,
		UploadPath:      remoteDir,
		Host:            _this.config.Host,
		EndPoint:        _this.config.Endpoint,
		Path:            remoteDir,
		CdnDomain:       _this.config.CdnDomain,
	}
	return res
}

type Action string

const PutObjectAction Action = "oss:PutObject"
const PutObjectAclAction Action = "oss:PutObjectAcl"
const GetObjectAction Action = "oss:GetObject"

func (_this *AliStorage) GetDirTokenWithAction(
	remoteDir string,
	actions ...Action,
) (bool, *StorageToken) {
	resource := fmt.Sprintf("acs:oss:*:*:%s/%s*", _this.config.Bucket, remoteDir)
	actionParam := make([]string, len(actions))
	for i := 0; i < len(actions); i++ {
		actionParam[i] = string(actions[i])
	}
	policy := Policy{Version: "1", Statement: []Statement{{
		Effect:   "Allow",
		Action:   actionParam,
		Resource: resource,
	}}}
	osstoken, err := _this.AssumeRole(policy)
	if err != nil {
		logger.Warn("get remoteDir ossToken %s fail %s", remoteDir, err)
		return false, nil
	}
	res := &StorageToken{
		AccessKeyID:     osstoken.AccessKeyId,
		AccessKeySecret: osstoken.AccessKeySecret,
		StsToken:        osstoken.SecurityToken,
		Bucket:          _this.config.Bucket,
		Region:          _this.config.Region,
		Provider:        _this.config.Provider,
		Expire:          osstoken.ExpireTimeStamp,
		UploadPath:      remoteDir,
		Host:            _this.config.Host,
		EndPoint:        _this.config.Endpoint,
		Path:            remoteDir,
	}
	return true, res
}

func (_this *AliStorage) GetObjectMeta(bucket string, key string) (*Content, error) {
	bucketObj, _ := _this.client.Bucket(bucket)

	props, err := bucketObj.GetObjectMeta(key)
	if err != nil {
		return nil, err
	}

	res := &Content{Key: key}
	res.ETag = props.Get("Etag")
	res.Size, _ = strconv.ParseInt(props.Get("Content-Length"), 10, 64)
	res.LastModified, err = time.Parse(http.TimeFormat, props.Get("Last-Modified"))
	return res, nil
}
