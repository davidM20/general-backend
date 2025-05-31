package customws

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/davidM20/micro-service-backend-go.git/pkg/customws/types"
	"github.com/davidM20/micro-service-backend-go.git/pkg/logger"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const (
	componentLog = "CUSTOMWS"
)

// UserData es un tipo genérico que el usuario de esta biblioteca puede definir
// para almacenar información específica de la aplicación junto con cada conexión.
// Ejemplo: type MyAppUserData struct { UserID int64; Email string; Roles []string }
// Luego se usaría ConnectionManager[MyAppUserData]

// Connection representa una conexión WebSocket activa para un usuario.
// TUserData es el tipo de datos específicos del usuario que se asociarán con esta conexión.
type Connection[TUserData any] struct {
	ID       int64 // ID del usuario (asumiendo que es int64, como en tu ejemplo original)
	conn     *websocket.Conn
	manager  *ConnectionManager[TUserData]
	SendChan chan types.ServerToClientMessage // Canal para enviar mensajes al cliente.
	UserData TUserData                        // Datos personalizados del usuario.
	ctx      context.Context
	cancel   context.CancelFunc
}

// Manager devuelve el ConnectionManager asociado con esta conexión.
func (c *Connection[TUserData]) Manager() *ConnectionManager[TUserData] {
	return c.manager
}

// Callbacks define las funciones que el usuario de la biblioteca debe implementar
// para manejar eventos y mensajes específicos de la aplicación.
type Callbacks[TUserData any] struct {
	// OnConnect se llama cuando un nuevo cliente establece una conexión exitosa.
	// El parámetro TUserData ya está poblado y asociado con la conexión.
	// Se puede usar para, por ejemplo, marcar al usuario como online en la BD, notificar a contactos, etc.
	OnConnect func(conn *Connection[TUserData]) error

	// OnDisconnect se llama cuando un cliente se desconecta, ya sea de forma limpia o por un error.
	// El error puede ser nil si la desconexión fue limpia.
	OnDisconnect func(conn *Connection[TUserData], err error)

	// ProcessClientMessage se llama para cada mensaje de tipo ClientToServerMessage recibido del cliente,
	// excepto los mensajes de tipo MessageTypeClientAck que son manejados internamente por la biblioteca.
	// La función debe procesar el mensaje y puede retornar un error si algo falla.
	// Si se necesita enviar una respuesta o ack, se deben usar los métodos de Connection (ej. SendMessage, SendServerAck).
	ProcessClientMessage func(conn *Connection[TUserData], msg types.ClientToServerMessage) error

	// AuthenticateAndGetUserData es llamado por ServeHTTP antes de actualizar la conexión a WebSocket.
	// Debe validar la petición (ej. token JWT, cookies) y retornar el ID del usuario (int64) y los datos TUserData.
	// Si la autenticación falla, debe retornar un error y ServeHTTP responderá con HTTP Unauthorized.
	AuthenticateAndGetUserData func(r *http.Request) (userID int64, userData TUserData, err error)

	// GeneratePID (opcional): Si se proporciona, se usará para generar PIDs para mensajes salientes.
	// Si es nil, se usará uuid.NewString().
	GeneratePID func() string
}

// ConnectionManager gestiona todas las conexiones WebSocket activas.
type ConnectionManager[TUserData any] struct {
	config    types.Config
	callbacks Callbacks[TUserData]
	upgrader  websocket.Upgrader

	// connections almacena las conexiones activas, mapeando UserID a *Connection.
	// Se usa sync.Map para concurrencia eficiente.
	connections sync.Map // map[int64]*Connection[TUserData]

	// pendingClientAcks almacena PIDs de mensajes enviados por el servidor que esperan un ClientAck.
	// map[pid string]*types.PendingClientAck
	pendingClientAcks sync.Map

	// pendingServerResponses almacena PIDs de ClientToServerMessage de tipo GenericRequest
	// que esperan una respuesta ServerToClientMessage con el mismo PID.
	// map[pid string]*types.PendingServerResponse
	pendingServerResponses sync.Map

	// ctx es el contexto raíz para el manager, usado para señalar el cierre.
	ctx    context.Context
	cancel context.CancelFunc

	mu sync.RWMutex

	// userConnections es un mapa para almacenar conexiones activas por UserID
	userConnections map[int64][]*Connection[TUserData]
}

// Callbacks devuelve la configuración de callbacks del ConnectionManager.
func (cm *ConnectionManager[TUserData]) Callbacks() Callbacks[TUserData] {
	return cm.callbacks
}

