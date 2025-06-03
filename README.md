# general-backend# Paquete `customws` - Gestor de Conexiones WebSocket Genérico en Go

## 1. Visión General

El paquete `customws` proporciona una base robusta, escalable y **completamente genérica** para construir servidores WebSocket en Go. Diseñado como una capa de infraestructura limpia, permite a los desarrolladores construir cualquier tipo de aplicación en tiempo real sin estar limitados a funcionalidades específicas.

El paquete maneja toda la complejidad del ciclo de vida de las conexiones WebSocket, concurrencia, envío y recepción de mensajes, control de pings/pongs, y un sistema avanzado de confirmaciones (acks) y solicitud/respuesta con PIDs (Process/Packet IDs). Su enfoque es ser la **base sólida** sobre la cual construir aplicaciones como chats, notificaciones push, sistemas de colaboración en tiempo real, juegos multijugador, monitoreo de estado, o cualquier servicio que requiera comunicación bidireccional persistente.

### Características Principales

✅ **Gestión de conexiones WebSocket** - Manejo automático del ciclo de vida de conexiones  
✅ **Sistema de acknowledgments (ACK)** - Confirmaciones bidireccionales confiables  
✅ **Mensajes genéricos con payloads flexibles** - Estructura de datos completamente personalizable  
✅ **Broadcasting y comunicación peer-to-peer** - Envío masivo y comunicación directa  
✅ **Manejo de timeouts y errores** - Control robusto de fallos y temporización  
✅ **Callbacks personalizables** - Lógica de aplicación completamente inyectable

## 2. Arquitectura

El paquete se centra en los siguientes componentes y principios arquitectónicos:

*   **`ConnectionManager[TUserData]`**: Es el orquestador central. Gestiona todas las conexiones activas, la configuración, y los callbacks proporcionados por la aplicación. Utiliza un tipo genérico `TUserData` que permite al desarrollador asociar datos personalizados (ej. información del perfil de usuario) con cada conexión.
*   **`Connection[TUserData]`**: Representa una única conexión WebSocket activa de un usuario. Cada conexión almacena su ID de usuario, la instancia de `*websocket.Conn`, un canal dedicado (`SendChan`) para mensajes salientes, el contexto para su ciclo de vida, y los datos `TUserData`.
*   **Concurrencia Segura**: 
    *   Se utiliza `sync.Map` para el almacenamiento concurrente y eficiente de conexiones activas y PIDs pendientes (para acks y respuestas).
    *   Cada conexión tiene sus propias goroutines `readPump` (para leer mensajes del cliente) y `writePump` (para escribir mensajes al cliente), asegurando que las operaciones por conexión no se bloqueen entre sí.
    *   Las escrituras a una misma conexión WebSocket se serializan a través del `SendChan` de la conexión y su `writePump`.
*   **Callbacks para Lógica de Aplicación**: La lógica específica del negocio se inyecta a través de la struct `Callbacks[TUserData]`:
    *   `AuthenticateAndGetUserData`: Valida la solicitud HTTP y obtiene el ID y los datos del usuario antes de actualizar a WebSocket.
    *   `OnConnect`: Se ejecuta cuando una nueva conexión se establece y autentica.
    *   `OnDisconnect`: Se ejecuta cuando una conexión se cierra (limpia o por error).
    *   `ProcessClientMessage`: Procesa los mensajes entrantes del cliente (excepto los `ClientAck` que se manejan internamente).
    *   `GeneratePID`: (Opcional) Permite personalizar la generación de IDs de mensajes.
*   **Protocolo de Mensajería Estructurado** (ver `pkg/customws/types/types.go`):
    *   `ClientToServerMessage` y `ServerToClientMessage`: Definen la estructura de los mensajes, incluyendo `PID` (para rastreo y correlación), `Type` (para enrutamiento de la lógica), y `Payload` (para datos arbitrarios en formato JSON).
    *   Soporte para diferentes `MessageType` predefinidos (datos genéricos, errores, acks, etc.) y la posibilidad de añadir más.
*   **Sistema de Confirmaciones (Acks) y Solicitud/Respuesta**:
    *   El servidor puede enviar mensajes y esperar una confirmación (`ClientAck`) del cliente usando el método `SendForClientAck`.
    *   El servidor puede enviar una solicitud y esperar una respuesta más específica del cliente (no solo un ack) usando `SendRequestAndWaitClientResponse`.
    *   Estos mecanismos utilizan PIDs para correlacionar mensajes y respuestas, y tienen timeouts configurables.
*   **Gestión del Ciclo de Vida**: 
    *   Se utiliza `context.Context` extensivamente para gestionar el ciclo de vida de las goroutines del `ConnectionManager` y de cada `Connection`.
    *   Se manejan pings/pongs para mantener vivas las conexiones y detectar cierres inesperados.
    *   Proporciona un método `Shutdown` para cerrar ordenadamente todas las conexiones y liberar recursos.
*   **Configuración**: Una struct `types.Config` permite ajustar timeouts (escritura, pong, ack, request), tamaños de buffer de mensajes y canales.

## 3. Características Principales en Detalle

### 3.1. ✅ Gestión de Conexiones WebSocket

El `ConnectionManager` proporciona gestión automática y robusta del ciclo de vida de conexiones WebSocket.

#### Características:
- **Conexiones concurrentes**: Manejo seguro de miles de conexiones simultáneas
- **Auto-limpieza**: Detección y limpieza automática de conexiones muertas
- **Gestión de contextos**: Cada conexión tiene su propio contexto cancelable
- **Prevención de duplicados**: Cierre automático de conexiones duplicadas del mismo usuario

#### Ejemplo de uso:

```go
// Definir datos personalizados para cada usuario
type MyUserData struct {
    UserID    int64
    Username  string
    Roles     []string
    LastSeen  time.Time
}

// Configurar el manager
cfg := types.DefaultConfig()
cfg.MaxMessageSize = 4096 // 4KB por mensaje
cfg.SendChannelBuffer = 1024 // Buffer más grande para mensajes salientes

callbacks := customws.Callbacks[MyUserData]{
    AuthenticateAndGetUserData: func(r *http.Request) (int64, MyUserData, error) {
        // Tu lógica de autenticación aquí
        token := r.Header.Get("Authorization")
        userID, username, roles, err := validateToken(token)
        if err != nil {
            return 0, MyUserData{}, err
        }
        
        userData := MyUserData{
            UserID:   userID,
            Username: username,
            Roles:    roles,
            LastSeen: time.Now(),
        }
        return userID, userData, nil
    },
    // ... otros callbacks
}

manager := customws.NewConnectionManager[MyUserData](cfg, callbacks)

// Obtener conexión de un usuario específico
if conn, exists := manager.GetConnection(userID); exists {
    fmt.Printf("Usuario %s está conectado\n", conn.UserData.Username)
}
```

