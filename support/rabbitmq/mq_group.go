package rabbitmq

import (
	"github.com/streadway/amqp"
	"log"
	"net"
	"runtime/debug"
	"strconv"
	"time"
	"tools-thinker/support/logger"
)

type mode int

const (
	OnlyProducer mode = 1
	OnlyConsumer mode = 2
	Mixed        mode = 3
)

// 一个MQGroup持有一个connection
// 一个connection多个channel,每个channel一个协程。
// 启动模式（只有生产者，只有消费者，混合模式）
// 断线重连
type AMQPGroup struct {
	config
	msgChan    chan mMsg
	connection *amqp.Connection
	monitor    *monitor
	producer   *producer
	consumers  []*consumer

	cls      chan struct{} // 所有协程退出信号
	chanErr  chan struct{} // mq role 汇报错误
	roleExit chan struct{} // 通知mq role退出

	mode    mode
	cNum    int
	handler MsgHandler
}

// 必要的配置
type config struct {
	que          string // 队列名称
	rk           string // 路由key
	exn          string // 交换机名称
	url          string // mq服务器url
	exchangeType string // 交换机类型

	rq        bool // 拒绝消息后是否重新入队,重新入队,一定要配置超时和dlx
	qos       int  // 下发的未ack的消息数量
	qosOnConn bool // qos是不是作用于connection上
}

func NewAMQPGroup(opts ...FuncOpt) *AMQPGroup {
	config := new(config)
	for _, option := range opts {
		option(config)
	}

	mq := new(AMQPGroup)
	mq.msgChan = make(chan mMsg, 1024)
	mq.cls = make(chan struct{}, 0)
	mq.config = *config
	mq.monitor = newMonitor(mq)
	go mq.monitor.work()
	return mq
}

type FuncOpt func(options *config)

func WithQueue(queue string) FuncOpt {
	return func(options *config) {
		options.que = queue
	}
}

func WithExchangeType(exchangeType string) FuncOpt {
	return func(options *config) {
		options.exchangeType = exchangeType
	}
}

func WithExchangeName(exchangeName string) FuncOpt {
	return func(options *config) {
		options.exn = exchangeName
	}
}

func WithUrl(url string) FuncOpt {
	return func(options *config) {
		options.url = url
	}
}

func WithRoutingKey(routingKey string) FuncOpt {
	return func(options *config) {
		options.rk = routingKey
	}
}

// 只生产
func (g *AMQPGroup) StartOnlyProducerMode() {
	g.start(OnlyProducer, 0, nil, true, 1, true, false)
}

// 只消费
// num 启动的consumer携程数量，建议最好不要超过10个
// rq : 拒绝的消息是否重新入队
// qos : 服务器发往client的未ack的消息总量
// qosOnConn : 生效在connection上
func (g *AMQPGroup) StartOnlyConsumerMode(
	num int,
	handler MsgHandler,
	rq bool,
	qos int,
	qosOnConn bool,
) {
	g.start(OnlyConsumer, num, handler, rq, qos, qosOnConn, false)
}

// 既生产又消费
func (g *AMQPGroup) StartMixedMode(num int, handler MsgHandler, rq bool, qos int, qosOnConn bool) {
	g.start(Mixed, num, handler, rq, qos, qosOnConn, false)
}

// 停止，关闭任务协程，关闭通信通道
func (g *AMQPGroup) Stop(isRecon bool) {
	if isRecon {
		close(g.roleExit)
	} else {
		close(g.cls)     // 协程关闭信号
		close(g.msgChan) // 关闭消息发送通道
	}
	g.closeAmqp()

}

