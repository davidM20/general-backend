# Plan de Migración a Arquitectura Escalable CustomWS

## Resumen Ejecutivo

Este documento describe la migración completa del sistema WebSocket actual al nuevo paquete `customws` escalable, diseñado para manejar hasta 1 millón de conexiones concurrentes. La migración involucra la refactorización de 14 archivos principales del websocket, eliminando el sistema Hub legacy y adoptando una arquitectura orientada a callbacks con gestión eficiente de recursos.

## Estado Actual de la Migración

### ✅ Archivos Completados
- `main.go` - Implementado completamente
- `manager_setup.go` - Implementado con callbacks de autenticación
- `app_context.go` - Implementado para dependencias compartidas
- `helpers.go` - Implementado con funciones utilitarias migradas
- `message_handlers.go` - Implementado parcialmente (handlers de chat migrados)

### 📋 Archivos Pendientes de Migración
- `hub.go` - **ELIMINAR** (reemplazado por customws)
- `client.go` - **ELIMINAR** (reemplazado por customws.Connection)
- `router.go` - **REFACTORIZAR** completamente
- `handlers_chat.go` - **MIGRAR** handlers restantes
- `handlers_contact.go` - **MIGRAR** completamente
- `handlers_curriculum.go` - **MIGRAR** completamente
- `handlers_list.go` - **MIGRAR** completamente
- `handlers_notification.go` - **MIGRAR** completamente
- `handlers_profile.go` - **MIGRAR** completamente
- `handlers_search.go` - **MIGRAR** completamente
- `hub_helpers.go` - **MIGRAR** funciones útiles
- `types.go` - **ACTUALIZAR** para usar customws.types

## Análisis Detallado Línea por Línea

### 1. `types.go` (Prioridad: CRÍTICA) - 318 líneas

**Estado:** Refactorización crítica para compatibilidad con customws

#### Sección 1: Imports (Líneas 1-6)
```go
// MANTENER - Sin cambios
package websocket

import (
	"time"
	"github.com/davidM20/micro-service-backend-go.git/internal/models"
)
```

#### Sección 2: Constantes de Tipos de Mensaje (Líneas 8-32)
**MANTENER** - Estas constantes siguen siendo válidas para identificar tipos en el nuevo sistema:

```go
// Líneas 8-32: MANTENER sin cambios
const (
	MessageTypeMessage                      = "message"         // ✅ Compatible
	MessageTypeList                         = "list"            // ✅ Compatible  
	MessageTypeReading                      = "reading"         // ✅ Compatible
	MessageTypeWriting                      = "writing"         // ✅ Compatible
	MessageTypeRecording                    = "recording"       // ✅ Compatible
	MessageTypeGetNotifications             = "get-notifications" // ✅ Compatible
	MessageTypeProfile                      = "profile"         // ✅ Compatible
	MessageTypeEditProfile                  = "edit-profile"    // ✅ Compatible
	MessageTypeDeleteItemCurriculum         = "delete-item-curriculum" // ✅ Compatible
	MessageTypeAddContact                   = "add-contact"     // ✅ Compatible
	MessageTypeDeletedContact               = "deleted-contact" // ✅ Compatible
	MessageTypeSearch                       = "search"          // ✅ Compatible
	MessageTypeReadMessages                 = "read_messages"   // ✅ Compatible
	MessageTypeGetMyProfile                 = "get_my_profile"  // ✅ Compatible
	// ... resto de constantes MANTENER
)
```

#### Sección 3: Estructuras Legacy (Líneas 36-45) - **ELIMINAR COMPLETAMENTE**
```go
// Líneas 36-40: ELIMINAR - Reemplazado por customws_types.ClientToServerMessage  
type IncomingMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

// Líneas 42-45: ELIMINAR - Reemplazado por customws_types.ServerToClientMessage
type OutgoingMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
	Error   string      `json:"error,omitempty"`
}
```

**⚠️ CAMBIO CRÍTICO:** Todas las referencias a `IncomingMessage` y `OutgoingMessage` en handlers deben cambiarse a:
- `IncomingMessage` → `customws_types.ClientToServerMessage`
- `OutgoingMessage` → `customws_types.ServerToClientMessage`

#### Sección 4: Payloads Específicos (Líneas 47-318) - **MANTENER CON MODIFICACIONES**

**Líneas 49-53: ErrorPayload - MANTENER**
```go
// ✅ MANTENER - Compatible con customws ErrorNotification
type ErrorPayload struct {
	Error string `json:"error"`
}
```

**Líneas 55-58: SuccessPayload - MANTENER**  
```go
// ✅ MANTENER - Útil para respuestas exitosas
type SuccessPayload struct {
	Message string `json:"message"`
}
```

**Líneas 60-70: ChatMessagePayload - MANTENER**
```go
// ✅ MANTENER - Core para funcionalidad de chat
type ChatMessagePayload struct {
	ChatID     string `json:"chatId"`
	Text       string `json:"text,omitempty"`
	MediaID    string `json:"mediaId,omitempty"`
	ResponseTo string `json:"responseTo,omitempty"`
	MessageID  string `json:"messageId,omitempty"`
	UserID     int64  `json:"userId,omitempty"`
	Timestamp  int64  `json:"timestamp,omitempty"`
}
```

**Líneas 72-76: ListRequestPayload - MANTENER**
```go
// ✅ MANTENER - Para requests de listas
type ListRequestPayload struct {
	ListType string `json:"listType"`
}
```

**Líneas 78-82: ChatStatusPayload - MANTENER**  
```go
// ✅ MANTENER - Para estados de chat (typing, reading, etc.)
type ChatStatusPayload struct {
	ChatID string `json:"chatId"`
	UserID int64  `json:"userId"`
}
```

**Líneas 84-87: ProfileRequestPayload - MANTENER**
```go
// ✅ MANTENER - Para solicitudes de perfil
type ProfileRequestPayload struct {
	TargetUserID int64 `json:"targetUserId,omitempty"`
}
```

**Líneas 89-105: EditProfilePayload - MANTENER**
```go
// ✅ MANTENER - Para edición de perfiles
type EditProfilePayload struct {
	Summary  *string `json:"summary,omitempty"`
	Address  *string `json:"address,omitempty"`
	Github   *string `json:"github,omitempty"`
	Linkedin *string `json:"linkedin,omitempty"`
	Picture  *string `json:"picture,omitempty"`
	Education      *[]models.Education      `json:"education,omitempty"`
	WorkExperience *[]models.WorkExperience `json:"workExperience,omitempty"`
	Certifications *[]models.Certifications `json:"certifications,omitempty"`
	Skills         *[]models.Skills         `json:"skills,omitempty"`
	Languages      *[]models.Languages      `json:"languages,omitempty"`
	Projects       *[]models.Project        `json:"projects,omitempty"`
}
```

**Líneas 107-111: DeleteItemPayload - MANTENER**
```go
// ✅ MANTENER - Para eliminación de items del currículum  
type DeleteItemPayload struct {
	ItemType string `json:"itemType"`
	ItemID   int64  `json:"itemId"`
}
```

**Líneas 113-116: AddContactPayload - MANTENER**
```go
// ✅ MANTENER - Para solicitudes de contacto
type AddContactPayload struct {
	TargetUserID int64 `json:"targetUserId"`
}
```

**Líneas 118-121: DeleteContactPayload - MANTENER**
```go
// ✅ MANTENER - Para eliminación de contactos
type DeleteContactPayload struct {
	TargetUserID int64 `json:"targetUserId"`
}
```

**Líneas 123-130: SearchPayload - MANTENER**
```go
// ✅ MANTENER - Para búsquedas
type SearchPayload struct {
	Query      string `json:"query,omitempty"`
	EntityType string `json:"entityType,omitempty"`
	Limit      int    `json:"limit,omitempty"`
	Offset     int    `json:"offset,omitempty"`
}
```

**Líneas 132-137: SearchResponsePayload - MANTENER**
```go
// ✅ MANTENER - Para respuestas de búsqueda
type SearchResponsePayload struct {
	Query      string        `json:"query"`
	EntityType string        `json:"entityType"`
	Results    []interface{} `json:"results"`
}
```

**Líneas 139-143: ReadMessagesPayload - MANTENER**
```go
// ✅ MANTENER - Para marcar mensajes como leídos
type ReadMessagesPayload struct {
	ChatID            string `json:"chatId"`
	LastReadTimestamp int64  `json:"lastReadTimestamp,omitempty"`
}
```

**Líneas 145-151: MessagesReadPayload - MANTENER**
```go
// ✅ MANTENER - Para notificaciones de lectura
type MessagesReadPayload struct {
	ChatID   string `json:"chatId"`
	ReaderID int64  `json:"readerId"`
}
```

**Líneas 153-159: OnlineStatusPayload - MANTENER**
```go
// ✅ MANTENER - Para notificaciones de estado online
type OnlineStatusPayload struct {
	UserID   int64  `json:"userId"`
	IsOnline bool   `json:"isOnline"`
	UserName string `json:"userName,omitempty"`
	ChatID   string `json:"chatId,omitempty"`
}
```

**Líneas 161-318: Response Structs - MANTENER TODAS**
```go
// ✅ MANTENER - Todas las estructuras de respuesta son compatibles:
// - EducationResponse (161-168)
// - WorkExperienceResponse (170-178) 
// - CertificationsResponse (180-185)
// - ProjectResponse (187-196)
// - SkillsResponse (213)
// - LanguagesResponse (214)
// - Curriculum (217-226)
// - UserSearchResult (228-242)
// - EnterpriseSearchResult (244-255)
// - MyProfileResponse (257-284)
// - ContactInfo (286-294)
// - ChatInfo (296-307)
// - OnlineUserInfo (309-312)
// - ListResponsePayload (314-318)
```

#### Cambios Requeridos en types.go:

1. **AGREGAR imports para customws:**
```go
// Línea 5: AGREGAR después de imports existentes
import (
	"time"
	"github.com/davidM20/micro-service-backend-go.git/internal/models"
	customws_types "github.com/davidM20/micro-service-backend-go.git/pkg/customws/types" // NUEVO
)
```

2. **ELIMINAR líneas 36-45 (IncomingMessage y OutgoingMessage)**

3. **AGREGAR nueva estructura AppUserData:**
```go
// AGREGAR después de las constantes (línea 33)
// AppUserData define los datos de usuario específicos almacenados en cada conexión
type AppUserData struct {
	ID       int64  `json:"id"`
	UserName string `json:"userName"`
	Email    string `json:"email"`
	RoleID   int    `json:"roleId"`
}
```

### 2. `router.go` (Prioridad: CRÍTICA) - 56 líneas

**Estado:** Refactorización completa requerida

#### Análisis línea por línea:

**Líneas 1-7: Package e imports**
```go
// CAMBIAR imports
package websocket

import (
	"encoding/json" // ELIMINAR - no necesario en nueva arquitectura
	"github.com/davidM20/micro-service-backend-go.git/pkg/logger" // MANTENER
	"fmt" // MANTENER
	customws_types "github.com/davidM20/micro-service-backend-go.git/pkg/customws/types" // AGREGAR
)
```

**Líneas 9-56: Función ProcessWebSocketMessage - REEMPLAZAR COMPLETAMENTE**
```go
// ELIMINAR líneas 9-56 completamente y REEMPLAZAR con:

// RouteClientMessage procesa mensajes desde el cliente usando la nueva arquitectura customws
func RouteClientMessage(appCtx *AppContext, conn *customws.Connection[AppUserData], msg customws_types.ClientToServerMessage) error {
	logger.Infof("WS", "[ROUTER] Info (UserID: %d): Routing message type '%s' with PID '%s'", 
		conn.UserData.ID, msg.Type, msg.PID)

	// Rutear basado en el tipo de mensaje  
	switch msg.Type {
	case MessageTypeMessage:
		return handleChatMessage(appCtx, conn, msg)
	case MessageTypeReading, MessageTypeWriting, MessageTypeRecording:
		return handleChatStatusUpdate(appCtx, conn, msg.Type, msg)
	case MessageTypeReadMessages:
		return handleReadMessages(appCtx, conn, msg)
	case MessageTypeList:
		return handleListRequest(appCtx, conn, msg)
	case MessageTypeSearch:
		return handleSearch(appCtx, conn, msg)
	case MessageTypeAddContact:
		return handleAddContact(appCtx, conn, msg)
	case MessageTypeDeletedContact:
		return handleDeleteContact(appCtx, conn, msg)
	case MessageTypeGetMyProfile:
		return handleGetMyProfile(appCtx, conn, msg)
	case MessageTypeProfile:
		return handleGetProfile(appCtx, conn, msg)
	case MessageTypeEditProfile:
		return handleEditProfile(appCtx, conn, msg)
	case MessageTypeDeleteItemCurriculum:
		return handleDeleteItemCurriculum(appCtx, conn, msg)
	case MessageTypeGetNotifications:
		return handleGetNotifications(appCtx, conn, msg)
	default:
		logger.Warnf("WS", "RouteClientMessage Warn (UserID: %d): Unknown message type '%s'", 
			conn.UserData.ID, msg.Type)
		conn.SendErrorNotification(msg.PID, 400, fmt.Sprintf("Unknown message type: %s", msg.Type))
		return fmt.Errorf("unknown message type: %s", msg.Type)
	}
}
```

**Cambios específicos requeridos:**
- **Línea 9:** `func ProcessWebSocketMessage(hub *Hub, client *Client, messageBytes []byte) error` → `func RouteClientMessage(appCtx *AppContext, conn *customws.Connection[AppUserData], msg customws_types.ClientToServerMessage) error`
- **Líneas 10-14:** Eliminar parsing JSON (ya manejado por customws)
- **Línea 16:** `client.User.ID` → `conn.UserData.ID`
- **Línea 18:** `msg.Type` sigue igual, pero viene de `customws_types.ClientToServerMessage`
- **Líneas 20-50:** Actualizar todas las llamadas de handlers para usar nueva firma
- **Línea 52:** `hub.sendErrorMessage(client, ...)` → `conn.SendErrorNotification(msg.PID, 400, ...)`

¿Es correcto este análisis detallado de `types.go` y `router.go`? ¿Continúo con el siguiente archivo handler?

## Análisis de Dependencias por Archivo

### 1. `router.go` (Prioridad: CRÍTICA)

**Estado:** Refactorización completa requerida
**Motivo:** Es el punto de entrada de todos los mensajes WebSocket

#### Cambios Necesarios:

1. **Eliminación del sistema de routing actual:**
   - Remover `handleIncomingMessage(client *Client, message []byte)`
   - Remover `processMessage(client *Client, msgType string, payload interface{})`

2. **Implementación del nuevo sistema:**
   ```go
   // Nueva función que será llamada desde manager_setup.go
   func RouteClientMessage(appCtx *AppContext, conn *customws.Connection[AppUserData], msg customws_types.ClientToServerMessage) error
   ```

3. **Migración del switch statement:**
   - Cambiar de `msgType string` a `msg.Type string`
   - Actualizar llamadas a handlers para usar nuevas firmas
   - Implementar manejo de errores consistente con customws

4. **Tipos de mensaje a migrar:**
   ```go
   case "message":          -> handleChatMessage(appCtx, conn, msg)
   case "reading":          -> handleChatStatusUpdate(appCtx, conn, "reading", msg)
   case "writing":          -> handleChatStatusUpdate(appCtx, conn, "writing", msg)
   case "recording":        -> handleChatStatusUpdate(appCtx, conn, "recording", msg)
   case "read_messages":    -> handleReadMessages(appCtx, conn, msg)
   case "contact_request":  -> handleContactRequest(appCtx, conn, msg)
   case "contact_response": -> handleContactResponse(appCtx, conn, msg)
   case "get_contacts":     -> handleGetContacts(appCtx, conn, msg)
   case "get_user_chats":   -> handleGetUserChats(appCtx, conn, msg)
   case "get_chat_messages": -> handleGetChatMessages(appCtx, conn, msg)
   case "search_users":     -> handleSearchUsers(appCtx, conn, msg)
   case "get_profile":      -> handleGetProfile(appCtx, conn, msg)
   case "update_profile":   -> handleUpdateProfile(appCtx, conn, msg)
   case "get_notifications": -> handleGetNotifications(appCtx, conn, msg)
   case "mark_notification_read": -> handleMarkNotificationRead(appCtx, conn, msg)
   case "get_curriculum":   -> handleGetCurriculum(appCtx, conn, msg)
   case "update_curriculum": -> handleUpdateCurriculum(appCtx, conn, msg)
   ```

### 2. `handlers_chat.go` (Prioridad: ALTA)

**Estado:** ⚠️ **DUPLICACIÓN CRÍTICA** - Todas las funciones ya están migradas en otros archivos

#### Análisis línea por línea:

**Líneas 1-11: Package e imports**
```go
// CAMBIAR imports
package websocket

import (
	"database/sql"     // MANTENER
	"encoding/json"    // ELIMINAR - no necesario con customws
	"fmt"              // MANTENER  
	"time"             // MANTENER
	
	"github.com/davidM20/micro-service-backend-go.git/pkg/logger" // MANTENER
	"github.com/google/uuid"  // MANTENER
	customws_types "github.com/davidM20/micro-service-backend-go.git/pkg/customws/types" // AGREGAR
)
```

#### Función 1: handleChatMessage (Líneas 13-95) - **ELIMINAR COMPLETAMENTE**

**Estado:** ⚠️ **DUPLICADA** - Esta función ya existe en `message_handlers.go`

**Acción requerida:** 
- **ELIMINAR** esta versión completamente (líneas 13-95)
- **MANTENER** la versión ya migrada en `message_handlers.go`

**Análisis de duplicación detectada:**
```go
// Línea 13: FUNCIÓN DUPLICADA
func (h *Hub) handleChatMessage(client *Client, payload interface{}) error

// ⚠️ ESTA FUNCIÓN YA EXISTE EN message_handlers.go COMO:
// func handleChatMessage(appCtx *AppContext, conn *customws.Connection[AppUserData], msg customws_types.ClientToServerMessage) error

// Líneas 15-25: Parsing JSON manual - YA NO NECESARIO en nueva arquitectura
var chatPayload ChatMessagePayload
payloadBytes, err := json.Marshal(payload)
if err := json.Unmarshal(payloadBytes, &chatPayload); err != nil

// Línea 33: Validación de chat - YA MIGRADA a helpers.go
otherUserID, err := h.validateChatParticipant(chatPayload.ChatID, client.User.ID)

// Línea 52: Inserción en DB - PATRÓN YA ACTUALIZADO en message_handlers.go
_, err = h.DB.Exec(`INSERT INTO Message...`)

// Línea 82: Envío a usuario - PATRÓN YA ACTUALIZADO
if err := h.SendToUser(otherUserID, jsonMsg); err != nil

// Línea 90: Envío a remitente - PATRÓN YA ACTUALIZADO  
client.SendToMe(jsonMsg)
```

#### Función 2: validateChatParticipant (Líneas 97-115) - **ELIMINAR COMPLETAMENTE**

**Estado:** ⚠️ **DUPLICADA** - Esta función ya fue migrada a `helpers.go`

**Acción requerida:**
- **ELIMINAR** esta versión completamente (líneas 97-115)  
- **USAR** la versión ya migrada en `helpers.go`

**Análisis de duplicación:**
```go
// Línea 97: FUNCIÓN DUPLICADA
func (h *Hub) validateChatParticipant(chatID string, userID int64) (int64, error)

// ⚠️ ESTA FUNCIÓN YA EXISTE EN helpers.go COMO:
// func validateChatParticipant(appCtx *AppContext, chatID string, userID int64) (int64, error)

// Líneas 99-115: Lógica de validación - YA MIGRADA correctamente
var user1ID, user2ID int64
query := "SELECT User1Id, User2Id FROM Contact WHERE ChatId = ? AND Status = 'accepted'"
err := h.DB.QueryRow(query, chatID).Scan(&user1ID, &user2ID)
// ... resto de lógica YA MIGRADA
```

#### Función 3: handleChatStatusUpdate (Líneas 117-164) - **ELIMINAR COMPLETAMENTE**

**Estado:** ⚠️ **DUPLICADA** - Esta función ya existe en `message_handlers.go`

**Acción requerida:**
- **ELIMINAR** esta versión completamente (líneas 117-164)
- **MANTENER** la versión ya migrada en `message_handlers.go`

**Análisis de duplicación:**
```go
// Línea 117: FUNCIÓN DUPLICADA
func (h *Hub) handleChatStatusUpdate(client *Client, msgType string, payload interface{}) error

// ⚠️ ESTA FUNCIÓN YA EXISTE EN message_handlers.go COMO:
// func handleChatStatusUpdate(appCtx *AppContext, conn *customws.Connection[AppUserData], msgType string, msg customws_types.ClientToServerMessage) error

// Líneas 120-131: Parsing payload - YA NO NECESARIO
var req ChatStatusPayload
payloadBytes, err := json.Marshal(payload)
if err := json.Unmarshal(payloadBytes, &req); err != nil

// Línea 136: Validación de chat - YA MIGRADA
otherUserID, err := h.validateChatParticipant(req.ChatID, client.User.ID)

// Línea 157: Envío a otro participante - PATRÓN YA ACTUALIZADO
if err := h.SendToUser(otherUserID, jsonMsg); err != nil
```

#### Función 4: handleReadMessagesList (Líneas 166-265) - **ELIMINAR COMPLETAMENTE**

**Estado:** ⚠️ **DUPLICADA** - Esta función ya existe como `handleReadMessages` en `message_handlers.go`

**Acción requerida:**
- **ELIMINAR** esta versión completamente (líneas 166-265)
- **MANTENER** la versión ya migrada en `message_handlers.go`

**Análisis de duplicación:**
```go
// Línea 166: FUNCIÓN DUPLICADA  
func (h *Hub) handleReadMessagesList(client *Client, payload interface{}) error

// ⚠️ ESTA FUNCIÓN YA EXISTE EN message_handlers.go COMO:
// func handleReadMessages(appCtx *AppContext, conn *customws.Connection[AppUserData], msg customws_types.ClientToServerMessage) error

// Líneas 170-181: Parsing payload - YA NO NECESARIO
var req ReadMessagesPayload
payloadBytes, err := json.Marshal(payload)
if err := json.Unmarshal(payloadBytes, &req); err != nil

// Línea 187: Validación de chat - YA MIGRADA
otherUserID, err := h.validateChatParticipant(req.ChatID, userID)

// Línea 194: Query de actualización - PATRÓN YA ACTUALIZADO
query := `UPDATE Message SET StatusMessage = 3 WHERE ChatId = ? AND UserId = ? AND StatusMessage < 3`
result, err := h.DB.Exec(query, args...)

// Líneas 216-235: Notificación de lectura - PATRÓN YA ACTUALIZADO
outgoingNotify := OutgoingMessage{
	Type:    MessageTypeMessagesRead,
	Payload: notificationPayload,
}
jsonNotify, err := json.Marshal(outgoingNotify)
if err := h.SendToUser(otherUserID, jsonNotify); err != nil
```

### Resolución Final para handlers_chat.go:

**Decisión:** **ELIMINAR ARCHIVO COMPLETO**

**Justificación:** Todas las funciones (4/4) están 100% duplicadas y ya migradas correctamente:

| Función Original | Estado | Ubicación Nueva | 
|------------------|---------|-----------------|
| `handleChatMessage` | ✅ Migrada | `message_handlers.go` |
| `validateChatParticipant` | ✅ Migrada | `helpers.go` |  
| `handleChatStatusUpdate` | ✅ Migrada | `message_handlers.go` |
| `handleReadMessagesList` | ✅ Migrada como `handleReadMessages` | `message_handlers.go` |

**Plan de acción específico:**
1. **VERIFICAR** que las versiones en `message_handlers.go` y `helpers.go` implementen toda la lógica
2. **ELIMINAR** archivo `handlers_chat.go` completamente
3. **ACTUALIZAR** imports en archivos que referencien estas funciones:
   ```go
   // CAMBIAR referencias de:
   // "handlers_chat.handleChatMessage" 
   // ↓ A:
   // "message_handlers.handleChatMessage"
   ```
4. **PROBAR** funcionalidad completa de chat después de eliminación

**Riesgos mitigados:**
- ✅ No hay pérdida de funcionalidad (todo ya migrado)
- ✅ No hay dependencias externas no resueltas
- ✅ Mejora la consistencia del código (elimina duplicación)
- ✅ Reduce la superficie de mantenimiento

### 4. `handlers_contact.go` (Prioridad: ALTA) - 216 líneas

**Estado:** Migración completa requerida - Contiene 5 funciones para gestión de contactos

#### Análisis línea por línea:

**Líneas 1-11: Package e imports**
```go
// CAMBIAR imports
package websocket

import (
	"database/sql"      // MANTENER
	"encoding/json"     // ELIMINAR - no necesario con customws
	"fmt"               // MANTENER
	"log"               // CAMBIAR - usar logger consistente

	"github.com/davidM20/micro-service-backend-go.git/internal/models" // MANTENER
	"github.com/google/uuid" // MANTENER
	"github.com/davidM20/micro-service-backend-go.git/pkg/logger"      // AGREGAR
	customws_types "github.com/davidM20/micro-service-backend-go.git/pkg/customws/types" // AGREGAR
	"time" // AGREGAR
)
```

#### Función 1: handleAddContact (Líneas 13-85) - **MIGRAR COMPLETAMENTE**

**Acción requerida:** Migrar función principal para agregar contactos

**Cambios específicos línea por línea:**

