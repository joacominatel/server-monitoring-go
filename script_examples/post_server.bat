@echo off
setlocal enabledelayedexpansion

:: Script para registrar un servidor Windows en la API con información detallada del sistema
:: Uso: post_server.bat [descripción]

:: Configuración
set API_URL=http://localhost:8080/api
set AUTH_TOKEN=tu_token_jwt_aqui

:: Obtener datos del sistema
echo Recopilando información del sistema...

:: Hostname e IP
for /f "tokens=2 delims=:" %%a in ('ipconfig ^| findstr /c:"IPv4"') do (
    set IP=%%a
    set IP=!IP:~1!
    goto :got_ip
)
:got_ip

for /f "tokens=2 delims==" %%a in ('wmic computersystem get name /format:list') do (
    set HOSTNAME=%%a
)

:: Descripción
if "%~1"=="" (
    for /f "tokens=2 delims==" %%a in ('wmic os get caption /format:list') do set DESC=Windows Server %%a
) else (
    set DESC=%~1
)

:: Sistema Operativo
for /f "tokens=2 delims==" %%a in ('wmic os get Caption /format:list') do set OS=%%a
for /f "tokens=2 delims==" %%a in ('wmic os get Version /format:list') do set OS_VERSION=%%a
for /f "tokens=2 delims==" %%a in ('wmic os get OSArchitecture /format:list') do (
    set OS_ARCH=%%a
    set OS_ARCH=!OS_ARCH:bit=!
    set OS_ARCH=!OS_ARCH: =!
    if "!OS_ARCH!"=="64" set OS_ARCH=x64
    if "!OS_ARCH!"=="32" set OS_ARCH=x86
)

:: Kernel (Windows no tiene un kernel separado, usamos buildnumber)
for /f "tokens=2 delims==" %%a in ('wmic os get BuildNumber /format:list') do set KERNEL=%%a

:: CPU Información
for /f "tokens=2 delims==" %%a in ('wmic cpu get Name /format:list') do set CPU_MODEL=%%a
for /f "tokens=2 delims==" %%a in ('wmic cpu get NumberOfCores /format:list') do set CPU_CORES=%%a
for /f "tokens=2 delims==" %%a in ('wmic cpu get NumberOfLogicalProcessors /format:list') do set CPU_THREADS=%%a

:: Memoria Total
for /f "tokens=2 delims==" %%a in ('wmic ComputerSystem get TotalPhysicalMemory /format:list') do set MEM_TOTAL=%%a

:: Disco Total (C:)
for /f "tokens=2 delims==" %%a in ('wmic logicaldisk where "DeviceID='C:'" get Size /format:list') do set DISK_TOTAL=%%a

:: Crear archivo JSON temporal
set JSON_FILE=%TEMP%\server_%RANDOM%.json

echo {> %JSON_FILE%
echo   "hostname": "%HOSTNAME%",>> %JSON_FILE%
echo   "ip": "%IP%",>> %JSON_FILE%
echo   "description": "%DESC%",>> %JSON_FILE%
echo   "is_active": true,>> %JSON_FILE%
echo   "os": "%OS%",>> %JSON_FILE%
echo   "os_version": "%OS_VERSION%",>> %JSON_FILE%
echo   "os_arch": "%OS_ARCH%",>> %JSON_FILE%
echo   "kernel": "%KERNEL%",>> %JSON_FILE%
echo   "cpu_model": "%CPU_MODEL%",>> %JSON_FILE%
echo   "cpu_cores": %CPU_CORES%,>> %JSON_FILE%
echo   "cpu_threads": %CPU_THREADS%,>> %JSON_FILE%
echo   "total_memory": %MEM_TOTAL%,>> %JSON_FILE%
echo   "total_disk": %DISK_TOTAL%>> %JSON_FILE%
echo }>> %JSON_FILE%

echo Enviando información del servidor a la API...
type %JSON_FILE%

:: Enviar a la API usando curl
curl -X POST "%API_URL%/servers" -H "Content-Type: application/json" -H "Authorization: Bearer %AUTH_TOKEN%" -d @%JSON_FILE%

echo.
echo Servidor registrado correctamente.

:: Limpiar
del %JSON_FILE% 