// NewConnectionManager crea una nueva instancia de ConnectionManager.
func NewConnectionManager[TUserData any](cfg types.Config, cbs Callbacks[TUserData]) *ConnectionManager[TUserData] {
	rootCtx, rootCancel := context.WithCancel(context.Background())

	if cbs.GeneratePID == nil {
		cbs.GeneratePID = func() string {
			return uuid.NewString()
		}
	}

	if cbs.AuthenticateAndGetUserData == nil {
		panic("customws: Callbacks.AuthenticateAndGetUserData no puede ser nil")
	}
	if cbs.ProcessClientMessage == nil {
		panic("customws: Callbacks.ProcessClientMessage no puede ser nil")
	}

	manager := &ConnectionManager[TUserData]{
		config:    cfg,
		callbacks: cbs,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				origin := r.Header.Get("Origin")
				if origin == "" {
					// Permitir conexiones sin origen (ej. Postman, scripts locales, otros servidores backend)
					logger.Infof(componentLog, "CheckOrigin: Permitiendo conexión sin origen (Origin header vacío).")
					return true
				}
				if len(cfg.AllowedOrigins) == 0 {
					logger.Warnf(componentLog, "CheckOrigin: No hay AllowedOrigins configurados. Rechazando origen: %s", origin)
					return false // Si no hay orígenes permitidos configurados, denegar todos los que tengan un Origin header.
				}

				// Verificar si "*" está en la lista de orígenes permitidos (wildcard para todos)
				for _, allowedOrigin := range cfg.AllowedOrigins {
					if allowedOrigin == "*" {
						logger.Infof(componentLog, "CheckOrigin: Permitiendo todos los orígenes (wildcard '*' configurado). Origen: %s", origin)
						return true
					}
					if origin == allowedOrigin {
						logger.Infof(componentLog, "CheckOrigin: Origen permitido: %s", origin)
						return true
					}
				}
				logger.Warnf(componentLog, "CheckOrigin: Origen %s no está en la lista de permitidos.", origin)
				return false
			},
		},
		ctx:    rootCtx,
		cancel: rootCancel,
	}

	go manager.cleanupRoutine()

	logger.Infof(componentLog, "ConnectionManager iniciado con UserData tipo: %T", *new(TUserData))
	return manager
}

// ServeHTTP maneja las solicitudes HTTP entrantes y las actualiza a conexiones WebSocket.
func (cm *ConnectionManager[TUserData]) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	userID, userData, err := cm.callbacks.AuthenticateAndGetUserData(r)
	if err != nil {
		logger.Errorf(componentLog, "Error de autenticación en ServeHTTP: %v", err)
		http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
		return
	}

	wsConn, err := cm.upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Errorf(componentLog, "Error al actualizar a WebSocket para UserID %d: %v", userID, err)
		return
	}

	logger.Infof(componentLog, "Conexión WebSocket establecida para UserID %d", userID)

	connCtx, connCancel := context.WithCancel(cm.ctx)

	connection := &Connection[TUserData]{
		ID:       userID,
		conn:     wsConn,
		manager:  cm,
		SendChan: make(chan types.ServerToClientMessage, cm.config.SendChannelBuffer),
		UserData: userData,
		ctx:      connCtx,
		cancel:   connCancel,
	}

	// Usar el mutex para modificar userConnections
	cm.mu.Lock()
	if cm.userConnections == nil {
		cm.userConnections = make(map[int64][]*Connection[TUserData])
	}
	cm.userConnections[userID] = append(cm.userConnections[userID], connection)
	cm.mu.Unlock()

	if oldConn, loaded := cm.connections.LoadAndDelete(userID); loaded {
		logger.Warnf(componentLog, "UserID %d ya tenía una conexión activa. Cerrando la anterior.", userID)
		if oldC, ok := oldConn.(*Connection[TUserData]); ok {
			oldC.Close()
		}
	}
	cm.connections.Store(userID, connection)

	if cm.callbacks.OnConnect != nil {
		if err := cm.callbacks.OnConnect(connection); err != nil {
			logger.Errorf(componentLog, "Error en callback OnConnect para UserID %d: %v. Cerrando conexión.", userID, err)
			connection.Close()
			return
		}
	}

	go connection.readPump()
	go connection.writePump()

	logger.Infof(componentLog, "Pumps de lectura/escritura iniciadas para UserID %d", userID)
}

// Close cierra la conexión WebSocket y cancela su contexto.
func (c *Connection[TUserData]) Close() {
	c.cancel()
	c.conn.Close()
	logger.Infof(componentLog, "Conexión cerrada explícitamente para UserID %d", c.ID)
}

