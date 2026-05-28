package driver

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"time"
	"tools-thinker/support/file"
	"tools-thinker/support/logger"
	"tools-thinker/support/storage/config"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/auth/global"
	iam "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/iam/v3"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/iam/v3/model"

	// 使用华为sdk来实现移动和云存储能力，基于以下原因：
	// 1.移动和云本质上基于华为云来实现的
	// 2.iam方面只能通过华为的sdk 来获取临时ak，sk
	// 3.华为云存储sdk 在使用过程中没有出现问题
	// 4.移动和云 通过的源码sdk ,在使用临时ak sk 和securitytoken 方式处理对象时，没法传入securitytoken的header参数,所以弃用
	//"tools-thinker/support/storage/storagecore/ecloud-obs-go-sdk/obs"
	"github.com/huaweicloud/huaweicloud-sdk-go-obs/obs"
)

const (
	ObsPutObjectAction   = "obs:object:PutObject"
	ObsGetObjectAction   = "obs:object:GetObject"
	OBS_MAX_DELETE_COUNT = 1000
)

type ecloudObsStorage struct {
	client    *obs.ObsClient
	config    *config.StorageConfig
	iamClient *iam.IamClient
}

func NewEcloudObsStorage(config *config.StorageConfig) (*ecloudObsStorage, error) {
	ak := config.AccessKeyId
	sk := config.AccessKeySecret
	endpoint := config.Endpoint
	hwObsClient, err := obs.New(ak, sk, endpoint)
	if err != nil {
		return nil, err
	}
	auth, err := global.NewCredentialsBuilder().
		WithAk(ak).
		WithSk(sk).
		SafeBuild()
	if err != nil {
		return nil, err
	}

	clientBuilder, err2 := iam.IamClientBuilder().
		WithEndpoints([]string{config.StsEndPoint}). // 注意这里的StsEndpoint与endpoint不同，配置文件中需要留意
		WithCredential(auth).
		SafeBuild()

	if err2 != nil {
		return nil, err
	}
	iamClient := iam.NewIamClient(clientBuilder)
	st := &ecloudObsStorage{
		client:    hwObsClient,
		config:    config,
		iamClient: iamClient,
	}
	return st, nil
}

func (s *ecloudObsStorage) GetObject(bucket string, key string) ([]byte, error) {
	param := &obs.GetObjectInput{}
	param.Key = key
	param.Bucket = bucket
	outPut, err := s.client.GetObject(param)
	if err != nil {
		logger.Warn("GetObject s.client.GetObject key %v fail,%s", key, err)
		return nil, err
	}
	// 流式下载对象
	// output.Body 在使用完毕后必须关闭，否则会造成连接泄漏。
	defer outPut.Body.Close()
	// 读取对象内容
	var body bytes.Buffer
	// 使用 io.Copy 更高效地读取内容
	_, err = io.Copy(&body, outPut.Body)
	// 正确处理 io.EOF
	// io.Copy 不会因为 EOF 而返回错误
	if err != nil {
		return nil, err
	}
	return body.Bytes(), nil
}

func (s *ecloudObsStorage) GetFile(bucket string, key string, localFile string) error {
	param := &obs.GetObjectInput{}
	param.Key = key
	param.Bucket = bucket
	outPut, err := s.client.GetObject(param)
	if err != nil {
		logger.Warn("GetFile s.client.GetObject key %v fail,%s", key, err)
		return err
	}
	// 流式下载对象
	// outPut.Body 在使用完毕后必须关闭，否则会造成连接泄漏。
	defer outPut.Body.Close()
	// 设置文件路径,初始化本地文件
	os.MkdirAll(path.Dir(localFile), 0755)
	fd, err := os.OpenFile(localFile, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0755)
	if err != nil {
		logger.Warn("open file %v fail,%s", localFile, err)
		return err
	}
	defer fd.Close()
	// 使用io.Copy直接将数据流写入文件，减少内存使用
	if _, err = io.Copy(fd, outPut.Body); err != nil {
		logger.Warn("write file %v fail, %s", localFile, err)
		return err
	}
	return nil
}

