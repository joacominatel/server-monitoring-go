/**
 * Servicio para gestión de servidores
 */
const serverService = {
  serversData: [],

  /**
   * Obtiene la lista de servidores
   * @returns {Promise<Array>} Lista de servidores
   */
  async getServers() {
    try {
      console.log("Obteniendo lista de servidores...");
      const data = await apiClient.get(API_SERVER_URL);
      console.log("Servidores recibidos:", data);

      this.serversData = data.servers || data || [];
      return this.serversData;
    } catch (error) {
      console.error("Error al obtener servidores:", error);
      throw error;
    }
  },

  /**
   * Obtiene un servidor por su ID
   * @param {number} serverId - ID del servidor
   * @returns {Promise<Object>} Datos del servidor
   */
  async getServerById(serverId) {
    try {
      return await apiClient.get(`${API_SERVER_URL}/${serverId}`);
    } catch (error) {
      console.error(`Error al obtener servidor ID ${serverId}:`, error);
      throw error;
    }
  },

  /**
   * Crea un nuevo servidor
   * @param {Object} serverData - Datos del servidor
   * @returns {Promise<Object>} Servidor creado
   */
  async createServer(serverData) {
    try {
      const result = await apiClient.post(API_SERVER_URL, serverData);
      // Actualizar lista local de servidores
      await this.getServers();
      return result;
    } catch (error) {
      console.error("Error al crear servidor:", error);
      throw error;
    }
  },

  /**
   * Actualiza un servidor existente
   * @param {number} serverId - ID del servidor
   * @param {Object} serverData - Datos actualizados
   * @returns {Promise<Object>} Servidor actualizado
   */
  async updateServer(serverId, serverData) {
    try {
      const result = await apiClient.put(`${API_SERVER_URL}/${serverId}`, serverData);
      // Actualizar lista local de servidores
      await this.getServers();
      return result;
    } catch (error) {
      console.error(`Error al actualizar servidor ID ${serverId}:`, error);
      throw error;
    }
  },

  /**
   * Elimina un servidor
   * @param {number} serverId - ID del servidor
   * @returns {Promise<Object>} Resultado de la operación
   */
  async deleteServer(serverId) {
    try {
      const result = await apiClient.delete(`${API_SERVER_URL}/${serverId}`);
      // Actualizar lista local de servidores
      await this.getServers();
      return result;
    } catch (error) {
      console.error(`Error al eliminar servidor ID ${serverId}:`, error);
      throw error;
    }
  },

  /**
   * Obtiene los grupos de servidores
   * @returns {Promise<Array>} Lista de grupos
   */
  async getServerGroups() {
    try {
      return await apiClient.get(`${API_SERVER_URL}/groups`);
    } catch (error) {
      console.error("Error al obtener grupos de servidores:", error);
      throw error;
    }
  },

  /**
   * Renderiza la lista de servidores en el DOM
   */
  renderServers() {
    const serverList = document.getElementById("serverList");
    serverList.innerHTML = "";

    if (this.serversData.length === 0) {
      serverList.innerHTML =
        '<p class="text-gray-500">No hay servidores disponibles</p>';
      return;
    }

    console.log(`Renderizando ${this.serversData.length} servidores`);

    this.serversData.forEach((server) => {
      const serverCard = document.createElement("div");
      serverCard.className = "border rounded-lg p-4 bg-gray-50";

      // Determinar estado del servidor
      const status = server.is_active ? "online" : "offline";
      const statusClass = status === "online" 
        ? "bg-green-100 text-green-800" 
        : "bg-red-100 text-red-800";

      serverCard.innerHTML = `
        <div class="flex justify-between items-center">
          <h3 class="text-lg font-semibold">${server.hostname || "Sin nombre"}</h3>
          <span class="px-2 py-1 text-xs ${statusClass} rounded-full">
            ${status}
          </span>
        </div>
        <p class="text-gray-600 text-sm">${server.description || "Sin descripción"}</p>
        <div class="text-sm mt-2">
          <p><span class="font-medium">IP:</span> ${server.ip || "N/A"}</p>
          <p><span class="font-medium">ID:</span> ${server.id}</p>
          ${server.location ? `<p><span class="font-medium">Ubicación:</span> ${server.location}</p>` : ''}
          ${server.os ? `<p><span class="font-medium">Sistema:</span> ${server.os} ${server.os_version || ''}</p>` : ''}
        </div>
        <div class="mt-3 flex space-x-2">
          <button onclick="metricsService.getServerMetrics(${server.id})" 
                  class="text-xs bg-blue-500 text-white px-2 py-1 rounded hover:bg-blue-600">
            Ver métricas
          </button>
          <button onclick="metricsService.startRealTimeMonitoring(${server.id})" 
                  class="text-xs bg-green-500 text-white px-2 py-1 rounded hover:bg-green-600">
            Monitoreo en tiempo real
          </button>
        </div>
      `;

      serverList.appendChild(serverCard);
    });
  }
}; 