# 部署与运维

## 概述

开发环境使用 Docker Compose 一键启动所有服务。生产环境中，设备端通过 Ansible 批量部署 Agent。固件通过 OpenWrt ImageBuilder 定制构建。

## 源码文件

```
deploy/
├── docker-compose.yml              # 开发环境编排
├── mosquitto/
│   └── mosquitto.conf              # MQTT Broker 配置
├── prometheus/
│   └── prometheus.yml              # 监控采集配置
└── ansible/
    ├── inventory.yml               # 设备清单
    └── playbooks/
        ├── setup-device.yml        # 部署 Agent
        └── upgrade-firmware.yml    # 远程固件升级

firmware/
├── imagebuilder/
│   └── build.sh                    # 固件构建脚本
├── profiles/
│   ├── x86-64.conf                 # x86-64 profile
│   ├── nanopi-r4s.conf             # NanoPi R4S profile
│   └── nanopi-r5s.conf             # NanoPi R5S profile
├── packages/
│   └── enterprise.txt              # 企业包列表
└── files/
    └── etc/config/nexusgate        # 默认 Agent 配置
```

---

## 一、Docker Compose (开发环境)

### 服务清单

| 服务 | 镜像 | 端口 | 说明 |
|------|------|------|------|
| postgres | postgres:16-alpine | 5432 | 数据库 |
| mosquitto | eclipse-mosquitto:2 | 1883 | MQTT Broker |
| server | 自建 (Go) | 8080 | 后端 API |
| web | 自建 (Vue) | 3000 | 前端面板 |
| prometheus | prom/prometheus:latest | 9090 | 监控采集 |
| grafana | grafana/grafana:latest | 3001 | 监控面板 |

### 启动命令

```bash
cd deploy
docker compose up -d
```

### PostgreSQL

```yaml
environment:
  POSTGRES_USER: nexusgate
  POSTGRES_PASSWORD: nexusgate
  POSTGRES_DB: nexusgate
volumes:
  - pgdata:/var/lib/postgresql/data
healthcheck:
  test: pg_isready -U nexusgate
  interval: 5s
  retries: 5
```

### Mosquitto

```
listener 1883
allow_anonymous true
persistence true
persistence_location /mosquitto/data/
```

- 开发环境允许匿名连接
- 生产环境需配置 ACL 和密码文件

### Server 环境变量

```yaml
environment:
  LISTEN_ADDR: ":8080"
  DB_HOST: postgres
  DB_PORT: "5432"
  DB_USER: nexusgate
  DB_PASSWORD: nexusgate
  DB_NAME: nexusgate
  MQTT_BROKER: "tcp://mosquitto:1883"
  JWT_SECRET: "change-me-in-production"
depends_on:
  postgres: { condition: service_healthy }
  mosquitto: { condition: service_started }
```

### Prometheus 采集配置

```yaml
scrape_configs:
  - job_name: "nexusgate-server"
    targets: ["server:8080"]

  - job_name: "openwrt-devices"
    file_sd_configs:
      - files: ["/etc/prometheus/targets/openwrt_*.json"]
        refresh_interval: 30s
```

### 数据持久化

| Volume | 用途 |
|--------|------|
| pgdata | PostgreSQL 数据 |
| mqttdata | Mosquitto 持久化消息 |
| promdata | Prometheus 时序数据 |
| grafanadata | Grafana Dashboard/配置 |

---

## 二、Ansible (设备部署)

### 设备清单 (`inventory.yml`)

```yaml
all:
  children:
    gateways:
      hosts:
        gw-hq:
          ansible_host: 192.168.1.1
        gw-branch-a:
          ansible_host: 10.0.1.1
      vars:
        ansible_connection: ssh
        ansible_user: root
        ansible_ssh_common_args: "-o StrictHostKeyChecking=no"
        nexusgate_server: "http://192.168.100.10:8080"
        mqtt_broker: "192.168.100.10"
```

### 部署 Agent (`setup-device.yml`)