func convertMapToMetaData(metadata map[string]string) *Metadata {
	var metadataParam *Metadata
	if metadata != nil {
		metadataParam = &Metadata{}
		for metaKey, metaValue := range metadata {
			if metaKey == "mime" {
				metadataParam.Mime = metaValue
				continue
			}
			if metaKey == "Content-Encoding" {
				metadataParam.ContentEncoding = metaValue
				continue
			}
			if metaKey == "acl" {
				metadataParam.Acl = metaValue
				continue
			}
			if metaKey == "Content-Disposition" {
				metadataParam.ContentDisposition = metaValue
				continue
			}
			logger.Warn("cant support metadata %s %s", metaKey, metaValue)
		}
	}
	return metadataParam
}

func (s *ecloudObsStorage) PutObject(
	bucket string,
	key string,
	data []byte,
	metadata map[string]string,
) error {
	err := s.PutObjectWithMeta(bucket, key, data, convertMapToMetaData(metadata))
	if err != nil {
		logger.Warn("PutObject %v fail,%s", key, err)
		return err
	}
	return nil
}

// https://support.huaweicloud.com/sdk-go-devg-obs/obs_23_0402.html
// 可以上传小于5GB的文件
func (s *ecloudObsStorage) PutObjectWithMeta(
	bucket string,
	key string,
	data []byte,
	metadata *Metadata,
) error {
	input := getPutObjectInput(bucket, key, data, metadata)
	output, err := s.client.PutObject(input)
	if err != nil {
		logger.Warn("PutObjectWithMeta %v fail,%s", key, err)
		return err
	} else {
		logger.Debug("PutObjectWithMeta %v success %v ", key, output.StatusCode)
	}
	return nil
}

func getPutObjectInput(
	bucket string,
	key string,
	data []byte,
	metadata *Metadata,
) *obs.PutObjectInput {
	input := &obs.PutObjectInput{}
	// 指定存储桶名称
	input.Bucket = bucket
	// 指定上传对象
	input.Key = key
	// 流式上传,[]byte转为 io.Reader
	input.Body = bytes.NewReader(data)
	input.PutObjectBasicInput = getPutObjectBasicInput(bucket, key, metadata)
	return input
}

// PutFile 大文件需要分块
// https://support.huaweicloud.com/sdk-go-devg-obs/obs_23_0403.html
// 支持0-5GB的文件
func (s *ecloudObsStorage) PutFile(
	bucket string,
	key string,
	localFile string,
	metadata map[string]string,
) error {
	metadataObj := convertMapToMetaData(metadata)
	err := s.PutFileWithMeta(bucket, key, localFile, metadataObj)
	if err != nil {
		logger.Warn("PutFile %v fail,%s", key, err)
		return err
	}
	return nil
}

// https://support.huaweicloud.com/sdk-go-devg-obs/obs_23_0402.html
// 可以上传小于5GB的文件
func (s *ecloudObsStorage) PutFileWithMeta(
	bucket string,
	key string,
	srcFile string,
	metadata *Metadata,
) error {
	var putFileParam = getPutFileInput(bucket, key, srcFile, metadata)
	outPut, err := s.client.PutFile(putFileParam)
	if err != nil {
		logger.Warn("PutFile %v fail,%s", key, err)
		return err
	} else {
		logger.Debug("PutFile %v success %v ", key, outPut.StatusCode)
	}
	return err
}

func getPutFileInput(
	bucket string,
	key string,
	localFile string,
	metadata *Metadata,
) *obs.PutFileInput {
	input := &obs.PutFileInput{}
	// 指定本地文件
	input.SourceFile = localFile
	input.PutObjectBasicInput = getPutObjectBasicInput(bucket, key, metadata)
	return input
}

func getPutObjectBasicInput(
	bucket string,
	key string,
	metadata *Metadata,
) obs.PutObjectBasicInput {
	var res = obs.PutObjectBasicInput{}
	// 指定存储桶名称
	res.Bucket = bucket
	// 指定上传对象
	res.Key = key
	if metadata != nil {
		if len(metadata.Mime) > 0 {
			res.HttpHeader.ContentType = metadata.Mime
		}
		if len(metadata.ContentEncoding) > 0 {
			res.HttpHeader.ContentEncoding = metadata.ContentEncoding
		}
		if len(metadata.Acl) > 0 {
			res.ACL = convertStorageAclToObsAcl(StorageAcl(metadata.Acl))
		}
		if len(metadata.ContentDisposition) > 0 {
			res.HttpHeader.ContentDisposition = metadata.ContentDisposition
		}
	}
	return res
}