```go
// Línea 13: CAMBIAR firma
func (h *Hub) handleAddContact(client *Client, payload interface{}) error
// ↓ CAMBIAR A:
func handleAddContact(appCtx *AppContext, conn *customws.Connection[AppUserData], msg customws_types.ClientToServerMessage) error

// Línea 14: CAMBIAR referencia usuario
userID := client.User.ID
// ↓ CAMBIAR A:
userID := conn.UserData.ID

// Línea 15: CAMBIAR logging
log.Printf("handleAddContact Info (UserID: %d): Received request", userID)
// ↓ CAMBIAR A:
logger.Infof("WS", "handleAddContact Info (UserID: %d): Processing add contact request", userID)

// Líneas 17-25: SIMPLIFICAR parsing payload
var req AddContactPayload
payloadBytes, err := json.Marshal(payload)
if err != nil {
	return h.sendErrorMessage(client, "Internal error processing payload")
}
if err := json.Unmarshal(payloadBytes, &req); err != nil {
	return h.sendErrorMessage(client, "Invalid add contact payload structure")
}
// ↓ CAMBIAR A:
var req AddContactPayload
if err := json.Unmarshal(msg.Payload, &req); err != nil {
	conn.SendErrorNotification(msg.PID, 400, "Invalid add contact payload structure")
	return fmt.Errorf("invalid payload: %w", err)
}

// Líneas 27-30: ACTUALIZAR validación y manejo de errores
if targetUserID == 0 || targetUserID == userID {
	return h.sendErrorMessage(client, "Invalid target user ID")
}
// ↓ CAMBIAR A:
if req.TargetUserID == 0 || req.TargetUserID == userID {
	conn.SendErrorNotification(msg.PID, 400, "Invalid target user ID")
	return fmt.Errorf("invalid target user ID")
}

// Línea 33: CAMBIAR acceso a DB
err = h.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM User WHERE Id = ? AND StatusAuthorizedId = 1)", targetUserID).Scan(&targetExists)
// ↓ CAMBIAR A:
err = appCtx.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM User WHERE Id = ? AND StatusAuthorizedId = 1)", req.TargetUserID).Scan(&targetExists)

// Líneas 34-40: ACTUALIZAR manejo de errores
if err != nil {
	log.Printf("handleAddContact DB Error (UserID: %d): Failed checking target user %d existence: %v", userID, targetUserID, err)
	return h.sendErrorMessage(client, "Database error checking target user")
}
if !targetExists {
	return h.sendErrorMessage(client, "Target user not found or not active")
}
// ↓ CAMBIAR A:
if err != nil {
	logger.Errorf("WS", "handleAddContact DB Error (UserID: %d): Failed checking target user %d existence: %v", userID, req.TargetUserID, err)
	conn.SendErrorNotification(msg.PID, 500, "Database error checking target user")
	return err
}
if !targetExists {
	conn.SendErrorNotification(msg.PID, 404, "Target user not found or not active")
	return fmt.Errorf("target user not found")
}

// Línea 43: ACTUALIZAR llamada a función auxiliar
isContact, currentStatus, err := h.getContactStatus(userID, targetUserID)
// ↓ CAMBIAR A:
isContact, currentStatus, err := getContactStatus(appCtx, userID, req.TargetUserID)

// Líneas 44-47: ACTUALIZAR manejo de errores
if err != nil {
	return h.sendErrorMessage(client, "Database error checking existing contact")
}
// ↓ CAMBIAR A:
if err != nil {
	conn.SendErrorNotification(msg.PID, 500, "Database error checking existing contact")
	return err
}

// Líneas 53-55: CAMBIAR acceso a DB
q := "SELECT User1Id FROM Contact WHERE ((User1Id = ? AND User2Id = ?) OR (User1Id = ? AND User2Id = ?)) AND Status = 'pending'"
err = h.DB.QueryRow(q, userID, targetUserID, targetUserID, userID).Scan(&initiatorID)
// ↓ CAMBIAR A:
q := "SELECT User1Id FROM Contact WHERE ((User1Id = ? AND User2Id = ?) OR (User1Id = ? AND User2Id = ?)) AND Status = 'pending'"
err = appCtx.DB.QueryRow(q, userID, req.TargetUserID, req.TargetUserID, userID).Scan(&initiatorID)

// Líneas 56-58: ACTUALIZAR manejo de errores
if err != nil {
	return h.sendErrorMessage(client, "DB Error checking pending request")
}
// ↓ CAMBIAR A:
if err != nil {
	conn.SendErrorNotification(msg.PID, 500, "DB Error checking pending request")
	return err
}

// Línea 60: CAMBIAR validación de iniciador
if initiatorID == targetUserID {
// ↓ CAMBIAR A:
if initiatorID == req.TargetUserID {

// Línea 63: ACTUALIZAR llamada a función auxiliar
err = h.updateContactStatus(userID, targetUserID, newStatus)
// ↓ CAMBIAR A:
err = updateContactStatus(appCtx, userID, req.TargetUserID, newStatus)

// Líneas 64-66: ACTUALIZAR manejo de errores
if err != nil {
	return h.sendErrorMessage(client, "Failed to accept contact request")
}
// ↓ CAMBIAR A:
if err != nil {
	conn.SendErrorNotification(msg.PID, 500, "Failed to accept contact request")
	return err
}

// Línea 69: ACTUALIZAR llamada a notificación
h.notifyContactStatusUpdate(userID, targetUserID, newStatus)
// ↓ CAMBIAR A:
notifyContactStatusUpdate(appCtx, userID, req.TargetUserID, newStatus)

// Línea 70: ACTUALIZAR envío de respuesta de éxito
return h.sendSuccessMessage(client, fmt.Sprintf("Contact request from user %d accepted.", targetUserID))
// ↓ CAMBIAR A:
response := customws_types.ServerToClientMessage{
	PID:       msg.PID,
	Type:      "contact_request_accepted",
	Payload:   SuccessPayload{Message: fmt.Sprintf("Contact request from user %d accepted.", req.TargetUserID)},
	Timestamp: time.Now().UnixMilli(),
}
return conn.SendMessage(response)

// Línea 73: ACTUALIZAR manejo de caso pending
return h.sendErrorMessage(client, "Contact request already sent and pending")
// ↓ CAMBIAR A:
conn.SendErrorNotification(msg.PID, 409, "Contact request already sent and pending")
return fmt.Errorf("contact request already pending")

// Línea 75: ACTUALIZAR caso already contacts
return h.sendErrorMessage(client, "Users are already contacts")
// ↓ CAMBIAR A:
conn.SendErrorNotification(msg.PID, 409, "Users are already contacts")
return fmt.Errorf("users are already contacts")

// Línea 78: ACTUALIZAR caso rejected/blocked
return h.sendErrorMessage(client, fmt.Sprintf("Cannot add contact, previous status: %s", currentStatus))
// ↓ CAMBIAR A:
conn.SendErrorNotification(msg.PID, 403, fmt.Sprintf("Cannot add contact, previous status: %s", currentStatus))
return fmt.Errorf("cannot add contact, previous status: %s", currentStatus)

// Línea 80: ACTUALIZAR logging
log.Printf("handleAddContact Warn (UserID: %d): Unknown existing contact status '%s' with UserID %d", userID, currentStatus, targetUserID)
// ↓ CAMBIAR A:
logger.Warnf("WS", "handleAddContact Unknown status '%s' between UserID %d and %d", currentStatus, userID, req.TargetUserID)

// Línea 81: ACTUALIZAR manejo unknown status
return h.sendErrorMessage(client, "Unknown contact status")
// ↓ CAMBIAR A:
conn.SendErrorNotification(msg.PID, 500, "Unknown contact status")
return fmt.Errorf("unknown contact status: %s", currentStatus)

// Línea 86: CAMBIAR acceso a DB para inserción
_, err = h.DB.Exec("INSERT INTO Contact (User1Id, User2Id, Status, ChatId) VALUES (?, ?, ?, ?)", userID, targetUserID, newStatus, chatID)
// ↓ CAMBIAR A:
_, err = appCtx.DB.Exec("INSERT INTO Contact (User1Id, User2Id, Status, ChatId) VALUES (?, ?, ?, ?)", userID, req.TargetUserID, newStatus, chatID)

// Líneas 87-90: ACTUALIZAR manejo de errores
if err != nil {
	log.Printf("handleAddContact DB Error (UserID: %d): Failed inserting new contact request for UserID %d: %v", userID, targetUserID, err)
	return h.sendErrorMessage(client, "Failed to send contact request")
}
// ↓ CAMBIAR A:
if err != nil {
	logger.Errorf("WS", "handleAddContact DB Error (UserID: %d): Failed inserting contact request for UserID %d: %v", userID, req.TargetUserID, err)
	conn.SendErrorNotification(msg.PID, 500, "Failed to send contact request")
	return err
}

// Línea 93: ACTUALIZAR notificación
h.notifyContactStatusUpdate(userID, targetUserID, newStatus)
// ↓ CAMBIAR A:
notifyContactStatusUpdate(appCtx, userID, req.TargetUserID, newStatus)

// Línea 94: ACTUALIZAR respuesta de éxito
return h.sendSuccessMessage(client, "Contact request sent successfully")
// ↓ CAMBIAR A:
response := customws_types.ServerToClientMessage{
	PID:       msg.PID,
	Type:      "contact_request_sent",
	Payload:   SuccessPayload{Message: "Contact request sent successfully"},
	Timestamp: time.Now().UnixMilli(),
}
return conn.SendMessage(response)
```

#### Función 2: handleDeleteContact (Líneas 97-127) - **MIGRAR COMPLETAMENTE**

**Acción requerida:** Migrar función para eliminar contactos

**Cambios específicos:**

```go
// Línea 97: CAMBIAR firma
func (h *Hub) handleDeleteContact(client *Client, payload interface{}) error
// ↓ CAMBIAR A:
func handleDeleteContact(appCtx *AppContext, conn *customws.Connection[AppUserData], msg customws_types.ClientToServerMessage) error

// Líneas 98: CAMBIAR referencia usuario
userID := client.User.ID
// ↓ CAMBIAR A:
userID := conn.UserData.ID

// Líneas 101-109: SIMPLIFICAR parsing payload (mismo patrón que handleAddContact)
var req DeleteContactPayload
payloadBytes, err := json.Marshal(payload)
if err != nil {
	return h.sendErrorMessage(client, "Internal error processing payload")
}
if err := json.Unmarshal(payloadBytes, &req); err != nil {
	return h.sendErrorMessage(client, "Invalid delete contact payload structure")
}
// ↓ CAMBIAR A:
var req DeleteContactPayload
if err := json.Unmarshal(msg.Payload, &req); err != nil {
	conn.SendErrorNotification(msg.PID, 400, "Invalid delete contact payload structure")
	return fmt.Errorf("invalid payload: %w", err)
}

// Líneas 111-114: ACTUALIZAR validación
if targetUserID == 0 || targetUserID == userID {
	return h.sendErrorMessage(client, "Invalid target user ID for deletion")
}
// ↓ CAMBIAR A:
if req.TargetUserID == 0 || req.TargetUserID == userID {
	conn.SendErrorNotification(msg.PID, 400, "Invalid target user ID for deletion")
	return fmt.Errorf("invalid target user ID")
}

// Línea 118: CAMBIAR acceso a DB
result, err := h.DB.Exec(query, userID, targetUserID, targetUserID, userID)
// ↓ CAMBIAR A:
result, err := appCtx.DB.Exec(query, userID, req.TargetUserID, req.TargetUserID, userID)

// Líneas 119-122: ACTUALIZAR manejo de errores
if err != nil {
	log.Printf("handleDeleteContact DB Error (UserID: %d): Failed deleting contact with UserID %d: %v", userID, targetUserID, err)
	return h.sendErrorMessage(client, "Failed to delete contact")
}
// ↓ CAMBIAR A:
if err != nil {
	logger.Errorf("WS", "handleDeleteContact DB Error (UserID: %d): Failed deleting contact with UserID %d: %v", userID, req.TargetUserID, err)
	conn.SendErrorNotification(msg.PID, 500, "Failed to delete contact")
	return err
}

// Líneas 125-127: ACTUALIZAR caso no encontrado
if rowsAffected == 0 {
	return h.sendErrorMessage(client, "Contact not found")
}
// ↓ CAMBIAR A:
if rowsAffected == 0 {
	conn.SendErrorNotification(msg.PID, 404, "Contact not found")
	return fmt.Errorf("contact not found")
}

// Línea 130: ACTUALIZAR notificación
h.notifyContactStatusUpdate(userID, targetUserID, "deleted")
// ↓ CAMBIAR A:
notifyContactStatusUpdate(appCtx, userID, req.TargetUserID, "deleted")

// Línea 132: ACTUALIZAR respuesta de éxito
return h.sendSuccessMessage(client, "Contact deleted successfully")
// ↓ CAMBIAR A:
response := customws_types.ServerToClientMessage{
	PID:       msg.PID,
	Type:      "contact_deleted",
	Payload:   SuccessPayload{Message: "Contact deleted successfully"},
	Timestamp: time.Now().UnixMilli(),
}
return conn.SendMessage(response)
```

#### Función 3: updateContactStatus (Líneas 135-143) - **MIGRAR COMPLETAMENTE**

**Acción requerida:** Migrar función auxiliar para actualizar estado de contactos

**Cambios específicos:**

```go
// Línea 135: CAMBIAR firma
func (h *Hub) updateContactStatus(user1ID, user2ID int64, newStatus string) error
// ↓ CAMBIAR A:
func updateContactStatus(appCtx *AppContext, user1ID, user2ID int64, newStatus string) error

// Línea 137: CAMBIAR acceso a DB
_, err := h.DB.Exec(query, newStatus, user1ID, user2ID, user2ID, user1ID)
// ↓ CAMBIAR A:
_, err := appCtx.DB.Exec(query, newStatus, user1ID, user2ID, user2ID, user1ID)

// Línea 139: CAMBIAR logging
log.Printf("updateContactStatus DB Error (Users %d, %d): Failed setting status to '%s': %v", user1ID, user2ID, newStatus, err)
// ↓ CAMBIAR A:
logger.Errorf("WS", "updateContactStatus DB Error (Users %d, %d): Failed setting status to '%s': %v", user1ID, user2ID, newStatus, err)
```

#### Función 4: notifyContactStatusUpdate (Líneas 146-192) - **REFACTORIZAR COMPLETAMENTE**

**Acción requerida:** Refactorizar para usar ConnectionManager y nuevos tipos de mensaje

**Cambio completo requerido:**

```go
// Líneas 146-192: REEMPLAZAR función completa
func (h *Hub) notifyContactStatusUpdate(initiatorID, targetUserID int64, status string) {
	// ... lógica original con OutgoingMessage y json.Marshal
}
// ↓ REEMPLAZAR COMPLETAMENTE CON:
func notifyContactStatusUpdate(appCtx *AppContext, initiatorID, targetUserID int64, status string) {
	// Obtener datos del iniciador para la notificación
	initiatorInfo, err := models.GetUserBaseInfo(appCtx.DB, initiatorID)
	if err != nil {
		logger.Errorf("WS", "notifyContactStatusUpdate Error: Failed fetching initiator %d info: %v", initiatorID, err)
		return
	}

	// Construir payload de la notificación
	notificationPayload := map[string]interface{}{
		"status":        status,
		"initiatorId":   initiatorID,
		"initiatorName": fmt.Sprintf("%s %s", initiatorInfo.FirstName, initiatorInfo.LastName),
	}

	// Crear mensaje usando customws types
	notification := customws_types.ServerToClientMessage{
		PID:       customws.GenerateNotificationPID(),
		Type:      "contact_status_update",
		Payload:   notificationPayload,
		Timestamp: time.Now().UnixMilli(),
	}

	// Enviar notificación usando ConnectionManager
	if err := appCtx.ConnectionManager.SendMessageToUser(targetUserID, notification); err != nil {
		logger.Warnf("WS", "notifyContactStatusUpdate: Failed sending notification to UserID %d: %v", targetUserID, err)
	} else {
		logger.Infof("WS", "notifyContactStatusUpdate: Sent status '%s' notification to UserID %d from UserID %d", status, targetUserID, initiatorID)
	}

	// Si el estado es 'accepted', también notificar al iniciador
	if status == "accepted" && initiatorID != targetUserID {
		// Obtener datos del target para notificar al iniciador
		targetInfo, err := models.GetUserBaseInfo(appCtx.DB, targetUserID)
		if err != nil {
			logger.Errorf("WS", "notifyContactStatusUpdate Error: Failed fetching target %d info: %v", targetUserID, err)
			return
		}

		initiatorNotificationPayload := map[string]interface{}{
			"status":      status,
			"initiatorId": targetUserID,
			"initiatorName": fmt.Sprintf("%s %s", targetInfo.FirstName, targetInfo.LastName),
		}

		initiatorNotification := customws_types.ServerToClientMessage{
			PID:       customws.GenerateNotificationPID(),
			Type:      "contact_status_update",
			Payload:   initiatorNotificationPayload,
			Timestamp: time.Now().UnixMilli(),
		}

		appCtx.ConnectionManager.SendMessageToUser(initiatorID, initiatorNotification)
	}
}
```

#### Función 5: getContactStatus (Líneas 195-216) - **MIGRAR COMPLETAMENTE**

**Acción requerida:** Migrar función auxiliar para verificar estado de contactos

**Cambios específicos:**

```go
// Línea 195: CAMBIAR firma
func (h *Hub) getContactStatus(userID1, userID2 int64) (isContact bool, status string, err error)
// ↓ CAMBIAR A:
func getContactStatus(appCtx *AppContext, userID1, userID2 int64) (isContact bool, status string, err error)

// Línea 197: CAMBIAR acceso a DB
err = h.DB.QueryRow(query, userID1, userID2, userID2, userID1).Scan(&status)
// ↓ CAMBIAR A:
err = appCtx.DB.QueryRow(query, userID1, userID2, userID2, userID1).Scan(&status)

// Línea 202: CAMBIAR logging
log.Printf("getContactStatus DB Error (Users %d, %d): %v", userID1, userID2, err)
// ↓ CAMBIAR A:
logger.Errorf("WS", "getContactStatus DB Error (Users %d, %d): %v", userID1, userID2, err)
```

### Cambios Críticos Resumidos para handlers_contact.go:

1.  **Imports:** AGREGAR `customws_types`, `time`. ELIMINAR `"encoding/json"`, `"log"` donde no se parseen payloads directamente (ahora lo hace customws).
2.  **Firmas de Handlers:** Actualizar todas las firmas a `func handlerName(appCtx *AppContext, conn *customws.Connection[AppUserData], msg customws_types.ClientToServerMessage) error`.
3.  **Firmas Auxiliares:** `getBaseUserProfile` y `getCurriculum` deben tomar `appCtx *AppContext`.
4.  **Acceso a DB:** Cambiar `h.DB` a `appCtx.DB`.
5.  **Payload Parsing:** Usar `json.Unmarshal(msg.Payload, &reqStruct)`.
6.  **User Info:** `client.User.ID` → `conn.UserData.ID`, `client.User.RoleID` → `conn.UserData.RoleID`.
7.  **Respuestas:** Usar `conn.SendErrorNotification()` y `conn.SendMessage()` con `customws_types.ServerToClientMessage`.
8.  **Nuevos Tipos de Mensaje:** `MessageTypeEditProfileResponse`, `MessageTypeDeleteItemCurriculumResponse`.
9.  **Estado Online:** `h.GetClient()` → `appCtx.ConnectionManager.IsUserConnected()`.
10. **Llamadas a `getContactStatus` y `getCurriculum`:** Actualizar para usar `appCtx`.
11. **Actualización de `client.User`:** Eliminar la lógica de `handleEditProfile` que actualizaba `client.User`.

### Esquema de BD utilizado por handlers_contact.go:
```sql
User (Id, FirstName, LastName, UserName, Email, Phone, Sex, DocId, NationalityId, Birthdate, Picture, DegreeId, UniversityId, RoleId, StatusAuthorizedId, Summary, Address, Github, Linkedin)
Nationality (Id, CountryName)
Degree (Id, DegreeName)
University (Id, Name)
Role (Id, Name)
-- Más tablas de Curriculum (Education, WorkExperience, etc. - llamadas vía getCurriculum)
Contact (User1Id, User2Id, Status) -- Usado por getContactStatus
```

### 5. `handlers_curriculum.go` (Prioridad: MEDIA) - 230 líneas

**Estado:** Funciones auxiliares bien estructuradas - Solo requiere agregar handlers WebSocket principales

**Observación crítica:** Este archivo contiene 6 funciones auxiliares puras para obtener datos curriculares, pero **NO contiene handlers WebSocket principales**. Las funciones existentes son:

#### Funciones Auxiliares Existentes (MANTENER SIN CAMBIOS):
- `getEducation(db *sql.DB, userID int64)` - Líneas 13-46
- `getWorkExperience(db *sql.DB, userID int64)` - Líneas 48-83  
- `getSkills(db *sql.DB, userID int64)` - Líneas 85-103
- `getCertifications(db *sql.DB, userID int64)` - Líneas 105-129
- `getLanguages(db *sql.DB, userID int64)` - Líneas 131-149
- `getProjects(db *sql.DB, userID int64)` - Líneas 151-190

**Recomendación:** **MANTENER todas las funciones auxiliares sin cambios** - están bien optimizadas y son funciones puras reutilizables.

#### Handlers WebSocket a AGREGAR:

1. **handleGetCurriculum** - Para obtener currículum completo:
   ```go
   func handleGetCurriculum(appCtx *AppContext, conn *customws.Connection[AppUserData], msg customws_types.ClientToServerMessage) error
   ```

2. **handleDeleteItemCurriculum** - Para eliminar items del currículum:
   ```go
   func handleDeleteItemCurriculum(appCtx *AppContext, conn *customws.Connection[AppUserData], msg customws_types.ClientToServerMessage) error
   ```

#### Cambios Requeridos:
- **AGREGAR imports:** `"encoding/json"`, `"time"`, `customws_types`
- **AGREGAR constantes:** `"curriculum_data"`, `"curriculum_item_deleted"`
- **Las funciones auxiliares existentes** se mantendrán y serán llamadas desde los nuevos handlers

### Esquema de BD utilizado por handlers_curriculum.go:
```sql
-- Currículum y perfil profesional
Education (Id, PersonId, Institution, Degree, Campus, GraduationDate, CountryId)
WorkExperience (Id, PersonId, Company, Position, StartDate, EndDate, Description, CountryId)
Skills (Id, PersonId, Skill, Level)
Certifications (Id, PersonId, Certification, Institution, DateObtained)
Languages (Id, PersonId, Language, Level)
Project (Id, PersonID, Title, Role, Description, Company, Document, ProjectStatus, StartDate, ExpectedEndDate)
```

## Orden de Migración Recomendado

### Fase 1: Infraestructura Core (CRÍTICA)
1. **`types.go`** - Actualizar tipos base
2. **`router.go`** - Implementar nuevo sistema de routing
3. **Probar conectividad básica**

### Fase 2: Funcionalidad Chat (ALTA)
1. **`handlers_chat.go`** - Completar migración restante
2. **`handlers_contact.go`** - Migrar gestión de contactos
3. **Probar funcionalidad de mensajería completa**

### Fase 3: Funcionalidad Extendida (MEDIA)
1. **`handlers_profile.go`** - Migrar gestión de perfiles
2. **`handlers_search.go`** - Migrar búsqueda de usuarios
3. **`handlers_list.go`** - Migrar listados
4. **`handlers_curriculum.go`** - Migrar gestión curricular
5. **`handlers_notification.go`** - Migrar notificaciones

### Fase 4: Limpieza (BAJA)
1. **`hub_helpers.go`** - Migrar funciones útiles restantes
2. **Eliminar archivos obsoletos:** `hub.go`, `client.go`
3. **Optimización y limpieza final**

## Patrones de Migración Establecidos

### 1. Firma de Funciones Handler

```go
// Antes:
func (h *Hub) handlerFunction(client *Client, payload interface{}) error

// Después:
func handlerFunction(appCtx *AppContext, conn *customws.Connection[AppUserData], msg customws_types.ClientToServerMessage) error
```

### 2. Acceso a Base de Datos

```go
// Antes:
h.DB.Query(...)

// Después:  
appCtx.DB.Query(...)
```

### 3. Envío de Mensajes

```go
// Antes:
h.SendToUser(userID, message)
client.SendToMe(message)

// Después:
appCtx.ConnectionManager.SendMessageToUser(userID, message)
conn.SendMessage(message)
```

### 4. Manejo de Errores

```go
// Antes:
return h.sendErrorMessage(client, "Error description")

// Después:
conn.SendErrorNotification(msg.PID, 400, "Error description")
return fmt.Errorf("error description")
```

### 5. Estructura de Respuesta

```go
// Crear respuesta usando customws_types
response := customws_types.ServerToClientMessage{
    PID:  customws.GenerateResponsePID(), // o usar msg.PID para respuesta directa
    Type: "response_type",
    Payload: payloadStruct,
    Timestamp: time.Now().UnixMilli(),
}
return conn.SendMessage(response)
```

## Consideraciones de Rendimiento

### 1. Gestión de Conexiones
- `customws` maneja automáticamente el pooling de conexiones
- Límites configurables en `types.Config`
- Cleanup automático de conexiones muertas

### 2. Manejo de Memoria
- Canales con buffer configurables
- Context-based cancellation para evitar goroutine leaks
- Sync.Map para acceso concurrente eficiente

### 3. Base de Datos
- Mantener el pool de conexiones existente en `db.go`
- Implementar prepared statements donde sea posible
- Añadir índices para consultas frecuentes de búsqueda

## Testing y Validación

### 1. Testing Unitario
- Crear tests para cada handler migrado
- Usar mocks para `AppContext` y `ConnectionManager`
- Validar manejo de errores y casos edge

### 2. Testing de Integración
- Probar flujos completos de mensajería
- Validar notificaciones entre usuarios
- Probar bajo carga con múltiples conexiones

### 3. Testing de Rendimiento
- Benchmark de handlers individuales
- Testing de concurrencia
- Profiling de memoria y CPU

## Riesgos y Mitigaciones

### 1. Riesgos Técnicos
- **Pérdida de mensajes durante migración:** Implementar migración por fases
- **Incompatibilidad de tipos:** Mantener compatibilidad en `types.go`
- **Regresión de funcionalidad:** Testing exhaustivo

### 2. Riesgos de Rendimiento
- **Overhead de migración:** Monitoreo de métricas
- **Memory leaks:** Uso correcto de contexts y cleanup
- **DB bottlenecks:** Optimización de queries

### 3. Mitigaciones
- Rollback plan para cada fase
- Feature flags para nueva funcionalidad
- Monitoring y alertas en producción

## Métricas de Éxito

- **Conectividad:** Soporte para 1M+ conexiones concurrentes teóricas
- **Latencia:** < 10ms para mensajes locales
- **Throughput:** > 10K mensajes/segundo
- **Memoria:** < 1KB por conexión activa
- **Disponibilidad:** 99.9% uptime

Este plan asegura una migración ordenada y sistemática hacia la nueva arquitectura escalable, manteniendo la funcionalidad existente mientras se aprovechan las ventajas de rendimiento y escalabilidad de `customws`. 

### 6. `handlers_profile.go` (Prioridad: MEDIA) - 412 líneas

**Estado:** Migración completa requerida - Contiene handlers para obtener y editar perfiles, y eliminar items del currículum, además de funciones auxiliares clave.

#### Análisis línea por línea:

**Líneas 1-10: Package e imports**
```go
// CAMBIAR imports
package websocket

import (
	"database/sql"      // MANTENER
	"encoding/json"     // ELIMINAR - no necesario con customws
	"fmt"               // MANTENER
	"sync"              // MANTENER (usado en getCurriculum)

	"github.com/davidM20/micro-service-backend-go.git/internal/models" // MANTENER
	"github.com/davidM20/micro-service-backend-go.git/pkg/logger"      // MANTENER
	customws_types "github.com/davidM20/micro-service-backend-go.git/pkg/customws/types" // AGREGAR
	"time"                                                               // AGREGAR
)
```

**Líneas 12-26: Structs locales**
- `ProfileResponse` (Líneas 15-20): **ELIMINAR COMENTADO** - Ya se indica que se usa `MyProfileResponse` de `types.go`.
- `DeleteCurriculumItemRequest` (Líneas 23-26): **MANTENER** - Es usado por `handleDeleteItemCurriculum`. Podría moverse a `types.go` como `DeleteItemPayload` si no existe ya con ese nombre.

#### Función 1: handleGetProfile (Líneas 31-168) - **MIGRAR COMPLETAMENTE**

**Acción requerida:** Migrar función principal para obtener perfiles de usuario (propio o de otros).

**Cambios específicos:**

