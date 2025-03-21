package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"net/http/cookiejar"
	"time"
)

const (
	baseURL    = "http://localhost:8080/api"
	username   = "admin"
	password   = "Argentina1"
	numServers = 3
)

// Servidor representa un servidor en la API
type Server struct {
	ID          int    `json:"id"`
	Hostname    string `json:"hostname"`
	IP          string `json:"ip"`
	Description string `json:"description"`
	Status      string `json:"status"`
}

// Métrica representa los datos de monitoreo de un servidor
type Metric struct {
	ServerID    int     `json:"server_id"`
	CPUUsage    float64 `json:"cpu_usage"`
	MemoryTotal int64   `json:"memory_total"`
	MemoryUsed  int64   `json:"memory_used"`
	MemoryFree  int64   `json:"memory_free"`
	DiskTotal   int64   `json:"disk_total"`
	DiskUsed    int64   `json:"disk_used"`
	DiskFree    int64   `json:"disk_free"`
	NetworkIn   int64   `json:"network_in,omitempty"`
	NetworkOut  int64   `json:"network_out,omitempty"`
	Timestamp   string  `json:"timestamp,omitempty"`
}

// Credenciales para login
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Cliente HTTP con cookies
var client *http.Client

// Servidores creados
var servers []Server

func init() {
	// Inicializar el generador de números aleatorios
	rand.Seed(time.Now().UnixNano())

	// Crear un jar de cookies para mantener la sesión
	jar, err := cookiejar.New(nil)
	if err != nil {
		panic(err)
	}

	// Configurar el cliente HTTP
	client = &http.Client{
		Jar:     jar,
		Timeout: 10 * time.Second,
	}
}

func main() {
	fmt.Println("🚀 Iniciando simulador de métricas para debugging")

	// Login para obtener cookie de autenticación
	err := login(username, password)
	if err != nil {
		fmt.Printf("Error en login: %v\n", err)
		return
	}
	fmt.Println("✅ Login exitoso")

	// Obtener o crear servidores
	servers, err = setupServers(numServers)
	if err != nil {
		fmt.Printf("Error configurando servidores: %v\n", err)
		return
	}
	fmt.Printf("✅ %d servidores configurados correctamente\n", len(servers))

	// Mostrar los servidores creados
	for _, server := range servers {
		fmt.Printf("   🖥️  Servidor #%d: %s (%s)\n", server.ID, server.Hostname, server.IP)
	}

	// Intervalo para enviar métricas
	interval := 2 * time.Second
	fmt.Printf("📊 Enviando métricas cada %v\n", interval)
	fmt.Println("⏱️  Presiona Ctrl+C para detener")

	// Ciclo infinito para enviar métricas
	ticker := time.NewTicker(interval)
	for range ticker.C {
		for _, server := range servers {
			metric := generateRandomMetric(server.ID)
			err := sendMetric(metric)
			if err != nil {
				fmt.Printf("❌ Error enviando métrica para servidor %d: %v\n", server.ID, err)
			} else {
				fmt.Printf("📤 Métrica enviada para servidor %d: CPU %.1f%%, MEM %.1f%%\n",
					server.ID,
					metric.CPUUsage,
					float64(metric.MemoryUsed)/float64(metric.MemoryTotal)*100)
			}
		}
		fmt.Println("---")
	}
}

// Login en la API para obtener cookie de autenticación
func login(username, password string) error {
	loginData := LoginRequest{
		Username: username,
		Password: password,
	}

	jsonData, err := json.Marshal(loginData)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", baseURL+"/auth/login", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("login fallido: código de estado %d", resp.StatusCode)
	}

	return nil
}

