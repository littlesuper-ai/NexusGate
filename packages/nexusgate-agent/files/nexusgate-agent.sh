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
    mosquitto_pub -h "$MQTT_BROKER" -p "$MQTT_PORT" -t "$topic" -m "$payload" -q 1
}

# Subscribe to commands from server
subscribe_commands() {
    local mac topic
    mac=$(get_mac)
    topic="nexusgate/devices/${mac}/command"

    mosquitto_sub -h "$MQTT_BROKER" -p "$MQTT_PORT" -t "$topic" -q 1 | while read -r msg; do
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
                local url
                url=$(echo "$msg" | jsonfilter -e '@.url' 2>/dev/null)
                if [ -n "$url" ]; then
                    logger -t nexusgate "Starting firmware upgrade from $url"
                    sysupgrade_url "$url"
                fi
                ;;
            *)
                logger -t nexusgate "Unknown command: $action"
                ;;
        esac
    done &
}

# Subscribe to config pushes
subscribe_config() {
    local mac topic
    mac=$(get_mac)
    topic="nexusgate/devices/${mac}/config"

    mosquitto_sub -h "$MQTT_BROKER" -p "$MQTT_PORT" -t "$topic" -q 1 | while read -r msg; do
        logger -t nexusgate "Applying pushed configuration"
        echo "$msg" | uci import 2>/dev/null && uci commit && /etc/init.d/network reload
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
