# 网络管理 (Multi-WAN / DHCP / VLAN)

## 概述

网络管理模块涵盖三大功能：多线负载均衡 (mwan3)、DHCP 服务管理、VLAN 隔离配置。均采用按设备管理的模式，支持 UCI 配置生成和 MQTT 下发。

## 源码文件

| 文件 | 说明 |
|------|------|
| `server/internal/model/network.go` | 6 个模型：WAN/Policy/Rule/DHCP/Lease/VLAN |
| `server/internal/handler/network.go` | 全部 CRUD + ApplyMWAN + UCI 生成 |
| `web/src/views/MWAN.vue` | 多线负载管理页面 |
| `web/src/views/DHCP.vue` | DHCP 管理页面 |
| `web/src/views/VLAN.vue` | VLAN 管理页面 |

---

## 一、Multi-WAN (mwan3)

### 数据模型

#### WANInterface

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | uint | PK | 主键 |
| device_id | uint | index, not null | 所属设备 |
| name | string | not null | WAN 名称 (wan1, wan2) |
| interface | string | not null | 物理接口 (eth1, pppoe-wan) |
| enabled | bool | default: true | 是否启用 |
| weight | int | default: 1 | 负载权重 |
| track_ips | string | - | 探测 IP, 逗号分隔 |
| reliability | int | default: 2 | 可靠性阈值 (需几个探测成功) |
| interval | int | default: 5 | 探测间隔 (秒) |
| down | int | default: 3 | 连续失败次数判定下线 |
| up | int | default: 3 | 连续成功次数判定上线 |

#### MWANPolicy

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | uint | PK | 主键 |
| device_id | uint | index, not null | 所属设备 |
| name | string | not null | 策略名 (balanced, failover) |
| members | string | - | JSON 成员: `[{"iface":"wan1","metric":1,"weight":1}]` |
| last_resort | string | default: default | 兜底策略: default / unreachable |

#### MWANRule

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | uint | PK | 主键 |
| device_id | uint | index, not null | 所属设备 |
| name | string | not null | 规则名 |
| src_ip | string | - | 源 IP/CIDR |
| dest_ip | string | - | 目标 IP/CIDR |
| proto | string | - | 协议: tcp/udp/all |
| src_port | string | - | 源端口 |
| dest_port | string | - | 目标端口 |
| policy | string | not null | 使用的策略名 |
| enabled | bool | default: true | 是否启用 |
| position | int | default: 0 | 优先级 |

### API 接口

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/v1/network/wan?device_id= | WAN 接口列表 |
| POST | /api/v1/network/wan | 添加 WAN 接口 |
| DELETE | /api/v1/network/wan/:id | 删除 WAN 接口 |
| GET | /api/v1/network/mwan/policies?device_id= | 策略列表 |
| POST | /api/v1/network/mwan/policies | 添加策略 |
| DELETE | /api/v1/network/mwan/policies/:id | 删除策略 |
| GET | /api/v1/network/mwan/rules?device_id= | 规则列表 |
| POST | /api/v1/network/mwan/rules | 添加规则 |
| DELETE | /api/v1/network/mwan/rules/:id | 删除规则 |
| POST | /api/v1/network/mwan/apply/:device_id | 应用配置到设备 |

### ApplyMWAN 流程

1. 查找设备 → 获取 MAC
2. 查询 enabled=true 的 WAN 接口
3. 查询所有策略
4. 查询 enabled=true 的规则 (按 position, id 排序)
5. 调用 `generateMWANUCI()` 生成配置
6. MQTT 发布到 `nexusgate/devices/{mac}/config`
7. 写入 device_configs 表

### UCI 生成示例

```
config interface 'wan1'
    option enabled '1'
    list track_ip '8.8.8.8'
    list track_ip '114.114.114.114'
    option reliability '2'
    option count '3'
    option timeout '3'
    option interval '5'
    option down '3'
    option up '3'

config policy 'balanced'
    option last_resort 'default'
    # members: [{"iface":"wan1","metric":1,"weight":1}]

config rule 'video_traffic'
    option dest_ip '0.0.0.0/0'
    option proto 'tcp'
    option dest_port '80,443'
    option use_policy 'balanced'
```

### 前端页面 (MWAN.vue)

顶部：设备选择器 + 应用到设备按钮

三个 Tab：
- **WAN 接口**: 名称、接口、启用、权重、探测IP、可靠性、间隔、Down/Up 阈值
- **负载策略**: 策略名、成员 JSON、兜底策略
- **路由规则**: 规则名、源IP、目标IP、协议、端口、策略、启用、优先级

---

## 二、DHCP 管理

### 数据模型

#### DHCPPool

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | uint | PK | 主键 |
| device_id | uint | index, not null | 所属设备 |
| interface | string | not null | 接口名 (lan/guest/iot) |
| start | int | default: 100 | 起始地址偏移 |
| limit | int | default: 150 | 地址池大小 |
| lease_time | string | default: 12h | 租期 |
| dns | string | - | DNS 服务器, 逗号分隔 |
| gateway | string | - | 默认网关 |
| enabled | bool | default: true | 是否启用 |

#### StaticLease

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | uint | PK | 主键 |
| device_id | uint | index, not null | 所属设备 |
| name | string | not null | 主机名 |
| mac | string | not null | MAC 地址 |
| ip | string | not null | 绑定 IP |
| created_at | time | auto | 创建时间 |

### API 接口

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/v1/network/dhcp/pools?device_id= | 地址池列表 |
| POST | /api/v1/network/dhcp/pools | 添加地址池 |
| DELETE | /api/v1/network/dhcp/pools/:id | 删除地址池 |
| GET | /api/v1/network/dhcp/leases?device_id= | 静态绑定列表 |
| POST | /api/v1/network/dhcp/leases | 添加静态绑定 |
| DELETE | /api/v1/network/dhcp/leases/:id | 删除静态绑定 |

### 前端页面 (DHCP.vue)

两个 Tab：
- **地址池**: 接口、起始、数量、租期、DNS、网关、启用
- **静态绑定**: 主机名、MAC 地址 (等宽字体)、IP 地址

---

## 三、VLAN 管理

### 数据模型

#### VLAN

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | uint | PK | 主键 |
| device_id | uint | index, not null | 所属设备 |
| vid | int | not null | 802.1Q VLAN ID (1-4094) |
| name | string | not null | 名称 (office/server/guest) |
| interface | string | - | 接口名 (br-lan.10) |
| ip_addr | string | - | IP 地址 (10.0.10.1) |
| netmask | string | default: 255.255.255.0 | 子网掩码 |
| isolated | bool | default: false | 是否禁用 VLAN 间路由 |

### API 接口

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/v1/network/vlans?device_id= | 列表 (按 vid 排序) |
| POST | /api/v1/network/vlans | 创建 |
| PUT | /api/v1/network/vlans/:id | 更新 |
| DELETE | /api/v1/network/vlans/:id | 删除 |

### 前端页面 (VLAN.vue)

单个表格：
- 列：VLAN ID、名称、接口、IP 地址、子网掩码、隔离状态
- 操作：编辑、删除
- 编辑弹窗支持修改全部字段
