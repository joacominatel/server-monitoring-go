# Simulador de Métricas para Servidores

Este es un simulador de métricas para servidores que permite generar y enviar datos de monitoreo a una API REST. Es útil para pruebas, desarrollo y debugging de sistemas de monitoreo.

## Características

- Simulación de métricas para múltiples servidores
- Soporte para diferentes sistemas operativos (Linux y Windows)
- Generación de métricas realistas para:
  - CPU (uso, temperatura, frecuencia)
  - Memoria (total, usada, libre, caché, buffers)
  - Swap
  - Disco (espacio, IO)
  - Red (tráfico, paquetes, errores)
  - Procesos y servicios
- Modo de estrés para pruebas de alertas
- Creación automática de umbrales de alerta
- Autenticación automática con la API

## Requisitos

- Go 1.16 o superior
- Acceso a una API REST en `http://localhost:8080/api`
- Credenciales de acceso (por defecto: admin/admin123)

## Configuración

El simulador se conecta por defecto a `http://localhost:8080/api`. Las credenciales predeterminadas son:
- Usuario: `admin`
- Contraseña: `admin123`

## Uso

### Ejecución Básica

```bash
go run debug.go
```

Esto iniciará el simulador con la configuración predeterminada:
- 3 servidores
- Envío de métricas cada 2 segundos
- Valores aleatorios realistas

### Modo de Estrés

Para probar alertas y umbrales, puedes usar el modo de estrés:

```bash
go run debug.go --stress --metric cpu --threshold 95
```

Opciones disponibles:
- `--stress`: Activa el modo de estrés
- `--metric`: Tipo de métrica a estresar (cpu, memory, disk, net_in, net_out, random)
- `--server`: ID del servidor a estresar (0 para todos)
- `--interval`: Intervalo en segundos entre envíos (default: 2)
- `--threshold`: Valor umbral para modo estrés (default: 95.0)
- `--create-thresholds`: Crea umbrales de alerta automáticamente

### Ejemplos de Uso

1. Estresar CPU en todos los servidores:
```bash
go run debug.go --stress --metric cpu --threshold 95
```

2. Estresar memoria en un servidor específico:
```bash
go run debug.go --stress --metric memory --server 1 --threshold 90
```

3. Estresar tráfico de red con umbrales automáticos:
```bash
go run debug.go --stress --metric net_in --threshold 50 --create-thresholds
```

## Métricas Generadas

El simulador genera las siguientes métricas:

### CPU
- Uso de CPU (%)
- Temperatura (°C)
- Frecuencia (MHz)
- Carga promedio (1, 5, 15 minutos)

### Memoria
- Memoria total
- Memoria usada
- Memoria libre
- Caché
- Buffers
- Swap (total, usado, libre)

### Disco
- Espacio total
- Espacio usado
- Espacio libre
- Operaciones de lectura/escritura
- Bytes leídos/escritos
- Tiempo de IO

### Red
- Subida/Descarga
- Paquetes entrantes/salientes
- Errores
- Paquetes descartados

### Sistema
- Número de procesos
- Número de hilos
- Número de handles
- Tiempo de actividad

## Umbrales de Alerta

Cuando se usa `--create-thresholds`, se crean automáticamente los siguientes umbrales:

- CPU: Alerta cuando > 90%
- Memoria: Alerta cuando > 90%
- Disco: Alerta cuando > 90%
- Red entrante: Alerta cuando > 50 MB/s
- Red saliente: Alerta cuando > 50 MB/s

## Notas

- Los valores generados son simulados y no representan métricas reales
- El modo de estrés es útil para probar sistemas de alertas
- Los servidores se crean automáticamente si no existen
- Las métricas se envían en formato JSON a la API 