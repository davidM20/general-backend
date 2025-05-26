# Guía del Cliente WebSocket para `customws` Genérico (TypeScript)

Este documento describe cómo un cliente WebSocket, preferiblemente escrito en TypeScript, debe interactuar con un servidor backend que utiliza el paquete `customws` genérico actualizado.

## 1. Conexión al Servidor WebSocket

El cliente debe establecer una conexión WebSocket estándar al endpoint expuesto por el servidor (ej. `ws://localhost:8082/ws` o `wss://yourdomain.com/ws`).

### 1.1. Autenticación

Con el protocolo genérico actualizado, la autenticación se realiza típicamente mediante un **token JWT en el header `Authorization`**. El método de validación depende de la implementación del callback `AuthenticateAndGetUserData` en el servidor.

```typescript
const authToken = "your-jwt-token-here"; // Obtener de tu sistema de autenticación
const socketUrl = `ws://localhost:8082/ws`;

let socket: WebSocket;

function connect() {
  // Nota: Los headers de WebSocket en el navegador son limitados
  // La autenticación generalmente se valida en el primer mensaje o mediante token en URL
  socket = new WebSocket(socketUrl);

  socket.onopen = (event) => {
    console.log("WebSocket Conectado! Evento:", event);
    // El servidor validará la autenticación mediante el callback AuthenticateAndGetUserData
  };

  socket.onclose = (event) => {
    console.log("WebSocket Desconectado. Código:", event.code, "Razón:", event.reason);
    // Intentar reconectar si es necesario, con backoff exponencial, etc.
  };

  socket.onerror = (error) => {
    console.error("Error de WebSocket:", error);
    // Manejar errores de conexión
  };

  socket.onmessage = (event) => {
    // Procesar mensajes entrantes del servidor (ver Sección 3)
    handleServerMessage(event.data);
  };
}

function disconnect() {
  if (socket && socket.readyState === WebSocket.OPEN) {
    socket.close(1000, "Cierre normal por el cliente");
  }
}

// Iniciar conexión
// connect();
```

## 2. Formato de Mensajes: Cliente → Servidor (Protocolo Genérico)

Todos los mensajes enviados por el cliente al servidor deben ser strings JSON que sigan la estructura `ClientToServerMessage` actualizada con tipos genéricos.

```typescript
// Definiciones de tipo actualizadas para el protocolo genérico
interface ClientToServerMessage {
  pid?: string;          // Opcional: ID de Proceso/Petición
  type: MessageType;     // Obligatorio: Tipo de mensaje genérico
  targetUserId?: number; // Opcional: Para mensajes peer-to-peer
  payload: any;          // Obligatorio: Contenido genérico (objeto JSON)
}

// MessageType genéricos actualizados
enum ClientMessageType {
  DATA_REQUEST = "data_request",        // Solicitud de datos genérica
  PRESENCE_UPDATE = "presence_update",  // Actualización de presencia
  CLIENT_ACK = "client_ack",           // Confirmación del cliente
  GENERIC_REQUEST = "generic_request"   // Solicitud genérica con respuesta esperada
}

// Payload genérico para data_request (compatible con map[string]interface{} de Go)
interface DataRequestPayload {
  action?: string;              // Acción a realizar (ej: "send_message", "get_profile")
  resource?: string;            // Recurso objetivo (ej: "chat", "file", "notification")
  data?: Record<string, any>;   // Datos específicos de la solicitud
  [key: string]: any;           // Permite campos adicionales para flexibilidad
}

// Ejemplo de envío de un mensaje de chat usando protocolo genérico
function sendChatMessage(text: string, targetUserId?: number): void {
  if (socket && socket.readyState === WebSocket.OPEN) {
    const chatPayload: DataRequestPayload = {
      action: "send_message",
      resource: "chat",
      data: {
        text: text,
        timestamp: new Date().toISOString()
      }
    };
    
    const message: ClientToServerMessage = {
      type: ClientMessageType.DATA_REQUEST,
      targetUserId: targetUserId, // Para P2P, especificar usuario destino
      payload: chatPayload,
      pid: generateUniquePID(), // Opcional, si esperas confirmación
    };
    
    socket.send(JSON.stringify(message));
    console.log("Mensaje de chat enviado:", message);
  } else {
    console.warn("Socket no conectado. No se pudo enviar el mensaje de chat.");
  }
}

// Ejemplo de upload de archivo usando protocolo genérico
function uploadFileChunk(fileName: string, chunkData: string, chunkNum: number): void {
  if (socket && socket.readyState === WebSocket.OPEN) {
    const filePayload: DataRequestPayload = {
      action: "upload_chunk",
      resource: "file",
      data: {
        fileName: fileName,
        chunkNum: chunkNum,
        data: chunkData // Base64 encoded
      }
    };
    
    const message: ClientToServerMessage = {
      type: ClientMessageType.DATA_REQUEST,
      payload: filePayload,
      pid: generateUniquePID(),
    };
    
    socket.send(JSON.stringify(message));
  }
}

