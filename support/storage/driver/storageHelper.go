package driver

import (
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"tools-thinker/support/concurrent"
	"tools-thinker/support/file"
	"tools-thinker/support/logger"
	"tools-thinker/support/storage/config"
)

var GO_LIMIT_COUNT = 20

type StorageHelper struct {
	st            Storage
	storageConfig *config.StorageConfig
}

// ReplaceEndPointInternal 将内网地址转换为外网地址
func (_this *StorageHelper) ReplaceEndPointInternal(in string) string {
	if _this.storageConfig.Provider == "ali" {
		return strings.ReplaceAll(
			in,
			"-internal.aliyuncs.com",
			".aliyuncs.com")
	}
	return in
}

// getStorageConfig
func (_this *StorageHelper) GetStorageConfig() *config.StorageConfig {
	return _this.storageConfig
}

func (_this *StorageHelper) GetCdnUrl() string {
	return _this.storageConfig.CdnProtocol + _this.storageConfig.CdnDomain
}

func (_this *StorageHelper) GetCdnDomain() string {
	return _this.storageConfig.CdnDomain
}

func (_this *StorageHelper) GetRootDir() string {
	return _this.storageConfig.Root
}

func (_this *StorageHelper) GetBucketName() string {
	return _this.storageConfig.Bucket
}

func (_this *StorageHelper) GetInternal() bool {
	return _this.storageConfig.Internal.GetValue()
}

func (_this *StorageHelper) GetRegion() string {
	return _this.storageConfig.Region
}

func (_this *StorageHelper) GetHost() string {
	return _this.storageConfig.Host
}

func (_this *StorageHelper) GetCustomHost() string {
	host := _this.storageConfig.CustomHost
	if len(host) == 0 {
		host = _this.storageConfig.Host
	}
	return host
}

func (_this *StorageHelper) GetTmpRoot() string {
	return _this.storageConfig.TmpRoot
}

func (_this *StorageHelper) GetEndPoint() string {
	return strings.TrimPrefix(_this.storageConfig.Endpoint, "https://")
}

func (_this *StorageHelper) GetEndpointInternal() string {
	return strings.TrimPrefix(_this.storageConfig.EndpointInternal, "https://")
}

func (_this *StorageHelper) GetTmpPath() []string {
	return _this.storageConfig.TmpPath
}

func (_this *StorageHelper) GetPath() []string {
	return _this.storageConfig.Path
}

func checkObjectKey(objectFullName string) {
	//如果objectFullName包含双斜杠，可能会导致文件无法下载和上传且不报错，所以这里预先检查打印日志
	if strings.Contains(objectFullName, "//") {
		logger.Info(
			"please check！objectFullName {%s} contain // , it  Cause things to fail",
			objectFullName,
		)
	}
}

func (_this *StorageHelper) GetObject(key string) ([]byte, error) {
	return _this.st.GetObject(_this.storageConfig.Bucket, key)
}
func (_this *StorageHelper) GetFile(key string, localFile string) error {
	return _this.st.GetFile(_this.storageConfig.Bucket, key, localFile)
}

func (_this *StorageHelper) PutObject(key string, data []byte, metadata map[string]string) error {
	return _this.st.PutObject(_this.storageConfig.Bucket, key, data, metadata)
}

func (_this *StorageHelper) PutObjectWithMeta(key string, data []byte, metadata *Metadata) error {
	return _this.st.PutObjectWithMeta(_this.storageConfig.Bucket, key, data, metadata)
}

func (_this *StorageHelper) PutFile(key string, srcFile string, metadata map[string]string) error {
	return _this.st.PutFile(_this.storageConfig.Bucket, key, srcFile, metadata)
}

func (_this *StorageHelper) PutFileFromFile(
	key string,
	srcFile string,
	metadata map[string]string,
) error {
	return _this.st.PutObjectFromFile(_this.storageConfig.Bucket, key, srcFile, metadata)
}
func (_this *StorageHelper) PutFileWithMeta(key string, srcFile string, metadata *Metadata) error {
	return _this.st.PutFileWithMeta(_this.storageConfig.Bucket, key, srcFile, metadata)
}

func (_this *StorageHelper) ListObjects(prefix string) ([]Content, error) {
	resultList, err := _this.st.ListObjects(_this.storageConfig.Bucket, prefix)
	return resultList, err
}