// PutObjectFromFile 注意小文件可以，大文件不能走这个函数
func (s *ecloudObsStorage) PutObjectFromFile(
	bucket string,
	key string,
	localFile string,
	metadata map[string]string,
) error {
	data, err := os.ReadFile(localFile)
	if err != nil {
		logger.Warn("oss put file %s fail %s", localFile, err)
		return err
	}
	err = s.PutObject(bucket, key, data, metadata)
	return err
}

// PutFileWithPart 支持分片上传，支持5GB以上的文件上传
// 注意 断点续传上传接口传入的文件大小至少要100K以上
func (s *ecloudObsStorage) PutFileWithPart(
	bucket string,
	key string,
	srcFile string,
	metadata *Metadata,
	partSize int64,
) error {
	uploadParam := &obs.UploadFileInput{
		UploadFile: srcFile,
	}
	uploadParam.EnableCheckpoint = true // 开启断点续传模式
	uploadParam.PartSize = partSize     // 指定分段大小为
	uploadParam.TaskNum = 5             // 指定分段上传时的最大并发数
	uploadParam.ObjectOperationInput = contructObjectOperationInputParam(bucket, key, metadata)

	// 重试配置
	maxRetries := 3
	retryInterval := time.Second * 2
	attempt := 0
	for {
		attempt++
		outPut, err := s.client.UploadFile(uploadParam)
		if err != nil {
			logger.Info(
				"ecloudStorage PutFileWithPart failed, will retry %d/%d in %s",
				attempt,
				maxRetries,
				retryInterval,
			)
			if attempt >= maxRetries {
				logger.Info(
					"ecloudStorage maxRetries failed",
				)
				return fmt.Errorf(
					"ecloudStorage PutFileWithPart error after %d attempts: %s",
					attempt,
					err,
				)
			}

			time.Sleep(retryInterval)
			continue
		}
		// 如果上传成功，跳出循环
		logger.Debug("ecloudStorage PutFileWithPart %v success %v ", key, outPut.StatusCode)
		break
	}
	return nil
}

func contructObjectOperationInputParam(
	bucket string,
	key string,
	metadata *Metadata,
) obs.ObjectOperationInput {
	var res = obs.ObjectOperationInput{}
	res.Metadata = make(map[string]string)
	res.Bucket = bucket
	res.Key = key
	if metadata != nil {
		if len(metadata.Mime) > 0 {
			res.Metadata["Content-Type"] = metadata.Mime
		}
		if len(metadata.ContentEncoding) > 0 {
			res.Metadata["Content-Encoding"] = metadata.ContentEncoding
		}
		if len(metadata.Acl) > 0 {
			res.ACL = convertStorageAclToObsAcl(StorageAcl(metadata.Acl))
		}
		if len(metadata.ContentDisposition) > 0 {
			res.Metadata["Content-Disposition"] = metadata.ContentDisposition
		}
	}
	return res
}

func (s *ecloudObsStorage) ListObjects(bucket string, prefix string) ([]Content, error) {
	input := &obs.ListObjectsInput{}
	a := [32]Content{}
	result := a[0:0]
	input.Bucket = bucket
	input.Prefix = prefix
	input.MaxKeys = 1000
	for {
		output, err := s.client.ListObjects(input)
		if err != nil {
			return result, err
		}
		for _, val := range output.Contents {
			content := Content{
				Key:          val.Key,
				Size:         val.Size,
				ETag:         val.ETag,
				LastModified: val.LastModified,
			}
			result = append(result, content)
		}

		if output.IsTruncated {
			input.Marker = output.NextMarker
		} else {
			break
		}

	}
	return result, nil
}

func (s *ecloudObsStorage) DeleteObject(bucket string, key string) error {
	input := &obs.DeleteObjectInput{}
	input.Bucket = bucket
	input.Key = key
	_, err := s.client.DeleteObject(input)
	return err
}

