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
	password   = "admin123"
	numServers = 3
)

// Servidor representa un servidor en la API
type Server struct {
	ID          int    `json:"id"`
	Hostname    string `json:"hostname"`
	IP          string `json:"ip"`
	Description string `json:"description"`
	IsActive    bool   `json:"is_active"`

	// Información del sistema operativo
	OS        string `json:"os"`
	OSVersion string `json:"os_version"`
	OSArch    string `json:"os_arch"`
	Kernel    string `json:"kernel"`

	// Información de hardware
	CPUModel    string `json:"cpu_model"`
	CPUCores    int    `json:"cpu_cores"`
	CPUThreads  int    `json:"cpu_threads"`
	TotalMemory int64  `json:"total_memory"`
	TotalDisk   int64  `json:"total_disk"`
}

// Métrica representa los datos de monitoreo de un servidor
type Metric struct {
	ServerID  int    `json:"server_id"`
	Timestamp string `json:"timestamp,omitempty"`

	// Métricas de CPU
	CPUUsage  float64 `json:"cpu_usage"`
	CPUTemp   float64 `json:"cpu_temp,omitempty"`
	CPUFreq   float64 `json:"cpu_freq,omitempty"`
	LoadAvg1  float64 `json:"load_avg_1,omitempty"`
	LoadAvg5  float64 `json:"load_avg_5,omitempty"`
	LoadAvg15 float64 `json:"load_avg_15,omitempty"`

	// Métricas de Memoria
	MemoryTotal   int64 `json:"memory_total"`
	MemoryUsed    int64 `json:"memory_used"`
	MemoryFree    int64 `json:"memory_free"`
	MemoryCache   int64 `json:"memory_cache,omitempty"`
	MemoryBuffers int64 `json:"memory_buffers,omitempty"`
	SwapTotal     int64 `json:"swap_total,omitempty"`
	SwapUsed      int64 `json:"swap_used,omitempty"`
	SwapFree      int64 `json:"swap_free,omitempty"`

	// Métricas de Disco
	DiskTotal      int64 `json:"disk_total"`
	DiskUsed       int64 `json:"disk_used"`
	DiskFree       int64 `json:"disk_free"`
	DiskReads      int64 `json:"disk_reads,omitempty"`
	DiskWrites     int64 `json:"disk_writes,omitempty"`
	DiskReadBytes  int64 `json:"disk_read_bytes,omitempty"`
	DiskWriteBytes int64 `json:"disk_write_bytes,omitempty"`
	DiskIOTime     int64 `json:"disk_io_time,omitempty"`

	// Métricas de Red
	NetUpload     int64 `json:"net_upload"`
	NetDownload   int64 `json:"net_download"`
	NetPacketsIn  int64 `json:"net_packets_in,omitempty"`
	NetPacketsOut int64 `json:"net_packets_out,omitempty"`
	NetErrorsIn   int64 `json:"net_errors_in,omitempty"`
	NetErrorsOut  int64 `json:"net_errors_out,omitempty"`
	NetDropsIn    int64 `json:"net_drops_in,omitempty"`
	NetDropsOut   int64 `json:"net_drops_out,omitempty"`

	// Procesos y servicios
	ProcessCount int `json:"process_count,omitempty"`
	ThreadCount  int `json:"thread_count,omitempty"`
	HandleCount  int `json:"handle_count,omitempty"`

	// Tiempo de actividad
	Uptime int64 `json:"uptime,omitempty"`
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
		fmt.Printf("   🖥️  Servidor #%d: %s (%s) - %s\n",
			server.ID,
			server.Hostname,
			server.IP,
			server.OS)
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
				fmt.Printf("📤 Métrica enviada para servidor %d: CPU %.1f%%, MEM %.1f%%, DISK %.1f%%\n",
					server.ID,
					metric.CPUUsage,
					float64(metric.MemoryUsed)/float64(metric.MemoryTotal)*100,
					float64(metric.DiskUsed)/float64(metric.DiskTotal)*100)
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
		// Generar diferentes tipos de servidores
		osType := "Linux"
		osVersion := "Ubuntu 22.04"
		kernel := "5.15.0-56-generic"
		arch := "x86_64"
		cpuModel := "Intel(R) Xeon(R) CPU E5-2680 v4 @ 2.40GHz"
		cpuCores := 4 + i
		cpuThreads := 8 + i*2

		// Para servidores pares, usar Windows
		if i%2 == 0 {
			osType = "Windows"
			osVersion = "Windows Server 2022"
			kernel = "10.0.20348"
			arch = "x64"
			cpuModel = "Intel(R) Xeon(R) Gold 6240R CPU @ 2.40GHz"
		}

		memTotal := int64(8+i*4) * 1024 * 1024 * 1024
		diskTotal := int64(100+i*50) * 1024 * 1024 * 1024

		server := Server{
			Hostname:    fmt.Sprintf("servidor-debug-%d", i),
			IP:          fmt.Sprintf("192.168.1.%d", i+100),
			Description: fmt.Sprintf("Servidor de prueba #%d para debugging", i),
			IsActive:    true,
			OS:          osType,
			OSVersion:   osVersion,
			OSArch:      arch,
			Kernel:      kernel,
			CPUModel:    cpuModel,
			CPUCores:    cpuCores,
			CPUThreads:  cpuThreads,
			TotalMemory: memTotal,
			TotalDisk:   diskTotal,
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
		fmt.Printf("✅ Servidor creado: %s (ID: %d, OS: %s)\n", createdServer.Hostname, createdServer.ID, createdServer.OS)
	}

	return createdServers, nil
}

