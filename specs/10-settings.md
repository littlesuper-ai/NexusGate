# 系统设置

## 概述

系统设置采用 Key-Value 模型，按类别分组，支持单个 upsert 和批量保存。前端提供分类 Tab 表单界面。

## 源码文件

| 文件 | 说明 |
|------|------|
| `server/internal/model/setting.go` | SystemSetting 模型 |
| `server/internal/handler/setting.go` | CRUD + Upsert + BatchUpsert |
| `web/src/views/Settings.vue` | 系统设置页面 |

## 数据模型

### SystemSetting

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | uint | PK | 主键 |
| key | string | unique index, not null | 设置键名 |
| value | string | text | 设置值 (字符串存储) |
| category | string | index, default: general | 分类 |
| updated_at | time | auto | 最后更新时间 |

Upsert 策略：使用 `ON CONFLICT (key) DO UPDATE` 实现幂等写入。

## API 接口

### GET /api/v1/settings

| 参数 | 说明 |
|------|------|
| category | 按分类过滤 (可选) |

返回设置列表，按 category + key 排序。

### GET /api/v1/settings/:key

按键名查询单个设置。

### POST /api/v1/settings

Upsert 单个设置：

```json
{ "key": "offline_threshold", "value": "120", "category": "general" }
```

- key 已存在 → 更新 value 和 category
- key 不存在 → 新建
- category 默认为 "general"

### POST /api/v1/settings/batch

批量 upsert：

```json
[
  { "key": "system_name", "value": "NexusGate", "category": "general" },
  { "key": "mqtt_broker", "value": "tcp://localhost:1883", "category": "mqtt" },
  { "key": "alert_cpu_threshold", "value": "85", "category": "alert" }
]
```

### DELETE /api/v1/settings/:key

按键名删除设置。

## 预定义设置项

### general — 基本设置

| Key | 默认值 | 说明 |
|-----|--------|------|
| system_name | NexusGate | 系统名称 |
| offline_threshold | 120 | 设备离线判定阈值 (秒) |
| metrics_retention_days | 30 | 指标保留天数 |
| page_size | 50 | 默认分页大小 |

### mqtt — MQTT 配置

| Key | 默认值 | 说明 |
|-----|--------|------|
| mqtt_broker | tcp://localhost:1883 | Broker 地址 |
| mqtt_client_id | nexusgate-server | 客户端 ID |
| mqtt_topic_prefix | nexusgate | Topic 前缀 |
| mqtt_keepalive | 60 | 心跳间隔 (秒) |

### alert — 告警阈值

| Key | 默认值 | 说明 |
|-----|--------|------|
| alert_cpu_threshold | 85 | CPU 告警阈值 (%) |
| alert_mem_threshold | 85 | 内存告警阈值 (%) |
| alert_conntrack_threshold | 50000 | 连接数告警阈值 |
| alert_notify_method | log | 通知方式: log/webhook/email |
| alert_webhook_url | (空) | Webhook URL |

### firmware — 固件设置

| Key | 默认值 | 说明 |
|-----|--------|------|
| firmware_store_path | ./firmware_store | 固件存储路径 |
| firmware_max_size_mb | 100 | 最大固件大小 (MB) |
| firmware_auto_upgrade | false | 是否启用自动升级 |

## 前端页面

### Settings.vue

4 个 Tab 表单：

**基本设置:**
- 系统名称 (input)
- 设备离线阈值 (input-number, 30-3600)
- 指标保留天数 (input-number, 1-365)
- 默认分页大小 (input-number, 10-200)

**MQTT 配置:**
- Broker 地址 (input)
- Client ID (input)
- Topic 前缀 (input)
- 心跳间隔 (input-number, 10-300)

**告警阈值:**
- CPU 告警阈值 (input-number, 50-100)
- 内存告警阈值 (input-number, 50-100)
- 连接数告警阈值 (input-number, 1000-500000)
- 告警通知方式 (select: 仅记录日志/Webhook/邮件)
- Webhook URL (条件显示，method=webhook 时)

**固件设置:**
- 固件存储路径 (input)
- 最大固件大小 (input-number, 1-500)
- 自动升级 (switch)

保存行为：
- 将所有字段序列化为 `{key, value, category}` 数组
- 调用 batchUpsertSettings 批量写入
- 值统一存储为字符串，前端按字段类型做类型转换