func (_this *StorageHelper) DeleteObject(key string) error {
	return _this.st.DeleteObject(_this.storageConfig.Bucket, key)
}

func (_this *StorageHelper) BatchDeleteObject(fileList []string) (successList []string, err error) {
	return _this.st.BatchDeleteObject(_this.storageConfig.Bucket, fileList)
}

func (_this *StorageHelper) CopyObject(srcKey string, destKey string) error {
	return _this.st.CopyObject(_this.storageConfig.Bucket, srcKey, destKey)
}

func (_this *StorageHelper) SetObjectAcl(key string, acl StorageAcl) error {
	return _this.st.SetObjectAcl(_this.storageConfig.Bucket, key, acl)
}

// 下载oss/obs中指定目录。 err!= nil时，返回失败的列表(nil为未获取到列表)。成功时列表为空
func (_this *StorageHelper) GetFolder(remoteDir string, localFolder string) ([]Content, error) {
	match := func(key string) bool {
		return true
	}

	return _this.GetFolderFilter(remoteDir, localFolder, match)
}

// 下载oss/obs中指定目录, pattern为使用key路径匹配的规则。 err!= nil时，返回失败的列表(nil为未获取到列表)。成功时列表为空
// 非匹配参考：https://www.cnblogs.com/asfeixue/p/lookahead.html
// 但go 不支持正则回溯如(?!，因此非匹配需要通过GetFolderFilter方法的match方法来实现。
func (_this *StorageHelper) GetFolderRegex(
	remoteDir string,
	localFolder string,
	pattern string,
) ([]Content, error) {
	match := func(key string) bool {
		ok, _ := regexp.MatchString(pattern, key)
		return ok
	}
	return _this.GetFolderFilter(remoteDir, localFolder, match)
}

func (_this *StorageHelper) GetFolderFilter(
	remoteDir string,
	localFolder string,
	match func(key string) bool,
) ([]Content, error) {
	resultList, err := _this.ListObjects(remoteDir)

	if err != nil {
		return nil, err
	}

	var failList []Content = make([]Content, 0)
	var result error = nil
	goLimit := concurrent.NewGoLimit(GO_LIMIT_COUNT)
	for _, contentTmp := range resultList {
		content := contentTmp
		if !match(content.Key) {
			continue
		}
		subfix := strings.Replace(content.Key, remoteDir, "", 1)

		destFile := file.JoinPath(localFolder, subfix)
		index := strings.LastIndex(destFile, "/")
		folder := destFile[0 : index+1]
		//对于对象存储，/的文件目录是可选的，因此先每次都创建(每次检测速度较慢，可以优化)
		os.MkdirAll(folder, 0755)
		if !strings.HasSuffix(content.Key, "/") {
			goLimit.Run(func() {
				err = _this.GetFile(content.Key, file.JoinPath(localFolder, subfix))
				if err != nil {
					result = err
					failList = append(failList, content)
					logger.Warn(
						"get storage file from %s to %s error: %s",
						content.Key,
						localFolder+subfix,
						err,
					)
				}
			})

		}
	}
	goLimit.Wait()
	return failList, result
}

// 上传文件夹 ，err != nil时，[]string为上传失败的列表
func (_this *StorageHelper) PutFolder(
	remoteDir string,
	localFolder string,
	optionMethod func(string) map[string]string,
) ([]string, error) {
	// prefix := localFolder
	localFolder = strings.ReplaceAll(localFolder, "\\", "/")
	var failList []string = make([]string, 0)
	goLimit := concurrent.NewGoLimit(GO_LIMIT_COUNT)
	err := filepath.Walk(localFolder, func(path string, info os.FileInfo, err error) error {
		if info == nil {
			logger.Warn("file %s is nil", path)
			return nil
		}
		if !info.IsDir() {
			goLimit.Run(func() {
				path = strings.ReplaceAll(path, "\\", "/")
				subfix := strings.Replace(path, localFolder, "", 1)
				if optionMethod != nil {
					err = _this.PutFile(file.JoinPath(remoteDir, subfix), path, optionMethod(path))
				} else {
					err = _this.PutFileWithMeta(file.JoinPath(remoteDir, subfix), path, &Metadata{})
				}
				if err != nil {
					logger.Warn("upload file %s failed,as err %v", path, err)
					failList = append(failList, path)
				}
			})
		}
		return err
	})
	goLimit.Wait()
	return failList, err
}