### 3.2. ✅ Sistema de Acknowledgments (ACK)

Sistema bidireccional de confirmaciones que garantiza la entrega confiable de mensajes.

#### Características:
- **ACKs automáticos**: El servidor puede requerir confirmación del cliente
- **Timeouts configurables**: Control de tiempo de espera para ACKs
- **Correlación por PID**: Cada mensaje tiene un ID único para rastreo
- **Manejo de errores**: Detección de fallos en ACKs con cleanup automático

#### Ejemplo de implementación:

```go
// Enviar mensaje con confirmación requerida
func sendImportantMessage(conn *customws.Connection[MyUserData], data interface{}) error {
    msg := types.ServerToClientMessage{
        Type:    types.MessageTypeDataEvent,
        Payload: data,
    }
    
    // SendForClientAck espera confirmación del cliente
    ackResponse, err := conn.Manager.SendForClientAck(conn, msg)
    if err != nil {
        return fmt.Errorf("error esperando ACK: %w", err)
    }
    
    // Procesar la confirmación
    var ackPayload types.AckPayload
    payloadBytes, _ := json.Marshal(ackResponse.Payload)
    json.Unmarshal(payloadBytes, &ackPayload)
    
    if ackPayload.Error != "" {
        return fmt.Errorf("cliente reportó error: %s", ackPayload.Error)
    }
    
    fmt.Printf("Mensaje confirmado con estado: %s\n", ackPayload.Status)
    return nil
}

// En ProcessClientMessage, manejar mensajes que requieren ACK del servidor
ProcessClientMessage: func(conn *customws.Connection[MyUserData], msg types.ClientToServerMessage) error {
    switch msg.Type {
    case types.MessageTypeDataRequest:
        // Procesar la solicitud
        result, err := processUserRequest(msg.Payload)
        
        // Enviar ACK del servidor (opcional pero recomendado)
        if msg.PID != "" {
            if err != nil {
                conn.SendServerAck(msg.PID, "failed", err)
            } else {
                conn.SendServerAck(msg.PID, "processed", nil)
            }
        }
        
        // Enviar resultado si es necesario
        if result != nil {
            responseMsg := types.ServerToClientMessage{
                Type:    types.MessageTypeDataEvent,
                Payload: result,
            }
            return conn.SendMessage(responseMsg)
        }
    }
    return nil
}
```

### 3.3. ✅ Mensajes Genéricos con Payloads Flexibles

Sistema de mensajería completamente flexible que permite cualquier estructura de datos.

#### Características:
- **Payloads JSON flexibles**: `interface{}` permite cualquier estructura
- **Tipos de mensaje extensibles**: Fácil adición de nuevos tipos
- **Validación personalizable**: Control total sobre validación de datos
- **Serialización automática**: Manejo transparente de JSON

#### Ejemplo de tipos de mensaje personalizados:

```go
// Definir tipos de mensaje personalizados
const (
    MessageTypeUserProfile    types.MessageType = "user_profile"
    MessageTypeFileUpload     types.MessageType = "file_upload"
    MessageTypeNotification   types.MessageType = "notification"
    MessageTypeGameMove       types.MessageType = "game_move"
)

// Definir estructuras de payload personalizadas
type UserProfilePayload struct {
    Action string                 `json:"action"` // "get", "update"
    Data   map[string]interface{} `json:"data"`
}

type FileUploadPayload struct {
    FileName string `json:"fileName"`
    FileSize int64  `json:"fileSize"`
    MimeType string `json:"mimeType"`
    ChunkNum int    `json:"chunkNum"`
    Data     string `json:"data"` // Base64 encoded chunk
}

type NotificationPayload struct {
    Type    string    `json:"type"`
    Title   string    `json:"title"`
    Message string    `json:"message"`
    Urgent  bool      `json:"urgent"`
    ExpireAt time.Time `json:"expireAt"`
}

// Implementar en ProcessClientMessage
ProcessClientMessage: func(conn *customws.Connection[MyUserData], msg types.ClientToServerMessage) error {
    switch msg.Type {
    case MessageTypeUserProfile:
        var payload UserProfilePayload
        if err := decodePayload(msg.Payload, &payload); err != nil {
            return err
        }
        return handleUserProfile(conn, payload)
        
    case MessageTypeFileUpload:
        var payload FileUploadPayload
        if err := decodePayload(msg.Payload, &payload); err != nil {
            return err
        }
        return handleFileUpload(conn, payload)
        
    case MessageTypeNotification:
        var payload NotificationPayload
        if err := decodePayload(msg.Payload, &payload); err != nil {
            return err
        }
        return handleNotification(conn, payload)
    }
    return nil
}

// Helper para decodificar payloads
func decodePayload(payload interface{}, target interface{}) error {
    payloadBytes, err := json.Marshal(payload)
    if err != nil {
        return err
    }
    return json.Unmarshal(payloadBytes, target)
}
```

### 3.4. ✅ Broadcasting y Comunicación Peer-to-Peer

Sistema robusto para envío masivo y comunicación directa entre usuarios.

#### Características:
- **Broadcasting masivo**: Envío a todos los usuarios conectados
- **Broadcasting selectivo**: Envío a grupos específicos de usuarios
- **Comunicación directa**: Mensajes peer-to-peer entre usuarios
- **Exclusiones**: Capacidad de excluir usuarios específicos del broadcasting
- **Manejo de errores por usuario**: Reporte detallado de fallos de envío

#### Ejemplo de implementación completa:

