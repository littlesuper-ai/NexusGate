# 设备管理

## 概述

设备是 NexusGate 的核心管理对象。每台 OpenWrt 路由器通过 Agent 向服务端注册，并持续上报心跳。服务端维护设备清单、在线状态、分组标签，支持远程重启和指标查询。

## 源码文件

| 文件 | 说明 |
|------|------|
| `server/internal/model/device.go` | Device、DeviceMetrics 模型 |
| `server/internal/handler/device.go` | 设备 CRUD、注册、重启、指标、仪表板 |
| `web/src/views/Devices.vue` | 设备列表页面 |
| `web/src/views/DeviceDetail.vue` | 设备详情页面 (4 个 Tab) |
| `web/src/views/Dashboard.vue` | 仪表板 (含设备概览) |

## 数据模型

### Device

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | uint | PK | 主键 |
| name | string | not null | 设备名称 |
| mac | string | unique index, not null | MAC 地址 (注册唯一键) |
| ip_address | string | - | IP 地址 |
| model | string | - | 设备型号 |
| firmware | string | - | 固件版本 |
| status | string | default: unknown | online / offline / unknown |
| group | string | index | 设备分组 (如：总部、分支A) |
| tags | string | - | 逗号分隔标签 (core,vpn,iot) |
| uptime_secs | int64 | - | 运行时间（秒） |
| cpu_usage | float64 | - | CPU 使用率 % |
| mem_usage | float64 | - | 内存使用率 % |
| last_seen_at | *time | - | 最后在线时间 |
| registered_at | time | autoCreateTime | 首次注册时间 |
| created_at | time | auto | 创建时间 |
| updated_at | time | auto | 更新时间 |
| deleted_at | time | soft delete | 软删除 |

### DeviceMetrics

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | uint | PK | 主键 |
| device_id | uint | index | 关联设备 |
| cpu_usage | float64 | - | CPU % |
| mem_usage | float64 | - | 内存 % |
| mem_total | int64 | - | 总内存 (Agent 上报单位为 KB) |
| mem_free | int64 | - | 可用内存 (Agent 上报 MemAvailable, 单位 KB) |
| rx_bytes | int64 | - | 接收字节数 |
| tx_bytes | int64 | - | 发送字节数 |
| conntrack | int | - | 连接追踪数 |
| uptime_secs | int64 | - | 运行时间 |
| load_avg | string | - | 负载均值 |
| collected_at | time | - | 采集时间 |

## API 接口

### POST /api/v1/devices/register (公开)

设备自注册接口，Agent 启动时调用。按 MAC 地址 upsert。

```json
// 请求
{
  "name": "branch-gw-01",
  "mac": "AA:BB:CC:DD:EE:FF",
  "ip_address": "192.168.1.1",
  "model": "NanoPi R4S",
  "firmware": "23.05.5-r1"
}

// 响应 200
{ "id": 1, "name": "branch-gw-01", "mac": "AA:BB:CC:DD:EE:FF", ... }
```

注意：无论新建还是更新均返回 200 (StatusOK)，使用 FirstOrCreate 实现 upsert。

### GET /api/v1/devices

| 参数 | 类型 | 说明 |
|------|------|------|
| group | query, optional | 按分组过滤 |
| status | query, optional | 按状态过滤 (online/offline) |

返回设备数组。

### GET /api/v1/devices/:id

返回单个设备详情。

### PUT /api/v1/devices/:id

更新设备基本信息（name、group、tags）。

### DELETE /api/v1/devices/:id

软删除设备。

### POST /api/v1/devices/:id/reboot

触发远程重启。通过 MQTT 向 `nexusgate/devices/{mac}/command` 发送 `{"action":"reboot"}`。

### GET /api/v1/devices/:id/metrics

返回最近 100 条采集指标，按时间倒序。

### GET /api/v1/dashboard/summary

```json
{
  "total_devices": 9,
  "online_devices": 6,
  "offline_devices": 3
}
```

## 前端页面

### Devices.vue — 设备列表

- 搜索：按名称、MAC、IP 模糊搜索
- 过滤：在线/离线/全部
- 表格列：名称、MAC、IP、型号、固件、状态、分组、CPU、内存
- 操作：查看详情、重启、删除
- 点击行跳转到 DeviceDetail

### DeviceDetail.vue — 设备详情

左侧固定面板：
- 设备信息描述列表 (名称、MAC、IP、型号、固件、分组、标签、状态、运行时间、最后在线)
- 操作区：重启设备、下发配置、编辑信息

右侧标签页：

| Tab | 内容 |
|-----|------|
| 实时状态 | CPU/内存进度条 + CPU 趋势迷你折线图 |
| 历史指标 | 4 个 ECharts 图 (CPU、内存、网络流量、连接追踪) |
| 配置历史 | 配置记录表格 + 内容预览弹窗 |
| 升级记录 | 固件升级历史表格 |

### Dashboard.vue — 仪表板

- 4 个统计卡片：设备总数、在线、离线、模板数
- 设备表格：带 WebSocket 实时更新 (CPU/内存/状态)
- 审计日志时间线：最近 10 条操作
