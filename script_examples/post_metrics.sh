#!/bin/bash

# Script para enviar métricas detalladas a la API desde sistemas Linux
# Uso: ./post_metrics.sh <server_id>

# Configuración
API_URL="http://localhost:8080/api"
AUTH_TOKEN="tu_token_jwt_aqui"  # O usa el método de login si es necesario

# Verificar argumentos
if [ -z "$1" ]; then
    echo "Uso: $0 <server_id>"
    exit 1
fi

SERVER_ID=$1
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

# Obtener métricas del sistema
echo "Recopilando métricas del sistema..."

# CPU
CPU_USAGE=$(top -bn1 | grep "Cpu(s)" | sed "s/.*, *\([0-9.]*\)%* id.*/\1/" | awk '{print 100 - $1}')
CPU_FREQ=$(cat /sys/devices/system/cpu/cpu0/cpufreq/scaling_cur_freq 2>/dev/null || echo "0")
if [ "$CPU_FREQ" != "0" ]; then
    CPU_FREQ=$(echo "scale=2; $CPU_FREQ / 1000" | bc)
fi
CPU_TEMP=$(sensors 2>/dev/null | grep "Core 0" | awk '{print $3}' | sed 's/[^0-9.]//g' || echo "0")

# Cargas promedio
LOAD_AVG=$(cat /proc/loadavg)
LOAD_AVG1=$(echo $LOAD_AVG | cut -d ' ' -f 1)
LOAD_AVG5=$(echo $LOAD_AVG | cut -d ' ' -f 2)
LOAD_AVG15=$(echo $LOAD_AVG | cut -d ' ' -f 3)

# Memoria
MEM_INFO=$(free -b)
MEM_TOTAL=$(echo "$MEM_INFO" | grep Mem | awk '{print $2}')
MEM_USED=$(echo "$MEM_INFO" | grep Mem | awk '{print $3}')
MEM_FREE=$(echo "$MEM_INFO" | grep Mem | awk '{print $4}')
MEM_CACHE=$(echo "$MEM_INFO" | grep Mem | awk '{print $6}')
MEM_BUFFERS=$(echo "$MEM_INFO" | grep Mem | awk '{print $5}')
SWAP_TOTAL=$(echo "$MEM_INFO" | grep Swap | awk '{print $2}')
SWAP_USED=$(echo "$MEM_INFO" | grep Swap | awk '{print $3}')
SWAP_FREE=$(echo "$MEM_INFO" | grep Swap | awk '{print $4}')

# Disco
DISK_INFO=$(df -B1 / | tail -1)
DISK_TOTAL=$(echo "$DISK_INFO" | awk '{print $2}')
DISK_USED=$(echo "$DISK_INFO" | awk '{print $3}')
DISK_FREE=$(echo "$DISK_INFO" | awk '{print $4}')

# Estadísticas de IO de disco (requiere iostat)
if command -v iostat >/dev/null 2>&1; then
    DISK_STATS=$(iostat -d -k 1 2 | grep -A 1 Device | tail -1)
    DISK_READS=$(echo "$DISK_STATS" | awk '{print $2}')
    DISK_WRITES=$(echo "$DISK_STATS" | awk '{print $3}')
    DISK_READ_BYTES=$(echo "$DISK_STATS" | awk '{print $4 * 1024}')
    DISK_WRITE_BYTES=$(echo "$DISK_STATS" | awk '{print $5 * 1024}')
else
    DISK_READS=0
    DISK_WRITES=0
    DISK_READ_BYTES=0
    DISK_WRITE_BYTES=0
fi

# Red - Requiere ifstat para estadísticas
NET_UPLOAD=0
NET_DOWNLOAD=0
NET_PACKETS_IN=0
NET_PACKETS_OUT=0
NET_ERRORS_IN=0
NET_ERRORS_OUT=0

if command -v ifstat >/dev/null 2>&1; then
    NET_STATS=$(ifstat -i eth0 -b 1 1 | tail -1)
    NET_DOWNLOAD=$(echo "$NET_STATS" | awk '{print $1 * 1024}')
    NET_UPLOAD=$(echo "$NET_STATS" | awk '{print $2 * 1024}')
elif [ -f "/proc/net/dev" ]; then
    # Alternativa usando /proc/net/dev
    IFACE=$(ip route | grep default | awk '{print $5}')
    NET_STATS1=$(grep $IFACE /proc/net/dev | awk '{print $2, $10}')
    sleep 1
    NET_STATS2=$(grep $IFACE /proc/net/dev | awk '{print $2, $10}')
    NET_DOWNLOAD=$(echo "$NET_STATS2" | awk '{print $1}' - echo "$NET_STATS1" | awk '{print $1}')
    NET_UPLOAD=$(echo "$NET_STATS2" | awk '{print $2}' - echo "$NET_STATS1" | awk '{print $2}')
    
    # Paquetes y errores
    NET_INFO=$(grep $IFACE /proc/net/dev | awk '{print $3, $4, $11, $12}')
    NET_PACKETS_IN=$(echo "$NET_INFO" | awk '{print $1}')
    NET_ERRORS_IN=$(echo "$NET_INFO" | awk '{print $2}')
    NET_PACKETS_OUT=$(echo "$NET_INFO" | awk '{print $3}')
    NET_ERRORS_OUT=$(echo "$NET_INFO" | awk '{print $4}')
fi

# Procesos
PROCESS_COUNT=$(ps -e | wc -l)
THREAD_COUNT=$(ps -eLf | wc -l)
HANDLE_COUNT=$(lsof | wc -l)

# Uptime
UPTIME=$(cat /proc/uptime | awk '{print $1}' | cut -d. -f1)

# Construir JSON
JSON_DATA=$(cat <<EOF
{
  "server_id": $SERVER_ID,
  "timestamp": "$TIMESTAMP",
  "cpu_usage": $CPU_USAGE,
  "cpu_temp": $CPU_TEMP,
  "cpu_freq": $CPU_FREQ,
  "load_avg_1": $LOAD_AVG1,
  "load_avg_5": $LOAD_AVG5,
  "load_avg_15": $LOAD_AVG15,
  "memory_total": $MEM_TOTAL,
  "memory_used": $MEM_USED,
  "memory_free": $MEM_FREE,
  "memory_cache": $MEM_CACHE,
  "memory_buffers": $MEM_BUFFERS,
  "swap_total": $SWAP_TOTAL,
  "swap_used": $SWAP_USED,
  "swap_free": $SWAP_FREE,
  "disk_total": $DISK_TOTAL,
  "disk_used": $DISK_USED,
  "disk_free": $DISK_FREE,
  "disk_reads": $DISK_READS,
  "disk_writes": $DISK_WRITES,
  "disk_read_bytes": $DISK_READ_BYTES,
  "disk_write_bytes": $DISK_WRITE_BYTES,
  "net_upload": $NET_UPLOAD,
  "net_download": $NET_DOWNLOAD,
  "net_packets_in": $NET_PACKETS_IN,
  "net_packets_out": $NET_PACKETS_OUT,
  "net_errors_in": $NET_ERRORS_IN,
  "net_errors_out": $NET_ERRORS_OUT,
  "process_count": $PROCESS_COUNT,
  "thread_count": $THREAD_COUNT,
  "handle_count": $HANDLE_COUNT,
  "uptime": $UPTIME
}
EOF
)

echo "Enviando métricas a la API..."
echo "$JSON_DATA"

# Enviar a la API
curl -X POST "$API_URL/metrics" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $AUTH_TOKEN" \
  -d "$JSON_DATA"

echo -e "\nMétricas enviadas correctamente." 