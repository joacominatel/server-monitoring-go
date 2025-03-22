#!/bin/bash

# Script para registrar un servidor Linux en la API con información detallada del sistema
# Uso: ./post_server.sh [descripción]

# Configuración
API_URL="http://localhost:8080/api"
AUTH_TOKEN="tu_token_jwt_aqui"  # O usa el método de login si es necesario

# Obtener datos del sistema
echo "Recopilando información del sistema..."

# Nombre de host
HOSTNAME=$(hostname)
# IP principal
IP=$(hostname -I | awk '{print $1}')
# Descripción por defecto o proporcionada por el usuario
DESCRIPTION=${1:-"Servidor Linux registrado automáticamente el $(date)"}

# Sistema Operativo
if [ -f /etc/os-release ]; then
    . /etc/os-release
    OS="${NAME}"
    OS_VERSION="${VERSION_ID}"
elif [ -f /etc/lsb-release ]; then
    . /etc/lsb-release
    OS="${DISTRIB_ID}"
    OS_VERSION="${DISTRIB_RELEASE}"
else
    OS=$(uname -s)
    OS_VERSION=$(uname -r)
fi

# Arquitectura
OS_ARCH=$(uname -m)

# Kernel
KERNEL=$(uname -r)

# CPU Información
CPU_MODEL=$(grep -m 1 "model name" /proc/cpuinfo | cut -d ":" -f2 | sed 's/^ *//')
CPU_CORES=$(grep -c ^processor /proc/cpuinfo)
CPU_THREADS=$(grep -c ^processor /proc/cpuinfo)

# En algunos sistemas multi-socket necesitamos ajustar
if [ -f /proc/cpuinfo ]; then
    PHYSICAL_CORES=$(grep "cpu cores" /proc/cpuinfo | head -1 | cut -d ":" -f2 | sed 's/^ *//')
    SOCKETS=$(grep "physical id" /proc/cpuinfo | sort -u | wc -l)
    if [ ! -z "$PHYSICAL_CORES" ] && [ ! -z "$SOCKETS" ]; then
        CPU_CORES=$((PHYSICAL_CORES * SOCKETS))
    fi
fi

# Memoria total
MEM_TOTAL=$(grep MemTotal /proc/meminfo | awk '{print $2 * 1024}')

# Disco total
DISK_TOTAL=$(df -B1 / | awk 'NR==2 {print $2}')

# Construir JSON
JSON_DATA=$(cat <<EOF
{
  "hostname": "$HOSTNAME",
  "ip": "$IP",
  "description": "$DESCRIPTION",
  "is_active": true,
  "os": "$OS",
  "os_version": "$OS_VERSION",
  "os_arch": "$OS_ARCH",
  "kernel": "$KERNEL",
  "cpu_model": "$CPU_MODEL",
  "cpu_cores": $CPU_CORES,
  "cpu_threads": $CPU_THREADS,
  "total_memory": $MEM_TOTAL,
  "total_disk": $DISK_TOTAL
}
EOF
)

echo "Enviando información del servidor a la API..."
echo "$JSON_DATA"

# Enviar a la API
curl -X POST "$API_URL/servers" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $AUTH_TOKEN" \
  -d "$JSON_DATA"

echo -e "\nServidor registrado correctamente." 