# 监控与实时推送

## 概述

NexusGate 通过 MQTT 心跳收集设备指标，存储到 device_metrics 表供历史查询，同时通过 WebSocket 将实时状态推送到前端面板。

## 源码文件

| 文件 | 说明 |
|------|------|
| `server/internal/mqtt/client.go` | MQTT 订阅处理 + WebSocket 广播 |
| `server/internal/ws/hub.go` | WebSocket Hub 实现 |
| `server/internal/model/device.go` | DeviceMetrics 模型 |
| `web/src/composables/useWebSocket.ts` | 前端 WebSocket composable |
| `web/src/views/Monitoring.vue` | 监控中心页面 |
| `web/src/views/DeviceDetail.vue` | 实时状态 Tab + 历史指标 Tab |
| `web/src/views/Dashboard.vue` | 仪表板实时更新 |
| `web/src/views/Topology.vue` | 网络拓扑 |

## 指标采集流程

```
Agent (30s 间隔)
  └─ 采集: CPU, 内存, 网络, 连接追踪, 运行时间, 负载
  └─ MQTT Publish → nexusgate/devices/{mac}/status

Server MQTT Subscriber
  ├─ 更新 devices 表 (status=online, cpu/mem, last_seen_at)
  ├─ 按 MAC 查找 device_id
  ├─ 写入 device_metrics 表
  └─ WebSocket Hub.Broadcast("device_status", {...})
```

### 心跳 Payload 格式

```json
{
  "mac": "AA:BB:CC:DD:EE:FF",
  "cpu_usage": 23.5,
  "mem_usage": 45.2,
  "mem_total": 1073741824,
  "mem_free": 587890688,
  "rx_bytes": 123456789,
  "tx_bytes": 98765432,
  "conntrack": 1234,
  "uptime_secs": 86400,
  "load_avg": "0.15 0.20 0.25"
}
```

### Agent 采集方式

| 指标 | 数据源 |
|------|--------|
| CPU | `/proc/stat` (awk 内联采样 1 秒, 计算 user+system 差值) |
| 内存 | `/proc/meminfo` (MemTotal, MemAvailable, 单位 KB) |
| 网络流量 | `/sys/class/net/eth0/statistics/rx_bytes` + tx_bytes |
| 连接追踪 | `/proc/sys/net/netfilter/nf_conntrack_count` |
| 运行时间 | `/proc/uptime` |
| 负载均值 | `/proc/loadavg` |

## WebSocket Hub

### 架构

```
                  ┌─── Browser A ◄──┐
Server Hub ───────┤─── Browser B ◄──┤── Broadcast
                  └─── Browser C ◄──┘
```

### 实现 (`server/internal/ws/hub.go`)

| 组件 | 说明 |
|------|------|
| Hub.clients | `map[*websocket.Conn]struct{}` 连接池 |
| Hub.HandleWS | Gin handler, 升级 HTTP→WS, 注册连接 |
| Hub.Broadcast | 序列化 JSON 后向所有客户端发送 |
| Hub.ClientCount | 当前在线客户端数 |

连接管理：
- 升级时注册到 clients map
- 读循环检测断线 → 从 map 移除
- Ping/Pong 保活：30s 发 Ping, 60s 读超时
- 并发安全：`sync.RWMutex`

### WebSocket 消息格式

```json
{
  "type": "device_status",
  "data": {
    "mac": "AA:BB:CC:DD:EE:FF",
    "device_id": 1,
    "cpu_usage": 23.5,
    "mem_usage": 45.2,
    "rx_bytes": 123456789,
    "tx_bytes": 98765432,
    "conntrack": 1234,
    "uptime_secs": 86400,
    "load_avg": "0.15 0.20 0.25",
    "status": "online"
  },
  "timestamp": "2026-02-25T10:30:00+08:00"
}
```

### WS 端点

```
GET /ws → WebSocket Upgrade
```

- 不需要 JWT（便于简单连接）
- CORS: CheckOrigin 允许所有来源

## 前端 WebSocket Composable

### useWebSocket (`web/src/composables/useWebSocket.ts`)

```typescript
const { connected, lastMessage, on, off, disconnect } = useWebSocket()
```

| 返回值 | 类型 | 说明 |
|--------|------|------|
| connected | Ref\<boolean\> | 连接状态 |
| lastMessage | Ref\<WSMessage\> | 最后一条消息 |
| on(type, handler) | function | 注册消息处理器 |
| off(type, handler) | function | 取消注册 |
| disconnect() | function | 主动断开 |

特性：
- 自动连接（组件 mount 时）
- 断线自动重连（3 秒间隔）
- 组件 unmount 时自动断开
- 按消息 type 分发到对应 handler

### Dashboard 集成

```typescript
const { connected: wsConnected, on: wsOn } = useWebSocket()

wsOn('device_status', (data) => {
  // 按 mac 或 device_id 找到设备 → 更新 cpu_usage, mem_usage, status
})
```

- 设备表格左上角显示 "实时更新" / "离线" 状态标签

## 前端监控页面

### Monitoring.vue — 监控中心

- 设备选择器下拉框
- 选择设备后加载指标 → 渲染 4 个 ECharts 图表

| 图表 | 数据 | 颜色 |
|------|------|------|
| CPU 使用率 | cpu_usage % | #409EFF (蓝) |
| 内存使用率 | mem_usage % | #E6A23C (橙) |
| 网络流量 | rx_bytes/tx_bytes (MB) | #67C23A/#F56C6C (绿/红) |
| 连接追踪 | conntrack | #909399 (灰) |

### DeviceDetail.vue — 实时状态 Tab

- CPU 进度条 (带颜色阈值：>85%红, >60%橙, 其余绿)
- 内存进度条 (同上)
- CPU 趋势迷你折线图 (近 2 小时)

### DeviceDetail.vue — 历史指标 Tab

同 Monitoring.vue 的 4 个图表，数据来源 GET /api/v1/devices/:id/metrics。

### Topology.vue — 网络拓扑

- ECharts force-directed graph
- 节点类型：Internet(蓝方块) → 分组(橙菱形) → 设备(圆形)
- 设备颜色：online=绿, offline=红, unknown=灰
- 核心设备 (tags 含 core)：加大 + 橙色边框
- VPN 设备间：紫色虚线
- 交互：拖拽、缩放、adjacency 高亮