// Helper para generar PIDs (ejemplo simple, usar UUID en producción)
function generateUniquePID(): string {
  return `client-${Date.now()}-${Math.random().toString(36).substring(2, 9)}`;
}
```

### 2.1. Campos del Mensaje del Cliente (Actualizados):
*   **`pid` (string, opcional)**: Identificador único generado por el cliente para este mensaje. Útil para correlacionar respuestas y confirmaciones.
*   **`type` (string, obligatorio)**: Tipo de mensaje genérico. Los principales son:
    *   `"data_request"`: Solicitud de datos genérica (reemplaza chat_request, job_search_request, etc.)
    *   `"presence_update"`: Actualización de estado de presencia
    *   `"client_ack"`: Confirmación de recepción de mensaje del servidor
    *   `"generic_request"`: Solicitud que espera respuesta específica
*   **`targetUserId` (number, opcional)**: **CLAVE PARA PEER-TO-PEER** - Especifica el ID del usuario destinatario para comunicación directa.
*   **`payload` (any, obligatorio)**: Objeto JSON genérico que contiene los datos. Para `data_request`, se recomienda usar la estructura `action/resource/data`.

## 3. Formato de Mensajes: Servidor → Cliente (Protocolo Genérico)

El cliente recibirá mensajes del servidor como strings JSON que siguen la estructura `ServerToClientMessage` actualizada.

```typescript
interface ServerToClientMessage {
  pid?: string;         // Opcional: ID del mensaje del servidor
  type: ServerMessageType; // Obligatorio: Tipo de mensaje genérico del servidor
  fromUserId?: number;  // Opcional: ID del usuario que originó el mensaje (importante para P2P)
  payload: any;         // Obligatorio: Contenido del mensaje
  error?: ErrorPayload; // Opcional: Detalles si es un mensaje de error
}

interface ErrorPayload {
  originalPid?: string;
  code?: number;
  message: string;
}

// MessageType genéricos del servidor
enum ServerMessageType {
  DATA_EVENT = "data_event",           // Evento de datos genérico (reemplaza chat_event, etc.)
  PRESENCE_EVENT = "presence_event",   // Evento de presencia
  SERVER_ACK = "server_ack",          // Confirmación del servidor
  GENERIC_RESPONSE = "generic_response", // Respuesta a generic_request
  ERROR_NOTIFICATION = "error_notification" // Notificación de error
}

function handleServerMessage(jsonData: string): void {
  try {
    const serverMsg: ServerToClientMessage = JSON.parse(jsonData);
    console.log("Mensaje recibido del servidor:", serverMsg);

    switch (serverMsg.type) {
      case ServerMessageType.DATA_EVENT:
        // Todos los eventos de datos llegan aquí (chat, archivos, notificaciones, etc.)
        console.log(`Evento de datos de usuario ${serverMsg.fromUserId}:`, serverMsg.payload);
        handleDataEvent(serverMsg);
        break;

      case ServerMessageType.PRESENCE_EVENT:
        // Eventos de presencia (online/offline/typing, etc.)
        console.log(`Actualización de presencia:`, serverMsg.payload);
        handlePresenceEvent(serverMsg);
        break;

      case ServerMessageType.SERVER_ACK:
        // Confirmación del servidor
        const ackData = serverMsg.payload as { acknowledgedPid: string, status?: string, error?: string };
        console.log(`ServerAck para PID ${ackData.acknowledgedPid}, Status: ${ackData.status}`);
        handleServerAck(ackData);
        break;

      case ServerMessageType.GENERIC_RESPONSE:
        if (serverMsg.pid) {
            console.log(`Respuesta genérica para PID ${serverMsg.pid}:`, serverMsg.payload);
            handleGenericResponse(serverMsg);
        }
        break;

      case ServerMessageType.ERROR_NOTIFICATION:
        console.error("Error del servidor:", serverMsg.error?.message);
        handleError(serverMsg.error);
        break;

      default:
        console.warn("Tipo de mensaje desconocido del servidor:", serverMsg.type);
    }

  } catch (e) {
    console.error("Error al procesar mensaje del servidor:", jsonData, e);
  }
}

function handleDataEvent(message: ServerToClientMessage): void {
  const payload = message.payload as any;
  
  // Identificar el tipo de datos por la estructura del payload
  if (payload.action === "send_message" && payload.resource === "chat") {
    // Es un mensaje de chat
    console.log("💬 Nuevo mensaje de chat:", payload.data);
    updateChatUI(message.fromUserId, payload.data);
  } else if (payload.action === "upload_complete" && payload.resource === "file") {
    // Es una notificación de archivo subido
    console.log("📁 Archivo subido:", payload.data);
    updateFileUI(payload.data);
  } else if (payload.resource === "notification") {
    // Es una notificación
    console.log("🔔 Nueva notificación:", payload.data);
    showNotification(payload.data);
  } else {
    // Evento genérico
    console.log("📦 Evento de datos genérico:", payload);
  }
}

function handlePresenceEvent(message: ServerToClientMessage): void {
  const payload = message.payload as any;
  console.log(`👤 Usuario ${payload.username || message.fromUserId} está ${payload.status}`);
  updatePresenceUI(message.fromUserId, payload.status);
}
```

## 4. 🔥 **Comunicación Peer-to-Peer (P2P) - ¿Cómo Funciona Realmente?**

### 4.1. **Arquitectura P2P en `customws`**

**⚠️ IMPORTANTE**: La comunicación "peer-to-peer" en `customws` **NO es P2P directo**. Los mensajes **SÍ pasan por el servidor** que actúa como un **intermediario inteligente** y **router central**.

#### **Flujo de Comunicación P2P:**

```
Cliente A ──1──> Servidor ──2──> Cliente B
   │                │              │
   │                │              │
 Envía          Valida,          Recibe
