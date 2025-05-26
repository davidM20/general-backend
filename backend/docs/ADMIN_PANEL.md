# Panel de Administración - Servidor WebSocket

## Descripción

El panel de administración proporciona una interfaz web completa para monitorear en tiempo real el estado del servidor WebSocket, conexiones activas, métricas de rendimiento y errores del sistema.

## Características Principales

### 🔐 Autenticación Segura
- Autenticación HTTP Basic
- Credenciales configurables por variables de entorno
- Acceso protegido a todas las rutas administrativas

### 📊 Métricas en Tiempo Real
- **Conexiones Activas**: Número de usuarios conectados
- **Total de Conexiones**: Contador histórico desde el inicio
- **Mensajes por Segundo**: Throughput actual del servidor
- **Mensajes por Tipo**: Desglose de tipos de mensajes procesados
- **Errores por Tipo**: Clasificación de errores del sistema

### 💻 Monitoreo del Sistema
- **Goroutines Activas**: Número de goroutines ejecutándose
- **Uso de Memoria**: Memoria asignada, total y del sistema
- **Garbage Collector**: Número de ejecuciones del GC
- **Latencia de BD**: Tiempo promedio de consultas a la base de datos

### 👥 Estadísticas de Usuarios
- **Usuarios Online**: Cantidad actual de usuarios conectados
- **Total de Usuarios**: Usuarios registrados en el sistema
- **Usuarios Activos 24h**: Usuarios que han estado activos en las últimas 24 horas

### 🔗 Gestión de Sesiones
- **Tabla de Sesiones Activas**: Lista detallada de usuarios conectados
- **Tiempo de Conexión**: Duración de cada sesión
- **Timestamp de Conexión**: Momento de inicio de cada sesión

## Configuración

### Variables de Entorno

Agregar a tu archivo `.env`:

```bash
# Sistema de Administración
ADMIN_USERNAME=admin
ADMIN_PASSWORD=tu_password_seguro_aqui
```

### Valores por Defecto

Si no se configuran las variables de entorno:
- **Username**: `admin`
- **Password**: `admin123`

⚠️ **Importante**: Cambiar las credenciales por defecto en producción.

## Integración en el Código

### 1. Inicialización en main.go

```go
import "backend/internal/websocket/admin"

// Inicializar sistema de administración
adminUser := os.Getenv("ADMIN_USERNAME")
adminPass := os.Getenv("ADMIN_PASSWORD")
if adminUser == "" {
    adminUser = "admin"
}
if adminPass == "" {
    adminPass = "admin123"
}

adminHandler := admin.InitializeAdmin(manager, dbConn, adminUser, adminPass)

// Configurar rutas
mux := http.NewServeMux()
mux.HandleFunc("/ws", manager.ServeHTTP)
mux.HandleFunc("/health", healthHandler)

// Registrar rutas administrativas
adminHandler.RegisterAdminRoutes(mux)
```

### 2. Registro de Métricas en Router

```go
// En ProcessClientMessage
collector := admin.GetCollector()
if collector != nil {
    collector.RecordMessage(string(msg.Type))
}

// Registrar errores
if err != nil && collector != nil {
    collector.RecordError(string(msg.Type) + "_error")
}
```

### 3. Registro de Conexiones en Callbacks

```go
// En OnConnect
collector := admin.GetCollector()
if collector != nil {
    collector.RecordConnection(conn.ID)
}

// En OnDisconnect
if collector != nil {
    collector.RecordDisconnection(conn.ID)
}
```

### 4. Métricas de Base de Datos

```go
import "backend/internal/db/queries"

// Configurar recorder de métricas
collector := admin.GetCollector()
queries.SetMetricsRecorder(collector)

// Usar funciones con métricas
user, err := queries.GetUserBySessionTokenWithMetrics(db, token)
```

## Acceso al Panel

### URL del Dashboard
```
http://localhost:8080/admin
```

### APIs REST Disponibles

