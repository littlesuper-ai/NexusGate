# NexusGate 项目概览

## 简介

NexusGate 是一个基于 OpenWrt 的企业级路由器网关集中管理平台。支持对多台 OpenWrt 设备进行统一注册、配置下发、固件升级、实时监控和网络策略管理。

## 技术栈

| 层级 | 技术 |
|------|------|
| 后端 API | Go 1.23 + Gin 1.10 |
| 数据库 | PostgreSQL 16 + GORM |
| 消息总线 | MQTT (Mosquitto 2) via paho.mqtt.golang |
| 实时推送 | WebSocket (gorilla/websocket) |
| 前端 | Vue 3 + TypeScript + Element Plus + Vite |
| 图表 | ECharts |
| 认证 | JWT (golang-jwt/v5) + bcrypt |
| 设备端 | Shell 脚本 (OpenWrt procd 管理) |
| 固件构建 | OpenWrt ImageBuilder 23.05.5 |
| 部署 | Docker Compose (开发) / Ansible (设备) |

## 项目结构

```
NexusGate/
├── server/                    # Go 后端
│   ├── cmd/nexusgate/main.go  # 入口
│   ├── internal/
│   │   ├── config/            # 环境变量配置
│   │   ├── handler/           # HTTP 路由 & 业务逻辑
│   │   │   ├── middleware/    # JWT / RBAC 中间件
│   │   │   ├── router.go      # 路由注册
│   │   │   ├── auth.go        # 认证 & 用户
│   │   │   ├── device.go      # 设备管理
│   │   │   ├── config.go      # 配置模板 & 下发
│   │   │   ├── firewall.go    # 防火墙
│   │   │   ├── vpn.go         # WireGuard VPN
│   │   │   ├── network.go     # MWAN / DHCP / VLAN
│   │   │   ├── firmware.go    # 固件 & OTA
│   │   │   └── setting.go     # 系统设置
│   │   ├── model/             # 数据模型 (GORM)
│   │   ├── mqtt/              # MQTT 客户端 & 订阅
│   │   ├── ws/                # WebSocket Hub
│   │   ├── ssh/               # SSH 远程执行
│   │   └── store/             # 数据库初始化 & 迁移
│   ├── go.mod
│   └── Dockerfile
├── web/                       # Vue 3 前端
│   ├── src/
│   │   ├── api/index.ts       # Axios API 封装
│   │   ├── composables/       # useWebSocket
│   │   ├── layout/Layout.vue  # 主布局 & 侧边栏
│   │   ├── router/index.ts    # 前端路由
│   │   └── views/             # 16 个页面
│   ├── package.json
│   └── vite.config.ts
├── packages/nexusgate-agent/  # OpenWrt 设备端代理
├── firmware/                  # 固件构建脚本 & Profile
├── deploy/                    # Docker Compose / Ansible
└── specs/                     # 本文档
```

## 模块索引

| 文档 | 模块 |
|------|------|
| [01-architecture](01-architecture.md) | 系统架构与通信流程 |
| [02-auth](02-auth.md) | 认证授权 (JWT / RBAC / 用户 / 审计) |
| [03-devices](03-devices.md) | 设备管理与注册 |
| [04-config](04-config.md) | 配置模板与下发 |
| [05-firewall](05-firewall.md) | 防火墙 (Zone / Rule) |
| [06-vpn](06-vpn.md) | VPN (WireGuard) |
| [07-network](07-network.md) | 网络管理 (MWAN / DHCP / VLAN) |
| [08-firmware](08-firmware.md) | 固件管理与 OTA 升级 |
| [09-monitoring](09-monitoring.md) | 监控与 WebSocket 实时推送 |
| [10-settings](10-settings.md) | 系统设置 |
| [11-agent](11-agent.md) | OpenWrt 设备端代理 |
| [12-deployment](12-deployment.md) | 部署与运维 |
| [13-api-reference](13-api-reference.md) | API 接口完整参考 |

## 数据库表清单

共 19 张表（含 GORM 自动迁移）：

| 表 | 模型 | 所属模块 |
|----|------|----------|
| users | User | 认证 |
| audit_logs | AuditLog | 审计 |
| devices | Device | 设备 |
| device_metrics | DeviceMetrics | 监控 |
| config_templates | ConfigTemplate | 配置 |
| device_configs | DeviceConfig | 配置 |
| firewall_zones | FirewallZone | 防火墙 |
| firewall_rules | FirewallRule | 防火墙 |
| wire_guard_interfaces | WireGuardInterface | VPN |
| wire_guard_peers | WireGuardPeer | VPN |
| wan_interfaces | WANInterface | 网络 |
| mwan_policies | MWANPolicy | 网络 |
| mwan_rules | MWANRule | 网络 |
| dhcp_pools | DHCPPool | 网络 |
| static_leases | StaticLease | 网络 |
| vlans | VLAN | 网络 |
| firmwares | Firmware | 固件 |
| firmware_upgrades | FirmwareUpgrade | 固件 |
| system_settings | SystemSetting | 设置 |