func (s *ecloudObsStorage) BatchDeleteObject(
	bucket string,
	filelist []string,
) (successList []string, e error) {
	successList = []string{}
	for len(filelist) > 0 {
		var splitSize int
		if len(filelist) > OBS_MAX_DELETE_COUNT {
			splitSize = OBS_MAX_DELETE_COUNT
		} else {
			splitSize = len(filelist)
		}
		toDelFileList := filelist[0:splitSize]
		filelist = filelist[splitSize:]
		var objects = make([]obs.ObjectToDelete, 0, len(toDelFileList))
		for _, file := range toDelFileList {
			objects = append(objects, obs.ObjectToDelete{
				Key: file,
			})
		}
		input := &obs.DeleteObjectsInput{
			Bucket:  bucket,
			Objects: nil,
		}
		input.Objects = objects
		delRes, err := s.client.DeleteObjects(input)
		if err != nil {
			return successList, err
		}
		for _, deleted := range delRes.Deleteds {
			successList = append(successList, deleted.Key)
		}
	}
	return successList, nil
}

func (s *ecloudObsStorage) CopyObject(bucket string, srcKey string, destKey string) error {
	input := &obs.CopyObjectInput{}
	input.Bucket = bucket
	input.Key = destKey
	input.CopySourceBucket = bucket
	input.CopySourceKey = srcKey
	_, err := s.client.CopyObject(input)
	return err
}

// 转换已有的acl key为obs的aclkey
func convertStorageAclToObsAcl(acl StorageAcl) obs.AclType {
	switch acl {
	case AclPrivate:
		return obs.AclPrivate
	case AclPublicRead:
		return obs.AclPublicRead
	case AclPublicReadWrite:
		return obs.AclPublicReadWrite
	case AclDefault:
		return obs.AclPrivate // 默认返回AclPrivate，因为华为不支持AclDefault
	default:
		return obs.AclPrivate // 如果没有匹配的值，返回默认的AclPrivate
	}
}

// public-read-write 和 public-read 针对对象效果一致，都是公共读权限
// 设在对象上，所有人可以获取该对象内容和元数据。
// https://support.huaweicloud.com/perms-cfg-obs/obs_40_0005.html#section3
func (s *ecloudObsStorage) SetObjectAcl(bucket string, key string, acl StorageAcl) error {
	input := &obs.SetObjectAclInput{}
	input.Bucket = bucket
	input.Key = key
	input.ACL = convertStorageAclToObsAcl(acl)
	_, err := s.client.SetObjectAcl(input)
	return err
}

func (s *ecloudObsStorage) SetObjectMetaData(bucket string, key string, metadata *Metadata) error {
	input := &obs.SetObjectMetadataInput{}
	input.Bucket = bucket
	input.Key = key
	if metadata != nil {
		if metadata.Mime != "" {
			input.ContentType = metadata.Mime
		}
		if metadata.ContentEncoding != "" {
			input.ContentEncoding = metadata.ContentEncoding
		}
		if metadata.ContentDisposition != "" {
			input.ContentDisposition = metadata.ContentDisposition
		}
	}
	if metadata.Acl != "" {
		err := s.SetObjectAcl(bucket, key, StorageAcl(metadata.Acl))
		if err != nil {
			return err
		}
	}
	_, err := s.client.SetObjectMetadata(input)
	return err
}

func (s *ecloudObsStorage) IsObjectExist(bucket string, key string) (bool, error) {
	input := &obs.HeadObjectInput{}
	input.Bucket = bucket
	input.Key = key
	_, err := s.client.HeadObject(input)
	return err == nil, nil

}

/*
*
OBS会把目录自己也作为一条记录返回  所以最少2条
*/
func (s *ecloudObsStorage) IsNotEmptyDirExist(bucket string, prefix string) (bool, error) {
	input := &obs.ListObjectsInput{}
	input.Bucket = bucket
	input.Prefix = prefix
	input.MaxKeys = 2
	output, err := s.client.ListObjects(input)
	if err != nil {
		return false, err
	}
	if len(output.Contents) == 2 {
		return true, nil
	}
	return false, nil
}

func (s *ecloudObsStorage) IsPublicAcl(bucket string, key string) (bool, error) {
	input := &obs.GetObjectAclInput{}
	input.Bucket = bucket
	input.Key = key
	output, err := s.client.GetObjectAcl(input)
	if err != nil {
		fmt.Printf("GetObjectAcl err : %v", err)
		return false, err
	}
	/*personJSON, err1 := json.Marshal(output)
	if err1 != nil {
		fmt.Printf("Marshal err : %v", err1)
		return false, err1
	}*/
	/*ss := string(personJSON)
	fmt.Println(ss)*/
	// 检查ACL中的Grant信息，判断是否有公共读权限
	isPublicRead := false
	for _, grant := range output.Grants {
		if (grant.Grantee.URI == obs.GroupAllUsers && grant.Permission == "READ") ||
			(grant.Grantee.Type == obs.GranteeGroup && grant.Permission == "READ") {
			isPublicRead = true
			break
		}
	}
	return isPublicRead, nil

}

