package rabbitmq

import (
	"encoding/json"
	"fmt"
	"tools-thinker/support/logger"
	"tools-thinker/support/util"
)

type Config struct {
	Provider        string
	AccessKeyId     string
	AccessKeySecret string
	Vhost           string
	Url             string
	ResourceOwnerId string
	User            string
	Password        string
	Port            int
}

func (m *Config) MarshalJSON() ([]byte, error) {
	type Alias Config
	return json.Marshal(&struct {
		*Alias
		AccessKeyId     string
		AccessKeySecret string
	}{
		Alias:           (*Alias)(m),
		AccessKeyId:     util.EncryptKey(m.AccessKeyId),
		AccessKeySecret: util.EncryptKey(m.AccessKeySecret),
	})
}

func GetMqUrl(c *Config) string {
	// return fmt.Sprintf("amqp://%s:%s@%s:5672/%s", "guest", "guest", "10.21.4.100", "")
	ak := c.AccessKeyId
	sk := c.AccessKeySecret
	vhost := c.Vhost
	url := c.Url

	userName := ""
	password := ""
	provider := c.Provider
	port := 5672
	if provider == "ali" {
		instanceId := c.ResourceOwnerId

		userName = GetUserName(ak, instanceId)
		password = GetPassword(sk)
		logger.Info("mq pass word:%v %v ", sk, password)
	} else {
		userName = c.User
		password = c.Password
		port = c.Port
	}

	lastUrl := fmt.Sprintf(
		"amqp://%s:%s@%s:%d/%s?heartbeat=0",
		userName,
		password,
		url,
		port,
		vhost,
	)
	return lastUrl
}

type QueueConfig struct {
	QueueName    string `json:"queueName"`
	RoutingKey   string `json:"routingKey"`
	ExchangeName string `json:"exchangeName"`
}
