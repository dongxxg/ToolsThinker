package driver

import (
	"fmt"
	"tools-thinker/support/logger"
	"tools-thinker/support/storage/config"

	openapi "github.com/alibabacloud-go/darabonba-openapi/client"
	sts20150401 "github.com/alibabacloud-go/sts-20150401/client"
	"github.com/alibabacloud-go/tea/tea"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

const (
	ProviderAli       = "ali"
	ProviderHuawei    = "huawei"
	ProviderEcloudObs = "ecloudObs"
	ProviderMinio     = "minio"
)

func New2(config *config.StorageConfig) (*StorageHelper, error) {
	Provider := config.Provider
	AccessKeyId := config.AccessKeyId
	AccessKeySecret := config.AccessKeySecret
	Endpoint := config.Endpoint
	StsEndpoint := config.StsEndPoint
	isInternal := config.Internal
	EndpointInternal := config.EndpointInternal
	var st Storage
	switch Provider {
	case ProviderAli:

		aliEndPoint := Endpoint
		if isInternal && len(EndpointInternal) != 0 {
			aliEndPoint = EndpointInternal
		}
		logger.Info("create ali oss client for %s", aliEndPoint)
		client, err := oss.New(aliEndPoint, AccessKeyId, AccessKeySecret)

		if err != nil {
			logger.Warn("create oss client error %s", err)
			client = nil
		}
		stsClient, err := CreateAliStsClient(StsEndpoint, AccessKeyId, AccessKeySecret)
		if err != nil {
			logger.Warn("create oss stsclient failed ，err = %s", err)
			stsClient = nil
		}
		st = &AliStorage{config: config, client: client, stsClient: stsClient}
		return &StorageHelper{
			st:            st,
			storageConfig: config,
		}, nil
	case ProviderEcloudObs:
		logger.Info("create ecloud obs client")
		ecloudObsStorageValue, err := NewEcloudObsStorage(config)
		if err != nil {
			logger.Warn("create obs client error %s", err)
			return nil, err
		}
		st = ecloudObsStorageValue
	case ProviderMinio:
		logger.Info("create minio client")
		minioClient, err := newMinioStorage(config)
		if err != nil {
			logger.Warn("create minio client error %s", err)
			return nil, err
		}
		st = minioClient
	default:
		return nil, fmt.Errorf("cant support provider %s", Provider)
	}

	return &StorageHelper{
		st:            st,
		storageConfig: config,
	}, nil
}

func CreateAliStsClient(
	endpoint string,
	accessKeyId string,
	accessKeySecret string,
) (client *sts20150401.Client, err error) {
	config := &openapi.Config{
		AccessKeyId:     tea.String(accessKeyId),
		AccessKeySecret: tea.String(accessKeySecret),
		Endpoint:        tea.String(endpoint),
	}
	client, err = sts20150401.NewClient(config)
	return client, err
}