// Obtener o crear servidores para simulación
func setupServers(count int) ([]Server, error) {
	// Primero intentamos obtener los servidores existentes
	req, err := http.NewRequest("GET", baseURL+"/servers", nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		// Intentar decodificar como un array directamente
		var servers []Server
		if err := json.NewDecoder(resp.Body).Decode(&servers); err != nil {
			// Si falla, intentar como estructura con campo "servers"
			resp.Body.Close()

			// Hacer una nueva solicitud ya que el cuerpo se ha consumido
			req, err := http.NewRequest("GET", baseURL+"/servers", nil)
			if err != nil {
				return nil, err
			}

			resp, err := client.Do(req)
			if err != nil {
				return nil, err
			}
			defer resp.Body.Close()

			var response struct {
				Servers []Server `json:"servers"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
				return nil, fmt.Errorf("error decodificando respuesta: %v", err)
			}
			servers = response.Servers
		}

		// Si ya hay suficientes servidores, los usamos
		if len(servers) >= count {
			fmt.Printf("✅ Usando %d servidores existentes\n", count)
			return servers[:count], nil
		}

		// Si hay algunos pero no suficientes, usamos los que hay y creamos el resto
		if len(servers) > 0 {
			count = count - len(servers)
			createdServers, err := createServers(count)
			if err != nil {
				return nil, err
			}
			return append(servers, createdServers...), nil
		}
	}

	// Si no hay servidores, creamos nuevos
	return createServers(count)
}

// Función auxiliar para crear servidores
func createServers(count int) ([]Server, error) {
	var createdServers []Server
	for i := 1; i <= count; i++ {
		server := Server{
			Hostname:    fmt.Sprintf("servidor-debug-%d", i),
			IP:          fmt.Sprintf("192.168.1.%d", i+100),
			Description: fmt.Sprintf("Servidor de prueba #%d para debugging", i),
			Status:      "online",
		}

		jsonData, err := json.Marshal(server)
		if err != nil {
			return nil, err
		}

		req, err := http.NewRequest("POST", baseURL+"/servers", bytes.NewBuffer(jsonData))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}

		// Leer y decodificar la respuesta para obtener el ID asignado
		var createdServer Server
		if err := json.NewDecoder(resp.Body).Decode(&createdServer); err != nil {
			resp.Body.Close()
			return nil, fmt.Errorf("error decodificando servidor creado: %v", err)
		}
		resp.Body.Close()

		if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("error al crear servidor %d: código de estado %d", i, resp.StatusCode)
		}

		createdServers = append(createdServers, createdServer)
		fmt.Printf("✅ Servidor creado: %s (ID: %d)\n", createdServer.Hostname, createdServer.ID)
	}

	return createdServers, nil
}

// Generar métricas aleatorias para un servidor
func generateRandomMetric(serverID int) Metric {
	// Valores base para cada servidor (para que tengan patrones distintos)
	baseCPU := 10.0 + float64(serverID*5)
	baseMemTotal := int64(8+serverID*4) * 1024 * 1024 * 1024     // GB en bytes
	baseDiskTotal := int64(100+serverID*50) * 1024 * 1024 * 1024 // GB en bytes

	// Fluctuación aleatoria
	cpuFluctuation := rand.Float64() * 30.0 // Fluctuación de hasta 30%
	cpuUsage := baseCPU + cpuFluctuation
	if cpuUsage > 100.0 {
		cpuUsage = 99.9
	}

	// Memoria
	memoryUsed := int64(float64(baseMemTotal) * (0.3 + rand.Float64()*0.5)) // Entre 30% y 80% de uso
	memoryFree := baseMemTotal - memoryUsed

	// Disco
	diskUsed := int64(float64(baseDiskTotal) * (0.2 + rand.Float64()*0.6)) // Entre 20% y 80% de uso
	diskFree := baseDiskTotal - diskUsed

	// Red (tráfico por segundo)
	networkIn := rand.Int63n(10 * 1024 * 1024) // 0-10 MB/s
	networkOut := rand.Int63n(5 * 1024 * 1024) // 0-5 MB/s

	return Metric{
		ServerID:    serverID,
		CPUUsage:    cpuUsage,
		MemoryTotal: baseMemTotal,
		MemoryUsed:  memoryUsed,
		MemoryFree:  memoryFree,
		DiskTotal:   baseDiskTotal,
		DiskUsed:    diskUsed,
		DiskFree:    diskFree,
		NetworkIn:   networkIn,
		NetworkOut:  networkOut,
		Timestamp:   time.Now().Format(time.RFC3339),
	}
}

// Enviar métrica al servidor
func sendMetric(metric Metric) error {
	jsonData, err := json.Marshal(metric)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", baseURL+"/metrics", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error al enviar métrica: código de estado %d", resp.StatusCode)
	}

	return nil
}
