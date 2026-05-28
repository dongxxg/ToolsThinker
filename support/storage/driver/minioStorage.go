package driver

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"
	"time"
	"tools-thinker/support"
	"tools-thinker/support/logger"
	"tools-thinker/support/storage/config"
	"tools-thinker/support/util"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// https://github.com/minio/minio-go/blob/v6.0.57/docs/API.md
// MinIO SDK API Functions

const (
	MAX_MINIO_BATCH_DELETE_COUNT = 1000
)

type minioStorage struct {
	client *minio.Client
	config *config.StorageConfig
}

func newMinioStorage(c *config.StorageConfig) (*minioStorage, error) {
	endpoint := util.If(c.Internal.GetValue(), c.EndpointInternal, c.Endpoint)
	isSecure := util.If(strings.HasPrefix(endpoint, "https"), true, false)
	endPoint := strings.ReplaceAll(endpoint, "http://", "")
	endPoint = strings.ReplaceAll(endPoint, "https://", "")
	client, err := minio.New(
		endPoint, &minio.Options{
			Creds:  credentials.NewStaticV4(c.AccessKeyId, c.AccessKeySecret, ""),
			Secure: isSecure,
		},
	)
	if err != nil {
		return nil, err
	}
	return &minioStorage{config: c, client: client}, nil
}
func (s *minioStorage) IsPublicAcl(bucket string, key string) (bool, error) {
	panic("implement me")
}
func (m *minioStorage) GetObject(bucket string, key string) ([]byte, error) {
	obj, err := m.client.GetObject(context.TODO(), bucket, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	defer obj.Close()
	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, obj)
	return buf.Bytes(), err
}

func (m *minioStorage) GetFile(bucket string, key string, localFile string) error {
	return m.client.FGetObject(
		context.TODO(),
		bucket,
		key,
		localFile,
		minio.GetObjectOptions{},
	)
}

func (m *minioStorage) PutObject(
	bucket string,
	key string,
	data []byte,
	metadata map[string]string,
) error {
	_, err := m.client.PutObject(
		context.TODO(),
		bucket,
		key,
		bytes.NewBuffer(data),
		int64(len(data)),
		mapToPutObjOptions(metadata),
	)
	return err
}

func (m *minioStorage) PutObjectWithMeta(
	bucket string,
	key string,
	data []byte,
	metadata *Metadata,
) error {
	_, err := m.client.PutObject(
		context.TODO(),
		bucket,
		key,
		bytes.NewBuffer(data),
		int64(len(data)),
		metadataToPutObjOptions(metadata),
	)
	return err
}

func (m *minioStorage) PutFile(
	bucket string,
	key string,
	localFile string,
	metadata map[string]string,
) error {
	_, err := m.client.FPutObject(
		context.TODO(),
		bucket,
		key,
		localFile,
		mapToPutObjOptions(metadata),
	)
	return err
}

func (m *minioStorage) PutFileWithMeta(
	bucket string,
	key string,
	srcFile string,
	metadata *Metadata,
) error {
	_, err := m.client.FPutObject(
		context.TODO(),
		bucket,
		key,
		srcFile,
		metadataToPutObjOptions(metadata),
	)
	return err
}

func (m *minioStorage) PutObjectFromFile(
	bucket string,
	key string,
	localFile string,
	metadata map[string]string,
) error {
	// 使用FPutObject上传文件
	_, err := m.client.FPutObject(
		context.TODO(), bucket, key, localFile, mapToPutObjOptions(metadata))
	if err != nil {
		return fmt.Errorf("failed to put object from file: %w", err)
	}
	return nil
}

// 在单个 PUT 操作中上传小于 128MiB 的对象。对于大于 128MiB 的对象，PutObject 会根据实际文件大小将对象无缝上传为 128MiB 或更大的部分。对象的最大上传大小为 5TB
func (m *minioStorage) PutFileWithPart(
	bucket string,
	key string,
	srcFile string,
	metadata *Metadata,
	partSize int64,
) error {
	file, err := os.Open(srcFile)
	if err != nil {
		return fmt.Errorf("PutFileWithPart os.Open error %w", err)
	}
	defer file.Close()

	fileStat, err := file.Stat()
	if err != nil {
		return fmt.Errorf("PutFileWithPart file.Stat error %w", err)
	}

	putOpt := metadataToPutObjOptions(metadata)
	// 这里是直接覆盖
	putOpt.ContentType = "application/octet-stream"
	_, err = m.client.PutObject(
		context.TODO(),
		bucket,
		key,
		file,
		fileStat.Size(),
		putOpt,
	)
	if err != nil {
		return fmt.Errorf("PutFileWithPart m.client.PutObject error %w", err)
	}
	return nil
}

func (m *minioStorage) ListObjects(bucket string, prefix string) ([]Content, error) {
	res := make([]Content, 0, 32)
	doneCh := make(chan struct{})
	defer close(doneCh)
	// List all objects from a bucket-name with a matching prefix.
	for object := range m.client.ListObjects(context.TODO(), bucket, minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: true,
		MaxKeys:   100, // 每次列出 100 个对象
	}) {
		if object.Err != nil {
			return res, object.Err
		}
		// protect
		if len(res) >= 100000 {
			return res, errors.New("too much data, break")
		}
		res = append(res, *objectInfoToContent(&object))
	}
	return res, nil
}