mensaje        procesa,         mensaje
P2P           guarda y          con info
              reenvía          del remitente
```

### 4.2. **Pasos Detallados del Flujo P2P**

#### **Paso 1: Cliente A envía mensaje P2P**
```typescript
// Cliente A quiere enviar mensaje privado al Cliente B (userID: 456)
const message: ClientToServerMessage = {
  type: "data_request",
  targetUserId: 456,  // ← CLAVE: Esto indica que es P2P
  payload: {
    action: "send_message",
    resource: "chat",
    data: {
      text: "Hola! Este es un mensaje privado",
      messageType: "private"
    }
  },
  pid: "client-msg-123"
};

socket.send(JSON.stringify(message));
```

#### **Paso 2: Servidor recibe y procesa el mensaje**
El servidor en el callback `ProcessClientMessage`:

1. **Recibe** el mensaje del Cliente A
2. **Valida** que el Cliente A tenga permisos para enviar mensajes P2P
3. **Verifica** que el `targetUserId` (456) sea válido y esté conectado
4. **Procesa** el mensaje (puede guardar en base de datos, aplicar filtros, etc.)
5. **Enriquece** el mensaje con metadata del servidor (timestamp, messageId, etc.)
6. **Reenvía** el mensaje al Cliente B usando `SendMessageToUser(456, forwardMessage)`

```go
// En el servidor Go (ProcessClientMessage)
func handlePeerToPeerMessage(conn *customws.Connection[MyUserData], msg types.ClientToServerMessage) error {
    if msg.TargetUserID == 0 {
        return errors.New("targetUserId requerido para mensajes P2P")
    }
    
    // 1. Validaciones de seguridad
    if !hasPermissionToSendP2P(conn.UserData) {
        return errors.New("sin permisos para enviar mensajes P2P")
    }
    
    // 2. Procesar y enriquecer mensaje
    enrichedPayload := map[string]interface{}{
        "originalPayload": msg.Payload,
        "senderName":     conn.UserData.Username,
        "timestamp":      time.Now(),
        "messageId":      uuid.NewString(),
    }
    
    // 3. Guardar en base de datos (opcional)
    if err := savePeerMessage(conn.ID, msg.TargetUserID, enrichedPayload); err != nil {
        return err
    }
    
    // 4. Reenviar al destinatario
    forwardMessage := types.ServerToClientMessage{
        Type:       types.MessageTypeDataEvent,
        FromUserID: conn.ID,  // ← Importante: el receptor sabe quién lo envió
        Payload:    enrichedPayload,
    }
    
    // ← AQUÍ ES LA MAGIA: SendMessageToUser para P2P
    if err := conn.Manager.SendMessageToUser(msg.TargetUserID, forwardMessage); err != nil {
        log.Printf("Error reenviando mensaje P2P: %v", err)
        return err
    }
    
    // 5. Confirmar al remitente
    if msg.PID != "" {
        conn.SendServerAck(msg.PID, "delivered", nil)
    }
    
    return nil
}
```

#### **Paso 3: Cliente B recibe el mensaje**
```typescript
// Cliente B recibe el mensaje en handleServerMessage
function handleDataEvent(message: ServerToClientMessage): void {
  if (message.fromUserId && message.fromUserId !== myUserId) {
    // Es un mensaje P2P de otro usuario
    console.log(`💬 Mensaje P2P de usuario ${message.fromUserId}:`, message.payload);
    
    const payload = message.payload as any;
    displayP2PMessage({
      senderId: message.fromUserId,
      senderName: payload.senderName,
      text: payload.originalPayload.data.text,
      timestamp: payload.timestamp,
      messageId: payload.messageId
    });
  }
}
```

### 4.3. **Ventajas de P2P "Mediado por Servidor"**

#### **✅ Seguridad y Control**
- **Validación**: El servidor puede validar permisos antes de reenviar
- **Filtrado**: Puede aplicar filtros de contenido, anti-spam, etc.
- **Autenticación**: Garantiza que los mensajes vienen de usuarios autenticados

#### **✅ Persistencia y Confiabilidad**
- **Guardado**: Los mensajes se pueden guardar en base de datos
- **Historial**: Permite recuperar conversaciones históricas
- **Delivery Status**: El servidor puede confirmar entrega al remitente

#### **✅ Funcionalidades Avanzadas**
- **Notificaciones Push**: Si el destinatario está offline, se puede enviar push notification
- **Encriptación**: El servidor puede manejar encriptación/desencriptación
- **Moderación**: Permite moderar contenido inapropiado
- **Analytics**: Puede generar métricas de uso

#### **✅ Escalabilidad**
- **Load Balancing**: Funciona con múltiples servidores (usando Redis, etc.)
- **Caching**: El servidor puede cachear mensajes frecuentes
- **Rate Limiting**: Control de velocidad por usuario

### 4.4. **Ejemplo Completo de Implementación P2P**

#### **Cliente: Envío de mensaje P2P**
```typescript
class P2PMessaging {
  constructor(private wsClient: WebSocketClient) {}