func (c *Connection[TUserData]) readPump() {
	defer func() {
		logger.Infof(componentLog, "readPump: Finalizando para UserID %d", c.ID)
		c.manager.unregisterConnection(c, errors.New("readPump finalizado"))
		c.cancel()
		c.conn.Close()
	}()

	c.conn.SetReadLimit(c.manager.config.MaxMessageSize)
	if err := c.conn.SetReadDeadline(time.Now().Add(c.manager.config.PongWait)); err != nil {
		logger.Errorf(componentLog, "readPump: Error al establecer ReadDeadline inicial para UserID %d: %v", c.ID, err)
		return
	}
	c.conn.SetPongHandler(func(string) error {
		if err := c.conn.SetReadDeadline(time.Now().Add(c.manager.config.PongWait)); err != nil {
			logger.Errorf(componentLog, "readPump: Error al establecer ReadDeadline en PongHandler para UserID %d: %v", c.ID, err)
			return err
		}
		return nil
	})

	for {
		select {
		case <-c.ctx.Done():
			logger.Infof(componentLog, "readPump: Contexto cancelado para UserID %d, terminando.", c.ID)
			return
		default:
			_, messageBytes, err := c.conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure, websocket.CloseNormalClosure) {
					logger.Warnf(componentLog, "readPump: Error de cierre inesperado para UserID %d: %v", c.ID, err)
				} else if e, ok := err.(*websocket.CloseError); ok && (e.Code == websocket.CloseNormalClosure || e.Code == websocket.CloseGoingAway) {
					logger.Infof(componentLog, "readPump: Cierre normal de WebSocket por cliente para UserID %d: %v", c.ID, err)
				} else {
					logger.Errorf(componentLog, "readPump: Error de lectura para UserID %d: %v", c.ID, err)
				}
				return
			}

			var clientMsg types.ClientToServerMessage
			if err := json.Unmarshal(messageBytes, &clientMsg); err != nil {
				logger.Errorf(componentLog, "readPump: Error al deserializar mensaje de UserID %d: %v. Mensaje: %s", c.ID, err, string(messageBytes))
				c.SendErrorNotification(clientMsg.PID, 0, fmt.Sprintf("Error deserializando tu mensaje: %v", err))
				continue
			}

			logger.Infof(componentLog, "readPump: Mensaje recibido de UserID %d, Tipo: %s, PID: %s", c.ID, clientMsg.Type, clientMsg.PID)

			if clientMsg.Type == types.MessageTypeClientAck {
				c.manager.handleClientAck(clientMsg)
				continue
			}

			// Si el mensaje del cliente tiene un PID y este PID está en nuestro mapa de respuestas pendientes,
			// entonces este mensaje es una respuesta a una solicitud que el servidor hizo previamente.
			if clientMsg.PID != "" {
				if pending, loaded := c.manager.pendingServerResponses.Load(clientMsg.PID); loaded {
					if pResp, castOk := pending.(*types.PendingServerResponse); castOk {
						select {
						case pResp.ResponseChan <- clientMsg: // Enviar la respuesta completa del cliente
							logger.Infof(componentLog, "readPump: Respuesta del cliente para PID %s reenviada al solicitante interno.", clientMsg.PID)
							// El solicitante (SendRequestAndWaitClientResponse) es responsable de eliminar de pendingServerResponses.
						default:
							logger.Warnf(componentLog, "readPump: Canal de ResponseChan para PID %s bloqueado o cerrado.", clientMsg.PID)
						}
						continue // Mensaje manejado como respuesta, no pasar a ProcessClientMessage general
					} else {
						logger.Errorf(componentLog, "readPump: Error al castear PendingServerResponse para PID %s.", clientMsg.PID)
						// No continuar, podría ser un mensaje normal que coincida con un PID antiguo.
					}
				}
			}

			// Procesar otros tipos de mensajes a través del callback (si no fue una respuesta manejada arriba)
			if err := c.manager.callbacks.ProcessClientMessage(c, clientMsg); err != nil {
				logger.Errorf(componentLog, "readPump: Error en callback ProcessClientMessage para UserID %d, PID %s: %v", c.ID, clientMsg.PID, err)
				c.SendErrorNotification(clientMsg.PID, 0, fmt.Sprintf("Error procesando tu mensaje: %v", err))
			}
		}
	}
}