```go
// Línea 31: CAMBIAR firma
func (h *Hub) handleGetProfile(client *Client, payload interface{}) error
// ↓ CAMBIAR A:
func handleGetProfile(appCtx *AppContext, conn *customws.Connection[AppUserData], msg customws_types.ClientToServerMessage) error

// Líneas 32-40: SIMPLIFICAR parsing payload
var req ProfileRequestPayload
if msg.Payload != nil {
	if err := json.Unmarshal(msg.Payload, &req); err != nil {
		logger.Infof("WS", "handleGetProfile Info (UserID: %d): Payload is not a ProfileRequestPayload or is empty: %v", conn.UserData.ID, err)
		// Continuar, targetUserID será el propio usuario si req está vacía o malformada
	}
}

// Líneas 42-45: ACTUALIZAR lógica de targetUserID
targetUserID := client.User.ID
if req.TargetUserID != 0 {
	targetUserID = req.TargetUserID
}
// ↓ CAMBIAR A:
targetUserID := conn.UserData.ID
if req.TargetUserID != 0 {
	targetUserID = req.TargetUserID
}

// Línea 48: ACTUALIZAR llamada a función auxiliar
userBase, err := h.getBaseUserProfile(targetUserID)
// ↓ CAMBIAR A:
userBase, err := getBaseUserProfile(appCtx, targetUserID)

// Líneas 49-55: ACTUALIZAR manejo de errores
if err == sql.ErrNoRows {
	return h.sendErrorMessage(client, "Requested user profile not found")
}
if err != nil {
	return h.sendErrorMessage(client, "Failed to retrieve profile data")
}
// ↓ CAMBIAR A:
if err == sql.ErrNoRows {
	conn.SendErrorNotification(msg.PID, 404, "Requested user profile not found")
	return fmt.Errorf("user profile not found: %d", targetUserID)
}
if err != nil {
	conn.SendErrorNotification(msg.PID, 500, "Failed to retrieve profile data")
	return err
}

// Líneas 57-107: MANTENER lógica de mapeo a MyProfileResponse (ya es correcta)

// Línea 110: ACTUALIZAR obtención de estado online
_, isOnline := h.GetClient(targetUserID)
// ↓ CAMBIAR A:
isOnline := appCtx.ConnectionManager.IsUserConnected(targetUserID)

// Líneas 112-125: ACTUALIZAR verificación de permisos
if targetUserID != client.User.ID {
	isContact, contactStatus, err := h.getContactStatus(client.User.ID, targetUserID)
	// ... lógica de error y admin check ...
	return h.sendErrorMessage(client, "Permission denied to view this profile")
}
// ↓ CAMBIAR A:
if targetUserID != conn.UserData.ID {
	isContact, contactStatus, err := getContactStatus(appCtx, conn.UserData.ID, targetUserID)
	if err != nil {
		conn.SendErrorNotification(msg.PID, 500, "Failed to check contact status")
		return err
	}
	isAdmin := conn.UserData.RoleID == 7 || conn.UserData.RoleID == 8 // Asumiendo que RoleID está en AppUserData
	if !isAdmin && !isContact { // y que getContactStatus devuelve `isContact` correctamente
		conn.SendErrorNotification(msg.PID, 403, "Permission denied to view this profile")
		return fmt.Errorf("permission denied to view profile %d", targetUserID)
	}
}

// Línea 128: ACTUALIZAR llamada a función auxiliar
curriculum, err := h.getCurriculum(targetUserID)
// ↓ CAMBIAR A:
curriculum, err := getCurriculum(appCtx, targetUserID) // getCurriculum necesita ser migrada también

// Líneas 135-138: MANTENER asignación de currículum (profileResp.Curriculum = *curriculum)

// Líneas 140-152: ACTUALIZAR envío de respuesta
outgoingMsg := OutgoingMessage{Type: MessageTypeGetProfileResponse, Payload: profileResp}
jsonMsg, err := json.Marshal(outgoingMsg)
client.SendToMe(jsonMsg)
// ↓ CAMBIAR A:
response := customws_types.ServerToClientMessage{
	PID:       msg.PID,
	Type:      MessageTypeGetProfileResponse, // Mantener la constante de types.go
	Payload:   profileResp,
	Timestamp: time.Now().UnixMilli(),
}
return conn.SendMessage(response)
```

#### Función 2: handleGetMyProfile (Líneas 170-175) - **MIGRAR COMPLETAMENTE**

**Acción requerida:** Simplificar y llamar al nuevo `handleGetProfile`.

```go
// Línea 170: CAMBIAR firma
func (h *Hub) handleGetMyProfile(client *Client) error
// ↓ CAMBIAR A:
func handleGetMyProfile(appCtx *AppContext, conn *customws.Connection[AppUserData], msg customws_types.ClientToServerMessage) error

// Línea 173: SIMPLIFICAR llamada
return h.handleGetProfile(client, nil) // Payload nil para indicar perfil propio
// ↓ CAMBIAR A:
// El payload de msg ya sería nil o un ProfileRequestPayload vacío por defecto.
// Si se quiere forzar que sea el perfil propio, se puede pasar un ProfileRequestPayload con TargetUserID = 0 o conn.UserData.ID
// O simplemente llamar a handleGetProfile, que ya tiene la lógica para deducir el targetUserID.
return handleGetProfile(appCtx, conn, msg)
```

#### Función 3: handleEditProfile (Líneas 177-228) - **MIGRAR COMPLETAMENTE**

**Acción requerida:** Migrar función para actualizar perfil de usuario.

```go
// Línea 177: CAMBIAR firma
func (h *Hub) handleEditProfile(client *Client, payload interface{}) error
// ↓ CAMBIAR A:
func handleEditProfile(appCtx *AppContext, conn *customws.Connection[AppUserData], msg customws_types.ClientToServerMessage) error

// Líneas 181-194: SIMPLIFICAR parsing payload
var updateReq models.User // Usar `EditProfilePayload` de `types.go` en su lugar
// ... parsing manual ...
// ↓ CAMBIAR A:
var req EditProfilePayload // Usar el tipo de types.go
if err := json.Unmarshal(msg.Payload, &req); err != nil {
	conn.SendErrorNotification(msg.PID, 400, "Invalid update payload format")
	return fmt.Errorf("invalid payload: %w", err)
}

// Líneas 196-207: CAMBIAR acceso a DB y parámetros
_, err := h.DB.Exec(query, updateReq.FirstName, ..., userID)
// ↓ CAMBIAR A:
// Esta parte necesita una refactorización cuidadosa para construir la query y los args
// a partir de los campos no nulos de `req` (EditProfilePayload).
// Ejemplo conceptual:
// updates := []string{}
// args := []interface{}{}
// if req.Summary != nil { updates = append(updates, "Summary = ?"); args = append(args, *req.Summary) }
// ... etc. para todos los campos ...
// if len(updates) == 0 { /* no hay nada que actualizar */ return ... }
// query := fmt.Sprintf("UPDATE User SET %s WHERE Id = ?", strings.Join(updates, ", "))
// args = append(args, conn.UserData.ID)
// _, err := appCtx.DB.Exec(query, args...)

// Líneas 209-212: ACTUALIZAR manejo de errores
if err != nil {
	return h.sendErrorMessage(client, "Failed to update profile information")
}
// ↓ CAMBIAR A:
if err != nil {
	conn.SendErrorNotification(msg.PID, 500, "Failed to update profile information")
	return err
}

// Líneas 216-222: ACTUALIZAR `client.User` (ya no existirá)
// Esta lógica de actualizar el `client.User` en memoria ya no es necesaria con `customws`,
// ya que `AppUserData` es más estático o se recarga con cada conexión.
// Si se necesita refrescar `conn.UserData`, se debería hacer explícitamente.
// ELIMINAR esta sección.

// Línea 224: ACTUALIZAR respuesta de éxito
return h.sendSuccessMessage(client, "Profile updated successfully")
// ↓ CAMBIAR A:
response := customws_types.ServerToClientMessage{
	PID:       msg.PID,
	Type:      MessageTypeEditProfileResponse, // AGREGAR constante en types.go
	Payload:   SuccessPayload{Message: "Profile updated successfully"},
	Timestamp: time.Now().UnixMilli(),
}
return conn.SendMessage(response)
```

#### Función 4: handleDeleteItemCurriculum (Líneas 230-293) - **MIGRAR COMPLETAMENTE**

**Acción requerida:** Migrar función para eliminar items del currículum.

```go
// Línea 230: CAMBIAR firma
func (h *Hub) handleDeleteItemCurriculum(client *Client, payload interface{}) error
// ↓ CAMBIAR A:
func handleDeleteItemCurriculum(appCtx *AppContext, conn *customws.Connection[AppUserData], msg customws_types.ClientToServerMessage) error

// Líneas 234-247: SIMPLIFICAR parsing payload
var req DeleteCurriculumItemRequest // Usar `DeleteItemPayload` de `types.go`
// ... parsing manual ...
// ↓ CAMBIAR A:
var req DeleteItemPayload // Usar el tipo de types.go
if err := json.Unmarshal(msg.Payload, &req); err != nil {
	conn.SendErrorNotification(msg.PID, 400, "Invalid delete payload format")
	return fmt.Errorf("invalid payload: %w", err)
}

// Líneas 249-252: MANTENER validación (req.ItemType y req.ItemID)
// Ajustar el mensaje de error para `conn.SendErrorNotification`

// Líneas 254-274: MANTENER lógica de switch para `tableName` y `personIdColumn`

// Línea 276: CAMBIAR acceso a DB
result, err := h.DB.Exec(query, req.ItemID, userID)
// ↓ CAMBIAR A:
result, err := appCtx.DB.Exec(query, req.ItemID, conn.UserData.ID)

// Líneas 277-286: ACTUALIZAR manejo de errores y rowsAffected
// Usar `conn.SendErrorNotification`

// Línea 290: ACTUALIZAR respuesta de éxito
return h.sendSuccessMessage(client, fmt.Sprintf("%s deleted successfully", req.ItemType))
// ↓ CAMBIAR A:
response := customws_types.ServerToClientMessage{
	PID:       msg.PID,
	Type:      MessageTypeDeleteItemCurriculumResponse, // AGREGAR constante en types.go
	Payload:   SuccessPayload{Message: fmt.Sprintf("%s deleted successfully", req.ItemType)},
	Timestamp: time.Now().UnixMilli(),
}
return conn.SendMessage(response)
```

#### Función 5: getBaseUserProfile (Líneas 298-344) - **MIGRAR (AUXILIAR)**

**Acción requerida:** Cambiar firma para usar `appCtx`.

```go
// Línea 298: CAMBIAR firma
func (h *Hub) getBaseUserProfile(userID int64) (*models.User, error)
// ↓ CAMBIAR A:
func getBaseUserProfile(appCtx *AppContext, userID int64) (*models.User, error)

// Línea 317: CAMBIAR acceso a DB
err := h.DB.QueryRow(query, userID).Scan(...)
// ↓ CAMBIAR A:
err := appCtx.DB.QueryRow(query, userID).Scan(...)

// Resto de la lógica MANTENER (scan y asignación de NullString)
```

#### Función 6: getCurriculum (Líneas 346-401) - **MIGRAR (AUXILIAR)**

**Acción requerida:** Cambiar firma para usar `appCtx` y ajustar llamadas a funciones auxiliares de currículum.

```go
// Línea 347: CAMBIAR firma
func (h *Hub) getCurriculum(userID int64) (*Curriculum, error)
// ↓ CAMBIAR A:
func getCurriculum(appCtx *AppContext, userID int64) (*Curriculum, error)

// Línea 356: CAMBIAR llamada a DB dentro de la goroutine
data, fetchErr := fetchFunc(h.DB, userID)
// ↓ CAMBIAR A:
data, fetchErr := fetchFunc(appCtx.DB, userID) // Asumiendo que las funciones get* de curriculum se adaptan para recibir appCtx.DB

// Líneas 383-388: ACTUALIZAR llamadas a fetch
// Las funciones getEducation, getWorkExperience, etc. (de handlers_curriculum.go) deben aceptar (appCtx.DB, userID)
// O mejor, que acepten (appCtx *AppContext, userID int64) y accedan a appCtx.DB internamente.
// Ejemplo:
fetch(func(db *sql.DB, id int64) (interface{}, error) { return getEducation(db, id) }, &curriculum.Education)
// ↓ CAMBIAR A (si getEducation toma AppContext):
fetch(func(ctx *AppContext, id int64) (interface{}, error) { return getEducation(ctx, id) }, &curriculum.Education)
// O (si getEducation toma *sql.DB, pasar appCtx.DB):
fetch(func(db *sql.DB, id int64) (interface{}, error) { return getEducation(db, id) }, &curriculum.Education)
// Este último es el patrón actual de handlers_curriculum.go, por lo que solo el `h.DB` en la línea 356 necesita cambiar a `appCtx.DB`.

// Resto de la lógica MANTENER (sync.WaitGroup, error handling, type assertions)
```

**Líneas 403-410: Stubs de actualización de currículum - ELIMINAR COMENTADOS**

### Cambios Críticos Resumidos para handlers_profile.go:

1.  **Imports:** AGREGAR `customws_types`, `time`. ELIMINAR `encoding/json` donde no se parseen payloads directamente (ahora lo hace customws).
2.  **Structs locales:** Mover `DeleteCurriculumItemRequest` a `types.go` (posiblemente como `DeleteItemPayload`).
3.  **Firmas de Handlers:** Actualizar todas las firmas a `func handlerName(appCtx *AppContext, conn *customws.Connection[AppUserData], msg customws_types.ClientToServerMessage) error`.
4.  **Firmas Auxiliares:** `getBaseUserProfile` y `getCurriculum` deben tomar `appCtx *AppContext`.
5.  **Acceso a DB:** Cambiar `h.DB` a `appCtx.DB`.
6.  **Payload Parsing:** Usar `json.Unmarshal(msg.Payload, &reqStruct)`.
7.  **User Info:** `client.User.ID` → `conn.UserData.ID`, `client.User.RoleID` → `conn.UserData.RoleID`.
8.  **Respuestas:** Usar `conn.SendErrorNotification()` y `conn.SendMessage()` con `customws_types.ServerToClientMessage`.
9.  **Nuevos Tipos de Mensaje:** `MessageTypeEditProfileResponse`, `MessageTypeDeleteItemCurriculumResponse`.
10. **Estado Online:** `h.GetClient()` → `appCtx.ConnectionManager.IsUserConnected()`.
11. **Llamadas a `getContactStatus` y `getCurriculum`:** Actualizar para usar `appCtx`.
12. **Actualización de `client.User`:** Eliminar la lógica de `handleEditProfile` que actualizaba `client.User`.

### Esquema de BD utilizado por handlers_profile.go:
```sql
User (Id, FirstName, LastName, UserName, Email, Phone, Sex, DocId, NationalityId, Birthdate, Picture, DegreeId, UniversityId, RoleId, StatusAuthorizedId, Summary, Address, Github, Linkedin)
Nationality (Id, CountryName)
Degree (Id, DegreeName)
University (Id, Name)
Role (Id, Name)
-- Más tablas de Curriculum (Education, WorkExperience, etc. - llamadas vía getCurriculum)
Contact (User1Id, User2Id, Status) -- Usado por getContactStatus
```

### 7. `handlers_search.go` (Prioridad: MEDIA)

**Estado:** Migración completa requerida - Contiene handlers para obtener y editar perfiles, y eliminar items del currículum, además de funciones auxiliares clave.

#### Análisis línea por línea:

**Líneas 1-10: Package e imports**
```go
// CAMBIAR imports
package websocket

import (
	"database/sql"      // MANTENER
	"encoding/json"     // ELIMINAR - no necesario con customws
	"fmt"               // MANTENER
	"sync"              // MANTENER (usado en getCurriculum)

	"github.com/davidM20/micro-service-backend-go.git/internal/models" // MANTENER
	"github.com/davidM20/micro-service-backend-go.git/pkg/logger"      // MANTENER
	customws_types "github.com/davidM20/micro-service-backend-go.git/pkg/customws/types" // AGREGAR
	"time"                                                               // AGREGAR
)
```

**Líneas 12-26: Structs locales**
- `ProfileResponse` (Líneas 15-20): **ELIMINAR COMENTADO** - Ya se indica que se usa `MyProfileResponse` de `types.go`.
- `DeleteCurriculumItemRequest` (Líneas 23-26): **MANTENER** - Es usado por `handleDeleteItemCurriculum`. Podría moverse a `types.go` como `DeleteItemPayload` si no existe ya con ese nombre.

#### Función 1: handleGetProfile (Líneas 31-168) - **MIGRAR COMPLETAMENTE**

**Acción requerida:** Migrar función principal para obtener perfiles de usuario (propio o de otros).

**Cambios específicos:**

```go
// Línea 31: CAMBIAR firma
func (h *Hub) handleGetProfile(client *Client, payload interface{}) error
// ↓ CAMBIAR A:
func handleGetProfile(appCtx *AppContext, conn *customws.Connection[AppUserData], msg customws_types.ClientToServerMessage) error

// Líneas 32-40: SIMPLIFICAR parsing payload
var req ProfileRequestPayload
if msg.Payload != nil {
	if err := json.Unmarshal(msg.Payload, &req); err != nil {
		logger.Infof("WS", "handleGetProfile Info (UserID: %d): Payload is not a ProfileRequestPayload or is empty: %v", conn.UserData.ID, err)
		// Continuar, targetUserID será el propio usuario si req está vacía o malformada
	}
}

// Líneas 42-45: ACTUALIZAR lógica de targetUserID
targetUserID := client.User.ID
if req.TargetUserID != 0 {
	targetUserID = req.TargetUserID
}
// ↓ CAMBIAR A:
targetUserID := conn.UserData.ID
if req.TargetUserID != 0 {
	targetUserID = req.TargetUserID
}

// Línea 48: ACTUALIZAR llamada a función auxiliar
userBase, err := h.getBaseUserProfile(targetUserID)
// ↓ CAMBIAR A:
userBase, err := getBaseUserProfile(appCtx, targetUserID)

// Líneas 49-55: ACTUALIZAR manejo de errores
if err == sql.ErrNoRows {
	return h.sendErrorMessage(client, "Requested user profile not found")
}
if err != nil {
	return h.sendErrorMessage(client, "Failed to retrieve profile data")
}
// ↓ CAMBIAR A:
if err == sql.ErrNoRows {
	conn.SendErrorNotification(msg.PID, 404, "Requested user profile not found")
	return fmt.Errorf("user profile not found: %d", targetUserID)
}
if err != nil {
	conn.SendErrorNotification(msg.PID, 500, "Failed to retrieve profile data")
	return err
}

// Líneas 57-107: MANTENER lógica de mapeo a MyProfileResponse (ya es correcta)

// Línea 110: ACTUALIZAR obtención de estado online
_, isOnline := h.GetClient(targetUserID)
// ↓ CAMBIAR A:
isOnline := appCtx.ConnectionManager.IsUserConnected(targetUserID)

// Líneas 112-125: ACTUALIZAR verificación de permisos
if targetUserID != client.User.ID {
	isContact, contactStatus, err := h.getContactStatus(client.User.ID, targetUserID)
	// ... lógica de error y admin check ...
	return h.sendErrorMessage(client, "Permission denied to view this profile")
}
// ↓ CAMBIAR A:
if targetUserID != conn.UserData.ID {
	isContact, contactStatus, err := getContactStatus(appCtx, conn.UserData.ID, targetUserID)
	if err != nil {
		conn.SendErrorNotification(msg.PID, 500, "Failed to check contact status")
		return err
	}
	isAdmin := conn.UserData.RoleID == 7 || conn.UserData.RoleID == 8 // Asumiendo que RoleID está en AppUserData
	if !isAdmin && !isContact { // y que getContactStatus devuelve `isContact` correctamente
		conn.SendErrorNotification(msg.PID, 403, "Permission denied to view this profile")
		return fmt.Errorf("permission denied to view profile %d", targetUserID)
	}
}

// Línea 128: ACTUALIZAR llamada a función auxiliar
curriculum, err := h.getCurriculum(targetUserID)
// ↓ CAMBIAR A:
curriculum, err := getCurriculum(appCtx, targetUserID) // getCurriculum necesita ser migrada también

// Líneas 135-138: MANTENER asignación de currículum (profileResp.Curriculum = *curriculum)

// Líneas 140-152: ACTUALIZAR envío de respuesta
outgoingMsg := OutgoingMessage{Type: MessageTypeGetProfileResponse, Payload: profileResp}
jsonMsg, err := json.Marshal(outgoingMsg)
client.SendToMe(jsonMsg)
// ↓ CAMBIAR A:
response := customws_types.ServerToClientMessage{
	PID:       msg.PID,
	Type:      MessageTypeGetProfileResponse, // Mantener la constante de types.go
	Payload:   profileResp,
	Timestamp: time.Now().UnixMilli(),
}
return conn.SendMessage(response)
```

#### Función 2: handleGetMyProfile (Líneas 170-175) - **MIGRAR COMPLETAMENTE**

**Acción requerida:** Simplificar y llamar al nuevo `handleGetProfile`.

```go
// Línea 170: CAMBIAR firma
func (h *Hub) handleGetMyProfile(client *Client) error
// ↓ CAMBIAR A:
func handleGetMyProfile(appCtx *AppContext, conn *customws.Connection[AppUserData], msg customws_types.ClientToServerMessage) error

// Línea 173: SIMPLIFICAR llamada
return h.handleGetProfile(client, nil) // Payload nil para indicar perfil propio
// ↓ CAMBIAR A:
// El payload de msg ya sería nil o un ProfileRequestPayload vacío por defecto.
// Si se quiere forzar que sea el perfil propio, se puede pasar un ProfileRequestPayload con TargetUserID = 0 o conn.UserData.ID
// O simplemente llamar a handleGetProfile, que ya tiene la lógica para deducir el targetUserID.
return handleGetProfile(appCtx, conn, msg)
```

#### Función 3: handleEditProfile (Líneas 177-228) - **MIGRAR COMPLETAMENTE**

**Acción requerida:** Migrar función para actualizar perfil de usuario.

```go
// Línea 177: CAMBIAR firma
func (h *Hub) handleEditProfile(client *Client, payload interface{}) error
// ↓ CAMBIAR A:
func handleEditProfile(appCtx *AppContext, conn *customws.Connection[AppUserData], msg customws_types.ClientToServerMessage) error

// Líneas 181-194: SIMPLIFICAR parsing payload
var updateReq models.User // Usar `EditProfilePayload` de `types.go` en su lugar
// ... parsing manual ...
// ↓ CAMBIAR A:
var req EditProfilePayload // Usar el tipo de types.go
if err := json.Unmarshal(msg.Payload, &req); err != nil {
	conn.SendErrorNotification(msg.PID, 400, "Invalid update payload format")
	return fmt.Errorf("invalid payload: %w", err)
}

// Líneas 196-207: CAMBIAR acceso a DB y parámetros
_, err := h.DB.Exec(query, updateReq.FirstName, ..., userID)
// ↓ CAMBIAR A:
// Esta parte necesita una refactorización cuidadosa para construir la query y los args
// a partir de los campos no nulos de `req` (EditProfilePayload).
// Ejemplo conceptual:
// updates := []string{}
// args := []interface{}{}
// if req.Summary != nil { updates = append(updates, "Summary = ?"); args = append(args, *req.Summary) }
// ... etc. para todos los campos ...
// if len(updates) == 0 { /* no hay nada que actualizar */ return ... }
// query := fmt.Sprintf("UPDATE User SET %s WHERE Id = ?", strings.Join(updates, ", "))
// args = append(args, conn.UserData.ID)
// _, err := appCtx.DB.Exec(query, args...)

// Líneas 209-212: ACTUALIZAR manejo de errores
if err != nil {
	return h.sendErrorMessage(client, "Failed to update profile information")
}
// ↓ CAMBIAR A:
if err != nil {
	conn.SendErrorNotification(msg.PID, 500, "Failed to update profile information")
	return err
}

// Líneas 216-222: ACTUALIZAR `client.User` (ya no existirá)
// Esta lógica de actualizar el `client.User` en memoria ya no es necesaria con `customws`,
// ya que `AppUserData` es más estático o se recarga con cada conexión.
// Si se necesita refrescar `conn.UserData`, se debería hacer explícitamente.
// ELIMINAR esta sección.

// Línea 224: ACTUALIZAR respuesta de éxito
return h.sendSuccessMessage(client, "Profile updated successfully")
// ↓ CAMBIAR A:
response := customws_types.ServerToClientMessage{
	PID:       msg.PID,
	Type:      MessageTypeEditProfileResponse, // AGREGAR constante en types.go
	Payload:   SuccessPayload{Message: "Profile updated successfully"},
	Timestamp: time.Now().UnixMilli(),
}
return conn.SendMessage(response)
```

#### Función 4: handleDeleteItemCurriculum (Líneas 230-293) - **MIGRAR COMPLETAMENTE**

**Acción requerida:** Migrar función para eliminar items del currículum.

```go
// Línea 230: CAMBIAR firma
func (h *Hub) handleDeleteItemCurriculum(client *Client, payload interface{}) error
// ↓ CAMBIAR A:
func handleDeleteItemCurriculum(appCtx *AppContext, conn *customws.Connection[AppUserData], msg customws_types.ClientToServerMessage) error

// Líneas 234-247: SIMPLIFICAR parsing payload
var req DeleteCurriculumItemRequest // Usar `DeleteItemPayload` de `types.go`
// ... parsing manual ...
// ↓ CAMBIAR A:
var req DeleteItemPayload // Usar el tipo de types.go
if err := json.Unmarshal(msg.Payload, &req); err != nil {
	conn.SendErrorNotification(msg.PID, 400, "Invalid delete payload format")
	return fmt.Errorf("invalid payload: %w", err)
}

// Líneas 249-252: MANTENER validación (req.ItemType y req.ItemID)
// Ajustar el mensaje de error para `conn.SendErrorNotification`

// Líneas 254-274: MANTENER lógica de switch para `tableName` y `personIdColumn`

// Línea 276: CAMBIAR acceso a DB
result, err := h.DB.Exec(query, req.ItemID, userID)
// ↓ CAMBIAR A:
result, err := appCtx.DB.Exec(query, req.ItemID, conn.UserData.ID)

// Líneas 277-286: ACTUALIZAR manejo de errores y rowsAffected
// Usar `conn.SendErrorNotification`

// Línea 290: ACTUALIZAR respuesta de éxito
return h.sendSuccessMessage(client, fmt.Sprintf("%s deleted successfully", req.ItemType))
// ↓ CAMBIAR A:
response := customws_types.ServerToClientMessage{
	PID:       msg.PID,
	Type:      MessageTypeDeleteItemCurriculumResponse, // AGREGAR constante en types.go
	Payload:   SuccessPayload{Message: fmt.Sprintf("%s deleted successfully", req.ItemType)},
	Timestamp: time.Now().UnixMilli(),
}
return conn.SendMessage(response)
```

#### Función 5: getBaseUserProfile (Líneas 298-344) - **MIGRAR (AUXILIAR)**

**Acción requerida:** Cambiar firma para usar `appCtx`.

```go
// Línea 298: CAMBIAR firma
func (h *Hub) getBaseUserProfile(userID int64) (*models.User, error)
// ↓ CAMBIAR A:
func getBaseUserProfile(appCtx *AppContext, userID int64) (*models.User, error)

// Línea 317: CAMBIAR acceso a DB
err := h.DB.QueryRow(query, userID).Scan(...)
// ↓ CAMBIAR A:
err := appCtx.DB.QueryRow(query, userID).Scan(...)

// Resto de la lógica MANTENER (scan y asignación de NullString)
```

#### Función 6: getCurriculum (Líneas 346-401) - **MIGRAR (AUXILIAR)**

**Acción requerida:** Cambiar firma para usar `appCtx` y ajustar llamadas a funciones auxiliares de currículum.

```go
// Línea 347: CAMBIAR firma
func (h *Hub) getCurriculum(userID int64) (*Curriculum, error)
// ↓ CAMBIAR A:
func getCurriculum(appCtx *AppContext, userID int64) (*Curriculum, error)

// Línea 356: CAMBIAR llamada a DB dentro de la goroutine
data, fetchErr := fetchFunc(h.DB, userID)
// ↓ CAMBIAR A:
data, fetchErr := fetchFunc(appCtx.DB, userID) // Asumiendo que las funciones get* de curriculum se adaptan para recibir appCtx.DB

// Líneas 383-388: ACTUALIZAR llamadas a fetch
// Las funciones getEducation, getWorkExperience, etc. (de handlers_curriculum.go) deben aceptar (appCtx.DB, userID)
// O mejor, que acepten (appCtx *AppContext, userID int64) y accedan a appCtx.DB internamente.
// Ejemplo:
fetch(func(db *sql.DB, id int64) (interface{}, error) { return getEducation(db, id) }, &curriculum.Education)
// ↓ CAMBIAR A (si getEducation toma AppContext):
fetch(func(ctx *AppContext, id int64) (interface{}, error) { return getEducation(ctx, id) }, &curriculum.Education)
// O (si getEducation toma *sql.DB, pasar appCtx.DB):
fetch(func(db *sql.DB, id int64) (interface{}, error) { return getEducation(db, id) }, &curriculum.Education)
// Este último es el patrón actual de handlers_curriculum.go, por lo que solo el `h.DB` en la línea 356 necesita cambiar a `appCtx.DB`.

// Resto de la lógica MANTENER (sync.WaitGroup, error handling, type assertions)
```

