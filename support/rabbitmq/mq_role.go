package rabbitmq

import (
	"errors"
	uuid "github.com/satori/go.uuid"
	"github.com/streadway/amqp"
	"time"
	"tools-thinker/support/logger"
)

type chanWarp struct {
	id      string
	channel *amqp.Channel
	g       *AMQPGroup
}

func (c *chanWarp) rebuild(id string, channel *amqp.Channel, g *AMQPGroup) {
	c.id = id
	c.channel = channel
	c.g = g
	if c.g.exchangeType != "" {
		err := c.channel.ExchangeDeclare(
			c.g.exn,
			c.g.exchangeType,
			true,
			false,
			false,
			false,
			nil)
		failOnError(err, "[AMQP] exchange declare error")
	}
	// 启动一个协程监听异常
	go c.listenNotifyCancelOrClosed()
}

func (c *chanWarp) listenNotifyCancelOrClosed() {
	notifyCloseChan := c.channel.NotifyClose(make(chan *amqp.Error, 1))
	notifyCancelChan := c.channel.NotifyCancel(make(chan string, 1))

	select {
	case err := <-notifyCloseChan:
		if err != nil {
			logger.Error("[AMQP] channel %s notify close:%s", c.id, err)
			c.g.chanErr <- struct{}{}
		}
		if err == nil {
			logger.Info("[AMQP] channel closed gracefully")
		}
	case err := <-notifyCancelChan:
		logger.Error("[AMQP] channel %s notify cancel:%s", c.id, err)
		c.g.chanErr <- struct{}{}
	}
}

type producer struct {
	chanWarp
}

func newProducer(id string, g *AMQPGroup, c *amqp.Channel) *producer {
	p := &producer{}
	p.rebuild(id, c, g)
	return p
}

func (p *producer) work() (retry bool) {
	defer func() {
		if err := recover(); err != nil {
			logger.Error("[AMQP] producer start failed:%s", p.id)
			retry = true
		}
	}()
	logger.Info("[AMQP] producer %s working!", p.id)
	reconnectTime := 0
loop:
	for {
		select {
		case _, ok := <-p.g.cls:
			if !ok {
				logger.Info("[AMQP] stop producer goroutine:%s... ", p.id)
				break loop
			}
		case _, ok := <-p.g.roleExit:
			if !ok {
				logger.Info("[AMQP] stop producer goroutine:%s... ", p.id)
				break loop
			}
		case message, ok := <-p.g.msgChan:
			if !ok {
				logger.Info("[AMQP] producer finish successful:%s", p.id)
				return
			}
			uuidVal := uuid.NewV4().String()
			sendErr := p.channel.Publish(p.g.exn, p.g.rk, false, false, amqp.Publishing{
				DeliveryMode: amqp.Persistent,
				ContentType:  "text/plain",
				Body:         message.msg,
				MessageId:    uuidVal,
				Timestamp:    time.Now(),
			})
			if sendErr != nil {
				logger.Error("[AMQP] %s,%s publish msg error, %d, msgId: %s, msg:%s error:%s \n", message.logPrefix, p.id, reconnectTime, uuidVal, message.msg, sendErr)
				if message.failTimes < 3 {
					message.failTimes++
					p.g.msgChan <- message
				} else {
					logger.Error("[AMQP] %s,%s publish discard:msgId: %s, msg:%s", message.logPrefix, p.id, uuidVal, message.msg)
				}
				p.g.chanErr <- struct{}{}
			} else {
				logger.Info("[AMQP] %s,%s publish successful msgId: %s msg:%s ", message.logPrefix, p.id, uuidVal, string(message.msg))
			}
		}
	}
	return false
}

type consumer struct {
	chanWarp
	handler MsgHandler
}

func newConsumer(id string, g *AMQPGroup, c *amqp.Channel, handler MsgHandler) *consumer {
	p := &consumer{}
	p.rebuild(id, c, g)
	p.handler = handler
	return p
}

func (c *consumer) rebuild(id string, channel *amqp.Channel, g *AMQPGroup) {
	c.chanWarp.rebuild(id, channel, g)
	// 绑定queue
	queue, err := c.channel.QueueDeclare(c.g.que,
		true,
		false,
		false,
		false,
		nil)
	failOnError(err, "[AMQP] consumer queue declare error")
	err = c.channel.QueueBind(queue.Name, c.g.rk, c.g.exn, true, nil)
	failOnError(err, "[AMQP] consumer bind queue error")
	// 限流参数含义：按数量（<=0不控制），按bytes（<=0不控制）,是否作用于conn（true：connection，false：channel）
	err2 := c.channel.Qos(c.g.qos, 0, c.g.qosOnConn)
	failOnError(err2, "consumer set channel pos failed")
	logger.Info("[AMQP] %s:qos=%d set success", c.id, c.g.qos)
}

func (c *consumer) work() (retry bool) {
	defer func() {
		if err := recover(); err != nil {
			logger.Error("[AMQP] consumer %s start failed:%s", c.id, err)
			retry = true
		}
	}()
	// push消费
	return c.push()
}

func (c *consumer) push() bool {
	msgChan, err := c.channel.Consume(c.g.que, c.g.que, false, false, false, true, nil)
	// 这里只会产生 call(req message, res ...message) error, args为nil
	if err != nil {
		logger.Error("[AMQP] consumer %s channel error:%s\n", c.id, err)
		c.g.chanErr <- struct{}{}
		return false
	}
	logger.Info("[AMQP] consumer %s working!", c.id)
	for {
		select {
		case _, ok := <-c.g.cls:
			if !ok {
				logger.Info("[AMQP] stop consumer goroutine:%s", c.id)
				return false
			}
		case _, ok := <-c.g.roleExit:
			if !ok {
				logger.Info("[AMQP] stop consumer goroutine:%s", c.id)
				return false
			}
		case msg, ok := <-msgChan:
			if ok {
				error := c.handleMsg(msg)
				if error != nil {
					logger.Error("[AMQP] consumer %s handle message error, msg: %s", c.id, msg.Body)
					msg.Reject(c.g.rq)
					c.afterReject()
				} else {
					// ack 参数含义，false：只ack当前的一-条msg。true: channel上已有的都会被ack
					msg.Ack(false)
					c.afterAck()
				}
			}
		}
	}
	return false
}

func (c *consumer) handleMsg(msg amqp.Delivery) (error error) {
	defer func() {
		if err := recover(); err != nil {
			error = errors.New("consume error")
		}
	}()
	// 处理数据
	logger.Info("[AMQP] consumer %s receive: %s, msg: %s", c.id, msg.Timestamp, msg.Body)
	error = c.handler.Handle(msg.Body)
	return error
}

func (c *consumer) afterAck() {
	defer func() {
		if err := recover(); err != nil {
			logger.Error("[AMQP] consumer %s execute afterAck failed", c.id)
		}
	}()
	c.handler.AfterAck()
}

func (c *consumer) afterReject() {
	defer func() {
		if err := recover(); err != nil {
			logger.Error("[AMQP] consumer %s execute afterReject failed", c.id)
		}
	}()
	c.handler.AfterReject()
}
