package ip2RegionUtil

import (
	"fmt"
	"testing"
	"time"
	"tools-thinker/support/util/http_util"

	"github.com/lionsoul2014/ip2region/binding/golang/xdb"
)

func TestIp2Region(t *testing.T) {

	var err error
	var downloadPath = "https://github.com/lionsoul2014/ip2region/raw/master/data/ip2region.xdb"
	var dbPath = "./ip2region.xdb"
	var cBuff []byte

	// 下载
	err = http_util.Down2File(downloadPath, dbPath)
	if err != nil {
		fmt.Printf("failed to download `%s`: %s\n", downloadPath, err)
		return
	}
	cBuff, err = xdb.LoadContentFromFile(dbPath)

	if err != nil {
		fmt.Printf("failed to load content from `%s`: %s\n", dbPath, err)
		return
	}

	searcher, err := xdb.NewWithBuffer(cBuff)
	if err != nil {
		fmt.Printf("failed to create searcher: %s\n", err.Error())
		return
	}

	defer searcher.Close()

	var ip = "39.156.66.10"
	//var ip = "49.93.193.180"
	//var ip = "18.136.204.28"
	var tStart = time.Now()
	region, err := searcher.SearchByStr(ip)
	if err != nil {
		fmt.Printf("failed to SearchIP(%s): %s\n", ip, err)
		return
	}

	fmt.Printf("region: %s, took: %s\n", region, time.Since(tStart))
}