func (c *Connection[TUserData]) writePump() {
	pingTicker := time.NewTicker(c.manager.config.PingPeriod)
	defer func() {
		pingTicker.Stop()
		logger.Infof(componentLog, "writePump: Finalizando para UserID %d", c.ID)
		c.conn.Close()
	}()

	for {
		select {
		case <-c.ctx.Done():
			logger.Infof(componentLog, "writePump: Contexto cancelado para UserID %d, enviando mensaje de cierre y terminando.", c.ID)
			_ = c.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseGoingAway, "Servidor cerrando conexión"))
			return

		case message, ok := <-c.SendChan:
			if !ok {
				logger.Infof(componentLog, "writePump: SendChan cerrado para UserID %d, enviando mensaje de cierre.", c.ID)
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.conn.SetWriteDeadline(time.Now().Add(c.manager.config.WriteWait)); err != nil {
				logger.Errorf(componentLog, "writePump: Error al establecer WriteDeadline para UserID %d: %v", c.ID, err)
				continue
			}

			messageBytes, err := json.Marshal(message)
			if err != nil {
				logger.Errorf(componentLog, "writePump: Error al serializar mensaje para UserID %d, PID %s: %v", c.ID, message.PID, err)
				continue
			}

			if err := c.conn.WriteMessage(websocket.TextMessage, messageBytes); err != nil {
				logger.Errorf(componentLog, "writePump: Error de escritura para UserID %d, PID %s: %v", c.ID, message.PID, err)
				return
			}
			logger.Infof(componentLog, "writePump: Mensaje enviado a UserID %d, Tipo: %s, PID: %s", c.ID, message.Type, message.PID)

		case <-pingTicker.C:
			if err := c.conn.SetWriteDeadline(time.Now().Add(c.manager.config.WriteWait)); err != nil {
				logger.Errorf(componentLog, "writePump: Error al establecer WriteDeadline para Ping (UserID %d): %v", c.ID, err)
				continue
			}
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				logger.Errorf(componentLog, "writePump: Error al enviar Ping a UserID %d: %v", c.ID, err)
				return
			}
		}
	}
}

// unregisterConnection es llamado para limpiar una conexión del manager.
func (cm *ConnectionManager[TUserData]) unregisterConnection(conn *Connection[TUserData], disconnectErr error) {
	cm.connections.Delete(conn.ID)
	close(conn.SendChan)

	// Usar el mutex para modificar userConnections
	cm.mu.Lock()
	if conns, exists := cm.userConnections[conn.ID]; exists {
		for i, c := range conns {
			if c == conn {
				// Eliminar esta conexión específica del slice
				cm.userConnections[conn.ID] = append(conns[:i], conns[i+1:]...)
				// Si el slice queda vacío, eliminar la entrada del mapa para limpiar
				if len(cm.userConnections[conn.ID]) == 0 {
					delete(cm.userConnections, conn.ID)
				}
				break // Asumiendo que una conexión solo aparece una vez por usuario
			}
		}
	}
	cm.mu.Unlock()

	logger.Infof(componentLog, "Conexión para UserID %d desregistrada.", conn.ID)

	if cm.callbacks.OnDisconnect != nil {
		cm.callbacks.OnDisconnect(conn, disconnectErr)
	}
}

// SendMessage encola un mensaje para ser enviado a este cliente específico.
func (c *Connection[TUserData]) SendMessage(msg types.ServerToClientMessage) error {
	select {
	case c.SendChan <- msg:
		return nil
	case <-c.ctx.Done():
		logger.Warnf(componentLog, "SendMessage: Intento de enviar a UserID %d pero su contexto está cerrado.", c.ID)
		return fmt.Errorf("conexión para UserID %d cerrada, no se puede enviar mensaje (PID: %s)", c.ID, msg.PID)
	case <-time.After(c.manager.config.WriteWait / 2):
		logger.Errorf(componentLog, "SendMessage: Timeout al intentar enviar a UserID %d (PID: %s). SendChan podría estar lleno o writePump detenida.", c.ID, msg.PID)
		return fmt.Errorf("timeout enviando mensaje a UserID %d (PID: %s)", c.ID, msg.PID)
	}
}

// SendErrorNotification es un helper para enviar un mensaje de error al cliente.
func (c *Connection[TUserData]) SendErrorNotification(originalPID string, code int, message string) {
	errMsg := types.ServerToClientMessage{
		PID:  c.manager.callbacks.GeneratePID(),
		Type: types.MessageTypeErrorNotification,
		Error: &types.ErrorPayload{
			OriginalPID: originalPID,
			Code:        code,
			Message:     message,
		},
	}
	if err := c.SendMessage(errMsg); err != nil {
		logger.Errorf(componentLog, "SendErrorNotification: No se pudo enviar notificación de error a UserID %d para PID original %s: %v", c.ID, originalPID, err)
	}
}

// SendServerAck envía un ack al cliente confirmando la recepción/procesamiento de un mensaje del cliente.
func (c *Connection[TUserData]) SendServerAck(acknowledgedPID string, status string, ackErr error) {
	payload := types.AckPayload{
		AcknowledgedPID: acknowledgedPID,
		Status:          status,
	}
	if ackErr != nil {
		payload.Error = ackErr.Error()
	}
	ackMsg := types.ServerToClientMessage{
		PID:     c.manager.callbacks.GeneratePID(),
		Type:    types.MessageTypeServerAck,
		Payload: payload,
	}
	if err := c.SendMessage(ackMsg); err != nil {
		logger.Errorf(componentLog, "SendServerAck: No se pudo enviar ServerAck a UserID %d para PID %s: %v", c.ID, acknowledgedPID, err)
	}
}