```go
// Broadcasting a todos los usuarios
func broadcastSystemNotification(manager *customws.ConnectionManager[MyUserData], message string) {
    notification := types.ServerToClientMessage{
        Type: MessageTypeNotification,
        Payload: NotificationPayload{
            Type:    "system",
            Title:   "Notificación del Sistema",
            Message: message,
            Urgent:  false,
        },
    }
    
    // Enviar a todos, excluyendo administradores si es necesario
    adminUserIDs := []int64{1, 2, 3} // IDs de administradores
    errors := manager.BroadcastToAll(notification, adminUserIDs...)
    
    // Manejar errores de envío
    if len(errors) > 0 {
        for userID, err := range errors {
            log.Printf("Error enviando notificación a usuario %d: %v", userID, err)
        }
    }
}

// Broadcasting selectivo a grupos de usuarios
func broadcastToRoleGroup(manager *customws.ConnectionManager[MyUserData], role string, message interface{}) {
    var targetUsers []int64
    
    // Iterar conexiones para encontrar usuarios con el rol específico
    manager.connections.Range(func(key, value interface{}) bool {
        userID := key.(int64)
        conn := value.(*customws.Connection[MyUserData])
        
        // Verificar si el usuario tiene el rol requerido
        for _, userRole := range conn.UserData.Roles {
            if userRole == role {
                targetUsers = append(targetUsers, userID)
                break
            }
        }
        return true
    })
    
    if len(targetUsers) > 0 {
        msg := types.ServerToClientMessage{
            Type:    types.MessageTypeDataEvent,
            Payload: message,
        }
        
        errors := manager.BroadcastToUsers(targetUsers, msg)
        log.Printf("Enviado mensaje a %d usuarios con rol '%s', %d errores", 
                   len(targetUsers), role, len(errors))
    }
}

// Comunicación peer-to-peer
func sendDirectMessage(manager *customws.ConnectionManager[MyUserData], fromUserID, toUserID int64, message interface{}) error {
    msg := types.ServerToClientMessage{
        Type:       types.MessageTypeDataEvent,
        FromUserID: fromUserID,
        Payload:    message,
    }
    
    return manager.SendMessageToUser(toUserID, msg)
}

// Sistema de chat peer-to-peer completo
type ChatMessage struct {
    Text      string    `json:"text"`
    Timestamp time.Time `json:"timestamp"`
    MessageID string    `json:"messageId"`
}

func handlePeerToPeerChat(conn *customws.Connection[MyUserData], msg types.ClientToServerMessage) error {
    if msg.TargetUserID == 0 {
        return errors.New("targetUserId es requerido para mensajes de chat")
    }
    
    // Decodificar mensaje de chat
    var chatData ChatMessage
    if err := decodePayload(msg.Payload, &chatData); err != nil {
        return err
    }
    
    // Agregar metadata del servidor
    chatData.Timestamp = time.Now()
    chatData.MessageID = uuid.NewString()
    
    // Guardar en base de datos (tu implementación)
    if err := saveChatMessage(conn.ID, msg.TargetUserID, chatData); err != nil {
        return err
    }
    
    // Reenviar mensaje al usuario destino
    forwardMsg := types.ServerToClientMessage{
        Type:       types.MessageTypeDataEvent,
        FromUserID: conn.ID,
        Payload:    chatData,
    }
    
    if err := conn.Manager.SendMessageToUser(msg.TargetUserID, forwardMsg); err != nil {
        log.Printf("Error reenviando mensaje de chat: %v", err)
        // El mensaje ya está guardado, así que no es un error crítico
    }
    
    // Confirmar al remitente si tiene PID
    if msg.PID != "" {
        conn.SendServerAck(msg.PID, "delivered", nil)
    }
    
    return nil
}
```

### 3.5. ✅ Manejo de Timeouts y Errores

Sistema robusto de control de errores y temporización que garantiza la estabilidad del servicio.

#### Características:
- **Timeouts configurables**: Control granular de tiempos de espera
- **Cleanup automático**: Limpieza automática de recursos huérfanos
- **Detección de conexiones muertas**: Ping/Pong automático para detectar desconexiones
- **Manejo graceful de errores**: Cierre ordenado de conexiones problemáticas
- **Logging detallado**: Trazabilidad completa de errores y eventos

#### Configuración de timeouts:

```go
func createProductionConfig() types.Config {
    return types.Config{
        WriteWait:         5 * time.Second,   // Timeout para escrituras WebSocket
        PongWait:          30 * time.Second,  // Timeout para recibir pong del cliente
        PingPeriod:        25 * time.Second,  // Frecuencia de envío de pings (< PongWait)
        MaxMessageSize:    8192,              // 8KB máximo por mensaje
        SendChannelBuffer: 256,               // Buffer del canal de envío
        AckTimeout:        10 * time.Second,  // Timeout para esperar ACKs
        RequestTimeout:    15 * time.Second,  // Timeout para respuestas de solicitudes
        AllowedOrigins:    []string{"https://myapp.com"}, // Solo dominios de confianza
    }
}
```

#### Manejo de errores personalizado:

```go
// Wrapper para manejo seguro de operaciones WebSocket
func safeWebSocketOperation(conn *customws.Connection[MyUserData], operation func() error) error {
    defer func() {
        if r := recover(); r != nil {
            log.Printf("Pánico recuperado en operación WebSocket para usuario %d: %v", conn.ID, r)
            conn.Close() // Cerrar conexión problemática
        }
    }()
    
    return operation()
}

// Implementación robusta de ProcessClientMessage con manejo de errores
ProcessClientMessage: func(conn *customws.Connection[MyUserData], msg types.ClientToServerMessage) error {
    // Validación de entrada
    if msg.Type == "" {
        conn.SendErrorNotification(msg.PID, 400, "Tipo de mensaje requerido")
        return errors.New("tipo de mensaje vacío")
    }
    
    // Timeout por operación
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    // Procesar mensaje con timeout
    errChan := make(chan error, 1)
    go func() {
        errChan <- processMessageSafely(ctx, conn, msg)
    }()
    
    select {
    case err := <-errChan:
        if err != nil {
            // Error específico por tipo
            switch {
            case errors.Is(err, ErrInvalidPayload):
                conn.SendErrorNotification(msg.PID, 422, "Payload inválido: "+err.Error())
            case errors.Is(err, ErrUnauthorized):
                conn.SendErrorNotification(msg.PID, 403, "No autorizado")
            case errors.Is(err, ErrRateLimit):
                conn.SendErrorNotification(msg.PID, 429, "Límite de velocidad excedido")
            default:
                conn.SendErrorNotification(msg.PID, 500, "Error interno del servidor")
                log.Printf("Error procesando mensaje de usuario %d: %v", conn.ID, err)
            }
        }
        return err
        
    case <-ctx.Done():
        conn.SendErrorNotification(msg.PID, 408, "Timeout procesando mensaje")
        return ctx.Err()
    }
}

// Errores específicos de la aplicación
var (
    ErrInvalidPayload = errors.New("payload inválido")
    ErrUnauthorized   = errors.New("no autorizado")
    ErrRateLimit      = errors.New("límite de velocidad excedido")
)

// Sistema de rate limiting por usuario
type RateLimiter struct {
    requests map[int64][]time.Time
    mutex    sync.RWMutex
    limit    int
    window   time.Duration
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
    return &RateLimiter{
        requests: make(map[int64][]time.Time),
        limit:    limit,
        window:   window,
    }
}

func (r *RateLimiter) Allow(userID int64) bool {
    r.mutex.Lock()
    defer r.mutex.Unlock()
    
    now := time.Now()
    cutoff := now.Add(-r.window)
    
    // Limpiar requests antiguos
    if requests, exists := r.requests[userID]; exists {
        validRequests := requests[:0]
        for _, reqTime := range requests {
            if reqTime.After(cutoff) {
                validRequests = append(validRequests, reqTime)
            }
        }
        r.requests[userID] = validRequests
    }
    
    // Verificar límite
    if len(r.requests[userID]) >= r.limit {
        return false
    }
    
    // Agregar nueva request
    r.requests[userID] = append(r.requests[userID], now)
    return true
}
```

### 3.6. ✅ Callbacks Personalizables

