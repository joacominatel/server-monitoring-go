/**
 * Servicio de interfaz de usuario
 */
const uiService = {
  currentTab: 'servers',
  
  /**
   * Inicializa los eventos de UI
   */
  init() {
    // Maneja eventos de cambio entre login y registro
    document.getElementById("loginTabBtn")?.addEventListener("click", () => this.showLoginForm());
    document.getElementById("registerTabBtn")?.addEventListener("click", () => this.showRegisterForm());
    
    // Manejador de errores de autenticación
    window.addEventListener('auth:unauthorized', () => {
      this.showLoggedOutState();
    });
    
    // Inicializar UI basada en sesión
    this.setupLoginForm();
    this.setupRegisterForm();
  },

  /**
   * Configura el formulario de login
   */
  setupLoginForm() {
    const loginForm = document.getElementById("loginForm");
    const loginUsername = document.getElementById("loginUsername");
    const loginPassword = document.getElementById("loginPassword");
    
    if (loginForm) {
      // Añadir evento de envío con tecla Enter
      loginForm.addEventListener("keypress", (e) => {
        if (e.key === "Enter") {
          this.handleLogin();
        }
      });
      
      // Limpiar errores al escribir
      loginUsername?.addEventListener("input", () => {
        document.getElementById("loginError").classList.add("hidden");
      });
      
      loginPassword?.addEventListener("input", () => {
        document.getElementById("loginError").classList.add("hidden");
      });
    }
  },

  /**
   * Configura el formulario de registro
   */
  setupRegisterForm() {
    const registerForm = document.getElementById("registerForm");
    const registerUsername = document.getElementById("registerUsername");
    const registerPassword = document.getElementById("registerPassword");
    const registerConfirm = document.getElementById("registerConfirm");
    
    if (registerForm) {
      // Añadir evento de envío con tecla Enter
      registerForm.addEventListener("keypress", (e) => {
        if (e.key === "Enter") {
          this.handleRegister();
        }
      });
      
      // Limpiar errores al escribir
      const inputs = [registerUsername, registerPassword, registerConfirm];
      inputs.forEach(input => {
        input?.addEventListener("input", () => {
          document.getElementById("registerError").classList.add("hidden");
        });
      });
    }
  },

  /**
   * Muestra el formulario de login
   */
  showLoginForm() {
    document.getElementById("loginTabBtn").className =
      "px-4 py-2 bg-blue-500 text-white rounded-t-lg";
    document.getElementById("registerTabBtn").className =
      "px-4 py-2 bg-gray-300 text-gray-700 rounded-t-lg";
    document.getElementById("loginForm").classList.remove("hidden");
    document.getElementById("registerForm").classList.add("hidden");
  },

  /**
   * Muestra el formulario de registro
   */
  showRegisterForm() {
    document.getElementById("loginTabBtn").className =
      "px-4 py-2 bg-gray-300 text-gray-700 rounded-t-lg";
    document.getElementById("registerTabBtn").className =
      "px-4 py-2 bg-blue-500 text-white rounded-t-lg";
    document.getElementById("loginForm").classList.add("hidden");
    document.getElementById("registerForm").classList.remove("hidden");
  },

  /**
   * Muestra mensaje de error en el formulario de login
   * @param {string} message - Mensaje de error
   */
  showLoginError(message) {
    const errorElement = document.getElementById("loginError");
    errorElement.textContent = message;
    errorElement.classList.remove("hidden");
  },

  /**
   * Muestra mensaje de error en el formulario de registro
   * @param {string} message - Mensaje de error
   */
  showRegisterError(message) {
    const errorElement = document.getElementById("registerError");
    errorElement.textContent = message;
    errorElement.classList.remove("hidden");
  },

  /**
   * Muestra mensaje de éxito en el formulario de registro
   * @param {string} message - Mensaje de éxito
   */
  showRegisterSuccess(message) {
    const successElement = document.getElementById("registerSuccess");
    successElement.textContent = message;
    successElement.classList.remove("hidden");
    document.getElementById("registerError").classList.add("hidden");
  },

  /**
   * Muestra estado de usuario autenticado
   */
  showLoggedInState() {
    document.getElementById("authSection").classList.add("hidden");
    document.getElementById("mainContent").classList.remove("hidden");
    document.getElementById("logoutBtn").classList.remove("hidden");

    if (authService.currentUser) {
      document.getElementById(
        "userInfo"
      ).textContent = `Usuario: ${authService.currentUser.username} (${authService.currentUser.role})`;
    }

    console.log("Vista de usuario autenticado activada");
  },

  /**
   * Muestra estado de usuario no autenticado
   */
  showLoggedOutState() {
    document.getElementById("authSection").classList.remove("hidden");
    document.getElementById("mainContent").classList.add("hidden");
    document.getElementById("logoutBtn").classList.add("hidden");
    document.getElementById("userInfo").textContent = "No has iniciado sesión";

    // Ocultar sección de métricas y cerrar WebSocket
    document.getElementById("metricsSection").classList.add("hidden");
    metricsService.closeWebSocket();

    this.showLoginForm();
    console.log("Vista de usuario no autenticado activada");
  },

  /**
   * Maneja el envío del formulario de login
   */
  async handleLogin() {
    const username = document.getElementById("loginUsername").value.trim();
    const password = document.getElementById("loginPassword").value;

    if (!username || !password) {
      this.showLoginError("Por favor, introduce usuario y contraseña");
      return;
    }

    try {
      await authService.login(username, password);
      this.showLoggedInState();
      await serverService.getServers();
      serverService.renderServers();
    } catch (error) {
      this.showLoginError(
        error.response?.data?.message || "Error al iniciar sesión"
      );
    }
  },

  /**
   * Maneja el envío del formulario de registro
   */
  async handleRegister() {
    const username = document.getElementById("registerUsername").value.trim();
    const password = document.getElementById("registerPassword").value;
    const confirmPass = document.getElementById("registerConfirm").value;

    // Validar campos
    if (!username || !password) {
      this.showRegisterError("Por favor, completa todos los campos");
      return;
    }

    if (password !== confirmPass) {
      this.showRegisterError("Las contraseñas no coinciden");
      return;
    }

    try {
      await authService.register(username, password);
      
      this.showRegisterSuccess("Usuario registrado correctamente. Puedes iniciar sesión.");
      
      // Limpiar formulario
      document.getElementById("registerUsername").value = "";
      document.getElementById("registerPassword").value = "";
      document.getElementById("registerConfirm").value = "";

      // Mostrar login después de 2 segundos
      setTimeout(() => {
        this.showLoginForm();
        document.getElementById("registerSuccess").classList.add("hidden");
      }, 2000);
    } catch (error) {
      this.showRegisterError(
        error.response?.data?.message || "Error al registrar usuario"
      );
    }
  },

  /**
   * Maneja el cierre de sesión
   */
  async handleLogout() {
    try {
      await authService.logout();
      this.showLoggedOutState();
    } catch (error) {
      console.error("Error durante el logout:", error);
      this.showLoggedOutState();
    }
  },

  /**
   * Cambia entre pestañas de la aplicación
   * @param {string} tabName - Nombre de la pestaña
   */
  switchTab(tabName) {
    // Ocultar todos los paneles
    document.querySelectorAll('.tab-panel').forEach(panel => {
      panel.classList.add('hidden');
    });
    
    // Desactivar todas las pestañas
    document.querySelectorAll('[role="tab"]').forEach(tab => {
      tab.classList.remove('border-blue-600');
      tab.classList.add('border-transparent');
      tab.classList.add('hover:border-gray-300');
      tab.setAttribute('aria-selected', 'false');
    });
    
    // Activar pestaña seleccionada
    const selectedTab = document.getElementById(`${tabName}-tab`);
    if (selectedTab) {
      selectedTab.classList.add('border-blue-600');
      selectedTab.classList.remove('border-transparent', 'hover:border-gray-300');
      selectedTab.setAttribute('aria-selected', 'true');
    }
    
    // Mostrar panel seleccionado
    const selectedPanel = document.getElementById(`${tabName}Panel`);
    if (selectedPanel) {
      selectedPanel.classList.remove('hidden');
    }
    
    this.currentTab = tabName;
    
    // Cargar datos específicos de la pestaña
    this.loadTabData(tabName);
  },

  /**
   * Carga los datos específicos de una pestaña
   * @param {string} tabName - Nombre de la pestaña
   */
  async loadTabData(tabName) {
    switch (tabName) {
      case 'servers':
        await serverService.getServers();
        serverService.renderServers();
        break;
      case 'alerts':
        await alertsService.getActiveAlerts();
        alertsService.renderAlerts(alertsService.activeAlerts);
        break;
      case 'thresholds':
        await alertsService.getThresholds();
        alertsService.renderThresholds();
        alertsService.prepareThresholdForm();
        break;
    }
  },

  /**
   * Muestra u oculta el formulario de umbral
   * @param {boolean} show - Indicador para mostrar u ocultar
   */
  toggleThresholdForm(show) {
    const thresholdForm = document.getElementById("thresholdForm");
    if (thresholdForm) {
      if (show) {
        thresholdForm.classList.remove("hidden");
        // Preparar opciones del formulario
        alertsService.prepareThresholdForm();
      } else {
        thresholdForm.classList.add("hidden");
      }
    }
  },

  /**
   * Aplica los filtros en la sección de alertas
   */
  applyAlertFilters() {
    const serverFilter = document.getElementById("alertServerFilter")?.value;
    const severityFilter = document.getElementById("alertSeverityFilter")?.value;
    const statusFilter = document.getElementById("alertStatusFilter")?.value;
    
    const filters = {};
    if (serverFilter) filters.server_id = serverFilter;
    if (severityFilter) filters.severity = severityFilter;
    if (statusFilter) filters.status = statusFilter;
    
    console.log("Aplicando filtros:", filters);
    
    alertsService.getAllAlerts(filters).then(alerts => {
      alertsService.renderAlerts(alerts);
    });
  }
}; 