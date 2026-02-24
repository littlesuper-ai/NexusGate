# 配置模板与下发

## 概述

NexusGate 支持维护 UCI 配置模板，并通过 MQTT 将配置推送到目标设备。每次下发记录在 device_configs 表中，可追溯配置变更历史。

## 源码文件

| 文件 | 说明 |
|------|------|
| `server/internal/model/config_template.go` | ConfigTemplate、DeviceConfig 模型 |
| `server/internal/handler/config.go` | 模板 CRUD、配置下发、历史查询 |
| `web/src/views/Templates.vue` | 配置模板管理页面 |
| `web/src/views/DeviceDetail.vue` | 配置历史 Tab + 下发弹窗 |

## 数据模型

### ConfigTemplate

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | uint | PK | 主键 |
| name | string | unique index, not null | 模板名称 |
| description | string | - | 描述 |
| category | string | index | 分类: network/firewall/vpn/qos/system |
| content | string | text, not null | UCI 配置内容 |
| version | int | default: 1 | 版本号 (更新时自增) |
| created_at | time | auto | 创建时间 |
| updated_at | time | auto | 更新时间 |
| deleted_at | time | soft delete | 软删除 |

### DeviceConfig

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | uint | PK | 主键 |
| device_id | uint | index, not null | 目标设备 |
| template_id | *uint | - | 来源模板 (可为空) |
| content | string | text, not null | 实际下发的配置内容 |
| version | int | default: 1 | 版本号 |
| applied_at | *time | - | 应用成功时间 |
| status | string | default: pending | pending / applied / failed |
| created_at | time | auto | 下发时间 |

## API 接口

### GET /api/v1/templates

| 参数 | 说明 |
|------|------|
| category | 按分类过滤 (可选) |

### POST /api/v1/templates

```json
{
  "name": "default-firewall",
  "description": "标准防火墙模板",
  "category": "firewall",
  "content": "config defaults\n\toption input 'ACCEPT'\n..."
}
```

### PUT /api/v1/templates/:id

更新模板内容，version 字段自增。

### DELETE /api/v1/templates/:id

软删除模板。

### POST /api/v1/devices/:id/config/push

```json
// 使用模板
{ "template_id": 3 }

// 或直接指定内容
{ "content": "config interface 'lan'\n\toption proto 'static'\n..." }
```

流程：
1. 查找设备 → 获取 MAC 地址
2. 如指定 template_id → 读取模板内容
3. 通过 MQTT 发布到 `nexusgate/devices/{mac}/config`
4. 创建 DeviceConfig 记录 (status=pending)
5. 返回 config_id

### GET /api/v1/devices/:id/config/history

返回最近 50 条配置记录，按时间倒序。

## 前端页面

### Templates.vue — 配置模板管理

- 分类过滤下拉框：network / firewall / vpn / qos / system
- 表格列：名称、分类、描述、版本、更新时间
- 新建/编辑弹窗：名称、分类、描述、内容 (textarea)
- 操作：编辑、删除

### DeviceDetail.vue — 配置下发

配置历史 Tab：
- 表格：ID、状态 (tag)、时间、查看按钮
- 状态颜色：applied=success, failed=danger, pending=warning

下发弹窗：
- 模板选择器 (可选)
- UCI 内容输入框 (textarea, 10 行)
- 两种模式：从模板加载 或 直接输入

## UCI 配置格式

NexusGate 生成的配置遵循 OpenWrt UCI 标准格式：

```
config <type> '<name>'
    option <key> '<value>'
    list <key> '<value>'
```

设备端 Agent 收到后执行：
```bash
echo "$config" | uci import
uci commit
/etc/init.d/network reload
```
