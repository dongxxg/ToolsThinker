package rabbitmq

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
	"tools-thinker/support/logger"
)

// EnsureVHost 检查并创建 vHost（如果不存在）
// 参数：
//
//	apiURL: RabbitMQ 管理接口地址 (例如 "http://localhost:15672")
//	username: 管理员用户名
//	password: 管理员密码
//	vhost: 要检查/创建的虚拟主机名称
//
// 返回值：
//
//	error: 操作错误信息，成功时返回 nil
func EnsureVHost(apiURL, username, password, vhost string) error {
	// 创建带超时的 HTTP 客户端
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	// 检查 vHost 是否存在
	exists, err := checkVHostExists(client, apiURL, username, password, vhost)
	if err != nil {
		return fmt.Errorf("检查 vHost 失败: %w", err)
	}

	if exists {
		logger.Info("[AMQP] vhost %s already exist", vhost)
		return nil
	}

	// 创建不存在的 vHost
	if err = createVHost(client, apiURL, username, password, vhost); err != nil {
		return fmt.Errorf("创建 vHost 失败: %w", err)
	}

	return nil
}

// checkVHostExists 检查 vHost 是否存在
func checkVHostExists(client *http.Client, apiURL, username, password, vhost string) (bool, error) {
	// 构造 API 路径并进行 URL 编码
	path := fmt.Sprintf("/api/vhosts/%s", url.PathEscape(vhost))
	req, err := http.NewRequest("GET", apiURL+path, nil)
	if err != nil {
		return false, fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置基础认证
	req.SetBasicAuth(username, password)

	resp, err := client.Do(req)
	if err != nil {
		return false, fmt.Errorf("请求失败: %w", err)
	}
	defer func() {
		err = resp.Body.Close()
	}()

	switch resp.StatusCode {
	case http.StatusOK:
		return true, nil
	case http.StatusNotFound:
		return false, nil
	default:
		return false, fmt.Errorf("意外的响应状态码: %d", resp.StatusCode)
	}
}

// createVHost 创建新的 vHost
func createVHost(client *http.Client, apiURL, username, password, vhost string) error {
	logger.Info("[AMQP] to create vhost %s", vhost)
	// 构造创建路径
	path := fmt.Sprintf("/api/vhosts/%s", url.PathEscape(vhost))

	// 准备请求体（RabbitMQ API 需要空 JSON 对象）
	body := bytes.NewBufferString("{}")

	req, err := http.NewRequest("PUT", apiURL+path, body)
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置认证和内容类型
	req.SetBasicAuth(username, password)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		logger.Error("[AMQP] create vhost %s request %s failed %s", vhost, path, err.Error())
		return fmt.Errorf("请求失败: %w", err)
	}
	defer func() {
		err = resp.Body.Close()
	}()

	// 验证响应状态
	if resp.StatusCode != http.StatusCreated {
		var errorResponse struct {
			Error string `json:"error"`
		}
		if err = json.NewDecoder(resp.Body).Decode(&errorResponse); err != nil {
			logger.Error("[AMQP] create vhost %s json decode failed %s", vhost, err.Error())
			return fmt.Errorf("创建失败 (状态码 %d)，且无法解析错误响应", resp.StatusCode)
		}
		logger.Error("[AMQP] create vhost %s failed %s", vhost, err.Error())
		return fmt.Errorf("创建失败 (状态码 %d): %s", resp.StatusCode, errorResponse.Error)
	}
	logger.Info("[AMQP] create vhost %s success", vhost)
	return nil
}