| Endpoint | Descripción |
|----------|-------------|
| `GET /admin` | Dashboard HTML principal |
| `GET /admin/api/metrics` | Métricas generales del servidor |
| `GET /admin/api/connections` | Información de conexiones activas |
| `GET /admin/api/users` | Estadísticas de usuarios |
| `GET /admin/api/errors` | Detalles de errores del sistema |
| `GET /admin/api/system` | Métricas del sistema (memoria, goroutines) |

### Ejemplos de Respuesta

#### /admin/api/metrics
```json
{
  "activeConnections": 15,
  "totalConnections": 342,
  "totalMessages": 5847,
  "totalErrors": 12,
  "messagesPerSecond": 23,
  "connectionsPerMinute": 5,
  "errorsByType": {
    "GET_CHAT_LIST_error": 3,
    "SEND_CHAT_MESSAGE_error": 9
  },
  "messagesByType": {
    "GET_CHAT_LIST": 1250,
    "SEND_CHAT_MESSAGE": 3890,
    "GET_MY_PROFILE": 707
  },
  "averageQueryTime": "15ms",
  "timestamp": 1703123456
}
```

#### /admin/api/system
```json
{
  "memory": {
    "allocMB": 45,
    "totalAllocMB": 892,
    "sysMB": 67,
    "numGC": 23
  },
  "goroutines": 48,
  "averageQueryMs": 15,
  "timestamp": 1703123456
}
```

## Funcionalidades del Dashboard

### 🔄 Auto-actualización
- Botón toggle para activar/desactivar
- Actualización automática cada 5 segundos
- Indicador visual del estado

### 📈 Visualización de Datos
- Cards con métricas principales
- Códigos de color para estado del sistema
- Tablas detalladas para sesiones activas
- Badges para errores y estado

### 🎨 Interfaz Moderna
- Diseño responsive
- Gradientes y efectos visuales
- Hover effects
- Iconos descriptivos

## Consideraciones de Seguridad

### Protección de Acceso
- ✅ Autenticación HTTP Basic requerida
- ✅ Credenciales configurables
- ✅ Sin exposición de datos sensibles

### Recomendaciones
- 🔐 Usar contraseñas seguras en producción
- 🛡️ Configurar HTTPS en producción
- 🚫 No exponer el panel en redes públicas
- 📝 Revisar logs regularmente

## Troubleshooting

### Panel no accesible
```bash
# Verificar que el servidor esté ejecutándose
curl http://localhost:8080/health

# Verificar credenciales
curl -u admin:admin123 http://localhost:8080/admin/api/metrics
```

### Métricas no se actualizan
```bash
# Verificar logs del servidor
# Las métricas se registran con cada mensaje/conexión/error
```

### Error 401 Unauthorized
- Verificar variables de entorno `ADMIN_USERNAME` y `ADMIN_PASSWORD`
- Confirmar que las credenciales en el cliente coinciden

### Error de conexión a BD en métricas de usuarios
- Verificar conectividad a MySQL
- Confirmar que las tablas `User` y `UserStatus` existen

## Extensión y Personalización

### Agregar Nuevas Métricas

```go
// En MetricsCollector
type MetricsCollector struct {
    // ... campos existentes ...
    CustomMetric int64 `json:"customMetric"`
}

// Agregar método de registro
func (mc *MetricsCollector) RecordCustomMetric(value int64) {
    atomic.AddInt64(&mc.CustomMetric, value)
}
```

### Nuevos Endpoints de API

```go
// En AdminHandler
func (ah *AdminHandler) HandleCustomAPI(w http.ResponseWriter, r *http.Request) {
    // Lógica personalizada
    response := map[string]interface{}{
        "customData": "valor",
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}

// Registrar en RegisterAdminRoutes
mux.HandleFunc("/admin/api/custom", ah.RequireAuth(ah.HandleCustomAPI))
```

## Conclusión

El panel de administración proporciona una herramienta completa para monitorear y diagnosticar el servidor WebSocket en tiempo real, facilitando la operación y mantenimiento del sistema de chat.

Para cualquier problema o mejora, revisar los logs del servidor y las métricas disponibles en el dashboard. 