func (m *minioStorage) DeleteObject(bucket string, key string) error {
	return m.client.RemoveObject(
		context.TODO(), bucket, key, minio.RemoveObjectOptions{})
}

func (m *minioStorage) BatchDeleteObject(
	bucket string,
	filelist []string,
) (successList []string, err error) {
	successList = make([]string, 0, len(filelist))

	for len(filelist) > 0 {
		var splitPos int
		if len(filelist) > MAX_MINIO_BATCH_DELETE_COUNT {
			splitPos = MAX_MINIO_BATCH_DELETE_COUNT
		} else {
			splitPos = len(filelist)
		}
		toDelFileList := filelist[:splitPos]
		filelist = filelist[splitPos:]

		deleteList := make([]minio.ObjectInfo, 0, len(toDelFileList))
		for _, key := range toDelFileList {
			deleteList = append(deleteList, minio.ObjectInfo{Key: key})
		}

		objectsCh := make(chan minio.ObjectInfo)

		// 发送对象以供删除
		go func() {
			defer close(objectsCh)
			for _, object := range deleteList {
				objectsCh <- object
			}
		}()

		// 执行批量删除
		fmt.Println("Deleting objects...")
		for err := range m.client.RemoveObjects(
			context.TODO(), bucket, objectsCh, minio.RemoveObjectsOptions{}) {
			if err.Err != nil {
				fmt.Printf("Error detected during deletion: %v\n", err)
			}
		}
		fmt.Println("Deletion completed for this batch")
		successList = append(successList, toDelFileList...)
	}

	return successList, nil
}

func (m *minioStorage) CopyObject(bucket string, srcKey string, destKey string) error {
	// Source object
	srcOpts := minio.CopySrcOptions{
		Bucket: bucket,
		Object: srcKey,
	}

	// Destination object
	dstOpts := minio.CopyDestOptions{
		Bucket: bucket,
		Object: destKey,
	}
	// Copy object call
	if _, err := m.client.CopyObject(context.TODO(), dstOpts, srcOpts); err != nil {
		return err
	}
	return nil
}

func (m *minioStorage) SetObjectAcl(bucket string, key string, acl StorageAcl) error {
	// return errors.New("minio not support SetObjectAcl")
	return nil
}

func (m *minioStorage) SetObjectMetaData(bucket string, key string, metadata *Metadata) error {
	// return errors.New("minio not support SetObjectMeta")
	return nil
}

func (m *minioStorage) IsObjectExist(bucket string, key string) (bool, error) {
	res, err := m.GetObjectMeta(bucket, key)
	switch err := err.(type) {
	case minio.ErrorResponse:
		if err.Code == "NoSuchKey" {
			return false, nil
		} else {
			return false, err
		}
	default:
		return res != nil, err
	}

}

/*
*
OBS会把目录自己也作为一条记录返回  所以最少2条
*/
func (m *minioStorage) IsNotEmptyDirExist(bucket string, prefix string) (bool, error) {
	res := make([]Content, 0, 32)
	doneCh := make(chan struct{})
	defer close(doneCh)
	// List all objects from a bucket-name with a matching prefix.
	for object := range m.client.ListObjects(context.TODO(), bucket, minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: false,
		MaxKeys:   2, // 每次列出 100 个对象
	}) {
		if object.Err != nil {
			return false, object.Err
		}
		// protect
		if len(res) >= 100000 {
			return false, errors.New("too much data, break")
		}
		res = append(res, *objectInfoToContent(&object))
	}

	if len(res) == 2 {
		return true, nil
	}
	return false, nil

}

func (m *minioStorage) SignFile(remoteDir string, expiredTime int64) (error, string) {
	return m.SignFile2(m.config.Bucket, remoteDir, expiredTime)
}

func (m *minioStorage) SignFile2(bucket, remoteDir string, expiredTime int64) (error, string) {
	// Set request parameters for content-disposition.
	reqParams := make(url.Values)
	// Generates a presigned url which expires in a day.
	presignedURL, err := m.client.PresignedGetObject(
		context.TODO(),
		bucket,
		remoteDir,
		time.Duration(expiredTime)*time.Second,
		reqParams,
	)
	if err != nil {
		return err, ""
	}
	return nil, presignedURL.String()
}

// 过期时间:秒
func (m *minioStorage) SignFileForDownload(
	remoteFilePath string,
	expiredTime int64,
	downLoadFilename string,
) string {
	// Set request parameters for content-disposition.
	reqParams := make(url.Values)
	reqParams.Set(
		"response-content-disposition",
		"attachment; filename=\""+url.PathEscape(downLoadFilename)+"\"",
	)

	// Generates a presigned url which expires in a day.
	presignedURL, err := m.client.PresignedGetObject(
		context.TODO(),
		m.config.Bucket,
		remoteFilePath,
		time.Duration(expiredTime)*time.Second,
		reqParams,
	)
	if err != nil {
		return ""
	}
	return presignedURL.String()
}