**Líneas 403-410: Stubs de actualización de currículum - ELIMINAR COMENTADOS**

### Cambios Críticos Resumidos para handlers_profile.go:

1.  **Imports:** AGREGAR `customws_types`, `time`. ELIMINAR `encoding/json` donde no se parseen payloads directamente (ahora lo hace customws).
2.  **Structs locales:** Mover `DeleteCurriculumItemRequest` a `types.go` (posiblemente como `DeleteItemPayload`).
3.  **Firmas de Handlers:** Actualizar todas las firmas a `func handlerName(appCtx *AppContext, conn *customws.Connection[AppUserData], msg customws_types.ClientToServerMessage) error`.
4.  **Firmas Auxiliares:** `getBaseUserProfile` y `getCurriculum` deben tomar `appCtx *AppContext`.
5.  **Acceso a DB:** Cambiar `h.DB` a `appCtx.DB`.
6.  **Payload Parsing:** Usar `json.Unmarshal(msg.Payload, &reqStruct)`.
7.  **User Info:** `client.User.ID` → `conn.UserData.ID`, `client.User.RoleID` → `conn.UserData.RoleID`.
8.  **Respuestas:** Usar `conn.SendErrorNotification()` y `conn.SendMessage()` con `customws_types.ServerToClientMessage`.
9.  **Nuevos Tipos de Mensaje:** `MessageTypeEditProfileResponse`, `MessageTypeDeleteItemCurriculumResponse`.
10. **Estado Online:** `h.GetClient()` → `appCtx.ConnectionManager.IsUserConnected()`.
11. **Llamadas a `getContactStatus` y `getCurriculum`:** Actualizar para usar `appCtx`.
12. **Actualización de `client.User`:** Eliminar la lógica de `handleEditProfile` que actualizaba `client.User`.

### Esquema de BD utilizado por handlers_profile.go:
```sql
User (Id, FirstName, LastName, UserName, Email, Phone, Sex, DocId, NationalityId, Birthdate, Picture, DegreeId, UniversityId, RoleId, StatusAuthorizedId, Summary, Address, Github, Linkedin)
Nationality (Id, CountryName)
Degree (Id, DegreeName)
University (Id, Name)
Role (Id, Name)
-- Más tablas de Curriculum (Education, WorkExperience, etc. - llamadas vía getCurriculum)
Contact (User1Id, User2Id, Status) -- Usado por getContactStatus
```

### 7. `handlers_search.go` (Prioridad: MEDIA)

**Estado:** Migración completa requerida - Contiene handlers para obtener y editar perfiles, y eliminar items del currículum, además de funciones auxiliares clave.

#### Análisis línea por línea:

**Líneas 1-10: Package e imports**
```go
// CAMBIAR imports
package websocket

import (
	"database/sql"      // MANTENER
	"encoding/json"     // ELIMINAR - no necesario con customws
	"fmt"               // MANTENER
	"sync"              // MANTENER (usado en getCurriculum)

	"github.com/davidM20/micro-service-backend-go.git/internal/models" // MANTENER
	"github.com/davidM20/micro-service-backend-go.git/pkg/logger"      // MANTENER
	customws_types "github.com/davidM20/micro-service-backend-go.git/pkg/customws/types" // AGREGAR
	"time"                                                               // AGREGAR
)
```

**Líneas 12-26: Structs locales**
- `ProfileResponse` (Líneas 15-20): **ELIMINAR COMENTADO** - Ya se indica que se usa `MyProfileResponse` de `types.go`.
- `DeleteCurriculumItemRequest` (Líneas 23-26): **MANTENER** - Es usado por `handleDeleteItemCurriculum`. Podría moverse a `types.go` como `DeleteItemPayload` si no existe ya con ese nombre.

#### Función 1: handleGetProfile (Líneas 31-168) - **MIGRAR COMPLETAMENTE**

**Acción requerida:** Migrar función principal para obtener perfiles de usuario (propio o de otros).

**Cambios específicos:**

```go
// Línea 31: CAMBIAR firma
func (h *Hub) handleGetProfile(client *Client, payload interface{}) error
// ↓ CAMBIAR A:
func handleGetProfile(appCtx *AppContext, conn *customws.Connection[AppUserData], msg customws_types.ClientToServerMessage) error

// Líneas 32-40: SIMPLIFICAR parsing payload
var req ProfileRequestPayload
if msg.Payload != nil {
	if err := json.Unmarshal(msg.Payload, &req); err != nil {
		logger.Infof("WS", "handleGetProfile Info (UserID: %d): Payload is not a ProfileRequestPayload or is empty: %v", conn.UserData.ID, err)
		// Continuar, targetUserID será el propio usuario si req está vacía o malformada
	}
}

// Líneas 42-45: ACTUALIZAR lógica de targetUserID
targetUserID := client.User.ID
if req.TargetUserID != 0 {
	targetUserID = req.TargetUserID
}
// ↓ CAMBIAR A:
targetUserID := conn.UserData.ID
if req.TargetUserID != 0 {
	targetUserID = req.TargetUserID
}

// Línea 48: ACTUALIZAR llamada a función auxiliar
userBase, err := h.getBaseUserProfile(targetUserID)
// ↓ CAMBIAR A:
userBase, err := getBaseUserProfile(appCtx, targetUserID)

// Líneas 49-55: ACTUALIZAR manejo de errores
if err == sql.ErrNoRows {
	return h.sendErrorMessage(client, "Requested user profile not found")
}
if err != nil {
	return h.sendErrorMessage(client, "Failed to retrieve profile data")
}
// ↓ CAMBIAR A:
if err == sql.ErrNoRows {
	conn.SendErrorNotification(msg.PID, 404, "Requested user profile not found")
	return fmt.Errorf("user profile not found: %d", targetUserID)
}
if err != nil {
	conn.SendErrorNotification(msg.PID, 500, "Failed to retrieve profile data")
	return err
}

// Líneas 57-107: MANTENER lógica de mapeo a MyProfileResponse (ya es correcta)

// Línea 110: ACTUALIZAR obtención de estado online
_, isOnline := h.GetClient(targetUserID)
// ↓ CAMBIAR A:
isOnline := appCtx.ConnectionManager.IsUserConnected(targetUserID)

// Líneas 112-125: ACTUALIZAR verificación de permisos
if targetUserID != client.User.ID {
	isContact, contactStatus, err := h.getContactStatus(client.User.ID, targetUserID)
	// ... lógica de error y admin check ...
	return h.sendErrorMessage(client, "Permission denied to view this profile")
}
// ↓ CAMBIAR A:
if targetUserID != conn.UserData.ID {
	isContact, contactStatus, err := getContactStatus(appCtx, conn.UserData.ID, targetUserID)
	if err != nil {
		conn.SendErrorNotification(msg.PID, 500, "Failed to check contact status")
		return err
	}
	isAdmin := conn.UserData.RoleID == 7 || conn.UserData.RoleID == 8 // Asumiendo que RoleID está en AppUserData
	if !isAdmin && !isContact { // y que getContactStatus devuelve `isContact` correctamente
		conn.SendErrorNotification(msg.PID, 403, "Permission denied to view this profile")
		return fmt.Errorf("permission denied to view profile %d", targetUserID)
	}
}

// Línea 128: ACTUALIZAR llamada a función auxiliar
curriculum, err := h.getCurriculum(targetUserID)
// ↓ CAMBIAR A:
curriculum, err := getCurriculum(appCtx, targetUserID) // getCurriculum necesita ser migrada también

// Líneas 135-138: MANTENER asignación de currículum (profileResp.Curriculum = *curriculum)

// Líneas 140-152: ACTUALIZAR envío de respuesta
outgoingMsg := OutgoingMessage{Type: MessageTypeGetProfileResponse, Payload: profileResp}
jsonMsg, err := json.Marshal(outgoingMsg)
client.SendToMe(jsonMsg)
// ↓ CAMBIAR A:
response := customws_types.ServerToClientMessage{
	PID:       msg.PID,
	Type:      MessageTypeGetProfileResponse, // Mantener la constante de types.go
	Payload:   profileResp,
	Timestamp: time.Now().UnixMilli(),
}
return conn.SendMessage(response)
```

#### Función 2: handleGetMyProfile (Líneas 170-175) - **MIGRAR COMPLETAMENTE**

**Acción requerida:** Simplificar y llamar al nuevo `handleGetProfile`.

```go
// Línea 170: CAMBIAR firma
func (h *Hub) handleGetMyProfile(client *Client) error
// ↓ CAMBIAR A:
func handleGetMyProfile(appCtx *AppContext, conn *customws.Connection[AppUserData], msg customws_types.ClientToServerMessage) error

// Línea 173: SIMPLIFICAR llamada
return h.handleGetProfile(client, nil) // Payload nil para indicar perfil propio
// ↓ CAMBIAR A:
// El payload de msg ya sería nil o un ProfileRequestPayload vacío por defecto.
// Si se quiere forzar que sea el perfil propio, se puede pasar un ProfileRequestPayload con TargetUserID = 0 o conn.UserData.ID
// O simplemente llamar a handleGetProfile, que ya tiene la lógica para deducir el targetUserID.
return handleGetProfile(appCtx, conn, msg)
```

#### Función 3: handleEditProfile (Líneas 177-228) - **MIGRAR COMPLETAMENTE**

**Acción requerida:** Migrar función para actualizar perfil de usuario.

```go
// Línea 177: CAMBIAR firma
func (h *Hub) handleEditProfile(client *Client, payload interface{}) error
// ↓ CAMBIAR A:
func handleEditProfile(appCtx *AppContext, conn *customws.Connection[AppUserData], msg customws_types.ClientToServerMessage) error

// Líneas 181-194: SIMPLIFICAR parsing payload
var updateReq models.User // Usar `EditProfilePayload` de `types.go` en su lugar
// ... parsing manual ...
// ↓ CAMBIAR A:
var req EditProfilePayload // Usar el tipo de types.go
if err := json.Unmarshal(msg.Payload, &req); err != nil {
	conn.SendErrorNotification(msg.PID, 400, "Invalid update payload format")
	return fmt.Errorf("invalid payload: %w", err)
}

// Líneas 196-207: CAMBIAR acceso a DB y parámetros
_, err := h.DB.Exec(query, updateReq.FirstName, ..., userID)
// ↓ CAMBIAR A:
// Esta parte necesita una refactorización cuidadosa para construir la query y los args
// a partir de los campos no nulos de `req` (EditProfilePayload).
// Ejemplo conceptual:
// updates := []string{}
// args := []interface{}{}
// if req.Summary != nil { updates = append(updates, "Summary = ?"); args = append(args, *req.Summary) }
// ... etc. para todos los campos ...
// if len(updates) == 0 { /* no hay nada que actualizar */ return ... }
// query := fmt.Sprintf("UPDATE User SET %s WHERE Id = ?", strings.Join(updates, ", "))
// args = append(args, conn.UserData.ID)
// _, err := appCtx.DB.Exec(query, args...)

// Líneas 209-212: ACTUALIZAR manejo de errores
if err != nil {
	return h.sendErrorMessage(client, "Failed to update profile information")
}
// ↓ CAMBIAR A:
if err != nil {
	conn.SendErrorNotification(msg.PID, 500, "Failed to update profile information")
	return err
}

// Líneas 216-222: ACTUALIZAR `client.User` (ya no existirá)
// Esta lógica de actualizar el `client.User` en memoria ya no es necesaria con `customws`,
// ya que `AppUserData` es más estático o se recarga con cada conexión.
// Si se necesita refrescar `conn.UserData`, se debería hacer explícitamente.
// ELIMINAR esta sección.

// Línea 224: ACTUALIZAR respuesta de éxito
return h.sendSuccessMessage(client, "Profile updated successfully")
// ↓ CAMBIAR A:
response := customws_types.ServerToClientMessage{
	PID:       msg.PID,
	Type:      MessageTypeEditProfileResponse, // AGREGAR constante en types.go
	Payload:   SuccessPayload{Message: "Profile updated successfully"},
	Timestamp: time.Now().UnixMilli(),
}
return conn.SendMessage(response)
```

#### Función 4: handleDeleteItemCurriculum (Líneas 230-293) - **MIGRAR COMPLETAMENTE**

**Acción requerida:** Migrar función para eliminar items del currículum.

```go
// Línea 230: CAMBIAR firma
func (h *Hub) handleDeleteItemCurriculum(client *Client, payload interface{}) error
// ↓ CAMBIAR A:
func handleDeleteItemCurriculum(appCtx *AppContext, conn *customws.Connection[AppUserData], msg customws_types.ClientToServerMessage) error

// Líneas 234-247: SIMPLIFICAR parsing payload
var req DeleteCurriculumItemRequest // Usar `DeleteItemPayload` de `types.go`
// ... parsing manual ...
// ↓ CAMBIAR A:
var req DeleteItemPayload // Usar el tipo de types.go
if err := json.Unmarshal(msg.Payload, &req); err != nil {
	conn.SendErrorNotification(msg.PID, 400, "Invalid delete payload format")
	return fmt.Errorf("invalid payload: %w", err)
}

// Líneas 249-252: MANTENER validación (req.ItemType y req.ItemID)
// Ajustar el mensaje de error para `conn.SendErrorNotification`

// Líneas 254-274: MANTENER lógica de switch para `tableName` y `personIdColumn`

// Línea 276: CAMBIAR acceso a DB
result, err := h.DB.Exec(query, req.ItemID, userID)
// ↓ CAMBIAR A:
result, err := appCtx.DB.Exec(query, req.ItemID, conn.UserData.ID)

// Líneas 277-286: ACTUALIZAR manejo de errores y rowsAffected
// Usar `conn.SendErrorNotification`

// Línea 290: ACTUALIZAR respuesta de éxito
return h.sendSuccessMessage(client, fmt.Sprintf("%s deleted successfully", req.ItemType))
// ↓ CAMBIAR A:
response := customws_types.ServerToClientMessage{
	PID:       msg.PID,
	Type:      MessageTypeDeleteItemCurriculumResponse, // AGREGAR constante en types.go
	Payload:   SuccessPayload{Message: fmt.Sprintf("%s deleted successfully", req.ItemType)},
	Timestamp: time.Now().UnixMilli(),
}
return conn.SendMessage(response)
```

#### Función 5: getBaseUserProfile (Líneas 298-344) - **MIGRAR (AUXILIAR)**

**Acción requerida:** Cambiar firma para usar `appCtx`.

```go
// Línea 298: CAMBIAR firma
func (h *Hub) getBaseUserProfile(userID int64) (*models.User, error)
// ↓ CAMBIAR A:
func getBaseUserProfile(appCtx *AppContext, userID int64) (*models.User, error)

// Línea 317: CAMBIAR acceso a DB
err := h.DB.QueryRow(query, userID).Scan(...)
// ↓ CAMBIAR A:
err := appCtx.DB.QueryRow(query, userID).Scan(...)

// Resto de la lógica MANTENER (scan y asignación de NullString)
```

#### Función 6: getCurriculum (Líneas 346-401) - **MIGRAR (AUXILIAR)**

**Acción requerida:** Cambiar firma para usar `appCtx` y ajustar llamadas a funciones auxiliares de currículum.

```go
// Línea 347: CAMBIAR firma
func (h *Hub) getCurriculum(userID int64) (*Curriculum, error)
// ↓ CAMBIAR A:
func getCurriculum(appCtx *AppContext, userID int64) (*Curriculum, error)

// Línea 356: CAMBIAR llamada a DB dentro de la goroutine
data, fetchErr := fetchFunc(h.DB, userID)
// ↓ CAMBIAR A:
data, fetchErr := fetchFunc(appCtx.DB, userID) // Asumiendo que las funciones get* de curriculum se adaptan para recibir appCtx.DB

// Líneas 383-388: ACTUALIZAR llamadas a fetch
// Las funciones getEducation, getWorkExperience, etc. (de handlers_curriculum.go) deben aceptar (appCtx.DB, userID)
// O mejor, que acepten (appCtx *AppContext, userID int64) y accedan a appCtx.DB internamente.
// Ejemplo:
fetch(func(db *sql.DB, id int64) (interface{}, error) { return getEducation(db, id) }, &curriculum.Education)
// ↓ CAMBIAR A (si getEducation toma AppContext):
fetch(func(ctx *AppContext, id int64) (interface{}, error) { return getEducation(ctx, id) }, &curriculum.Education)
// O (si getEducation toma *sql.DB, pasar appCtx.DB):
fetch(func(db *sql.DB, id int64) (interface{}, error) { return getEducation(db, id) }, &curriculum.Education)
// Este último es el patrón actual de handlers_curriculum.go, por lo que solo el `h.DB` en la línea 356 necesita cambiar a `appCtx.DB`.

// Resto de la lógica MANTENER (sync.WaitGroup, error handling, type assertions)
```

**Líneas 403-410: Stubs de actualización de currículum - ELIMINAR COMENTADOS**

### Cambios Críticos Resumidos para handlers_profile.go:

1.  **Imports:** AGREGAR `customws_types`, `time`. ELIMINAR `encoding/json` donde no se parseen payloads directamente (ahora lo hace customws).
2.  **Structs locales:** Mover `DeleteCurriculumItemRequest` a `types.go` (posiblemente como `DeleteItemPayload`).
3.  **Firmas de Handlers:** Actualizar todas las firmas a `func handlerName(appCtx *AppContext, conn *customws.Connection[AppUserData], msg customws_types.ClientToServerMessage) error`.
4.  **Firmas Auxiliares:** `getBaseUserProfile` y `getCurriculum` deben tomar `appCtx *AppContext`.
5.  **Acceso a DB:** Cambiar `h.DB` a `appCtx.DB`.
6.  **Payload Parsing:** Usar `json.Unmarshal(msg.Payload, &reqStruct)`.
7.  **User Info:** `client.User.ID` → `conn.UserData.ID`, `client.User.RoleID` → `conn.UserData.RoleID`.
8.  **Respuestas:** Usar `conn.SendErrorNotification()` y `conn.SendMessage()` con `customws_types.ServerToClientMessage`.
9.  **Nuevos Tipos de Mensaje:** `MessageTypeEditProfileResponse`, `MessageTypeDeleteItemCurriculumResponse`.
10. **Estado Online:** `h.GetClient()` → `appCtx.ConnectionManager.IsUserConnected()`.
11. **Llamadas a `getContactStatus` y `getCurriculum`:** Actualizar para usar `appCtx`.
12. **Actualización de `client.User`:** Eliminar la lógica de `handleEditProfile` que actualizaba `client.User`.

### Esquema de BD utilizado por handlers_profile.go:
```sql
User (Id, FirstName, LastName, UserName, Email, Phone, Sex, DocId, NationalityId, Birthdate, Picture, DegreeId, UniversityId, RoleId, StatusAuthorizedId, Summary, Address, Github, Linkedin)
Nationality (Id, CountryName)
Degree (Id, DegreeName)
University (Id, Name)
Role (Id, Name)
-- Más tablas de Curriculum (Education, WorkExperience, etc. - llamadas vía getCurriculum)
Contact (User1Id, User2Id, Status) -- Usado por getContactStatus
```

### 7. `handlers_search.go` (Prioridad: MEDIA)

**Estado:** Migración completa requerida - Contiene handlers para obtener y editar perfiles, y eliminar items del currículum, además de funciones auxiliares clave.

#### Análisis línea por línea:

**Líneas 1-10: Package e imports**
```go
// CAMBIAR imports
package websocket

import (
	"database/sql"      // MANTENER
	"encoding/json"     // ELIMINAR - no necesario con customws
	"fmt"               // MANTENER
	"sync"              // MANTENER (usado en getCurriculum)

	"github.com/davidM20/micro-service-backend-go.git/internal/models" // MANTENER
	"github.com/davidM20/micro-service-backend-go.git/pkg/logger"      // MANTENER
	customws_types "github.com/davidM20/micro-service-backend-go.git/pkg/customws/types" // AGREGAR
	"time"                                                               // AGREGAR
)
```

**Líneas 12-26: Structs locales**
- `ProfileResponse` (Líneas 15-20): **ELIMINAR COMENTADO** - Ya se indica que se usa `MyProfileResponse` de `types.go`.
- `DeleteCurriculumItemRequest` (Líneas 23-26): **MANTENER** - Es usado por `handleDeleteItemCurriculum`. Podría moverse a `types.go` como `DeleteItemPayload` si no existe ya con ese nombre.

#### Función 1: handleGetProfile (Líneas 31-168) - **MIGRAR COMPLETAMENTE**

**Acción requerida:** Migrar función principal para obtener perfiles de usuario (propio o de otros).

**Cambios específicos:**

```go
// Línea 31: CAMBIAR firma
func (h *Hub) handleGetProfile(client *Client, payload interface{}) error
// ↓ CAMBIAR A:
func handleGetProfile(appCtx *AppContext, conn *customws.Connection[AppUserData], msg customws_types.ClientToServerMessage) error

// Líneas 32-40: SIMPLIFICAR parsing payload
var req ProfileRequestPayload
if msg.Payload != nil {
	if err := json.Unmarshal(msg.Payload, &req); err != nil {
		logger.Infof("WS", "handleGetProfile Info (UserID: %d): Payload is not a ProfileRequestPayload or is empty: %v", conn.UserData.ID, err)
		// Continuar, targetUserID será el propio usuario si req está vacía o malformada
	}
}

// Líneas 42-45: ACTUALIZAR lógica de targetUserID
targetUserID := client.User.ID
if req.TargetUserID != 0 {
	targetUserID = req.TargetUserID
}
// ↓ CAMBIAR A:
targetUserID := conn.UserData.ID
if req.TargetUserID != 0 {
	targetUserID = req.TargetUserID
}

// Línea 48: ACTUALIZAR llamada a función auxiliar
userBase, err := h.getBaseUserProfile(targetUserID)
// ↓ CAMBIAR A:
userBase, err := getBaseUserProfile(appCtx, targetUserID)

// Líneas 49-55: ACTUALIZAR manejo de errores
if err == sql.ErrNoRows {
	return h.sendErrorMessage(client, "Requested user profile not found")
}
if err != nil {
	return h.sendErrorMessage(client, "Failed to retrieve profile data")
}
// ↓ CAMBIAR A:
if err == sql.ErrNoRows {
	conn.SendErrorNotification(msg.PID, 404, "Requested user profile not found")
	return fmt.Errorf("user profile not found: %d", targetUserID)
}
if err != nil {
	conn.SendErrorNotification(msg.PID, 500, "Failed to retrieve profile data")
	return err
}

// Líneas 57-107: MANTENER lógica de mapeo a MyProfileResponse (ya es correcta)

// Línea 110: ACTUALIZAR obtención de estado online
_, isOnline := h.GetClient(targetUserID)
// ↓ CAMBIAR A:
isOnline := appCtx.ConnectionManager.IsUserConnected(targetUserID)

// Líneas 112-125: ACTUALIZAR verificación de permisos
if targetUserID != client.User.ID {
	isContact, contactStatus, err := h.getContactStatus(client.User.ID, targetUserID)
	// ... lógica de error y admin check ...
	return h.sendErrorMessage(client, "Permission denied to view this profile")
}
// ↓ CAMBIAR A:
if targetUserID != conn.UserData.ID {
	isContact, contactStatus, err := getContactStatus(appCtx, conn.UserData.ID, targetUserID)
	if err != nil {
		conn.SendErrorNotification(msg.PID, 500, "Failed to check contact status")
		return err
	}
	isAdmin := conn.UserData.RoleID == 7 || conn.UserData.RoleID == 8 // Asumiendo que RoleID está en AppUserData
	if !isAdmin && !isContact { // y que getContactStatus devuelve `isContact` correctamente
		conn.SendErrorNotification(msg.PID, 403, "Permission denied to view this profile")
		return fmt.Errorf("permission denied to view profile %d", targetUserID)
	}
}

// Línea 128: ACTUALIZAR llamada a función auxiliar
curriculum, err := h.getCurriculum(targetUserID)
// ↓ CAMBIAR A:
curriculum, err := getCurriculum(appCtx, targetUserID) // getCurriculum necesita ser migrada también

// Líneas 135-138: MANTENER asignación de currículum (profileResp.Curriculum = *curriculum)

// Líneas 140-152: ACTUALIZAR envío de respuesta
outgoingMsg := OutgoingMessage{Type: MessageTypeGetProfileResponse, Payload: profileResp}
jsonMsg, err := json.Marshal(outgoingMsg)
client.SendToMe(jsonMsg)
// ↓ CAMBIAR A:
response := customws_types.ServerToClientMessage{
	PID:       msg.PID,
	Type:      MessageTypeGetProfileResponse, // Mantener la constante de types.go
	Payload:   profileResp,
	Timestamp: time.Now().UnixMilli(),
}
return conn.SendMessage(response)
```

#### Función 2: handleGetMyProfile (Líneas 170-175) - **MIGRAR COMPLETAMENTE**

**Acción requerida:** Simplificar y llamar al nuevo `handleGetProfile`.

```go
// Línea 170: CAMBIAR firma
func (h *Hub) handleGetMyProfile(client *Client) error
// ↓ CAMBIAR A:
func handleGetMyProfile(appCtx *AppContext, conn *customws.Connection[AppUserData], msg customws_types.ClientToServerMessage) error

// Línea 173: SIMPLIFICAR llamada
return h.handleGetProfile(client, nil) // Payload nil para indicar perfil propio
// ↓ CAMBIAR A:
// El payload de msg ya sería nil o un ProfileRequestPayload vacío por defecto.
// Si se quiere forzar que sea el perfil propio, se puede pasar un ProfileRequestPayload con TargetUserID = 0 o conn.UserData.ID
// O simplemente llamar a handleGetProfile, que ya tiene la lógica para deducir el targetUserID.
return handleGetProfile(appCtx, conn, msg)
```

#### Función 3: handleEditProfile (Líneas 177-228) - **MIGRAR COMPLETAMENTE**

**Acción requerida:** Migrar función para actualizar perfil de usuario.