  // Enviar mensaje P2P de chat
  async sendPrivateMessage(targetUserId: number, text: string): Promise<void> {
    const message: ClientToServerMessage = {
      type: "data_request",
      targetUserId: targetUserId,
      payload: {
        action: "send_message",
        resource: "chat",
        data: {
          text: text,
          messageType: "private",
          clientTimestamp: Date.now()
        }
      },
      pid: generateUniquePID()
    };

    try {
      const ackResponse = await this.wsClient.sendMessageWithAck(message);
      console.log("✅ Mensaje P2P enviado con confirmación:", ackResponse);
    } catch (error) {
      console.error("❌ Error enviando mensaje P2P:", error);
    }
  }

  // Enviar invitación de colaboración P2P
  async sendCollaborationInvite(targetUserId: number, projectData: any): Promise<void> {
    const message: ClientToServerMessage = {
      type: "data_request",
      targetUserId: targetUserId,
      payload: {
        action: "send_invitation",
        resource: "collaboration",
        data: {
          projectId: projectData.id,
          projectName: projectData.name,
          role: "editor",
          message: "¿Te gustaría colaborar en este proyecto?",
          invitationExpiry: Date.now() + (24 * 60 * 60 * 1000) // 24 horas
        }
      }
    };

    await this.wsClient.sendMessage(message);
  }

  // Iniciar llamada de video P2P
  async initiateVideoCall(targetUserId: number, sdpOffer: string): Promise<void> {
    const message: ClientToServerMessage = {
      type: "data_request",
      targetUserId: targetUserId,
      payload: {
        action: "initiate_call",
        resource: "webrtc",
        data: {
          callType: "video",
          sdpOffer: sdpOffer,
          callId: generateUniquePID()
        }
      }
    };

    await this.wsClient.sendMessage(message);
  }
}
```

#### **Cliente: Recepción de mensajes P2P**
```typescript
function handleP2PMessages(message: ServerToClientMessage): void {
  if (!message.fromUserId) return; // No es P2P

  const payload = message.payload as any;
  const senderId = message.fromUserId;

  switch (payload.action) {
    case "send_message":
      if (payload.resource === "chat") {
        displayPrivateMessage({
          senderId: senderId,
          senderName: payload.senderName,
          text: payload.data.text,
          timestamp: payload.timestamp,
          messageId: payload.messageId
        });
      }
      break;

    case "send_invitation":
      if (payload.resource === "collaboration") {
        showCollaborationInvitation({
          senderId: senderId,
          projectName: payload.data.projectName,
          role: payload.data.role,
          message: payload.data.message
        });
      }
      break;

    case "initiate_call":
      if (payload.resource === "webrtc") {
        handleIncomingVideoCall({
          callerId: senderId,
          callId: payload.data.callId,
          sdpOffer: payload.data.sdpOffer
        });
      }
      break;

    default:
      console.log("📦 Mensaje P2P genérico:", payload);
  }
}
```

### 4.5. **P2P vs Broadcasting - Comparación**

| Aspecto | **P2P (targetUserId)** | **Broadcasting (sin targetUserId)** |
|---------|------------------------|-------------------------------------|
| **Destinatarios** | Un usuario específico | Todos los usuarios conectados |
| **Método servidor** | `SendMessageToUser(userID, msg)` | `BroadcastToAll(msg)` |
| **Privacidad** | ✅ Privado | ❌ Público |
| **Uso de red** | ✅ Mínimo | ❌ Alto |
| **Campo `fromUserId`** | ✅ Siempre presente | ✅ Opcional |
| **Casos de uso** | Chat privado, notificaciones personales, colaboración directa | Anuncios globales, actualizaciones de sistema |

### 4.6. **Consideraciones de Seguridad P2P**

#### **Validaciones Recomendadas en el Servidor:**
```go
func validateP2PMessage(conn *customws.Connection[MyUserData], targetUserID int64) error {
    // 1. Verificar que el usuario tenga permisos P2P
    if !hasP2PPermissions(conn.UserData.Roles) {
        return errors.New("sin permisos para mensajes P2P")
    }
    
    // 2. Verificar que el destinatario existe y está activo
    if !isValidActiveUser(targetUserID) {
        return errors.New("usuario destinatario no válido")
    }
    
    // 3. Verificar relaciones (ej: amistad, mismo workspace)
    if !canSendMessageTo(conn.ID, targetUserID) {
        return errors.New("no autorizado para enviar mensaje a este usuario")
    }
    
    // 4. Rate limiting específico para P2P
    if !checkP2PRateLimit(conn.ID) {
        return errors.New("límite de mensajes P2P excedido")
    }
    
    return nil
}
```

En resumen, el P2P en `customws` es **P2P mediado por servidor** que ofrece todas las ventajas de seguridad, persistencia y control de un sistema centralizado, mientras proporciona la experiencia de comunicación directa entre usuarios.

## 5. Envío de Confirmaciones (`ClientAck`) al Servidor

El protocolo de confirmaciones permanece igual, pero ahora usa los tipos genéricos:

```typescript
interface AckPayload {
  acknowledgedPid: string; // PID del mensaje del servidor que se está confirmando
  status?: string;         // Ej: "received", "processed", "read"
  error?: string;          // Si hubo un error al procesar el mensaje original en el cliente
}