func (s *ecloudObsStorage) execCreateSignedUrl(
	bucket string,
	key string,
	expires int,
	method obs.HttpMethodType,
) (*obs.CreateSignedUrlOutput, error) {
	input := &obs.CreateSignedUrlInput{}
	input.Bucket = bucket
	input.Key = key
	input.Expires = expires
	input.Method = method
	return s.client.CreateSignedUrl(input)
}

func getSignFileCreateSignedUrlInput(
	bucket string,
	key string,
	expires int,
) *obs.CreateSignedUrlInput {
	input := getCreateSignedUrlInput(
		bucket,
		key,
		expires,
		obs.HttpMethodGet,
	)
	return input
}

func getSignFileForDownloadCreateSignedUrlInput(
	bucket string,
	key string,
	expires int,
	fileName string,
) *obs.CreateSignedUrlInput {
	input := getCreateSignedUrlInput(
		bucket,
		key,
		expires,
		obs.HttpMethodGet,
	)
	params := file.GetContentDisposition(fileName)
	input.QueryParams = map[string]string{
		"response-content-disposition": params,
	}
	return input
}

func getCreateSignedUrlInput(
	bucket string,
	key string,
	expires int,
	method obs.HttpMethodType,
) *obs.CreateSignedUrlInput {
	input := &obs.CreateSignedUrlInput{}
	input.Bucket = bucket
	input.Key = key
	input.Expires = expires
	input.Method = method
	return input
}

func (s *ecloudObsStorage) getTmpToken(
	key string,
	osbAction []string,
	expirationDate int32,
) (*StorageToken, error) {
	request := &model.CreateTemporaryAccessKeyByTokenRequest{}
	request.Body = &model.CreateTemporaryAccessKeyByTokenRequestBody{}
	request.Body.Auth = &model.TokenAuth{}
	// var DurationSeconds int32 = 3600 * 6
	resource := []string{fmt.Sprintf("obs:*:*:object:%s/%s*", s.config.Bucket, key)}
	request.Body.Auth.Identity = &model.TokenAuthIdentity{
		Methods: []model.TokenAuthIdentityMethods{model.GetTokenAuthIdentityMethodsEnum().TOKEN},
		Token: &model.IdentityToken{
			DurationSeconds: &expirationDate,
		},
		Policy: &model.ServicePolicy{
			Version: "1.1",
			Statement: []model.ServiceStatement{{
				Action:    osbAction,
				Effect:    model.GetServiceStatementEffectEnum().ALLOW,
				Condition: nil,
				Resource:  &resource,
			}},
		},
	}
	response, err := s.iamClient.CreateTemporaryAccessKeyByToken(request)
	if err != nil {
		logger.Warn(
			"ecloudObsStorage getTmpToken %s fail %s",
			key,
			err,
		)
		return nil, err
	}
	res := &StorageToken{
		AccessKeyID:     response.Credential.Access,
		AccessKeySecret: response.Credential.Secret,
		StsToken:        response.Credential.Securitytoken,
		Bucket:          s.config.Bucket,
		Region:          s.config.Region,
		Provider:        s.config.Provider,
		Expire:          int64(validTime),
		UploadPath:      key,
		Host:            s.config.Host,
		EndPoint:        s.config.Endpoint,
		Path:            key,
		CdnDomain:       s.config.CdnDomain,
	}
	return res, nil
}

func (s *ecloudObsStorage) GetDirToken2(remoteDir string) *StorageToken {
	var (
		durationSeconds int32 = 3600 * 6
		execActions           = []string{
			ObsPutObjectAction,
			ObsGetObjectAction,
		}
	)
	res, err := s.getTmpToken(
		remoteDir,
		execActions,
		durationSeconds,
	)
	if err != nil {
		logger.Warn(
			"ecloudObsStorage GetDirToken2 %s fail %s",
			remoteDir,
			err,
		)
		return nil
	}
	return res
}