```go
// Línea 177: CAMBIAR firma
func (h *Hub) handleEditProfile(client *Client, payload interface{}) error
// ↓ CAMBIAR A:
func handleEditProfile(appCtx *AppContext, conn *customws.Connection[AppUserData], msg customws_types.ClientToServerMessage) error

// Líneas 181-194: SIMPLIFICAR parsing payload
var updateReq models.User // Usar `EditProfilePayload` de `types.go` en su lugar
// ... parsing manual ...
// ↓ CAMBIAR A:
var req EditProfilePayload // Usar el tipo de types.go
if err := json.Unmarshal(msg.Payload, &req); err != nil {
	conn.SendErrorNotification(msg.PID, 400, "Invalid update payload format")
	return fmt.Errorf("invalid payload: %w", err)
}

// Líneas 196-207: CAMBIAR acceso a DB y parámetros
_, err := h.DB.Exec(query, updateReq.FirstName, ..., userID)
// ↓ CAMBIAR A:
// Esta parte necesita una refactorización cuidadosa para construir la query y los args
// a partir de los campos no nulos de `req` (EditProfilePayload).
// Ejemplo conceptual:
// updates := []string{}
// args := []interface{}{}
// if req.Summary != nil { updates = append(updates, "Summary = ?"); args = append(args, *req.Summary) }
// ... etc. para todos los campos ...
// if len(updates) == 0 { /* no hay nada que actualizar */ return ... }
// query := fmt.Sprintf("UPDATE User SET %s WHERE Id = ?", strings.Join(updates, ", "))
// args = append(args, conn.UserData.ID)
// _, err := appCtx.DB.Exec(query, args...)

// Líneas 209-212: ACTUALIZAR manejo de errores
if err != nil {
	return h.sendErrorMessage(client, "Failed to update profile information")
}
// ↓ CAMBIAR A:
if err != nil {
	conn.SendErrorNotification(msg.PID, 500, "Failed to update profile information")
	return err
}

// Líneas 216-222: ACTUALIZAR `client.User` (ya no existirá)
// Esta lógica de actualizar el `client.User` en memoria ya no es necesaria con `customws`,
// ya que `AppUserData` es más estático o se recarga con cada conexión.
// Si se necesita refrescar `conn.UserData`, se debería hacer explícitamente.
// ELIMINAR esta sección.

// Línea 224: ACTUALIZAR respuesta de éxito
return h.sendSuccessMessage(client, "Profile updated successfully")
// ↓ CAMBIAR A:
response := customws_types.ServerToClientMessage{
	PID:       msg.PID,
	Type:      MessageTypeEditProfileResponse, // AGREGAR constante en types.go
	Payload:   SuccessPayload{Message: "Profile updated successfully"},
	Timestamp: time.Now().UnixMilli(),
}
return conn.SendMessage(response)
```

#### Función 4: handleDeleteItemCurriculum (Líneas 230-293) - **MIGRAR COMPLETAMENTE**

**Acción requerida:** Migrar función para eliminar items del currículum.

```go
// Línea 230: CAMBIAR firma
func (h *Hub) handleDeleteItemCurriculum(client *Client, payload interface{}) error
// ↓ CAMBIAR A:
func handleDeleteItemCurriculum(appCtx *AppContext, conn *customws.Connection[AppUserData], msg customws_types.ClientToServerMessage) error

// Líneas 234-247: SIMPLIFICAR parsing payload
var req DeleteCurriculumItemRequest // Usar `DeleteItemPayload` de `types.go`
// ... parsing manual ...
// ↓ CAMBIAR A:
var req DeleteItemPayload // Usar el tipo de types.go
if err := json.Unmarshal(msg.Payload, &req); err != nil {
	conn.SendErrorNotification(msg.PID, 400, "Invalid delete payload format")
	return fmt.Errorf("invalid payload: %w", err)
}

// Líneas 249-252: MANTENER validación (req.ItemType y req.ItemID)
// Ajustar el mensaje de error para `conn.SendErrorNotification`

// Líneas 254-274: MANTENER lógica de switch para `tableName` y `personIdColumn`

// Línea 276: CAMBIAR acceso a DB
result, err := h.DB.Exec(query, req.ItemID, userID)
// ↓ CAMBIAR A:
result, err := appCtx.DB.Exec(query, req.ItemID, conn.UserData.ID)

// Líneas 277-286: ACTUALIZAR manejo de errores y rowsAffected
// Usar `conn.SendErrorNotification`

// Línea 290: ACTUALIZAR respuesta de éxito
return h.sendSuccessMessage(client, fmt.Sprintf("%s deleted successfully", req.ItemType))
// ↓ CAMBIAR A:
response := customws_types.ServerToClientMessage{
	PID:       msg.PID,
	Type:      MessageTypeDeleteItemCurriculumResponse, // AGREGAR constante en types.go
	Payload:   SuccessPayload{Message: fmt.Sprintf("%s deleted successfully", req.ItemType)},
	Timestamp: time.Now().UnixMilli(),
}
return conn.SendMessage(response)
```

#### Función 5: getBaseUserProfile (Líneas 298-344) - **MIGRAR (AUXILIAR)**

**Acción requerida:** Cambiar firma para usar `appCtx`.

```go
// Línea 298: CAMBIAR firma
func (h *Hub) getBaseUserProfile(userID int64) (*models.User, error)
// ↓ CAMBIAR A:
func getBaseUserProfile(appCtx *AppContext, userID int64) (*models.User, error)

// Línea 317: CAMBIAR acceso a DB
err := h.DB.QueryRow(query, userID).Scan(...)
// ↓ CAMBIAR A:
err := appCtx.DB.QueryRow(query, userID).Scan(...)

// Resto de la lógica MANTENER (scan y asignación de NullString)
```

#### Función 6: getCurriculum (Líneas 346-401) - **MIGRAR (AUXILIAR)**

**Acción requerida:** Cambiar firma para usar `appCtx` y ajustar llamadas a funciones auxiliares de currículum.

```go
// Línea 347: CAMBIAR firma
func (h *Hub) getCurriculum(userID int64) (*Curriculum, error)
// ↓ CAMBIAR A:
func getCurriculum(appCtx *AppContext, userID int64) (*Curriculum, error)

// Línea 356: CAMBIAR llamada a DB dentro de la goroutine
data, fetchErr := fetchFunc(h.DB, userID)
// ↓ CAMBIAR A:
data, fetchErr := fetchFunc(appCtx.DB, userID) // Asumiendo que las funciones get* de curriculum se adaptan para recibir appCtx.DB

// Líneas 383-388: ACTUALIZAR llamadas a fetch
// Las funciones getEducation, getWorkExperience, etc. (de handlers_curriculum.go) deben aceptar (appCtx.DB, userID)
// O mejor, que acepten (appCtx *AppContext, userID int64) y accedan a appCtx.DB internamente.
// Ejemplo:
fetch(func(db *sql.DB, id int64) (interface{}, error) { return getEducation(db, id) }, &curriculum.Education)
// ↓ CAMBIAR A (si getEducation toma AppContext):
fetch(func(ctx *AppContext, id int64) (interface{}, error) { return getEducation(ctx, id) }, &curriculum.Education)
// O (si getEducation toma *sql.DB, pasar appCtx.DB):
fetch(func(db *sql.DB, id int64) (interface{}, error) { return getEducation(db, id) }, &curriculum.Education)
// Este último es el patrón actual de handlers_curriculum.go, por lo que solo el `h.DB` en la línea 356 necesita cambiar a `appCtx.DB`.

// Resto de la lógica MANTENER (sync.WaitGroup, error handling, type assertions)
```

**Líneas 403-410: Stubs de actualización de currículum - ELIMINAR COMENTADOS**

### Cambios Críticos Resumidos para handlers_profile.go:

1.  **Imports:** AGREGAR `customws_types`, `time`. ELIMINAR `encoding/json` donde no se parseen payloads directamente (ahora lo hace customws).
2.  **Structs locales:** Mover `DeleteCurriculumItemRequest` a `types.go` (posiblemente como `DeleteItemPayload`).
3.  **Firmas de Handlers:** Actualizar todas las firmas a `func handlerName(appCtx *AppContext, conn *customws.Connection[AppUserData], msg customws_types.ClientToServerMessage) error`.
4.  **Firmas Auxiliares:** `getBaseUserProfile` y `getCurriculum` deben tomar `appCtx *AppContext`.
5.  **Acceso a DB:** Cambiar `h.DB` a `appCtx.DB`.
6.  **Payload Parsing:** Usar `json.Unmarshal(msg.Payload, &reqStruct)`.
7.  **User Info:** `client.User.ID` → `conn.UserData.ID`, `client.User.RoleID` → `conn.UserData.RoleID`.
8.  **Respuestas:** Usar `conn.SendErrorNotification()` y `conn.SendMessage()` con `customws_types.ServerToClientMessage`.
9.  **Nuevos Tipos de Mensaje:** `MessageTypeEditProfileResponse`, `MessageTypeDeleteItemCurriculumResponse`.
10. **Estado Online:** `h.GetClient()` → `appCtx.ConnectionManager.IsUserConnected()`.
11. **Llamadas a `getContactStatus` y `getCurriculum`:** Actualizar para usar `appCtx`.
12. **Actualización de `client.User`:** Eliminar la lógica de `handleEditProfile` que actualizaba `client.User`.

### Esquema de BD utilizado por handlers_profile.go:
```sql
User (Id, FirstName, LastName, UserName, Email, Phone, Sex, DocId, NationalityId, Birthdate, Picture, DegreeId, UniversityId, RoleId, StatusAuthorizedId, Summary, Address, Github, Linkedin)
Nationality (Id, CountryName)
Degree (Id, DegreeName)
University (Id, Name)
Role (Id, Name)
-- Más tablas de Curriculum (Education, WorkExperience, etc. - llamadas vía getCurriculum)
Contact (User1Id, User2Id, Status) -- Usado por getContactStatus
```

### 7. `handlers_search.go` (Prioridad: MEDIA)

**Estado:** Migración completa requerida - Contiene handlers para obtener y editar perfiles, y eliminar items del currículum, además de funciones auxiliares clave.

#### Análisis línea por línea:

**Líneas 1-10: Package e imports**
```go
// CAMBIAR imports
package websocket

import (
	"database/sql"      // MANTENER
	"encoding/json"     // ELIMINAR - no necesario con customws
	"fmt"               // MANTENER
	"sync"              // MANTENER (usado en getCurriculum)

	"github.com/davidM20/micro-service-backend-go.git/internal/models" // MANTENER
	"github.com/davidM20/micro-service-backend-go.git/pkg/logger"      // MANTENER
	customws_types "github.com/davidM20/micro-service-backend-go.git/pkg/customws/types" // AGREGAR
	"time"                                                               // AGREGAR
)
```

**Líneas 12-26: Structs locales**
- `ProfileResponse` (Líneas 15-20): **ELIMINAR COMENTADO** - Ya se indica que se usa `MyProfileResponse` de `types.go`.
- `DeleteCurriculumItemRequest` (Líneas 23-26): **MANTENER** - Es usado por `handleDeleteItemCurriculum`. Podría moverse a `types.go` como `DeleteItemPayload` si no existe ya con ese nombre.

#### Función 1: handleGetProfile (Líneas 31-168) - **MIGRAR COMPLETAMENTE**

**Acción requerida:** Migrar función principal para obtener perfiles de usuario (propio o de otros).

**Cambios específicos:**

```go
// Línea 31: CAMBIAR firma
func (h *Hub) handleGetProfile(client *Client, payload interface{}) error
// ↓ CAMBIAR A:
func handleGetProfile(appCtx *AppContext, conn *customws.Connection[AppUserData], msg customws_types.ClientToServerMessage) error

// Líneas 32-40: SIMPLIFICAR parsing payload
var req ProfileRequestPayload
if msg.Payload != nil {
	if err := json.Unmarshal(msg.Payload, &req); err != nil {
		logger.Infof("WS", "handleGetProfile Info (UserID: %d): Payload is not a ProfileRequestPayload or is empty: %v", conn.UserData.ID, err)
		// Continuar, targetUserID será el propio usuario si req está vacía o malformada
	}
}

// Líneas 42-45: ACTUALIZAR lógica de targetUserID
targetUserID := client.User.ID
if req.TargetUserID != 0 {
	targetUserID = req.TargetUserID
}
// ↓ CAMBIAR A:
targetUserID := conn.UserData.ID
if req.TargetUserID != 0 {
	targetUserID = req.TargetUserID
}

// Línea 48: ACTUALIZAR llamada a función auxiliar
userBase, err := h.getBaseUserProfile(targetUserID)
// ↓ CAMBIAR A:
userBase, err := getBaseUserProfile(appCtx, targetUserID)

// Líneas 49-55: ACTUALIZAR manejo de errores
if err == sql.ErrNoRows {
	return h.sendErrorMessage(client, "Requested user profile not found")
}
if err != nil {
	return h.sendErrorMessage(client, "Failed to retrieve profile data")
}
// ↓ CAMBIAR A:
if err == sql.ErrNoRows {
	conn.SendErrorNotification(msg.PID, 404, "Requested user profile not found")
	return fmt.Errorf("user profile not found: %d", targetUserID)
}
if err != nil {
	conn.SendErrorNotification(msg.PID, 500, "Failed to retrieve profile data")
	return err
}

// Líneas 57-107: MANTENER lógica de mapeo a MyProfileResponse (ya es correcta)

// Línea 110: ACTUALIZAR obtención de estado online
_, isOnline := h.GetClient(targetUserID)
// ↓ CAMBIAR A:
isOnline := appCtx.ConnectionManager.IsUserConnected(targetUserID)

// Líneas 112-125: ACTUALIZAR verificación de permisos
if targetUserID != client.User.ID {
	isContact, contactStatus, err := h.getContactStatus(client.User.ID, targetUserID)
	// ... lógica de error y admin check ...
	return h.sendErrorMessage(client, "Permission denied to view this profile")
}
// ↓ CAMBIAR A:
if targetUserID != conn.UserData.ID {
	isContact, contactStatus, err := getContactStatus(appCtx, conn.UserData.ID, targetUserID)
	if err != nil {
		conn.SendErrorNotification(msg.PID, 500, "Failed to check contact status")
		return err
	}
	isAdmin := conn.UserData.RoleID == 7 || conn.UserData.RoleID == 8 // Asumiendo que RoleID está en AppUserData
	if !isAdmin && !isContact { // y que getContactStatus devuelve `isContact` correctamente
		conn.SendErrorNotification(msg.PID, 403, "Permission denied to view this profile")
		return fmt.Errorf("permission denied to view profile %d", targetUserID)
	}
}

// Línea 128: ACTUALIZAR llamada a función auxiliar
curriculum, err := h.getCurriculum(targetUserID)
// ↓ CAMBIAR A:
curriculum, err := getCurriculum(appCtx, targetUserID) // getCurriculum necesita ser migrada también

// Líneas 135-138: MANTENER asignación de currículum (profileResp.Curriculum = *curriculum)

// Líneas 140-152: ACTUALIZAR envío de respuesta
outgoingMsg := OutgoingMessage{Type: MessageTypeGetProfileResponse, Payload: profileResp}
jsonMsg, err := json.Marshal(outgoingMsg)
client.SendToMe(jsonMsg)
// ↓ CAMBIAR A:
response := customws_types.ServerToClientMessage{
	PID:       msg.PID,
	Type:      MessageTypeGetProfileResponse, // Mantener la constante de types.go
	Payload:   profileResp,
	Timestamp: time.Now().UnixMilli(),
}
return conn.SendMessage(response)
```

#### Función 2: handleGetMyProfile (Líneas 170-175) - **MIGRAR COMPLETAMENTE**

**Acción requerida:** Simplificar y llamar al nuevo `handleGetProfile`.

```go
// Línea 170: CAMBIAR firma
func (h *Hub) handleGetMyProfile(client *Client) error
// ↓ CAMBIAR A:
func handleGetMyProfile(appCtx *AppContext, conn *customws.Connection[AppUserData], msg customws_types.ClientToServerMessage) error

// Línea 173: SIMPLIFICAR llamada
return h.handleGetProfile(client, nil) // Payload nil para indicar perfil propio
// ↓ CAMBIAR A:
// El payload de msg ya sería nil o un ProfileRequestPayload vacío por defecto.
// Si se quiere forzar que sea el perfil propio, se puede pasar un ProfileRequestPayload con TargetUserID = 0 o conn.UserData.ID
// O simplemente llamar a handleGetProfile, que ya tiene la lógica para deducir el targetUserID.
return handleGetProfile(appCtx, conn, msg)
```

#### Función 3: handleEditProfile (Líneas 177-228) - **MIGRAR COMPLETAMENTE**

**Acción requerida:** Migrar función para actualizar perfil de usuario.

```go
// Línea 177: CAMBIAR firma
func (h *Hub) handleEditProfile(client *Client, payload interface{}) error
// ↓ CAMBIAR A:
func handleEditProfile(appCtx *AppContext, conn *customws.Connection[AppUserData], msg customws_types.ClientToServerMessage) error

// Líneas 181-194: SIMPLIFICAR parsing payload
var updateReq models.User // Usar `EditProfilePayload` de `types.go` en su lugar
// ... parsing manual ...
// ↓ CAMBIAR A:
var req EditProfilePayload // Usar el tipo de types.go
if err := json.Unmarshal(msg.Payload, &req); err != nil {
	conn.SendErrorNotification(msg.PID, 400, "Invalid update payload format")
	return fmt.Errorf("invalid payload: %w", err)
}

// Líneas 196-207: CAMBIAR acceso a DB y parámetros
_, err := h.DB.Exec(query, updateReq.FirstName, ..., userID)
// ↓ CAMBIAR A:
// Esta parte necesita una refactorización cuidadosa para construir la query y los args
// a partir de los campos no nulos de `req` (EditProfilePayload).
// Ejemplo conceptual:
// updates := []string{}
// args := []interface{}{}
// if req.Summary != nil { updates = append(updates, "Summary = ?"); args = append(args, *req.Summary) }
// ... etc. para todos los campos ...
// if len(updates) == 0 { /* no hay nada que actualizar */ return ... }
// query := fmt.Sprintf("UPDATE User SET %s WHERE Id = ?", strings.Join(updates, ", "))
// args = append(args, conn.UserData.ID)
// _, err := appCtx.DB.Exec(query, args...)

// Líneas 209-212: ACTUALIZAR manejo de errores
if err != nil {
	return h.sendErrorMessage(client, "Failed to update profile information")
}
// ↓ CAMBIAR A:
if err != nil {
	conn.SendErrorNotification(msg.PID, 500, "Failed to update profile information")
	return err
}

// Líneas 216-222: ACTUALIZAR `client.User` (ya no existirá)
// Esta lógica de actualizar el `client.User` en memoria ya no es necesaria con `customws`,
// ya que `AppUserData` es más estático o se recarga con cada conexión.
// Si se necesita refrescar `conn.UserData`, se debería hacer explícitamente.
// ELIMINAR esta sección.

// Línea 224: ACTUALIZAR respuesta de éxito
return h.sendSuccessMessage(client, "Profile updated successfully")
// ↓ CAMBIAR A:
response := customws_types.ServerToClientMessage{
	PID:       msg.PID,
	Type:      MessageTypeEditProfileResponse, // AGREGAR constante en types.go
	Payload:   SuccessPayload{Message: "Profile updated successfully"},
	Timestamp: time.Now().UnixMilli(),
}
return conn.SendMessage(response)
```

#### Función 4: handleDeleteItemCurriculum (Líneas 230-293) - **MIGRAR COMPLETAMENTE**

**Acción requerida:** Migrar función para eliminar items del currículum.

```go
// Línea 230: CAMBIAR firma
func (h *Hub) handleDeleteItemCurriculum(client *Client, payload interface{}) error
// ↓ CAMBIAR A:
func handleDeleteItemCurriculum(appCtx *AppContext, conn *customws.Connection[AppUserData], msg customws_types.ClientToServerMessage) error

// Líneas 234-247: SIMPLIFICAR parsing payload
var req DeleteCurriculumItemRequest // Usar `DeleteItemPayload` de `types.go`
// ... parsing manual ...
// ↓ CAMBIAR A:
var req DeleteItemPayload // Usar el tipo de types.go
if err := json.Unmarshal(msg.Payload, &req); err != nil {
	conn.SendErrorNotification(msg.PID, 400, "Invalid delete payload format")
	return fmt.Errorf("invalid payload: %w", err)
}

// Líneas 249-252: MANTENER validación (req.ItemType y req.ItemID)
// Ajustar el mensaje de error para `conn.SendErrorNotification`

// Líneas 254-274: MANTENER lógica de switch para `tableName` y `personIdColumn`

// Línea 276: CAMBIAR acceso a DB
result, err := h.DB.Exec(query, req.ItemID, userID)
// ↓ CAMBIAR A:
result, err := appCtx.DB.Exec(query, req.ItemID, conn.UserData.ID)

// Líneas 277-286: ACTUALIZAR manejo de errores y rowsAffected
// Usar `conn.SendErrorNotification`

// Línea 290: ACTUALIZAR respuesta de éxito
return h.sendSuccessMessage(client, fmt.Sprintf("%s deleted successfully", req.ItemType))
// ↓ CAMBIAR A:
response := customws_types.ServerToClientMessage{
	PID:       msg.PID,
	Type:      MessageTypeDeleteItemCurriculumResponse, // AGREGAR constante en types.go
	Payload:   SuccessPayload{Message: fmt.Sprintf("%s deleted successfully", req.ItemType)},
	Timestamp: time.Now().UnixMilli(),
}
return conn.SendMessage(response)
```

#### Función 5: getBaseUserProfile (Líneas 298-344) - **MIGRAR (AUXILIAR)**

**Acción requerida:** Cambiar firma para usar `appCtx`.

```go
// Línea 298: CAMBIAR firma
func (h *Hub) getBaseUserProfile(userID int64) (*models.User, error)
// ↓ CAMBIAR A:
func getBaseUserProfile(appCtx *AppContext, userID int64) (*models.User, error)

// Línea 317: CAMBIAR acceso a DB
err := h.DB.QueryRow(query, userID).Scan(...)
// ↓ CAMBIAR A:
err := appCtx.DB.QueryRow(query, userID).Scan(...)

// Resto de la lógica MANTENER (scan y asignación de NullString)
```

#### Función 6: getCurriculum (Líneas 346-401) - **MIGRAR (AUXILIAR)**

**Acción requerida:** Cambiar firma para usar `appCtx` y ajustar llamadas a funciones auxiliares de currículum.

```go
// Línea 347: CAMBIAR firma
func (h *Hub) getCurriculum(userID int64) (*Curriculum, error)
// ↓ CAMBIAR A:
func getCurriculum(appCtx *AppContext, userID int64) (*Curriculum, error)

// Línea 356: CAMBIAR llamada a DB dentro de la goroutine
data, fetchErr := fetchFunc(h.DB, userID)
// ↓ CAMBIAR A:
data, fetchErr := fetchFunc(appCtx.DB, userID) // Asumiendo que las funciones get* de curriculum se adaptan para recibir appCtx.DB

// Líneas 383-388: ACTUALIZAR llamadas a fetch
// Las funciones getEducation, getWorkExperience, etc. (de handlers_curriculum.go) deben aceptar (appCtx.DB, userID)
// O mejor, que acepten (appCtx *AppContext, userID int64) y accedan a appCtx.DB internamente.
// Ejemplo:
fetch(func(db *sql.DB, id int64) (interface{}, error) { return getEducation(db, id) }, &curriculum.Education)
// ↓ CAMBIAR A (si getEducation toma AppContext):
fetch(func(ctx *AppContext, id int64) (interface{}, error) { return getEducation(ctx, id) }, &curriculum.Education)
// O (si getEducation toma *sql.DB, pasar appCtx.DB):
fetch(func(db *sql.DB, id int64) (interface{}, error) { return getEducation(db, id) }, &curriculum.Education)
// Este último es el patrón actual de handlers_curriculum.go, por lo que solo el `h.DB` en la línea 356 necesita cambiar a `appCtx.DB`.

// Resto de la lógica MANTENER (sync.WaitGroup, error handling, type assertions)
```

**Líneas 403-410: Stubs de actualización de currículum - ELIMINAR COMENTADOS**

### Cambios Críticos Resumidos para handlers_profile.go:

1.  **Imports:** AGREGAR `customws_types`, `time`. ELIMINAR `encoding/json` donde no se parseen payloads directamente (ahora lo hace customws).
2.  **Structs locales:** Mover `DeleteCurriculumItemRequest` a `types.go` (posiblemente como `DeleteItemPayload`).
3.  **Firmas de Handlers:** Actualizar todas las firmas a `func handlerName(appCtx *AppContext, conn *customws.Connection[AppUserData], msg customws_types.ClientToServerMessage) error`.
4.  **Firmas Auxiliares:** `getBaseUserProfile` y `getCurriculum` deben tomar `appCtx *AppContext`.
5.  **Acceso a DB:** Cambiar `h.DB` a `appCtx.DB`.
6.  **Payload Parsing:** Usar `json.Unmarshal(msg.Payload, &reqStruct)`.
7.  **User Info:** `client.User.ID` → `conn.UserData.ID`, `client.User.RoleID` → `conn.UserData.RoleID`.
8.  **Respuestas:** Usar `conn.SendErrorNotification()` y `conn.SendMessage()` con `customws_types.ServerToClientMessage`.
9.  **Nuevos Tipos de Mensaje:** `MessageTypeEditProfileResponse`, `MessageTypeDeleteItemCurriculumResponse`.
10. **Estado Online:** `h.GetClient()` → `appCtx.ConnectionManager.IsUserConnected()`.
11. **Llamadas a `getContactStatus` y `getCurriculum`:** Actualizar para usar `appCtx`.
12. **Actualización de `client.User`:** Eliminar la lógica de `handleEditProfile` que actualizaba `client.User`.

### Esquema de BD utilizado por handlers_profile.go:
```sql
User (Id, FirstName, LastName, UserName, Email, Phone, Sex, DocId, NationalityId, Birthdate, Picture, DegreeId, UniversityId, RoleId, StatusAuthorizedId, Summary, Address, Github, Linkedin)
Nationality (Id, CountryName)
Degree (Id, DegreeName)
University (Id, Name)
Role (Id, Name)
-- Más tablas de Curriculum (Education, WorkExperience, etc. - llamadas vía getCurriculum)
Contact (User1Id, User2Id, Status) -- Usado por getContactStatus
```

### 7. `handlers_search.go` (Prioridad: MEDIA)

**Estado:** Migración completa requerida - Contiene handlers para obtener y editar perfiles, y eliminar items del currículum, además de funciones auxiliares clave.

#### Análisis línea por línea:

**Líneas 1-10: Package e imports**
```go
// CAMBIAR imports
package websocket

import (
	"database/sql"      // MANTENER
	"encoding/json"     // ELIMINAR - no necesario con customws
	"fmt"               // MANTENER
	"sync"              // MANTENER (usado en getCurriculum)

	"github.com/davidM20/micro-service-backend-go.git/internal/models" // MANTENER
	"github.com/davidM20/micro-service-backend-go.git/pkg/logger"      // MANTENER
	customws_types "github.com/davidM20/micro-service-backend-go.git/pkg/customws/types" // AGREGAR
	"time"                                                               // AGREGAR
)
```

**Líneas 12-26: Structs locales**
- `ProfileResponse` (Líneas 15-20): **ELIMINAR COMENTADO** - Ya se indica que se usa `MyProfileResponse` de `types.go`.
- `DeleteCurriculumItemRequest` (Líneas 23-26): **MANTENER** - Es usado por `handleDeleteItemCurriculum`. Podría moverse a `types.go` como `DeleteItemPayload` si no existe ya con ese nombre.

#### Función 1: handleGetProfile (Líneas 31-168) - **MIGRAR COMPLETAMENTE**

**Acción requerida:** Migrar función principal para obtener perfiles de usuario (propio o de otros).

**Cambios específicos:**

