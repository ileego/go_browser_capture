# Browser DOM Capture - 智能网页元素抓取系统

<div align="center">

![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat\&logo=go)
![Chrome Extension](https://img.shields.io/badge/Chrome-Extension-green?style=flat\&logo=google-chrome)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-15+-336791?style=flat\&logo=postgresql)
![Redis](https://img.shields.io/badge/Redis-7.0+-DC382D?style=flat\&logo=redis)
![License](https://img.shields.io/badge/License-MIT-blue.svg)

**🚀 基于Chrome插件的分布式网页元素抓取解决方案**

[功能特性](#-功能特性) • [快速开始](#-快速开始) • [使用示例](#-使用示例) • [API文档](#-api文档)

</div>

***

## 📖 简介

Browser DOM Capture 是一个强大的网页元素抓取系统，通过 Chrome 插件与 Go 后端的协同工作，实现了高效、灵活、可扩展的网页数据抓取。

### 核心亮点

- 🎯 **智能匹配**：基于域名自动匹配抓取配置，无需手动指定选择器
- ⚡ **并行处理**：支持多插件并行抓取，大幅提升抓取效率
- 🔄 **负载均衡**：采用最少任务优先策略，智能分配抓取任务
- 🛡️ **反爬保护**：内置随机延迟机制，有效防止被识别为爬虫
- 📦 **批量操作**：支持批量URL抓取，一次请求处理多个目标
- 🎨 **可视化配置**：提供友好的插件界面，支持实时测试和配置

***

## ✨ 功能特性

### 后端服务

- **RESTful API**：基于 Gin 框架构建，提供完整的 CRUD 接口
- **WebSocket 通信**：实时与插件双向通信，支持任务分发和结果回传
- **插件池管理**：支持多插件连接，自动管理插件状态和健康检查
- **任务调度器**：智能任务调度，支持重试机制和超时控制
- **配置管理**：基于 DDD 架构的 Selector 配置管理，支持域名、选择器、正则表达式
- **批量抓取**：一次请求处理多个URL，自动匹配配置并并行执行
- **API 文档**：集成 Swagger，自动生成交互式 API 文档

### Chrome 插件

- **DOM 元素捕获**：精确捕获指定选择器的 HTML 和文本内容
- **新标签页执行**：在后台标签页打开目标页面，不影响用户浏览
- **智能队列管理**：自动管理抓取请求队列，防止请求过载
- **随机延迟**：可配置的随机延迟范围，防止被网站识别为爬虫
- **手动测试**：支持手动输入 URL 和选择器进行测试，方便调试
- **实时状态**：显示连接状态、插件 ID 等信息
- **配置持久化**：配置自动保存到本地存储

***

## 🏗️ 系统架构

```
┌─────────────────────────────────────────────────────────────┐
│                         用户请求                               │
│                    (批量URL + ID)                             │
└────────────────────────┬────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────┐
│                      Go 后端服务                              │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │  API 层      │  │  应用层      │  │  领域层      │      │
│  │  (Gin)       │  │  (Service)   │  │  (Domain)    │      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
│         │                 │                 │                │
│         ▼                 ▼                 ▼                │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │ 任务调度器   │  │ 插件池管理   │  │ 配置管理     │      │
│  │ (Scheduler)  │  │ (Pool)       │  │ (Config)     │      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
│         │                 │                                  │
│         └─────────────────┼──────────────────┐               │
│                           ▼                  │               │
│                  ┌─────────────────┐        │               │
│                  │ WebSocket 服务器 │        │               │
│                  └─────────────────┘        │               │
└─────────────────────────────────────────────┼───────────────┘
                                               │
                                               ▼
┌─────────────────────────────────────────────────────────────┐
│                    Chrome 插件池                             │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐   │
│  │ 插件 1   │  │ 插件 2   │  │ 插件 3   │  │ 插件 N   │   │
│  │ (队列)   │  │ (队列)   │  │ (队列)   │  │ (队列)   │   │
│  └──────────┘  └──────────┘  └──────────┘  └──────────┘   │
└─────────────────────────────────────────────────────────────┘
```

***

## 🚀 快速开始

### 环境要求

- Go 1.21+
- PostgreSQL 15+
- Redis 7.0+
- Chrome 浏览器（最新版）

### 1. 克隆项目

```bash
git clone https://github.com/ileego/go_browser_capture.git
cd go_browser_capture
```

### 2. 启动数据库服务

使用 Docker Compose 快速启动 PostgreSQL 和 Redis：

```bash
docker-compose up -d
```

数据库将运行在：

- PostgreSQL: `localhost:15432`
- Redis: `localhost:16379`

### 3. 启动后端服务

```bash
cd backend
go mod download
go run main.go
```

服务将在 `http://localhost:8080` 启动

### 4. 安装 Chrome 插件

1. 打开 Chrome 浏览器，访问 `chrome://extensions/`
2. 启用右上角的"开发者模式"
3. 点击"加载已解压的扩展程序"
4. 选择项目根目录下的 `extension` 文件夹

### 5. 验证安装

- 访问 `http://localhost:8080/swagger/index.html` 查看 API 文档
- 点击浏览器工具栏中的插件图标，查看连接状态

***

## 📚 使用示例

### 1. 创建 Selector 配置

```bash
curl -X POST http://localhost:8080/api/v1/selector-configs \
  -H "Content-Type: application/json" \
  -d '{
    "name": "京东商品价格",
    "domain": "item.jd.com",
    "selector": "#root > div > div.page-container > div.page-content > div.page-content-main > div.page-content-right > div > div.page-right-information > div.page-right-content > div.page-right-sticky-floor > div.page-right-price > div > span.product-price--main > span.product-price--value",
    "regex": "",
    "is_active": true
  }'
```

### 2. 批量抓取

```bash
curl -X POST http://localhost:8080/api/v1/batch-capture \
  -H "Content-Type: application/json" \
  -d '{
    "urls": [
      {"id": "1", "url": "https://item.jd.com/100247240555.html"},
      {"id": "2", "url": "https://item.jd.com/100114788370.html"},
      {"id": "3", "url": "https://item.jd.com/100231638932.html"}
    ],
    "timeout": 60000
  }'
```

**响应示例**：

```json
{
  "success": true,
  "results": [
    {
      "id": "1",
      "url": "https://item.jd.com/100247240555.html",
      "success": true,
      "data": {
        "id": "batch_1_549800",
        "html": "398.82",
        "text": "398.82",
        "timestamp": "2026-06-22T02:56:41.577Z"
      },
      "config_id": "bfd3d59b-317a-41db-b2bc-b26174a9b518",
      "config_name": "京东商品价格"
    }
  ],
  "total": 3,
  "success_count": 3,
  "failed_count": 0,
  "timestamp": "2026-06-22T10:56:42+08:00"
}
```

### 3. 手动测试抓取

在插件界面中：

1. 输入目标 URL
2. 输入 DOM 选择器（如 `.price`, `#main`, `h1`）
3. 点击"开始捕捉"
4. 查看抓取结果

***

## 📖 API 文档

启动后端服务后，访问以下地址查看完整的 API 文档：

```
http://localhost:8080/swagger/index.html
```

### 主要 API 端点

| 方法     | 端点                             | 描述     |
| ------ | ------------------------------ | ------ |
| GET    | `/api/v1/health`               | 健康检查   |
| GET    | `/api/v1/status`               | 系统状态   |
| POST   | `/api/v1/capture`              | 单次抓取   |
| POST   | `/api/v1/batch-capture`        | 批量抓取   |
| GET    | `/api/v1/selector-configs`     | 获取所有配置 |
| POST   | `/api/v1/selector-configs`     | 创建配置   |
| PUT    | `/api/v1/selector-configs/:id` | 更新配置   |
| DELETE | `/api/v1/selector-configs/:id` | 删除配置   |

***

## ⚙️ 配置说明

### 插件配置

在插件界面中可以配置以下参数：

- **WebSocket 地址**：后端 WebSocket 服务地址（默认：`ws://localhost:8080/ws`）
- **最小延迟**：随机延迟的最小值（毫秒，默认：0）
- **最大延迟**：随机延迟的最大值（毫秒，默认：2000）
- **超时时间**：抓取超时时间（毫秒，默认：30000）

### 随机延迟机制

当两个抓取请求到达时间间隔小于 1 秒时，系统会自动在配置的延迟范围内添加随机延迟，防止被网站识别为爬虫行为。

***

## 🛠️ 技术栈

### 后端

- **Gin** - 高性能 Web 框架
- **pgx** - PostgreSQL 驱动
- **Redis** - 缓存和会话管理
- **WebSocket** - 实时双向通信
- **Swagger** - API 文档生成
- **DDD** - 领域驱动设计架构

### 插件

- **Manifest V3** - Chrome 扩展最新规范
- **Chrome Scripting API** - 动态脚本注入
- **Chrome Tabs API** - 标签页管理
- **WebSocket** - 与后端实时通信

***

## 📦 项目结构

```
go_chrome/
├── backend/                    # Go 后端
│   ├── application/           # 应用层
│   │   └── batch_capture_app_service.go
│   ├── domain/                # 领域层
│   │   ├── selector_config.go
│   │   └── selector_config_service.go
│   ├── infrastructure/        # 基础设施层
│   │   ├── postgres_selector_config_repository.go
│   │   └── redis_repository.go
│   ├── handlers/              # HTTP 处理器
│   ├── pool/                  # 插件池管理
│   ├── scheduler/             # 任务调度器
│   ├── websocket/             # WebSocket 服务
│   └── main.go
├── extension/                 # Chrome 插件
│   ├── manifest.json
│   ├── background.js
│   ├── popup.html
│   ├── popup.js
│   ├── content.js
│   └── icons/
├── docker-compose.yml         # Docker 编排
└── README.md
```

***

## 🤝 贡献指南

欢迎贡献代码！请遵循以下步骤：

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 提交 Pull Request

***

## 📝 许可证

本项目采用 MIT 许可证 - 详见 [LICENSE](LICENSE) 文件

***

## 🙏 致谢

感谢所有为这个项目做出贡献的开发者！

***

<div align="center">

**如果这个项目对你有帮助，请给个 ⭐️ Star 支持一下！**

Made with ❤️ by \[Your Name]

</div>
