# Usamos Go 1.24 que es la última versión estable según el go.mod
FROM golang:1.24-alpine AS builder

# Instalamos dependencias necesarias
RUN apk add --no-cache git make build-base

# Configuramos el directorio de trabajo
WORKDIR /app

# Copiamos los archivos de dependencias
COPY go.mod go.sum ./

# Descargamos dependencias
RUN go mod download

# Copiamos el código fuente
COPY . .

# Compilamos la aplicación con optimizaciones
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o server .

# Imagen final más ligera
FROM alpine:3.19

# Instalamos certificados y dependencias
RUN apk --no-cache add ca-certificates tzdata curl

# Creamos un usuario no privilegiado
RUN adduser -D -g 'appuser' appuser

# Configuramos el directorio de trabajo
WORKDIR /app

# Copiamos el binario compilado desde la etapa builder
COPY --from=builder /app/server .

# Copiamos archivos de configuración si existen
COPY --from=builder /app/config ./config

# Creamos directorio para logs
RUN mkdir -p /app/logs && chown -R appuser:appuser /app

# Cambiamos al usuario no privilegiado
USER appuser

# Exponemos el puerto de la aplicación
EXPOSE 8080

# Punto de entrada
CMD ["./server"]