步骤：
1. **安装依赖**: `opkg install mosquitto-client-ssl jq curl`
2. **复制脚本**: Agent 主脚本 → `/usr/bin/nexusgate-agent`
3. **复制 init**: procd 启动脚本 → `/etc/init.d/nexusgate-agent`
4. **UCI 配置**:
   ```bash
   uci set nexusgate.settings=agent
   uci set nexusgate.settings.enabled='1'
   uci set nexusgate.settings.server_url='{{ nexusgate_server }}'
   uci set nexusgate.settings.mqtt_broker='{{ mqtt_broker }}'
   uci set nexusgate.settings.mqtt_port='1883'
   uci set nexusgate.settings.heartbeat_interval='30'
   uci commit nexusgate
   ```
5. **启动服务**: `service nexusgate-agent enable && service nexusgate-agent start`

### 固件升级 (`upgrade-firmware.yml`)

步骤：
1. **上传固件**: 通过 SCP 复制到 `/tmp/firmware.img.gz`
2. **校验 SHA256**: `sha256sum /tmp/firmware.img.gz`
3. **执行升级**: `sysupgrade -n /tmp/firmware.img.gz`

---

## 三、固件构建 (ImageBuilder)

### 构建脚本 (`firmware/imagebuilder/build.sh`)

| 变量 | 默认值 | 说明 |
|------|--------|------|
| OPENWRT_VERSION | 23.05.5 | OpenWrt 版本 |
| BUILD_DIR | /tmp/nexusgate-build | 构建临时目录 |
| OUTPUT_DIR | firmware/output | 输出目录 |

### 构建流程

```bash
./firmware/imagebuilder/build.sh x86-64
```

1. 读取 Profile: `firmware/profiles/x86-64.conf`
   ```bash
   TARGET=x86
   SUBTARGET=64
   DEVICE_PROFILE=generic
   ```
2. 下载 ImageBuilder (若未缓存)
3. 读取包列表: `firmware/packages/enterprise.txt`
4. 复制自定义文件: `firmware/files/` → 固件 rootfs
5. 执行构建:
   ```bash
   make image PROFILE=$DEVICE_PROFILE PACKAGES="$PACKAGES" FILES=$FILES_DIR
   ```
6. 输出到 `firmware/output/`

### Hardware Profile

| Profile | TARGET | SUBTARGET | DEVICE_PROFILE |
|---------|--------|-----------|----------------|
| x86-64 | x86 | 64 | generic |
| nanopi-r4s | rockchip | armv8 | friendlyarm_nanopi-r4s |
| nanopi-r5s | rockchip | armv8 | friendlyarm_nanopi-r5s |

### 企业包列表 (`enterprise.txt`)

**网络:**
- mwan3, luci-app-mwan3 — 多线负载均衡
- kmod-8021q — VLAN 支持
- tcpdump-mini, iperf3 — 网络诊断

**防火墙:**
- firewall4, nftables

**DNS/DHCP:**
- dnsmasq-full (替换 dnsmasq)

**VPN:**
- wireguard-tools, kmod-wireguard, luci-proto-wireguard
- openvpn-openssl, luci-app-openvpn

**QoS:**
- sqm-scripts, luci-app-sqm — SQM/CAKE

**监控:**
- collectd + 模块 (cpu, interface, load, memory, network, conntrack)
- luci-app-statistics
- vnstat2, luci-app-vnstat2
- snmpd

**MQTT & 工具:**
- mosquitto-client-ssl, jq, jsonfilter
- openssh-sftp-server
- curl, wget-ssl, nano, htop, ethtool

**Web UI:**
- luci, luci-ssl, luci-theme-openwrt-2020

### 默认 Agent 配置

`firmware/files/etc/config/nexusgate` 预置在固件中，设备开机即可使用默认配置启动 Agent。

---

## 四、本地开发 (无 Docker)

```bash
# 启动 PostgreSQL + Mosquitto (Colima)
colima start --cpu 4 --memory 4 --disk 30
docker compose -f deploy/docker-compose.yml up -d postgres mosquitto

# 启动后端
cd server
export DB_HOST=localhost DB_USER=nexusgate DB_PASSWORD=nexusgate DB_NAME=nexusgate
export MQTT_BROKER=tcp://localhost:1883 JWT_SECRET=dev-secret
go run ./cmd/nexusgate

# 启动前端
cd web
npm install
npm run dev
```

访问:
- 前端: http://localhost:3000
- API: http://localhost:8080/api/v1
- 默认登录: admin / admin123