Sistema flexible de callbacks que permite inyectar completamente la lógica de aplicación.

#### Características:
- **Autenticación personalizable**: Control total sobre validación de usuarios
- **Hooks de ciclo de vida**: Eventos de conexión y desconexión
- **Procesamiento de mensajes**: Lógica de negocio completamente inyectable
- **Generación de IDs**: Personalización de generación de PIDs
- **Validación por contexto**: Acceso completo a datos de request y usuario

#### Implementación completa de callbacks:

```go
// Implementación avanzada de callbacks para una aplicación de colaboración
func createCollaborationAppCallbacks(db *sql.DB, redis *redis.Client) customws.Callbacks[MyUserData] {
    rateLimiter := NewRateLimiter(100, time.Minute) // 100 requests por minuto
    
    return customws.Callbacks[MyUserData]{
        // Autenticación JWT con verificación de base de datos
        AuthenticateAndGetUserData: func(r *http.Request) (int64, MyUserData, error) {
            // Múltiples métodos de autenticación
            var token string
            
            // 1. Authorization header
            if auth := r.Header.Get("Authorization"); auth != "" {
                if strings.HasPrefix(auth, "Bearer ") {
                    token = auth[7:]
                }
            }
            
            // 2. Query parameter (para casos especiales)
            if token == "" {
                token = r.URL.Query().Get("token")
            }
            
            // 3. Cookie (para web apps)
            if token == "" {
                if cookie, err := r.Cookie("auth_token"); err == nil {
                    token = cookie.Value
                }
            }
            
            if token == "" {
                return 0, MyUserData{}, errors.New("token de autenticación requerido")
            }
            
            // Validar JWT
            claims, err := validateJWT(token)
            if err != nil {
                return 0, MyUserData{}, fmt.Errorf("token inválido: %w", err)
            }
            
            // Verificar en base de datos
            userData, err := getUserFromDB(db, claims.UserID)
            if err != nil {
                return 0, MyUserData{}, fmt.Errorf("usuario no encontrado: %w", err)
            }
            
            // Verificar que el usuario esté activo
            if !userData.IsActive {
                return 0, MyUserData{}, errors.New("cuenta de usuario desactivada")
            }
            
            // Verificar permisos para WebSocket
            if !hasWebSocketPermission(userData.Roles) {
                return 0, MyUserData{}, errors.New("sin permisos para conexión WebSocket")
            }
            
            return userData.UserID, userData, nil
        },
        
        // Hook de conexión con lógica de negocio
        OnConnect: func(conn *customws.Connection[MyUserData]) error {
            log.Printf("Usuario conectado: %s (ID: %d) desde %s", 
                      conn.UserData.Username, conn.ID, "IP_ADDRESS")
            
            // 1. Actualizar estado en base de datos
            if err := updateUserOnlineStatus(db, conn.ID, true); err != nil {
                log.Printf("Error actualizando estado online: %v", err)
            }
            
            // 2. Notificar a contactos/colaboradores
            if err := notifyUserOnline(conn); err != nil {
                log.Printf("Error notificando conexión: %v", err)
            }
            
            // 3. Enviar datos de inicialización
            initData := map[string]interface{}{
                "user":         conn.UserData,
                "serverTime":   time.Now(),
                "capabilities": []string{"file_upload", "video_call", "screen_share"},
            }
            
            initMsg := types.ServerToClientMessage{
                Type:    MessageTypeUserProfile,
                Payload: initData,
            }
            
            if err := conn.SendMessage(initMsg); err != nil {
                log.Printf("Error enviando datos de inicialización: %v", err)
                return err
            }
            
            // 4. Cargar mensajes pendientes desde Redis
            go loadPendingMessages(redis, conn)
            
            return nil
        },
        
        // Hook de desconexión con cleanup
        OnDisconnect: func(conn *customws.Connection[MyUserData], err error) {
            if err != nil {
                log.Printf("Usuario desconectado por error: %s (ID: %d) - %v", 
                          conn.UserData.Username, conn.ID, err)
            } else {
                log.Printf("Usuario desconectado limpiamente: %s (ID: %d)", 
                          conn.UserData.Username, conn.ID)
            }
            
            // 1. Actualizar estado en base de datos
            updateUserOnlineStatus(db, conn.ID, false)
            
            // 2. Notificar desconexión
            notifyUserOffline(conn)
            
            // 3. Limpiar recursos específicos del usuario
            cleanupUserResources(redis, conn.ID)
            
            // 4. Guardar estadísticas de sesión
            saveSessionStats(db, conn.ID, conn.UserData.LastSeen, time.Now())
        },
        
        // Procesamiento de mensajes con lógica de negocio completa
        ProcessClientMessage: func(conn *customws.Connection[MyUserData], msg types.ClientToServerMessage) error {
            // Rate limiting por usuario
            if !rateLimiter.Allow(conn.ID) {
                return ErrRateLimit
            }
            
            // Logging detallado
            log.Printf("Procesando mensaje: Usuario %d, Tipo %s, PID %s", 
                      conn.ID, msg.Type, msg.PID)
            
            // Actualizar última actividad
            conn.UserData.LastSeen = time.Now()
            
            // Enrutar por tipo de mensaje
            switch msg.Type {
            case types.MessageTypeDataRequest:
                return handleDataRequest(conn, msg)
            case types.MessageTypePresenceUpdate:
                return handlePresenceUpdate(conn, msg)
            case MessageTypeFileUpload:
                return handleFileUpload(conn, msg)
            case MessageTypeNotification:
                return handleNotificationRequest(conn, msg)
            default:
                return fmt.Errorf("tipo de mensaje no soportado: %s", msg.Type)
            }
        },
        
        // Generación personalizada de PIDs con metadatos
        GeneratePID: func() string {
            // Incluir timestamp y número aleatorio para debugging
            timestamp := time.Now().UnixNano()
            random := rand.Intn(10000)
            return fmt.Sprintf("%d-%04d-%s", timestamp, random, uuid.NewString()[:8])
        },
    }
}

// Funciones de soporte para los callbacks
func validateJWT(token string) (*JWTClaims, error) {
    // Tu implementación de validación JWT
    return nil, nil
}

func getUserFromDB(db *sql.DB, userID int64) (MyUserData, error) {
    // Tu implementación de consulta a BD
    return MyUserData{}, nil
}

func hasWebSocketPermission(roles []string) bool {
    allowedRoles := []string{"user", "admin", "moderator"}
    for _, role := range roles {
        for _, allowed := range allowedRoles {
            if role == allowed {
                return true
            }
        }
    }
    return false
}

func notifyUserOnline(conn *customws.Connection[MyUserData]) error {
    // Notificar a contactos que el usuario está online
    onlineNotification := types.ServerToClientMessage{
        Type:       types.MessageTypePresenceEvent,
        FromUserID: conn.ID,
        Payload: map[string]interface{}{
            "status": "online",
            "user":   conn.UserData.Username,
        },
    }
    
    // Obtener lista de contactos del usuario y notificarles
    contactIDs := getUserContacts(conn.ID) // Tu implementación
    errors := conn.Manager.BroadcastToUsers(contactIDs, onlineNotification)
    
    if len(errors) > 0 {
        log.Printf("Errores notificando conexión: %v", errors)
    }
    
    return nil
}
```