func (_this *StorageHelper) PutFolderToOss(
	remoteDir string,
	localFolder string,
) (failList []string, err error) {
	return _this.PutFolder(remoteDir, localFolder, nil)
}

// @return fList 上传失败的文件列表时
// @return sList 上传成功的文件列表
// @return err 报错信息
func (_this *StorageHelper) PutFolder2(
	remoteDir string,
	localFolder string,
	optionMethod func(string) map[string]string,
) (fList []string, sList []string, err error) {
	// prefix := localFolder
	localFolder = strings.ReplaceAll(localFolder, "\\", "/")
	failList := concurrent.SafeStringSlice{}
	sucList := concurrent.SafeStringSlice{}
	goLimit := concurrent.NewGoLimit(GO_LIMIT_COUNT)
	err = filepath.Walk(localFolder, func(path string, info os.FileInfo, err error) error {
		if info == nil {
			logger.Warn("file %s is nil", path)
			return nil
		}
		if !info.IsDir() {
			goLimit.Run(func() {
				path = strings.ReplaceAll(path, "\\", "/")
				subfix := strings.Replace(path, localFolder, "", 1)
				var pErr error
				if optionMethod != nil {
					pErr = _this.PutFile(file.JoinPath(remoteDir, subfix), path, optionMethod(path))
				} else {
					pErr = _this.PutFile(file.JoinPath(remoteDir, subfix), path, nil)
				}
				if pErr != nil {
					failList.Append(path)
					logger.Warn(
						"upload %s to %s failed err = %v",
						file.JoinPath(remoteDir, subfix),
						path,
						pErr,
					)
				} else {
					sucList.Append(path)
				}
			})
		}
		return err
	})
	goLimit.Wait()
	fList = failList.GetSlice()
	sList = sucList.GetSlice()
	return fList, sList, err
}

func (_this *StorageHelper) GetFolderSize(remoteDir string) (int64, error) {
	contents, err := _this.ListObjects(remoteDir)
	if err != nil {
		return 0, err
	}
	var sum int64 = 0
	for _, content := range contents {
		sum += content.Size
	}
	return sum, nil
}

func (_this *StorageHelper) DeleteFolder(remoteDir string) error {
	contents, err := _this.ListObjects(remoteDir)
	if err != nil {
		return err
	}
	goLimit := concurrent.NewGoLimit(GO_LIMIT_COUNT)
	for _, content := range contents {
		contentTmp := content
		goLimit.Run(func() {
			_this.DeleteObject(contentTmp.Key)
		})
	}
	goLimit.Wait()
	return nil
}

func (_this *StorageHelper) CopyFolder(remoteDir string, remoteDistDir string) error {
	contents, err := _this.ListObjects(remoteDir)

	if err != nil {
		return err
	}
	goLimit := concurrent.NewGoLimit(GO_LIMIT_COUNT)
	for _, content := range contents {
		contentTmp := content
		if !strings.HasSuffix(contentTmp.Key, "/") {
			var f = func() {
				subfix := strings.Replace(contentTmp.Key, remoteDir, "", 1)
				_this.CopyObject(contentTmp.Key, file.JoinPath(remoteDistDir, subfix))
			}
			goLimit.Run(f)
		}
	}
	goLimit.Wait()
	return nil
}

func (_this *StorageHelper) CopyFolderWithAcl(
	remoteDir string,
	remoteDistDir string,
	acl StorageAcl,
) error {
	contents, err := _this.ListObjects(remoteDir)

	if err != nil {
		return err
	}
	goLimit := concurrent.NewGoLimit(GO_LIMIT_COUNT)
	for _, content := range contents {
		contentTmp := content
		if !strings.HasSuffix(contentTmp.Key, "/") {
			var f = func() {
				subfix := strings.Replace(contentTmp.Key, remoteDir, "", 1)
				desKey := file.JoinPath(remoteDistDir, subfix)
				tmpError := _this.CopyObject(contentTmp.Key, desKey)
				if tmpError != nil {
					logger.Warn(
						"CopyObject from %s to %s failed!,err %v",
						contentTmp.Key,
						desKey,
						tmpError,
					)
					return
				}
				tmpError2 := _this.SetObjectAcl(desKey, acl)
				if tmpError2 != nil {
					logger.Warn("set file public read %s failed!,err %v", desKey, tmpError2)
					return
				}
			}
			goLimit.Run(f)
		}
	}
	goLimit.Wait()
	return nil
}

