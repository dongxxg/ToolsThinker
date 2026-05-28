package manager

import (
	"tools-thinker/support/storage/config"
)

func GetStorageConfigs(inputConfigs *config.OssConfigs) GetCloudyStorageConfigs {
	var configs []*config.StorageConfig
	for _, preConf := range inputConfigs.Configs {
		configs = append(configs, preConf)
	}
	return func() []*config.StorageConfig {
		return configs
	}
}
