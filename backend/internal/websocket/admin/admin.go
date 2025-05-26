package admin

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/davidM20/micro-service-backend-go.git/internal/websocket/wsmodels"
	"github.com/davidM20/micro-service-backend-go.git/pkg/customws"
	"github.com/davidM20/micro-service-backend-go.git/pkg/logger"
)

// AdminAuth estructura para autenticación admin
type AdminAuth struct {
	Username string
	Password string
}

// MetricsCollector recolecta métricas del sistema
type MetricsCollector struct {
	// Contadores atómicos
	TotalConnections      int64
	TotalMessages         int64
	TotalErrors           int64
	MessagesPerSecond     int64
	ConnectionsPerMinute  int64
	LastSecondMessages    int64
	LastMinuteConnections int64

	// Mapas protegidos por mutex
	mutex                sync.RWMutex
	ErrorsByType         map[string]int64
	MessagesByType       map[string]int64
	UserSessions         map[int64]time.Time
	DatabaseQueryTimes   []time.Duration
	LastNDatabaseQueries int // mantener últimas N consultas para promedio

	// Referencias
	manager *customws.ConnectionManager[wsmodels.WsUserData]
	db      *sql.DB

	// Timers para cálculos periódicos
	lastSecondTime time.Time
	lastMinuteTime time.Time
}

// AdminHandler maneja todas las rutas administrativas
type AdminHandler struct {
	auth      AdminAuth
	collector *MetricsCollector
}

var (
	// Instancia global del collector
	globalCollector *MetricsCollector
	once            sync.Once
)

// InitializeAdmin inicializa el sistema de administración
func InitializeAdmin(manager *customws.ConnectionManager[wsmodels.WsUserData], db *sql.DB, adminUser, adminPass string) *AdminHandler {
	once.Do(func() {
		globalCollector = &MetricsCollector{
			ErrorsByType:         make(map[string]int64),
			MessagesByType:       make(map[string]int64),
			UserSessions:         make(map[int64]time.Time),
			DatabaseQueryTimes:   make([]time.Duration, 0, 100), // Buffer para 100 consultas
			LastNDatabaseQueries: 100,
			manager:              manager,
			db:                   db,
			lastSecondTime:       time.Now(),
			lastMinuteTime:       time.Now(),
		}

		// Iniciar goroutine para calcular métricas periódicas
		go globalCollector.startMetricsCalculation()

		logger.Info("ADMIN", "Sistema de administración inicializado")
	})

	return &AdminHandler{
		auth: AdminAuth{
			Username: adminUser,
			Password: adminPass,
		},
		collector: globalCollector,
	}
}

// GetCollector retorna la instancia global del collector
func GetCollector() *MetricsCollector {
	return globalCollector
}

