package ip2RegionUtil

import (
	"errors"
	"fmt"
	"strings"
	"time"
	"tools-thinker/support/logger"

	"github.com/lionsoul2014/ip2region/binding/golang/xdb"
)

type Ip2RegionUtil struct {
	cBuff []byte
}

type Region struct {
	Province string
	City     string
}

type RegionType int

// 江苏省|南京市|电信
const (
	RegionTypeProvince RegionType = 0 // 省
	RegionTypeCity     RegionType = 1 // 市
)

// 加载ip2region.xdb文件到内存(10M左右)
func NewIp2RegionUtil(dbPath string) (*Ip2RegionUtil, error) {
	util := &Ip2RegionUtil{}
	var err error
	util.cBuff, err = xdb.LoadContentFromFile(dbPath)
	if err != nil {
		logger.Error("failed to load content from `%s`: %s\n", dbPath, err)
		return nil, err
	}
	return util, nil
}

// 根据ip获取地区,返回结构体
func (i *Ip2RegionUtil) GetRegion(ip string) (*Region, error) {
	regionArr, err := i.getRegionSlice(ip)
	if err != nil {
		return nil, err
	}

	if regionArr == nil {
		return nil, errors.New("regionArr is nil")
	}
	region := &Region{
		Province: regionArr[RegionTypeProvince],
		City:     regionArr[RegionTypeCity],
	}

	return region, nil
}

// 根据ip获取指定级别的地区信息,返回拼接好的字符串
func (i *Ip2RegionUtil) GetRegionStrWithSeparator(
	ip string,
	regionType []RegionType,
	separator string,
) (string, error) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered from panic:", r)
		}
	}()

	region, err := i.getRegionSlice(ip)
	if err != nil {
		return "", err
	}

	regionStr := ""
	if region != nil {
		regionLength := len(region)
		for index, v := range regionType {
			if index >= regionLength {
				continue
			}
			regionStr += region[v] + separator
		}
	}
	// 去除可能多出来的/
	regionStr = strings.Trim(regionStr, "/")

	return regionStr, nil
}

func (i *Ip2RegionUtil) getRegionSlice(ip string) ([]string, error) {
	searcher, err := xdb.NewWithBuffer(i.cBuff)
	if err != nil {
		logger.Error("failed to create searcher: %s\n", err.Error())
		return nil, err
	}

	defer searcher.Close()

	var tStart = time.Now()
	regionStr, err := searcher.SearchByStr(ip)
	if err != nil {
		logger.Error("failed to SearchIP(%s): %s\n", ip, err)
		return nil, err
	}

	logger.Info("region: %s, took: %s\n", regionStr, time.Since(tStart))

	// 中国|0|江苏省|南京市|电信
	splitRegion := strings.Split(regionStr, "|")
	/*if len(splitRegion) != 5 {
		return nil, errors.New("regionStr split error")
	}*/
	for j := range splitRegion {
		if splitRegion[j] == "0" {
			splitRegion[j] = ""
		}
	}
	return splitRegion, nil
}