// Generar métricas aleatorias para un servidor
func generateRandomMetric(serverID int) Metric {
	// Detectar servidor en la lista
	var server Server
	for _, s := range servers {
		if s.ID == serverID {
			server = s
			break
		}
	}

	// Valores base para cada servidor (para que tengan patrones distintos)
	baseCPU := 10.0 + float64(serverID*5)
	baseMemTotal := server.TotalMemory
	baseDiskTotal := server.TotalDisk

	// Fluctuación aleatoria
	cpuFluctuation := rand.Float64() * 30.0 // Fluctuación de hasta 30%
	cpuUsage := baseCPU + cpuFluctuation
	if cpuUsage > 100.0 {
		cpuUsage = 99.9
	}

	// CPU Temperatura y frecuencia
	cpuTemp := 35.0 + rand.Float64()*20.0
	cpuFreq := 2400.0 + rand.Float64()*600.0

	// Cargas promedio (más relevantes en Linux pero las generamos igual)
	loadAvg1 := cpuUsage / 100.0 * float64(server.CPUCores) * (0.7 + rand.Float64()*0.6)
	loadAvg5 := loadAvg1 * (0.8 + rand.Float64()*0.4)
	loadAvg15 := loadAvg5 * (0.8 + rand.Float64()*0.4)

	// Memoria
	memoryUsed := int64(float64(baseMemTotal) * (0.3 + rand.Float64()*0.5)) // Entre 30% y 80% de uso
	memoryFree := baseMemTotal - memoryUsed
	memoryCache := int64(float64(memoryUsed) * 0.2)
	memoryBuffers := int64(float64(memoryUsed) * 0.1)

	// Swap
	swapTotal := baseMemTotal / 2
	swapUsed := int64(float64(swapTotal) * rand.Float64() * 0.3) // Uso de swap entre 0% y 30%
	swapFree := swapTotal - swapUsed

	// Disco
	diskUsed := int64(float64(baseDiskTotal) * (0.2 + rand.Float64()*0.6)) // Entre 20% y 80% de uso
	diskFree := baseDiskTotal - diskUsed

	// IO de disco
	diskReads := rand.Int63n(500)
	diskWrites := rand.Int63n(200)
	diskReadBytes := diskReads * 4096 * (1 + rand.Int63n(10))
	diskWriteBytes := diskWrites * 4096 * (1 + rand.Int63n(10))
	diskIOTime := (diskReads + diskWrites) * (5 + rand.Int63n(20))

	// Red
	netDownload := rand.Int63n(10 * 1024 * 1024) // 0-10 MB/s
	netUpload := rand.Int63n(5 * 1024 * 1024)    // 0-5 MB/s
	netPacketsIn := netDownload / (500 + rand.Int63n(1000))
	netPacketsOut := netUpload / (500 + rand.Int63n(1000))
	netErrorsIn := int64(float64(netPacketsIn) * 0.001 * rand.Float64()) // Errores ocasionales
	netErrorsOut := int64(float64(netPacketsOut) * 0.001 * rand.Float64())
	netDropsIn := int64(float64(netPacketsIn) * 0.002 * rand.Float64())
	netDropsOut := int64(float64(netPacketsOut) * 0.002 * rand.Float64())

	// Procesos y threads
	processCount := 100 + rand.Intn(150)
	threadCount := processCount * (3 + rand.Intn(5))
	handleCount := threadCount * (5 + rand.Intn(10))

	// Uptime (entre 1 hora y 90 días en segundos)
	uptime := 3600 + rand.Int63n(90*24*3600)

	return Metric{
		ServerID:       serverID,
		Timestamp:      time.Now().Format(time.RFC3339),
		CPUUsage:       cpuUsage,
		CPUTemp:        cpuTemp,
		CPUFreq:        cpuFreq,
		LoadAvg1:       loadAvg1,
		LoadAvg5:       loadAvg5,
		LoadAvg15:      loadAvg15,
		MemoryTotal:    baseMemTotal,
		MemoryUsed:     memoryUsed,
		MemoryFree:     memoryFree,
		MemoryCache:    memoryCache,
		MemoryBuffers:  memoryBuffers,
		SwapTotal:      swapTotal,
		SwapUsed:       swapUsed,
		SwapFree:       swapFree,
		DiskTotal:      baseDiskTotal,
		DiskUsed:       diskUsed,
		DiskFree:       diskFree,
		DiskReads:      diskReads,
		DiskWrites:     diskWrites,
		DiskReadBytes:  diskReadBytes,
		DiskWriteBytes: diskWriteBytes,
		DiskIOTime:     diskIOTime,
		NetUpload:      netUpload,
		NetDownload:    netDownload,
		NetPacketsIn:   netPacketsIn,
		NetPacketsOut:  netPacketsOut,
		NetErrorsIn:    netErrorsIn,
		NetErrorsOut:   netErrorsOut,
		NetDropsIn:     netDropsIn,
		NetDropsOut:    netDropsOut,
		ProcessCount:   processCount,
		ThreadCount:    threadCount,
		HandleCount:    handleCount,
		Uptime:         uptime,
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
