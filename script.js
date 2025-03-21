// Configuración de la API
const apiBaseUrl = "http://localhost:8080/api";
const apiAuthUrl = `${apiBaseUrl}/auth`;
const apiServerUrl = `${apiBaseUrl}/servers`;
const apiMetricsUrl = `${apiBaseUrl}/metrics`;

// Configure axios defaults
axios.defaults.withCredentials = true;
// No necesitamos esta línea ya que el backend lo manejará correctamente
// axios.defaults.headers.common['Access-Control-Allow-Credentials'] = true;

// Variables globales
let currentUser = null;
let serversData = [];
let activeWebSocket = null;
let activeServerId = null;

// Variables para gráficos
let cpuChart = null;
let memChart = null;
const cpuData = {
  labels: [],
  datasets: [
    {
      label: "CPU Usage (%)",
      data: [],
      borderColor: "rgb(59, 130, 246)",
      tension: 0.1,
    },
  ],
};
const memData = {
  labels: [],
  datasets: [
    {
      label: "Memory Usage (%)",
      data: [],
      borderColor: "rgb(16, 185, 129)",
      tension: 0.1,
    },
  ],
};

// Elementos DOM
document.addEventListener("DOMContentLoaded", () => {
  // Cambiar entre formularios de login y registro
  document
    .getElementById("loginTabBtn")
    .addEventListener("click", () => showLoginForm());
  document
    .getElementById("registerTabBtn")
    .addEventListener("click", () => showRegisterForm());

  // Inicializar gráficos
  initCharts();

  // Comprobar si ya hay sesión activa
  checkSession();

  console.log("Aplicación inicializada");
});

// Inicializar gráficos ChartJS
function initCharts() {
  const cpuCtx = document.getElementById("cpuChart").getContext("2d");
  const memCtx = document.getElementById("memChart").getContext("2d");

  cpuChart = new Chart(cpuCtx, {
    type: "line",
    data: cpuData,
    options: {
      responsive: true,
      animation: false,
      scales: {
        y: {
          beginAtZero: true,
          max: 100,
        },
      },
    },
  });

  memChart = new Chart(memCtx, {
    type: "line",
    data: memData,
    options: {
      responsive: true,
      animation: false,
      scales: {
        y: {
          beginAtZero: true,
          max: 100,
        },
      },
    },
  });
}

// Funciones de gestión de interfaz
function showLoginForm() {
  document.getElementById("loginTabBtn").className =
    "px-4 py-2 bg-blue-500 text-white rounded-t-lg";
  document.getElementById("registerTabBtn").className =
    "px-4 py-2 bg-gray-300 text-gray-700 rounded-t-lg";
  document.getElementById("loginForm").classList.remove("hidden");
  document.getElementById("registerForm").classList.add("hidden");
}

function showRegisterForm() {
  document.getElementById("loginTabBtn").className =
    "px-4 py-2 bg-gray-300 text-gray-700 rounded-t-lg";
  document.getElementById("registerTabBtn").className =
    "px-4 py-2 bg-blue-500 text-white rounded-t-lg";
  document.getElementById("loginForm").classList.add("hidden");
  document.getElementById("registerForm").classList.remove("hidden");
}

function showLoggedInState() {
  document.getElementById("authSection").classList.add("hidden");
  document.getElementById("mainContent").classList.remove("hidden");
  document.getElementById("logoutBtn").classList.remove("hidden");

  if (currentUser) {
    document.getElementById(
      "userInfo"
    ).textContent = `Usuario: ${currentUser.username} (${currentUser.role})`;
  }

  console.log("Vista de usuario autenticado activada");
}

function showLoggedOutState() {
  document.getElementById("authSection").classList.remove("hidden");
  document.getElementById("mainContent").classList.add("hidden");
  document.getElementById("logoutBtn").classList.add("hidden");
  document.getElementById("userInfo").textContent = "No has iniciado sesión";

  // Ocultar sección de métricas y cerrar WebSocket
  document.getElementById("metricsSection").classList.add("hidden");
  closeWebSocket();

  showLoginForm();
  console.log("Vista de usuario no autenticado activada");
}

