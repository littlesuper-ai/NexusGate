# 防火墙管理

## 概述

基于 OpenWrt firewall4 (nftables) 的区域防火墙管理。支持创建防火墙区域 (Zone) 和规则 (Rule)，并生成 UCI 配置下发到设备。

## 源码文件

| 文件 | 说明 |
|------|------|
| `server/internal/model/firewall.go` | FirewallZone、FirewallRule 模型 |
| `server/internal/handler/firewall.go` | Zone/Rule CRUD + UCI 生成 + 应用 |
| `web/src/views/Firewall.vue` | 防火墙管理页面 |

## 数据模型

### FirewallZone

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | uint | PK | 主键 |
| device_id | uint | index, not null | 所属设备 |
| name | string | not null | 区域名 (lan/wan/guest/iot) |
| input | string | default: REJECT | 入站策略: ACCEPT/REJECT/DROP |
| output | string | default: ACCEPT | 出站策略 |
| forward | string | default: REJECT | 转发策略 |
| masq | bool | - | 是否启用 NAT 伪装 |
| networks | string | - | 关联网络接口, 逗号分隔 |
| created_at | time | auto | 创建时间 |
| updated_at | time | auto | 更新时间 |
| deleted_at | time | soft delete | 软删除 |

### FirewallRule

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | uint | PK | 主键 |
| device_id | uint | index, not null | 所属设备 |
| name | string | not null | 规则名 |
| src | string | - | 源区域 (如 lan) |
| dest | string | - | 目标区域 (如 wan) |
| proto | string | - | 协议: tcp/udp/icmp/any |
| src_ip | string | - | 源 IP/CIDR |
| dest_ip | string | - | 目标 IP/CIDR |
| dest_port | string | - | 目标端口 (支持逗号/范围) |
| target | string | not null | 动作: ACCEPT/REJECT/DROP |
| enabled | bool | default: true | 是否启用 |
| position | int | default: 0 | 规则优先级 (越小越优先) |
| created_at | time | auto | 创建时间 |
| updated_at | time | auto | 更新时间 |
| deleted_at | time | soft delete | 软删除 |

## API 接口

### Zone CRUD

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/v1/firewall/zones?device_id= | 列表 (按设备过滤) |
| POST | /api/v1/firewall/zones | 创建 |
| PUT | /api/v1/firewall/zones/:id | 更新 |
| DELETE | /api/v1/firewall/zones/:id | 删除 |

### Rule CRUD

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/v1/firewall/rules?device_id= | 列表 (按 position, id 排序) |
| POST | /api/v1/firewall/rules | 创建 |
| PUT | /api/v1/firewall/rules/:id | 更新 |
| DELETE | /api/v1/firewall/rules/:id | 删除 |

### POST /api/v1/firewall/apply/:device_id

应用防火墙配置到设备。

流程：
1. 查找设备 → 获取 MAC
2. 查询该设备的所有 zones 和 rules (rules 只取 enabled=true)
3. 调用 `generateFirewallUCI()` 生成 UCI 配置
4. MQTT 发布到 `nexusgate/devices/{mac}/config`
5. 写入 device_configs 表
6. 返回 UCI 配置预览

## UCI 生成规则

```
# 全局默认 (始终生成)
config defaults
    option syn_flood '1'
    option input 'REJECT'
    option output 'ACCEPT'
    option forward 'REJECT'

# Zone 示例
config zone 'lan'
    option name 'lan'
    option input 'ACCEPT'
    option output 'ACCEPT'
    option forward 'REJECT'
    option masq '1'
    option mtu_fix '1'
    list network 'lan'

# Rule 示例
config rule 'allow_ssh'
    option name 'allow_ssh'
    option src 'lan'
    option dest 'wan'
    option proto 'tcp'
    option dest_port '22'
    option target 'ACCEPT'
```

生成逻辑 (`generateFirewallUCI`)：
1. 首先生成 `config defaults` 块 (syn_flood, input=REJECT, output=ACCEPT, forward=REJECT)
2. 遍历 zones → 生成 `config zone '{name}'` 块, masq=true 时附加 `mtu_fix`, networks 拆分为多条 `list network`
3. 遍历 rules (仅 enabled=true) → 生成 `config rule '{name}'` 块, 仅在字段非空时输出对应 option (proto 为 "any" 时跳过)

## 前端页面

### Firewall.vue

顶部：设备选择器 + 应用按钮

两个 Tab：

**区域 (Zones) Tab:**
- 表格列：名称、关联网络、入站策略、出站策略、转发策略、NAT
- 新建/编辑弹窗：名称、网络、策略选择器 (ACCEPT/REJECT/DROP)、NAT 开关

**规则 (Rules) Tab:**
- 表格列：优先级、名称、源区域、目标区域、协议、源IP、目标端口、动作、启用
- 新建/编辑弹窗：名称、源/目标区域(从已有zone中选择)、协议、IP、端口、动作、启用开关、优先级