// handleClientAck procesa un ClientAck recibido.
func (cm *ConnectionManager[TUserData]) handleClientAck(ackMsg types.ClientToServerMessage) {
	// Necesitamos decodificar el payload correctamente ya que json.Unmarshal a interface{} crea un map[string]interface{}
	var ackPayload types.AckPayload
	payloadBytes, err := json.Marshal(ackMsg.Payload)
	if err != nil {
		logger.Errorf(componentLog, "handleClientAck: Error al re-serializar AckPayload (PID %s): %v", ackMsg.PID, err)
		return
	}
	if err := json.Unmarshal(payloadBytes, &ackPayload); err != nil {
		logger.Errorf(componentLog, "handleClientAck: Error al decodificar AckPayload para mensaje con PID %s (AckedPID: %s): %v", ackMsg.PID, ackPayload.AcknowledgedPID, err)
		return
	}

	if pending, loaded := cm.pendingClientAcks.Load(ackPayload.AcknowledgedPID); loaded {
		if pAck, castOk := pending.(*types.PendingClientAck); castOk {
			select {
			case pAck.AckChan <- ackMsg:
				logger.Infof(componentLog, "handleClientAck: ClientAck para PID %s (mensaje original %s) reenviado al solicitante.", ackMsg.PID, ackPayload.AcknowledgedPID)
			default:
				logger.Warnf(componentLog, "handleClientAck: Canal de AckChan para PID original %s bloqueado o cerrado.", ackPayload.AcknowledgedPID)
			}
			// El solicitante (que llamó a un método SendWithClientAck) es responsable de eliminar de pendingClientAcks.
		} else {
			logger.Errorf(componentLog, "handleClientAck: Error al castear PendingClientAck para PID original %s.", ackPayload.AcknowledgedPID)
		}
	} else {
		logger.Warnf(componentLog, "handleClientAck: Recibido ClientAck para PID original %s pero no estaba pendiente (quizás timeout o ya procesado).", ackPayload.AcknowledgedPID)
	}
}

// cleanupRoutine periódicamente revisa acks y requests pendientes para timeouts.
func (cm *ConnectionManager[TUserData]) cleanupRoutine() {
	ticker := time.NewTicker(cm.config.AckTimeout / 2)
	defer ticker.Stop()

	for {
		select {
		case <-cm.ctx.Done():
			logger.Infof(componentLog, "cleanupRoutine: Deteniéndose.")
			return
		case <-ticker.C:
			now := time.Now()
			cm.pendingClientAcks.Range(func(key, value interface{}) bool {
				pid := key.(string)
				pAck, ok := value.(*types.PendingClientAck)
				if !ok {
					logger.Errorf(componentLog, "cleanupRoutine: Tipo incorrecto en pendingClientAcks para PID %s", pid)
					cm.pendingClientAcks.Delete(pid) // Eliminar el elemento corrupto
					return true
				}
				if now.Sub(pAck.Timestamp) > cm.config.AckTimeout {
					logger.Warnf(componentLog, "cleanupRoutine: Timeout para ClientAck PID %s. Eliminando y cerrando canal.", pid)
					close(pAck.AckChan)
					cm.pendingClientAcks.Delete(pid)
				}
				return true
			})

			cm.pendingServerResponses.Range(func(key, value interface{}) bool {
				pid := key.(string)
				pResp, ok := value.(*types.PendingServerResponse)
				if !ok {
					logger.Errorf(componentLog, "cleanupRoutine: Tipo incorrecto en pendingServerResponses para PID %s", pid)
					cm.pendingServerResponses.Delete(pid) // Eliminar el elemento corrupto
					return true
				}
				if now.Sub(pResp.Timestamp) > cm.config.RequestTimeout {
					logger.Warnf(componentLog, "cleanupRoutine: Timeout para ServerResponse PID %s. Eliminando y cerrando canal.", pid)
					close(pResp.ResponseChan)
					cm.pendingServerResponses.Delete(pid)
				}
				return true
			})
		}
	}
}

// GetConnection recupera una conexión activa por UserID.
// Devuelve la conexión y un booleano indicando si se encontró.
func (cm *ConnectionManager[TUserData]) GetConnection(userID int64) (*Connection[TUserData], bool) {
	if conn, ok := cm.connections.Load(userID); ok {
		if c, castOk := conn.(*Connection[TUserData]); castOk {
			return c, true
		}
		logger.Errorf(componentLog, "GetConnection: Se encontró un tipo inesperado en el mapa de conexiones para UserID %d", userID)
		cm.connections.Delete(userID) // Eliminar el dato corrupto
		return nil, false
	}
	return nil, false
}