function sendClientAck(acknowledgedServerPid: string, status: string, error?: string): void {
  if (socket && socket.readyState === WebSocket.OPEN) {
    const ackPayload: AckPayload = {
      acknowledgedPid: acknowledgedServerPid,
      status: status,
    };
    if (error) {
      ackPayload.error = error;
    }
    const message: ClientToServerMessage = {
      type: "client_ack",  // Tipo genérico actualizado
      payload: ackPayload,
      pid: generateUniquePID(),
    };
    socket.send(JSON.stringify(message));
    console.log("ClientAck enviado para PID del servidor:", acknowledgedServerPid);
  }
}
```

## 6. Envío de Respuestas a Solicitudes Específicas del Servidor

Con el protocolo genérico, las respuestas a solicitudes del servidor siguen el mismo patrón pero con tipos actualizados:

```typescript
// Ejemplo: El servidor envió una solicitud con PID "server-request-123" y tipo "generic_request"
function sendGenericResponse(originalServerPid: string, responseData: any): void {
  if (socket && socket.readyState === WebSocket.OPEN) {
    const responseMessage: ClientToServerMessage = {
      pid: originalServerPid, // ¡Importante! Usar el PID de la solicitud del servidor
      type: "generic_response", // Tipo genérico de respuesta
      payload: responseData
    };

    socket.send(JSON.stringify(responseMessage));
    console.log("Respuesta genérica enviada:", responseMessage);
  }
}

// Ejemplo de manejo en handleServerMessage:
if (serverMsg.type === "generic_request" && serverMsg.pid) {
  // El servidor está pidiendo información específica
  const requestPayload = serverMsg.payload as any;
  
  if (requestPayload.action === "get_client_info") {
    const clientInfo = {
      browserInfo: navigator.userAgent,
      screenResolution: `${screen.width}x${screen.height}`,
      timezone: Intl.DateTimeFormat().resolvedOptions().timeZone,
      timestamp: Date.now()
    };
    sendGenericResponse(serverMsg.pid, clientInfo);
  }
}
```

## 7. Manteniendo la Conexión Activa

El manejo de Ping/Pong permanece igual que en la versión anterior:

*   **No necesitas implementar el envío de Pings desde el cliente.**
*   **No necesitas implementar el envío explícito de Pongs**; la biblioteca del cliente generalmente lo hace.
*   El servidor desconectará al cliente si no recibe un Pong dentro del `PongWait` configurado.

## 8. Ejemplos de Uso Avanzados con Protocolo Genérico

### 8.1. Sistema de Chat Completo
```typescript
class ChatSystem {
  constructor(private wsClient: WebSocketClient) {}

  // Enviar mensaje público
  async sendPublicMessage(text: string): Promise<void> {
    await this.wsClient.sendMessage({
      type: "data_request",
      payload: {
        action: "send_message",
        resource: "chat",
        data: { text, messageType: "public" }
      }
    });
  }

  // Enviar mensaje privado (P2P)
  async sendPrivateMessage(targetUserId: number, text: string): Promise<void> {
    await this.wsClient.sendMessage({
      type: "data_request",
      targetUserId: targetUserId,
      payload: {
        action: "send_message",
        resource: "chat",
        data: { text, messageType: "private" }
      }
    });
  }

  // Obtener historial
  async getChatHistory(chatId: string): Promise<any> {
    return await this.wsClient.sendGenericRequest({
      action: "get_history",
      resource: "chat",
      data: { chatId, limit: 50 }
    });
  }
}
```

### 8.2. Sistema de Notificaciones
```typescript
class NotificationSystem {
  constructor(private wsClient: WebSocketClient) {}

  // Enviar notificación a usuarios específicos
  async sendNotification(targetUsers: number[], notification: any): Promise<void> {
    await this.wsClient.sendMessage({
      type: "data_request",
      payload: {
        action: "send_notification",
        resource: "notification",
        data: {
          targetUsers,
          ...notification
        }
      }
    });
  }

  // Marcar notificación como leída
  async markAsRead(notificationId: string): Promise<void> {
    await this.wsClient.sendMessage({
      type: "data_request",
      payload: {
        action: "mark_read",
        resource: "notification",
        data: { notificationId }
      }
    });
  }
}
```

## 9. Cierre de Conexión

El cierre de conexión permanece igual:

```typescript
if (socket && socket.readyState === WebSocket.OPEN) {
  socket.close(1000, "Usuario cerró la sesión"); // Código 1000 es cierre normal
}
```

---

Esta guía actualizada proporciona una base completa para desarrollar clientes WebSocket que interactúen efectivamente con el backend `customws` genérico, con especial énfasis en la comprensión correcta del modelo de comunicación peer-to-peer mediado por servidor.

## 10. 👤 **Sistema de Presencia - ¿Qué es y para qué sirve?**

### 10.1. **¿Qué es el Sistema de Presencia?**

El **sistema de presencia** es una funcionalidad que permite **monitorear y comunicar el estado en tiempo real** de los usuarios conectados al sistema WebSocket. Incluye información como:

- **Estado de conexión**: online, offline, away, busy
- **Actividad específica**: typing, viewing, editing, idle
- **Ubicación virtual**: en qué página/sección/documento está el usuario
- **Información contextual**: último mensaje visto, progreso en tareas, etc.

### 10.2. **Objetivo Principal**

El objetivo es crear **experiencias colaborativas fluidas** donde los usuarios puedan:

1. **Ver quién está online** en tiempo real
2. **Saber qué están haciendo** otros usuarios
3. **Coordinar actividades** de manera natural
4. **Evitar conflictos** en trabajo colaborativo
5. **Mejorar la comunicación** con indicadores visuales

### 10.3. **¿Cómo Funciona en `customws`?**

#### **Flujo de Presencia:**
```
Cliente A ──presence_update──> Servidor ──presence_event──> Todos los Clientes
    │                              │                              │