## 4. Cómo Empezar

### 4.1. Definir UserData (Opcional pero Recomendado)

Crea una struct que contenga la información de tu aplicación que quieres asociar con cada conexión. Este será tu tipo `TUserData`.

```go
package main

// Ejemplo de UserData para una aplicación de colaboración
type MyUserData struct {
    UserID    int64     `json:"userId"`
    Username  string    `json:"username"`
    Email     string    `json:"email"`
    Roles     []string  `json:"roles"`
    LastSeen  time.Time `json:"lastSeen"`
    IsActive  bool      `json:"isActive"`
    Workspace string    `json:"workspace"`
}
```

### 4.2. Implementar los Callbacks

```go
import (
    "net/http"
    "fmt"
    "errors"
    "github.com/user/repo/backend/pkg/customws"
    "github.com/user/repo/backend/pkg/customws/types"
    // ... tus otros imports (ej. para BD, auth)
)

func setupCallbacks() customws.Callbacks[MyUserData] {
    return customws.Callbacks[MyUserData]{
        AuthenticateAndGetUserData: func(r *http.Request) (userID int64, userData MyUserData, err error) {
            // Ejemplo: Validar token JWT del header Authorization
            authHeader := r.Header.Get("Authorization")
            if !strings.HasPrefix(authHeader, "Bearer ") {
                return 0, MyUserData{}, errors.New("token Bearer requerido")
            }
            
            token := authHeader[7:] // Remover "Bearer "

            // Aquí tu lógica para validar el token y obtener info del usuario de la BD
            // claims, err := myauth.ValidateToken(token)
            // if err != nil { return 0, MyUserData{}, err }
            // userInfoFromDB, err := mydatabase.GetUserByID(claims.UserID)
            // if err != nil { return 0, MyUserData{}, err }

            // Ejemplo simplificado:
            if token == "valid-token-user1" {
                userData = MyUserData{
                    UserID:    1,
                    Username:  "UserOne",
                    Email:     "user1@example.com",
                    Roles:     []string{"user", "collaborator"},
                    LastSeen:  time.Now(),
                    IsActive:  true,
                    Workspace: "main",
                }
                return userData.UserID, userData, nil
            }
            return 0, MyUserData{}, errors.New("token inválido")
        },

        OnConnect: func(conn *customws.Connection[MyUserData]) error {
            fmt.Printf("Usuario conectado: ID %d, Username %s, Workspace %s\n", 
                      conn.ID, conn.UserData.Username, conn.UserData.Workspace)
            
            // Enviar datos de inicialización
            initData := map[string]interface{}{
                "welcome":     "Bienvenido al sistema",
                "serverTime":  time.Now(),
                "userData":    conn.UserData,
                "capabilities": []string{"messaging", "file_sharing", "notifications"},
            }
            
            initMsg := types.ServerToClientMessage{
                Type:    types.MessageTypeDataEvent,
                Payload: initData,
            }
            
            if err := conn.SendMessage(initMsg); err != nil {
                fmt.Printf("Error enviando mensaje de bienvenida: %v\n", err)
                return err
            }
            
            // Notificar a otros usuarios del workspace
            presenceMsg := types.ServerToClientMessage{
                Type:       types.MessageTypePresenceEvent,
                FromUserID: conn.ID,
                Payload: map[string]interface{}{
                    "status":    "online",
                    "username":  conn.UserData.Username,
                    "workspace": conn.UserData.Workspace,
                },
            }
            
            // Broadcast a usuarios del mismo workspace (implementar lógica de filtrado)
            // conn.Manager.BroadcastToAll(presenceMsg, conn.ID) // Excluir al usuario actual
            
            return nil
        },

        OnDisconnect: func(conn *customws.Connection[MyUserData], err error) {
            if err != nil {
                fmt.Printf("Usuario desconectado por error: ID %d, Username %s. Error: %v\n", 
                          conn.ID, conn.UserData.Username, err)
            } else {
                fmt.Printf("Usuario desconectado limpiamente: ID %d, Username %s\n", 
                          conn.ID, conn.UserData.Username)
            }
            
            // Notificar desconexión a otros usuarios
            offlineMsg := types.ServerToClientMessage{
                Type:       types.MessageTypePresenceEvent,
                FromUserID: conn.ID,
                Payload: map[string]interface{}{
                    "status":   "offline",
                    "username": conn.UserData.Username,
                },
            }
            
            // Broadcast de desconexión
            // conn.Manager.BroadcastToAll(offlineMsg, conn.ID)
        },

        ProcessClientMessage: func(conn *customws.Connection[MyUserData], msg types.ClientToServerMessage) error {
            fmt.Printf("Mensaje recibido de UserID %d (PID: %s, Tipo: %s)\n", 
                      conn.ID, msg.PID, msg.Type)

            switch msg.Type {
            case types.MessageTypeDataRequest:
                return handleDataRequest(conn, msg)
                
            case types.MessageTypePresenceUpdate:
                return handlePresenceUpdate(conn, msg)

            case types.MessageTypeGenericRequest:
                return handleGenericRequest(conn, msg)
                
            default:
                fmt.Printf("Tipo de mensaje desconocido: %s\n", msg.Type)
                if msg.PID != "" {
                    conn.SendErrorNotification(msg.PID, 400, "Tipo de mensaje no soportado")
                }
                return errors.New("tipo de mensaje no soportado")
            }
        },
        
        // GeneratePID personalizado (opcional)
        GeneratePID: func() string {
            return fmt.Sprintf("msg_%d_%s", time.Now().UnixNano(), uuid.NewString()[:8])
        },
    }
}

// Funciones de manejo de mensajes
func handleDataRequest(conn *customws.Connection[MyUserData], msg types.ClientToServerMessage) error {
    // Decodificar payload
    var requestData map[string]interface{}
    payloadBytes, _ := json.Marshal(msg.Payload)
    if err := json.Unmarshal(payloadBytes, &requestData); err != nil {
        return err
    }
    
    fmt.Printf("Solicitud de datos de %s: %v\n", conn.UserData.Username, requestData)
    
    // Procesar solicitud (tu lógica aquí)
    responseData := map[string]interface{}{
        "status":    "success",
        "data":      "Datos solicitados",
        "timestamp": time.Now(),
    }
    
    // Enviar respuesta
    responseMsg := types.ServerToClientMessage{
        Type:    types.MessageTypeDataEvent,
        Payload: responseData,
    }
    
    if err := conn.SendMessage(responseMsg); err != nil {
        return err
    }
    
    // Enviar ACK si el cliente lo espera
    if msg.PID != "" {
        conn.SendServerAck(msg.PID, "processed", nil)
    }
    
    return nil
}

func handlePresenceUpdate(conn *customws.Connection[MyUserData], msg types.ClientToServerMessage) error {
    var presenceData types.PresenceUpdatePayload
    payloadBytes, _ := json.Marshal(msg.Payload)
    if err := json.Unmarshal(payloadBytes, &presenceData); err != nil {
        return err
    }
    
    fmt.Printf("Actualización de presencia de %s: %s\n", conn.UserData.Username, presenceData.Status)
    
    // Retransmitir a otros usuarios
    presenceEvent := types.ServerToClientMessage{
        Type:       types.MessageTypePresenceEvent,
        FromUserID: conn.ID,
        Payload:    presenceData,
    }
    
    // Si es para un usuario específico
    if presenceData.TargetUserID != 0 {
        return conn.Manager.SendMessageToUser(presenceData.TargetUserID, presenceEvent)
    }
    
    // Broadcast general
    conn.Manager.BroadcastToAll(presenceEvent, conn.ID)
    return nil
}

func handleGenericRequest(conn *customws.Connection[MyUserData], msg types.ClientToServerMessage) error {
    // El cliente está esperando una respuesta específica con el mismo PID
    
    // Procesar la solicitud
    responseData := map[string]interface{}{
        "requestProcessed": true,
        "result":          "Operación completada exitosamente",
        "timestamp":       time.Now(),
    }
    
    // Responder con el mismo PID
    response := types.ServerToClientMessage{
        PID:     msg.PID, // ¡Importante: mismo PID!
        Type:    types.MessageTypeGenericResponse,
        Payload: responseData,
    }
    
    return conn.SendMessage(response)
}
```

