package config

import (
	"tools-thinker/support"
)

type URLPath struct {
	ID         int64
	UniqueName string
	Name       string
	Region     string
	Bucket     string
	Path       string
	Provider   string
	Host       *string // cdn域名
	Note       *string // 备注
}

type StorageConfig struct {
	Provider         string          `json:"provider"` // ali ecloudObs minio
	Protocol         string          `json:"protocol"` // oss:// obs:// minio://
	AccessKeyId      string          `json:"accessKeyId"`
	AccessKeySecret  string          `json:"accessKeySecret"`
	Root             string          `json:"root"`
	BaseDir          string          `json:"baseDir"`
	RoleArn          string          `json:"roleArn"`
	Bucket           string          `json:"bucket"`
	Region           string          `json:"region"`
	Internal         support.BoolStr `json:"internal"`
	Endpoint         string          `json:"endpoint"`
	EndpointInternal string          `json:"endpointInternal"`
	StsEndPoint      string          `json:"stsEndPoint"`
	Host             string          `json:"host"`
	CustomHost       string          `json:"customHost"`
	Path             []string        `json:"path"`
	TmpPath          []string        `json:"tmpPath"`
	TmpRoot          string          `json:"tmpRoot"`
	CdnDomain        string          `json:"cdnDomain"`
	CdnProtocol      string          `json:"cdnProtocol"`
}

/*
OssConfigs 配置中心下发的配置文件
*/
type OssConfigs struct {
	Configs []*StorageConfig `json:"configs"`
}

// StorageConfigProvider 对象存储配置获取接口，外部使用方实现此接口
type StorageConfigProvider interface {
	GetOssConfigs() *OssConfigs
	GetStorageConfig() *StorageConfig
	GetUrlPath() *URLPath
}