Cambia estado                 Procesa y                    Actualizan UI
(ej: "typing")              distribuye cambio             mostrando estado
```

#### **Tipos de Mensajes de Presencia:**
```typescript
// Cliente → Servidor: Actualizar mi presencia
{
  type: "presence_update",
  payload: {
    status: "typing",
    targetUserId: 456, // Opcional: presencia específica hacia alguien
    context: {
      location: "chat-room-general",
      activity: "writing_message"
    }
  }
}

// Servidor → Clientes: Notificar cambio de presencia
{
  type: "presence_event",
  fromUserId: 123,
  payload: {
    userId: 123,
    username: "juan_dev",
    status: "typing",
    context: {
      location: "chat-room-general",
      activity: "writing_message"
    },
    timestamp: "2024-01-15T10:30:00Z"
  }
}
```

### 10.4. **🎯 Escenarios Ideales de Uso**

#### **💬 1. Aplicaciones de Chat/Mensajería**
```typescript
// Mostrar "Juan está escribiendo..."
client.sendPresenceUpdate("typing", targetUserId);

setTimeout(() => {
  client.sendPresenceUpdate("online", targetUserId);
}, 3000); // Parar después de 3 segundos
```

**Casos específicos:**
- Indicador "escribiendo..." en conversaciones
- Estado "leyendo mensajes" / "mensaje visto"
- Mostrar quién está activo en el chat

#### **📝 2. Colaboración en Documentos**
```typescript
// Usuario editando una sección específica
client.sendPresenceUpdate("editing", null, {
  documentId: "doc_123",
  section: "paragraph_5",
  cursor_position: 245
});

// Usuario viendo documento sin editar
client.sendPresenceUpdate("viewing", null, {
  documentId: "doc_123",
  scroll_position: "60%"
});
```

**Casos específicos:**
- Mostrar cursores de otros usuarios en tiempo real
- Evitar edición simultánea de la misma sección
- Indicar quién está revisando qué parte

#### **🎮 3. Aplicaciones Interactivas/Gaming**
```typescript
// Usuario en lobby esperando partida
client.sendPresenceUpdate("waiting", null, {
  game: "chess",
  lobby: "beginner_room",
  seeking_opponent: true
});

// Usuario en partida activa
client.sendPresenceUpdate("playing", opponentUserId, {
  game: "chess",
  match_id: "match_789",
  turn: "opponent"
});
```

#### **👥 4. Plataformas de E-learning**
```typescript
// Estudiante viendo clase en vivo
client.sendPresenceUpdate("attending", null, {
  class_id: "math_101",
  video_timestamp: "15:23",
  attention_level: "focused"
});

// Profesor dando clase
client.sendPresenceUpdate("teaching", null, {
  class_id: "math_101",
  current_slide: 15,
  students_count: 24
});
```

#### **💼 5. Aplicaciones de Trabajo Remoto**
```typescript
// Desarrollador trabajando en feature
client.sendPresenceUpdate("coding", null, {
  project: "webapp_v2",
  branch: "feature/user-auth",
  file: "auth.service.ts",
  last_commit: "2h ago"
});

// En reunión virtual
client.sendPresenceUpdate("in_meeting", null, {
  meeting_id: "standup_daily",
  role: "participant",
  camera: true,
  microphone: false
});
```

### 10.5. **Implementación Completa del Cliente**

#### **Clase de Gestión de Presencia:**
```typescript
class PresenceManager {
  private currentStatus: string = 'offline';
  private presenceTimer: number | null = null;
  private heartbeatInterval: number = 30000; // 30 segundos

  constructor(private wsClient: WebSocketClient) {
    this.setupHeartbeat();
    this.setupVisibilityTracking();
  }

  // === MÉTODOS PRINCIPALES ===

  // Actualizar estado general
  updateStatus(status: 'online' | 'away' | 'busy' | 'offline'): void {
    this.currentStatus = status;
    this.wsClient.sendPresenceUpdate(status);
    console.log(`🟢 Estado actualizado a: ${status}`);
  }

  // Actualizar estado con contexto específico
  updateStatusWithContext(status: string, context: any, targetUserId?: number): void {
    this.wsClient.sendMessage({
      type: 'presence_update',
      targetUserId,
      payload: {
        status,
        context,
        timestamp: Date.now()
      }
    });
  }

  // === MÉTODOS DE CONVENIENCIA ===

