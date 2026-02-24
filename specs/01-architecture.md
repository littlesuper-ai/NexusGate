# 系统架构

## 整体架构

```
┌───────────────┐       ┌───────────────┐       ┌───────────────┐
│  Vue 3 SPA    │◄─────►│  Go REST API  │◄─────►│  PostgreSQL   │
│  (前端面板)    │ HTTP  │  (Gin 框架)    │ GORM  │  (数据存储)    │
│  :3000        │  WS   │  :8080        │       │  :5432        │
└───────────────┘       └───────┬───────┘       └───────────────┘
                                │
                    ┌───────────┼───────────┐
                    │     MQTT Pub/Sub      │
                    │                       │
              ┌─────▼──────┐         ┌──────▼─────┐
              │ Mosquitto  │         │ Mosquitto  │
              │ Broker     │◄───────►│ (同一实例)  │
              │ :1883      │         │            │
              └─────┬──────┘         └──────┬─────┘
                    │                       │
         ┌──────────┼──────────┐           │
         │          │          │           │
    ┌────▼───┐ ┌────▼───┐ ┌───▼────┐     │
    │ Agent  │ │ Agent  │ │ Agent  │     │
    │ 设备 A │ │ 设备 B │ │ 设备 C │ ... │
    │ OpenWrt│ │ OpenWrt│ │ OpenWrt│     │
    └────────┘ └────────┘ └────────┘
```

## 通信流程

### 设备注册

```
Agent ──HTTP POST──► Server /api/v1/devices/register
       { mac, name, ip_address, model, firmware }
       ◄── 200 { device object }
```

- 首次注册创建新记录，后续按 MAC 地址 upsert

### 心跳上报 (MQTT)

```
Agent ──MQTT Publish──► nexusgate/devices/{mac}/status
       { cpu_usage, mem_usage, mem_total, mem_free,
         rx_bytes, tx_bytes, conntrack, uptime_secs, load_avg }

Server 订阅 nexusgate/devices/+/status
  ├─ 更新 devices 表 (status, cpu_usage, mem_usage, last_seen_at)
  ├─ 写入 device_metrics 表
  └─ 广播到 WebSocket 客户端 (type: device_status)
```

- 默认间隔：30 秒
- QoS: 1 (至少一次)

### 命令下发 (MQTT)

```
Server ──MQTT Publish──► nexusgate/devices/{mac}/command
       { "action": "reboot" | "upgrade", ... }

Agent 订阅 nexusgate/devices/{mac}/command
  ├─ reboot → 执行 reboot
  ├─ upgrade → wget + sysupgrade
  └─ apply_config → uci import
```

### 配置下发 (MQTT)

```
Server ──MQTT Publish──► nexusgate/devices/{mac}/config
       UCI 配置文本

Agent 订阅 nexusgate/devices/{mac}/config
  └─ uci import → uci commit → /etc/init.d/network reload
```

### WebSocket 实时推送

```
浏览器 ──WS Upgrade──► Server /ws
       ◄── { type: "device_status", data: {...}, timestamp }

Hub 模式：
  - 所有 WS 客户端注册到 Hub
  - MQTT 消息到达 → Hub.Broadcast() → 所有客户端
  - Ping/Pong 保活 (30s ping, 60s 超时)
  - 断线自动重连 (3s 间隔)
```

## 关键依赖

| 包 | 版本 | 用途 |
|----|------|------|
| github.com/gin-gonic/gin | v1.10.0 | HTTP 框架 |
| github.com/gin-contrib/cors | v1.7.3 | 跨域支持 |
| gorm.io/gorm | v1.25.12 | ORM |
| gorm.io/driver/postgres | v1.5.11 | PostgreSQL 驱动 |
| github.com/eclipse/paho.mqtt.golang | v1.5.0 | MQTT 客户端 |
| github.com/gorilla/websocket | v1.5.3 | WebSocket (go.mod 标记为 indirect) |
| github.com/golang-jwt/jwt/v5 | v5.2.1 | JWT |
| golang.org/x/crypto | v0.31.0 | bcrypt |

## 前端依赖

| 包 | 用途 |
|----|------|
| vue@3 | UI 框架 |
| vue-router@4 | 路由 |
| pinia | 状态管理 |
| axios | HTTP 客户端 |
| element-plus | UI 组件库 |
| echarts | 图表可视化 |
| dayjs | 时间处理 |

## 环境变量

| 变量 | 默认值 | 说明 |
|------|--------|------|
| LISTEN_ADDR | :8080 | 服务监听地址 |
| DB_HOST | localhost | PostgreSQL 主机 |
| DB_PORT | 5432 | PostgreSQL 端口 |
| DB_USER | nexusgate | 数据库用户名 |
| DB_PASSWORD | nexusgate | 数据库密码 |
| DB_NAME | nexusgate | 数据库名 |
| MQTT_BROKER | tcp://localhost:1883 | MQTT Broker 地址 |
| JWT_SECRET | (必填, 无默认值) | JWT 签名密钥 (为空时服务拒绝启动) |
