# 固件管理与 OTA 升级

## 概述

支持固件文件上传、版本管理、SHA256 校验、单设备/批量 OTA 升级，以及升级记录追踪。固件通过 MQTT 命令通知设备下载升级。

## 源码文件

| 文件 | 说明 |
|------|------|
| `server/internal/model/firmware.go` | Firmware、FirmwareUpgrade 模型 |
| `server/internal/handler/firmware.go` | 上传/下载/删除/升级/批量升级 |
| `web/src/views/Firmware.vue` | 固件管理页面 |
| `web/src/views/DeviceDetail.vue` | 升级记录 Tab |

## 数据模型

### Firmware

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | uint | PK | 主键 |
| version | string | not null | 版本号 (如 23.05.5-r2) |
| target | string | not null | 目标平台 (x86-64/nanopi-r4s/nanopi-r5s) |
| filename | string | not null | 文件名 |
| file_size | int64 | - | 文件大小 (bytes) |
| sha256 | string | - | SHA256 校验和 |
| download_url | string | - | 下载 URL |
| changelog | string | text | 更新日志 |
| is_stable | bool | default: false | 是否标记为稳定版 |
| created_at | time | auto | 上传时间 |
| updated_at | time | auto | 更新时间 |
| deleted_at | time | soft delete | 软删除 |

### FirmwareUpgrade

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | uint | PK | 主键 |
| device_id | uint | index, not null | 目标设备 |
| firmware_id | uint | not null | 固件 ID |
| status | string | default: pending | pending/downloading/upgrading/success/failed |
| started_at | *time | - | 开始时间 |
| finished_at | *time | - | 完成时间 |
| error_msg | string | - | 错误信息 |
| created_at | time | auto | 创建时间 |

## API 接口

### GET /api/v1/firmware

| 参数 | 说明 |
|------|------|
| target | 按平台过滤 (可选) |

返回固件列表，按上传时间倒序。

### POST /api/v1/firmware/upload

Multipart 表单上传：

| 字段 | 说明 |
|------|------|
| file | 固件文件 (必须) |
| version | 版本号 (必须) |
| target | 目标平台 (必须) |
| changelog | 更新日志 (可选) |

处理流程：
1. 解析 multipart 文件
2. 保存到 `./firmware_store/` 目录
3. 计算 SHA256 哈希
4. 创建 Firmware 记录（含 file_size, sha256, download_url）

### GET /api/v1/firmware/download/:filename

直接下载固件文件。

### DELETE /api/v1/firmware/:id

删除固件文件和数据库记录。

### POST /api/v1/firmware/:id/stable

标记为稳定版（is_stable = true）。

### POST /api/v1/firmware/upgrade

单设备升级：

```json
{
  "device_id": 1,
  "firmware_id": 3
}
```

流程：
1. 查找设备 → 获取 MAC
2. 查找固件 → 获取下载 URL 和 SHA256
3. 创建 FirmwareUpgrade 记录 (status=pending)
4. MQTT 发布到 `nexusgate/devices/{mac}/command`:
   ```json
   {
     "action": "upgrade",
     "url": "/api/v1/firmware/download/openwrt-23.05.5-x86-64.img.gz",
     "sha256": "abc123...",
     "version": "23.05.5-r2"
   }
   ```
5. 返回 upgrade_id

### POST /api/v1/firmware/upgrade/batch

批量升级：

```json
{
  "firmware_id": 3,
  "group": "分支",      // 可选，按分组过滤
  "model": "NanoPi R4S" // 可选，按型号过滤
}
```

对每个匹配的设备执行单设备升级流程。

### GET /api/v1/firmware/upgrades

| 参数 | 说明 |
|------|------|
| device_id | 按设备过滤 (可选) |

返回最近 100 条升级记录，按时间倒序。

## 前端页面

### Firmware.vue

- 目标平台过滤：x86-64 / NanoPi R4S / NanoPi R5S
- 固件表格：版本、平台、文件名、大小、SHA256、稳定标记
- 操作：下载、标记稳定、推送升级、批量升级、删除
- 上传弹窗：版本号、目标平台、更新日志、文件选择
- 批量升级弹窗：选择固件 + 按分组/型号过滤

## 设备端升级流程

Agent 收到 upgrade 命令后执行：
```bash
wget -O /tmp/firmware.img.gz "$url"
sha256sum /tmp/firmware.img.gz | grep -q "$sha256"
sysupgrade -n /tmp/firmware.img.gz
```

`sysupgrade -n` 表示不保留配置（全新安装）。根据需要可改为 `sysupgrade` (保留配置)。