  // Chat: Usuario escribiendo
  startTyping(chatId: string, targetUserId?: number): void {
    this.updateStatusWithContext('typing', { chatId }, targetUserId);
    
    // Auto-parar después de 5 segundos
    if (this.presenceTimer) clearTimeout(this.presenceTimer);
    this.presenceTimer = setTimeout(() => {
      this.stopTyping(chatId, targetUserId);
    }, 5000);
  }

  stopTyping(chatId: string, targetUserId?: number): void {
    this.updateStatusWithContext('online', { chatId }, targetUserId);
    if (this.presenceTimer) {
      clearTimeout(this.presenceTimer);
      this.presenceTimer = null;
    }
  }

  // Documento: Editando sección específica
  startEditing(documentId: string, section: string, cursorPosition?: number): void {
    this.updateStatusWithContext('editing', {
      documentId,
      section,
      cursorPosition,
      startedAt: Date.now()
    });
  }

  stopEditing(documentId: string): void {
    this.updateStatusWithContext('viewing', {
      documentId,
      stoppedEditingAt: Date.now()
    });
  }

  // Ubicación: Cambio de página/sección
  updateLocation(page: string, section?: string): void {
    this.updateStatusWithContext(this.currentStatus, {
      location: page,
      section,
      navigatedAt: Date.now()
    });
  }

  // === HEARTBEAT Y DETECCIÓN DE INACTIVIDAD ===

  private setupHeartbeat(): void {
    // Enviar heartbeat periódico para mantener presencia
    setInterval(() => {
      if (this.currentStatus !== 'offline') {
        this.updateStatus(this.currentStatus);
      }
    }, this.heartbeatInterval);
  }

  private setupVisibilityTracking(): void {
    // Detectar cuando el usuario cambia de tab/ventana
    document.addEventListener('visibilitychange', () => {
      if (document.hidden) {
        this.updateStatus('away');
      } else {
        this.updateStatus('online');
      }
    });

    // Detectar inactividad del mouse/teclado
    let inactivityTimer: number;
    const resetInactivityTimer = () => {
      clearTimeout(inactivityTimer);
      if (this.currentStatus === 'away') {
        this.updateStatus('online');
      }
      
      inactivityTimer = setTimeout(() => {
        this.updateStatus('away');
      }, 300000); // 5 minutos de inactividad
    };

    document.addEventListener('mousemove', resetInactivityTimer);
    document.addEventListener('keypress', resetInactivityTimer);
    resetInactivityTimer(); // Inicializar
  }

  // === CLEANUP ===
  destroy(): void {
    if (this.presenceTimer) clearTimeout(this.presenceTimer);
    this.updateStatus('offline');
  }
}
```

#### **Manejo de Eventos de Presencia:**
```typescript
// Configurar callback para eventos de presencia
const presenceCallbacks: WSClientCallbacks = {
  onPresenceEvent: (message) => {
    const payload = message.payload as any;
    const userId = message.fromUserId;
    
    console.log(`👤 ${payload.username} está ${payload.status}`);
    
    // Actualizar UI según el tipo de presencia
    switch (payload.status) {
      case 'typing':
        showTypingIndicator(userId, payload.context?.chatId);
        break;
        
      case 'editing':
        showEditingCursor(userId, payload.context?.documentId, payload.context?.section);
        break;
        
      case 'online':
        updateUserStatus(userId, 'online');
        hideTypingIndicator(userId);
        break;
        
      case 'away':
        updateUserStatus(userId, 'away');
        break;
        
      case 'offline':
        updateUserStatus(userId, 'offline');
        removeAllIndicators(userId);
        break;
    }
    
    // Actualizar lista de usuarios online
    updateOnlineUsersList();
  }
};

// Funciones de UI (ejemplos)
function showTypingIndicator(userId: number, chatId?: string): void {
  const indicator = document.getElementById(`typing-${userId}`);
  if (indicator) {
    indicator.style.display = 'block';
    indicator.textContent = 'escribiendo...';
  }
}

function showEditingCursor(userId: number, docId?: string, section?: string): void {
  const cursor = document.getElementById(`cursor-${userId}`);
  if (cursor && section) {
    cursor.style.display = 'block';
    // Posicionar cursor en la sección correspondiente
    const sectionElement = document.getElementById(section);
    if (sectionElement) {
      cursor.style.top = sectionElement.offsetTop + 'px';
    }
  }
}
```

### 10.6. **✅ Ventajas del Sistema de Presencia**

#### **🚀 Experiencia de Usuario**
- **Inmediatez**: Los usuarios ven cambios en tiempo real
- **Coordinación**: Facilita trabajo colaborativo sin conflictos
- **Engagement**: Aumenta la sensación de "estar acompañado"
- **Feedback Visual**: Indicadores claros de actividad

#### **💼 Casos de Negocio**
- **Productividad**: Menos interrupciones y mejor coordinación
- **Retención**: Los usuarios se sienten más conectados
- **Soporte**: Agentes pueden ver cuando clientes están activos
- **Analytics**: Métricas de engagement y patrones de uso

#### **🔧 Técnicas**
- **Eficiencia**: Solo se envían cambios (no estado completo)
- **Escalabilidad**: El servidor puede filtrar por contexto
- **Flexibilidad**: Sistema genérico adaptable a cualquier caso

### 10.7. **❌ Desventajas y Consideraciones**

#### **📊 Consumo de Recursos**
```typescript
// Problema: Demasiadas actualizaciones
// ❌ MAL: Actualizar presencia en cada keystroke
document.addEventListener('keydown', () => {
  client.sendPresenceUpdate('typing'); // ¡Spam al servidor!
});

