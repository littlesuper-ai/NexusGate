# OpenWrt 设备端代理

## 概述

nexusgate-agent 是运行在 OpenWrt 设备上的 Shell 脚本，负责设备注册、心跳上报、远程命令接收和配置同步。通过 procd 管理生命周期，支持自动重启。

## 源码文件

```
packages/nexusgate-agent/
├── Makefile                              # OpenWrt 包构建定义
└── files/
    ├── nexusgate-agent.sh                # 主脚本
    ├── nexusgate-agent.init              # procd init 脚本
    └── nexusgate.conf                    # 默认 UCI 配置
```

## 包构建 (Makefile)

| 字段 | 值 |
|------|-----|
| PKG_NAME | nexusgate-agent |
| SECTION | net |
| CATEGORY | Network |
| DEPENDS | +curl +mosquitto-client-ssl +jq |

安装位置：
- `/usr/bin/nexusgate-agent` — 主脚本
- `/etc/init.d/nexusgate-agent` — init 脚本
- `/etc/config/nexusgate` — UCI 配置

## UCI 配置

```
config agent 'settings'
    option enabled '1'
    option server_url 'http://192.168.1.100:8080'
    option mqtt_broker '192.168.1.100'
    option mqtt_port '1883'
    option heartbeat_interval '30'
    option device_name ''
```

| 选项 | 默认值 | 说明 |
|------|--------|------|
| enabled | 1 | 是否启用 |
| server_url | http://localhost:8080 | 服务端 HTTP 地址 |
| mqtt_broker | localhost | MQTT Broker 主机 |
| mqtt_port | 1883 | MQTT 端口 |
| heartbeat_interval | 30 | 心跳间隔 (秒) |
| device_name | (空) | 自定义设备名 (为空则用 hostname) |

## init 脚本

```bash
START=99          # 启动顺序 (最后启动)
USE_PROCD=1       # 使用 procd

start_service()   # procd_open_instance → 设置 respawn → procd_close_instance
stop_service()    # 自动处理
```

procd respawn 配置：进程异常退出后自动重启。

## 主脚本功能

### 1. 配置加载 (`get_config`)

从 UCI 读取配置：
```bash
config_load nexusgate
config_get SERVER_URL settings server_url "http://localhost:8080"
config_get MQTT_BROKER settings mqtt_broker "localhost"
# ...
```

### 2. 设备信息采集

| 函数 | 数据源 | 返回值 |
|------|--------|--------|
| `get_mac` | br-lan / eth0 | MAC 地址 |
| `get_device_name` | UCI 或 hostname | 设备名称 |
| `get_model` | `/tmp/sysinfo/model` | 型号 |
| `get_firmware` | `/etc/openwrt_release` DISTRIB_REVISION | 固件版本 |

### 3. 设备注册 (`register`)

```bash
curl -s -X POST "$SERVER_URL/api/v1/devices/register" \
  -H "Content-Type: application/json" \
  -d '{"name":"...","mac":"...","ip_address":"...","model":"...","firmware":"..."}'
```

- IP 地址从 br-lan 接口获取
- 启动时执行一次

### 4. 心跳上报 (`publish_heartbeat`)

循环采集并发布到 MQTT：

```bash
mosquitto_pub -h "$MQTT_BROKER" -p "$MQTT_PORT" \
  -t "nexusgate/devices/$MAC/status" \
  -m "$payload" -q 1
```

**CPU 采集算法：**
```bash
# awk 内联采样, 读取 /proc/stat 两次 (间隔 1 秒), 计算 user+system 占比
cpu_usage=$(awk '{u=$2+$4; t=$2+$4+$5; if(NR==1){pu=u;pt=t} else {printf "%.1f", (u-pu)*100/(t-pt)}}' \
    <(grep 'cpu ' /proc/stat; sleep 1; grep 'cpu ' /proc/stat) 2>/dev/null || echo "0")
```

**内存采集：**
```bash
# 从 /proc/meminfo 读取, 单位为 KB (未转换为 bytes)
mem_total=$(awk '/MemTotal/{print $2}' /proc/meminfo)
mem_free=$(awk '/MemAvailable/{print $2}' /proc/meminfo)   # 注意: 使用 MemAvailable, 非 MemFree
mem_usage=$(awk "BEGIN{printf \"%.1f\", ($mem_total-$mem_free)/$mem_total*100}")
```

注意：心跳 payload 中 mem_total 和 mem_free 的值是 KB 单位 (来自 /proc/meminfo 原始值)，而非 bytes。

**网络流量：**
```bash
rx_bytes=$(cat /sys/class/net/eth0/statistics/rx_bytes)
tx_bytes=$(cat /sys/class/net/eth0/statistics/tx_bytes)
```

**连接追踪：**
```bash
conntrack=$(cat /proc/sys/net/netfilter/nf_conntrack_count)
```

### 5. 命令接收 (`subscribe_commands`)

```bash
mosquitto_sub -h "$MQTT_BROKER" -p "$MQTT_PORT" \
  -t "nexusgate/devices/$MAC/command" -q 1
```

| 命令 | 处理 |
|------|------|
| `{"action":"reboot"}` | 执行 `reboot` |
| `{"action":"upgrade","url":"...","sha256":"..."}` | 调用 `sysupgrade_url "$url"` (**TODO: 函数未定义, 需实现**) |
| `{"action":"apply_config"}` | 记录日志 (实际配置通过 config topic 推送) |

### 6. 配置同步 (`subscribe_config`)

```bash
mosquitto_sub -h "$MQTT_BROKER" -p "$MQTT_PORT" \
  -t "nexusgate/devices/$MAC/config" -q 1
```

收到 UCI 配置文本后：
```bash
echo "$config" | uci import
uci commit
/etc/init.d/network reload
```

### 7. 主流程 (`main`)

```
get_config()
register()
subscribe_commands() &    # 后台进程
subscribe_config() &      # 后台进程
while true; do
    publish_heartbeat()
    sleep $HEARTBEAT_INTERVAL
done
```

## MQTT Topic 清单

| Topic | 方向 | QoS | 用途 |
|-------|------|-----|------|
| nexusgate/devices/{mac}/status | Agent → Server | 1 | 心跳指标 |
| nexusgate/devices/{mac}/command | Server → Agent | 1 | 远程命令 |
| nexusgate/devices/{mac}/config | Server → Agent | 1 | 配置下发 |

## 依赖软件包

| 包 | 用途 |
|----|------|
| curl | HTTP 注册请求 |
| mosquitto-client-ssl | MQTT 发布/订阅 |
| jq / jsonfilter | JSON 解析 |
