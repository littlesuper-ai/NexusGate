# API 接口参考

Base URL: `/api/v1`

认证: 除公开接口外，所有请求需携带 `Authorization: Bearer <JWT>` Header。

---

## 公开接口

### POST /auth/login

登录获取 JWT Token。

```
Request:  { "username": "admin", "password": "admin123" }
Response: { "token": "eyJ...", "user": { "id": 1, "username": "admin", "role": "admin" } }
```

### POST /devices/register

设备自注册（Agent 调用），按 MAC upsert。

```
Request:  { "name": "gw-01", "mac": "AA:BB:CC:DD:EE:FF", "ip_address": "192.168.1.1", "model": "NanoPi R4S", "firmware": "23.05.5" }
Response: 200 { Device object }
```

注意：使用 FirstOrCreate 实现 upsert，新建和更新均返回 200。

---

## WebSocket

### GET /ws

WebSocket 升级端点（无需 JWT）。

消息格式:
```json
{ "type": "device_status", "data": { ... }, "timestamp": "2026-02-25T10:30:00Z" }
```

---

## 设备管理

| 方法 | 路径 | 说明 | Query 参数 |
|------|------|------|-----------|
| GET | /devices | 设备列表 | group, status |
| GET | /devices/:id | 设备详情 | - |
| PUT | /devices/:id | 更新设备 | - |
| DELETE | /devices/:id | 删除设备 | - |
| POST | /devices/:id/reboot | 远程重启 | - |
| GET | /devices/:id/metrics | 设备指标 | - |
| GET | /dashboard/summary | 仪表板统计 | - |

---

## 配置管理

| 方法 | 路径 | 说明 | Query 参数 |
|------|------|------|-----------|
| GET | /templates | 模板列表 | category |
| POST | /templates | 创建模板 | - |
| PUT | /templates/:id | 更新模板 | - |
| DELETE | /templates/:id | 删除模板 | - |
| POST | /devices/:id/config/push | 下发配置 | - |
| GET | /devices/:id/config/history | 配置历史 | - |

**POST /devices/:id/config/push:**
```json
{ "template_id": 3 }
// 或
{ "content": "config interface 'lan'\n..." }
```

---

## 用户管理 (admin)

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /users | 用户列表 |
| POST | /users | 创建用户 |
| DELETE | /users/:id | 删除用户 |
| GET | /audit-logs | 审计日志 |

**POST /users:**
```json
{ "username": "ops1", "password": "securepass", "role": "operator", "email": "ops@ex.com" }
```

---

## 防火墙

| 方法 | 路径 | 说明 | Query 参数 |
|------|------|------|-----------|
| GET | /firewall/zones | Zone 列表 | device_id |
| POST | /firewall/zones | 创建 Zone | - |
| PUT | /firewall/zones/:id | 更新 Zone | - |
| DELETE | /firewall/zones/:id | 删除 Zone | - |
| GET | /firewall/rules | Rule 列表 | device_id |
| POST | /firewall/rules | 创建 Rule | - |
| PUT | /firewall/rules/:id | 更新 Rule | - |
| DELETE | /firewall/rules/:id | 删除 Rule | - |
| POST | /firewall/apply/:device_id | 应用到设备 | - |

**FirewallZone:**
```json
{ "device_id": 1, "name": "lan", "input": "ACCEPT", "output": "ACCEPT", "forward": "REJECT", "masq": false, "networks": "lan" }
```

**FirewallRule:**
```json
{ "device_id": 1, "name": "allow_ssh", "src": "lan", "dest": "wan", "proto": "tcp", "dest_port": "22", "target": "ACCEPT", "enabled": true, "position": 10 }
```

---

## VPN (WireGuard)

| 方法 | 路径 | 说明 | Query 参数 |
|------|------|------|-----------|
| GET | /vpn/interfaces | 接口列表 | device_id |
| POST | /vpn/interfaces | 创建接口 | - |
| PUT | /vpn/interfaces/:id | 更新接口 | - |
| DELETE | /vpn/interfaces/:id | 删除接口 (含 peers) | - |
| GET | /vpn/peers | Peer 列表 | interface_id |
| POST | /vpn/peers | 创建 Peer | - |
| PUT | /vpn/peers/:id | 更新 Peer | - |
| DELETE | /vpn/peers/:id | 删除 Peer | - |
| POST | /vpn/apply/:device_id | 应用到设备 | - |

**WireGuardInterface:**
```json
{ "device_id": 1, "name": "wg0", "private_key": "xxx", "public_key": "yyy", "address": "10.99.0.1/24", "listen_port": 51820, "enabled": true }
```

**WireGuardPeer:**
```json
{ "interface_id": 1, "description": "branch-a", "public_key": "zzz", "allowed_ips": "10.99.1.0/24,192.168.10.0/24", "endpoint": "203.0.113.1:51820", "keepalive": 25, "enabled": true }
```

---

## 网络管理 - Multi-WAN