// ✅ BIEN: Debounce y throttling
let typingTimeout: number;
document.addEventListener('keydown', () => {
  if (!typingTimeout) {
    client.sendPresenceUpdate('typing');
  }
  
  clearTimeout(typingTimeout);
  typingTimeout = setTimeout(() => {
    client.sendPresenceUpdate('online');
    typingTimeout = null;
  }, 2000);
});
```

**Problemas comunes:**
- **Spam de mensajes**: Actualizaciones demasiado frecuentes
- **Ancho de banda**: En aplicaciones con muchos usuarios
- **Procesamiento**: El servidor debe manejar muchos eventos
- **Batería**: En dispositivos móviles puede agotar la batería

#### **🔒 Privacidad y Seguridad**
- **Información personal**: Puede revelar patrones de actividad
- **Stalking digital**: Usuarios pueden sentirse "vigilados"
- **GDPR/Compliance**: Necesita consentimiento explícito
- **Configurabilidad**: Los usuarios deben poder desactivarlo

#### **🐛 Complejidad Técnica**
- **Estados inconsistentes**: Sincronizar estado entre múltiples dispositivos
- **Conexiones perdidas**: Manejar usuarios que se desconectan abruptamente
- **Escalabilidad**: Con miles de usuarios simultáneos
- **Debugging**: Difícil debuggear problemas de presencia intermitentes

### 10.8. **🛠️ Mejores Prácticas**

#### **Optimización de Performance:**
```typescript
class OptimizedPresenceManager {
  private lastSentStatus: string = '';
  private debounceTimer: number = 0;
  private readonly DEBOUNCE_DELAY = 1000; // 1 segundo

  // Solo enviar si realmente cambió el estado
  updateStatus(newStatus: string): void {
    if (this.lastSentStatus === newStatus) return;
    
    clearTimeout(this.debounceTimer);
    this.debounceTimer = setTimeout(() => {
      this.wsClient.sendPresenceUpdate(newStatus);
      this.lastSentStatus = newStatus;
    }, this.DEBOUNCE_DELAY);
  }

  // Batch updates para múltiples cambios
  batchUpdate(updates: Array<{status: string, context?: any}>): void {
    // Agrupar múltiples updates en uno solo
    const latestUpdate = updates[updates.length - 1];
    this.updateStatus(latestUpdate.status);
  }
}
```

#### **Configuración de Privacidad:**
```typescript
interface PresenceSettings {
  enabled: boolean;
  showOnlineStatus: boolean;
  showActivity: boolean;
  showLocation: boolean;
  allowedUsers: number[]; // Lista blanca de usuarios que pueden ver mi presencia
}

class PrivacyAwarePresence {
  constructor(private settings: PresenceSettings) {}

  updateStatus(status: string, context?: any): void {
    if (!this.settings.enabled) return;
    
    const filteredContext = this.filterContext(context);
    this.wsClient.sendPresenceUpdate(status, undefined, filteredContext);
  }

  private filterContext(context: any): any {
    if (!this.settings.showActivity) {
      delete context?.activity;
    }
    if (!this.settings.showLocation) {
      delete context?.location;
    }
    return context;
  }
}
```

### 10.9. **Comparación: Presencia vs Sin Presencia**

| Aspecto | **Con Presencia** | **Sin Presencia** |
|---------|------------------|-------------------|
| **UX Colaborativo** | ✅ Excelente - usuarios coordinados | ❌ Confuso - ediciones simultáneas |
| **Engagement** | ✅ Alto - sensación de comunidad | ⚠️ Medio - interacción menos fluida |
| **Complejidad** | ❌ Alta - más código y lógica | ✅ Baja - implementación simple |
| **Performance** | ❌ Mayor uso de recursos | ✅ Mínimo overhead |
| **Privacidad** | ❌ Expone información personal | ✅ Máxima privacidad |
| **Debugging** | ❌ Más puntos de falla | ✅ Menos superficie de error |
| **Escalabilidad** | ❌ Limitada por eventos frecuentes | ✅ Ilimitada escalabilidad |

### 10.10. **¿Cuándo Usar Presencia?**

#### **✅ SÍ usar presencia cuando:**
- Aplicación colaborativa (docs, chat, gaming)
- Base de usuarios < 10,000 simultáneos
- UX es prioritario sobre performance
- Usuarios esperan ver actividad de otros
- Hay recursos para implementar bien

#### **❌ NO usar presencia cuando:**
- Aplicación principalmente individual
- Crítico el performance/ancho de banda
- Usuarios priorizan privacidad
- Equipo no tiene capacidad para mantener la complejidad
- Presupuesto limitado de infraestructura

---

El sistema de presencia es una **herramienta poderosa pero compleja** que puede transformar aplicaciones simples en experiencias colaborativas inmersivas, pero requiere **implementación cuidadosa** y **consideración de trade-offs** importantes. 