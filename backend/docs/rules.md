# Estructura de Action y Resource

## 1. Orden de Procesamiento
El sistema procesa los mensajes en este orden:
1. Primero verifica la `Action`
2. Luego verifica el `Resource` (excepto para "ping")

## 2. Combinaciones Válidas

### Action: "ping"
- No requiere Resource
- Es la única acción que puede funcionar sin Resource

### Action: "get_list"
- **Resources válidos**:
  - `"chat"` → Obtiene lista de chats
  - `"notification"` → Obtiene lista de notificaciones

### Action: "get_pending"
- **Resources válidos**:
  - `"notification"` → Obtiene notificaciones pendientes
  - Añade automáticamente `onlyUnread: true`

### Action: "get_history"
- **Resources válidos**:
  - `"chat"` → Obtiene historial de chat
  - Requiere datos adicionales: `chatId`, `limit`, `beforeMessageId`

### Action: "send_message"
- **Resources válidos**:
  - `"chat"` → Envía mensaje de chat
  - Requiere datos adicionales: `chatId`, `text`

## 3. Ejemplo de Estructura
```json
{
    "action": "get_list",     // Primero se verifica la acción
    "resource": "chat",       // Luego se verifica el recurso
    "data": {                 // Datos específicos según la combinación
        // campos específicos
    }
}
```

## 4. Reglas Importantes
1. Toda acción (excepto "ping") DEBE tener un Resource
2. Si no se especifica un Resource, se devuelve error 400
3. Las combinaciones Action-Resource deben ser exactamente las especificadas
4. Cada combinación puede requerir datos específicos en el campo `data`

Esta estructura jerárquica (Action → Resource → Data) permite un sistema de enrutamiento claro y predecible para el manejo de mensajes.
