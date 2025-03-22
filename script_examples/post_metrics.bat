@echo off
setlocal enabledelayedexpansion

:: Script para enviar métricas detalladas a la API desde sistemas Windows
:: Uso: post_metrics.bat <server_id>

:: Configuración
set API_URL=http://localhost:8080/api
set AUTH_TOKEN=tu_token_jwt_aqui

:: Verificar argumentos
if "%~1"=="" (
    echo Uso: %0 ^<server_id^>
    exit /b 1
)

set SERVER_ID=%~1

:: Obtener fecha y hora en formato ISO
for /f "tokens=2 delims==" %%I in ('wmic os get LocalDateTime /format:list') do set DATETIME=%%I
set TIMESTAMP=%DATETIME:~0,4%-%DATETIME:~4,2%-%DATETIME:~6,2%T%DATETIME:~8,2%:%DATETIME:~10,2%:%DATETIME:~12,2%Z

echo Recopilando métricas del sistema...

:: CPU
for /f "tokens=2 delims==" %%I in ('wmic cpu get LoadPercentage /format:list') do set CPU_USAGE=%%I
:: Si no funciona, alternativa con typeperf
if "%CPU_USAGE%"=="" (
    for /f "skip=2 tokens=3" %%I in ('typeperf "\Processor(_Total)\%% Processor Time" -sc 1') do set CPU_USAGE=%%I
    set CPU_USAGE=!CPU_USAGE:"=!
)

:: CPU Temperatura (requiere herramientas adicionales como OpenHardwareMonitor)
set CPU_TEMP=0

:: CPU Frecuencia
for /f "tokens=2 delims==" %%I in ('wmic cpu get CurrentClockSpeed /format:list') do set CPU_FREQ=%%I

:: Memoria
for /f "tokens=2 delims==" %%I in ('wmic OS get TotalVisibleMemorySize /format:list') do set MEM_TOTAL=%%I
set /a MEM_TOTAL=MEM_TOTAL*1024

for /f "tokens=2 delims==" %%I in ('wmic OS get FreePhysicalMemory /format:list') do set MEM_FREE=%%I
set /a MEM_FREE=MEM_FREE*1024

set /a MEM_USED=MEM_TOTAL-MEM_FREE

:: Memoria caché y buffers (no disponible directamente en WMI)
set MEM_CACHE=0
set MEM_BUFFERS=0

:: Swap
for /f "tokens=2 delims==" %%I in ('wmic OS get TotalVirtualMemorySize /format:list') do set SWAP_TOTAL=%%I
set /a SWAP_TOTAL=SWAP_TOTAL*1024

for /f "tokens=2 delims==" %%I in ('wmic OS get FreeVirtualMemory /format:list') do set SWAP_FREE=%%I
set /a SWAP_FREE=SWAP_FREE*1024

set /a SWAP_USED=SWAP_TOTAL-SWAP_FREE

:: Disco (unidad C:)
for /f "tokens=2 delims==" %%I in ('wmic logicaldisk where "DeviceID='C:'" get Size /format:list') do set DISK_TOTAL=%%I
for /f "tokens=2 delims==" %%I in ('wmic logicaldisk where "DeviceID='C:'" get FreeSpace /format:list') do set DISK_FREE=%%I
set /a DISK_USED=DISK_TOTAL-DISK_FREE

:: Disco IO (requiere typeperf)
set DISK_READS=0
set DISK_WRITES=0
set DISK_READ_BYTES=0
set DISK_WRITE_BYTES=0
for /f "skip=2 tokens=3" %%I in ('typeperf "\PhysicalDisk(_Total)\Disk Reads/sec" -sc 1 2^>nul') do set DISK_READS=%%I
set DISK_READS=!DISK_READS:"=!

for /f "skip=2 tokens=3" %%I in ('typeperf "\PhysicalDisk(_Total)\Disk Writes/sec" -sc 1 2^>nul') do set DISK_WRITES=%%I
set DISK_WRITES=!DISK_WRITES:"=!

for /f "skip=2 tokens=3" %%I in ('typeperf "\PhysicalDisk(_Total)\Disk Read Bytes/sec" -sc 1 2^>nul') do set DISK_READ_BYTES=%%I
set DISK_READ_BYTES=!DISK_READ_BYTES:"=!

for /f "skip=2 tokens=3" %%I in ('typeperf "\PhysicalDisk(_Total)\Disk Write Bytes/sec" -sc 1 2^>nul') do set DISK_WRITE_BYTES=%%I
set DISK_WRITE_BYTES=!DISK_WRITE_BYTES:"=!

:: Red (typeperf)
set NET_UPLOAD=0
set NET_DOWNLOAD=0
set NET_PACKETS_IN=0
set NET_PACKETS_OUT=0
set NET_ERRORS_IN=0
set NET_ERRORS_OUT=0