```go
// Línea 31: CAMBIAR firma
func (h *Hub) handleGetProfile(client *Client, payload interface{}) error
// ↓ CAMBIAR A:
func handleGetProfile(appCtx *AppContext, conn *customws.Connection[AppUserData], msg customws_types.ClientToServerMessage) error

// Líneas 32-40: SIMPLIFICAR parsing payload
var req ProfileRequestPayload
if msg.Payload != nil {
	if err := json.Unmarshal(msg.Payload, &req); err != nil {
		logger.Infof("WS", "handleGetProfile Info (UserID: %d): Payload is not a ProfileRequestPayload or is empty: %v", conn.UserData.ID, err)
		// Continuar, targetUserID será el propio usuario si req está vacía o malformada
	}
}

// Líneas 42-45: ACTUALIZAR lógica de targetUserID
targetUserID := client.User.ID
if req.TargetUserID != 0 {
	targetUserID = req.TargetUserID
}
// ↓ CAMBIAR A:
targetUserID := conn.UserData.ID
if req.TargetUserID != 0 {
	targetUserID = req.TargetUserID
}

// Línea 48: ACTUALIZAR llamada a función auxiliar
userBase, err := h.getBaseUserProfile(targetUserID)
// ↓ CAMBIAR A:
userBase, err := getBaseUserProfile(appCtx, targetUserID)

// Líneas 49-55: ACTUALIZAR manejo de errores
if err == sql.ErrNoRows {
	return h.sendErrorMessage(client, "Requested user profile not found")
}
if err != nil {
	return h.sendErrorMessage(client, "Failed to retrieve profile data")
}
// ↓ CAMBIAR A:
if err == sql.ErrNoRows {
	conn.SendErrorNotification(msg.PID, 404, "Requested user profile not found")
	return fmt.Errorf("user profile not found: %d", targetUserID)
}
if err != nil {
	conn.SendErrorNotification(msg.PID, 500, "Failed to retrieve profile data")
	return err
}

// Líneas 57-107: MANTENER lógica de mapeo a MyProfileResponse (ya es correcta)

// Línea 110: ACTUALIZAR obtención de estado online
_, isOnline := h.GetClient(targetUserID)
// ↓ CAMBIAR A:
isOnline := appCtx.ConnectionManager.IsUserConnected(targetUserID)

// Líneas 112-125: ACTUALIZAR verificación de permisos
if targetUserID != client.User.ID {
	isContact, contactStatus, err := h.getContactStatus(client.User.ID, targetUserID)
	// ... lógica de error y admin check ...
	return h.sendErrorMessage(client, "Permission denied to view this profile")
}
// ↓ CAMBIAR A:
if targetUserID != conn.UserData.ID {
	isContact, contactStatus, err := getContactStatus(appCtx, conn.UserData.ID, targetUserID)
	if err != nil {
		conn.SendErrorNotification(msg.PID, 500, "Failed to check contact status")
		return err
	}
	isAdmin := conn.UserData.RoleID == 7 || conn.UserData.RoleID == 8 // Asumiendo que RoleID está en AppUserData
	if !isAdmin && !isContact { // y que getContactStatus devuelve `isContact` correctamente
		conn.SendErrorNotification(msg.PID, 403, "Permission denied to view this profile")
		return fmt.Errorf("permission denied to view profile %d", targetUserID)
	}
}

// Línea 128: ACTUALIZAR llamada a función auxiliar
curriculum, err := h.getCurriculum(targetUserID)
// ↓ CAMBIAR A:
curriculum, err := getCurriculum(appCtx, targetUserID) // getCurriculum necesita ser migrada también

// Líneas 135-138: MANTENER asignación de currículum (profileResp.Curriculum = *curriculum)

// Líneas 140-152: ACTUALIZAR envío de respuesta
outgoingMsg := OutgoingMessage{Type: MessageTypeGetProfileResponse, Payload: profileResp}
jsonMsg, err := json.Marshal(outgoingMsg)
client.SendToMe(jsonMsg)
// ↓ CAMBIAR A:
response := customws_types.ServerToClientMessage{
	PID:       msg.PID,
	Type:      MessageTypeGetProfileResponse, // Mantener la constante de types.go
	Payload:   profileResp,
	Timestamp: time.Now().UnixMilli(),
}
return conn.SendMessage(response)
```

#### Función 2: handleGetMyProfile (Líneas 170-175) - **MIGRAR COMPLETAMENTE**

**Acción requerida:** Simplificar y llamar al nuevo `handleGetProfile`.

```go
// Línea 170: CAMBIAR firma
func (h *Hub) handleGetMyProfile(client *Client) error
// ↓ CAMBIAR A:
func handleGetMyProfile(appCtx *AppContext, conn *customws.Connection[AppUserData], msg customws_types.ClientToServerMessage) error

// Línea 173: SIMPLIFICAR llamada
return h.handleGetProfile(client, nil) // Payload nil para indicar perfil propio
// ↓ CAMBIAR A:
// El payload de msg ya sería nil o un ProfileRequestPayload vacío por defecto.
// Si se quiere forzar que sea el perfil propio, se puede pasar un ProfileRequestPayload con TargetUserID = 0 o conn.UserData.ID
// O simplemente llamar a handleGetProfile, que ya tiene la lógica para deducir el targetUserID.
return handleGetProfile(appCtx, conn, msg)
```

#### Función 3: handleEditProfile (Líneas 177-228) - **MIGRAR COMPLETAMENTE**

**Acción requerida:** Migrar función para actualizar perfil de usuario.

```go
// Línea 177: CAMBIAR firma
func (h *Hub) handleEditProfile(client *Client, payload interface{}) error
// ↓ CAMBIAR A:
func handleEditProfile(appCtx *AppContext, conn *customws.Connection[AppUserData], msg customws_types.ClientToServerMessage) error

// Líneas 181-194: SIMPLIFICAR parsing payload
var updateReq models.User // Usar `EditProfilePayload` de `types.go` en su lugar
// ... parsing manual ...
// ↓ CAMBIAR A:
var req EditProfilePayload // Usar el tipo de types.go
if err := json.Unmarshal(msg.Payload, &req); err != nil {
	conn.SendErrorNotification(msg.PID, 400, "Invalid update payload format")
	return fmt.Errorf("invalid payload: %w", err)
}

// Líneas 196-207: CAMBIAR acceso a DB y parámetros
_, err := h.DB.Exec(query, updateReq.FirstName, ..., userID)
// ↓ CAMBIAR A:
// Esta parte necesita una refactorización cuidadosa para construir la query y los args
// a partir de los campos no nulos de `req` (EditProfilePayload).
// Ejemplo conceptual:
// updates := []string{}
// args := []interface{}{}
// if req.Summary != nil { updates = append(updates, "Summary = ?"); args = append(args, *req.Summary) }
// ... etc. para todos los campos ...
// if len(updates) == 0 { /* no hay nada que actualizar */ return ... }
// query := fmt.Sprintf("UPDATE User SET %s WHERE Id = ?", strings.Join(updates, ", "))
// args = append(args, conn.UserData.ID)
// _, err := appCtx.DB.Exec(query, args...)

// Líneas 209-212: ACTUALIZAR manejo de errores
if err != nil {
	return h.sendErrorMessage(client, "Failed to update profile information")
}
// ↓ CAMBIAR A:
if err != nil {
	conn.SendErrorNotification(msg.PID, 500, "Failed to update profile information")
	return err
}

// Líneas 216-222: ACTUALIZAR `client.User` (ya no existirá)
// Esta lógica de actualizar el `client.User` en memoria ya no es necesaria con `customws`,
// ya que `AppUserData` es más estático o se recarga con cada conexión.
// Si se necesita refrescar `conn.UserData`, se debería hacer explícitamente.
// ELIMINAR esta sección.

// Línea 224: ACTUALIZAR respuesta de éxito
return h.sendSuccessMessage(client, "Profile updated successfully")
// ↓ CAMBIAR A:
response := customws_types.ServerToClientMessage{
	PID:       msg.PID,
	Type:      MessageTypeEditProfileResponse, // AGREGAR constante en types.go
	Payload:   SuccessPayload{Message: "Profile updated successfully"},
	Timestamp: time.Now().UnixMilli(),
}
return conn.SendMessage(response)
```

#### Función 4: handleDeleteItemCurriculum (Líneas 230-293) - **MIGRAR COMPLETAMENTE**

**Acción requerida:** Migrar función para eliminar items del currículum.

```go
// Línea 230: CAMBIAR firma
func (h *Hub) handleDeleteItemCurriculum(client *Client, payload interface{}) error
// ↓ CAMBIAR A:
func handleDeleteItemCurriculum(appCtx *AppContext, conn *customws.Connection[AppUserData], msg customws_types.ClientToServerMessage) error

// Líneas 234-247: SIMPLIFICAR parsing payload
var req DeleteCurriculumItemRequest // Usar `DeleteItemPayload` de `types.go`
// ... parsing manual ...
// ↓ CAMBIAR A:
var req DeleteItemPayload // Usar el tipo de types.go
if err := json.Unmarshal(msg.Payload, &req); err != nil {
	conn.SendErrorNotification(msg.PID, 400, "Invalid delete payload format")
	return fmt.Errorf("invalid payload: %w", err)
}

// Líneas 249-252: MANTENER validación (req.ItemType y req.ItemID)
// Ajustar el mensaje de error para `conn.SendErrorNotification`

// Líneas 254-274: MANTENER lógica de switch para `tableName` y `personIdColumn`

// Línea 276: CAMBIAR acceso a DB
result, err := h.DB.Exec(query, req.ItemID, userID)
// ↓ CAMBIAR A:
result, err := appCtx.DB.Exec(query, req.ItemID, conn.UserData.ID)

// Líneas 277-286: ACTUALIZAR manejo de errores y rowsAffected
// Usar `conn.SendErrorNotification`

// Línea 290: ACTUALIZAR respuesta de éxito
return h.sendSuccessMessage(client, fmt.Sprintf("%s deleted successfully", req.ItemType))
// ↓ CAMBIAR A:
response := customws_types.ServerToClientMessage{
	PID:       msg.PID,
	Type:      MessageTypeDeleteItemCurriculumResponse, // AGREGAR constante en types.go
	Payload:   SuccessPayload{Message: fmt.Sprintf("%s deleted successfully", req.ItemType)},
	Timestamp: time.Now().UnixMilli(),
}
return conn.SendMessage(response)
```

#### Función 5: getBaseUserProfile (Líneas 298-344) - **MIGRAR (AUXILIAR)**

**Acción requerida:** Cambiar firma para usar `appCtx`.

```go
// Línea 298: CAMBIAR firma
func (h *Hub) getBaseUserProfile(userID int64) (*models.User, error)
// ↓ CAMBIAR A:
func getBaseUserProfile(appCtx *AppContext, userID int64) (*models.User, error)

// Línea 317: CAMBIAR acceso a DB
err := h.DB.QueryRow(query, userID).Scan(...)
// ↓ CAMBIAR A:
err := appCtx.DB.QueryRow(query, userID).Scan(...)

// Resto de la lógica MANTENER (scan y asignación de NullString)
```

#### Función 6: getCurriculum (Líneas 346-401) - **MIGRAR (AUXILIAR)**

**Acción requerida:** Cambiar firma para usar `appCtx` y ajustar llamadas a funciones auxiliares de currículum.

```go
// Línea 347: CAMBIAR firma
func (h *Hub) getCurriculum(userID int64) (*Curriculum, error)
// ↓ CAMBIAR A:
func getCurriculum(appCtx *AppContext, userID int64) (*Curriculum, error)

// Línea 356: CAMBIAR llamada a DB dentro de la goroutine
data, fetchErr := fetchFunc(h.DB, userID)
// ↓ CAMBIAR A:
data, fetchErr := fetchFunc(appCtx.DB, userID) // Asumiendo que las funciones get* de curriculum se adaptan para recibir appCtx.DB

// Líneas 383-388: ACTUALIZAR llamadas a fetch
// Las funciones getEducation, getWorkExperience, etc. (de handlers_curriculum.go) deben aceptar (appCtx.DB, userID)
// O mejor, que acepten (appCtx *AppContext, userID int64) y accedan a appCtx.DB internamente.
// Ejemplo:
fetch(func(db *sql.DB, id int64) (interface{}, error) { return getEducation(db, id) }, &curriculum.Education)
// ↓ CAMBIAR A (si getEducation toma AppContext):
fetch(func(ctx *AppContext, id int64) (interface{}, error) { return getEducation(ctx, id) }, &curriculum.Education)
// O (si getEducation toma *sql.DB, pasar appCtx.DB):
fetch(func(db *sql.DB, id int64) (interface{}, error) { return getEducation(db, id) }, &curriculum.Education)
// Este último es el patrón actual de handlers_curriculum.go, por lo que solo el `h.DB` en la línea 356 necesita cambiar a `appCtx.DB`.

// Resto de la lógica MANTENER (sync.WaitGroup, error handling, type assertions)
```

**Líneas 403-410: Stubs de actualización de currículum - ELIMINAR COMENTADOS**

### Cambios Críticos Resumidos para handlers_profile.go:

1.  **Imports:** AGREGAR `customws_types`, `time`. ELIMINAR `encoding/json` donde no se parseen payloads directamente (ahora lo hace customws).
2.  **Structs locales:** Mover `DeleteCurriculumItemRequest` a `types.go` (posiblemente como `DeleteItemPayload`).
3.  **Firmas de Handlers:** Actualizar todas las firmas a `func handlerName(appCtx *AppContext, conn *customws.Connection[AppUserData], msg customws_types.ClientToServerMessage) error`.
4.  **Firmas Auxiliares:** `getBaseUserProfile` y `getCurriculum` deben tomar `appCtx *AppContext`.
5.  **Acceso a DB:** Cambiar `h.DB` a `appCtx.DB`.
6.  **Payload Parsing:** Usar `json.Unmarshal(msg.Payload, &reqStruct)`.
7.  **User Info:** `client.User.ID` → `conn.UserData.ID`, `client.User.RoleID` → `conn.UserData.RoleID`.
8.  **Respuestas:** Usar `conn.SendErrorNotification()` y `conn.SendMessage()` con `customws_types.ServerToClientMessage`.
9.  **Nuevos Tipos de Mensaje:** `MessageTypeEditProfileResponse`, `MessageTypeDeleteItemCurriculumResponse`.
10. **Estado Online:** `h.GetClient()` → `appCtx.ConnectionManager.IsUserConnected()`.
11. **Llamadas a `getContactStatus` y `getCurriculum`:** Actualizar para usar `appCtx`.
12. **Actualización de `client.User`:** Eliminar la lógica de `handleEditProfile` que actualizaba `client.User`.

### Esquema de BD utilizado por handlers_profile.go:
```sql
User (Id, FirstName, LastName, UserName, Email, Phone, Sex, DocId, NationalityId, Birthdate, Picture, DegreeId, UniversityId, RoleId, StatusAuthorizedId, Summary, Address, Github, Linkedin)
Nationality (Id, CountryName)
Degree (Id, DegreeName)
University (Id, Name)
Role (Id, Name)
-- Más tablas de Curriculum (Education, WorkExperience, etc. - llamadas vía getCurriculum)
Contact (User1Id, User2Id, Status) -- Usado por getContactStatus
```

### 7. `handlers_search.go` (Prioridad: MEDIA)

**Estado:** Migración completa requerida - Contiene handlers para obtener y editar perfiles, y eliminar items del currículum, además de funciones auxiliares clave.

#### Análisis línea por línea:

**Líneas 1-10: Package e imports**
```go
// CAMBIAR imports
package websocket

import (
	"database/sql"      // MANTENER
	"encoding/json"     // ELIMINAR - no necesario con customws
	"fmt"               // MANTENER
	"sync"              // MANTENER (usado en getCurriculum)

	"github.com/davidM20/micro-service-backend-go.git/internal/models" // MANTENER
	"github.com/davidM20/micro-service-backend-go.git/pkg/logger"      // MANTENER
	customws_types "github.com/davidM20/micro-service-backend-go.git/pkg/customws/types" // AGREGAR
	"time"                                                               // AGREGAR
)
```

**Líneas 12-26: Structs locales**
- `ProfileResponse` (Líneas 15-20): **ELIMINAR COMENTADO** - Ya se indica que se usa `MyProfileResponse` de `types.go`.
- `DeleteCurriculumItemRequest` (Líneas 23-26): **MANTENER** - Es usado por `handleDeleteItemCurriculum`. Podría moverse a `types.go` como `DeleteItemPayload` si no existe ya con ese nombre.

#### Función 1: handleGetProfile (Líneas 31-168) - **MIGRAR COMPLETAMENTE**

**Acción requerida:** Migrar función principal para obtener perfiles de usuario (propio o de otros).

**Cambios específicos:**

```go
// Línea 31: CAMBIAR firma
func (h *Hub) handleGetProfile(client *Client, payload interface{}) error
// ↓ CAMBIAR A:
func handleGetProfile(appCtx *AppContext, conn *customws.Connection[AppUserData], msg customws_types.ClientToServerMessage) error

// Líneas 32-40: SIMPLIFICAR parsing payload
var req ProfileRequestPayload
if msg.Payload != nil {
	if err := json.Unmarshal(msg.Payload, &req); err != nil {
		logger.Infof("WS", "handleGetProfile Info (UserID: %d): Payload is not a ProfileRequestPayload or is empty: %v", conn.UserData.ID, err)
		// Continuar, targetUserID será el propio usuario si req está vacía o malformada
	}
}

// Líneas 42-45: ACTUALIZAR lógica de targetUserID
targetUserID := client.User.ID
if req.TargetUserID != 0 {
	targetUserID = req.TargetUserID
}
// ↓ CAMBIAR A:
targetUserID := conn.UserData.ID
if req.TargetUserID != 0 {
	targetUserID = req.TargetUserID
}

// Línea 48: ACTUALIZAR llamada a función auxiliar
userBase, err := h.getBaseUserProfile(targetUserID)
// ↓ CAMBIAR A:
userBase, err := getBaseUserProfile(appCtx, targetUserID)

// Líneas 49-55: ACTUALIZAR manejo de errores
if err == sql.ErrNoRows {
	return h.sendErrorMessage(client, "Requested user profile not found")
}
if err != nil {
	return h.sendErrorMessage(client, "Failed to retrieve profile data")
}
// ↓ CAMBIAR A:
if err == sql.ErrNoRows {
	conn.SendErrorNotification(msg.PID, 404, "Requested user profile not found")
	return fmt.Errorf("user profile not found: %d", targetUserID)
}
if err != nil {
	conn.SendErrorNotification(msg.PID, 500, "Failed to retrieve profile data")
	return err
}

// Líneas 57-107: MANTENER lógica de mapeo a MyProfileResponse (ya es correcta)

// Línea 110: ACTUALIZAR obtención de estado online
_, isOnline := h.GetClient(targetUserID)
// ↓ CAMBIAR A:
isOnline := appCtx.ConnectionManager.IsUserConnected(targetUserID)

// Líneas 112-125: ACTUALIZAR verificación de permisos
if targetUserID != client.User.ID {
	isContact, contactStatus, err := h.getContactStatus(client.User.ID, targetUserID)
	// ... lógica de error y admin check ...
	return h.sendErrorMessage(client, "Permission denied to view this profile")
}
// ↓ CAMBIAR A:
if targetUserID != conn.UserData.ID {
	isContact, contactStatus, err := getContactStatus(appCtx, conn.UserData.ID, targetUserID)
	if err != nil {
		conn.SendErrorNotification(msg.PID, 500, "Failed to check contact status")
		return err
	}
	isAdmin := conn.UserData.RoleID == 7 || conn.UserData.RoleID == 8 // Asumiendo que RoleID está en AppUserData
	if !isAdmin && !isContact { // y que getContactStatus devuelve `isContact` correctamente
		conn.SendErrorNotification(msg.PID, 403, "Permission denied to view this profile")
		return fmt.Errorf("permission denied to view profile %d", targetUserID)
	}
}

// Línea 128: ACTUALIZAR llamada a función auxiliar
curriculum, err := h.getCurriculum(targetUserID)
// ↓ CAMBIAR A:
curriculum, err := getCurriculum(appCtx, targetUserID) // getCurriculum necesita ser migrada también

// Líneas 135-138: MANTENER asignación de currículum (profileResp.Curriculum = *curriculum)

// Líneas 140-152: ACTUALIZAR envío de respuesta
outgoingMsg := OutgoingMessage{Type: MessageTypeGetProfileResponse, Payload: profileResp}
jsonMsg, err := json.Marshal(outgoingMsg)
client.SendToMe(jsonMsg)
// ↓ CAMBIAR A:
response := customws_types.ServerToClientMessage{
	PID:       msg.PID,
	Type:      MessageTypeGetProfileResponse, // Mantener la constante de types.go
	Payload:   profileResp,
	Timestamp: time.Now().UnixMilli(),
}
return conn.SendMessage(response)
```

#### Función 2: handleGetMyProfile (Líneas 170-175) - **MIGRAR COMPLETAMENTE**

**Acción requerida:** Simplificar y llamar al nuevo `handleGetProfile`.

```go
// Línea 170: CAMBIAR firma
func (h *Hub) handleGetMyProfile(client *Client) error
// ↓ CAMBIAR A:
func handleGetMyProfile(appCtx *AppContext, conn *customws.Connection[AppUserData], msg customws_types.ClientToServerMessage) error

// Línea 173: SIMPLIFICAR llamada
return h.handleGetProfile(client, nil) // Payload nil para indicar perfil propio
// ↓ CAMBIAR A:
// El payload de msg ya sería nil o un ProfileRequestPayload vacío por defecto.
// Si se quiere forzar que sea el perfil propio, se puede pasar un ProfileRequestPayload con TargetUserID = 0 o conn.UserData.ID
// O simplemente llamar a handleGetProfile, que ya tiene la lógica para deducir el targetUserID.
return handleGetProfile(appCtx, conn, msg)
```

#### Función 3: handleEditProfile (Líneas 177-228) - **MIGRAR COMPLETAMENTE**

**Acción requerida:** Migrar función para actualizar perfil de usuario.

```go
// Línea 177: CAMBIAR firma
func (h *Hub) handleEditProfile(client *Client, payload interface{}) error
// ↓ CAMBIAR A:
func handleEditProfile(appCtx *AppContext, conn *customws.Connection[AppUserData], msg customws_types.ClientToServerMessage) error

// Líneas 181-194: SIMPLIFICAR parsing payload
var updateReq models.User // Usar `EditProfilePayload` de `types.go` en su lugar
// ... parsing manual ...
// ↓ CAMBIAR A:
var req EditProfilePayload // Usar el tipo de types.go
if err := json.Unmarshal(msg.Payload, &req); err != nil {
	conn.SendErrorNotification(msg.PID, 400, "Invalid update payload format")
	return fmt.Errorf("invalid payload: %w", err)
}

// Líneas 196-207: CAMBIAR acceso a DB y parámetros
_, err := h.DB.Exec(query, updateReq.FirstName, ..., userID)
// ↓ CAMBIAR A:
// Esta parte necesita una refactorización cuidadosa para construir la query y los args
// a partir de los campos no nulos de `req` (EditProfilePayload).
// Ejemplo conceptual:
// updates := []string{}
// args := []interface{}{}
// if req.Summary != nil { updates = append(updates, "Summary = ?"); args = append(args, *req.Summary) }
// ... etc. para todos los campos ...
// if len(updates) == 0 { /* no hay nada que actualizar */ return ... }
// query := fmt.Sprintf("UPDATE User SET %s WHERE Id = ?", strings.Join(updates, ", "))
// args = append(args, conn.UserData.ID)
// _, err := appCtx.DB.Exec(query, args...)

// Líneas 209-212: ACTUALIZAR manejo de errores
if err != nil {
	return h.sendErrorMessage(client, "Failed to update profile information")
}
// ↓ CAMBIAR A:
if err != nil {
	conn.SendErrorNotification(msg.PID, 500, "Failed to update profile information")
	return err
}

// Líneas 216-222: ACTUALIZAR `client.User` (ya no existirá)
// Esta lógica de actualizar el `client.User` en memoria ya no es necesaria con `customws`,
// ya que `AppUserData` es más estático o se recarga con cada conexión.
// Si se necesita refrescar `conn.UserData`, se debería hacer explícitamente.
// ELIMINAR esta sección.

// Línea 224: ACTUALIZAR respuesta de éxito
return h.sendSuccessMessage(client, "Profile updated successfully")
// ↓ CAMBIAR A:
response := customws_types.ServerToClientMessage{
	PID:       msg.PID,
	Type:      MessageTypeEditProfileResponse, // AGREGAR constante en types.go
	Payload:   SuccessPayload{Message: "Profile updated successfully"},
	Timestamp: time.Now().UnixMilli(),
}
return conn.SendMessage(response)
```

#### Función 4: handleDeleteItemCurriculum (Líneas 230-293) - **MIGRAR COMPLETAMENTE**

**Acción requerida:** Migrar función para eliminar items del currículum.

```go
// Línea 230: CAMBIAR firma
func (h *Hub) handleDeleteItemCurriculum(client *Client, payload interface{}) error
// ↓ CAMBIAR A:
func handleDeleteItemCurriculum(appCtx *AppContext, conn *customws.Connection[AppUserData], msg customws_types.ClientToServerMessage) error

// Líneas 234-247: SIMPLIFICAR parsing payload
var req DeleteCurriculumItemRequest // Usar `DeleteItemPayload` de `types.go`
// ... parsing manual ...
// ↓ CAMBIAR A:
var req DeleteItemPayload // Usar el tipo de types.go
if err := json.Unmarshal(msg.Payload, &req); err != nil {
	conn.SendErrorNotification(msg.PID, 400, "Invalid delete payload format")
	return fmt.Errorf("invalid payload: %w", err)
}

// Líneas 249-252: MANTENER validación (req.ItemType y req.ItemID)
// Ajustar el mensaje de error para `conn.SendErrorNotification`

// Líneas 254-274: MANTENER lógica de switch para `tableName` y `personIdColumn`

// Línea 276: CAMBIAR acceso a DB
result, err := h.DB.Exec(query, req.ItemID, userID)
// ↓ CAMBIAR A:
result, err := appCtx.DB.Exec(query, req.ItemID, conn.UserData.ID)

// Líneas 277-286: ACTUALIZAR manejo de errores y rowsAffected
// Usar `conn.SendErrorNotification`

// Línea 290: ACTUALIZAR respuesta de éxito
return h.sendSuccessMessage(client, fmt.Sprintf("%s deleted successfully", req.ItemType))
// ↓ CAMBIAR A:
response := customws_types.ServerToClientMessage{
	PID:       msg.PID,
	Type:      MessageTypeDeleteItemCurriculumResponse, // AGREGAR constante en types.go
	Payload:   SuccessPayload{Message: fmt.Sprintf("%s deleted successfully", req.ItemType)},
	Timestamp: time.Now().UnixMilli(),
}
return conn.SendMessage(response)
```

#### Función 5: getBaseUserProfile (Líneas 298-344) - **MIGRAR (AUXILIAR)**

**Acción requerida:** Cambiar firma para usar `appCtx`.

```go
// Línea 298: CAMBIAR firma
func (h *Hub) getBaseUserProfile(userID int64) (*models.User, error)
// ↓ CAMBIAR A:
func getBaseUserProfile(appCtx *AppContext, userID int64) (*models.User, error)

// Línea 317: CAMBIAR acceso a DB
err := h.DB.QueryRow(query, userID).Scan(...)
// ↓ CAMBIAR A:
err := appCtx.DB.QueryRow(query, userID).Scan(...)

// Resto de la lógica MANTENER (scan y asignación de NullString)
```

#### Función 6: getCurriculum (Líneas 346-401) - **MIGRAR (AUXILIAR)**

**Acción requerida:** Cambiar firma para usar `appCtx` y ajustar llamadas a funciones auxiliares de currículum.