// Middleware de autenticación para rutas admin
func (ah *AdminHandler) RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		if !ok || username != ah.auth.Username || password != ah.auth.Password {
			w.Header().Set("WWW-Authenticate", `Basic realm="Admin Area"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next(w, r)
	}
}

// RegisterAdminRoutes registra todas las rutas administrativas
func (ah *AdminHandler) RegisterAdminRoutes(mux *http.ServeMux) {
	// Dashboard principal
	mux.HandleFunc("/admin", ah.RequireAuth(ah.HandleDashboard))

	// API endpoints
	mux.HandleFunc("/admin/api/metrics", ah.RequireAuth(ah.HandleMetricsAPI))
	mux.HandleFunc("/admin/api/connections", ah.RequireAuth(ah.HandleConnectionsAPI))
	mux.HandleFunc("/admin/api/users", ah.RequireAuth(ah.HandleUsersAPI))
	mux.HandleFunc("/admin/api/errors", ah.RequireAuth(ah.HandleErrorsAPI))
	mux.HandleFunc("/admin/api/system", ah.RequireAuth(ah.HandleSystemAPI))

	logger.Info("ADMIN", "Rutas administrativas registradas")
}

// HandleDashboard sirve el dashboard HTML principal
func (ah *AdminHandler) HandleDashboard(w http.ResponseWriter, r *http.Request) {
	html := ah.generateDashboardHTML()
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

// HandleMetricsAPI devuelve métricas generales
func (ah *AdminHandler) HandleMetricsAPI(w http.ResponseWriter, r *http.Request) {
	ah.collector.mutex.RLock()
	defer ah.collector.mutex.RUnlock()

	metrics := map[string]interface{}{
		"activeConnections":    ah.getActiveConnectionsCount(),
		"totalConnections":     atomic.LoadInt64(&ah.collector.TotalConnections),
		"totalMessages":        atomic.LoadInt64(&ah.collector.TotalMessages),
		"totalErrors":          atomic.LoadInt64(&ah.collector.TotalErrors),
		"messagesPerSecond":    atomic.LoadInt64(&ah.collector.MessagesPerSecond),
		"connectionsPerMinute": atomic.LoadInt64(&ah.collector.ConnectionsPerMinute),
		"errorsByType":         ah.collector.ErrorsByType,
		"messagesByType":       ah.collector.MessagesByType,
		"averageQueryTime":     ah.collector.getAverageQueryTime(),
		"timestamp":            time.Now().Unix(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

// HandleConnectionsAPI devuelve información de conexiones activas
func (ah *AdminHandler) HandleConnectionsAPI(w http.ResponseWriter, r *http.Request) {
	ah.collector.mutex.RLock()
	sessions := make(map[string]interface{})
	for userID, connectTime := range ah.collector.UserSessions {
		sessions[fmt.Sprintf("%d", userID)] = map[string]interface{}{
			"userId":      userID,
			"connectedAt": connectTime.Unix(),
			"duration":    time.Since(connectTime).Seconds(),
		}
	}
	ah.collector.mutex.RUnlock()

	response := map[string]interface{}{
		"activeConnections": ah.getActiveConnectionsCount(),
		"sessions":          sessions,
		"timestamp":         time.Now().Unix(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleUsersAPI devuelve estadísticas de usuarios
func (ah *AdminHandler) HandleUsersAPI(w http.ResponseWriter, r *http.Request) {
	// Consultar estadísticas de usuarios desde la BD
	onlineUsers := ah.getOnlineUsersCount()
	totalUsers := ah.getTotalUsersCount()
	recentUsers := ah.getRecentUsersCount(24) // últimas 24 horas

	response := map[string]interface{}{
		"onlineUsers":    onlineUsers,
		"totalUsers":     totalUsers,
		"recentUsers24h": recentUsers,
		"timestamp":      time.Now().Unix(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleErrorsAPI devuelve estadísticas detalladas de errores
func (ah *AdminHandler) HandleErrorsAPI(w http.ResponseWriter, r *http.Request) {
	ah.collector.mutex.RLock()
	defer ah.collector.mutex.RUnlock()

	response := map[string]interface{}{
		"totalErrors":  atomic.LoadInt64(&ah.collector.TotalErrors),
		"errorsByType": ah.collector.ErrorsByType,
		"timestamp":    time.Now().Unix(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleSystemAPI devuelve métricas del sistema
func (ah *AdminHandler) HandleSystemAPI(w http.ResponseWriter, r *http.Request) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	response := map[string]interface{}{
		"memory": map[string]interface{}{
			"allocMB":      bToMb(m.Alloc),
			"totalAllocMB": bToMb(m.TotalAlloc),
			"sysMB":        bToMb(m.Sys),
			"numGC":        m.NumGC,
		},
		"goroutines":     runtime.NumGoroutine(),
		"averageQueryMs": ah.collector.getAverageQueryTime().Milliseconds(),
		"timestamp":      time.Now().Unix(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Métodos del MetricsCollector

// RecordMessage registra un mensaje procesado
func (mc *MetricsCollector) RecordMessage(messageType string) {
	atomic.AddInt64(&mc.TotalMessages, 1)
	atomic.AddInt64(&mc.LastSecondMessages, 1)

	mc.mutex.Lock()
	mc.MessagesByType[messageType]++
	mc.mutex.Unlock()
}

// RecordError registra un error
func (mc *MetricsCollector) RecordError(errorType string) {
	atomic.AddInt64(&mc.TotalErrors, 1)

	mc.mutex.Lock()
	mc.ErrorsByType[errorType]++
	mc.mutex.Unlock()
}

// RecordConnection registra una nueva conexión
func (mc *MetricsCollector) RecordConnection(userID int64) {
	atomic.AddInt64(&mc.TotalConnections, 1)
	atomic.AddInt64(&mc.LastMinuteConnections, 1)

	mc.mutex.Lock()
	mc.UserSessions[userID] = time.Now()
	mc.mutex.Unlock()
}

// RecordDisconnection registra una desconexión
func (mc *MetricsCollector) RecordDisconnection(userID int64) {
	mc.mutex.Lock()
	delete(mc.UserSessions, userID)
	mc.mutex.Unlock()
}

// RecordDatabaseQuery registra el tiempo de una consulta a BD
func (mc *MetricsCollector) RecordDatabaseQuery(duration time.Duration) {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()

	if len(mc.DatabaseQueryTimes) >= mc.LastNDatabaseQueries {
		// Eliminar el más antiguo
		mc.DatabaseQueryTimes = mc.DatabaseQueryTimes[1:]
	}
	mc.DatabaseQueryTimes = append(mc.DatabaseQueryTimes, duration)
}

// startMetricsCalculation inicia el cálculo periódico de métricas
func (mc *MetricsCollector) startMetricsCalculation() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		now := time.Now()

		// Calcular mensajes por segundo
		if now.Sub(mc.lastSecondTime) >= time.Second {
			atomic.StoreInt64(&mc.MessagesPerSecond, atomic.SwapInt64(&mc.LastSecondMessages, 0))
			mc.lastSecondTime = now
		}

		// Calcular conexiones por minuto
		if now.Sub(mc.lastMinuteTime) >= time.Minute {
			atomic.StoreInt64(&mc.ConnectionsPerMinute, atomic.SwapInt64(&mc.LastMinuteConnections, 0))
			mc.lastMinuteTime = now
		}
	}
}

// getAverageQueryTime calcula el tiempo promedio de consultas a BD
func (mc *MetricsCollector) getAverageQueryTime() time.Duration {
	if len(mc.DatabaseQueryTimes) == 0 {
		return 0
	}

	var total time.Duration
	for _, duration := range mc.DatabaseQueryTimes {
		total += duration
	}
	return total / time.Duration(len(mc.DatabaseQueryTimes))
}

// Funciones auxiliares para consultas a BD

func (ah *AdminHandler) getOnlineUsersCount() int {
	var count int
	err := ah.collector.db.QueryRow("SELECT COUNT(*) FROM UserStatus WHERE IsOnline = 1").Scan(&count)
	if err != nil {
		logger.Errorf("ADMIN", "Error consultando usuarios online: %v", err)
		return 0
	}
	return count
}

func (ah *AdminHandler) getTotalUsersCount() int {
	var count int
	err := ah.collector.db.QueryRow("SELECT COUNT(*) FROM User").Scan(&count)
	if err != nil {
		logger.Errorf("ADMIN", "Error consultando total de usuarios: %v", err)
		return 0
	}
	return count
}

func (ah *AdminHandler) getRecentUsersCount(hours int) int {
	var count int
	query := "SELECT COUNT(*) FROM UserStatus WHERE LastSeen >= DATE_SUB(NOW(), INTERVAL ? HOUR)"
	err := ah.collector.db.QueryRow(query, hours).Scan(&count)
	if err != nil {
		logger.Errorf("ADMIN", "Error consultando usuarios recientes: %v", err)
		return 0
	}
	return count
}

// Función auxiliar para convertir bytes a MB
func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}

// generateDashboardHTML genera el HTML del dashboard administrativo
func (ah *AdminHandler) generateDashboardHTML() string {
	return `<!DOCTYPE html>
<html lang="es">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>WebSocket Server - Panel de Administración</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        
        body {
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            min-height: 100vh;
            color: #333;
        }
        
        .container {
            max-width: 1400px;
            margin: 0 auto;
            padding: 20px;
        }
        
        .header {
            background: rgba(255, 255, 255, 0.95);
            padding: 20px;
            border-radius: 10px;
            margin-bottom: 30px;
            box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
        }
        
        .header h1 {
            color: #2c3e50;
            margin-bottom: 10px;
        }
        
        .header p {
            color: #7f8c8d;
            font-size: 14px;
        }
        
        .metrics-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
            gap: 20px;
            margin-bottom: 30px;
        }
        
        .metric-card {
            background: rgba(255, 255, 255, 0.95);
            border-radius: 10px;
            padding: 20px;
            box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
            transition: transform 0.2s;
        }
        
        .metric-card:hover {
            transform: translateY(-2px);
        }
        
        .metric-card h3 {
            color: #2c3e50;
            margin-bottom: 15px;
            font-size: 18px;
            border-bottom: 2px solid #3498db;
            padding-bottom: 5px;
        }
        
        .metric-value {
            font-size: 32px;
            font-weight: bold;
            color: #3498db;
            margin: 10px 0;
        }
        
        .metric-label {
            color: #7f8c8d;
            font-size: 14px;
        }
        
        .metric-list {
            list-style: none;
        }
        
        .metric-list li {
            padding: 5px 0;
            border-bottom: 1px solid #ecf0f1;
            display: flex;
            justify-content: space-between;
        }
        
        .metric-list li:last-child {
            border-bottom: none;
        }
        
        .status-indicator {
            display: inline-block;
            width: 10px;
            height: 10px;
            border-radius: 50%;
            margin-right: 10px;
        }
        
        .status-online {
            background: #27ae60;
        }
        
        .status-warning {
            background: #f39c12;
        }
        
        .status-error {
            background: #e74c3c;
        }
        
        .refresh-btn {
            background: #3498db;
            color: white;
            border: none;
            padding: 10px 20px;
            border-radius: 5px;
            cursor: pointer;
            font-size: 14px;
            margin: 10px 5px;
            transition: background 0.2s;
        }
        
        .refresh-btn:hover {
            background: #2980b9;
        }
        
        .auto-refresh {
            color: #27ae60;
            font-size: 12px;
            margin-left: 10px;
        }
        
        .chart-container {
            background: rgba(255, 255, 255, 0.95);
            border-radius: 10px;
            padding: 20px;
            margin-bottom: 20px;
            box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
        }
        
        .sessions-table {
            width: 100%;
            border-collapse: collapse;
            margin-top: 15px;
        }
        
        .sessions-table th,
        .sessions-table td {
            padding: 10px;
            text-align: left;
            border-bottom: 1px solid #ecf0f1;
        }
        
        .sessions-table th {
            background: #f8f9fa;
            font-weight: 600;
            color: #2c3e50;
        }
        
        .sessions-table tr:hover {
            background: #f8f9fa;
        }
        
        .error-badge {
            background: #e74c3c;
            color: white;
            padding: 2px 8px;
            border-radius: 12px;
            font-size: 12px;
        }
        
        .success-badge {
            background: #27ae60;
            color: white;
            padding: 2px 8px;
            border-radius: 12px;
            font-size: 12px;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🔧 Panel de Administración - Servidor WebSocket</h1>
            <p>Monitoreo en tiempo real del servidor de chat</p>
            <button class="refresh-btn" onclick="refreshAll()">🔄 Actualizar Todo</button>
            <button class="refresh-btn" onclick="toggleAutoRefresh()" id="autoRefreshBtn">⏱️ Auto-actualizar: OFF</button>
            <span class="auto-refresh" id="autoRefreshStatus"></span>
        </div>

        <div class="metrics-grid">
            <!-- Métricas Generales -->
            <div class="metric-card">
                <h3>📊 Métricas Generales</h3>
                <div class="metric-value" id="activeConnections">-</div>
                <div class="metric-label">Conexiones Activas</div>
                <div style="margin-top: 15px;">
                    <div><strong>Total Conexiones:</strong> <span id="totalConnections">-</span></div>
                    <div><strong>Total Mensajes:</strong> <span id="totalMessages">-</span></div>
                    <div><strong>Mensajes/seg:</strong> <span id="messagesPerSecond">-</span></div>
                </div>
            </div>

            <!-- Estado del Sistema -->
            <div class="metric-card">
                <h3>💻 Estado del Sistema</h3>
                <div class="metric-value" id="goroutines">-</div>
                <div class="metric-label">Goroutines Activas</div>
                <div style="margin-top: 15px;">
                    <div><strong>Memoria (MB):</strong> <span id="memoryAlloc">-</span></div>
                    <div><strong>GC Runs:</strong> <span id="numGC">-</span></div>
                    <div><strong>Query Avg (ms):</strong> <span id="avgQueryTime">-</span></div>
                </div>
            </div>

            <!-- Usuarios -->
            <div class="metric-card">
                <h3>👥 Estadísticas de Usuarios</h3>
                <div class="metric-value" id="onlineUsers">-</div>
                <div class="metric-label">Usuarios Online</div>
                <div style="margin-top: 15px;">
                    <div><strong>Total Usuarios:</strong> <span id="totalUsers">-</span></div>
                    <div><strong>Activos 24h:</strong> <span id="recentUsers">-</span></div>
                </div>
            </div>

            <!-- Errores -->
            <div class="metric-card">
                <h3>🚨 Errores</h3>
                <div class="metric-value" id="totalErrors">-</div>
                <div class="metric-label">Total de Errores</div>
                <ul class="metric-list" id="errorsByType">
                    <li>Cargando...</li>
                </ul>
            </div>
        </div>

        <!-- Tipos de Mensajes -->
        <div class="chart-container">
            <h3>📨 Tipos de Mensajes</h3>
            <ul class="metric-list" id="messagesByType">
                <li>Cargando...</li>
            </ul>
        </div>

        <!-- Sesiones Activas -->
        <div class="chart-container">
            <h3>🔗 Sesiones Activas</h3>
            <table class="sessions-table">
                <thead>
                    <tr>
                        <th>Usuario ID</th>
                        <th>Tiempo Conectado</th>
                        <th>Conectado Desde</th>
                        <th>Estado</th>
                    </tr>
                </thead>
                <tbody id="sessionsTable">
                    <tr><td colspan="4">Cargando...</td></tr>
                </tbody>
            </table>
        </div>
    </div>

    <script>
        let autoRefresh = false;
        let autoRefreshInterval;

        function formatBytes(bytes) {
            return bytes + ' MB';
        }

        function formatDuration(seconds) {
            const hours = Math.floor(seconds / 3600);
            const minutes = Math.floor((seconds % 3600) / 60);
            const secs = Math.floor(seconds % 60);
            
            if (hours > 0) {
                return hours + 'h ' + minutes + 'm';
            } else if (minutes > 0) {
                return minutes + 'm ' + secs + 's';
            } else {
                return secs + 's';
            }
        }

        function formatTimestamp(timestamp) {
            return new Date(timestamp * 1000).toLocaleString();
        }

        async function fetchMetrics() {
            try {
                const response = await fetch('/admin/api/metrics');
                const data = await response.json();
                
                document.getElementById('activeConnections').textContent = data.activeConnections;
                document.getElementById('totalConnections').textContent = data.totalConnections;
                document.getElementById('totalMessages').textContent = data.totalMessages;
                document.getElementById('totalErrors').textContent = data.totalErrors;
                document.getElementById('messagesPerSecond').textContent = data.messagesPerSecond;

                // Mensajes por tipo
                const messagesList = document.getElementById('messagesByType');
                messagesList.innerHTML = '';
                for (const [type, count] of Object.entries(data.messagesByType || {})) {
                    const li = document.createElement('li');
                    li.innerHTML = '<span>' + type + '</span><span>' + count + '</span>';
                    messagesList.appendChild(li);
                }
                
                if (Object.keys(data.messagesByType || {}).length === 0) {
                    messagesList.innerHTML = '<li>No hay datos disponibles</li>';
                }

            } catch (error) {
                console.error('Error fetching metrics:', error);
            }
        }

        async function fetchSystemInfo() {
            try {
                const response = await fetch('/admin/api/system');
                const data = await response.json();
                
                document.getElementById('goroutines').textContent = data.goroutines;
                document.getElementById('memoryAlloc').textContent = formatBytes(data.memory.allocMB);
                document.getElementById('numGC').textContent = data.memory.numGC;
                document.getElementById('avgQueryTime').textContent = data.averageQueryMs + ' ms';

            } catch (error) {
                console.error('Error fetching system info:', error);
            }
        }

        async function fetchUsers() {
            try {
                const response = await fetch('/admin/api/users');
                const data = await response.json();
                
                document.getElementById('onlineUsers').textContent = data.onlineUsers;
                document.getElementById('totalUsers').textContent = data.totalUsers;
                document.getElementById('recentUsers').textContent = data.recentUsers24h;

            } catch (error) {
                console.error('Error fetching users:', error);
            }
        }

        async function fetchErrors() {
            try {
                const response = await fetch('/admin/api/errors');
                const data = await response.json();
                
                document.getElementById('totalErrors').textContent = data.totalErrors;

                // Errores por tipo
                const errorsList = document.getElementById('errorsByType');
                errorsList.innerHTML = '';
                for (const [type, count] of Object.entries(data.errorsByType || {})) {
                    const li = document.createElement('li');
                    li.innerHTML = '<span>' + type + '</span><span class="error-badge">' + count + '</span>';
                    errorsList.appendChild(li);
                }
                
                if (Object.keys(data.errorsByType || {}).length === 0) {
                    errorsList.innerHTML = '<li>No hay errores registrados</li>';
                }

            } catch (error) {
                console.error('Error fetching errors:', error);
            }
        }

        async function fetchConnections() {
            try {
                const response = await fetch('/admin/api/connections');
                const data = await response.json();
                
                const table = document.getElementById('sessionsTable');
                table.innerHTML = '';
                
                for (const [userId, session] of Object.entries(data.sessions || {})) {
                    const row = table.insertRow();
                    row.innerHTML = 
                        '<td>' + session.userId + '</td>' +
                        '<td>' + formatDuration(session.duration) + '</td>' +
                        '<td>' + formatTimestamp(session.connectedAt) + '</td>' +
                        '<td><span class="status-indicator status-online"></span>Online</td>';
                }
                
                if (Object.keys(data.sessions || {}).length === 0) {
                    const row = table.insertRow();
                    row.innerHTML = '<td colspan="4">No hay sesiones activas</td>';
                }

            } catch (error) {
                console.error('Error fetching connections:', error);
            }
        }

        function refreshAll() {
            fetchMetrics();
            fetchSystemInfo();
            fetchUsers();
            fetchErrors();
            fetchConnections();
        }

        function toggleAutoRefresh() {
            autoRefresh = !autoRefresh;
            const btn = document.getElementById('autoRefreshBtn');
            const status = document.getElementById('autoRefreshStatus');
            
            if (autoRefresh) {
                btn.textContent = '⏱️ Auto-actualizar: ON';
                btn.style.background = '#27ae60';
                status.textContent = '(actualiza cada 5 segundos)';
                autoRefreshInterval = setInterval(refreshAll, 5000);
            } else {
                btn.textContent = '⏱️ Auto-actualizar: OFF';
                btn.style.background = '#3498db';
                status.textContent = '';
                clearInterval(autoRefreshInterval);
            }
        }

        // Cargar datos iniciales
        refreshAll();
    </script>
</body>
</html>`
}

// getActiveConnectionsCount cuenta manualmente las conexiones activas
func (ah *AdminHandler) getActiveConnectionsCount() int64 {
	ah.collector.mutex.RLock()
	defer ah.collector.mutex.RUnlock()
	return int64(len(ah.collector.UserSessions))
}