func (_this *StorageHelper) SetFolderAcl(remoteDir string, acl StorageAcl) error {
	contents, err := _this.ListObjects(remoteDir)

	if err != nil {
		return err
	}
	goLimit := concurrent.NewGoLimit(GO_LIMIT_COUNT)
	for _, content := range contents {
		contentTmp := content
		goLimit.Run(func() {
			_this.SetObjectAcl(contentTmp.Key, acl)
		})
	}

	goLimit.Wait()
	return nil
}

func (_this *StorageHelper) Move(srcKey string, destKey string) error {
	err := _this.CopyObject(srcKey, destKey)
	if err != nil {
		return err
	}
	err = _this.DeleteObject(srcKey)
	return err
}

func (_this *StorageHelper) MoveWithAcl(srcKey string, destKey string, acl StorageAcl) error {
	err := _this.CopyObject(srcKey, destKey)
	if err != nil {
		return err
	}
	if acl != "" {
		_this.SetObjectAcl(destKey, acl)
	}
	_this.DeleteObject(srcKey)
	return nil
}

func (_this *StorageHelper) MoveFolder(remoteDir string, remoteDistDir string) error {
	contents, err := _this.ListObjects(remoteDir)
	if err != nil {
		return err
	}
	var moveError error
	goLimit := concurrent.NewGoLimit(GO_LIMIT_COUNT)
	for _, content := range contents {
		if !strings.HasSuffix(content.Key, "/") {
			contentTmp := content
			goLimit.Run(func() {
				subfix := strings.Replace(contentTmp.Key, remoteDir, "", 1)
				tmpError := _this.CopyObject(contentTmp.Key, file.JoinPath(remoteDistDir, subfix))
				tmpError = _this.DeleteObject(contentTmp.Key)
				if tmpError != nil {
					logger.Warn(
						"move file from %s to %s failed!,err %v",
						contentTmp.Key,
						file.JoinPath(remoteDistDir, subfix),
						tmpError,
					)
					moveError = tmpError
				}
			})
		}
	}
	goLimit.Wait()
	return moveError
}

func (_this *StorageHelper) MoveFolderWithAcl(
	remoteDir string,
	remoteDistDir string,
	acl StorageAcl,
) error {
	contents, err := _this.ListObjects(remoteDir)
	if err != nil {
		return err
	}
	goLimit := concurrent.NewGoLimit(GO_LIMIT_COUNT)
	for _, content := range contents {
		if !strings.HasSuffix(content.Key, "/") {
			contentTmp := content
			goLimit.Run(func() {
				subfix := strings.Replace(contentTmp.Key, remoteDir, "", 1)
				desKey := file.JoinPath(remoteDistDir, subfix)
				tmpError1 := _this.CopyObject(contentTmp.Key, desKey)
				if tmpError1 != nil {
					logger.Warn(
						"copy file from %s to %s failed!,err %v",
						contentTmp.Key,
						file.JoinPath(remoteDistDir, subfix),
						tmpError1,
					)
					return
				}
				tmpError2 := _this.SetObjectAcl(desKey, acl)
				if tmpError2 != nil {
					logger.Warn("set file public read %s failed!,err %v", desKey, tmpError2)
				}
				tmpError3 := _this.DeleteObject(contentTmp.Key)
				if tmpError3 != nil {
					logger.Warn("delete file  %s failed!,err %v", contentTmp.Key, tmpError3)
				}
			})
		}
	}
	goLimit.Wait()
	return nil
}

const ACL_REQ_HEADER = "acl"

func GetOption(fileName string) map[string]string {
	metadata := make(map[string]string)
	if strings.HasSuffix(fileName, "info.plist") {
		metadata["mime"] = "application/json"
		metadata["Content-Encoding"] = "gzip"
		metadata[ACL_REQ_HEADER] = string(AclPrivate)
		return metadata
	}
	var noPublicFiles = "pptx?$"
	match, _ := regexp.MatchString(noPublicFiles, fileName)
	if match {
		metadata[ACL_REQ_HEADER] = string(AclPrivate)
	} else {
		metadata[ACL_REQ_HEADER] = string(AclPublicRead)
	}
	return metadata
}

func GetInfoPlistOption() map[string]string {
	metadata := make(map[string]string)
	metadata["mime"] = "application/json"
	metadata["Content-Encoding"] = "gzip"
	metadata["acl"] = string(AclPrivate)
	return metadata
}

