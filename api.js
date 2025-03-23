// Configurar axios para incluir credenciales
axios.defaults.withCredentials = true;

/**
 * Cliente API para realizar peticiones al backend
 */
const apiClient = {
  /**
   * Realiza una petición GET a la API
   * @param {string} url - URL del endpoint
   * @param {Object} params - Parámetros de la petición
   * @returns {Promise} Promesa con la respuesta
   */
  async get(url, params = {}) {
    try {
      const response = await axios.get(url, { params });
      return response.data;
    } catch (error) {
      this.handleError(error);
      throw error;
    }
  },

  /**
   * Realiza una petición POST a la API
   * @param {string} url - URL del endpoint
   * @param {Object} data - Datos a enviar
   * @returns {Promise} Promesa con la respuesta
   */
  async post(url, data = {}) {
    try {
      const response = await axios.post(url, data);
      return response.data;
    } catch (error) {
      this.handleError(error);
      throw error;
    }
  },

  /**
   * Realiza una petición PUT a la API
   * @param {string} url - URL del endpoint
   * @param {Object} data - Datos a enviar
   * @returns {Promise} Promesa con la respuesta
   */
  async put(url, data = {}) {
    try {
      const response = await axios.put(url, data);
      return response.data;
    } catch (error) {
      this.handleError(error);
      throw error;
    }
  },

  /**
   * Realiza una petición DELETE a la API
   * @param {string} url - URL del endpoint
   * @param {Object} params - Parámetros de la petición
   * @returns {Promise} Promesa con la respuesta
   */
  async delete(url, params = {}) {
    try {
      const response = await axios.delete(url, { params });
      return response.data;
    } catch (error) {
      this.handleError(error);
      throw error;
    }
  },

  /**
   * Maneja errores de la API
   * @param {Error} error - Error producido
   */
  handleError(error) {
    console.error('Error API:', error);
    
    // Redirigir a la pantalla de login si hay error de autenticación
    if (error.response && error.response.status === 401) {
      console.log('Error de autenticación, redirigiendo a login');
      
      // Emitir evento para que la aplicación maneje la desconexión
      window.dispatchEvent(new CustomEvent('auth:unauthorized'));
    }
    
    return error;
  }
}; 