// SendMessageToUser envía un mensaje a un usuario específico si está conectado.
func (cm *ConnectionManager[TUserData]) SendMessageToUser(userID int64, msg types.ServerToClientMessage) error {
	if conn, found := cm.GetConnection(userID); found {
		return conn.SendMessage(msg)
	}
	return fmt.Errorf("usuario %d no conectado o no encontrado", userID)
}

// BroadcastToAll envía un mensaje a todas las conexiones activas.
// Devuelve un mapa de errores, donde la clave es el UserID y el valor es el error ocurrido al enviar a ese usuario.
// Si no hubo errores, el mapa estará vacío.
func (cm *ConnectionManager[TUserData]) BroadcastToAll(msg types.ServerToClientMessage, excludeUserIDs ...int64) map[int64]error {
	errorsMap := make(map[int64]error)
	var wg sync.WaitGroup
	var mu sync.Mutex // Para proteger errorsMap

	excludeSet := make(map[int64]struct{})
	for _, id := range excludeUserIDs {
		excludeSet[id] = struct{}{}
	}

	cm.connections.Range(func(key, value interface{}) bool {
		userID := key.(int64)
		conn := value.(*Connection[TUserData])

		if _, excluded := excludeSet[userID]; excluded {
			return true // Continuar iterando, pero no enviar a este usuario
		}

		wg.Add(1)
		go func(c *Connection[TUserData], m types.ServerToClientMessage) {
			defer wg.Done()
			if err := c.SendMessage(m); err != nil {
				mu.Lock()
				errorsMap[c.ID] = err
				mu.Unlock()
				logger.Errorf(componentLog, "BroadcastToAll: Error enviando a UserID %d: %v", c.ID, err)
			}
		}(conn, msg) // Pasar una copia de msg si se modifica o si la goroutine vive mucho tiempo
		return true // Continuar iterando
	})

	wg.Wait() // Esperar a que todos los envíos (goroutines) terminen
	return errorsMap
}

// BroadcastToUsers envía un mensaje a una lista específica de UserIDs si están conectados.
// Devuelve un mapa de errores, donde la clave es el UserID y el valor es el error ocurrido al enviar a ese usuario.
func (cm *ConnectionManager[TUserData]) BroadcastToUsers(userIDs []int64, msg types.ServerToClientMessage, excludeUserIDs ...int64) map[int64]error {
	errorsMap := make(map[int64]error)
	var wg sync.WaitGroup
	var mu sync.Mutex // Para proteger errorsMap

	excludeSet := make(map[int64]struct{})
	for _, id := range excludeUserIDs {
		excludeSet[id] = struct{}{}
	}

	for _, userID := range userIDs {
		if _, excluded := excludeSet[userID]; excluded {
			continue
		}

		if conn, found := cm.GetConnection(userID); found {
			wg.Add(1)
			go func(c *Connection[TUserData], m types.ServerToClientMessage) {
				defer wg.Done()
				if err := c.SendMessage(m); err != nil {
					mu.Lock()
					errorsMap[c.ID] = err
					mu.Unlock()
					logger.Errorf(componentLog, "BroadcastToUsers: Error enviando a UserID %d: %v", c.ID, err)
				}
			}(conn, msg)
		} else {
			mu.Lock()
			errorsMap[userID] = errors.New("usuario no conectado")
			mu.Unlock()
			logger.Warnf(componentLog, "BroadcastToUsers: UserID %d no encontrado para envío.", userID)
		}
	}

	wg.Wait()
	return errorsMap
}

