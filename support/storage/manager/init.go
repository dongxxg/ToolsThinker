package manager

import (
	"tools-thinker/support/storage/driver"
)

// 通过配置 manager.URLPath 来初始化对应的 ossHelper
func InitCloudyManager(
	getCloudyUrlPath GetCloudyUrlPath,
	GetConfigs GetCloudyStorageConfigs,
) error {
	return initCloudyOss(getCloudyUrlPath, GetConfigs)
}

// 使用 *storage.StorageConfig 来直接获取 ossHelper
func InitSpecialManager(GetConfig GetStorageConfig) (*driver.StorageHelper, error) {
	return initSpecialOss(GetConfig)
}
