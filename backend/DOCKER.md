# Despliegue con Docker de la Plataforma de Monitoreo

Este documento describe cómo desplegar el backend de la plataforma de monitoreo utilizando Docker y Docker Compose.

## Requisitos

- Docker Engine 25.0+ 
- Docker Compose 2.24+

## Servicios incluidos

El despliegue incluye los siguientes servicios:

1. **app** - Aplicación backend en Go
2. **postgres** - Base de datos PostgreSQL 16
3. **redis** - Redis 7 para escalabilidad y WebSockets

## Iniciar los servicios

Para iniciar todos los servicios, ejecuta:

```bash
cd backend
docker compose up -d
```

Para ver los logs en tiempo real:

```bash
docker compose logs -f
```

## Configuración

Todas las variables de entorno están preconfiguradas en el archivo `docker-compose.yml`. 

Aspectos importantes:

1. **JWT_SECRET**: Considera cambiar este valor en producción
2. **ADMIN_PASSWORD**: Cambia esta contraseña después del primer inicio
3. **REDIS_ENABLED**: Establecido en `true` para aprovechar Redis en el entorno Docker

## Acceso a los servicios

- **API REST**: http://localhost:8080/api
- **WebSockets**: ws://localhost:8080/api/metrics/live/{server_id}
- **PostgreSQL**: localhost:5432
- **Redis**: localhost:6379

## Persistencia de datos

Los datos se almacenan en volúmenes Docker:

- `postgres_data`: Datos de PostgreSQL
- `redis_data`: Datos de Redis
- `./logs:/app/logs`: Logs de la aplicación

## Gestión de contenedores

### Detener los servicios

```bash
docker compose down
```

### Reconstruir imágenes

```bash
docker compose build --no-cache
```

### Reiniciar un servicio específico

```bash
docker compose restart app
```

## Solución de problemas

1. **Error de conexión a PostgreSQL**:
   - Verifica que el contenedor de postgres esté en estado "healthy"
   - Comprueba los logs: `docker compose logs postgres`

2. **Error de conexión a Redis**:
   - Verifica que el contenedor de redis esté en estado "healthy"
   - Comprueba los logs: `docker compose logs redis`

3. **La aplicación no arranca**:
   - Revisa los logs: `docker compose logs app`
   - Verifica que PostgreSQL y Redis estén disponibles

## Monitoreo

Puedes monitorear el estado de los servicios con:

```bash
docker compose ps
```

Para verificar el estado de salud de los contenedores:

```bash
docker inspect --format "{{.Name}} - {{.State.Health.Status}}" $(docker compose ps -q)
```

## Seguridad

Para entornos de producción, considera:

1. Cambiar las credenciales por defecto
2. Usar secretos de Docker para gestionar contraseñas
3. Restringir la exposición de puertos
4. Configurar SSL/TLS para las conexiones 