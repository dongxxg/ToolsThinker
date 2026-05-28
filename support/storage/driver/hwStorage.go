package driver

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path"
	"strings"
	"tools-thinker/support/logger"
	"tools-thinker/support/storage/config"

	"github.com/north-team/huawei-obs-sdk-go/obs"
)

const (
	validTimeHuaweiObs      int = 3600 //秒
	HUAWEI_MAX_DELETE_COUNT     = 1000
)

type hwStorage struct {
	config *config.StorageConfig
	client *obs.ObsClient
}

func (s *hwStorage) IsPublicAcl(bucket string, key string) (bool, error) {
	panic("implement me")
}
func (this *hwStorage) GetObject(bucket string, key string) ([]byte, error) {
	input := &obs.GetObjectInput{}
	input.Bucket = bucket
	input.Key = key
	output, err := this.client.GetObject(input)
	if err != nil {
		logger.Warn("obs get remotefile %s fail %s", key, err.Error())
		return nil, err
	}
	defer output.Body.Close()
	length := output.ContentLength
	if length > 1000*1000 {
		logger.Warn(
			"obs file is large then 1M ,you should download file then process. %s/%s",
			bucket,
			key,
		)
	}

	buf := new(bytes.Buffer)
	_, readErr := io.Copy(buf, output.Body)
	if readErr != nil {
		logger.Warn("obs get remotefile %s success ,but read fail %s", key, readErr.Error())
		return nil, readErr
	}
	return buf.Bytes(), nil
}

func (this *hwStorage) GetFile(bucket string, key string, localFile string) error {
	input := &obs.GetObjectInput{}
	input.Bucket = bucket
	input.Key = key
	output, err := this.client.GetObject(input)
	if err != nil {
		logger.Warn("obs get remotefile %v fail %s", key, err.Error())
		return err
	}
	defer output.Body.Close()
	os.MkdirAll(path.Dir(localFile), 0755)
	fd, err := os.OpenFile(localFile, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0755)
	if err != nil {
		logger.Warn("open file %v fail", key)
		return err
	}
	defer fd.Close()
	io.Copy(fd, output.Body)
	return nil
}

func (this *hwStorage) PutObject(
	bucket string,
	key string,
	data []byte,
	metadata map[string]string,
) error {
	input := &obs.PutObjectInput{}
	input.Bucket = bucket
	input.Key = key
	input.Body = bytes.NewReader(data)
	if metadata != nil {
		if mime, ok := metadata["mime"]; ok {
			input.ContentType = mime
			delete(metadata, "mime")
		}
		if encoding, ok := metadata["Content-Encoding"]; ok {
			delete(metadata, "Content-Encoding")
			metadata["ContentEncoding"] = encoding
		}
		if acl, ok := metadata["x-oss-object-acl"]; ok {
			input.ACL = obs.AclType(acl)
			delete(metadata, "x-oss-object-acl")
		}
		input.Metadata = metadata
	}
	_, err := this.client.PutObject(input)

	return err
}

func (this *hwStorage) PutObjectFromFile(
	bucket string,
	key string,
	filePath string,
	metadata map[string]string,
) error {
	return this.PutFile(bucket, key, filePath, metadata)
}

func (this *hwStorage) PutObjectWithMeta(
	bucket string,
	key string,
	data []byte,
	meta *Metadata,
) error {
	input := &obs.PutObjectInput{}
	input.Bucket = bucket
	input.Key = key
	input.Body = bytes.NewReader(data)
	if meta.Mime != "" {
		input.ContentType = meta.Mime
	}
	if meta.ContentEncoding != "" {
		input.Metadata = make(map[string]string)
		input.Metadata["ContentEncoding"] = meta.ContentEncoding
	}
	if meta.Acl != "" {
		input.ACL = obs.AclType(meta.Acl)
	}
	_, err := this.client.PutObject(input)

	return err
}

func (this *hwStorage) PutFile(
	bucket string,
	key string,
	localFile string,
	metadata map[string]string,
) error {
	input := &obs.PutFileInput{}
	input.Bucket = bucket
	input.Key = key
	input.SourceFile = localFile // localfile为待上传的本地文件路径，需要指定到具体的文件名
	if metadata != nil {
		if mime, ok := metadata["mime"]; ok {
			input.ContentType = mime
			delete(metadata, "mime")
		}
		if encoding, ok := metadata["Content-Encoding"]; ok {
			delete(metadata, "Content-Encoding")
			metadata["ContentEncoding"] = encoding
		}
		if acl, ok := metadata["acl"]; ok {
			input.ACL = obs.AclType(acl)
			delete(metadata, "acl")
		}
		input.Metadata = metadata
	}
	_, err := this.client.PutFile(input)
	return err
}

