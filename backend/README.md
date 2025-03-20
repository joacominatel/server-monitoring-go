# Backend de Plataforma de Monitoreo de Servidores

Este es el backend para la Plataforma de Monitoreo de Servidores, desarrollado en Go.

## Características

- API REST para recibir y consultar métricas de servidores
- Almacenamiento en PostgreSQL usando GORM
- Gestión de servidores y métricas
- Consulta de métricas por rango de tiempo
- Logs estructurados con persistencia en base de datos
- Consulta y gestión de logs históricos vía API
- Autenticación y autorización con JWT en cookies
- Control de acceso basado en roles (RBAC)

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
JWT_SECRET=mi_clave_secreta_jwt_para_desarrollo
ADMIN_PASSWORD=admin123
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
  │   ├── middleware/      # Middleware (autenticación, etc.)
  │   ├── models/          # Modelos de datos
  │   ├── services/        # Lógica de negocio
  ├── pkg/                 # Bibliotecas compartidas
  │   ├── logger/          # Sistema de logging
  │   ├── database/        # Conexión a la base de datos
```

## Autenticación y Autorización

El sistema utiliza JWT (JSON Web Tokens) almacenados en cookies HttpOnly para la autenticación. Esto proporciona mayor seguridad que almacenar tokens en localStorage o sessionStorage.

### Roles de usuario

- **Admin**: Acceso completo a todas las funcionalidades
- **User**: Acceso para consultar y gestionar servidores y métricas
- **Viewer**: Acceso de solo lectura

### Usuario administrador por defecto

Al iniciar la aplicación por primera vez, se crea un usuario administrador por defecto:
- **Username**: admin
- **Password**: admin123 (configurable mediante la variable ADMIN_PASSWORD)

Se recomienda cambiar esta contraseña inmediatamente después del primer inicio de sesión.

## API Endpoints

### Autenticación

- `POST /api/auth/login` - Iniciar sesión y obtener token
- `POST /api/auth/logout` - Cerrar sesión
- `POST /api/auth/register` - Registrar un nuevo usuario
- `GET /api/auth/me` - Obtener información del usuario actual
- `POST /api/auth/change-password` - Cambiar contraseña

### Usuarios (solo admin)

- `GET /api/users` - Obtener todos los usuarios
- `GET /api/users/:id` - Obtener un usuario por ID
- `POST /api/users` - Crear un usuario
- `PUT /api/users/:id` - Actualizar un usuario
- `DELETE /api/users/:id` - Eliminar un usuario

### Servidores

- `GET /api/servers` - Obtener todos los servidores
- `GET /api/servers/:id` - Obtener un servidor por ID
- `POST /api/servers` - Crear un nuevo servidor (requiere admin o user)
- `PUT /api/servers/:id` - Actualizar un servidor (requiere admin o user)
- `DELETE /api/servers/:id` - Eliminar un servidor (requiere admin o user)

### Métricas

- `POST /api/metrics` - Crear una nueva métrica
- `GET /api/metrics/server/:server_id` - Obtener métricas por ID de servidor
- `GET /api/metrics/server/:server_id/latest` - Obtener la última métrica de un servidor
- `GET /api/metrics/server/:server_id/timerange` - Obtener métricas por rango de tiempo

### Logs (solo admin)

- `GET /api/logs` - Obtener logs con filtros (nivel, fuente, fecha)
- `DELETE /api/logs/cleanup` - Eliminar logs antiguos

## Ejemplos de uso

### Iniciar sesión

```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  --cookie-jar cookies.txt \
  -d '{
    "username": "admin",
    "password": "admin123"
  }'
```

### Crear un servidor (utilizando cookie de autenticación)

```bash
curl -X POST http://localhost:8080/api/servers \
  -H "Content-Type: application/json" \
  --cookie cookies.txt \
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
  --cookie cookies.txt \
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
curl "http://localhost:8080/api/metrics/server/1/timerange?start=2023-10-01T00:00:00Z&end=2023-10-02T23:59:59Z" \
  --cookie cookies.txt
```

### Consultar logs (solo admin)

```bash
# Obtener los últimos 50 logs de nivel ERROR
curl "http://localhost:8080/api/logs?level=ERROR&limit=50" \
  --cookie cookies.txt

# Obtener logs por fuente y rango de fechas
curl "http://localhost:8080/api/logs?source=system&start_date=2023-10-01T00:00:00Z&end_date=2023-10-02T23:59:59Z" \
  --cookie cookies.txt

# Limpiar logs antiguos (más de 30 días)
curl -X DELETE "http://localhost:8080/api/logs/cleanup?days=30" \
  --cookie cookies.txt
``` 