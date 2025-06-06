package types

import "time"

// Apiresponse es una estructura de respuesta genérica.
type Apiresponse struct {
	ApiOrigin string `json:"apiOrigin"`
	Status    string `json:"status"`
}

// MessageType define el propósito del mensaje.
type MessageType string

// Constantes para MessageType
const (
	// Tipos de mensajes Cliente -> Servidor
	MessageTypeDataRequest    MessageType = "data_request"    // Una solicitud de datos enviada por un usuario
	MessageTypePresenceUpdate MessageType = "presence_update" // Ej: typing, focus
	MessageTypeClientAck      MessageType = "client_ack"      // Cliente confirma recepción/procesamiento de un mensaje del servidor
	MessageTypeGenericRequest MessageType = "generic_request" // Solicitud genérica del cliente que espera una respuesta con el mismo PID

	// --- Chat --- Client -> Server
	MessageTypeGetChatList        MessageType = "get_chat_list"
	MessageTypeSendChatMessage    MessageType = "send_chat_message"
	MessageTypeMessagesRead       MessageType = "messages_read"        // Cliente notifica que ha leído mensajes en un chat
	MessageTypeTypingIndicatorOn  MessageType = "typing_indicator_on"  // Usuario comenzó a escribir
	MessageTypeTypingIndicatorOff MessageType = "typing_indicator_off" // Usuario dejó de escribir

	// --- Perfil --- Client -> Server
	MessageTypeGetMyProfile    MessageType = "get_my_profile"
	MessageTypeUpdateMyProfile MessageType = "update_my_profile"
	MessageTypeGetUserProfile  MessageType = "get_user_profile"
	// Para añadir/editar/eliminar items del perfil (educación, experiencia, etc.)
	// Se podría usar un tipo genérico o tipos específicos.
	MessageTypeUpdateProfileSection MessageType = "update_profile_section"

	// --- Notificaciones --- Client -> Server
	MessageTypeGetNotifications     MessageType = "get_notifications"
	MessageTypeMarkNotificationRead MessageType = "mark_notification_read"

	// --- Contactos y Búsqueda --- Client -> Server
	MessageTypeSearchUsers           MessageType = "search_users"
	MessageTypeSearchEnterprises     MessageType = "search_enterprises"
	MessageTypeSendContactRequest    MessageType = "send_contact_request"
	MessageTypeRespondContactRequest MessageType = "respond_contact_request"

	// Tipos de mensajes Servidor -> Cliente
	MessageTypeDataEvent         MessageType = "data_event"         // Un nuevo evento de datos para entregar al cliente
	MessageTypePresenceEvent     MessageType = "presence_event"     // Notificación de cambio de presencia de otro usuario
	MessageTypeServerAck         MessageType = "server_ack"         // Servidor confirma recepción/procesamiento de un mensaje del cliente
	MessageTypeGenericResponse   MessageType = "generic_response"   // Respuesta del servidor a una GenericRequest
	MessageTypeErrorNotification MessageType = "error_notification" // Notificación de error (ej. fallo al procesar un mensaje previo)

	// --- Chat --- Server -> Client
	MessageTypeChatList             MessageType = "chat_list"
	MessageTypeNewChatMessage       MessageType = "new_chat_message"
	MessageTypeChatHistory          MessageType = "get_history"            // Nuevo: Para enviar el historial de mensajes de un chat
	MessageTypeMessageStatusUpdated MessageType = "message_status_updated" // Ej: delivered_to_recipient, read_by_recipient
	MessageTypeTypingEvent          MessageType = "typing_event"           // Evento de "está escribiendo"

	// --- Perfil --- Server -> Client
	MessageTypeMyProfileData         MessageType = "my_profile_data"
	MessageTypeUserProfileData       MessageType = "user_profile_data"
	MessageTypeProfileUpdateResult   MessageType = "profile_update_result"
	MessageTypeProfileSectionUpdated MessageType = "profile_section_updated"

	// --- Notificaciones --- Server -> Client
	MessageTypeNotificationList MessageType = "notification_list"
	MessageTypeNewNotification  MessageType = "new_notification"
	MessageTypeNotificationRead MessageType = "notification_read"

	// --- Contactos y Búsqueda --- Server -> Client
	MessageTypeSearchResultsUsers       MessageType = "search_results_users"
	MessageTypeSearchResultsEnterprises MessageType = "search_results_enterprises"
	MessageTypeContactRequestReceived   MessageType = "contact_request_received"
	MessageTypeContactRequestResponded  MessageType = "contact_request_responded"
	MessageTypeContactStatusChanged     MessageType = "contact_status_changed" // Ej: amigo añadido, eliminado

	// --- Mensajes del Cliente al Servidor ---
	MessageTypeAcceptFriendRequest MessageType = "accept_request"
	MessageTypeRejectFriendRequest MessageType = "reject_request"

	// --- Mensajes del Servidor al Cliente ---
	MessageTypeNewMessage  MessageType = "new_message"
	MessageTypeProfileData MessageType = "profile_data"

	// Tipos de mensaje para el CV
	MessageTypeSetSkill          = "set_skill"
	MessageTypeSetLanguage       = "set_language"
	MessageTypeSetWorkExperience = "set_work_experience"
	MessageTypeSetCertification  = "set_certification"
	MessageTypeSetProject        = "set_project"
	MessageTypeGetCV             = "get_cv"
)

