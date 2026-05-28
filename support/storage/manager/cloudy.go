package manager

import (
	"fmt"
	"sync"
	"tools-thinker/support/logger"
	"tools-thinker/support/storage/config"
	"tools-thinker/support/storage/driver"
)

// 构建所有查询结构
var (
	nameByOssHelperMap sync.Map
	idByOssHelperMap   sync.Map
)

type MultiCloudOssHelp struct {
	Helper  *driver.StorageHelper
	UrlPath *config.URLPath
}

type GetCloudyUrlPath func() []*config.URLPath
type GetCloudyStorageConfigs func() []*config.StorageConfig

// initCloudyOss 初始化云存储服务配置
func initCloudyOss(getCloudyUrlPath GetCloudyUrlPath, GetConfigs GetCloudyStorageConfigs) error {
	configs := GetConfigs()
	if len(configs) == 0 {
		return fmt.Errorf("no oss config")
	}
	err := initStorageManager(configs)
	if err != nil {
		logger.Warn("initNormalOss InitStorageManager error %s ", err)
		return err
	}
	// 针对urlpath获取
	urlPaths := getCloudyUrlPath()
	for _, urlPath := range urlPaths {
		helper, err := getOssHelper(urlPath.Provider, urlPath.Bucket)
		if err != nil {
			return err
		}
		m := &MultiCloudOssHelp{
			Helper:  helper,
			UrlPath: urlPath,
		}
		key := setProviderNameKey(m.UrlPath.Name, m.UrlPath.Provider)
		nameByOssHelperMap.Store(key, m)
		idByOssHelperMap.Store(urlPath.ID, m)
	}
	return nil
}

/*
GetUseOssHelperByName
@Description: 通过名称和云服务提供商获取云存储助手
@param name 名称 resource history
@param cloudProvider 云服务提供商 ALIYUN ECLOUD
@param storageServerProvider 存储提供商 ali ecloudObs minio 优先级最高
*/
func GetUseOssHelperByName(
	name,
	cloudProvider string,
	storageServerProviderMap map[string]string,
) (*MultiCloudOssHelp, error) {
	storageServerProvider := storageServerProviderMap[cloudProvider]
	if storageServerProvider == "" {
		return nil, fmt.Errorf("not find storageServerProvider %s", cloudProvider)
	}
	// resource-ali
	key := setProviderNameKey(name, storageServerProvider)
	errNotFound := fmt.Errorf("not find oss helper %s cloudProvider %s", key, cloudProvider)
	if v, has := nameByOssHelperMap.Load(key); has {
		if helper, ok := v.(*MultiCloudOssHelp); ok {
			return helper, nil
		}
		return nil, errNotFound
	}
	return nil, errNotFound
}

func GetUseOssHelperById(id int64) (*MultiCloudOssHelp, error) {
	if v, has := idByOssHelperMap.Load(id); has {
		if helper, ok := v.(*MultiCloudOssHelp); ok {
			return helper, nil
		}
		return nil, fmt.Errorf("not find oss helper %d", id)
	}
	return nil, fmt.Errorf("not find oss helper %d", id)
}

func setProviderNameKey(name, provider string) string {
	return fmt.Sprintf("%s-%s", name, provider)
}