### 4.3. Crear e Iniciar el `ConnectionManager`

```go
package main

import (
    "net/http"
    "log"
    "time"
    "context"
    "os"
    "os/signal"
    "syscall"
    "github.com/user/repo/backend/pkg/customws"
    "github.com/user/repo/backend/pkg/customws/types"
)

func main() {
    // Configuración personalizada
    cfg := types.Config{
        WriteWait:         10 * time.Second,
        PongWait:          60 * time.Second,
        PingPeriod:        50 * time.Second,
        MaxMessageSize:    4096, // 4KB por mensaje
        SendChannelBuffer: 512,
        AckTimeout:        5 * time.Second,
        RequestTimeout:    10 * time.Second,
        AllowedOrigins:    []string{"http://localhost:5173", "https://myapp.com"}, // ¡IMPORTANTE!
    }

    callbacks := setupCallbacks()
    
    // Crear el ConnectionManager con tu tipo de UserData
    connManager := customws.NewConnectionManager[MyUserData](cfg, callbacks)

    // Configurar endpoints HTTP
    http.HandleFunc("/ws", connManager.ServeHTTP)
    
    // Endpoint de salud (opcional)
    http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("OK"))
    })

    serverAddr := ":8082"
    log.Printf("Servidor WebSocket escuchando en %s/ws", serverAddr)

    // Configurar servidor HTTP
    server := &http.Server{
        Addr:         serverAddr,
        ReadTimeout:  15 * time.Second,
        WriteTimeout: 15 * time.Second,
        IdleTimeout:  60 * time.Second,
    }

    // Manejar señales para shutdown graceful
    go func() {
        sigChan := make(chan os.Signal, 1)
        signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
        <-sigChan
        
        log.Println("Señal de cierre recibida, iniciando shutdown graceful...")
        
        // Shutdown del ConnectionManager
        shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
        defer cancel()
        
        if err := connManager.Shutdown(shutdownCtx); err != nil {
            log.Printf("Error durante shutdown del ConnectionManager: %v", err)
        } else {
            log.Println("ConnectionManager cerrado exitosamente")
        }
        
        // Shutdown del servidor HTTP
        if err := server.Shutdown(shutdownCtx); err != nil {
            log.Printf("Error durante shutdown del servidor HTTP: %v", err)
        }
    }()

    // Iniciar servidor
    if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
        log.Fatalf("No se pudo iniciar el servidor HTTP: %v", err)
    }
    
    log.Println("Servidor cerrado exitosamente")
}
```

### 4.4. Dependencias

Asegúrate de tener las dependencias necesarias en tu `go.mod`:

```bash
go get github.com/gorilla/websocket
go get github.com/google/uuid
```

Tu `go.mod` debería incluir:
```go
module tu-proyecto

go 1.21

require (
    github.com/gorilla/websocket v1.5.0
    github.com/google/uuid v1.3.0
)
```

## 5. Protocolo de Mensajería para el Cliente WebSocket

Un cliente que desee conectarse a un servidor `customws` debe adherirse al siguiente protocolo:

### 5.1. Conexión
*   Establecer una conexión WebSocket al endpoint definido (ej. `ws://localhost:8082/ws`).
*   La autenticación se maneja según lo implementado en el callback `AuthenticateAndGetUserData` (ej. un token JWT en el header `Authorization: Bearer <token>`).

### 5.2. Formato de Mensajes (Cliente -> Servidor)

El cliente debe enviar mensajes JSON que sigan la estructura `types.ClientToServerMessage`:

```json
{
  "pid": "opcional-client-generated-uuid-123", // Opcional: ID de Proceso/Petición. Usar si se espera un ServerAck o para rastreo.
  "type": "data_request",                     // Obligatorio: Uno de los MessageType definidos o uno personalizado.
  "targetUserId": 456,                       // Opcional: Para mensajes directos (ej. en comunicación P2P).
  "payload": {                               // Obligatorio: Contenido del mensaje.
    "action": "get_user_profile",
    "data": {
      "userId": 123,
      "fields": ["username", "email", "roles"]
    }
  }
}
```

#### Tipos de mensaje predefinidos (Cliente -> Servidor):

- **`data_request`**: Solicitud de datos genérica
- **`presence_update`**: Actualización de estado de presencia
- **`client_ack`**: Confirmación de recepción de mensaje del servidor
- **`generic_request`**: Solicitud genérica que espera respuesta específica

#### Ejemplo de mensajes por tipo:

**Solicitud de datos:**
```json
{
  "type": "data_request",
  "payload": {
    "resource": "user_profile",
    "action": "update",
    "data": {
      "username": "nuevo_nombre",
      "bio": "Nueva biografía"
    }
  }
}
```

**Actualización de presencia:**
```json
{
  "type": "presence_update",
  "payload": {
    "status": "typing",
    "targetUserId": 789
  }
}
```