// Funciones de autenticación
async function checkSession() {
  try {
    console.log("Verificando sesión...");
    const response = await axios.get(`${apiAuthUrl}/me`);
    console.log("Respuesta de sesión:", response.data);

    if (response.data && response.data.user) {
      currentUser = response.data.user;
      showLoggedInState();
      getServers(); // Cargar servidores automáticamente
      console.log("Sesión activa encontrada para:", currentUser.username);
    }
  } catch (error) {
    console.log("No hay sesión activa:", error.message);
  }
}

async function login() {
  const username = document.getElementById("loginUsername").value.trim();
  const password = document.getElementById("loginPassword").value;

  if (!username || !password) {
    showLoginError("Por favor, introduce usuario y contraseña");
    return;
  }

  try {
    console.log(`Intentando login para usuario: ${username}`);
    const response = await axios.post(`${apiAuthUrl}/login`, {
      username,
      password,
    });

    console.log("Respuesta de login:", response.data);

    if (response.data) {
      currentUser = response.data.user || (await getUserInfo());
      showLoggedInState();
      getServers();
      console.log("Login exitoso para:", username);
    }
  } catch (error) {
    console.error("Error durante el login:", error);
    showLoginError(error.response?.data?.message || "Error al iniciar sesión");
  }
}

async function getUserInfo() {
  try {
    console.log("Obteniendo información del usuario...");
    const response = await axios.get(`${apiAuthUrl}/me`);
    console.log("Información del usuario:", response.data);
    return response.data.user;
  } catch (error) {
    console.error("Error al obtener información del usuario:", error);
    return null;
  }
}

async function register() {
  const username = document.getElementById("registerUsername").value.trim();
  const password = document.getElementById("registerPassword").value;
  const confirmPass = document.getElementById("registerConfirm").value;

  // Validar campos
  if (!username || !password) {
    showRegisterError("Por favor, completa todos los campos");
    return;
  }

  if (password !== confirmPass) {
    showRegisterError("Las contraseñas no coinciden");
    return;
  }

  try {
    console.log(`Intentando registrar usuario: ${username}`);
    const response = await axios.post(`${apiAuthUrl}/register`, {
      username,
      password,
    });

    console.log("Respuesta de registro:", response.data);

    document.getElementById("registerSuccess").textContent =
      "Usuario registrado correctamente. Puedes iniciar sesión.";
    document.getElementById("registerSuccess").classList.remove("hidden");
    document.getElementById("registerError").classList.add("hidden");

    // Limpiar formulario
    document.getElementById("registerUsername").value = "";
    document.getElementById("registerPassword").value = "";
    document.getElementById("registerConfirm").value = "";

    // Mostrar login después de 2 segundos
    setTimeout(() => {
      showLoginForm();
      document.getElementById("registerSuccess").classList.add("hidden");
    }, 2000);
  } catch (error) {
    console.error("Error durante el registro:", error);
    showRegisterError(
      error.response?.data?.message || "Error al registrar usuario"
    );
  }
}

async function logout() {
  try {
    console.log("Cerrando sesión...");
    await axios.post(`${apiAuthUrl}/logout`);
    console.log("Sesión cerrada correctamente");
  } catch (error) {
    console.error("Error durante el logout:", error);
  } finally {
    currentUser = null;
    showLoggedOutState();
  }
}

// Funciones de manejo de errores
function showLoginError(message) {
  const errorElement = document.getElementById("loginError");
  errorElement.textContent = message;
  errorElement.classList.remove("hidden");
}

function showRegisterError(message) {
  const errorElement = document.getElementById("registerError");
  errorElement.textContent = message;
  errorElement.classList.remove("hidden");
}

// Funciones de servidores
async function getServers() {
  try {
    console.log("Obteniendo lista de servidores...");
    const response = await axios.get(apiServerUrl);
    console.log("Servidores recibidos:", response.data);

    serversData = response.data.servers || response.data || [];
    renderServers();
  } catch (error) {
    console.error("Error al obtener servidores:", error);

    // Si es error de autenticación, volver al estado de no autenticado
    if (error.response && error.response.status === 401) {
      console.log(
        "Error de autenticación al obtener servidores, cerrando sesión"
      );
      showLoggedOutState();
    }
  }
}