func GetReadOption(fileName string) map[string]string {
	metadata := make(map[string]string)
	if strings.HasSuffix(fileName, "info.plist") {
		metadata["mime"] = "application/json"
		metadata["Content-Encoding"] = "gzip"
		metadata["acl"] = string(AclPublicRead)
		return metadata
	}
	metadata["acl"] = string(AclPublicRead)
	return metadata
}

func (_this *StorageHelper) IsObjectExist(key string) (bool, error) {
	return _this.st.IsObjectExist(_this.storageConfig.Bucket, key)
}

func (_this *StorageHelper) IsNotEmptyDirExist(key string) (bool, error) {
	return _this.st.IsNotEmptyDirExist(_this.storageConfig.Bucket, key)
}

func (_this *StorageHelper) GetDirToken(remoteDir string) map[string]interface{} {
	return _this.st.GetDirToken(remoteDir)
}

func (_this *StorageHelper) GetDirToken2(remoteDir string) *StorageToken {
	token := _this.st.GetDirToken2(remoteDir)
	endpointUrl, err := url.Parse(token.EndPoint)
	if err != nil {
		logger.Error("parse endpoint url failed, %v", err)
		return token
	}
	port := endpointUrl.Port()
	schema := endpointUrl.Scheme
	token.EndPoint = endpointUrl.Hostname()
	if port == "" {
		if schema == "https" {
			token.Port = 443
		} else {
			token.Port = 80
		}
	} else {
		portInt, err := strconv.ParseInt(port, 10, 64)
		if err == nil {
			token.Port = portInt
		} else {
			logger.Error("parse port failed, %v", err)
		}
	}
	return token
}

func (_this *StorageHelper) GetDirTokenWithAction(
	remoteDir string,
	actions Action,
) (bool, *StorageToken) {
	return _this.st.GetDirTokenWithAction(remoteDir, actions)
}

func (_this *StorageHelper) GetRemoteDirPath(remoteDir string) string {
	processedPath := remoteDir
	// 检查 processedPath 是否以 "/" 开头，如果没有则添加
	if !strings.HasPrefix(processedPath, "/") {
		processedPath = "/" + processedPath
	}
	// minio的接入域名不是直接指向Bucket的
	//if _this.storageConfig.Provider == ProviderMinio {
	//	processedPath = "/" + _this.GetBucketName() + processedPath
	//}
	return processedPath
}

// expiredTime 秒
func (_this *StorageHelper) SignFile(remoteDir string, expiredTime int64) string {
	err, signUrl := _this.st.SignFile(remoteDir, expiredTime)
	if err != nil {
		logger.Error(
			"driver:%s SignFile key %s fail %s",
			_this.storageConfig.Provider,
			remoteDir,
			err.Error(),
		)
	}
	return signUrl
}

func (_this *StorageHelper) RemoteDirPathWithRoot(remoteDir string) string {
	rootDir := _this.GetRootDir()

	// 检查 remoteDir 是否已经以 rootDir 为前缀
	if !strings.HasPrefix(remoteDir, rootDir) {
		// 检查 remoteDir 是否以 "/" 开头，如果不是，则加上 "/"
		if !strings.HasPrefix(remoteDir, "/") {
			remoteDir = "/" + remoteDir
		}
		return rootDir + remoteDir
	}
	return remoteDir
}

// ViewUrlWithRemoteDir 返回完整的 URL
func (_this *StorageHelper) ViewUrlWithRemoteDir(remoteDir string) string {
	processedPath := _this.RemoteDirPathWithRoot(remoteDir)

	// 检查 processedPath 是否以 "/" 开头，如果没有则添加
	if !strings.HasPrefix(processedPath, "/") {
		processedPath = "/" + processedPath
	}

	return "https://" + _this.storageConfig.Host + processedPath
}

// 生成自定义的域名前缀 （目前有特殊用法，推荐使用ViewUrlWithRemoteDir函数）
func (_this *StorageHelper) ViewFileCustomUrl(remoteDir string) string {
	processedPath := _this.RemoteDirPathWithRoot(remoteDir)

	processedPath = _this.GetRemoteDirPath(processedPath)

	return "https://" + _this.GetCustomHost() + processedPath
}

func (_this *StorageHelper) ViewCdnUrlWithRemoteDir(remoteDir string) string {
	processedPath := _this.RemoteDirPathWithRoot(remoteDir)

	processedPath = _this.GetRemoteDirPath(processedPath)

	return _this.GetCdnUrl() + processedPath
}

