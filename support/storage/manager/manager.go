package manager

import (
	"fmt"
	"sync"
	"tools-thinker/support/storage/config"
	"tools-thinker/support/storage/driver"
)

type storageManager struct {
	Helpers sync.Map
}

type msgChan struct {
	provider string
	bucket   string
	helper   *driver.StorageHelper
}

var (
	sm *storageManager
)

// 初始化存储管理器
func initStorageManager(configs []*config.StorageConfig) (err error) {
	sm, err = newStorageManager(configs)
	if err != nil {
		return err
	}
	return nil
}

// 根据provider获取对应的helper
func getOssHelper(provider, bucket string) (*driver.StorageHelper, error) {
	notFoundErr := fmt.Errorf("provider %s not found", provider)
	key := getProviderRegionKey(provider, bucket)
	if v, has := sm.Helpers.Load(key); has {
		if helper, ok := v.(*driver.StorageHelper); ok {
			return helper, nil
		}
		return nil, notFoundErr
	}
	return nil, notFoundErr
}

// 获取存储助手
// 这里是获取所有支持的云对象存储, 通过provider来获取对应的helper
// 目前针对阿里云oss要使用两个区域进行了修改兼容
// 并发处理各自云的对象存储客户端
func newStorageManager(configs []*config.StorageConfig) (*storageManager, error) {
	// 初始化一个storageManager对象
	sm := &storageManager{}
	// 创建一个通道，用于传递helper对象和对应的provider
	helpersCh := make(chan msgChan, len(configs))
	errorsCh := make(chan error, len(configs)) // 使用缓冲通道，避免阻塞
	var wg sync.WaitGroup
	for _, conf := range configs {
		wg.Add(1)
		go func(c *config.StorageConfig) {
			defer wg.Done()
			helper, err := driver.New2(c)
			if err != nil {
				errorsCh <- fmt.Errorf("create %s storage helper error: %s", c.Provider, err)
				return
			}
			select {
			case helpersCh <- msgChan{bucket: c.Bucket, provider: c.Provider, helper: helper}:
			case <-errorsCh: // 如果errorsCh中有错误，停止发送新的helper
				return
			}
		}(conf)
	}

	// 等待所有goroutine完成
	wg.Wait()
	close(helpersCh)
	close(errorsCh) // 在所有goroutine完成后关闭errorsCh

	// 检查是否有错误发生
	if err, ok := <-errorsCh; ok {
		return nil, err // 返回遇到的第一个错误
	}

	// 获取已经放入通道的数据
	for msg := range helpersCh {
		key := getProviderRegionKey(msg.provider, msg.bucket)
		sm.Helpers.Store(key, msg.helper)
	}

	return sm, nil
}

func getProviderRegionKey(provider, bucket string) string {
	if len(bucket) == 0 {
		return provider // 如果bucket为空，使用provider作为key
	}
	return fmt.Sprintf("%s-%s", provider, bucket)
}
