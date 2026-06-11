package common

// @Author dx
// @Date 2025-07-20 22:26:00
// @Desc UI 日志适配器，同时写入控制台和 UI 面板

import (
	"fmt"

	"tools-thinker/support/logger"
)

// UILogger 同时写入控制台日志和 UI 日志面板的适配器
type UILogger struct {
	Prefix  string
	LogToUI func(string)
	log     logger.LogPrefix
}

// NewUILogger 创建 UI 日志适配器
func NewUILogger(prefix string, logToUI func(string)) *UILogger {
	return &UILogger{
		Prefix:  prefix,
		LogToUI: logToUI,
		log:     logger.LogPrefix("[" + prefix + "]"),
	}
}

// Info 记录信息日志
func (l *UILogger) Info(format string, v ...any) {
	msg := fmt.Sprintf(format, v...)
	l.log.Info(msg)
	if l.LogToUI != nil {
		l.LogToUI(msg)
	}
}

// Warn 记录警告日志
func (l *UILogger) Warn(format string, v ...any) {
	msg := fmt.Sprintf(format, v...)
	l.log.Warn(msg)
	if l.LogToUI != nil {
		l.LogToUI("[警告] " + msg)
	}
}

// Error 记录错误日志
func (l *UILogger) Error(format string, v ...any) {
	msg := fmt.Sprintf(format, v...)
	l.log.Error(msg)
	if l.LogToUI != nil {
		l.LogToUI("[错误] " + msg)
	}
}

// Progress 记录进度信息
func (l *UILogger) Progress(current, total int, detail string) {
	msg := fmt.Sprintf("[%d/%d] %s", current, total, detail)
	l.log.Info(msg)
	if l.LogToUI != nil {
		l.LogToUI(msg)
	}
}
