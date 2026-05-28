# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 项目概述

**tools-thinker** 是一个基于 Fyne 的 Go 桌面 GUI 应用，提供办公工具集（Excel 合并、Word 处理等），采用左右分栏布局。

## 构建与运行

```bash
# 运行应用
go run main.go

# 构建可执行文件
go build -o tools-thinker.exe .

# 运行全部测试
go test ./...

# 运行测试（竞态检测 + 覆盖率）
go test -race -cover ./...

# 运行单个测试
go test -run TestName -v ./support/collection/_map/...
```

## 架构

```
main.go                  → 入口，调用 internal.InitApp()
internal/
  mainApp.go             → 应用引导：创建 Fyne 应用/窗口，组装左右分栏布局（左 22% : 右 78%）
  layout/
    left/whole.go        → 左侧菜单栏（Excel、Word 按钮）
    right/whole.go       → 右侧面板：上部内容区 + 下部日志输出（70/30 VSplit）
    excel/excel.go       → Excel 功能界面：合并按钮、文件夹选择器，委托给 merge handler
  excel/feature/
    merge/handler.go     → Excel 合并逻辑：读取 data/ 下所有 .xlsx，合并输出为 merged.xlsx
  common/file.go         → 共用的文件打开对话框工具
myTheme/
  myTheme.go             → 自定义 Fyne 主题，通过 //go:embed 嵌入微软雅黑字体（msyh.ttf）
support/                 → 独立工具库（不依赖 Fyne）
  collection/            → 通用数据结构：map、set、slice、tree 工具集
  concurrent/            → 并发原语：协程池、限流器、队列、互斥锁
  crypt/                 → 加密工具：AES、RSA、Base64
  docx/                  → Word 文档生成（go-docx 封装）
  logger/                → 自定义日志：彩色输出 + goroutine ID
  storage/               → 多云对象存储（阿里云、华为云、天翼云、MinIO），工厂模式
  rabbitmq/              → RabbitMQ 客户端封装
  http_util/             → HTTP 客户端工具
  media/                 → M3U8 播放列表解析/生成
  file/                  → 文件系统工具（按扩展名枚举文件）
  ...                    → ip2region、密码生成、健康检查、图表等
```

### 核心模式

- **UI 结构**：左右 `HSplit` 布局——左侧菜单导航，右侧渲染功能界面。右侧本身是 `VSplit`（内容 70% / 日志 30%）。
- **功能导航**：左侧按钮调用 `internal/layout/<feature>` 包，直接更新 `right.RefreshContent` 和 `right.PrintLog`。
- **support/ 解耦**：`support/` 是不依赖 Fyne 的通用 Go 工具库，可独立使用。
- **存储工厂**：`support/storage/driver/storageFactory.go` 通过工厂模式创建存储后端（阿里云 OSS、华为云 OBS、MinIO、天翼云）。
- **自定义主题**：`myTheme/` 通过 `//go:embed` 嵌入 `msyh.ttf`，解决 Fyne 中文渲染问题。

## 模块信息

- 模块名：`tools-thinker`（Go 1.24）
- GUI 框架：[Fyne v2](https://fyne.io/) 跨平台桌面应用
- 主要依赖：excelize（Excel）、go-docx（Word）、go-charts（图表）、minio-go、aliyun-oss-sdk
