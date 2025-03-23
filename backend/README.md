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
- Transmisión de métricas en tiempo real mediante WebSockets
- Control de acceso basado en roles (RBAC)
- Escalabilidad horizontal con Redis Pub/Sub
- Sistema de alertas basado en umbrales configurables
- Notificaciones a través de Discord, Email y Webhooks personalizados

## Requisitos

- Go 1.18 o superior
- PostgreSQL 12 o superior
- Redis (opcional, para escalabilidad horizontal)

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

# Configuración de Redis para WebSockets
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0
REDIS_ENABLED=false

# Configuración de WebSockets
WS_PING_INTERVAL=30
WS_ALLOWED_ORIGINS=*

# Configuración de notificaciones para alertas
EMAIL_ENABLED=false
EMAIL_FROM=alertas@sistema.local
EMAIL_SMTP=smtp.example.com
EMAIL_PORT=587
EMAIL_USER=
EMAIL_PASSWORD=
DISCORD_ENABLED=false
DISCORD_WEBHOOK_URL=
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
  │   ├── interfaces/      # Interfaces para evitar dependencias circulares
  │   ├── logger/          # Sistema de logging
  │   ├── database/        # Conexión a la base de datos
  │   ├── websocket/       # Sistema de WebSockets para tiempo real
```

## WebSockets para métricas en tiempo real

El sistema implementa WebSockets para transmitir métricas en tiempo real, eliminando la necesidad de polling y mejorando la experiencia del usuario.

### Arquitectura de WebSockets

- **Hub**: Gestiona todas las conexiones activas y distribuye mensajes
- **Client**: Representa una conexión individual y maneja lectura/escritura
- **Interfaces**: Evita dependencias circulares entre paquetes
- **Redis Pub/Sub**: Opcional para escalabilidad horizontal con múltiples instancias

### Puntos de conexión WebSocket

- `ws://servidor/api/metrics/live/:server_id` - Transmite métricas en tiempo real para un servidor específico

### Autenticación para WebSockets

Las conexiones WebSocket requieren autenticación. Hay dos formas de proporcionar el token JWT:

1. **Headers**: Enviar el token en el header `Authorization: Bearer <token>`
2. **Query Parameter**: Añadir `?token=<token>` a la URL de conexión

### Comunicación

- El servidor envía actualizaciones de métricas a todos los clientes conectados para un servidor específico
- Se implementa ping/pong para mantener las conexiones activas
- Las desconexiones se detectan y manejan automáticamente

### Configuración WebSockets

En el archivo `.env` se pueden configurar:
```env
WS_PING_INTERVAL=30 # Intervalo de ping en segundos
WS_ALLOWED_ORIGINS=* # Orígenes permitidos (separados por coma)
```

### Escalabilidad con Redis Pub/Sub

Para entornos con múltiples instancias del backend, se puede habilitar Redis para distribuir las métricas:

```env
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0
REDIS_ENABLED=false # Cambiar a true para habilitar Redis
```

### Ejemplo de cliente JavaScript

```javascript
// Conexión al WebSocket con autenticación
const token = obtenerTokenJWT();
const ws = new WebSocket(`ws://localhost:8080/api/metrics/live/1?token=${token}`);

// Manejo de eventos
ws.onopen = () => console.log('Conexión establecida');
ws.onclose = () => console.log('Conexión cerrada');
ws.onerror = (error) => console.error('Error de WebSocket:', error);

// Recepción de métricas
ws.onmessage = (event) => {
  const metrica = JSON.parse(event.data);
  console.log('Nueva métrica recibida:', metrica);
  // Actualizar UI con la nueva métrica
};
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

## Sistema de Alertas y Umbrales

El sistema implementa un mecanismo de alertas basado en umbrales configurables para métricas de servidores. Cada vez que se recibe una nueva métrica, se evalúa contra los umbrales definidos y se generan alertas cuando corresponde.

### Funcionamiento