// 起动失败，该方法会panic
func (g *AMQPGroup) start(
	mode mode,
	num int,
	handler MsgHandler,
	rq bool,
	qos int,
	qosOnConn bool,
	isRecon bool,
) {
	defer func() {
		if err := recover(); err != nil {
			logger.Error("[AMQP] start caused panic:%s", err)
			if !isRecon {
				g.Stop(false)
			} else {
				panic(err)
			}
		}
	}()
	// 消费者配置
	g.config.rq = rq
	g.config.qos = qos
	g.config.qosOnConn = qosOnConn
	g.mode = mode
	g.cNum = num
	g.handler = handler
	// 重连属性
	g.chanErr = make(chan struct{}, num)
	g.monitor.refresh()
	g.roleExit = make(chan struct{}, 0)
	// 创建connection
	g.createConn()
	// 判断模式
	switch mode {
	case OnlyProducer:
		g.createProducer()
	case OnlyConsumer:
		g.createConsumers(num, handler)
	case Mixed:
		g.createProducer()
		g.createConsumers(num, handler)
	default:
		panic("[AMQP] unknown amqp group start mode")
	}
	// 启动
	if mode != OnlyConsumer {
		go func() {
			for g.producer.work() {

			}
			logger.Info("[AMQP] producer goroutine stopped")
		}()
	}
	if mode != OnlyProducer {
		for _, consumer := range g.consumers {
			temp := consumer
			go func() {
				for temp.work() {

				}
				logger.Info("[AMQP] consumer %s goroutine stopped", temp.id)
			}()
		}
	}
	logger.Info("[AMQP] start amqp group success")
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

func newConn(a *AMQPGroup) (*amqp.Connection, error) {
	logger.Info("[AMQP] start create amqp connection, wait at most 1 minute...")
	newConn, err := amqp.DialConfig(a.url, amqp.Config{
		Dial: func(network, addr string) (net.Conn, error) {
			return net.DialTimeout(network, addr, 1*time.Minute)
		},
	})
	if err == nil {
		logger.Info("[AMQP] create amqp connection success")
	}
	return newConn, err
}

func (g *AMQPGroup) createConn() {
	newConn, err := newConn(g)
	failOnError(err, "[AMQP] create base connection error")
	g.connection = newConn
}

func (g *AMQPGroup) createProducer() {
	newChan, err := g.connection.Channel()
	failOnError(err, "[AMQP] create producer channel error")
	g.producer = newProducer(g.que, g, newChan)
}

func (g *AMQPGroup) createConsumers(num int, handler MsgHandler) {
	consumers := make([]*consumer, num)
	for id := 0; id < num; id++ {
		newChan, err := g.connection.Channel()
		failOnError(err, "[AMQP] create consumer channel error")
		// handler
		var handlerInstance MsgHandler
		if id == 0 {
			handlerInstance = handler
		} else {
			handlerInstance = handler.NewInstance()
		}
		consumers[id] = newConsumer(g.que+strconv.Itoa(id), g, newChan, handlerInstance)
		logger.Info("[AMQP] create consumer success, channel: %d", id)
	}
	g.consumers = consumers
}

func (g *AMQPGroup) closeAmqp() {
	logger.Info("[AMQP] start close amqp connection...")
	for _, consumer := range g.consumers {
		err := consumer.channel.Close()
		if err != nil {
			logger.Warn("[AMQP] close consumer %s channel error:%s\n", consumer.id, err)
		}
	}
	if g.producer != nil {
		err := g.producer.channel.Close()
		if err != nil {
			logger.Warn("[AMQP] close producer %s connection error:%s\n", g.producer.id, err)
		}
	}
	if g.connection != nil || !g.connection.IsClosed() {
		err := g.connection.Close()
		if err != nil {
			logger.Warn("[AMQP] close amqp connection error:%s \n", err)
		}
	}
	logger.Info("[AMQP] close amqp connection success")
}

func (g *AMQPGroup) roleNum() int {
	var res int
	if g.producer != nil {
		res++
	}
	if g.consumers != nil {
		res += len(g.consumers)
	}
	return res
}

type MsgHandler interface {
	Handle([]byte) error     // 处理消息逻辑
	AfterAck()               // 消息消费成功之后
	AfterReject()            // 消息拒绝之后执行
	NewInstance() MsgHandler // 返回一个全新的 msgHandler, amqp实现中，每个channel单独使用一个msgHandler
}

// 业务发送消息方法
func (g *AMQPGroup) SendMessage(message []byte, logPrefix string) {
	defer func() {
		if err := recover(); err != nil {
			logger.Error("[AMQP] send msg on close channel, %s\n%s", message, debug.Stack())
		}
	}()
	msg := mMsg{logPrefix: logPrefix, msg: message}
	g.msgChan <- msg
}

type mMsg struct {
	logPrefix string
	msg       []byte
	failTimes int
}