```go
// Línea 347: CAMBIAR firma
func (h *Hub) getCurriculum(userID int64) (*Curriculum, error)
// ↓ CAMBIAR A:
func getCurriculum(appCtx *AppContext, userID int64) (*Curriculum, error)

// Línea 356: CAMBIAR llamada a DB dentro de la goroutine
data, fetchErr := fetchFunc(h.DB, userID)
// ↓ CAMBIAR A:
data, fetchErr := fetchFunc(appCtx.DB, userID) // Asumiendo que las funciones get* de curriculum se adaptan para recibir appCtx.DB

// Líneas 383-388: ACTUALIZAR llamadas a fetch
// Las funciones getEducation, getWorkExperience, etc. (de handlers_curriculum.go) deben aceptar (appCtx.DB, userID)
// O mejor, que acepten (appCtx *AppContext, userID int64) y accedan a appCtx.DB internamente.
// Ejemplo:
fetch(func(db *sql.DB, id int64) (interface{}, error) { return getEducation(db, id) }, &curriculum.Education)
// ↓ CAMBIAR A (si getEducation toma AppContext):
fetch(func(ctx *AppContext, id int64) (interface{}, error) { return getEducation(ctx, id) }, &curriculum.Education)
// O (si getEducation toma *sql.DB, pasar appCtx.DB):
fetch(func(db *sql.DB, id int64) (interface{}, error) { return getEducation(db, id) }, &curriculum.Education)
// Este último es el patrón actual de handlers_curriculum.go, por lo que solo el `h.DB` en la línea 356 necesita cambiar a `appCtx.DB`.

// Resto de la lógica MANTENER (sync.WaitGroup, error handling, type assertions)
```

**Líneas 403-410: Stubs de actualización de currículum - ELIMINAR COMENTADOS**

### Cambios Críticos Resumidos para handlers_profile.go:

1.  **Imports:** AGREGAR `customws_types`, `time`. ELIMINAR `encoding/json` donde no se parseen payloads directamente (ahora lo hace customws).
2.  **Structs locales:** Mover `DeleteCurriculumItemRequest` a `types.go` (posiblemente como `DeleteItemPayload`).
3.  **Firmas de Handlers:** Actualizar todas las firmas a `func handlerName(appCtx *AppContext, conn *customws.Connection[AppUserData], msg customws_types.ClientToServerMessage) error`.
4.  **Firmas Auxiliares:** `getBaseUserProfile` y `getCurriculum` deben tomar `appCtx *AppContext`.
5.  **Acceso a DB:** Cambiar `h.DB` a `appCtx.DB`.
6.  **Payload Parsing:** Usar `json.Unmarshal(msg.Payload, &reqStruct)`.
7.  **User Info:** `client.User.ID` → `conn.UserData.ID`, `client.User.RoleID` → `conn.UserData.RoleID`.
8.  **Respuestas:** Usar `conn.SendErrorNotification()` y `conn.SendMessage()` con `customws_types.ServerToClientMessage`.
9.  **Nuevos Tipos de Mensaje:** `MessageTypeEditProfileResponse`, `MessageTypeDeleteItemCurriculumResponse`.
10. **Estado Online:** `h.GetClient()` → `appCtx.ConnectionManager.IsUserConnected()`.
11. **Llamadas a `getContactStatus` y `getCurriculum`:** Actualizar para usar `appCtx`.
12. **Actualización de `client.User`:** Eliminar la lógica de `handleEditProfile` que actualizaba `client.User`.

### Esquema de BD utilizado por handlers_profile.go:
```sql
User (Id, FirstName, LastName, UserName, Email, Phone, Sex, DocId, NationalityId, Birthdate, Picture, DegreeId, UniversityId, RoleId, StatusAuthorizedId, Summary, Address, Github, Linkedin)
Nationality (Id, CountryName)
Degree (Id, DegreeName)
University (Id, Name)
Role (Id, Name)
-- Más tablas de Curriculum (Education, WorkExperience, etc. - llamadas vía getCurriculum)
Contact (User1Id, User2Id, Status) -- Usado por getContactStatus
```

### 7. `handlers_search.go` (Prioridad: MEDIA)

**Estado:** Migración completa requerida - Contiene handlers para obtener y editar perfiles, y eliminar items del currículum, además de funciones auxiliares clave.

#### Análisis línea por línea:

**Líneas 1-10: Package e imports**
```go
// CAMBIAR imports
package websocket

import (
	"database/sql"      // MANTENER
	"encoding/json"     // ELIMINAR - no necesario con customws
	"fmt"               // MANTENER
	"sync"              // MANTENER (usado en getCurriculum)

	"github.com/davidM20/micro-service-backend-go.git/internal/models" // MANTENER
	"github.com/davidM20/micro-service-backend-go.git/pkg/logger"      // MANTENER
	customws_types "github.com/davidM20/micro-service-backend-go.git/pkg/customws/types" // AGREGAR
	"time"                                                               // AGREGAR
)
```

**Líneas 12-26: Structs locales**
- `ProfileResponse` (Líneas 15-20): **ELIMINAR COMENTADO** - Ya se indica que se usa `MyProfileResponse` de `types.go`.
- `DeleteCurriculumItemRequest` (Líneas 23-26): **MANTENER** - Es usado por `handleDeleteItemCurriculum`. Podría moverse a `types.go` como `DeleteItemPayload` si no existe ya con ese nombre.

#### Función 1: handleGetProfile (Líneas 31-168) - **MIGRAR COMPLETAMENTE**

**Acción requerida:** Migrar función principal para obtener perfiles de usuario (propio o de otros).

**Cambios específicos:**

```go
// Línea 31: CAMBIAR firma
func (h *Hub) handleGetProfile(client *Client, payload interface{}) error
// ↓ CAMBIAR A:
func handleGetProfile(appCtx *AppContext, conn *customws.Connection[AppUserData], msg customws_types.ClientToServerMessage) error

// Líneas 32-40: SIMPLIFICAR parsing payload
var req ProfileRequestPayload
if msg.Payload != nil {
	if err := json.Unmarshal(msg.Payload, &req); err != nil {
		logger.Infof("WS", "handleGetProfile Info (UserID: %d): Payload is not a ProfileRequestPayload or is empty: %v", conn.UserData.ID, err)
		// Continuar, targetUserID será el propio usuario si req está vacía o malformada
	}
}

// Líneas 42-45: ACTUALIZAR lógica de targetUserID
targetUserID := client.User.ID
if req.TargetUserID != 0 {
	targetUserID = req.TargetUserID
}
// ↓ CAMBIAR A:
targetUserID := conn.UserData.ID
if req.TargetUserID != 0 {
	targetUserID = req.TargetUserID
}

// Línea 48: ACTUALIZAR llamada a función auxiliar
userBase, err := h.getBaseUserProfile(targetUserID)
// ↓ CAMBIAR A:
userBase, err := getBaseUserProfile(appCtx, targetUserID)

// Líneas 49-55: ACTUALIZAR manejo de errores
if err == sql.ErrNoRows {
	return h.sendErrorMessage(client, "Requested user profile not found")
}
if err != nil {
	return h.sendErrorMessage(client, "Failed to retrieve profile data")
}
// ↓ CAMBIAR A:
if err == sql.ErrNoRows {
	conn.SendErrorNotification(msg.PID, 404, "Requested user profile not found")
	return fmt.Errorf("user profile not found: %d", targetUserID)
}
if err != nil {
	conn.SendErrorNotification(msg.PID, 500, "Failed to retrieve profile data")
	return err
}

// Líneas 57-107: MANTENER lógica de mapeo a MyProfileResponse (ya es correcta)

// Línea 110: ACTUALIZAR obtención de estado online
_, isOnline := h.GetClient(targetUserID)
// ↓ CAMBIAR A:
isOnline := appCtx.ConnectionManager.IsUserConnected(targetUserID)

// Líneas 112-125: ACTUALIZAR verificación de permisos
if targetUserID != client.User.ID {
	isContact, contactStatus, err := h.getContactStatus(client.User.ID, targetUserID)
	// ... lógica de error y admin check ...
	return h.sendErrorMessage(client, "Permission denied to view this profile")
}
// ↓ CAMBIAR A:
if targetUserID != conn.UserData.ID {
	isContact, contactStatus, err := getContactStatus(appCtx, conn.UserData.ID, targetUserID)
	if err != nil {
		conn.SendErrorNotification(msg.PID, 500, "Failed to check contact status")
		return err
	}
	isAdmin := conn.UserData.RoleID == 7 || conn.UserData.RoleID == 8 // Asumiendo que RoleID está en AppUserData
	if !isAdmin && !isContact { // y que getContactStatus devuelve `isContact` correctamente
		conn.SendErrorNotification(msg.PID, 403, "Permission denied to view this profile")
		return fmt.Errorf("permission denied to view profile %d", targetUserID)
	}
}

// Línea 128: ACTUALIZAR llamada a función auxiliar
curriculum, err := h.getCurriculum(targetUserID)
// ↓ CAMBIAR A:
curriculum, err := getCurriculum(appCtx, targetUserID) // getCurriculum necesita ser migrada también

// Líneas 135-138: MANTENER asignación de currículum (profileResp.Curriculum = *curriculum)

// Líneas 140-152: ACTUALIZAR envío de respuesta
outgoingMsg := OutgoingMessage{Type: MessageTypeGetProfileResponse, Payload: profileResp}
jsonMsg, err := json.Marshal(outgoingMsg)
client.SendToMe(jsonMsg)
// ↓ CAMBIAR A:
response := customws_types.ServerToClientMessage{
	PID:       msg.PID,
	Type:      MessageTypeGetProfileResponse, // Mantener la constante de types.go
	Payload:   profileResp,
	Timestamp: time.Now().UnixMilli(),
}
return conn.SendMessage(response)
```

#### Función 2: handleGetMyProfile (Líneas 170-175) - **MIGRAR COMPLETAMENTE**

**Acción requerida:** Simplificar y llamar al nuevo `handleGetProfile`.

```go
// Línea 170: CAMBIAR firma
func (h *Hub) handleGetMyProfile(client *Client) error
// ↓ CAMBIAR A:
func handleGetMyProfile(appCtx *AppContext, conn *customws.Connection[AppUserData], msg customws_types.ClientToServerMessage) error

// Línea 173: SIMPLIFICAR llamada
return h.handleGetProfile(client, nil) // Payload nil para indicar perfil propio
// ↓ CAMBIAR A:
// El payload de msg ya sería nil o un ProfileRequestPayload vacío por defecto.
// Si se quiere forzar que sea el perfil propio, se puede pasar un ProfileRequestPayload con TargetUserID = 0 o conn.UserData.ID
// O simplemente llamar a handleGetProfile, que ya tiene la lógica para deducir el targetUserID.
return handleGetProfile(appCtx, conn, msg)
```

#### Función 3: handleEditProfile (Líneas 177-228) - **MIGRAR COMPLETAMENTE**

**Acción requerida:** Migrar función para actualizar perfil de usuario.

```go
// Línea 177: CAMBIAR firma
func (h *Hub) handleEditProfile(client *Client, payload interface{}) error
// ↓ CAMBIAR A:
func handleEditProfile(appCtx *AppContext, conn *customws.Connection[AppUserData], msg customws_types.ClientToServerMessage) error

// Líneas 181-194: SIMPLIFICAR parsing payload
var updateReq models.User // Usar `EditProfilePayload` de `types.go` en su lugar
// ... parsing manual ...
// ↓ CAMBIAR A:
var req EditProfilePayload // Usar el tipo de types.go
if err := json.Unmarshal(msg.Payload, &req); err != nil {
	conn.SendErrorNotification(msg.PID, 400, "Invalid update payload format")
	return fmt.Errorf("invalid payload: %w", err)
}

// Líneas 196-207: CAMBIAR acceso a DB y parámetros
_, err := h.DB.Exec(query, updateReq.FirstName, ..., userID)
// ↓ CAMBIAR A:
// Esta parte necesita una refactorización cuidadosa para construir la query y los args
// a partir de los campos no nulos de `req` (EditProfilePayload).
// Ejemplo conceptual:
// updates := []string{}
// args := []interface{}{}
// if req.Summary != nil { updates = append(updates, "Summary = ?"); args = append(args, *req.Summary) }
// ... etc. para todos los campos ...
// if len(updates) == 0 { /* no hay nada que actualizar */ return ... }
// query := fmt.Sprintf("UPDATE User SET %s WHERE Id = ?", strings.Join(updates, ", "))
// args = append(args, conn.UserData.ID)
// _, err := appCtx.DB.Exec(query, args...)

// Líneas 209-212: ACTUALIZAR manejo de errores
if err != nil {
	return h.sendErrorMessage(client, "Failed to update profile information")
}
// ↓ CAMBIAR A:
if err != nil {
	conn.SendErrorNotification(msg.PID, 500, "Failed to update profile information")
	return err
}

// Líneas 216-222: ACTUALIZAR `client.User` (ya no existirá)
// Esta lógica de actualizar el `client.User` en memoria ya no es necesaria con `customws`,
// ya que `AppUserData` es más estático o se recarga con cada conexión.
// Si se necesita refrescar `conn.UserData`, se debería hacer explícitamente.
// ELIMINAR esta sección.

// Línea 224: ACTUALIZAR respuesta de éxito
return h.sendSuccessMessage(client, "Profile updated successfully")
// ↓ CAMBIAR A:
response := customws_types.ServerToClientMessage{
	PID:       msg.PID,
	Type:      MessageTypeEditProfileResponse, // AGREGAR constante en types.go
	Payload:   SuccessPayload{Message: "Profile updated successfully"},
	Timestamp: time.Now().UnixMilli(),
}
return conn.SendMessage(response)
```

#### Función 4: handleDeleteItemCurriculum (Líneas 230-293) - **MIGRAR COMPLETAMENTE**

**Acción requerida:** Migrar función para eliminar items del currículum.

```go
// Línea 230: CAMBIAR firma
func (h *Hub) handleDeleteItemCurriculum(client *Client, payload interface{}) error
// ↓ CAMBIAR A:
func handleDeleteItemCurriculum(appCtx *AppContext, conn *customws.Connection[AppUserData], msg customws_types.ClientToServerMessage) error

// Líneas 234-247: SIMPLIFICAR parsing payload
var req DeleteCurriculumItemRequest // Usar `DeleteItemPayload` de `types.go`
// ... parsing manual ...
// ↓ CAMBIAR A:
var req DeleteItemPayload // Usar el tipo de types.go
if err := json.Unmarshal(msg.Payload, &req); err != nil {
	conn.SendErrorNotification(msg.PID, 400, "Invalid delete payload format")
	return fmt.Errorf("invalid payload: %w", err)
}

// Líneas 249-252: MANTENER validación (req.ItemType y req.ItemID)
// Ajustar el mensaje de error para `conn.SendErrorNotification`

// Líneas 254-274: MANTENER lógica de switch para `tableName` y `personIdColumn`

// Línea 276: CAMBIAR acceso a DB
result, err := h.DB.Exec(query, req.ItemID, userID)
// ↓ CAMBIAR A:
result, err := appCtx.DB.Exec(query, req.ItemID, conn.UserData.ID)

// Líneas 277-286: ACTUALIZAR manejo de errores y rowsAffected
// Usar `conn.SendErrorNotification`

// Línea 290: ACTUALIZAR respuesta de éxito
return h.sendSuccessMessage(client, fmt.Sprintf("%s deleted successfully", req.ItemType))
// ↓ CAMBIAR A:
response := customws_types.ServerToClientMessage{
	PID:       msg.PID,
	Type:      MessageTypeDeleteItemCurriculumResponse, // AGREGAR constante en types.go
	Payload:   SuccessPayload{Message: fmt.Sprintf("%s deleted successfully", req.ItemType)},
	Timestamp: time.Now().UnixMilli(),
}
return conn.SendMessage(response)
```

#### Función 5: getBaseUserProfile (Líneas 298-344) - **MIGRAR (AUXILIAR)**

**Acción requerida:** Cambiar firma para usar `appCtx`.

```go
// Línea 298: CAMBIAR firma
func (h *Hub) getBaseUserProfile(userID int64) (*models.User, error)
// ↓ CAMBIAR A:
func getBaseUserProfile(appCtx *AppContext, userID int64) (*models.User, error)

// Línea 317: CAMBIAR acceso a DB
err := h.DB.QueryRow(query, userID).Scan(...)
// ↓ CAMBIAR A:
err := appCtx.DB.QueryRow(query, userID).Scan(...)

// Resto de la lógica MANTENER (scan y asignación de NullString)
```

#### Función 6: getCurriculum (Líneas 346-401) - **MIGRAR (AUXILIAR)**

**Acción requerida:** Cambiar firma para usar `appCtx` y ajustar llamadas a funciones auxiliares de currículum.

```go
// Línea 347: CAMBIAR firma
func (h *Hub) getCurriculum(userID int64) (*Curriculum, error)
// ↓ CAMBIAR A:
func getCurriculum(appCtx *AppContext, userID int64) (*Curriculum, error)

// Línea 356: CAMBIAR llamada a DB dentro de la goroutine
data, fetchErr := fetchFunc(h.DB, userID)
// ↓ CAMBIAR A:
data, fetchErr := fetchFunc(appCtx.DB, userID) // Asumiendo que las funciones get* de curriculum se adaptan para recibir appCtx.DB

// Líneas 383-388: ACTUALIZAR llamadas a fetch
// Las funciones getEducation, getWorkExperience, etc. (de handlers_curriculum.go) deben aceptar (appCtx.DB, userID)
// O mejor, que acepten (appCtx *AppContext, userID int64) y accedan a appCtx.DB internamente.
// Ejemplo:
fetch(func(db *sql.DB, id int64) (interface{}, error) { return getEducation(db, id) }, &curriculum.Education)
// ↓ CAMBIAR A (si getEducation toma AppContext):
fetch(func(ctx *AppContext, id int64) (interface{}, error) { return getEducation(ctx, id) }, &curriculum.Education)
// O (si getEducation toma *sql.DB, pasar appCtx.DB):
fetch(func(db *sql.DB, id int64) (interface{}, error) { return getEducation(db, id) }, &curriculum.Education)
// Este último es el patrón actual de handlers_curriculum.go, por lo que solo el `h.DB` en la línea 356 necesita cambiar a `appCtx.DB`.

// Resto de la lógica MANTENER (sync.WaitGroup, error handling, type assertions)
```

**Líneas 403-410: Stubs de actualización de currículum - ELIMINAR COMENTADOS**

### Cambios Críticos Resumidos para handlers_profile.go:

1.  **Imports:** AGREGAR `customws_types`, `time`. ELIMINAR `encoding/json` donde no se parseen payloads directamente (ahora lo hace customws).
2.  **Structs locales:** Mover `DeleteCurriculumItemRequest` a `types.go` (posiblemente como `DeleteItemPayload`).
3.  **Firmas de Handlers:** Actualizar todas las firmas a `func handlerName(appCtx *AppContext, conn *customws.Connection[AppUserData], msg customws_types.ClientToServerMessage) error`.
4.  **Firmas Auxiliares:** `getBaseUserProfile` y `getCurriculum` deben tomar `appCtx *AppContext`.
5.  **Acceso a DB:** Cambiar `h.DB` a `appCtx.DB`.
6.  **Payload Parsing:** Usar `json.Unmarshal(msg.Payload, &reqStruct)`.
7.  **User Info:** `client.User.ID` → `conn.UserData.ID`, `client.User.RoleID` → `conn.UserData.RoleID`.
8.  **Respuestas:** Usar `conn.SendErrorNotification()` y `conn.SendMessage()` con `customws_types.ServerToClientMessage`.
9.  **Nuevos Tipos de Mensaje:** `MessageTypeEditProfileResponse`, `MessageTypeDeleteItemCurriculumResponse`.
10. **Estado Online:** `h.GetClient()` → `appCtx.ConnectionManager.IsUserConnected()`.
11. **Llamadas a `getContactStatus` y `getCurriculum`:** Actualizar para usar `appCtx`.
12. **Actualización de `client.User`:** Eliminar la lógica de `handleEditProfile` que actualizaba `client.User`.

### Esquema de BD utilizado por handlers_profile.go:
```sql
User (Id, FirstName, LastName, UserName, Email, Phone, Sex, DocId, NationalityId, Birthdate, Picture, DegreeId, UniversityId, RoleId, StatusAuthorizedId, Summary, Address, Github, Linkedin)
Nationality (Id, CountryName)
Degree (Id, DegreeName)
University (Id, Name)
Role (Id, Name)
-- Más tablas de Curriculum (Education, WorkExperience, etc. - llamadas vía getCurriculum)
Contact (User1Id, User2Id, Status) -- Usado por getContactStatus
```

### 7. `handlers_search.go` (Prioridad: MEDIA)

**Estado:** Migración completa requerida - Contiene handlers para obtener y editar perfiles, y eliminar items del currículum, además de funciones auxiliares clave.

#### Análisis línea por línea:

**Líneas 1-10: Package e imports**
```go
// CAMBIAR imports
package websocket

import (
	"database/sql"      // MANTENER
	"encoding/json"     // ELIMINAR - no necesario con customws
	"fmt"               // MANTENER
	"sync"              // MANTENER (usado en getCurriculum)

	"github.com/davidM20/micro-service-backend-go.git/internal/models" // MANTENER
	"github.com/davidM20/micro-service-backend-go.git/pkg/logger"      // MANTENER
	customws_types "github.com/davidM20/micro-service-backend-go.git/pkg/customws/types" // AGREGAR
	"time"                                                               // AGREGAR
)
```

**Líneas 12-26: Structs locales**
- `ProfileResponse` (Líneas 15-20): **ELIMINAR COMENTADO** - Ya se indica que se usa `MyProfileResponse` de `types.go`.
- `DeleteCurriculumItemRequest` (Líneas 23-26): **MANTENER** - Es usado por `handleDeleteItemCurriculum`. Podría moverse a `types.go` como `DeleteItemPayload` si no existe ya con ese nombre.

#### Función 1: handleGetProfile (Líneas 31-168) - **MIGRAR COMPLETAMENTE**

**Acción requerida:** Migrar función principal para obtener perfiles de usuario (propio o de otros).

**Cambios específicos:**

```go
// Línea 31: CAMBIAR firma
func (h *Hub) handleGetProfile(client *Client, payload interface{}) error
// ↓ CAMBIAR A:
func handleGetProfile(appCtx *AppContext, conn *customws.Connection[AppUserData], msg customws_types.ClientToServerMessage) error

// Líneas 32-40: SIMPLIFICAR parsing payload
var req ProfileRequestPayload
if msg.Payload != nil {
	if err := json.Unmarshal(msg.Payload, &req); err != nil {
		logger.Infof("WS", "handleGetProfile Info (UserID: %d): Payload is not a ProfileRequestPayload or is empty: %v", conn.UserData.ID, err)
		// Continuar, targetUserID será el propio usuario si req está vacía o malformada
	}
}

// Líneas 42-45: ACTUALIZAR lógica de targetUserID
targetUserID := client.User.ID
if req.TargetUserID != 0 {
	targetUserID = req.TargetUserID
}
// ↓ CAMBIAR A:
targetUserID := conn.UserData.ID
if req.TargetUserID != 0 {
	targetUserID = req.TargetUserID
}

// Línea 48: ACTUALIZAR llamada a función auxiliar
userBase, err := h.getBaseUserProfile(targetUserID)
// ↓ CAMBIAR A:
userBase, err := getBaseUserProfile(appCtx, targetUserID)

// Líneas 49-55: ACTUALIZAR manejo de errores
if err == sql.ErrNoRows {
	return h.sendErrorMessage(client, "Requested user profile not found")
}
if err != nil {
	return h.sendErrorMessage(client, "Failed to retrieve profile data")
}
// ↓ CAMBIAR A:
if err == sql.ErrNoRows {
	conn.SendErrorNotification(msg.PID, 404, "Requested user profile not found")
	return fmt.Errorf("user profile not found: %d", targetUserID)
}
if err != nil {
	conn.SendErrorNotification(msg.PID, 500, "Failed to retrieve profile data")
	return err
}

// Líneas 57-107: MANTENER lógica de mapeo a MyProfileResponse (ya es correcta)

// Línea 110: ACTUALIZAR obtención de estado online
_, isOnline := h.GetClient(targetUserID)
// ↓ CAMBIAR A:
isOnline := appCtx.ConnectionManager.IsUserConnected(targetUserID)

// Líneas 112-125: ACTUALIZAR verificación de permisos
if targetUserID != client.User.ID {
	isContact, contactStatus, err := h.getContactStatus(client.User.ID, targetUserID)
	// ... lógica de error y admin check ...
	return h.sendErrorMessage(client, "Permission denied to view this profile")
}
// ↓ CAMBIAR A:
if targetUserID != conn.UserData.ID {
	isContact, contactStatus, err := getContactStatus(appCtx, conn.UserData.ID, targetUserID)
	if err != nil {
		conn.SendErrorNotification(msg.PID, 500, "Failed to check contact status")
		return err
	}
	isAdmin := conn.UserData.RoleID == 7 || conn.UserData.RoleID == 8 // Asumiendo que RoleID está en AppUserData
	if !isAdmin && !isContact { // y que getContactStatus devuelve `isContact` correctamente
		conn.SendErrorNotification(msg.PID, 403, "Permission denied to view this profile")
		return fmt.Errorf("permission denied to view profile %d", targetUserID)
	}
}

// Línea 128: ACTUALIZAR llamada a función auxiliar
curriculum, err := h.getCurriculum(targetUserID)
// ↓ CAMBIAR A:
curriculum, err := getCurriculum(appCtx, targetUserID) // getCurriculum necesita ser migrada también

// Líneas 135-138: MANTENER asignación de currículum (profileResp.Curriculum = *curriculum)

// Líneas 140-152: ACTUALIZAR envío de respuesta
outgoingMsg := OutgoingMessage{Type: MessageTypeGetProfileResponse, Payload: profileResp}
jsonMsg, err := json.Marshal(outgoingMsg)
client.SendToMe(jsonMsg)
// ↓ CAMBIAR A:
response := customws_types.ServerToClientMessage{
	PID:       msg.PID,
	Type:      MessageTypeGetProfileResponse, // Mantener la constante de types.go
	Payload:   profileResp,
	Timestamp: time.Now().UnixMilli(),
}
return conn.SendMessage(response)
```

#### Función 2: handleGetMyProfile (Líneas 170-175) - **MIGRAR COMPLETAMENTE**

**Acción requerida:** Simplificar y llamar al nuevo `handleGetProfile`.

```go
// Línea 170: CAMBIAR firma
func (h *Hub) handleGetMyProfile(client *Client) error
// ↓ CAMBIAR A:
func handleGetMyProfile(appCtx *AppContext, conn *customws.Connection[AppUserData], msg customws_types.ClientToServerMessage) error

// Línea 173: SIMPLIFICAR llamada
return h.handleGetProfile(client, nil) // Payload nil para indicar perfil propio
// ↓ CAMBIAR A:
// El payload de msg ya sería nil o un ProfileRequestPayload vacío por defecto.
// Si se quiere forzar que sea el perfil propio, se puede pasar un ProfileRequestPayload con TargetUserID = 0 o conn.UserData.ID
// O simplemente llamar a handleGetProfile, que ya tiene la lógica para deducir el targetUserID.
return handleGetProfile(appCtx, conn, msg)
```

#### Función 3: handleEditProfile (Líneas 177-228) - **MIGRAR COMPLETAMENTE**

**Acción requerida:** Migrar función para actualizar perfil de usuario.

