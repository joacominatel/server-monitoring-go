/**
 * Servicio de autenticación y gestión de usuarios
 */
const authService = {
  currentUser: null,

  /**
   * Verifica si hay una sesión activa
   * @returns {Promise<Object|null>} Usuario actual o null
   */
  async checkSession() {
    try {
      console.log("Verificando sesión...");
      const data = await apiClient.get(`${API_AUTH_URL}/me`);
      
      if (data && data.user) {
        this.currentUser = data.user;
        console.log("Sesión activa encontrada para:", this.currentUser.username);
        return this.currentUser;
      }
      return null;
    } catch (error) {
      console.log("No hay sesión activa:", error.message);
      return null;
    }
  },

  /**
   * Inicia sesión en el sistema
   * @param {string} username - Nombre de usuario
   * @param {string} password - Contraseña
   * @returns {Promise<Object>} Datos del usuario
   */
  async login(username, password) {
    try {
      console.log(`Intentando login para usuario: ${username}`);
      const data = await apiClient.post(`${API_AUTH_URL}/login`, {
        username,
        password
      });

      if (data && data.user) {
        this.currentUser = data.user;
      } else {
        this.currentUser = await this.getUserInfo();
      }
      
      console.log("Login exitoso para:", username);
      return this.currentUser;
    } catch (error) {
      console.error("Error durante el login:", error);
      throw error;
    }
  },

  /**
   * Obtiene información del usuario actual
   * @returns {Promise<Object|null>} Datos del usuario
   */
  async getUserInfo() {
    try {
      console.log("Obteniendo información del usuario...");
      const data = await apiClient.get(`${API_AUTH_URL}/me`);
      console.log("Información del usuario:", data);
      return data.user;
    } catch (error) {
      console.error("Error al obtener información del usuario:", error);
      return null;
    }
  },

  /**
   * Registra un nuevo usuario
   * @param {string} username - Nombre de usuario
   * @param {string} password - Contraseña
   * @returns {Promise<Object>} Resultado del registro
   */
  async register(username, password) {
    try {
      console.log(`Intentando registrar usuario: ${username}`);
      const data = await apiClient.post(`${API_AUTH_URL}/register`, {
        username,
        password
      });
      
      console.log("Respuesta de registro:", data);
      return data;
    } catch (error) {
      console.error("Error durante el registro:", error);
      throw error;
    }
  },

  /**
   * Cierra la sesión actual
   * @returns {Promise<void>}
   */
  async logout() {
    try {
      console.log("Cerrando sesión...");
      await apiClient.post(`${API_AUTH_URL}/logout`);
      console.log("Sesión cerrada correctamente");
      this.currentUser = null;
    } catch (error) {
      console.error("Error durante el logout:", error);
      this.currentUser = null;
      throw error;
    }
  },

  /**
   * Comprueba si el usuario tiene un rol específico
   * @param {string} role - Rol a comprobar
   * @returns {boolean} True si tiene el rol
   */
  hasRole(role) {
    return this.currentUser && this.currentUser.role === role;
  },

  /**
   * Comprueba si el usuario es administrador
   * @returns {boolean} True si es administrador
   */
  isAdmin() {
    return this.hasRole(USER_ROLES.ADMIN);
  },

  /**
   * Cambia la contraseña del usuario
   * @param {string} currentPassword - Contraseña actual
   * @param {string} newPassword - Nueva contraseña
   * @returns {Promise<Object>} Resultado de la operación
   */
  async changePassword(currentPassword, newPassword) {
    try {
      return await apiClient.post(`${API_AUTH_URL}/change-password`, {
        current_password: currentPassword,
        new_password: newPassword
      });
    } catch (error) {
      console.error("Error al cambiar contraseña:", error);
      throw error;
    }
  }
}; 