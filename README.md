# NexusGate

基于 OpenWrt 的企业级路由器网关集中管理平台。支持对多台 OpenWrt 设备进行统一注册、配置下发、固件升级、实时监控和网络策略管理。

## 功能特性

- **设备管理** — 设备注册、状态追踪、分组标签、批量操作
- **配置下发** — 配置模板系统，通过 MQTT 推送至设备，支持 ACK 确认
- **防火墙管理** — Zone / Rule 管理，一键下发防火墙策略
- **VPN 管理** — WireGuard 隧道配置与对端管理
- **网络管理** — 多线负载(MWAN)、DHCP 地址池与静态租约、VLAN
- **固件升级** — OTA 升级，SHA256 校验，批量与自动升级策略
- **实时监控** — WebSocket 实时推送 CPU / 内存 / 网络指标，ECharts 图表展示
- **告警系统** — 告警创建、解决、Webhook 通知
- **审计日志** — 操作审计追踪，支持筛选与分页
- **认证授权** — JWT 令牌认证，RBAC 三级角色 (admin / operator / viewer)
- **安全加固** — 输入校验、密码复杂度、速率限制、请求追踪

## 技术栈

| 层级 | 技术 |
|------|------|
| 后端 API | Go 1.23 + Gin |
| 数据库 | PostgreSQL 16 + GORM |
| 消息总线 | MQTT (Mosquitto 2) |
| 实时推送 | WebSocket |
| 前端 | Vue 3 + TypeScript + Element Plus + Vite |
| 图表 | ECharts |
| 设备端 | Shell 脚本 (OpenWrt procd) |
| 固件构建 | OpenWrt ImageBuilder 23.05.5 |
| 部署 | Docker Compose / Ansible |
| 监控 | Prometheus + Grafana |

## 项目结构

```
NexusGate/
├── server/                    # Go 后端 API
│   ├── cmd/nexusgate/         # 入口
│   └── internal/
│       ├── config/            # 环境变量配置
│       ├── handler/           # HTTP 路由 & 业务逻辑
│       │   └── middleware/    # JWT / RBAC / 限流
│       ├── model/             # 数据模型 (GORM)
│       ├── mqtt/              # MQTT 客户端
│       ├── ws/                # WebSocket Hub
│       ├── jobs/              # 后台任务
│       └── store/             # 数据库初始化 & 迁移
├── web/                       # Vue 3 前端 SPA
│   └── src/
│       ├── api/               # Axios API 封装
│       ├── composables/       # WebSocket 组合式函数
│       ├── layout/            # 主布局
│       ├── router/            # 前端路由
│       └── views/             # 页面组件
├── packages/nexusgate-agent/  # OpenWrt 设备端代理脚本
├── firmware/                  # 固件构建脚本 & Profile
├── deploy/                    # Docker Compose / Ansible / Mosquitto / Prometheus / Grafana
└── specs/                     # 设计文档 (14 篇)
```

## 快速开始

### 环境要求

- Docker & Docker Compose
- (开发) Go 1.23+, Node.js 22+

### 使用 Docker Compose 启动

```bash
cd deploy

# 可选：设置环境变量
export JWT_SECRET="your-secure-secret"
export ADMIN_PASSWORD="your-admin-password"

docker compose up -d
```

启动后访问：

| 服务 | 地址 |
|------|------|
| 前端 | http://localhost:3000 |
| API | http://localhost:8080 |
| Prometheus | http://localhost:9090 |
| Grafana | http://localhost:3001 |

首次启动会自动创建 admin 用户。如未设置 `ADMIN_PASSWORD`，随机密码将打印在 server 容器日志中：

```bash
docker compose logs server | grep "password"
```

### 本地开发

**后端：**

```bash
cd server

# 设置环境变量
export JWT_SECRET="dev-secret"
export DB_HOST=localhost
export MQTT_BROKER="tcp://localhost:1883"

go run ./cmd/nexusgate
```

**前端：**

```bash
cd web
npm install
npm run dev
```

前端开发服务器运行在 http://localhost:3000，API 请求自动代理到 http://localhost:8080。

## 测试

**前端测试 (Vitest)：**

```bash
cd web
npm test            # 运行一次
npm run test:watch  # 监听模式
```

**后端测试：**

```bash
cd server
go test ./... -v
```

## 环境变量

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `JWT_SECRET` | (必填) | JWT 签名密钥 |
| `LISTEN_ADDR` | `:8080` | 后端监听地址 |
| `DB_HOST` | `localhost` | PostgreSQL 主机 |
| `DB_PORT` | `5432` | PostgreSQL 端口 |
| `DB_USER` | `nexusgate` | 数据库用户 |
| `DB_PASSWORD` | `nexusgate` | 数据库密码 |
| `DB_NAME` | `nexusgate` | 数据库名称 |
| `DB_SSLMODE` | `disable` | SSL 模式 |
| `MQTT_BROKER` | `tcp://localhost:1883` | MQTT 代理地址 |
| `CORS_ORIGINS` | `*` | 允许的跨域源 (逗号分隔) |
| `ADMIN_PASSWORD` | (随机生成) | 初始 admin 用户密码 |

## 设计文档

详细的系统设计文档位于 [`specs/`](specs/) 目录，包含架构设计、API 参考、各模块详细说明等 14 篇文档。

## License

MIT
