 \
# Documentación: Funcionalidad de Guardado de Mensajes de Chat

Esta documentación describe la implementación del backend para guardar mensajes de chat en la base de datos y notificar al cliente.

## Flujo General

1.  **Recepción del Mensaje (Handler)**:
    *   Un handler de WebSocket (ej. `HandleSendChatMessage`) recibe un mensaje del cliente con un payload que contiene los datos del nuevo mensaje (ej. `chatId`, `text`, `mediaId`, `responseTo`).
    *   El handler genera un ID único para el mensaje (ej. UUID).

2.  **Procesamiento del Servicio (Capa de Servicio)**:
    *   El handler llama a la función `services.ProcessAndSaveChatMessage`.
    *   Esta función recibe:
        *   `userID`: ID del usuario remitente.
        *   `payload`: Un `map[string]interface{}` con los datos del mensaje extraídos por el handler.
        *   `messageID`: El ID único generado por el handler.
    *   Responsabilidades de `ProcessAndSaveChatMessage`:
        *   Validar la presencia de campos requeridos (`chatId`, y `text` o `mediaId`).
        *   Determinar el `TypeMessageId` (ej. 1 para texto, 2 para media).
        *   Crear una instancia de `models.Message` con todos los datos, incluyendo la fecha actual (`time.Now().UTC()`) y el estado inicial (`queries.StatusMessageSent`).
        *   Llamar a `queries.CreateMessage` para persistir el mensaje en la base de datos.
        *   Devolver el `models.Message` guardado (con el ID confirmado por la BD) o un error.

3.  **Acceso a Datos (Capa de Queries)**:
    *   La función `queries.CreateMessage` recibe la instancia de `models.Message`.
    *   Responsabilidades de `CreateMessage`:
        *   Asegurar que el mensaje tenga un ID (genera un UUID si está vacío).
        *   Establecer valores por defecto si no se proporcionan (`Date`, `TypeMessageId`, `StatusMessage`).
        *   Manejar campos opcionales que pueden ser `NULL` en la BD (ej. `MediaId`, `ResponseTo`) utilizando `sql.NullString`.
        *   Construir y ejecutar la sentencia SQL `INSERT` en la tabla `Message`.
        *   Devolver el ID del mensaje insertado o un error.
    *   Se han definido constantes para los estados de los mensajes (`StatusMessageSent`, `StatusMessageDelivered`, etc.) en `queries.go`.

4.  **Respuesta al Cliente (Handler)**:
    *   Una vez que `ProcessAndSaveChatMessage` devuelve el mensaje guardado (o un error):
        *   Si hay un error, el handler envía una notificación de error al cliente.
        *   Si el guardado es exitoso, el handler:
            *   Envía el mensaje al destinatario (si está online).
            *   Envía una confirmación (ACK) al remitente, incluyendo los datos del mensaje guardado para que pueda actualizar su UI.

## Estructuras de Datos Clave

*   **`models.Message`**: Define la estructura del mensaje tal como se mapea a la tabla `Message` en la base de datos.
    *   Campos importantes: `Id`, `TypeMessageId`, `Text`, `MediaId`, `Date`, `StatusMessage`, `UserId`, `ChatId`, `ResponseTo`.
    *   `ChatIdGroup` existe en el esquema SQL pero no en `models.Message` actualmente; se inserta como `NULL` por defecto.
*   **Payload del Cliente (WebSocket)**: Se espera un JSON con campos como:
    *   `chatId` (string, requerido)
    *   `text` (string, opcional si `mediaId` presente)
    *   `mediaId` (string, opcional si `text` presente)
    *   `responseTo` (string, opcional)
    *   `clientTempId` (string, opcional, para que el cliente pueda rastrear el mensaje antes de recibir el ID final del servidor)

## Archivos Modificados/Creados

*   `backend/internal/db/queries/queries.go`:
    *   Actualizada la función `CreateMessage`.
    *   Añadidas constantes de estado de mensaje.
    *   Corregida `CreateMessageFromChatParams`.
*   `backend/internal/websocket/services/chat_service.go`:
    *   Creada la función `ProcessAndSaveChatMessage`.
    *   Ajustada para recibir `map[string]interface{}` como payload.
*   `backend/internal/models/models.go`: Revisado para entender la estructura de `Message`. (No se modificó en este flujo, pero es relevante).
*   `backend/schema.sql`: Revisado para entender la estructura de la tabla `Message`. (No se modificó en este flujo).

## Prompt para IA (Recreación de la Funcionalidad)

