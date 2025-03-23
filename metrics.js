/**
 * Servicio para gestión de métricas y monitoreo en tiempo real
 */
const metricsService = {
  activeWebSocket: null,
  activeServerId: null,

  /**
   * Obtiene las métricas más recientes de un servidor
   * @param {number} serverId - ID del servidor
   * @returns {Promise<Object>} Métricas del servidor
   */
  async getServerMetrics(serverId) {
    try {
      console.log(`Obteniendo métricas para servidor ID: ${serverId}`);
      const data = await apiClient.get(`${API_METRICS_URL}/server/${serverId}/latest`);
      console.log("Métricas del servidor:", data);

      // Aquí se podría mostrar las métricas en una ventana modal o en otra sección
      alert(
        `Métricas más recientes del servidor ${serverId}\n\n${JSON.stringify(
          data,
          null,
          2
        )}`
      );
      
      return data;
    } catch (error) {
      console.error("Error al obtener métricas:", error);
      throw error;
    }
  },

  /**
   * Obtiene métricas históricas por rango de tiempo
   * @param {number} serverId - ID del servidor
   * @param {Date} startDate - Fecha inicial
   * @param {Date} endDate - Fecha final
   * @returns {Promise<Array>} Lista de métricas
   */
  async getMetricsByTimeRange(serverId, startDate, endDate) {
    try {
      const params = {
        start: startDate.toISOString(),
        end: endDate.toISOString()
      };
      
      return await apiClient.get(
        `${API_METRICS_URL}/server/${serverId}/timerange`, 
        params
      );
    } catch (error) {
      console.error("Error al obtener métricas por rango de tiempo:", error);
      throw error;
    }
  },

  /**
   * Inicia monitoreo en tiempo real para un servidor
   * @param {number} serverId - ID del servidor
   */
  startRealTimeMonitoring(serverId) {
    // Cerrar WebSocket previo si existe
    this.closeWebSocket();
    
    console.log(`Iniciando monitoreo en tiempo real para servidor ID: ${serverId}`);
    
    // Encontrar y mostrar información del servidor
    const server = serverService.serversData.find(s => s.id === serverId);
    if (server) {
      document.getElementById('serverInfo').innerHTML = `
        <p><strong>Servidor:</strong> ${server.hostname || 'Sin nombre'}</p>
        <p><strong>IP:</strong> ${server.ip || 'N/A'}</p>
        <p><strong>Descripción:</strong> ${server.description || 'Sin descripción'}</p>
        ${server.os ? `<p><strong>Sistema:</strong> ${server.os} ${server.os_version || ''}</p>` : ''}
        ${server.location ? `<p><strong>Ubicación:</strong> ${server.location}</p>` : ''}
      `;
    }
    
    // Mostrar sección de métricas
    document.getElementById('metricsSection').classList.remove('hidden');
    this.activeServerId = serverId;
    
    // Inicializar los datos de las gráficas
    chartsService.resetCharts();
    
    // Construir la URL de WebSocket
    const wsProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsUrl = `${wsProtocol}//localhost:8080/api/metrics/live/${serverId}`;
    
    // Obtener token para autenticación
    this.connectWebSocket(wsUrl, serverId);
  },

  /**
   * Establece la conexión WebSocket
   * @param {string} wsUrl - URL del WebSocket
   * @param {number} serverId - ID del servidor
   */
  connectWebSocket(wsUrl, serverId) {
    try {
      // Extraer token de la cookie para autenticación
      const authToken = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxLCJ1c2VybmFtZSI6ImFkbWluIiwiZW1haWwiOiJhZG1pbkBzaXN0ZW1hLmxvY2FsIiwicm9sZSI6IkFETUlOIiwic3ViIjoiMSIsImV4cCI6MjA1MzcyMTc2NiwiaWF0IjoxNzQyNjgxNzY2fQ.odU-p2iHLAhmHTeftmSYqXR1tJF8teSqePGMxCJmNUk'
      console.log(`Conectando a WebSocket: ${wsUrl}`);
      
      // Crear conexión WebSocket con autenticación
      this.activeWebSocket = new WebSocket(`${wsUrl}?token=${authToken}`);
      
      this.activeWebSocket.onopen = (event) => {
        console.log('Conexión WebSocket establecida:', event);
        document.getElementById('socketStatus').textContent = 'Conectado';
        document.getElementById('socketStatus').className = 'px-2 py-1 bg-green-100 text-green-800 rounded-full text-xs';
      };

      this.activeWebSocket.onmessage = (event) => {
        try {
          const metrics = JSON.parse(event.data);
          console.log("Nuevas métricas recibidas:", metrics);
          this.updateMetricsDisplay(metrics);
        } catch (error) {
          console.error("Error al procesar métricas:", error);
        }
      };

      this.activeWebSocket.onerror = (error) => {
        console.error("Error de WebSocket:", error);
        document.getElementById("socketStatus").textContent = "Error";
        document.getElementById("socketStatus").className =
          "px-2 py-1 bg-red-100 text-red-800 rounded-full text-xs";
      };

      this.activeWebSocket.onclose = (event) => {
        console.log("Conexión WebSocket cerrada:", event);
        document.getElementById("socketStatus").textContent = "Desconectado";
        document.getElementById("socketStatus").className =
          "px-2 py-1 bg-gray-200 text-gray-700 rounded-full text-xs";

        // Si no fue cerrado manualmente, intentar reconectar
        if (this.activeServerId && event.code !== 1000) {
          console.log("Intentando reconectar en 3 segundos...");
          setTimeout(() => {
            if (this.activeServerId) {
              this.startRealTimeMonitoring(this.activeServerId);
            }
          }, 3000);
        }
      };
    } catch (error) {
      console.error("Error al crear conexión WebSocket:", error);
    }
  },

  /**
   * Cierra la conexión WebSocket actual
   */
  closeWebSocket() {
    if (this.activeWebSocket) {
      console.log("Cerrando conexión WebSocket");
      this.activeWebSocket.close(1000, "Cerrado por el usuario");
      this.activeWebSocket = null;
      this.activeServerId = null;
    }
  },

  /**
   * Obtiene el valor de una cookie por su nombre
   * @param {string} name - Nombre de la cookie
   * @returns {string|null} Valor de la cookie o null
   */
  getCookieValue(name) {
    const regex = new RegExp(`(^| )${name}=([^;]+)`);
    const match = document.cookie.match(regex);
    if (match) {
      return match[2];
    }
    return null;
  },

  /**
   * Actualiza la visualización de métricas en la UI
   * @param {Object} metrics - Datos de métricas
   */
  updateMetricsDisplay(metrics) {
    const timestamp = new Date().toLocaleTimeString();

    // Actualizar valores actuales
    document.getElementById(
      "cpuValue"
    ).textContent = `${metrics.cpu_usage.toFixed(1)}%`;

    // Calcular porcentaje de memoria
    const memUsedPercent = (
      (metrics.memory_used / metrics.memory_total) *
      100
    ).toFixed(1);
    document.getElementById("memValue").textContent = `${memUsedPercent}%`;

    // Calcular porcentaje de disco
    const diskUsedPercent = (
      (metrics.disk_used / metrics.disk_total) *
      100
    ).toFixed(1);
    document.getElementById("diskValue").textContent = `${diskUsedPercent}%`;

    // Actualizar valor de red (nombres actualizados)
    if (metrics.net_upload !== undefined && metrics.net_download !== undefined) {
      const netIn = this.formatBytes(metrics.net_download);
      const netOut = this.formatBytes(metrics.net_upload);
      document.getElementById(
        "networkValue"
      ).textContent = `↓${netIn}/s ↑${netOut}/s`;
    }

    // Actualizar detalles
    this.updateDetailedMetrics(metrics);

    // Actualizar gráficos
    chartsService.updateCharts(timestamp, metrics.cpu_usage, memUsedPercent, 
      diskUsedPercent, metrics.net_download, metrics.net_upload);
  },

  /**
   * Actualiza los detalles de métricas en la UI
   * @param {Object} metrics - Datos de métricas
   */
  updateDetailedMetrics(metrics) {
    // Actualizar detalles de memoria
    const memTotal = this.formatBytes(metrics.memory_total);
    const memUsed = this.formatBytes(metrics.memory_used);
    const memFree = this.formatBytes(metrics.memory_free);
    
    document.getElementById("memDetails").textContent = 
      `${memUsed} usado de ${memTotal}`;
    
    // Actualizar detalles de disco
    const diskTotal = this.formatBytes(metrics.disk_total);
    const diskUsed = this.formatBytes(metrics.disk_used);
    const diskFree = this.formatBytes(metrics.disk_free);
    
    document.getElementById("diskDetails").textContent = 
      `${diskUsed} usado de ${diskTotal}`;
    
    // Actualizar detalles de CPU
    let cpuDetails = `Carga: ${metrics.load_avg_1 || 'N/A'} (1m)`;
    if (metrics.cpu_temp) {
      cpuDetails += `, Temp: ${metrics.cpu_temp.toFixed(1)}°C`;
    }
    if (metrics.cpu_freq) {
      cpuDetails += `, ${(metrics.cpu_freq/1000).toFixed(2)} GHz`;
    }
    
    document.getElementById("cpuDetails").textContent = cpuDetails;
    
    // Actualizar detalles del sistema
    if (document.getElementById("systemDetails")) {
      const uptimeStr = metrics.uptime ? this.formatUptime(metrics.uptime) : 'N/A';
      
      document.getElementById("systemDetails").innerHTML = `
        <div>Tiempo activo: ${uptimeStr}</div>
        ${metrics.cpu_cores ? `<div>CPU Cores: ${metrics.cpu_cores}</div>` : ''}
        ${metrics.cpu_threads ? `<div>CPU Threads: ${metrics.cpu_threads}</div>` : ''}
      `;
    }
    
    // Actualizar detalles de procesos
    if (document.getElementById("processDetails")) {
      document.getElementById("processDetails").innerHTML = `
        ${metrics.process_count ? `<div>Procesos: ${metrics.process_count}</div>` : ''}
        ${metrics.thread_count ? `<div>Threads: ${metrics.thread_count}</div>` : ''}
        ${metrics.handle_count ? `<div>Handles: ${metrics.handle_count}</div>` : ''}
      `;
    }
  },

  /**
   * Formatea bytes en unidades legibles
   * @param {number} bytes - Valor en bytes
   * @param {number} decimals - Decimales a mostrar
   * @returns {string} Valor formateado
   */
  formatBytes(bytes, decimals = 1) {
    if (bytes === 0) return "0 B";

    const k = 1024;
    const sizes = ["B", "KB", "MB", "GB", "TB"];
    const i = Math.floor(Math.log(bytes) / Math.log(k));

    return (
      parseFloat((bytes / Math.pow(k, i)).toFixed(decimals)) + " " + sizes[i]
    );
  },

  /**
   * Formatea segundos en formato de tiempo activo
   * @param {number} seconds - Segundos
   * @returns {string} Tiempo formateado
   */
  formatUptime(seconds) {
    const days = Math.floor(seconds / 86400);
    const hours = Math.floor((seconds % 86400) / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);
    
    let result = "";
    if (days > 0) result += `${days}d `;
    if (hours > 0 || days > 0) result += `${hours}h `;
    result += `${minutes}m`;
    
    return result;
  }
}; 