| 方法 | 路径 | 说明 | Query 参数 |
|------|------|------|-----------|
| GET | /network/wan | WAN 接口列表 | device_id |
| POST | /network/wan | 添加 WAN 接口 | - |
| DELETE | /network/wan/:id | 删除 WAN 接口 | - |
| GET | /network/mwan/policies | 策略列表 | device_id |
| POST | /network/mwan/policies | 添加策略 | - |
| DELETE | /network/mwan/policies/:id | 删除策略 | - |
| GET | /network/mwan/rules | 规则列表 | device_id |
| POST | /network/mwan/rules | 添加规则 | - |
| DELETE | /network/mwan/rules/:id | 删除规则 | - |
| POST | /network/mwan/apply/:device_id | 应用到设备 | - |

**WANInterface:**
```json
{ "device_id": 1, "name": "wan1", "interface": "eth1", "enabled": true, "weight": 1, "track_ips": "8.8.8.8,114.114.114.114", "reliability": 2, "interval": 5, "down": 3, "up": 3 }
```

**MWANPolicy:**
```json
{ "device_id": 1, "name": "balanced", "members": "[{\"iface\":\"wan1\",\"metric\":1,\"weight\":1}]", "last_resort": "default" }
```

**MWANRule:**
```json
{ "device_id": 1, "name": "video", "src_ip": "", "dest_ip": "0.0.0.0/0", "proto": "tcp", "dest_port": "80,443", "policy": "balanced", "enabled": true, "position": 10 }
```

---

## 网络管理 - DHCP

| 方法 | 路径 | 说明 | Query 参数 |
|------|------|------|-----------|
| GET | /network/dhcp/pools | 地址池列表 | device_id |
| POST | /network/dhcp/pools | 添加地址池 | - |
| DELETE | /network/dhcp/pools/:id | 删除地址池 | - |
| GET | /network/dhcp/leases | 静态绑定列表 | device_id |
| POST | /network/dhcp/leases | 添加静态绑定 | - |
| DELETE | /network/dhcp/leases/:id | 删除静态绑定 | - |

**DHCPPool:**
```json
{ "device_id": 1, "interface": "lan", "start": 100, "limit": 150, "lease_time": "12h", "dns": "8.8.8.8,223.5.5.5", "gateway": "192.168.1.1", "enabled": true }
```

**StaticLease:**
```json
{ "device_id": 1, "name": "printer-1", "mac": "AA:BB:CC:DD:EE:FF", "ip": "192.168.1.50" }
```

---

## 网络管理 - VLAN

| 方法 | 路径 | 说明 | Query 参数 |
|------|------|------|-----------|
| GET | /network/vlans | VLAN 列表 | device_id |
| POST | /network/vlans | 创建 VLAN | - |
| PUT | /network/vlans/:id | 更新 VLAN | - |
| DELETE | /network/vlans/:id | 删除 VLAN | - |

**VLAN:**
```json
{ "device_id": 1, "vid": 10, "name": "office", "interface": "br-lan.10", "ip_addr": "10.0.10.1", "netmask": "255.255.255.0", "isolated": false }
```

---

## 固件管理

| 方法 | 路径 | 说明 | Query 参数 |
|------|------|------|-----------|
| GET | /firmware | 固件列表 | target |
| POST | /firmware/upload | 上传固件 (multipart) | - |
| GET | /firmware/download/:filename | 下载固件 | - |
| DELETE | /firmware/:id | 删除固件 | - |
| POST | /firmware/:id/stable | 标记稳定版 | - |
| POST | /firmware/upgrade | 单设备升级 | - |
| POST | /firmware/upgrade/batch | 批量升级 | - |
| GET | /firmware/upgrades | 升级记录 | device_id |

**Upload (multipart form-data):**
```
file: <binary>
version: "23.05.5-r2"
target: "x86-64"
changelog: "Bug fixes..."
```

**PushUpgrade:**
```json
{ "device_id": 1, "firmware_id": 3 }
```

**BatchUpgrade:**
```json
{ "firmware_id": 3, "group": "分支", "model": "NanoPi R4S" }
```

---

## 系统设置

| 方法 | 路径 | 说明 | Query 参数 |
|------|------|------|-----------|
| GET | /settings | 设置列表 | category |
| GET | /settings/:key | 查询设置 | - |
| POST | /settings | Upsert 设置 | - |
| POST | /settings/batch | 批量 Upsert | - |
| DELETE | /settings/:key | 删除设置 | - |

**Upsert:**
```json
{ "key": "offline_threshold", "value": "120", "category": "general" }
```

**Batch:**
```json
[
  { "key": "system_name", "value": "NexusGate", "category": "general" },
  { "key": "alert_cpu_threshold", "value": "85", "category": "alert" }
]
```

---

## 错误响应格式

所有错误返回统一格式：

```json
{ "error": "error description" }
```

| HTTP 状态码 | 含义 |
|-------------|------|
| 400 | 请求参数错误 |
| 401 | 未认证 / Token 过期 |
| 403 | 权限不足 |
| 404 | 资源不存在 |
| 500 | 服务端错误 |

## 路由总计

| 类别 | 端点数 |
|------|--------|
| 公开接口 | 2 |
| WebSocket | 1 |
| 设备管理 | 7 |
| 配置管理 | 6 |
| 用户管理 (admin) | 4 |
| 防火墙 | 9 |
| VPN | 9 |
| Multi-WAN | 10 |
| DHCP | 6 |
| VLAN | 4 |
| 固件管理 | 8 |
| 系统设置 | 5 |
| **总计** | **71** |
