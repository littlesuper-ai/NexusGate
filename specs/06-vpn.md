# VPN 管理 (WireGuard)

## 概述

支持通过 WireGuard 建立站点间 VPN 隧道。管理 WireGuard 接口和 Peer 配置，生成 UCI 格式并下发到 OpenWrt 设备。

## 源码文件

| 文件 | 说明 |
|------|------|
| `server/internal/model/vpn.go` | WireGuardInterface、WireGuardPeer 模型 |
| `server/internal/handler/vpn.go` | Interface/Peer CRUD + UCI 生成 + 应用 |
| `web/src/views/VPN.vue` | VPN 管理页面 |

## 数据模型

### WireGuardInterface

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | uint | PK | 主键 |
| device_id | uint | index, not null | 所属设备 |
| name | string | not null | 接口名 (wg0, wg1) |
| private_key | string | json:"-", not null | 私钥 (不输出到 JSON) |
| public_key | string | - | 公钥 |
| address | string | - | 地址段 (10.99.0.1/24) |
| listen_port | int | default: 51820 | 监听端口 |
| enabled | bool | default: true | 是否启用 |
| created_at | time | auto | 创建时间 |
| updated_at | time | auto | 更新时间 |
| deleted_at | time | soft delete | 软删除 |

### WireGuardPeer

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | uint | PK | 主键 |
| interface_id | uint | index, not null | 所属 WireGuard 接口 |
| description | string | - | 描述 (如：总部网关) |
| public_key | string | not null | 对端公钥 |
| preshared_key | string | json:"-" | 预共享密钥 |
| allowed_ips | string | - | 允许的 IP 段, 逗号分隔 CIDR |
| endpoint | string | - | 对端地址 (host:port) |
| keepalive | int | default: 25 | 保活间隔 (秒) |
| enabled | bool | default: true | 是否启用 |
| last_handshake | *time | - | 最后握手时间 |
| tx_bytes | int64 | - | 发送字节数 |
| rx_bytes | int64 | - | 接收字节数 |
| created_at | time | auto | 创建时间 |
| updated_at | time | auto | 更新时间 |
| deleted_at | time | soft delete | 软删除 |

## API 接口

### Interface CRUD

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/v1/vpn/interfaces?device_id= | 列表 |
| POST | /api/v1/vpn/interfaces | 创建 |
| PUT | /api/v1/vpn/interfaces/:id | 更新 |
| DELETE | /api/v1/vpn/interfaces/:id | 删除 (级联删除 peers) |

### Peer CRUD

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/v1/vpn/peers?interface_id= | 列表 |
| POST | /api/v1/vpn/peers | 创建 |
| PUT | /api/v1/vpn/peers/:id | 更新 |
| DELETE | /api/v1/vpn/peers/:id | 删除 |

### POST /api/v1/vpn/apply/:device_id

应用 WireGuard VPN 配置到设备。

流程：
1. 查找设备 → 获取 MAC
2. 查询该设备所有 WireGuard 接口 (enabled=true) 及其 Peer
3. 调用 `generateWireGuardUCI()` 生成配置
4. MQTT 发布配置
5. 写入 device_configs 表

## UCI 生成规则

```
# Interface
config interface 'wg0'
    option proto 'wireguard'
    option private_key 'xxxxxxx'
    list addresses '10.99.0.1/24'
    option listen_port '51820'

# Peer
config wireguard_wg0
    option description '分支A网关'
    option public_key 'yyyyyyy'
    list allowed_ips '10.99.1.0/24'
    list allowed_ips '192.168.10.0/24'
    option endpoint_host '203.0.113.1:51820'
    option persistent_keepalive '25'
```

生成逻辑 (`generateWireGuardUCI`)：
- 遍历 interfaces (enabled=true) → 生成 `config interface '{name}'` 块
- 遍历每个接口的 peers (enabled=true, 按 interface_id 过滤) → 生成 `config wireguard_{name}` 块
- allowed_ips 按逗号拆分为多条 `list allowed_ips`
- endpoint 直接作为 `endpoint_host` 输出 (不拆分 host/port)
- preshared_key 目前不输出到 UCI (模型中存储但未生成)

## 前端页面

### VPN.vue

布局：设备选择器 + 应用按钮

双列布局：
- 左侧：WireGuard 接口列表
  - 表格列：名称、地址段、监听端口、启用状态
  - 新建/编辑弹窗
- 右侧：选中接口的 Peer 列表
  - 表格列：描述、公钥、允许 IP、端点、保活、启用、流量统计
  - 新建/编辑弹窗

## 网络拓扑集成

Topology.vue 中，带 `vpn` 标签的设备之间会绘制紫色虚线，表示 VPN 隧道连接：
- 样式：`{ color: '#9b59b6', type: 'dashed', curveness: 0.3 }`
- 标签：显示 "VPN"
