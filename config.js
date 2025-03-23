// Configuración de la API
const API_BASE_URL = "http://localhost:8080/api";
const API_AUTH_URL = `${API_BASE_URL}/auth`;
const API_SERVER_URL = `${API_BASE_URL}/servers`;
const API_METRICS_URL = `${API_BASE_URL}/metrics`;
const API_ALERTS_URL = `${API_BASE_URL}/alerts`;
const API_THRESHOLDS_URL = `${API_BASE_URL}/alert-thresholds`;
const API_LOGS_URL = `${API_BASE_URL}/logs`;

// Configuración WebSocket
const WS_PING_INTERVAL = 30000; // 30 segundos
const MAX_CHART_POINTS = 20; // Número máximo de puntos en gráficas

// Severidad de alertas
const ALERT_SEVERITY = {
  INFO: "info",
  WARNING: "warning",
  CRITICAL: "critical"
};

// Estados de alertas
const ALERT_STATUS = {
  ACTIVE: "active",
  ACKNOWLEDGED: "acknowledged",
  RESOLVED: "resolved",
  SUPPRESSED: "suppressed"
};

// Tipos de métricas
const METRIC_TYPE = {
  CPU: "cpu",
  MEMORY: "memory",
  DISK: "disk",
  NETWORK_IN: "network_in",
  NETWORK_OUT: "network_out"
};

// Operadores para umbrales
const THRESHOLD_OPERATORS = {
  GT: ">",
  LT: "<",
  GTE: ">=",
  LTE: "<=",
  EQ: "=="
};

// Roles de usuario
const USER_ROLES = {
  ADMIN: "ADMIN",
  USER: "USER",
  VIEWER: "VIEWER"
}; 