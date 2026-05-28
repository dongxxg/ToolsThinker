/**
 * @author  zhaoliang.liang
 * @date  2025/1/22 15:26
 */

package storage

import (
	"tools-thinker/support/logger"
	"tools-thinker/support/storage/config"
	"tools-thinker/support/storage/driver"
	"tools-thinker/support/storage/manager"
)

// UrlPaths 不同对象存储驱动对应的Url信息
var UrlPaths []*config.URLPath

// GlobalStorageConfigs 全局对象存储配置
var GlobalStorageConfigs *config.OssConfigs

// OssHelper 对象存储助手
var OssHelper *driver.StorageHelper

// initStorageConfig 初始化对象存储全局配置
func initStorageConfig(configs []*config.StorageConfig) {
	if len(configs) == 0 {
		panic("生成ossConfig失败,storageConfig为空")
	}
	var ossConfig []*config.StorageConfig
	for _, preConf := range configs {
		ossConfig = append(ossConfig, preConf)
	}
	GlobalStorageConfigs = &config.OssConfigs{
		Configs: ossConfig,
	}

}

// InitOssManual 手动初始化对象存储实例
func InitOssManual(storageConfig *config.StorageConfig) {
	var err error
	getStorageConfigFunc := func() *config.StorageConfig {
		storageConfig := storageConfig
		if storageConfig.Internal && storageConfig.EndpointInternal != "" {
			storageConfig.Endpoint = storageConfig.EndpointInternal
		}
		return storageConfig
	}

	OssHelper, err = manager.InitSpecialManager(getStorageConfigFunc)
	if err != nil {
		logger.Error("InitSpecialOss InitOssManual error %s ", err)
		panic(err)
	}
}

// InitOssByInsert 通过插入初始化对象存储实例
func InitOssByInsert(helper *driver.StorageHelper) {
	OssHelper = helper
}

// InitStorageManagerManual 手动初始化对象管理器
func InitStorageManagerManual(urlPaths []*config.URLPath, configs config.OssConfigs) {
	getCloudyUrlPath := func() []*config.URLPath {
		return urlPaths
	}
	err := manager.InitCloudyManager(getCloudyUrlPath, manager.GetStorageConfigs(&configs))
	if err != nil {
		logger.Error("InitSpecialOss InitCloudyManual error %s ", err)
		panic(err)
	}
}

// InitStorageHelper 初始化对象存储助手
func InitStorageHelper(storageConfig *config.StorageConfig) (*driver.StorageHelper, error) {
	getStorageConfigFunc := func() *config.StorageConfig {
		return storageConfig
	}
	return manager.InitSpecialManager(getStorageConfigFunc)
}