**Solicitud genérica con respuesta esperada:**
```json
{
  "pid": "client-req-123",
  "type": "generic_request",
  "payload": {
    "operation": "search",
    "query": "documentos importantes",
    "filters": {
      "dateRange": "last_week",
      "type": "pdf"
    }
  }
}
```

### 5.3. Formato de Mensajes (Servidor -> Cliente)

El cliente recibirá mensajes JSON que sigan la estructura `types.ServerToClientMessage`:

```json
{
  "pid": "server-generated-uuid-abc-789", // Opcional: ID del mensaje del servidor. Útil si el servidor espera un ClientAck.
  "type": "data_event",
  "fromUserId": 123,
  "payload": {
    "event": "profile_updated",
    "data": {
      "userId": 123,
      "username": "nuevo_nombre",
      "timestamp": "2024-01-15T10:30:00Z"
    }
  },
  "error": null // O un objeto ErrorPayload si algo falló
}
```

#### Tipos de mensaje predefinidos (Servidor -> Cliente):

- **`data_event`**: Evento de datos genérico
- **`presence_event`**: Notificación de cambio de presencia
- **`server_ack`**: Confirmación del servidor
- **`generic_response`**: Respuesta a una solicitud genérica
- **`error_notification`**: Notificación de error

#### Ejemplos de mensajes del servidor:

**Evento de datos:**
```json
{
  "type": "data_event",
  "fromUserId": 456,
  "payload": {
    "event": "new_document",
    "document": {
      "id": "doc_123",
      "title": "Documento importante",
      "author": "Usuario456",
      "createdAt": "2024-01-15T10:30:00Z"
    }
  }
}
```

**Evento de presencia:**
```json
{
  "type": "presence_event",
  "fromUserId": 789,
  "payload": {
    "status": "online",
    "username": "UsuarioEjemplo",
    "lastSeen": "2024-01-15T10:30:00Z"
  }
}
```

**Respuesta genérica:**
```json
{
  "pid": "client-req-123", // Mismo PID de la solicitud
  "type": "search_response", // Un tipo que el servidor espera como respuesta
  "payload": {
    "results": [
      {
        "id": "local_doc_1",
        "title": "Documento local encontrado",
        "content": "Contenido relevante..."
      }
    ],
    "searchTime": 150,
    "source": "client_cache"
  }
}
```

**Notificación de error:**
```json
{
  "type": "error_notification",
  "error": {
    "originalPid": "client-req-123",
    "code": 422,
    "message": "Datos de entrada inválidos: el campo 'query' es requerido"
  }
}
```

### 5.4. Envío de Confirmaciones (ClientAck)

Si el servidor envía un mensaje con un `PID` y espera una confirmación (porque el servidor usó `SendForClientAck`), el cliente debe enviar un `ClientToServerMessage` estructurado de la siguiente manera:

```json
{
  "pid": "client-ack-pid-xyz-456", // PID opcional para este mensaje de ack en sí.
  "type": "client_ack",
  "payload": {
    "acknowledgedPid": "server-generated-uuid-abc-789", // PID del mensaje del servidor que se está confirmando.
    "status": "received", // Opcional: "processed", "read", "failed", etc.
    "error": null // O un string si hubo un error al procesar el mensaje original del servidor en el cliente.
  }
}
```

### 5.5. Envío de Respuestas a Solicitudes del Servidor

Si el servidor envía una solicitud que espera una respuesta específica (porque el servidor usó `SendRequestAndWaitClientResponse`), el cliente debe responder con un `ClientToServerMessage` que:

1.  Tenga el **mismo `PID`** que la solicitud original del servidor.
2.  Tenga un `Type` que el servidor espere como respuesta (esto es específico de la aplicación).
3.  Contenga el `Payload` con los datos de la respuesta.

Ejemplo de respuesta del cliente a una solicitud del servidor con PID `server-req-pid-123`:

```json
{
  "pid": "server-req-pid-123", // ¡MISMO PID que la solicitud del servidor!
  "type": "search_response", // Un tipo que el servidor espera como respuesta
  "payload": {
    "results": [
      {
        "id": "local_doc_1",
        "title": "Documento local encontrado",
        "content": "Contenido relevante..."
      }
    ],
    "searchTime": 150,
    "source": "client_cache"
  }
}
```

## 6. Consideraciones de Escalabilidad y Producción

### 6.1. Seguridad

#### Configuración de Orígenes (CORS)
```go
cfg := types.Config{
    // ¡CRÍTICO! Configurar orígenes permitidos para producción
    AllowedOrigins: []string{
        "https://myapp.com",
        "https://app.mycompany.com",
        // NO incluir "*" o URLs de desarrollo en producción
    },
    // ... otras configuraciones
}
```

#### Validación de entrada robusta
```go
ProcessClientMessage: func(conn *customws.Connection[MyUserData], msg types.ClientToServerMessage) error {
    // 1. Validar tamaño del payload
    if payloadSize := getPayloadSize(msg.Payload); payloadSize > maxPayloadSize {
        return errors.New("payload demasiado grande")
    }
    
    // 2. Validar campos requeridos
    if msg.Type == "" {
        return errors.New("tipo de mensaje requerido")
    }
    
    // 3. Sanitizar datos de entrada
    if err := sanitizePayload(msg.Payload); err != nil {
        return err
    }
    
    // 4. Rate limiting por usuario
    if !rateLimiter.Allow(conn.ID) {
        return errors.New("límite de velocidad excedido")
    }
    
    // ... procesar mensaje
}
```

### 6.2. Optimización de Rendimiento

#### Configuración optimizada para alta concurrencia
```go
func createHighPerformanceConfig() types.Config {
    return types.Config{
        WriteWait:         5 * time.Second,   // Más agresivo para detectar conexiones lentas
        PongWait:          30 * time.Second,  // Reducido para liberar conexiones muertas más rápido
        PingPeriod:        25 * time.Second,
        MaxMessageSize:    2048,              // Limitado para evitar DoS
        SendChannelBuffer: 1024,              // Buffer grande para picos de tráfico
        AckTimeout:        3 * time.Second,   // Más rápido para alta frecuencia
        RequestTimeout:    8 * time.Second,
    }
}
```

#### Pool de objetos para reducir GC pressure
```go
var messagePool = sync.Pool{
    New: func() interface{} {
        return &types.ServerToClientMessage{}
    },
}

func sendOptimizedMessage(conn *customws.Connection[MyUserData], msgType types.MessageType, payload interface{}) error {
    msg := messagePool.Get().(*types.ServerToClientMessage)
    defer messagePool.Put(msg)
    
    // Limpiar mensaje reutilizado
    *msg = types.ServerToClientMessage{
        Type:    msgType,
        Payload: payload,
    }
    
    return conn.SendMessage(*msg)
}
```

### 6.3. Escalabilidad Horizontal

