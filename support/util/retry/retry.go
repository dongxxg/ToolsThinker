package retry

import (
	"tools-thinker/support/logger"
	"time"
)

// 指定重试时间组
func Do(f func() bool, rules []int) {
	doRes := make(chan bool)
	defer close(doRes)
	index := 0
	for {
		go time.AfterFunc(time.Duration(rules[index])*time.Second, func() {
			doRes <- f()
		})
		if <-doRes {
			if index != 0 {
				logger.Info("retry success at %d", index)
			}
			return
		} else {
			if index != 0 {
				logger.Warn("retry failed at %d", index)
			}
		}
		if index == len(rules)-1 {
			return
		}
		index++
	}
}

// 固定间隔重试
func TickDo(f func() bool, tick int) {
	doRes := make(chan bool)
	defer close(doRes)
	index := 0
	t := 0 // 第一次直接执行
	for {
		go time.AfterFunc(time.Duration(t)*time.Second, func() {
			doRes <- f()
		})
		if <-doRes {
			if index != 0 {
				logger.Info("retry success at %d", index)
			}
			return
		} else {
			if index != 0 {
				logger.Warn("retry failed at %d", index)
			}
		}
		index++
		t = tick
	}
}