// expiredTime 秒
func (_this *StorageHelper) SignFile2(bucket string, remoteDir string, expiredTime int64) string {
	err, signUrl := _this.st.SignFile2(bucket, remoteDir, expiredTime)
	if err != nil {
		logger.Error(
			"driver:%s SignFile key %s fail %s",
			_this.storageConfig.Provider,
			remoteDir,
			err.Error(),
		)
	}
	return signUrl
}

// expiredTime 秒
func (_this *StorageHelper) SignFileForDownload(
	remoteDir string,
	expiredTime int64,
	downloadName string,
) string {
	return _this.st.SignFileForDownload(remoteDir, expiredTime, downloadName)
}

func (_this *StorageHelper) SignFileWithCdn(
	bucket string,
	remoteDir string,
	expiredTime int64,
	ssl bool,
) string {

	err, signFileUrl := _this.st.SignFile2(bucket, remoteDir, expiredTime)
	if err != nil {
		logger.Error(
			"driver:%s SignFile key %s fail %s",
			_this.storageConfig.Provider,
			remoteDir,
			err.Error(),
		)
	}

	signUrl := strings.ReplaceAll(
		signFileUrl,
		_this.GetBucketName()+"."+_this.GetEndPoint(),
		_this.GetCdnDomain(),
	)
	if ssl {
		signUrl = strings.ReplaceAll(signUrl, "http://", "https://")
	}
	return signUrl
}

func (_this *StorageHelper) GetObjectMeta(key string) (*Content, error) {
	return _this.st.GetObjectMeta(_this.storageConfig.Bucket, key)
}

// 解冻，触发成功为true，否则为false
func (_this *StorageHelper) RestoreArchive(key string) (bool, error) {
	return _this.st.RestoreArchive(_this.storageConfig.Bucket, key)
}

// 是否为归档文件
func (_this *StorageHelper) IsArchive(key string) (bool, error) {
	return _this.st.IsArchive(_this.storageConfig.Bucket, key)
}

func (_this *StorageHelper) PutFileWithPart(
	key string,
	localFile string,
	meta *Metadata,
	partSize int64,
) error {
	if partSize == 0 || partSize < 100*1024 {
		partSize = 10 * 1024 * 1024
	}
	return _this.st.PutFileWithPart(
		_this.storageConfig.Bucket,
		key,
		localFile,
		meta,
		partSize,
	)
}

func (_this *StorageHelper) SetObjectMetaData(key string, metadata *Metadata) error {
	return _this.st.SetObjectMetaData(_this.storageConfig.Bucket, key, metadata)
}

func (_this *StorageHelper) FindM3u8(ossPath string) []string {
	contents, err := _this.ListObjects(ossPath)
	if err != nil {
		logger.Warn("list %s for %s err, %v", ossPath)
		return nil
	}
	startIndex := len(ossPath)
	var res []string
	for _, content := range contents {
		//fmt.Println(content.Key)
		name := content.Key
		subName := name[startIndex:]
		basename := filepath.Base(subName)
		if strings.HasSuffix(basename, ".m3u8") {
			res = append(res, content.Key)
		}
	}
	return res
}

// 后续由前端自己配置图片处理参数
func (_this *StorageHelper) ProcessResizeImage(filePath string) string {
	// 生成带处理参数的 URL
	switch _this.storageConfig.Provider {
	case ProviderAli:
		return filePath + "?x-oss-process=image/resize,l_200"
	case ProviderEcloudObs:
		return filePath + "?x-image-process=image/resize,l_200"
	default:
		return filePath
	}
}

// 迁移文件跨对象存储
func (_this *StorageHelper) MigrationFileObject(
	key string,
	targetHelper *StorageHelper,
	metadata *Metadata,
) error {
	localTmpFilePath := "/tmp/" + filepath.Base(key)
	// 从源头流式获取文件
	err := _this.GetFile(key, localTmpFilePath)
	if err != nil {
		return err
	}
	// 确保在函数结束时删除临时文件
	defer os.Remove(localTmpFilePath)
	// 流式上传文件到目标
	err = targetHelper.PutFileWithMeta(key, localTmpFilePath, metadata)
	return err
}

func (_this *StorageHelper) IsPublicAcl(key string) (bool, error) {
	return _this.st.IsPublicAcl(_this.storageConfig.Bucket, key)
}