// SendForClientAck envía un mensaje a la conexión especificada y espera un ClientAck.
// El PID en msgToSend se usará para la correlación. Si está vacío, se generará uno.
// Devuelve el ClientToServerMessage de ack o un error si hay timeout o fallo.
func (cm *ConnectionManager[TUserData]) SendForClientAck(conn *Connection[TUserData], msgToSend types.ServerToClientMessage) (types.ClientToServerMessage, error) {
	if conn == nil {
		return types.ClientToServerMessage{}, errors.New("conexión es nil")
	}

	pidToAck := msgToSend.PID
	if pidToAck == "" {
		pidToAck = cm.callbacks.GeneratePID()
		msgToSend.PID = pidToAck // Asegurar que el mensaje saliente tenga el PID
	}

	ackChannel := make(chan types.ClientToServerMessage, 1) // Buffer de 1 para evitar bloqueo si el ack llega antes de que empecemos a escuchar
	pendingAck := &types.PendingClientAck{
		AckChan:   ackChannel,
		Timestamp: time.Now(),
		MessageID: pidToAck,
	}

	cm.pendingClientAcks.Store(pidToAck, pendingAck)
	defer cm.pendingClientAcks.Delete(pidToAck) // Asegurar limpieza al salir

	logger.Infof(componentLog, "SendForClientAck: Enviando mensaje (PID: %s) a UserID %d, esperando ClientAck.", pidToAck, conn.ID)
	err := conn.SendMessage(msgToSend)
	if err != nil {
		close(ackChannel) // Cerrar para que el select de abajo no se bloquee si SendMessage falla
		return types.ClientToServerMessage{}, fmt.Errorf("error al enviar mensaje inicial (PID: %s) a UserID %d: %w", pidToAck, conn.ID, err)
	}

	select {
	case ack, ok := <-ackChannel:
		if !ok {
			// Canal cerrado por cleanupRoutine debido a timeout
			logger.Warnf(componentLog, "SendForClientAck: Canal de Ack cerrado (probablemente timeout) para PID %s, UserID %d.", pidToAck, conn.ID)
			return types.ClientToServerMessage{}, fmt.Errorf("timeout o error esperando ClientAck para PID %s", pidToAck)
		}
		logger.Infof(componentLog, "SendForClientAck: ClientAck recibido para PID %s de UserID %d.", pidToAck, conn.ID)
		return ack, nil

	case <-time.After(cm.config.AckTimeout):
		logger.Warnf(componentLog, "SendForClientAck: Timeout esperando ClientAck para PID %s de UserID %d.", pidToAck, conn.ID)
		// No necesitamos cerrar ackChannel aquí porque el defer de pendingClientAcks.Delete lo hará la cleanupRoutine eventualmente,
		// o si el ack llega tarde y la cleanupRoutine ya lo eliminó, el select en handleClientAck no encontrará el canal.
		// Sin embargo, es bueno tener un cleanup explícito si la función retorna por timeout.
		// La cleanupRoutine cerrará el canal si detecta el timeout.
		return types.ClientToServerMessage{}, fmt.Errorf("timeout esperando ClientAck para PID %s", pidToAck)

	case <-conn.ctx.Done(): // Si la conexión se cierra mientras esperamos el ack
		logger.Warnf(componentLog, "SendForClientAck: Contexto de conexión para UserID %d cerrado mientras se esperaba Ack para PID %s.", conn.ID, pidToAck)
		return types.ClientToServerMessage{}, fmt.Errorf("conexión cerrada esperando ClientAck para PID %s", pidToAck)
	}
}

// SendRequestAndWaitClientResponse envía un mensaje (solicitud) a la conexión especificada
// y espera una respuesta específica del cliente (no solo un ClientAck).
// El PID en requestMsg se usará para la correlación. Si está vacío, se generará uno.
// Devuelve el ClientToServerMessage de respuesta o un error si hay timeout o fallo.
func (cm *ConnectionManager[TUserData]) SendRequestAndWaitClientResponse(conn *Connection[TUserData], requestMsg types.ServerToClientMessage) (types.ClientToServerMessage, error) {
	if conn == nil {
		return types.ClientToServerMessage{}, errors.New("conexión es nil")
	}

	requestPID := requestMsg.PID
	if requestPID == "" {
		requestPID = cm.callbacks.GeneratePID()
		requestMsg.PID = requestPID // Asegurar que el mensaje saliente tenga el PID
	}

	responseChannel := make(chan types.ClientToServerMessage, 1)
	pendingReq := &types.PendingServerResponse{ // Usamos PendingServerResponse, pero es para una *respuesta del cliente*
		ResponseChan: responseChannel,
		Timestamp:    time.Now(),
	}

	cm.pendingServerResponses.Store(requestPID, pendingReq)
	defer cm.pendingServerResponses.Delete(requestPID) // Asegurar limpieza al salir

	logger.Infof(componentLog, "SendRequestAndWaitClientResponse: Enviando solicitud (PID: %s) a UserID %d, esperando respuesta específica del cliente.", requestPID, conn.ID)
	err := conn.SendMessage(requestMsg)
	if err != nil {
		close(responseChannel)
		return types.ClientToServerMessage{}, fmt.Errorf("error al enviar solicitud inicial (PID: %s) a UserID %d: %w", requestPID, conn.ID, err)
	}

	select {
	case response, ok := <-responseChannel:
		if !ok {
			logger.Warnf(componentLog, "SendRequestAndWaitClientResponse: Canal de respuesta cerrado (probablemente timeout) para PID %s, UserID %d.", requestPID, conn.ID)
			return types.ClientToServerMessage{}, fmt.Errorf("timeout o error esperando respuesta del cliente para PID %s", requestPID)
		}
		logger.Infof(componentLog, "SendRequestAndWaitClientResponse: Respuesta del cliente recibida para PID %s de UserID %d.", requestPID, conn.ID)
		return response, nil

	case <-time.After(cm.config.RequestTimeout): // Usar el RequestTimeout configurado
		logger.Warnf(componentLog, "SendRequestAndWaitClientResponse: Timeout esperando respuesta del cliente para PID %s de UserID %d.", requestPID, conn.ID)
		return types.ClientToServerMessage{}, fmt.Errorf("timeout esperando respuesta del cliente para PID %s", requestPID)

	case <-conn.ctx.Done():
		logger.Warnf(componentLog, "SendRequestAndWaitClientResponse: Contexto de conexión para UserID %d cerrado mientras se esperaba respuesta para PID %s.", conn.ID, requestPID)
		return types.ClientToServerMessage{}, fmt.Errorf("conexión cerrada esperando respuesta del cliente para PID %s", requestPID)
	}
}