1. **Umbrales configurables**: Define condiciones como CPU > 90%, memoria > 80%, etc.
2. **Evaluación automática**: Cada nueva métrica se verifica contra los umbrales aplicables
3. **Generación de alertas**: Se crean alertas cuando los valores superan los umbrales establecidos
4. **Notificaciones**: Envío de notificaciones por canales configurados (Discord, Email, Webhooks)
5. **Resolución automática**: Las alertas se resuelven automáticamente cuando los valores vuelven a la normalidad

### Tipos de métricas monitorizables

- **CPU**: Porcentaje de uso de CPU
- **Memoria**: Porcentaje de uso de memoria
- **Disco**: Porcentaje de uso de disco
- **Red (entrada)**: Tráfico de red entrante
- **Red (salida)**: Tráfico de red saliente

### Configuración de umbrales

Los umbrales se pueden configurar con:

- **Severidad**: Info, Warning, Critical
- **Operador**: >, <, >=, <=, ==
- **Valor**: Umbral numérico
- **Cooldown**: Tiempo mínimo entre alertas (evita tormentas de alertas)
- **Canales de notificación**: Discord, Email, Webhook personalizado
- **Alcance**: Por servidor específico, por grupo de servidores o global

### Notificaciones

El sistema puede enviar notificaciones por diferentes canales:

1. **Discord**: Mediante webhooks de Discord con mensajes formateados
2. **Email**: A través de SMTP (pendiente de implementar completamente)
3. **Webhooks**: Para integración con sistemas externos

Para habilitar Discord, configura:
```env
DISCORD_ENABLED=true
DISCORD_WEBHOOK_URL=https://discord.com/api/webhooks/tu_webhook_url
```

### Estados de alertas

- **Active**: La alerta está activa y sin atender
- **Acknowledged**: La alerta ha sido reconocida pero no resuelta
- **Resolved**: La alerta ha sido resuelta (manual o automáticamente)
- **Suppressed**: La alerta está temporalmente suprimida

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
- `GET /api/metrics/live/:server_id` - **WebSocket** para métricas en tiempo real

### Logs (solo admin)

- `GET /api/logs` - Obtener logs con filtros (nivel, fuente, fecha)
- `DELETE /api/logs/cleanup` - Eliminar logs antiguos

### Alertas

- `GET /api/alerts` - Obtener todas las alertas (con filtros opcionales)
- `GET /api/alerts/active` - Obtener solo alertas activas
- `GET /api/alerts/:id` - Obtener una alerta por ID
- `POST /api/alerts/:id/acknowledge` - Reconocer una alerta
- `POST /api/alerts/:id/resolve` - Resolver una alerta manualmente

### Umbrales

- `GET /api/alerts/thresholds` - Obtener todos los umbrales configurados
- `GET /api/alerts/thresholds/:id` - Obtener un umbral por ID
- `POST /api/alerts/thresholds` - Crear un nuevo umbral (requiere admin)
- `PUT /api/alerts/thresholds/:id` - Actualizar un umbral (requiere admin)
- `DELETE /api/alerts/thresholds/:id` - Eliminar un umbral (requiere admin)
- `GET /api/alerts/thresholds/server/:server_id` - Obtener umbrales aplicables a un servidor

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

### Crear un umbral de alerta para CPU

```bash
curl -X POST http://localhost:8080/api/alerts/thresholds \
  -H "Content-Type: application/json" \
  --cookie cookies.txt \
  -d '{
    "name": "CPU crítico",
    "description": "Alerta cuando la CPU supera el 90%",
    "metric_type": "cpu",
    "operator": ">",
    "value": 90.0,
    "severity": "critical",
    "enable_discord": true,
    "cooldown_minutes": 15,
    "server_id": 1
  }'
```

### Obtener alertas activas

```bash
curl "http://localhost:8080/api/alerts/active" \
  --cookie cookies.txt
```

### Reconocer una alerta

```bash
curl -X POST http://localhost:8080/api/alerts/1/acknowledge \
  -H "Content-Type: application/json" \
  --cookie cookies.txt \
  -d '{
    "notes": "Investigando el problema de CPU"
  }' 