func (this *hwStorage) PutFileWithMeta(
	bucket string,
	key string,
	localFile string,
	meta *Metadata,
) error {
	input := &obs.PutFileInput{}
	input.Bucket = bucket
	input.Key = key
	input.SourceFile = localFile // localfile为待上传的本地文件路径，需要指定到具体的文件名
	if meta.Mime != "" {
		input.ContentType = meta.Mime
	}
	if meta.ContentEncoding != "" {
		input.Metadata = make(map[string]string)
		input.Metadata["ContentEncoding"] = meta.ContentEncoding
	}
	if meta.Acl != "" {
		input.ACL = obs.AclType(meta.Acl)
	}
	_, err := this.client.PutFile(input)
	return err
}

func (this *hwStorage) PutFileWithPart(
	bucketStr string,
	key string,
	srcFile string,
	metadata *Metadata,
	partSize int64,
) error {
	return errors.New("hwstorage can not support")
}

func (this *hwStorage) ListObjects(bucket string, prefix string) ([]Content, error) {
	input := &obs.ListObjectsInput{}
	a := [32]Content{}
	result := a[0:0]
	input.Bucket = bucket
	input.Prefix = prefix
	input.MaxKeys = 1000
	for {
		output, err := this.client.ListObjects(input)
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

func (this *hwStorage) DeleteObject(bucket string, key string) error {
	input := &obs.DeleteObjectInput{}
	input.Bucket = bucket
	input.Key = key
	_, err := this.client.DeleteObject(input)
	return err
}

func (this *hwStorage) BatchDeleteObject(
	bucket string,
	filelist []string,
) (successList []string, e error) {
	successList = make([]string, 0, 0)
	for len(filelist) > 0 {
		var splitSize int
		if len(filelist) > HUAWEI_MAX_DELETE_COUNT {
			splitSize = HUAWEI_MAX_DELETE_COUNT
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
		delRes, err := this.client.DeleteObjects(input)
		if err != nil {
			return successList, err
		}
		for _, deleted := range delRes.Deleteds {
			successList = append(successList, deleted.Key)
		}
	}
	return successList, nil
}

func (this *hwStorage) CopyObject(bucket string, srcKey string, destKey string) error {
	input := &obs.CopyObjectInput{}
	input.Bucket = bucket
	input.Key = destKey
	input.CopySourceBucket = bucket
	input.CopySourceKey = srcKey
	_, err := this.client.CopyObject(input)
	return err
}

func (this *hwStorage) SetObjectAcl(bucket string, key string, acl StorageAcl) error {
	input := &obs.SetObjectAclInput{}
	input.Bucket = bucket
	input.Key = key
	input.ACL = obs.AclType(string(acl))
	// 设置对象访问权限为公共读
	// input.ACL = obs.AclPublicRead
	_, err := this.client.SetObjectAcl(input)
	return err
}

func (this *hwStorage) IsObjectExist(bucket string, key string) (bool, error) {
	input := &obs.GetObjectMetadataInput{}
	input.Bucket = bucket
	input.Key = key
	output, err := this.client.GetObjectMetadata(input)

	return output != nil, err
}

/*
*
OBS会把目录自己也作为一条记录返回  所以最少2条
*/
func (s *hwStorage) IsNotEmptyDirExist(bucket string, prefix string) (bool, error) {
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
func (this *hwStorage) SetObjectMetaData(bucket string, key string, metadata *Metadata) error {
	inputMeta := &obs.SetObjectMetadataInput{}
	inputMeta.Bucket = bucket
	inputMeta.Key = key
	if metadata.Mime != "" {
		inputMeta.ContentType = metadata.Mime
	}
	if metadata.ContentEncoding != "" {
		inputMeta.ContentEncoding = metadata.ContentEncoding
	}
	if metadata.ContentDisposition != "" {
		inputMeta.ContentDisposition = metadata.ContentDisposition
	}
	_, err := this.client.SetObjectMetadata(inputMeta)
	if err != nil {
		logger.Warn("huawei-obs-sdk SetObjectMetadata key %s fail %s", key, err.Error())
		return err
	}

	if metadata.Acl != "" {
		if metadata.Acl == "default" {
			return errors.New("huawei-obs-sdk SetObjectMetaData ACL is default")
		}
		aclInput := &obs.SetObjectAclInput{}
		aclInput.Bucket = bucket
		aclInput.Key = key
		aclInput.ACL = obs.AclType(metadata.Acl)
		_, err := this.client.SetObjectAcl(aclInput)
		if err != nil {
			logger.Warn("huawei-obs-sdk SetObjectMetadata ACL key %s fail %s", key, err.Error())
			return err
		}
	}
	return nil
}

func (this *hwStorage) execCreateSignedUrl(
	bucket string,
	remoteDir string,
	method obs.HttpMethodType,
	expires int,
) (*obs.CreateSignedUrlOutput, error) {
	putObjectInput := &obs.CreateSignedUrlInput{}
	putObjectInput.Method = method
	putObjectInput.Bucket = bucket
	putObjectInput.Key = remoteDir
	putObjectInput.Expires = expires
	putObjectOutput, err := this.client.CreateSignedUrl(putObjectInput)
	return putObjectOutput, err
}

func (this *hwStorage) GetDirToken(remoteDir string) map[string]interface{} {
	res := make(map[string]interface{})
	info := this.GetDirToken2(remoteDir)
	res["accessKeyId"] = info.AccessKeyID
	res["accessKeySecret"] = info.AccessKeySecret
	res["stsToken"] = info.StsToken
	res["bucket"] = info.Bucket
	res["region"] = info.Region
	res["provider"] = info.Provider
	res["expire"] = info.Expire
	res["uploadPath"] = remoteDir
	res["host"] = this.config.Host
	return res
}

func (this *hwStorage) GetDirToken2(remoteDir string) *StorageToken {
	// 华为云的 上传 下载授权url
	// 详见: https://support.huaweicloud.com/sdk-go-devg-obs/obs_33_0601.html
	logger.Warn("huawei-obs-sdk GetDirToken2 not implement")
	return nil
	// res := &StorageToken{
	// 	AccessKeyID:     this.config.AccessKeyId,
	// 	AccessKeySecret: this.config.AccessKeySecret,
	// 	Bucket:          this.config.Bucket,
	// 	Region:          this.config.Region,
	// 	Provider:        this.config.Provider,
	// 	Expire:          int64(validTimeHuaweiObs),
	// 	UploadPath:      remoteDir,
	// 	Host:            this.config.Domain,
	// 	Path:            remoteDir,
	// }
	// putObjectOutput, err := this.execCreateSignedUrl(
	// 	this.config.Bucket,
	// 	remoteDir,
	// 	obs.HttpMethodPut,
	// 	validTimeHuaweiObs,
	// )
	// if err != nil {
	// 	logger.Warn("huawei-obs-sdk GetDirToken2 key %s fail %s", remoteDir, err.Error())
	// } else {
	// 	res.PutSignedUrl = putObjectOutput.SignedUrl
	// }

	// getObjectOutput, err := this.execCreateSignedUrl(
	// 	this.config.Bucket,
	// 	remoteDir,
	// 	obs.HttpMethodGet,
	// 	validTimeHuaweiObs,
	// )
	// if err != nil {
	// 	logger.Warn("huawei-obs-sdk GetDirToken2 key %s fail %s", remoteDir, err.Error())
	// } else {
	// 	res.GetSignedUrl = getObjectOutput.SignedUrl
	// }
	// return res
}

func (this *hwStorage) GetDirTokenWithAction(
	remoteDir string,
	actions ...Action,
) (bool, *StorageToken) {
	logger.Warn("huawei-obs-sdk GetDirTokenWithAction not implement")
	return false, nil
	// res := &StorageToken{
	// 	AccessKeyID:     this.config.AccessKeyId,
	// 	AccessKeySecret: this.config.AccessKeySecret,
	// 	Bucket:          this.config.Bucket,
	// 	Region:          this.config.Region,
	// 	Provider:        this.config.Provider,
	// 	Expire:          int64(validTimeHuaweiObs),
	// 	UploadPath:      remoteDir,
	// 	Host:            this.config.Domain,
	// 	Path:            remoteDir,
	// }
	// flag := false
	// for _, action := range actions {
	// 	switch action {
	// 	case PutObjectAction:
	// 		flag = true
	// 		putObjectOutput, err := this.execCreateSignedUrl(
	// 			this.config.Bucket,
	// 			remoteDir,
	// 			obs.HttpMethodPut,
	// 			validTimeHuaweiObs,
	// 		)
	// 		if err != nil {
	// 			logger.Warn(
	// 				"huawei-obs-sdk GetDirTokenWithAction PutObjectAction key %s fail %s",
	// 				remoteDir,
	// 				err.Error(),
	// 			)
	// 		}
	// 		res.PutSignedUrl = putObjectOutput.SignedUrl
	// 	case GetObjectAction:
	// 		flag = true
	// 		getObjectOutput, err := this.execCreateSignedUrl(
	// 			this.config.Bucket,
	// 			remoteDir,
	// 			obs.HttpMethodGet,
	// 			validTimeHuaweiObs,
	// 		)
	// 		if err != nil {
	// 			logger.Warn(
	// 				"huawei-obs-sdk GetDirTokenWithAction GetObjectAction key %s fail %s",
	// 				remoteDir,
	// 				err.Error(),
	// 			)
	// 		}
	// 		res.GetSignedUrl = getObjectOutput.SignedUrl
	// 	default:
	// 		logger.Warn(
	// 			"huawei-obs-sdk GetDirTokenWithAction unknown action %s",
	// 			action,
	// 		)
	// 	}
	// }
	// if !flag {
	// 	return false, nil
	// }
	// return flag, res
}

func (this *hwStorage) SignFile(remoteDir string, expiredTime int64) (error, string) {
	return this.SignFile2(this.config.Bucket, remoteDir, expiredTime)
}

func (this *hwStorage) SignFile2(b, remoteFilePath string, expiredTime int64) (error, string) {
	expre := int(expiredTime)
	getObjectOutput, err := this.execCreateSignedUrl(
		b,
		remoteFilePath,
		obs.HttpMethodGet,
		expre,
	)
	if err != nil {
		return err, ""
	}
	return nil, getObjectOutput.SignedUrl
}

// 过期时间:秒
func (this *hwStorage) SignFileForDownload(
	remoteFilePath string,
	expiredTime int64,
	downLoadFilename string,
) string {
	logger.Warn("hwStorage SignFileForDownload key %s fail %s", remoteFilePath, "cant support")
	//objectName := remoteFilePath
	//// 获取存储空间。
	//bucket, err := this.client.Bucket(this.config.Bucket)
	//if err != nil {
	//	return "err"
	//}
	//contentDispositionOption := oss.ContentDisposition("attachment; filename=" + downLoadFilename + "")
	//// 使用签名URL将OSS文件下载到流。
	//signedURL, err := bucket.SignURL(objectName, oss.HTTPGet, expiredTime, contentDispositionOption)
	return ""
}
func (this *hwStorage) GetObjectMeta(bucket string, key string) (*Content, error) {
	// 创建GetObjectMetadataInput实例
	input := &obs.GetObjectMetadataInput{
		Bucket: bucket,
		Key:    key,
	}
	// 调用GetObjectMetadata方法获取对象元数据
	output, err := this.client.GetObjectMetadata(input)
	if err != nil {
		logger.Warn("huawei-obs-sdk GetObjectMetadata key %s fail %s", key, err.Error())
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
func (this *hwStorage) RestoreArchive(bucket string, key string) (bool, error) {
	// 检查是否为归档类型文件
	input := &obs.GetObjectMetadataInput{
		Bucket: bucket,
		Key:    key,
	}
	output, err := this.client.GetObjectMetadata(input)
	if err != nil {
		logger.Warn("huawei-obs-sdk GetObjectMetadata key %s fail %s", key, err.Error())
		return false, err
	}

	// 检查存储类别是否为归档存储
	if output.StorageClass == obs.StorageClassCold {
		input := &obs.RestoreObjectInput{}
		// 指定存储桶名称
		input.Bucket = bucket
		// 指定归档对象名称，此处以example/objectname为例。
		input.Key = key
		// 指定待取回归档对象的对应版本号
		// input.VersionId = "G001117FCE89978B0000401205D5DC9A"
		// 指定恢复对象的保存时间，此处以1为例，单位天，取值范围：[1, 30]。
		// 必填
		input.Days = 1
		// 指定恢复选项，此处以obs.RestoreTierExpedited为例，默认为标准恢复。
		// input.Tier = obs.RestoreTierExpedited
		// 取回归档对象
		_, err := this.client.RestoreObject(input)
		if err == nil {
			return true, nil
		}
		if obsError, ok := err.(obs.ObsError); ok {
			return false, obsError
		}
		return false, err
	} else {
		// 如果不是归档存储类别，直接返回true，表示不需要解冻
		return true, nil
	}
}

func (this *hwStorage) IsArchive(bucket string, key string) (bool, error) {
	// 创建GetObjectMetadataInput实例
	input := &obs.GetObjectMetadataInput{}
	// 指定存储桶名称
	input.Bucket = bucket
	// 指定对象。
	input.Key = key
	// 调用GetObjectMetadata方法获取对象元数据
	output, err := this.client.GetObjectMetadata(input)
	if err != nil {
		logger.Warn("huawei-obs-sdk GetObjectMetadata key %s fail %s", key, err.Error())
		return false, err
	}
	// 检查存储类别是否为归档存储
	// https://support.huaweicloud.com/sdk-go-devg-obs/obs_33_0508.html#obs_33_0508__table997454612315
	return output.StorageClass == obs.StorageClassCold, nil
}