for /f "skip=2 tokens=3" %%I in ('typeperf "\Network Interface(*)\Bytes Sent/sec" -sc 1 2^>nul') do set NET_UPLOAD=%%I
set NET_UPLOAD=!NET_UPLOAD:"=!

for /f "skip=2 tokens=3" %%I in ('typeperf "\Network Interface(*)\Bytes Received/sec" -sc 1 2^>nul') do set NET_DOWNLOAD=%%I
set NET_DOWNLOAD=!NET_DOWNLOAD:"=!

for /f "skip=2 tokens=3" %%I in ('typeperf "\Network Interface(*)\Packets Received/sec" -sc 1 2^>nul') do set NET_PACKETS_IN=%%I
set NET_PACKETS_IN=!NET_PACKETS_IN:"=!

for /f "skip=2 tokens=3" %%I in ('typeperf "\Network Interface(*)\Packets Sent/sec" -sc 1 2^>nul') do set NET_PACKETS_OUT=%%I
set NET_PACKETS_OUT=!NET_PACKETS_OUT:"=!

:: Procesos
for /f %%I in ('wmic process get Caption /format:list ^| find /c "Caption"') do set PROCESS_COUNT=%%I
for /f %%I in ('wmic thread get /format:list ^| find /c "ThreadId"') do set THREAD_COUNT=%%I
for /f %%I in ('wmic process get HandleCount /format:list ^| find /c /v ""') do set /a HANDLE_COUNT=%%I-3

:: Uptime
for /f "tokens=1" %%I in ('wmic os get LastBootUpTime /format:list ^| findstr "="') do set LAST_BOOT=%%I
set LAST_BOOT=!LAST_BOOT:~13,17!
for /f "tokens=2 delims==" %%J in ('wmic os get LocalDateTime /format:list') do set NOW=%%J
set NOW=!NOW:~0,17!

:: Crear archivo JSON temporal
set JSON_FILE=%TEMP%\metrics_%RANDOM%.json

echo {> %JSON_FILE%
echo   "server_id": %SERVER_ID%,>> %JSON_FILE%
echo   "timestamp": "%TIMESTAMP%",>> %JSON_FILE%
echo   "cpu_usage": %CPU_USAGE%,>> %JSON_FILE%
echo   "cpu_temp": %CPU_TEMP%,>> %JSON_FILE%
echo   "cpu_freq": %CPU_FREQ%,>> %JSON_FILE%
echo   "memory_total": %MEM_TOTAL%,>> %JSON_FILE%
echo   "memory_used": %MEM_USED%,>> %JSON_FILE%
echo   "memory_free": %MEM_FREE%,>> %JSON_FILE%
echo   "memory_cache": %MEM_CACHE%,>> %JSON_FILE%
echo   "memory_buffers": %MEM_BUFFERS%,>> %JSON_FILE%
echo   "swap_total": %SWAP_TOTAL%,>> %JSON_FILE%
echo   "swap_used": %SWAP_USED%,>> %JSON_FILE%
echo   "swap_free": %SWAP_FREE%,>> %JSON_FILE%
echo   "disk_total": %DISK_TOTAL%,>> %JSON_FILE%
echo   "disk_used": %DISK_USED%,>> %JSON_FILE%
echo   "disk_free": %DISK_FREE%,>> %JSON_FILE%
echo   "disk_reads": %DISK_READS%,>> %JSON_FILE%
echo   "disk_writes": %DISK_WRITES%,>> %JSON_FILE%
echo   "disk_read_bytes": %DISK_READ_BYTES%,>> %JSON_FILE%
echo   "disk_write_bytes": %DISK_WRITE_BYTES%,>> %JSON_FILE%
echo   "net_upload": %NET_UPLOAD%,>> %JSON_FILE%
echo   "net_download": %NET_DOWNLOAD%,>> %JSON_FILE%
echo   "net_packets_in": %NET_PACKETS_IN%,>> %JSON_FILE%
echo   "net_packets_out": %NET_PACKETS_OUT%,>> %JSON_FILE%
echo   "net_errors_in": %NET_ERRORS_IN%,>> %JSON_FILE%
echo   "net_errors_out": %NET_ERRORS_OUT%,>> %JSON_FILE%
echo   "process_count": %PROCESS_COUNT%,>> %JSON_FILE%
echo   "thread_count": %THREAD_COUNT%,>> %JSON_FILE%
echo   "handle_count": %HANDLE_COUNT%>> %JSON_FILE%
echo }>> %JSON_FILE%

echo Enviando métricas a la API...
type %JSON_FILE%

:: Enviar a la API usando curl
curl -X POST "%API_URL%/metrics" -H "Content-Type: application/json" -H "Authorization: Bearer %AUTH_TOKEN%" -d @%JSON_FILE%

echo.
echo Métricas enviadas correctamente.

:: Limpiar
del %JSON_FILE% 