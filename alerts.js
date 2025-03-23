/**
 * Servicio para gestión de alertas y umbrales
 */
const alertsService = {
  activeAlerts: [],
  allAlerts: [],
  thresholds: [],

  /**
   * Obtiene todas las alertas activas
   * @param {Object} filters - Filtros opcionales (serverID, severity, etc.)
   * @returns {Promise<Array>} Lista de alertas activas
   */
  async getActiveAlerts(filters = {}) {
    try {
      console.log("Obteniendo alertas activas...");
      const data = await apiClient.get(`${API_ALERTS_URL}/active`, filters);
      console.log("Alertas activas recibidas:", data);

      this.activeAlerts = data.alerts || data || [];
      this.updateActiveAlertsBadge();
      return this.activeAlerts;
    } catch (error) {
      console.error("Error al obtener alertas activas:", error);
      throw error;
    }
  },

  /**
   * Obtiene todas las alertas con filtros opcionales
   * @param {Object} filters - Filtros opcionales (serverID, severity, status, etc.)
   * @returns {Promise<Array>} Lista de alertas
   */
  async getAllAlerts(filters = {}) {
    try {
      console.log("Obteniendo todas las alertas...");
      const data = await apiClient.get(API_ALERTS_URL, filters);
      console.log("Alertas recibidas:", data);

      this.allAlerts = data.alerts || data || [];
      return this.allAlerts;
    } catch (error) {
      console.error("Error al obtener todas las alertas:", error);
      throw error;
    }
  },

  /**
   * Obtiene una alerta por su ID
   * @param {number} alertId - ID de la alerta
   * @returns {Promise<Object>} Datos de la alerta
   */
  async getAlertById(alertId) {
    try {
      return await apiClient.get(`${API_ALERTS_URL}/${alertId}`);
    } catch (error) {
      console.error(`Error al obtener alerta ID ${alertId}:`, error);
      throw error;
    }
  },

  /**
   * Reconoce una alerta
   * @param {number} alertId - ID de la alerta
   * @param {string} notes - Notas sobre el reconocimiento
   * @returns {Promise<Object>} Resultado de la operación
   */
  async acknowledgeAlert(alertId, notes = "") {
    try {
      console.log(`Reconociendo alerta ID: ${alertId}`);
      const result = await apiClient.post(`${API_ALERTS_URL}/${alertId}/acknowledge`, { notes });
      
      // Actualizar listas de alertas
      await this.getActiveAlerts();
      return result;
    } catch (error) {
      console.error(`Error al reconocer alerta ID ${alertId}:`, error);
      throw error;
    }
  },

  /**
   * Resuelve una alerta manualmente
   * @param {number} alertId - ID de la alerta
   * @param {string} notes - Notas sobre la resolución
   * @returns {Promise<Object>} Resultado de la operación
   */
  async resolveAlert(alertId, notes = "") {
    try {
      console.log(`Resolviendo alerta ID: ${alertId}`);
      const result = await apiClient.post(`${API_ALERTS_URL}/${alertId}/resolve`, { notes });
      
      // Actualizar listas de alertas
      await this.getActiveAlerts();
      return result;
    } catch (error) {
      console.error(`Error al resolver alerta ID ${alertId}:`, error);
      throw error;
    }
  },

  /**
   * Obtiene todos los umbrales de alerta
   * @returns {Promise<Array>} Lista de umbrales
   */
  async getThresholds() {
    try {
      console.log("Obteniendo umbrales de alerta...");
      const data = await apiClient.get(API_THRESHOLDS_URL);
      console.log("Umbrales recibidos:", data);

      this.thresholds = data.thresholds || data || [];
      return this.thresholds;
    } catch (error) {
      console.error("Error al obtener umbrales:", error);
      throw error;
    }
  },

  /**
   * Obtiene un umbral de alerta por su ID
   * @param {number} thresholdId - ID del umbral
   * @returns {Promise<Object>} Datos del umbral
   */
  async getThresholdById(thresholdId) {
    try {
      return await apiClient.get(`${API_THRESHOLDS_URL}/${thresholdId}`);
    } catch (error) {
      console.error(`Error al obtener umbral ID ${thresholdId}:`, error);
      throw error;
    }
  },

  /**
   * Crea un nuevo umbral de alerta
   * @param {Object} thresholdData - Datos del umbral
   * @returns {Promise<Object>} Umbral creado
   */
  async createThreshold(thresholdData) {
    try {
      console.log("Creando nuevo umbral de alerta:", thresholdData);
      const result = await apiClient.post(API_THRESHOLDS_URL, thresholdData);
      
      // Actualizar lista de umbrales
      await this.getThresholds();
      return result;
    } catch (error) {
      console.error("Error al crear umbral:", error);
      throw error;
    }
  },

  /**
   * Actualiza un umbral existente
   * @param {number} thresholdId - ID del umbral
   * @param {Object} thresholdData - Datos actualizados
   * @returns {Promise<Object>} Umbral actualizado
   */
  async updateThreshold(thresholdId, thresholdData) {
    try {
      console.log(`Actualizando umbral ID ${thresholdId}:`, thresholdData);
      const result = await apiClient.put(`${API_THRESHOLDS_URL}/${thresholdId}`, thresholdData);
      
      // Actualizar lista de umbrales
      await this.getThresholds();
      return result;
    } catch (error) {
      console.error(`Error al actualizar umbral ID ${thresholdId}:`, error);
      throw error;
    }
  },

  /**
   * Elimina un umbral
   * @param {number} thresholdId - ID del umbral
   * @returns {Promise<Object>} Resultado de la operación
   */
  async deleteThreshold(thresholdId) {
    try {
      console.log(`Eliminando umbral ID: ${thresholdId}`);
      const result = await apiClient.delete(`${API_THRESHOLDS_URL}/${thresholdId}`);
      
      // Actualizar lista de umbrales
      await this.getThresholds();
      return result;
    } catch (error) {
      console.error(`Error al eliminar umbral ID ${thresholdId}:`, error);
      throw error;
    }
  },

  /**
   * Obtiene los umbrales aplicables a un servidor
   * @param {number} serverId - ID del servidor
   * @returns {Promise<Array>} Lista de umbrales aplicables
   */
  async getThresholdsByServer(serverId) {
    try {
      return await apiClient.get(`${API_THRESHOLDS_URL}/server/${serverId}`);
    } catch (error) {
      console.error(`Error al obtener umbrales para servidor ID ${serverId}:`, error);
      throw error;
    }
  },

  /**
   * Actualiza el contador de alertas activas en la UI
   */
  updateActiveAlertsBadge() {
    const badge = document.getElementById("activeAlertsBadge");
    if (badge) {
      if (this.activeAlerts.length > 0) {
        badge.textContent = this.activeAlerts.length;
        badge.classList.remove("hidden");
      } else {
        badge.classList.add("hidden");
      }
    }
  },

  /**
   * Renderiza la lista de alertas en el DOM
   * @param {Array} alerts - Lista de alertas
   */
  renderAlerts(alerts) {
    const alertsList = document.getElementById("alertsList");
    alertsList.innerHTML = "";

    if (alerts.length === 0) {
      alertsList.innerHTML =
        '<p class="text-gray-500">No hay alertas que mostrar</p>';
      return;
    }

    console.log(`Renderizando ${alerts.length} alertas`);

    alerts.forEach((alert) => {
      const alertDiv = document.createElement("div");
      
      // Determinar clases de estilo según severidad
      let severityClass;
      switch (alert.severity) {
        case ALERT_SEVERITY.CRITICAL:
          severityClass = "border-red-500 bg-red-50";
          break;
        case ALERT_SEVERITY.WARNING:
          severityClass = "border-yellow-500 bg-yellow-50";
          break;
        default:
          severityClass = "border-blue-500 bg-blue-50";
      }
      
      // Determinar clases de estilo según estado
      let statusClass, statusText;
      switch (alert.status) {
        case ALERT_STATUS.ACTIVE:
          statusClass = "bg-red-100 text-red-800";
          statusText = "Activa";
          break;
        case ALERT_STATUS.ACKNOWLEDGED:
          statusClass = "bg-yellow-100 text-yellow-800";
          statusText = "Reconocida";
          break;
        case ALERT_STATUS.RESOLVED:
          statusClass = "bg-green-100 text-green-800";
          statusText = "Resuelta";
          break;
        case ALERT_STATUS.SUPPRESSED:
          statusClass = "bg-gray-100 text-gray-800";
          statusText = "Suprimida";
          break;
        default:
          statusClass = "bg-gray-100 text-gray-800";
          statusText = alert.status;
      }
      
      alertDiv.className = `border-l-4 ${severityClass} p-4 mb-4 rounded-lg shadow-sm`;
      
      // Formatear fechas
      const triggeredDate = new Date(alert.triggered_at).toLocaleString();
      const acknowledgedDate = alert.acknowledged_at 
        ? new Date(alert.acknowledged_at).toLocaleString() 
        : null;
      const resolvedDate = alert.resolved_at 
        ? new Date(alert.resolved_at).toLocaleString() 
        : null;
      
      // Construir HTML para la alerta
      alertDiv.innerHTML = `
        <div class="flex justify-between items-start">
          <div>
            <h3 class="text-lg font-semibold">${alert.title}</h3>
            <p class="text-sm text-gray-600">Servidor: ${alert.server ? alert.server.hostname : `ID: ${alert.server_id}`}</p>
          </div>
          <span class="px-2 py-1 text-xs ${statusClass} rounded-full">
            ${statusText}
          </span>
        </div>
        <div class="mt-2">
          <p>${alert.message}</p>
          <div class="mt-2 text-sm">
            <p><span class="font-medium">Métrica:</span> ${alert.metric_type}, Valor: ${alert.metric_value}, Umbral: ${alert.operator} ${alert.threshold}</p>
            <p><span class="font-medium">Activada:</span> ${triggeredDate}</p>
            ${acknowledgedDate ? `<p><span class="font-medium">Reconocida:</span> ${acknowledgedDate}</p>` : ''}
            ${resolvedDate ? `<p><span class="font-medium">Resuelta:</span> ${resolvedDate}</p>` : ''}
            ${alert.notes ? `<p class="mt-1 italic">"${alert.notes}"</p>` : ''}
          </div>
        </div>
        ${alert.status === ALERT_STATUS.ACTIVE ? `
          <div class="mt-3 flex space-x-2">
            <button onclick="alertsService.acknowledgeAlert(${alert.id})" 
                    class="text-xs bg-yellow-500 text-white px-2 py-1 rounded hover:bg-yellow-600">
              Reconocer
            </button>
            <button onclick="alertsService.resolveAlert(${alert.id})" 
                    class="text-xs bg-green-500 text-white px-2 py-1 rounded hover:bg-green-600">
              Resolver
            </button>
          </div>
        ` : ''}
        ${alert.status === ALERT_STATUS.ACKNOWLEDGED ? `
          <div class="mt-3 flex space-x-2">
            <button onclick="alertsService.resolveAlert(${alert.id})" 
                    class="text-xs bg-green-500 text-white px-2 py-1 rounded hover:bg-green-600">
              Resolver
            </button>
          </div>
        ` : ''}
      `;

      alertsList.appendChild(alertDiv);
    });
  },

  /**
   * Renderiza la lista de umbrales en el DOM
   */
  renderThresholds() {
    const thresholdsList = document.getElementById("thresholdsList");
    thresholdsList.innerHTML = "";

    if (this.thresholds.length === 0) {
      thresholdsList.innerHTML =
        '<p class="text-gray-500">No hay umbrales definidos</p>';
      return;
    }

    console.log(`Renderizando ${this.thresholds.length} umbrales`);

    this.thresholds.forEach((threshold) => {
      const thresholdDiv = document.createElement("div");
      thresholdDiv.className = "border rounded-lg p-4 bg-gray-50";

      // Determinar color por severidad
      let severityBadgeClass;
      switch (threshold.severity) {
        case ALERT_SEVERITY.CRITICAL:
          severityBadgeClass = "bg-red-100 text-red-800";
          break;
        case ALERT_SEVERITY.WARNING:
          severityBadgeClass = "bg-yellow-100 text-yellow-800";
          break;
        default:
          severityBadgeClass = "bg-blue-100 text-blue-800";
      }

      thresholdDiv.innerHTML = `
        <div class="flex justify-between items-center">
          <h3 class="text-lg font-semibold">${threshold.name}</h3>
          <span class="px-2 py-1 text-xs ${severityBadgeClass} rounded-full">
            ${threshold.severity}
          </span>
        </div>
        <p class="text-gray-600 text-sm">${threshold.description || "Sin descripción"}</p>
        <div class="mt-2 grid grid-cols-2 gap-2 text-sm">
          <p><span class="font-medium">Métrica:</span> ${threshold.metric_type}</p>
          <p><span class="font-medium">Condición:</span> ${threshold.operator} ${threshold.value}</p>
          ${threshold.server_id ? `<p><span class="font-medium">Servidor:</span> ${threshold.server_id}</p>` : '<p><span class="font-medium">Ámbito:</span> Global</p>'}
          <p><span class="font-medium">Cooldown:</span> ${threshold.cooldown_minutes} min</p>
        </div>
        <div class="mt-3 flex justify-end space-x-2">
          <button onclick="alertsService.showEditThresholdForm(${threshold.id})" 
                  class="text-xs bg-blue-500 text-white px-2 py-1 rounded hover:bg-blue-600">
            Editar
          </button>
          <button onclick="alertsService.confirmDeleteThreshold(${threshold.id})" 
                  class="text-xs bg-red-500 text-white px-2 py-1 rounded hover:bg-red-600">
            Eliminar
          </button>
        </div>
      `;

      thresholdsList.appendChild(thresholdDiv);
    });
  },

  /**
   * Muestra el formulario para editar un umbral
   * @param {number} thresholdId - ID del umbral a editar
   */
  async showEditThresholdForm(thresholdId) {
    console.log(`Preparando edición de umbral ID: ${thresholdId}`);
    // Aquí se implementaría la lógica para mostrar el formulario de edición
    // con los datos precargados del umbral especificado
    alert(`Funcionalidad para editar umbral ID: ${thresholdId} no implementada aún`);
  },

  /**
   * Solicita confirmación para eliminar un umbral
   * @param {number} thresholdId - ID del umbral a eliminar
   */
  confirmDeleteThreshold(thresholdId) {
    if (confirm(`¿Estás seguro de que deseas eliminar el umbral ID: ${thresholdId}?`)) {
      this.deleteThreshold(thresholdId);
    }
  },

  /**
   * Prepara el formulario de nuevo umbral con los servidores disponibles
   */
  prepareThresholdForm() {
    const serverSelect = document.getElementById("thresholdServer");
    if (serverSelect && serverService.serversData.length > 0) {
      // Limpiar opciones actuales excepto la global
      while (serverSelect.options.length > 1) {
        serverSelect.remove(1);
      }
      
      // Añadir opciones de servidores
      serverService.serversData.forEach(server => {
        const option = document.createElement("option");
        option.value = server.id;
        option.text = server.hostname;
        serverSelect.appendChild(option);
      });
    }
  }
}; 