```go
// Línea 177: CAMBIAR firma
func (h *Hub) handleEditProfile(client *Client, payload interface{}) error
// ↓ CAMBIAR A:
func handleEditProfile(appCtx *AppContext, conn *customws.Connection[AppUserData], msg customws_types.ClientToServerMessage) error

// Líneas 181-194: SIMPLIFICAR parsing payload
var updateReq models.User // Usar `EditProfilePayload` de `types.go` en su lugar
// ... parsing manual ...
// ↓ CAMBIAR A:
var req EditProfilePayload // Usar el tipo de types.go
if err := json.Unmarshal(msg.Payload, &req); err != nil {
	conn.SendErrorNotification(msg.PID, 400, "Invalid update payload format")
	return fmt.Errorf("invalid payload: %w", err)
}

// Líneas 196-207: CAMBIAR acceso a DB y parámetros
_, err := h.DB.Exec(query, updateReq.FirstName, ..., userID)
// ↓ CAMBIAR A:
// Esta parte necesita una refactorización cuidadosa para construir la query y los args
// a partir de los campos no nulos de `req` (EditProfilePayload).
// Ejemplo conceptual:
// updates := []string{}
// args := []interface{}{}
// if req.Summary != nil { updates = append(updates, "Summary = ?"); args = append(args, *req.Summary) }
// ... etc. para todos los campos ...
// if len(updates) == 0 { /* no hay nada que actualizar */ return ... }
// query := fmt.Sprintf("UPDATE User SET %s WHERE Id = ?", strings.Join(updates, ", "))
// args = append(args, conn.UserData.ID)
// _, err := appCtx.DB.Exec(query, args...)

// Líneas 209-212: ACTUALIZAR manejo de errores
if err != nil {
	return h.sendErrorMessage(client, "Failed to update profile information")
}
// ↓ CAMBIAR A:
if err != nil {
	conn.SendErrorNotification(msg.PID, 500, "Failed to update profile information")
	return err
}

// Líneas 216-222: ACTUALIZAR `client.User` (ya no existirá)
// Esta lógica de actualizar el `client.User` en memoria ya no es necesaria con `customws`,
// ya que `AppUserData` es más estático o se recarga con cada conexión.
// Si se necesita refrescar `conn.UserData`, se debería hacer explícitamente.
// ELIMINAR esta sección.

// Línea 224: ACTUALIZAR respuesta de éxito
return h.sendSuccessMessage(client, "Profile updated successfully")
// ↓ CAMBIAR A:
response := customws_types.ServerToClientMessage{
	PID:       msg.PID,
	Type:      MessageTypeEditProfileResponse, // AGREGAR constante en types.go
	Payload:   SuccessPayload{Message: "Profile updated successfully"},
	Timestamp: time.Now().UnixMilli(),
}
return conn.SendMessage(response)
```

#### Función 4: handleDeleteItemCurriculum (Líneas 230-293) - **MIGRAR COMPLETAMENTE**

**Acción requerida:** Migrar función para eliminar items del currículum.

```go
// Línea 230: CAMBIAR firma
func (h *Hub) handleDeleteItemCurriculum(client *Client, payload interface{}) error
// ↓ CAMBIAR A:
func handleDeleteItemCurriculum(appCtx *AppContext, conn *customws.Connection[AppUserData], msg customws_types.ClientToServerMessage) error

// Líneas 234-247: SIMPLIFICAR parsing payload
var req DeleteCurriculumItemRequest // Usar `DeleteItemPayload` de `types.go`
// ... parsing manual ...
// ↓ CAMBIAR A:
var req DeleteItemPayload // Usar el tipo de types.go
if err := json.Unmarshal(msg.Payload, &req); err != nil {
	conn.SendErrorNotification(msg.PID, 400, "Invalid delete payload format")
	return fmt.Errorf("invalid payload: %w", err)
}

// Líneas 249-252: MANTENER validación (req.ItemType y req.ItemID)
// Ajustar el mensaje de error para `conn.SendErrorNotification`

// Líneas 254-274: MANTENER lógica de switch para `tableName` y `personIdColumn`

// Línea 276: CAMBIAR acceso a DB
result, err := h.DB.Exec(query, req.ItemID, userID)
// ↓ CAMBIAR A:
result, err := appCtx.DB.Exec(query, req.ItemID, conn.UserData.ID)

// Líneas 277-286: ACTUALIZAR manejo de errores y rowsAffected
// Usar `conn.SendErrorNotification`

// Línea 290: ACTUALIZAR respuesta de éxito
return h.sendSuccessMessage(client, fmt.Sprintf("%s deleted successfully", req.ItemType))
// ↓ CAMBIAR A:
response := customws_types.ServerToClientMessage{
	PID:       msg.PID,
	Type:      MessageTypeDeleteItemCurriculumResponse, // AGREGAR constante en types.go
	Payload:   SuccessPayload{Message: fmt.Sprintf("%s deleted successfully", req.ItemType)},
	Timestamp: time.Now().UnixMilli(),
}
return conn.SendMessage(response)
```

#### Función 5: getBaseUserProfile (Líneas 298-344) - **MIGRAR (AUXILIAR)**

**Acción requerida:** Cambiar firma para usar `appCtx`.

```go
// Línea 298: CAMBIAR firma
func (h *Hub) getBaseUserProfile(userID int64) (*models.User, error)
// ↓ CAMBIAR A:
func getBaseUserProfile(appCtx *AppContext, userID int64) (*models.User, error)

// Línea 317: CAMBIAR acceso a DB
err := h.DB.QueryRow(query, userID).Scan(...)
// ↓ CAMBIAR A:
err := appCtx.DB.QueryRow(query, userID).Scan(...)

// Resto de la lógica MANTENER (scan y asignación de NullString)
```

#### Función 6: getCurriculum (Líneas 346-401) - **MIGRAR (AUXILIAR)**

**Acción requerida:** Cambiar firma para usar `appCtx` y ajustar llamadas a funciones auxiliares de currículum.

```go
// Línea 347: CAMBIAR firma
func (h *Hub) getCurriculum(userID int64) (*Curriculum, error)
// ↓ CAMBIAR A:
func getCurriculum(appCtx *AppContext, userID int64) (*Curriculum, error)

// Línea 356: CAMBIAR llamada a DB dentro de la goroutine
data, fetchErr := fetchFunc(h.DB, userID)
// ↓ CAMBIAR A:
data, fetchErr := fetchFunc(appCtx.DB, userID) // Asumiendo que las funciones get* de curriculum se adaptan para recibir appCtx.DB

// Líneas 383-388: ACTUALIZAR llamadas a fetch
// Las funciones getEducation, getWorkExperience, etc. (de handlers_curriculum.go) deben aceptar (appCtx.DB, userID)
// O mejor, que acepten (appCtx *AppContext, userID int64) y accedan a appCtx.DB internamente.
// Ejemplo:
fetch(func(db *sql.DB, id int64) (interface{}, error) { return getEducation(db, id) }, &curriculum.Education)
// ↓ CAMBIAR A (si getEducation toma AppContext):
fetch(func(ctx *AppContext, id int64) (interface{}, error) { return getEducation(ctx, id) }, &curriculum.Education)
// O (si getEducation toma *sql.DB, pasar appCtx.DB):
fetch(func(db *sql.DB, id int64) (interface{}, error) { return getEducation(db, id) }, &curriculum.Education)
// Este último es el patrón actual de handlers_curriculum.go, por lo que solo el `h.DB` en la línea 356 necesita cambiar a `appCtx.DB`.

// Resto de la lógica MANTENER (sync.WaitGroup, error handling, type assertions)
```

**Líneas 403-410: Stubs de actualización de currículum - ELIMINAR COMENTADOS**

### Cambios Críticos Resumidos para handlers_profile.go:

1.  **Imports:** AGREGAR `customws_types`, `time`. ELIMINAR `encoding/json` donde no se parseen payloads directamente (ahora lo hace customws).
2.  **Structs locales:** Mover `DeleteCurriculumItemRequest` a `types.go` (posiblemente como `DeleteItemPayload`).
3.  **Firmas de Handlers:** Actualizar todas las firmas a `func handlerName(appCtx *AppContext, conn *customws.Connection[AppUserData], msg customws_types.ClientToServerMessage) error`.
4.  **Firmas Auxiliares:** `getBaseUserProfile` y `getCurriculum` deben tomar `appCtx *AppContext`.
5.  **Acceso a DB:** Cambiar `h.DB` a `appCtx.DB`.
6.  **Payload Parsing:** Usar `json.Unmarshal(msg.Payload, &reqStruct)`.
7.  **User Info:** `client.User.ID` → `conn.UserData.ID`, `client.User.RoleID` → `conn.UserData.RoleID`.
8.  **Respuestas:** Usar `conn.SendErrorNotification()` y `conn.SendMessage()` con `customws_types.ServerToClientMessage`.
9.  **Nuevos Tipos de Mensaje:** `MessageTypeEditProfileResponse`, `MessageTypeDeleteItemCurriculumResponse`.
10. **Estado Online:** `h.GetClient()` → `appCtx.ConnectionManager.IsUserConnected()`.
11. **Llamadas a `getContactStatus` y `getCurriculum`:** Actualizar para usar `appCtx`.
12. **Actualización de `client.User`:** Eliminar la lógica de `handleEditProfile` que actualizaba `client.User`.

### Esquema de BD utilizado por handlers_profile.go:
```sql
User (Id, FirstName, LastName, UserName, Email, Phone, Sex, DocId, NationalityId, Birthdate, Picture, DegreeId, UniversityId, RoleId, StatusAuthorizedId, Summary, Address, Github, Linkedin)
Nationality (Id, CountryName)
Degree (Id, DegreeName)
University (Id, Name)
Role (Id, Name)
-- Más tablas de Curriculum (Education, WorkExperience, etc. - llamadas vía getCurriculum)
Contact (User1Id, User2Id, Status) -- Usado por getContactStatus
```

### 7. `handlers_search.go` (Prioridad: MEDIA)

**Estado:** Migración completa requerida - Contiene handlers para obtener y editar perfiles, y eliminar items del currículum, además de funciones auxiliares clave.

#### Análisis línea por línea:

**Líneas 1-10: Package e imports**
```go
// CAMBIAR imports
package websocket

import (
	"database/sql"      // MANTENER
	"encoding/json"     // ELIMINAR - no necesario con customws
	"fmt"               // MANTENER
	"sync"              // MANTENER (usado en getCurriculum)

	"github.com/davidM20/micro-service-backend-go.git/internal/models" // MANTENER
	"github.com/davidM20/micro-service-backend-go.git/pkg/logger"      // MANTENER
	customws_types "github.com/davidM20/micro-service-backend-go.git/pkg/customws/types" // AGREGAR
	"time"                                                               // AGREGAR
)
```

**Líneas 12-26: Structs locales**
- `ProfileResponse` (Líneas 15-20): **ELIMINAR COMENTADO** - Ya se indica que se usa `MyProfileResponse` de `types.go`.
- `DeleteCurriculumItemRequest` (Líneas 23-26): **MANTENER** - Es usado por `handleDeleteItemCurriculum`. Podría moverse a `types.go` como `DeleteItemPayload` si no existe ya con ese nombre.

#### Función 1: handleGetProfile (Líneas 31-168) - **MIGRAR COMPLETAMENTE**

**Acción requerida:** Migrar función principal para obtener perfiles de usuario (propio o de otros).

**Cambios específicos:**

```go
// Línea 31: CAMBIAR firma
func (h *Hub) handleGetProfile(client *Client, payload interface{}) error
// ↓ CAMBIAR A:
func handleGetProfile(appCtx *AppContext, conn *customws.Connection[AppUserData], msg customws_types.ClientToServerMessage) error

// Líneas 32-40: SIMPLIFICAR parsing payload
var req ProfileRequestPayload
if msg.Payload != nil {
	if err := json.Unmarshal(msg.Payload, &req); err != nil {
		logger.Infof("WS", "handleGetProfile Info (UserID: %d): Payload is not a ProfileRequestPayload or is empty: %v", conn.UserData.ID, err)
		// Continuar, targetUserID será el propio usuario si req está vacía o malformada
	}
}

// Líneas 42-45: ACTUALIZAR lógica de targetUserID
targetUserID := client.User.ID
if req.TargetUserID != 0 {
	targetUserID = req.TargetUserID
}
// ↓ CAMBIAR A:
targetUserID := conn.UserData.ID
if req.TargetUserID != 0 {
	targetUserID = req.TargetUserID
}

// Línea 48: ACTUALIZAR llamada a función auxiliar
userBase, err := h.getBaseUserProfile(targetUserID)
// ↓ CAMBIAR A:
userBase, err := getBaseUserProfile(appCtx, targetUserID)

// Líneas 49-55: ACTUALIZAR manejo de errores
if err == sql.ErrNoRows {
	return h.sendErrorMessage(client, "Requested user profile not found")
}
if err != nil {
	return h.sendErrorMessage(client, "Failed to retrieve profile data")
}
// ↓ CAMBIAR A:
if err == sql.ErrNoRows {
	conn.SendErrorNotification(msg.PID, 404, "Requested user profile not found")
	return fmt.Errorf("user profile not found: %d", targetUserID)
}
if err != nil {
	conn.SendErrorNotification(msg.PID, 500, "Failed to retrieve profile data")
	return err
}

// Líneas 57-107: MANTENER lógica de mapeo a MyProfileResponse (ya es correcta)

// Línea 110: ACTUALIZAR obtención de estado online
_, isOnline := h.GetClient(targetUserID)
// ↓ CAMBIAR A:
isOnline := appCtx.ConnectionManager.IsUserConnected(targetUserID)

// Líneas 112-125: ACTUALIZAR verificación de permisos
if targetUserID != client.User.ID {
	isContact, contactStatus, err := h.getContactStatus(client.User.ID, targetUserID)
	// ... lógica de error y admin check ...
	return h.sendErrorMessage(client, "Permission denied to view this profile")
}
// ↓ CAMBIAR A:
if targetUserID != conn.UserData.ID {
	isContact, contactStatus, err := getContactStatus(appCtx, conn.UserData.ID, targetUserID)
	if err != nil {
		conn.SendErrorNotification(msg.PID, 500, "Failed to check contact status")
		return err
	}
	isAdmin := conn.UserData.RoleID == 7 || conn.UserData.RoleID == 8 // Asumiendo que RoleID está en AppUserData
	if !isAdmin && !isContact { // y que getContactStatus devuelve `isContact` correctamente
		conn.SendErrorNotification(msg.PID, 403, "Permission denied to view this profile")
		return fmt.Errorf("permission denied to view profile %d", targetUserID)
	}
}

// Línea 128: ACTUALIZAR llamada a función auxiliar
curriculum, err := h.getCurriculum(targetUserID)
// ↓ CAMBIAR A:
curriculum, err := getCurriculum(appCtx, targetUserID) // getCurriculum necesita ser migrada también

// Líneas 135-138: MANTENER asignación de currículum (profileResp.Curriculum = *curriculum)

// Líneas 140-152: ACTUALIZAR envío de respuesta
outgoingMsg := OutgoingMessage{Type: MessageTypeGetProfileResponse, Payload: profileResp}
jsonMsg, err := json.Marshal(outgoingMsg)
client.SendToMe(jsonMsg)
// ↓ CAMBIAR A:
response := customws_types.ServerToClientMessage{
	PID:       msg.PID,
	Type:      MessageTypeGetProfileResponse, // Mantener la constante de types.go
	Payload:   profileResp,
	Timestamp: time.Now().UnixMilli(),
}
return conn.SendMessage(response)
```

#### Función 2: handleGetMyProfile (Líneas 170-175) - **MIGRAR COMPLETAMENTE**

**Acción requerida:** Simplificar y llamar al nuevo `handleGetProfile`.

```go
// Línea 170: CAMBIAR firma
func (h *Hub) handleGetMyProfile(client *Client) error
// ↓ CAMBIAR A:
func handleGetMyProfile(appCtx *AppContext, conn *customws.Connection[AppUserData], msg customws_types.ClientToServerMessage) error

// Línea 173: SIMPLIFICAR llamada
return h.handleGetProfile(client, nil) // Payload nil para indicar perfil propio
// ↓ CAMBIAR A:
// El payload de msg ya sería nil o un ProfileRequestPayload vacío por defecto.
// Si se quiere forzar que sea el perfil propio, se puede pasar un ProfileRequestPayload con TargetUserID = 0 o conn.UserData.ID
// O simplemente llamar a handleGetProfile, que ya tiene la lógica para deducir el targetUserID.
return handleGetProfile(appCtx, conn, msg)
```

#### Función 3: handleEditProfile (Líneas 177-228) - **MIGRAR COMPLETAMENTE**

**Acción requerida:** Migrar función para actualizar perfil de usuario.

```go
// Línea 177: CAMBIAR firma
func (h *Hub) handleEditProfile(client *Client, payload interface{}) error
// ↓ CAMBIAR A:
func handleEditProfile(appCtx *AppContext, conn *customws.Connection[AppUserData], msg customws_types.ClientToServerMessage) error

// Líneas 181-194: SIMPLIFICAR parsing payload
var updateReq models.User // Usar `EditProfilePayload` de `types.go` en su lugar
// ... parsing manual ...
// ↓ CAMBIAR A:
var req EditProfilePayload // Usar el tipo de types.go
if err := json.Unmarshal(msg.Payload, &req); err != nil {
	conn.SendErrorNotification(msg.PID, 400, "Invalid update payload format")
	return fmt.Errorf("invalid payload: %w", err)
}

// Líneas 196-207: CAMBIAR acceso a DB y parámetros
_, err := h.DB.Exec(query, updateReq.FirstName, ..., userID)
// ↓ CAMBIAR A:
// Esta parte necesita una refactorización cuidadosa para construir la query y los args
// a partir de los campos no nulos de `req` (EditProfilePayload).
// Ejemplo conceptual:
// updates := []string{}
// args := []interface{}{}
// if req.Summary != nil { updates = append(updates, "Summary = ?"); args = append(args, *req.Summary) }
// ... etc. para todos los campos ...
// if len(updates) == 0 { /* no hay nada que actualizar */ return ... }
// query := fmt.Sprintf("UPDATE User SET %s WHERE Id = ?", strings.Join(updates, ", "))
// args = append(args, conn.UserData.ID)
// _, err := appCtx.DB.Exec(query, args...)

// Líneas 209-212: ACTUALIZAR manejo de errores
if err != nil {
	return h.sendErrorMessage(client, "Failed to update profile information")
}
// ↓ CAMBIAR A:
if err != nil {
	conn.SendErrorNotification(msg.PID, 500, "Failed to update profile information")
	return err
}

// Líneas 216-222: ACTUALIZAR `client.User` (ya no existirá)
// Esta lógica de actualizar el `client.User` en memoria ya no es necesaria con `customws`,
// ya que `AppUserData` es más estático o se recarga con cada conexión.
// Si se necesita refrescar `conn.UserData`, se debería hacer explícitamente.
// ELIMINAR esta sección.

// Línea 224: ACTUALIZAR respuesta de éxito
return h.sendSuccessMessage(client, "Profile updated successfully")
// ↓ CAMBIAR A:
response := customws_types.ServerToClientMessage{
	PID:       msg.PID,
	Type:      MessageTypeEditProfileResponse, // AGREGAR constante en types.go
	Payload:   SuccessPayload{Message: "Profile updated successfully"},
	Timestamp: time.Now().UnixMilli(),
}
return conn.SendMessage(response)
```

#### Función 4: handleDeleteItemCurriculum (Líneas 230-293) - **MIGRAR COMPLETAMENTE**

**Acción requerida:** Migrar función para eliminar items del currículum.

```go
// Línea 230: CAMBIAR firma
func (h *Hub) handleDeleteItemCurriculum(client *Client, payload interface{}) error
// ↓ CAMBIAR A:
func handleDeleteItemCurriculum(appCtx *AppContext, conn *customws.Connection[AppUserData], msg customws_types.ClientToServerMessage) error

// Líneas 234-247: SIMPLIFICAR parsing payload
var req DeleteCurriculumItemRequest // Usar `DeleteItemPayload` de `types.go`
// ... parsing manual ...
// ↓ CAMBIAR A:
var req DeleteItemPayload // Usar el tipo de types.go
if err := json.Unmarshal(msg.Payload, &req); err != nil {
	conn.SendErrorNotification(msg.PID, 400, "Invalid delete payload format")
	return fmt.Errorf("invalid payload: %w", err)
}

// Líneas 249-252: MANTENER validación (req.ItemType y req.ItemID)
// Ajustar el mensaje de error para `conn.SendErrorNotification`

// Líneas 254-274: MANTENER lógica de switch para `tableName` y `personIdColumn`

// Línea 276: CAMBIAR acceso a DB
result, err := h.DB.Exec(query, req.ItemID, userID)
// ↓ CAMBIAR A:
result, err := appCtx.DB.Exec(query, req.ItemID, conn.UserData.ID)

// Líneas 277-286: ACTUALIZAR manejo de errores y rowsAffected
// Usar `conn.SendErrorNotification`

// Línea 290: ACTUALIZAR respuesta de éxito
return h.sendSuccessMessage(client, fmt.Sprintf("%s deleted successfully", req.ItemType))
// ↓ CAMBIAR A:
response := customws_types.ServerToClientMessage{
	PID:       msg.PID,
	Type:      MessageTypeDeleteItemCurriculumResponse, // AGREGAR constante en types.go
	Payload:   SuccessPayload{Message: fmt.Sprintf("%s deleted successfully", req.ItemType)},
	Timestamp: time.Now().UnixMilli(),
}
return conn.SendMessage(response)
```

#### Función 5: getBaseUserProfile (Líneas 298-344) - **MIGRAR (AUXILIAR)**

**Acción requerida:** Cambiar firma para usar `appCtx`.

```go
// Línea 298: CAMBIAR firma
func (h *Hub) getBaseUserProfile(userID int64) (*models.User, error)
// ↓ CAMBIAR A:
func getBaseUserProfile(appCtx *AppContext, userID int64) (*models.User, error)

// Línea 317: CAMBIAR acceso a DB
err := h.DB.QueryRow(query, userID).Scan(...)
// ↓ CAMBIAR A:
err := appCtx.DB.QueryRow(query, userID).Scan(...)

// Resto de la lógica MANTENER (scan y asignación de NullString)
```

#### Función 6: getCurriculum (Líneas 346-401) - **MIGRAR (AUXILIAR)**

**Acción requerida:** Cambiar firma para usar `appCtx` y ajustar llamadas a funciones auxiliares de currículum.

```go
// Línea 347: CAMBIAR firma
func (h *Hub) getCurriculum(userID int64) (*Curriculum, error)
// ↓ CAMBIAR A:
func getCurriculum(appCtx *AppContext, userID int64) (*Curriculum, error)

// Línea 356: CAMBIAR llamada a DB dentro de la goroutine
data, fetchErr := fetchFunc(h.DB, userID)
// ↓ CAMBIAR A:
data, fetchErr := fetchFunc(appCtx.DB, userID) // Asumiendo que las funciones get* de curriculum se adaptan para recibir appCtx.DB

// Líneas 383-388: ACTUALIZAR llamadas a fetch
// Las funciones getEducation, getWorkExperience, etc. (de handlers_curriculum.go) deben aceptar (appCtx.DB, userID)
// O mejor, que acepten (appCtx *AppContext, userID int64) y accedan a appCtx.DB internamente.
// Ejemplo:
fetch(func(db *sql.DB, id int64) (interface{}, error) { return getEducation(db, id) }, &curriculum.Education)
// ↓ CAMBIAR A (si getEducation toma AppContext):
fetch(func(ctx *AppContext, id int64) (interface{}, error) { return getEducation(ctx, id) }, &curriculum.Education)
// O (si getEducation toma *sql.DB, pasar appCtx.DB):
fetch(func(db *sql.DB, id int64) (interface{}, error) { return getEducation(db, id) }, &curriculum.Education)
// Este último es el patrón actual de handlers_curriculum.go, por lo que solo el `h.DB` en la línea 356 necesita cambiar a `appCtx.DB`.

// Resto de la lógica MANTENER (sync.WaitGroup, error handling, type assertions)
```

**Líneas 403-410: Stubs de actualización de currículum - ELIMINAR COMENTADOS**

### Cambios Críticos Resumidos para handlers_profile.go:

1.  **Imports:** AGREGAR `customws_types`, `time`. ELIMINAR `encoding/json` donde no se parseen payloads directamente (ahora lo hace customws).
2.  **Structs locales:** Mover `DeleteCurriculumItemRequest` a `types.go` (posiblemente como `DeleteItemPayload`).
3.  **Firmas de Handlers:** Actualizar todas las firmas a `func handlerName(appCtx *AppContext, conn *customws.Connection[AppUserData], msg customws_types.ClientToServerMessage) error`.
4.  **Firmas Auxiliares:** `getBaseUserProfile` y `getCurriculum` deben tomar `appCtx *AppContext`.
5.  **Acceso a DB:** Cambiar `h.DB` a `appCtx.DB`.
6.  **Payload Parsing:** Usar `json.Unmarshal(msg.Payload, &reqStruct)`.
7.  **User Info:** `client.User.ID` → `conn.UserData.ID`, `client.User.RoleID` → `conn.UserData.RoleID`.
8.  **Respuestas:** Usar `conn.SendErrorNotification()` y `conn.SendMessage()` con `customws_types.ServerToClientMessage`.
9.  **Nuevos Tipos de Mensaje:** `MessageTypeEditProfileResponse`, `MessageTypeDeleteItemCurriculumResponse`.
10. **Estado Online:** `h.GetClient()` → `appCtx.ConnectionManager.IsUserConnected()`.
11. **Llamadas a `getContactStatus` y `getCurriculum`:** Actualizar para usar `appCtx`.
12. **Actualización de `client.User`:** Eliminar la lógica de `handleEditProfile` que actualizaba `client.User`.

### Esquema de BD utilizado por handlers_profile.go:
```sql
User (Id, FirstName, LastName, UserName, Email, Phone, Sex, DocId, NationalityId, Birthdate, Picture, DegreeId, UniversityId, RoleId, StatusAuthorizedId, Summary, Address, Github, Linkedin)
Nationality (Id, CountryName)
Degree (Id, DegreeName)
University (Id, Name)
Role (Id, Name)
-- Más tablas de Curriculum (Education, WorkExperience, etc. - llamadas vía getCurriculum)
Contact (User1Id, User2Id, Status) -- Usado por getContactStatus
```

### 7. `handlers_search.go` (Prioridad: MEDIA)

**Estado:** Migración completa requerida - Contiene handlers para obtener y editar perfiles, y eliminar items del currículum, además de funciones auxiliares clave.

#### Análisis línea por línea:

**Líneas 1-10: Package e imports**
```go
// CAMBIAR imports
package websocket

import (
	"database/sql"      // MANTENER
	"encoding/json"     // ELIMINAR - no necesario con customws
	"fmt"               // MANTENER
	"sync"              // MANTENER (usado en getCurriculum)

	"github.com/davidM20/micro-service-backend-go.git/internal/models" // MANTENER
	"github.com/davidM20/micro-service-backend-go.git/pkg/logger"      // MANTENER
	customws_types "github.com/davidM20/micro-service-backend-go.git/pkg/customws/types" // AGREGAR
	"time"                                                               // AGREGAR
)
```

**Líneas 12-26: Structs locales**
- `ProfileResponse` (Líneas 15-20): **ELIMINAR COMENTADO** - Ya se indica que se usa `MyProfileResponse` de `types.go`.
- `DeleteCurriculumItemRequest` (Líneas 23-26): **MANTENER** - Es usado por `handleDeleteItemCurriculum`. Podría moverse a `types.go` como `DeleteItemPayload` si no existe ya con ese nombre.

#### Función 1: handleGetProfile (Líneas 31-168) - **MIGRAR COMPLETAMENTE**

**Acción requerida:** Migrar función principal para obtener perfiles de usuario (propio o de otros).

**Cambios específicos:**

```go
// Línea 31: CAMBIAR firma
func (h *Hub) handleGetProfile(client *Client, payload interface{}) error
// ↓ CAMBIAR A:
func handleGetProfile(appCtx *AppContext, conn *customws.Connection[AppUserData], msg customws_types.ClientToServerMessage) error

// Líneas 32-40: SIMPLIFICAR parsing payload
var req ProfileRequestPayload
if msg.Payload != nil {
	if err := json.Unmarshal(msg.Payload, &req); err != nil {
2.  **Structs locales:** Mover `