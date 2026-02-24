#!/bin/sh
# NexusGate Agent - runs on OpenWrt devices
# Handles: registration, heartbeat, config sync, remote commands

. /lib/functions.sh

CONFIG_FILE="/etc/config/nexusgate"
AGENT_ID_FILE="/etc/nexusgate/agent_id"

get_config() {
    config_load nexusgate
    config_get SERVER_URL settings server_url "http://localhost:8080"
    config_get MQTT_BROKER settings mqtt_broker "localhost"
    config_get MQTT_PORT settings mqtt_port "1883"
    config_get HEARTBEAT_INTERVAL settings heartbeat_interval "30"
    config_get DEVICE_NAME settings device_name ""
}

get_mac() {
    local mac
    mac=$(cat /sys/class/net/br-lan/address 2>/dev/null || \
          cat /sys/class/net/eth0/address 2>/dev/null || \
          echo "00:00:00:00:00:00")
    echo "$mac"
}

get_device_name() {
    if [ -n "$DEVICE_NAME" ]; then
        echo "$DEVICE_NAME"
    else
        uci get system.@system[0].hostname 2>/dev/null || echo "openwrt"
    fi
}

get_model() {
    cat /tmp/sysinfo/model 2>/dev/null || echo "unknown"
}

get_firmware() {
    . /etc/openwrt_release 2>/dev/null
    echo "${DISTRIB_REVISION:-unknown}"
}

# MQTT client ID for persistent sessions
get_client_id() {
    local mac
    mac=$(get_mac | tr -d ':')
    echo "nexusgate-${mac}"
}

# Register device with NexusGate server
register() {
    local mac name model firmware
    mac=$(get_mac)
    name=$(get_device_name)
    model=$(get_model)
    firmware=$(get_firmware)

    local payload
    payload=$(cat <<EOF
{
    "name": "$name",
    "mac": "$mac",
    "ip_address": "$(ip -4 addr show br-lan 2>/dev/null | grep -oP 'inet \K[\d.]+')",
    "model": "$model",
    "firmware": "$firmware"
}
EOF
)
    curl -sf -X POST \
        -H "Content-Type: application/json" \
        -d "$payload" \
        "${SERVER_URL}/api/v1/devices/register" > /dev/null 2>&1
}

# Collect and publish system metrics
publish_heartbeat() {
    local mac cpu_usage mem_total mem_free mem_usage uptime_secs load_avg
    local rx_bytes tx_bytes conntrack

    mac=$(get_mac)

    # CPU usage (1s sample)
    cpu_usage=$(awk '{u=$2+$4; t=$2+$4+$5; if(NR==1){pu=u;pt=t} else {printf "%.1f", (u-pu)*100/(t-pt)}}' \
        <(grep 'cpu ' /proc/stat; sleep 1; grep 'cpu ' /proc/stat) 2>/dev/null || echo "0")

    # Memory
    mem_total=$(awk '/MemTotal/{print $2}' /proc/meminfo)
    mem_free=$(awk '/MemAvailable/{print $2}' /proc/meminfo)
    mem_usage=$(awk "BEGIN{printf \"%.1f\", ($mem_total-$mem_free)/$mem_total*100}")

    # Uptime
    uptime_secs=$(cut -d. -f1 /proc/uptime)

    # Load average
    load_avg=$(cat /proc/loadavg | cut -d' ' -f1-3)

    # Network (wan interface)
    rx_bytes=$(cat /sys/class/net/eth0/statistics/rx_bytes 2>/dev/null || echo 0)
    tx_bytes=$(cat /sys/class/net/eth0/statistics/tx_bytes 2>/dev/null || echo 0)

    # Conntrack
    conntrack=$(cat /proc/sys/net/netfilter/nf_conntrack_count 2>/dev/null || echo 0)

    local topic="nexusgate/devices/${mac}/status"
    local payload
    payload=$(cat <<EOF
{"mac":"$mac","cpu_usage":$cpu_usage,"mem_usage":$mem_usage,"mem_total":$mem_total,"mem_free":$mem_free,"rx_bytes":$rx_bytes,"tx_bytes":$tx_bytes,"conntrack":$conntrack,"uptime_secs":$uptime_secs,"load_avg":"$load_avg"}
EOF
)
    mosquitto_pub -h "$MQTT_BROKER" -p "$MQTT_PORT" \
        -i "$(get_client_id)-pub" \
        -t "$topic" -m "$payload" -q 1
}