// ClientToServerMessage es la estructura para mensajes enviados por el cliente al servidor.
type ClientToServerMessage struct {
	PID          string      `json:"pid,omitempty"`          // ID de Proceso/Petición, opcional para el cliente, pero útil para rastrear o si el cliente espera un ServerAck específico.
	Type         MessageType `json:"type"`                   // Tipo de mensaje para enrutamiento en el servidor.
	TargetUserID int64       `json:"targetUserId,omitempty"` // Para mensajes directos (ej. en comunicación peer-to-peer).
	Payload      interface{} `json:"payload,omitempty"`      // Contenido del mensaje, puede ser cualquier struct JSON.
}

// ServerToClientMessage es la estructura para mensajes enviados por el servidor al cliente.
type ServerToClientMessage struct {
	PID        string        `json:"pid,omitempty"` // ID de Proceso/Petición, para que el cliente pueda correlacionar respuestas o confirmar con un ClientAck.
	Type       MessageType   `json:"type"`
	FromUserID int64         `json:"fromUserId,omitempty"` // Quién originó el mensaje (ej. en comunicación peer-to-peer).
	Payload    interface{}   `json:"payload,omitempty"`
	Error      *ErrorPayload `json:"error,omitempty"` // Para reportar errores específicos de la operación.
}

// ErrorPayload define la estructura para errores.
type ErrorPayload struct {
	OriginalPID string `json:"originalPid,omitempty"` // PID del mensaje que causó el error, si aplica.
	Code        int    `json:"code"`                  // Código de error interno o HTTP status-like.
	Message     string `json:"message"`               // Mensaje de error legible.
}

// AckPayload es un payload común para mensajes de tipo ack (tanto ClientAck como ServerAck).
type AckPayload struct {
	AcknowledgedPID string `json:"acknowledgedPid"` // PID del mensaje que se está confirmando.
	Status          string `json:"status"`
	Error           string `json:"error"` // Ej: "received", "processed", "read".
}

// PresenceUpdatePayload es un ejemplo para MessageTypePresenceUpdate.
type PresenceUpdatePayload struct {
	Status       string `json:"status"`                 // "typing", "idle", "online", "offline" (aunque online/offline se maneja por conexión)
	TargetUserID int64  `json:"targetUserId,omitempty"` // Si es "typing" en una comunicación específica.
}

// DataRequestPayload es un ejemplo para MessageTypeDataRequest.
type DataRequestPayload struct {
	Data map[string]interface{} `json:"data"` // Datos genéricos para la solicitud
	// Cualquier otro metadato del mensaje, como ID de mensaje temporal del cliente.
}

// Configuration para el ConnectionManager.
type Config struct {
	WriteWait         time.Duration // Tiempo máximo para una escritura al peer.
	PongWait          time.Duration // Tiempo máximo para leer el siguiente pong del peer.
	PingPeriod        time.Duration // Frecuencia de envío de pings al peer. (Debe ser menor que PongWait)
	MaxMessageSize    int64         // Tamaño máximo permitido para un mensaje leído del peer.
	SendChannelBuffer int           // Tamaño del buffer para el canal de envío de cada conexión.
	AckTimeout        time.Duration // Timeout para esperar una confirmación (ack) de un mensaje enviado con SendWithAck.
	RequestTimeout    time.Duration // Timeout genérico para solicitudes que esperan una respuesta.
	AllowedOrigins    []string      // Lista de orígenes permitidos. Si es nil o vacía, se denegarán todos los orígenes no locales por defecto.
}

// DefaultConfig retorna una configuración por defecto.
func DefaultConfig() Config {
	return Config{
		WriteWait:         10 * time.Second,
		PongWait:          60 * time.Second,
		PingPeriod:        (60 * time.Second * 9) / 10, // Debe ser menor que PongWait
		MaxMessageSize:    2048,                        // Aumentado a 2KB, ajustar según necesidad
		SendChannelBuffer: 512,                         // Buffer más grande para el canal de envío
		AckTimeout:        5 * time.Second,
		RequestTimeout:    10 * time.Second,
		AllowedOrigins:    nil, // Por defecto, nil. El CheckOrigin lo interpretará.
	}
}

// UserData es una interfaz vacía para permitir al usuario de la biblioteca definir
// qué datos específicos del usuario quiere asociar con cada conexión.
// Por ejemplo, podría ser un struct con UserID, Nombre, Roles, etc.
// type UserData interface{} // No es necesario definirlo como interfaz aquí, se usará un tipo genérico T en ConnectionManager.

// Estructura para las operaciones pendientes de ACK por parte del servidor.
// Se usa para cuando el servidor envía un mensaje y espera una confirmación del cliente.
type PendingClientAck struct {
	AckChan   chan ClientToServerMessage // Canal para recibir el ClientAck.
	Timestamp time.Time                  // Para gestionar timeouts.
	MessageID string                     // PID del mensaje original enviado por el servidor.
}

// Estructura para solicitudes genéricas del cliente que esperan una respuesta del servidor,
// O para solicitudes del servidor que esperan una respuesta específica del cliente.
type PendingServerResponse struct {
	ResponseChan chan ClientToServerMessage // Corregido: Debe ser ClientToServerMessage si esperamos respuesta del cliente.
	Timestamp    time.Time                  // Para gestionar timeouts.
}