func (m *minioStorage) GetObjectMeta(bucket string, key string) (*Content, error) {
	objInfo, err := m.client.StatObject(context.TODO(), bucket, key, minio.StatObjectOptions{})
	if err != nil {
		return nil, err
	}
	return objectInfoToContent(&objInfo), nil
}

// 获取临时token
// https://github.com/minio/minio/blob/master/docs/sts/assume-role.md
// 使用了minio的账号密码实现,相当于最大权限
func (m *minioStorage) GetDirToken2(remoteDir string) *StorageToken {
	var expires int = 6 * 3600
	if strings.HasSuffix(remoteDir, "/") {
		remoteDir = remoteDir[:len(remoteDir)-1]
	}

	policy := fmt.Sprintf(`
	{
		"Version": "2012-10-17",
		"Statement": [
			{
				"Effect": "Allow",
				"Action": ["s3:GetObject", "s3:PutObject"],
				"Resource": ["arn:aws:s3:::%s/%s/*"]
			}
		]
	}`, m.config.Bucket, remoteDir)

	role, err := credentials.NewSTSAssumeRole(m.config.Endpoint, credentials.STSAssumeRoleOptions{
		AccessKey:       m.config.AccessKeyId,
		SecretKey:       m.config.AccessKeySecret,
		SessionToken:    "",
		Policy:          policy,
		Location:        "",
		DurationSeconds: expires,
		RoleARN:         "",
		RoleSessionName: "",
		ExternalID:      "",
	})

	if err != nil {
		logger.Error("Error creating MinIO client: %v", err)
		return nil
	}
	tempRole, err := role.GetWithContext(nil)
	if err != nil {
		logger.Error("Error creating MinIO client: %v", err)
		return nil
	}

	return &StorageToken{
		AccessKeyID:     tempRole.AccessKeyID,
		AccessKeySecret: tempRole.SecretAccessKey,
		StsToken:        tempRole.SessionToken,
		Bucket:          m.config.Bucket,
		Region:          m.config.Region,
		Provider:        m.config.Provider,
		Expire:          support.NowMs() + int64(expires)*1000,
		UploadPath:      remoteDir,
		Host:            m.config.Host,
		EndPoint:        m.config.Endpoint,
		Path:            remoteDir,
		CdnDomain:       m.config.CdnDomain,
	}
}

func (m *minioStorage) GetDirToken(remoteDir string) map[string]any {
	t := m.GetDirToken2(remoteDir)
	if t == nil {
		return nil
	}
	res := map[string]any{
		"accessKeyId":     t.AccessKeyID,
		"accessKeySecret": t.AccessKeySecret,
		"stsToken":        t.StsToken,
		"bucket":          t.Bucket,
		"region":          t.Region,
		"provider":        t.Provider,
		"expire":          t.Expire,
		"uploadPath":      t.UploadPath,
		"host":            t.Host,
		"endPoint":        t.EndPoint,
	}
	return res

}

func (m *minioStorage) GetDirTokenWithAction(
	remoteDir string,
	actions ...Action,
) (bool, *StorageToken) {
	t := m.GetDirToken2(remoteDir)
	if t == nil {
		return false, nil
	}
	return true, t
}

// minio not support
func (m *minioStorage) RestoreArchive(bucket string, key string) (bool, error) {
	logger.Warn("minio RestoreArchive method is not implemented")
	return false, errors.New("cant support RestoreArchive")
}

// minio not support
func (m *minioStorage) IsArchive(bucket string, key string) (bool, error) {
	logger.Warn("minio IsArchive method is not implemented")
	return false, errors.New("cant support IsArchive")
}

func mapToPutObjOptions(metadata map[string]string) minio.PutObjectOptions {
	ops := minio.PutObjectOptions{}
	if len(metadata) == 0 {
		return ops
	}
	for key, value := range metadata {
		if len(value) == 0 {
			continue
		}
		switch key {
		case "Content-Encoding":
			ops.ContentEncoding = value
		case "Content-Disposition":
			ops.ContentDisposition = value
		case "mime":
			ops.ContentType = value
		default:
			// do nothing
			logger.Warn("minio cant support metadata %s %s", key, value)
		}
	}
	return ops
}

func metadataToPutObjOptions(metadata *Metadata) minio.PutObjectOptions {
	ops := minio.PutObjectOptions{}
	if metadata != nil {
		if len(metadata.Mime) != 0 {
			ops.ContentType = metadata.Mime
		}
		if len(metadata.ContentEncoding) != 0 {
			ops.ContentEncoding = metadata.ContentEncoding
		}
		if len(metadata.ContentDisposition) != 0 {
			ops.ContentDisposition = metadata.ContentDisposition
		}
	}
	return ops
}

func objectInfoToContent(obj *minio.ObjectInfo) *Content {
	res := &Content{
		Key:          obj.Key,
		Size:         obj.Size,
		ETag:         obj.ETag,
		LastModified: obj.LastModified,
	}
	return res
}