function renderServers() {
  const serverList = document.getElementById("serverList");
  serverList.innerHTML = "";

  if (serversData.length === 0) {
    serverList.innerHTML =
      '<p class="text-gray-500">No hay servidores disponibles</p>';
    return;
  }

  console.log(`Renderizando ${serversData.length} servidores`);

  serversData.forEach((server) => {
    const serverCard = document.createElement("div");
    serverCard.className = "border rounded-lg p-4 bg-gray-50";

    serverCard.innerHTML = `
            <div class="flex justify-between items-center">
                <h3 class="text-lg font-semibold">${
                  server.hostname || "Sin nombre"
                }</h3>
                <span class="px-2 py-1 text-xs ${
                  server.status === "online"
                    ? "bg-green-100 text-green-800"
                    : "bg-red-100 text-red-800"
                } rounded-full">
                    ${server.status || "desconocido"}
                </span>
            </div>
            <p class="text-gray-600 text-sm">${
              server.description || "Sin descripción"
            }</p>
            <div class="text-sm mt-2">
                <p><span class="font-medium">IP:</span> ${
                  server.ip || "N/A"
                }</p>
                <p><span class="font-medium">ID:</span> ${server.id}</p>
            </div>
            <div class="mt-3 flex space-x-2">
                <button onclick="getServerMetrics(${
                  server.id
                })" class="text-xs bg-blue-500 text-white px-2 py-1 rounded hover:bg-blue-600">
                    Ver métricas
                </button>
                <button onclick="startRealTimeMonitoring(${
                  server.id
                })" class="text-xs bg-green-500 text-white px-2 py-1 rounded hover:bg-green-600">
                    Monitoreo en tiempo real
                </button>
            </div>
        `;

    serverList.appendChild(serverCard);
  });
}

// Función para obtener métricas de un servidor específico
async function getServerMetrics(serverId) {
  try {
    console.log(`Obteniendo métricas para servidor ID: ${serverId}`);
    const response = await axios.get(
      `${apiMetricsUrl}/server/${serverId}/latest`
    );
    console.log("Métricas del servidor:", response.data);

    // Aquí se podría mostrar las métricas en una ventana modal o en otra sección
    alert(
      `Métricas más recientes del servidor ${serverId}\n\n${JSON.stringify(
        response.data,
        null,
        2
      )}`
    );
  } catch (error) {
    console.error("Error al obtener métricas:", error);
  }
}

// WebSockets para métricas en tiempo real
function startRealTimeMonitoring(serverId) {
  // Cerrar WebSocket previo si existe
  closeWebSocket();
  
  console.log(`Iniciando monitoreo en tiempo real para servidor ID: ${serverId}`);
  
  // Encontrar y mostrar información del servidor
  const server = serversData.find(s => s.id === serverId);
  if (server) {
    document.getElementById('serverInfo').innerHTML = `
        <p><strong>Servidor:</strong> ${server.hostname || 'Sin nombre'}</p>
        <p><strong>IP:</strong> ${server.ip || 'N/A'}</p>
        <p><strong>Descripción:</strong> ${server.description || 'Sin descripción'}</p>
    `;
  }
  
  // Mostrar sección de métricas
  document.getElementById('metricsSection').classList.remove('hidden');
  activeServerId = serverId;
  
  // Inicializar los datos de las gráficas
  resetCharts();
  
  // Construir la URL de WebSocket
  let wsProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
  let authToken = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxLCJ1c2VybmFtZSI6ImFkbWluIiwiZW1haWwiOiJhZG1pbkBzaXN0ZW1hLmxvY2FsIiwicm9sZSI6IkFETUlOIiwic3ViIjoiMSIsImV4cCI6MjA1MzYxMjQ2NiwiaWF0IjoxNzQyNTcyNDY2fQ.2EmQP5abtVbVTGyLMS5dCJeqvsSxk5YWgJtoG5fwkyk'
  let wsUrl = `${wsProtocol}//localhost:8080/api/metrics/live/${serverId}`;

  function getCookieValue(name) 
    {
      const regex = new RegExp(`(^| )${name}=([^;]+)`)
      const match = document.cookie.match(regex)
      if (match) {
        return match[2]
      }
   }
  
  // Extraer token de la cookie auth_token para añadirlo como parámetro de consulta
  // Según el archivo auth.go, el backend espera un parámetro ?token= en la URL
  // const authToken = getCookieValue('auth_token');
  console.log('Token de autenticación:', authToken);
    
  console.log(`Conectando WebSocket a: ${wsUrl}`);
  
  // Crear conexión WebSocket
  try {
    activeWebSocket = new WebSocket(wsUrl + `?token=${authToken}`);
    
    activeWebSocket.onopen = (event) => {
      console.log('Conexión WebSocket establecida:', event);
      document.getElementById('socketStatus').textContent = 'Conectado';
      document.getElementById('socketStatus').className = 'px-2 py-1 bg-green-100 text-green-800 rounded-full text-xs';
    };

    activeWebSocket.onmessage = (event) => {
      try {
        const metrics = JSON.parse(event.data);
        console.log("Nuevas métricas recibidas:", metrics);
        updateMetricsDisplay(metrics);
      } catch (error) {
        console.error("Error al procesar métricas:", error);
      }
    };

    activeWebSocket.onerror = (error) => {
      console.error("Error de WebSocket:", error);
      document.getElementById("socketStatus").textContent = "Error";
      document.getElementById("socketStatus").className =
        "px-2 py-1 bg-red-100 text-red-800 rounded-full text-xs";
    };

    activeWebSocket.onclose = (event) => {
      console.log("Conexión WebSocket cerrada:", event);
      document.getElementById("socketStatus").textContent = "Desconectado";
      document.getElementById("socketStatus").className =
        "px-2 py-1 bg-gray-200 text-gray-700 rounded-full text-xs";

      // Si no fue cerrado manualmente, intentar reconectar
      if (activeServerId && event.code !== 1000) {
        console.log("Intentando reconectar en 3 segundos...");
        setTimeout(() => {
          if (activeServerId) {
            startRealTimeMonitoring(activeServerId);
          }
        }, 3000);
      }
    };
  } catch (error) {
    console.error("Error al crear conexión WebSocket:", error);
  }
}

