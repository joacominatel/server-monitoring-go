# Backend de Plataforma de Monitoreo de Servidores

Este es el backend para la Plataforma de Monitoreo de Servidores, desarrollado en Go.

## Características

- API REST para recibir y consultar métricas de servidores
- Almacenamiento en PostgreSQL usando GORM
- Gestión de servidores y métricas
- Consulta de métricas por rango de tiempo
- Logs estructurados

## Requisitos

- Go 1.18 o superior
- PostgreSQL 12 o superior

## Configuración

1. Clonar el repositorio
2. Configurar las variables de entorno:

```
# Archivo .env (ya está creado con valores por defecto)
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=server_monitoring
SERVER_PORT=8080
ENV=development
```

## Ejecución

```bash
# Compilar
go build -o server

# Ejecutar
./server
```

Alternativamente, puedes ejecutar directamente con:

```bash
go run main.go
```

## Estructura del proyecto

```
/backend
  ├── main.go              # Punto de entrada principal
  ├── config/              # Configuración de la aplicación
  ├── internal/            # Código interno de la aplicación
  │   ├── handlers/        # Manejadores HTTP
  │   ├── models/          # Modelos de datos
  │   ├── services/        # Lógica de negocio
  ├── pkg/                 # Bibliotecas compartidas
  │   ├── logger/          # Sistema de logging
  │   ├── database/        # Conexión a la base de datos
```

## API Endpoints

### Servidores

- `GET /api/servers` - Obtener todos los servidores
- `GET /api/servers/:id` - Obtener un servidor por ID
- `POST /api/servers` - Crear un nuevo servidor
- `PUT /api/servers/:id` - Actualizar un servidor
- `DELETE /api/servers/:id` - Eliminar un servidor

### Métricas

- `POST /api/metrics` - Crear una nueva métrica
- `GET /api/metrics/server/:server_id` - Obtener métricas por ID de servidor
- `GET /api/metrics/server/:server_id/latest` - Obtener la última métrica de un servidor
- `GET /api/metrics/server/:server_id/timerange` - Obtener métricas por rango de tiempo

## Ejemplos de uso

### Crear un servidor

```bash
curl -X POST http://localhost:8080/api/servers \
  -H "Content-Type: application/json" \
  -d '{
    "hostname": "servidor1.example.com",
    "ip": "192.168.1.100",
    "description": "Servidor de producción"
  }'
```

### Crear una métrica

```bash
curl -X POST http://localhost:8080/api/metrics \
  -H "Content-Type: application/json" \
  -d '{
    "server_id": 1,
    "cpu_usage": 45.5,
    "memory_total": 16000000000,
    "memory_used": 8000000000,
    "memory_free": 8000000000,
    "disk_total": 1000000000000,
    "disk_used": 300000000000,
    "disk_free": 700000000000
  }'
```

### Obtener métricas por rango de tiempo

```bash
curl "http://localhost:8080/api/metrics/server/1/timerange?start=2023-10-01T00:00:00Z&end=2023-10-02T23:59:59Z"
``` 