# Firmware upgrade: download, verify SHA256, and flash
sysupgrade_url() {
    local url="$1"
    local expected_sha256="$2"
    local firmware_path="/tmp/firmware.bin"

    logger -t nexusgate "Downloading firmware from $url"
    wget -q -O "$firmware_path" "$url" 2>/dev/null
    if [ $? -ne 0 ]; then
        logger -t nexusgate "ERROR: firmware download failed"
        return 1
    fi

    # Verify SHA256 if provided
    if [ -n "$expected_sha256" ]; then
        local actual_sha256
        actual_sha256=$(sha256sum "$firmware_path" | cut -d' ' -f1)
        if [ "$actual_sha256" != "$expected_sha256" ]; then
            logger -t nexusgate "ERROR: SHA256 mismatch (expected=$expected_sha256, got=$actual_sha256)"
            rm -f "$firmware_path"
            return 1
        fi
        logger -t nexusgate "SHA256 verified OK"
    fi

    logger -t nexusgate "Starting sysupgrade..."
    sysupgrade "$firmware_path"
}

# Subscribe to commands from server
subscribe_commands() {
    local mac topic client_id
    mac=$(get_mac)
    topic="nexusgate/devices/${mac}/command"
    client_id="$(get_client_id)-cmd"

    mosquitto_sub -h "$MQTT_BROKER" -p "$MQTT_PORT" \
        -i "$client_id" -q 1 -t "$topic" | while read -r msg; do
        local action
        action=$(echo "$msg" | jsonfilter -e '@.action' 2>/dev/null)

        case "$action" in
            reboot)
                logger -t nexusgate "Received reboot command"
                reboot
                ;;
            apply_config)
                logger -t nexusgate "Received config update"
                # Config content arrives on the config topic
                ;;
            upgrade)
                local url sha256 upgrade_id
                url=$(echo "$msg" | jsonfilter -e '@.url' 2>/dev/null)
                sha256=$(echo "$msg" | jsonfilter -e '@.sha256' 2>/dev/null)
                upgrade_id=$(echo "$msg" | jsonfilter -e '@.upgrade_id' 2>/dev/null)
                if [ -n "$url" ]; then
                    logger -t nexusgate "Starting firmware upgrade from $url"
                    if sysupgrade_url "$url" "$sha256"; then
                        # ACK success (this won't run if sysupgrade reboots — that's OK)
                        mosquitto_pub -h "$MQTT_BROKER" -p "$MQTT_PORT" \
                            -t "nexusgate/devices/${mac}/upgrade/ack" \
                            -m "{\"upgrade_id\":$upgrade_id,\"status\":\"success\"}" -q 1
                    else
                        mosquitto_pub -h "$MQTT_BROKER" -p "$MQTT_PORT" \
                            -t "nexusgate/devices/${mac}/upgrade/ack" \
                            -m "{\"upgrade_id\":$upgrade_id,\"status\":\"failed\",\"error\":\"download or verification failed\"}" -q 1
                    fi
                fi
                ;;
            *)
                logger -t nexusgate "Unknown command: $action"
                ;;
        esac
    done &
}

# Subscribe to config pushes (JSON envelope: {"config_id": N, "content": "..."})
subscribe_config() {
    local mac topic client_id
    mac=$(get_mac)
    topic="nexusgate/devices/${mac}/config"
    client_id="$(get_client_id)-cfg"

    mosquitto_sub -h "$MQTT_BROKER" -p "$MQTT_PORT" \
        -i "$client_id" -q 1 -t "$topic" | while read -r msg; do
        local config_id content
        config_id=$(echo "$msg" | jsonfilter -e '@.config_id' 2>/dev/null)
        content=$(echo "$msg" | jsonfilter -e '@.content' 2>/dev/null)

        if [ -z "$content" ]; then
            # Legacy: raw UCI text without envelope
            content="$msg"
        fi

        logger -t nexusgate "Applying pushed configuration (config_id=$config_id)"

        local status="applied"
        local error_msg=""
        echo "$content" | uci import 2>/tmp/nexusgate_uci_err
        if [ $? -ne 0 ]; then
            status="failed"
            error_msg=$(cat /tmp/nexusgate_uci_err 2>/dev/null)
            logger -t nexusgate "ERROR: uci import failed: $error_msg"
        else
            uci commit
            /etc/init.d/network reload 2>/dev/null
            logger -t nexusgate "Configuration applied successfully"
        fi

        # Send ACK if we have a config_id
        if [ -n "$config_id" ] && [ "$config_id" != "0" ]; then
            mosquitto_pub -h "$MQTT_BROKER" -p "$MQTT_PORT" \
                -t "nexusgate/devices/${mac}/config/ack" \
                -m "{\"config_id\":$config_id,\"status\":\"$status\",\"error\":\"$error_msg\"}" -q 1
        fi
    done &
}

# Main loop
main() {
    get_config

    logger -t nexusgate "NexusGate agent starting"

    # Initial registration
    register

    # Start MQTT subscriptions
    subscribe_commands
    subscribe_config

    # Heartbeat loop
    while true; do
        publish_heartbeat
        sleep "$HEARTBEAT_INTERVAL"
    done
}

main "$@"