function closeWebSocket() {
  if (activeWebSocket) {
    console.log("Cerrando conexión WebSocket");
    activeWebSocket.close(1000, "Cerrado por el usuario");
    activeWebSocket = null;
    activeServerId = null;
  }
}

function resetCharts() {
  // Reiniciar datos de gráficos
  cpuData.labels = [];
  cpuData.datasets[0].data = [];
  memData.labels = [];
  memData.datasets[0].data = [];

  if (cpuChart && memChart) {
    cpuChart.update();
    memChart.update();
  }

  // Reiniciar valores de métricas
  document.getElementById("cpuValue").textContent = "-%";
  document.getElementById("memValue").textContent = "-%";
  document.getElementById("diskValue").textContent = "-%";
  document.getElementById("networkValue").textContent = "-";
}

function updateMetricsDisplay(metrics) {
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

  // Actualizar valor de red (si está disponible)
  if (metrics.network_in !== undefined && metrics.network_out !== undefined) {
    const netIn = formatBytes(metrics.network_in);
    const netOut = formatBytes(metrics.network_out);
    document.getElementById(
      "networkValue"
    ).textContent = `↓${netIn}/s ↑${netOut}/s`;
  }

  // Actualizar gráficos
  updateCharts(timestamp, metrics.cpu_usage, memUsedPercent);
}

function updateCharts(timestamp, cpuValue, memValue) {
  // Mantener solo los últimos 20 puntos
  if (cpuData.labels.length > 20) {
    cpuData.labels.shift();
    cpuData.datasets[0].data.shift();
    memData.labels.shift();
    memData.datasets[0].data.shift();
  }

  // Añadir nuevos datos
  cpuData.labels.push(timestamp);
  cpuData.datasets[0].data.push(cpuValue);

  memData.labels.push(timestamp);
  memData.datasets[0].data.push(memValue);

  // Actualizar gráficos
  cpuChart.update();
  memChart.update();
}

// Función auxiliar para formatear bytes
function formatBytes(bytes, decimals = 1) {
  if (bytes === 0) return "0 B";

  const k = 1024;
  const sizes = ["B", "KB", "MB", "GB", "TB"];
  const i = Math.floor(Math.log(bytes) / Math.log(k));

  return (
    parseFloat((bytes / Math.pow(k, i)).toFixed(decimals)) + " " + sizes[i]
  );
}
