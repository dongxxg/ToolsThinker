// 阿里提供的生成用户名和密码：https://github.com/AliwareMQ/amqp-demos
package rabbitmq

import (
	"bytes"
	"encoding/base64"
	"strconv"
	"strings"
	"time"
	"tools-thinker/support/util"
)

const (
	ACCESS_FROM_USER = 0
	COLON            = ":"
)

func GetUserName(ak string, instanceId string) string {
	var buffer bytes.Buffer
	buffer.WriteString(strconv.Itoa(ACCESS_FROM_USER))
	buffer.WriteString(COLON)
	buffer.WriteString(instanceId)
	buffer.WriteString(COLON)
	buffer.WriteString(ak)
	return base64.StdEncoding.EncodeToString(buffer.Bytes())
}

func GetUserNameBySTSToken(ak string, instanceId string, stsToken string) string {
	var buffer bytes.Buffer
	buffer.WriteString(strconv.Itoa(ACCESS_FROM_USER))
	buffer.WriteString(COLON)
	buffer.WriteString(instanceId)
	buffer.WriteString(COLON)
	buffer.WriteString(ak)
	buffer.WriteString(COLON)
	buffer.WriteString(stsToken)
	return base64.StdEncoding.EncodeToString(buffer.Bytes())
}

func GetPassword(sk string) string {
	now := time.Now()
	currentMillis := strconv.FormatInt(now.UnixNano()/1000000, 10)
	return GetPassword2(sk, currentMillis)
}

func GetPassword2(sk string, ms string) string {

	currentMillis := ms
	var buffer bytes.Buffer
	// 阿里原来写反了参数。buffer.WriteString(strings.ToUpper(HmacSha1(sk,currentMillis)))
	buffer.WriteString(strings.ToUpper(util.HmacSha1(currentMillis, sk)))
	buffer.WriteString(COLON)
	buffer.WriteString(currentMillis)
	// fmt.Println(currentMillis)
	// fmt.Println(HmacSha1(sk, currentMillis))
	return base64.StdEncoding.EncodeToString(buffer.Bytes())
}
