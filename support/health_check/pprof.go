package hc

import (
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"tools-thinker/support/logger"
	"time"
)

// 获取Cpu使用率百分比，例如使用20%，就返回20
func GetCpuUsage() float64 {
	if percent, err := cpu.Percent(time.Second, false); err != nil || len(percent) == 0 {
		return 100 //默认为100，不会再往这台服务器分配新的连接
	} else {
		return percent[0]
	}
}

func GetMemUsage() float64 {
	if memStat, err := mem.VirtualMemory(); err == nil && memStat != nil {
		return memStat.UsedPercent
	} else {
		logger.Error("get memstat failed,as memstat is nil or err = %v", err)
		return 0
	}
}
