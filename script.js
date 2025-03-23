/**
 * Archivo principal que coordina los diferentes módulos de la aplicación
 */

// Variables globales para facilitar el acceso desde la consola en desarrollo
window.authService = authService;
window.serverService = serverService;
window.metricsService = metricsService;
window.chartsService = chartsService;
window.alertsService = alertsService;
window.uiService = uiService;

// Función principal que se ejecuta cuando el DOM está listo
document.addEventListener("DOMContentLoaded", async () => {
  console.log("Inicializando aplicación de monitoreo de servidores...");
  
  // Inicializar servicios de UI
  uiService.init();
  
  // Inicializar gráficos
  chartsService.initCharts();
  
  // Comprobar si hay sesión activa y cargar datos iniciales
  try {
    const user = await authService.checkSession();
    if (user) {
      uiService.showLoggedInState();
      
      // Cargar datos iniciales
      await serverService.getServers();
      serverService.renderServers();
      
      // Verificar si hay alertas activas
      await alertsService.getActiveAlerts();
    }
  } catch (error) {
    console.error("Error al inicializar la aplicación:", error);
  }

  // Configurar manejadores de eventos globales
  setupEventHandlers();
  
  console.log("Aplicación inicializada correctamente");
});

/**
 * Configura los manejadores de eventos para elementos DOM
 */
function setupEventHandlers() {
  // Manejadores para autenticación
  document.getElementById("logoutBtn")?.addEventListener("click", () => uiService.handleLogout());
  
  // Botones de formularios
  const loginBtn = document.querySelector('button[onclick="login()"]');
  if (loginBtn) {
    loginBtn.onclick = () => uiService.handleLogin();
  }
  
  const registerBtn = document.querySelector('button[onclick="register()"]');
  if (registerBtn) {
    registerBtn.onclick = () => uiService.handleRegister();
  }
  
  // Tabs de navegación
  const tabButtons = document.querySelectorAll('[role="tab"]');
  tabButtons.forEach(button => {
    const tabName = button.id.replace('-tab', '');
    button.onclick = () => uiService.switchTab(tabName);
  });
  
  // Botones de alertas
  const getActiveAlertsBtn = document.querySelector('button[onclick="getActiveAlerts()"]');
  if (getActiveAlertsBtn) {
    getActiveAlertsBtn.onclick = async () => {
      const alerts = await alertsService.getActiveAlerts();
      alertsService.renderAlerts(alerts);
    };
  }
  
  const getAllAlertsBtn = document.querySelector('button[onclick="getAllAlerts()"]');
  if (getAllAlertsBtn) {
    getAllAlertsBtn.onclick = async () => {
      const alerts = await alertsService.getAllAlerts();
      alertsService.renderAlerts(alerts);
    };
  }
  
  // Botón para aplicar filtros
  const applyFiltersBtn = document.querySelector('button[onclick="applyAlertFilters()"]');
  if (applyFiltersBtn) {
    applyFiltersBtn.onclick = () => uiService.applyAlertFilters();
  }
  
  // Formulario de umbrales
  const showThresholdFormBtn = document.querySelector('button[onclick="showThresholdForm()"]');
  if (showThresholdFormBtn) {
    showThresholdFormBtn.onclick = () => uiService.toggleThresholdForm(true);
  }
  
  const hideThresholdFormBtn = document.querySelector('button[onclick="hideThresholdForm()"]');
  if (hideThresholdFormBtn) {
    hideThresholdFormBtn.onclick = () => uiService.toggleThresholdForm(false);
  }
  
  const createThresholdBtn = document.querySelector('button[onclick="createThreshold()"]');
  if (createThresholdBtn) {
    createThresholdBtn.onclick = async () => {
      // Recoger datos del formulario
      const thresholdData = {
        name: document.getElementById("thresholdName").value,
        description: document.getElementById("thresholdDescription").value,
        metric_type: document.getElementById("thresholdMetricType").value,
        operator: document.getElementById("thresholdOperator").value,
        value: parseFloat(document.getElementById("thresholdValue").value),
        server_id: document.getElementById("thresholdServer").value || null,
        severity: document.getElementById("thresholdSeverity").value,
        cooldown_minutes: parseInt(document.getElementById("thresholdCooldown").value),
        enable_discord: document.getElementById("thresholdEnableDiscord").checked
      };
      
      try {
        await alertsService.createThreshold(thresholdData);
        alert("Umbral creado correctamente");
        uiService.toggleThresholdForm(false);
        
        // Limpiar formulario
        document.getElementById("thresholdName").value = "";
        document.getElementById("thresholdDescription").value = "";
        document.getElementById("thresholdValue").value = "";
        
        // Actualizar lista de umbrales
        await alertsService.getThresholds();
        alertsService.renderThresholds();
      } catch (error) {
        alert(error.response?.data?.message || "Error al crear umbral");
      }
    };
  }
  
  // Botón para cerrar WebSocket
  const closeSocketBtn = document.getElementById("closeSocketBtn");
  if (closeSocketBtn) {
    closeSocketBtn.onclick = () => {
      metricsService.closeWebSocket();
      document.getElementById("metricsSection").classList.add("hidden");
    };
  }
}