// Shutdown cierra ordenadamente el ConnectionManager y todas las conexiones activas.
func (cm *ConnectionManager[TUserData]) Shutdown(ctx context.Context) error {
	logger.Infof(componentLog, "Iniciando shutdown del ConnectionManager...")

	// Señalar a todas las rutinas internas (como cleanupRoutine) que deben detenerse.
	cm.cancel() // Cierra cm.ctx

	// Cerrar todas las conexiones activas.
	// Esto señalará a sus readPump/writePump que deben terminar a través de conn.ctx.Done().
	var wg sync.WaitGroup
	cm.connections.Range(func(key, value interface{}) bool {
		conn, ok := value.(*Connection[TUserData])
		if !ok {
			logger.Errorf(componentLog, "Shutdown: Tipo inesperado en connections map para key %v", key)
			return true
		}
		wg.Add(1)
		go func(c *Connection[TUserData]) {
			defer wg.Done()
			logger.Infof(componentLog, "Shutdown: Cerrando conexión para UserID %d...", c.ID)
			c.Close() // Esto llama a c.cancel() y c.conn.Close()
			// La unregisterConnection se llamará desde el defer de readPump.
		}(conn)
		return true
	})

	// Esperar a que todas las goroutines de cierre de conexión terminen o haya timeout.
	shutdownComplete := make(chan struct{})
	go func() {
		wg.Wait()
		close(shutdownComplete)
	}()

	select {
	case <-shutdownComplete:
		logger.Infof(componentLog, "Todas las conexiones han sido señaladas para cerrar.")
	case <-ctx.Done():
		logger.Errorf(componentLog, "Shutdown: Timeout durante el cierre de conexiones: %v", ctx.Err())
		return ctx.Err()
	}

	// Limpiar mapas de PIDs pendientes (aunque las goroutines que esperan deberían haber terminado por timeout o cierre de canal)
	clearSyncMap := func(m *sync.Map, mapName string) {
		m.Range(func(key, value interface{}) bool {
			logger.Warnf(componentLog, "Shutdown: Limpiando PID pendiente %s de %s", key, mapName)
			m.Delete(key)
			// Intentar cerrar canales si son del tipo esperado, aunque podrían estar ya cerrados.
			switch p := value.(type) {
			case *types.PendingClientAck:
				close(p.AckChan)
			case *types.PendingServerResponse:
				close(p.ResponseChan)
			}
			return true
		})
	}
	clearSyncMap(&cm.pendingClientAcks, "pendingClientAcks")
	clearSyncMap(&cm.pendingServerResponses, "pendingServerResponses")

	logger.Infof(componentLog, "ConnectionManager shutdown completado.")
	return nil
}

// IsUserOnline verifica si un usuario con el UserID dado tiene al menos una conexión activa.
func (cm *ConnectionManager[TUserData]) IsUserOnline(userID int64) bool {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	conns, exists := cm.userConnections[userID]
	return exists && len(conns) > 0
}

// HandlePeerToPeerMessage maneja el envío de mensajes directos entre usuarios.
// Verifica si el destinatario está en línea y envía el mensaje si es posible.
func (cm *ConnectionManager[TUserData]) HandlePeerToPeerMessage(fromConn *Connection[TUserData], toUserID int64, msg types.ServerToClientMessage) error {
	if fromConn == nil {
		return errors.New("conexión de origen es nil")
	}

	// Verificar si el destinatario está en línea
	if !cm.IsUserOnline(toUserID) {
		return fmt.Errorf("usuario %d no está en línea", toUserID)
	}

	// Obtener la conexión del destinatario
	toConn, found := cm.GetConnection(toUserID)
	if !found {
		return fmt.Errorf("no se encontró conexión para usuario %d", toUserID)
	}

	// Enviar el mensaje al destinatario
	err := toConn.SendMessage(msg)
	if err != nil {
		logger.Errorf(componentLog, "Error enviando mensaje de UserID %d a UserID %d: %v", fromConn.ID, toUserID, err)
		return fmt.Errorf("error enviando mensaje: %w", err)
	}

	logger.Infof(componentLog, "Mensaje enviado exitosamente de UserID %d a UserID %d", fromConn.ID, toUserID)
	return nil
}
