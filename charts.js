/**
 * Servicio para gestión de gráficos
 */
const chartsService = {
  // Gráficos
  cpuChart: null,
  memChart: null,
  diskChart: null,
  networkChart: null,
  
  // Datos para gráficos
  cpuData: {
    labels: [],
    datasets: [
      {
        label: "CPU (%)",
        data: [],
        borderColor: "rgb(59, 130, 246)",
        backgroundColor: "rgba(59, 130, 246, 0.2)",
        tension: 0.2,
        fill: true,
      },
    ],
  },
  
  memData: {
    labels: [],
    datasets: [
      {
        label: "Memoria (%)",
        data: [],
        borderColor: "rgb(16, 185, 129)",
        backgroundColor: "rgba(16, 185, 129, 0.2)",
        tension: 0.2,
        fill: true,
      },
    ],
  },
  
  diskData: {
    labels: [],
    datasets: [
      {
        label: "Disco (%)",
        data: [],
        borderColor: "rgb(139, 92, 246)",
        backgroundColor: "rgba(139, 92, 246, 0.2)",
        tension: 0.2,
        fill: true,
      },
    ],
  },
  
  networkData: {
    labels: [],
    datasets: [
      {
        label: "Descarga (KB/s)",
        data: [],
        borderColor: "rgb(245, 158, 11)",
        backgroundColor: "rgba(245, 158, 11, 0.1)",
        tension: 0.2,
        yAxisID: 'y',
      },
      {
        label: "Subida (KB/s)",
        data: [],
        borderColor: "rgb(239, 68, 68)",
        backgroundColor: "rgba(239, 68, 68, 0.1)",
        tension: 0.2,
        yAxisID: 'y',
      },
    ],
  },

  /**
   * Inicializa los gráficos
   */
  initCharts() {
    // Definir opciones comunes para los gráficos
    const commonOptions = {
      responsive: true,
      animation: false,
      plugins: {
        legend: {
          display: true,
          position: 'top',
        },
        tooltip: {
          mode: 'index',
          intersect: false,
        },
      },
      scales: {
        x: {
          grid: {
            display: false,
          },
        },
        y: {
          beginAtZero: true,
          grid: {
            color: 'rgba(0, 0, 0, 0.05)',
          },
        },
      },
    };
    
    // Inicializar gráfico de CPU
    const cpuCtx = document.getElementById("cpuChart")?.getContext("2d");
    if (cpuCtx) {
      this.cpuChart = new Chart(cpuCtx, {
        type: "line",
        data: this.cpuData,
        options: {
          ...commonOptions,
          scales: {
            ...commonOptions.scales,
            y: {
              ...commonOptions.scales.y,
              max: 100,
              title: {
                display: true,
                text: 'Porcentaje'
              }
            },
          },
        },
      });
    }

    // Inicializar gráfico de memoria
    const memCtx = document.getElementById("memChart")?.getContext("2d");
    if (memCtx) {
      this.memChart = new Chart(memCtx, {
        type: "line",
        data: this.memData,
        options: {
          ...commonOptions,
          scales: {
            ...commonOptions.scales,
            y: {
              ...commonOptions.scales.y,
              max: 100,
              title: {
                display: true,
                text: 'Porcentaje'
              }
            },
          },
        },
      });
    }

    // Inicializar gráfico de disco
    const diskCtx = document.getElementById("diskChart")?.getContext("2d");
    if (diskCtx) {
      this.diskChart = new Chart(diskCtx, {
        type: "line",
        data: this.diskData,
        options: {
          ...commonOptions,
          scales: {
            ...commonOptions.scales,
            y: {
              ...commonOptions.scales.y,
              max: 100,
              title: {
                display: true,
                text: 'Porcentaje'
              }
            },
          },
        },
      });
    }

    // Inicializar gráfico de red
    const networkCtx = document.getElementById("networkChart")?.getContext("2d");
    if (networkCtx) {
      this.networkChart = new Chart(networkCtx, {
        type: "line",
        data: this.networkData,
        options: {
          ...commonOptions,
          scales: {
            ...commonOptions.scales,
            y: {
              ...commonOptions.scales.y,
              title: {
                display: true,
                text: 'KB/s'
              }
            },
          },
        },
      });
    }
  },

  /**
   * Reinicia los datos de los gráficos
   */
  resetCharts() {
    // Reiniciar datos de gráficos
    this.cpuData.labels = [];
    this.cpuData.datasets[0].data = [];
    
    this.memData.labels = [];
    this.memData.datasets[0].data = [];
    
    this.diskData.labels = [];
    this.diskData.datasets[0].data = [];
    
    this.networkData.labels = [];
    this.networkData.datasets[0].data = [];
    this.networkData.datasets[1].data = [];

    // Actualizar gráficos
    this.updateAllCharts();
  },

  /**
   * Actualiza todos los gráficos
   */
  updateAllCharts() {
    if (this.cpuChart) this.cpuChart.update();
    if (this.memChart) this.memChart.update();
    if (this.diskChart) this.diskChart.update();
    if (this.networkChart) this.networkChart.update();
  },

  /**
   * Actualiza los datos de los gráficos
   * @param {string} timestamp - Marca de tiempo
   * @param {number} cpuValue - Valor de CPU
   * @param {number} memValue - Valor de memoria
   * @param {number} diskValue - Valor de disco
   * @param {number} netDown - Tráfico de descarga
   * @param {number} netUp - Tráfico de subida
   */
  updateCharts(timestamp, cpuValue, memValue, diskValue, netDown, netUp) {
    // Limitar número de puntos en los gráficos
    if (this.cpuData.labels.length > MAX_CHART_POINTS) {
      this.cpuData.labels.shift();
      this.cpuData.datasets[0].data.shift();
      
      this.memData.labels.shift();
      this.memData.datasets[0].data.shift();
      
      this.diskData.labels.shift();
      this.diskData.datasets[0].data.shift();
      
      this.networkData.labels.shift();
      this.networkData.datasets[0].data.shift();
      this.networkData.datasets[1].data.shift();
    }

    // Añadir nuevos datos
    this.cpuData.labels.push(timestamp);
    this.cpuData.datasets[0].data.push(parseFloat(cpuValue));

    this.memData.labels.push(timestamp);
    this.memData.datasets[0].data.push(parseFloat(memValue));
    
    this.diskData.labels.push(timestamp);
    this.diskData.datasets[0].data.push(parseFloat(diskValue));
    
    // Convertir bytes a KB para gráfico de red
    const netDownKB = netDown ? Math.round(netDown / 1024) : 0;
    const netUpKB = netUp ? Math.round(netUp / 1024) : 0;
    
    this.networkData.labels.push(timestamp);
    this.networkData.datasets[0].data.push(netDownKB);
    this.networkData.datasets[1].data.push(netUpKB);

    // Actualizar gráficos
    this.updateAllCharts();
  }
}; 