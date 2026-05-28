package rabbitmq

import (
	"time"
	"tools-thinker/support/logger"
)

// 负责重连
type monitor struct {
	id          string
	g           *AMQPGroup
	retry       int           // 一次重连过程中，重试了几次
	breakTimes  int           // 启动后，断线总次数
	refreshChan chan struct{} // 刷新监听errChan
}

func newMonitor(g *AMQPGroup) *monitor {
	res := &monitor{
		id:          g.que,
		g:           g,
		retry:       0,
		breakTimes:  0,
		refreshChan: make(chan struct{})}
	return res
}

func (m *monitor) work() {
	for m.doWork() {

	}
	logger.Info("[AMQP] monitor goroutine stopped")
}
func (m *monitor) doWork() (retry bool) {
	defer func() {
		if err := recover(); err != nil {
			retry = true
		}
	}()
	logger.Info("[AMQP] monitor working!")
	for {
		select {
		case _, ok := <-m.g.chanErr:
			if ok {
				logger.Info("receive chan err, recon.........")
				m.recon()
			}
		case _, ok := <-m.g.cls:
			if !ok {
				return false
			}
		case <-m.refreshChan:
			// 更新监听的m.g.chanErr
			logger.Info("[AMQP] monitor %s receive refresh chan", m.id)
		}
	}
}

func (m *monitor) refresh() {
	// 阻塞时跳过
	select {
	case m.refreshChan <- struct{}{}:
	default:
	}

}

func (m *monitor) recon() {
	logger.Info("[AMQP] monitor start reconnect mq connection")
	for !m.sleepAndRecon() {
		logger.Info("[AMQP] retry connect next time")
	}
	m.breakTimes++
	logger.Info(
		"[AMQP] monitor reconnect success, retry:%d, amqp break:%d",
		m.retry,
		m.breakTimes,
	)
	m.retry = 0
}

func (m *monitor) sleepAndRecon() (success bool) {
	defer func() {
		if err := recover(); err != nil {
			logger.Error("[AMQP] reconnect failed, at time :%d", m.retry)
			success = false
		}
	}()
	// 先停止
	m.g.Stop(true)
	// 再sleep
	logger.Info("[AMQP] monitor goroutine sleep 5s... at:%d", m.retry)
	time.Sleep(5 * time.Second)

	if !m.g.connection.IsClosed() {
		logger.Info("[AMQP] amqp group already reconnected")
		return true
	}
	m.retry++
	m.g.start(m.g.mode, m.g.cNum, m.g.handler, m.g.rq, m.g.qos, m.g.qosOnConn, true)

	return true
}