```
Eres un asistente de IA experto en Go. Necesito implementar una funcionalidad para guardar mensajes de chat en una aplicación existente.

**Contexto del Proyecto:**

*   Backend en Go.
*   Uso de WebSocket para comunicación en tiempo real (`customws` y `wsmodels`).
*   Arquitectura en capas: handlers, services, queries (acceso a BD).
*   Base de datos SQL (MySQL, la sintaxis SQL debe ser compatible).
*   Ya existe un archivo `backend/internal/db/queries/queries.go` para consultas SQL.
*   Ya existe un archivo `backend/internal/websocket/services/chat_service.go` para la lógica de negocio del chat.
*   Ya existe un archivo `backend/internal/models/models.go` con las estructuras de datos que mapean a tablas de BD.
*   Ya existe un archivo `backend/internal/websocket/wsmodels/types.go` para modelos específicos de WebSocket.
*   Logger (`pkg/logger`) disponible.

**Tarea Principal:**

Modificar/Crear el código necesario para que cuando un handler de WebSocket reciba un nuevo mensaje de chat, este se guarde en la base de datos.

**Pasos Detallados a Implementar:**

1.  **Constantes de Estado del Mensaje (en `queries.go`):**
    *   Define constantes para los estados de los mensajes: `StatusMessageSent` (1), `StatusMessageDelivered` (2), `StatusMessageRead` (3), `StatusMessageError` (4), `StatusMessagePending` (0), `StatusMessageNotSentYet` (-1).

2.  **Función `CreateMessage` (en `queries.go`):**
    *   Debe existir o ser actualizada.
    *   Firma: `func CreateMessage(db *sql.DB, msg *models.Message) (string, error)`
    *   Debe generar un UUID para `msg.Id` si está vacío.
    *   Debe establecer `msg.Date` a `time.Now().UTC()` si es zero.
    *   Debe establecer `msg.TypeMessageId` a 1 (texto) por defecto si es 0.
    *   Debe establecer `msg.StatusMessage` a `StatusMessageSent` por defecto si es 0.
    *   Debe manejar `msg.MediaId` y `msg.ResponseTo` como `sql.NullString` para la inserción, ya que pueden ser opcionales/nulos.
    *   El campo `ChatIdGroup` de la tabla `Message` no está presente en `models.Message`. La consulta debe insertar `NULL` para `ChatIdGroup`.
    *   Debe ejecutar una sentencia `INSERT INTO Message (...) VALUES (...)`.
    *   Debe devolver el `msg.Id` y `nil` en caso de éxito, o `""` y un error en caso de fallo.
    *   Asegúrate de que cualquier función que llame a `CreateMessage` (como `CreateMessageFromChatParams`) se actualice para manejar la nueva firma de retorno (string, error).

3.  **Función `ProcessAndSaveChatMessage` (en `services/chat_service.go`):**
    *   Firma: `func ProcessAndSaveChatMessage(userID int64, payload map[string]interface{}, messageID string) (*models.Message, error)`
    *   Debe tomar el `userID` del remitente, un `payload map[string]interface{}` con los datos del mensaje (extraídos del mensaje WebSocket del cliente), y un `messageID` (pre-generado en el handler).
    *   Debe extraer `chatId` (string, requerido), `text` (string), `mediaId` (string), y `responseTo` (string) del `payload`.
    *   Debe validar que `chatId` no esté vacío y que al menos `text` o `mediaId` estén presentes.
    *   Debe determinar `typeMessageID`: 1 si `mediaId` está vacío, 2 si `mediaId` está presente (ajustar si la lógica de tipos es más compleja).
    *   Debe construir una instancia de `*models.Message` populando:
        *   `Id` con el `messageID` recibido.
        *   `ChatId`, `Text`, `MediaId`, `ResponseTo` con los valores extraídos/derivados.
        *   `UserId` con el `userID` recibido.
        *   `Date` con `time.Now().UTC()`.
        *   `TypeMessageId` con el valor determinado.
        *   `StatusMessage` con `queries.StatusMessageSent`.
    *   Debe llamar a `queries.CreateMessage(chatDB, &newMessage)` para guardar el mensaje.
    *   Debe actualizar `newMessage.Id` con el ID devuelto por `CreateMessage` (por si la BD lo modifica, aunque con UUIDs no debería).
    *   Debe loguear errores o éxito.
    *   Debe devolver el `*models.Message` guardado y `nil` en caso de éxito, o `nil` y un error en caso de fallo.

**Consideraciones Adicionales:**

*   El `payload` del cliente (mensaje WebSocket) contendrá campos como `chatId`, `text`, `mediaId`, `responseTo`. La función de servicio debe ser robusta para extraer estos campos del `map[string]interface{}`.
*   La variable `chatDB *sql.DB` ya está inicializada en `services/chat_service.go`.
*   Asegúrate de que las importaciones necesarias estén presentes en los archivos modificados.
*   Presta atención a la gestión de errores y al logging informativo.

**Archivos relevantes (para verificar estructuras y funciones existentes):**

*   `backend/internal/websocket/handlers/HandleSendChatMessage.go` (para ver cómo un handler podría usar el servicio).
*   `backend/internal/models/models.go` (para la estructura `models.Message`).
*   `backend/schema.sql` (para la definición de la tabla `Message`, incluyendo `ChatIdGroup`).

Comienza implementando las constantes y la función `CreateMessage` en `queries.go`, luego la función `ProcessAndSaveChatMessage` en `chat_service.go`.
```