func (s *ecloudObsStorage) GetDirToken(remoteDir string) map[string]interface{} {
	t := s.GetDirToken2(remoteDir)
	if t == nil {
		return nil
	}
	res := make(map[string]interface{})
	res["accessKeyId"] = t.AccessKeyID
	res["accessKeySecret"] = t.AccessKeySecret
	res["stsToken"] = t.StsToken
	res["bucket"] = t.Bucket
	res["region"] = t.Region
	res["provider"] = t.Provider
	res["expire"] = t.Expire
	res["uploadPath"] = t.UploadPath
	res["host"] = t.Host
	res["endPoint"] = t.EndPoint
	return res
}

func (s *ecloudObsStorage) GetDirTokenWithAction(
	remoteDir string,
	actions ...Action,
) (bool, *StorageToken) {
	var (
		execActions   = []string{}
		execActionMap = map[string]bool{ // 用于去重
			ObsPutObjectAction: false,
			ObsGetObjectAction: false,
		}
		durationSeconds int32 = 3600 * 6
	)
	for _, action := range actions {
		switch action {
		case PutObjectAction, PutObjectAclAction:
			if !execActionMap[ObsPutObjectAction] {
				execActions = append(execActions, ObsPutObjectAction)
				execActionMap[ObsPutObjectAction] = true
			}
		default:
			if !execActionMap[ObsGetObjectAction] {
				execActions = append(execActions, ObsGetObjectAction)
				execActionMap[ObsGetObjectAction] = true
			}
		}
	}
	res, err := s.getTmpToken(remoteDir, execActions, durationSeconds)
	if err != nil {
		logger.Warn(
			"ecloudObsStorage GetDirTokenWithAction %s fail %s",
			remoteDir,
			err,
		)
		return false, nil
	}
	return true, res
}

func (s *ecloudObsStorage) SignFile(remoteDir string, expiredTime int64) (error, string) {
	return s.SignFile2(
		s.config.Bucket,
		remoteDir,
		int64(expiredTime),
	)
}

func (s *ecloudObsStorage) SignFile2(bucket, remoteDir string, expiredTime int64) (error, string) {
	expre := int(expiredTime)
	input := getSignFileCreateSignedUrlInput(bucket, remoteDir, expre)
	getObjectOutput, err := s.client.CreateSignedUrl(input)
	if err != nil {
		return err, ""
	}
	return nil, getObjectOutput.SignedUrl
}

// 过期时间:秒
func (s *ecloudObsStorage) SignFileForDownload(
	remoteFilePath string,
	expiredTime int64,
	downLoadFilename string,
) string {
	expre := int(expiredTime)
	input := getSignFileForDownloadCreateSignedUrlInput(
		s.config.Bucket,
		remoteFilePath,
		expre,
		downLoadFilename,
	)
	getObjectOutput, err := s.client.CreateSignedUrl(input)
	if err != nil {
		errStr := fmt.Sprintf(
			"ecloudObsStorage SignFileForDownload key %s fail %s",
			remoteFilePath,
			err.Error(),
		)
		logger.Warn(errStr)
		return ""
	}
	return getObjectOutput.SignedUrl
}

func (s *ecloudObsStorage) GetObjectMeta(bucket string, key string) (*Content, error) {
	output, err := s.client.GetObjectMetadata(&obs.GetObjectMetadataInput{
		Bucket: bucket,
		Key:    key,
	})
	if err != nil {
		logger.Warn("ecloudObsStorage GetObjectMeta key %s fail %s", key, err.Error())
		return nil, err
	}
	// 将获取到的元数据转换为Content结构体
	res := &Content{
		Key: key,
		// ETag值去除双引号
		ETag: strings.Trim(output.ETag, "\""),
		Size: output.ContentLength,
		// LastModified为time.Time类型，无需转换
		LastModified: output.LastModified,
	}
	return res, nil
}

func (s *ecloudObsStorage) RestoreArchive(bucket string, key string) (bool, error) {
	logger.Warn("RestoreArchive method is not implemented")
	return false, errors.New("cant support RestoreArchive")
}

func (s *ecloudObsStorage) IsArchive(bucket string, key string) (bool, error) {
	logger.Warn("IsArchive method is not implemented")
	return false, errors.New("cant support IsArchive")
}

// 判断字符串是否包含中文字符
func containsChinese(str string) bool {
	for _, r := range str {
		if r >= '\u4e00' && r <= '\u9fff' {
			return true
		}
	}
	return false
}
