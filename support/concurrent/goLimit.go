package concurrent

import (
	"runtime/debug"
	"sync"
	"tools-thinker/support/logger"
)

// GoLimit 限制一个同步任务中并发的协程数
type GoLimit struct {
	c  chan struct{}
	wg *sync.WaitGroup
}

func NewGoLimit(size int) *GoLimit {
	return &GoLimit{
		c:  make(chan struct{}, size),
		wg: &sync.WaitGroup{},
	}
}

func (g *GoLimit) Run(f func()) *GoLimit {
	g.wg.Add(1)
	g.c <- struct{}{}
	go func() {
		defer func() {
			if err := recover(); err != nil {
				logger.Error("go limit run err: %s\n%s", err, debug.Stack())
			}
			g.wg.Done()
			<-g.c
		}()
		f()
	}()
	return g
}

func (g *GoLimit) Wait() {
	g.wg.Wait()
}