#### Arquitectura distribuida con Redis
```go
// Implementación de broadcasting distribuido
type DistributedBroadcaster struct {
    localManager *customws.ConnectionManager[MyUserData]
    redisClient  *redis.Client
    nodeID       string
}

func (db *DistributedBroadcaster) BroadcastGlobal(msg types.ServerToClientMessage) error {
    // 1. Enviar a conexiones locales
    localErrors := db.localManager.BroadcastToAll(msg)
    
    // 2. Publicar en Redis para otros nodos
    msgBytes, _ := json.Marshal(msg)
    if err := db.redisClient.Publish(context.Background(), "websocket_broadcast", msgBytes).Err(); err != nil {
        log.Printf("Error publicando en Redis: %v", err)
    }
    
    return nil
}

func (db *DistributedBroadcaster) SubscribeToRedis() {
    pubsub := db.redisClient.Subscribe(context.Background(), "websocket_broadcast")
    defer pubsub.Close()
    
    for msg := range pubsub.Channel() {
        var wsMsg types.ServerToClientMessage
        if err := json.Unmarshal([]byte(msg.Payload), &wsMsg); err != nil {
            continue
        }
        
        // Enviar a conexiones locales (sin republicar en Redis)
        db.localManager.BroadcastToAll(wsMsg)
    }
}
```

#### Load balancing con sticky sessions
```nginx
# Configuración Nginx para WebSocket load balancing
upstream websocket_backend {
    ip_hash; # Sticky sessions por IP
    server backend1:8082;
    server backend2:8082;
    server backend3:8082;
}

server {
    listen 443 ssl;
    server_name myapp.com;
    
    location /ws {
        proxy_pass http://websocket_backend;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        # Timeouts para WebSocket
        proxy_read_timeout 3600s;
        proxy_send_timeout 3600s;
    }
}
```

### 6.4. Monitoreo y Observabilidad

#### Métricas de rendimiento
```go
type ConnectionMetrics struct {
    ActiveConnections  int64
    TotalMessages      int64
    ErrorCount         int64
    AverageLatency     float64
    LastUpdate         time.Time
}

func (cm *ConnectionManager[TUserData]) GetMetrics() ConnectionMetrics {
    var activeConnections int64
    cm.connections.Range(func(key, value interface{}) bool {
        activeConnections++
        return true
    })
    
    return ConnectionMetrics{
        ActiveConnections: activeConnections,
        TotalMessages:     atomic.LoadInt64(&cm.totalMessages),
        ErrorCount:        atomic.LoadInt64(&cm.errorCount),
        LastUpdate:        time.Now(),
    }
}

// Endpoint de métricas para Prometheus
func metricsHandler(manager *customws.ConnectionManager[MyUserData]) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        metrics := manager.GetMetrics()
        
        fmt.Fprintf(w, "# HELP websocket_active_connections Current active WebSocket connections\n")
        fmt.Fprintf(w, "# TYPE websocket_active_connections gauge\n")
        fmt.Fprintf(w, "websocket_active_connections %d\n", metrics.ActiveConnections)
        
        fmt.Fprintf(w, "# HELP websocket_total_messages Total messages processed\n")
        fmt.Fprintf(w, "# TYPE websocket_total_messages counter\n")
        fmt.Fprintf(w, "websocket_total_messages %d\n", metrics.TotalMessages)
        
        // ... más métricas
    }
}
```

#### Logging estructurado
```go
import "github.com/sirupsen/logrus"

func setupStructuredLogging() {
    logrus.SetFormatter(&logrus.JSONFormatter{})
    logrus.SetLevel(logrus.InfoLevel)
}

// En los callbacks, usar logging estructurado
OnConnect: func(conn *customws.Connection[MyUserData]) error {
    logrus.WithFields(logrus.Fields{
        "event":     "user_connected",
        "userId":    conn.ID,
        "username":  conn.UserData.Username,
        "userAgent": conn.UserData.UserAgent, // Si lo tienes
        "ip":        getClientIP(conn),        // Si lo tienes
    }).Info("Usuario conectado")
    
    return nil
}
```

### 6.5. Límites del Sistema Operativo

Para manejar conexiones masivas (objetivo de 100k+ conexiones concurrentes):

#### Configuración del sistema
```bash
# /etc/security/limits.conf
* soft nofile 1000000
* hard nofile 1000000

# /etc/sysctl.conf  
net.core.somaxconn = 65535
net.ipv4.tcp_max_syn_backlog = 65535
net.core.netdev_max_backlog = 5000
net.ipv4.tcp_fin_timeout = 30
net.ipv4.tcp_keepalive_time = 120
net.ipv4.tcp_keepalive_probes = 3
net.ipv4.tcp_keepalive_intvl = 15
```

#### Configuración de Go runtime
```go
func init() {
    // Configurar GOMAXPROCS basado en CPUs disponibles
    runtime.GOMAXPROCS(runtime.NumCPU())
    
    // Ajustar GC para alta concurrencia
    debug.SetGCPercent(100) // Valor por defecto, ajustar según necesidad
}

func main() {
    // Configurar límites de memoria
    debug.SetMemoryLimit(8 << 30) // 8GB limit, por ejemplo
    
    // ... resto del código
}
```

### 6.6. Testing de Carga

#### Script de prueba con múltiples conexiones
```go
func loadTest() {
    const numConnections = 10000
    const messagesPerConnection = 100
    
    var wg sync.WaitGroup
    for i := 0; i < numConnections; i++ {
        wg.Add(1)
        go func(connID int) {
            defer wg.Done()
            
            // Crear conexión WebSocket
            ws, _, err := websocket.DefaultDialer.Dial("ws://localhost:8082/ws", http.Header{
                "Authorization": []string{"Bearer valid-token-user1"},
            })
            if err != nil {
                log.Printf("Error conectando %d: %v", connID, err)
                return
            }
            defer ws.Close()
            
            // Enviar mensajes
            for j := 0; j < messagesPerConnection; j++ {
                msg := types.ClientToServerMessage{
                    Type: types.MessageTypeDataRequest,
                    Payload: map[string]interface{}{
                        "test": fmt.Sprintf("mensaje_%d_%d", connID, j),
                    },
                }
                
                if err := ws.WriteJSON(msg); err != nil {
                    log.Printf("Error enviando mensaje %d-%d: %v", connID, j, err)
                    return
                }
                
                time.Sleep(10 * time.Millisecond) // Simular tráfico realista
            }
        }(i)
    }
    
    wg.Wait()
    log.Printf("Prueba de carga completada: %d conexiones, %d mensajes cada una", 
              numConnections, messagesPerConnection)
}
```

Con estas configuraciones y consideraciones, el paquete `customws` puede escalar para manejar decenas de miles de conexiones concurrentes en un solo servidor, y millones de conexiones en una arquitectura distribuida.

---

Este `README.md` proporciona una guía completa y actualizada para entender, implementar y escalar aplicaciones usando el paquete `customws` genérico. 