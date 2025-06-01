# Guía para Agregar Nuevas Acciones y Recursos WebSocket

Este documento describe el proceso paso a paso para añadir nuevas combinaciones de `action` y `resource` al sistema de comunicación WebSocket, afectando tanto al backend (Go) como al frontend (React Native / TypeScript).

## Contexto General

La comunicación WebSocket se basa en un mensaje principal estructurado con `action`, `resource` y `data`:

```json
{
  "action": "nombre_accion",
  "resource": "nombre_recurso",
  "data": { /* datos específicos */ }
}
```

El backend despacha las solicitudes en `handelDataRequest.go` y los datos específicos del payload se definen en `wsmodels/types.go`. El frontend utiliza hooks personalizados para interactuar con el WebSocket, encapsulando la lógica de envío y recepción.

---

## Backend (Go)

### 1. Definir Tipos de Payload (si es necesario)

**Archivo:** `backend/internal/websocket/wsmodels/types.go`

Si la nueva acción o los datos de respuesta requieren estructuras específicas, añádelas aquí.

*   **Ejemplo (Nueva estructura para respuesta):**

    ```go
    // wsmodels/types.go
    // ... (otros structs)

    type MiNuevoRecursoDataPayload struct {
        Campo1 string `json:"campo1"`
        Campo2 int    `json:"campo2"`
        // ... más campos
    }
    ```

### 2. Crear el Handler Específico

**Directorio:** `backend/internal/websocket/handlers/`

Crea un nuevo archivo `.go` para el handler de tu nueva acción/recurso, o modifica uno existente si es una extensión lógica.

*   **Nombre del archivo:** Por convención, `handle[Accion][Recurso].go` (ej. `handleGetMiNuevoRecurso.go`).
*   **Contenido del Handler:**
    *   Debe ser una función que reciba `*customws.Connection[wsmodels.WsUserData]` y `types.ClientToServerMessage`.
    *   Procesará `msg.Payload` (que contendrá el campo `data` de la solicitud del cliente).
    *   Realizará la lógica de negocio (ej. consultas a la BD a través de `queries` o lógica en `services`).
    *   Construirá y enviará una respuesta al cliente usando `conn.SendMessage()` o `conn.SendErrorNotification()`.

*   **Ejemplo (Handler para `get_info` / `mi_nuevo_recurso`):**

    ```go
    // backend/internal/websocket/handlers/handleGetMiNuevoRecurso.go
    package handlers

    import (
        "fmt"
        "github.com/davidM20/micro-service-backend-go.git/internal/websocket/wsmodels"
        "github.com/davidM20/micro-service-backend-go.git/pkg/customws"
        "github.com/davidM20/micro-service-backend-go.git/pkg/customws/types"
        "github.com/davidM20/micro-service-backend-go.git/pkg/logger"
    )

    const miNuevoRecursoLogComponent = "HANDLER_MI_NUEVO_RECURSO"

    func HandleGetMiNuevoRecurso(conn *customws.Connection[wsmodels.WsUserData], msg types.ClientToServerMessage) error {
        logger.Infof(miNuevoRecursoLogComponent, "Solicitud para mi_nuevo_recurso recibida de UserID %d, PID: %s", conn.ID, msg.PID)

        // 1. Procesar msg.Payload si es necesario (ej. obtener parámetros de msg.Payload.(map[string]interface{}))
        // var requestParams struct { Param1 string `json:"param1"` }
        // payloadBytes, _ := json.Marshal(msg.Payload)
        // json.Unmarshal(payloadBytes, &requestParams)

        // 2. Lógica para obtener los datos (ej. de la BD o un servicio)
        data := wsmodels.MiNuevoRecursoDataPayload{
            Campo1: "Valor1",
            Campo2: 123,
        }

        // 3. Preparar respuesta
        responsePayload := map[string]interface{}{
            "origin":          "get_info_mi_nuevo_recurso", // Identificador para el frontend
            "mi_nuevo_recurso": data,
        }

        responseMsg := types.ServerToClientMessage{
            Type:       types.MessageTypeDataEvent, // O un tipo más específico si se crea
            FromUserID: 0, // Sistema
            Payload:    responsePayload,
            PID:        conn.Manager().Callbacks().GeneratePID(),
        }

        if err := conn.SendMessage(responseMsg); err != nil {
            logger.Errorf(miNuevoRecursoLogComponent, "Error enviando datos de mi_nuevo_recurso a UserID %d: %v", conn.ID, err)
            return fmt.Errorf("error enviando datos: %w", err)
        }

        logger.Successf(miNuevoRecursoLogComponent, "Datos de mi_nuevo_recurso enviados a UserID %d", conn.ID)
        return nil
    }
    ```

### 3. Registrar la Nueva Acción/Recurso en el Dispatcher Principal

**Archivo:** `backend/internal/websocket/handlers/handelDataRequest.go`

Añade un nuevo `case` o modifica uno existente en el `switch` para dirigir la combinación `action`/`resource` a tu nuevo handler.

*   **Lógica de ACK:**
    *   Si la operación puede tomar tiempo o involucra enviar un `data_event` posterior (como `get_info/dashboard`), envía un `server_ack` inmediato al cliente desde `handelDataRequest.go` y luego llama al handler específico (posiblemente en una goroutine).
    *   Si el handler específico envía una respuesta directa y rápida (como un error o un pequeño payload en respuesta a una acción como `send_message`), puede que no necesites un ACK explícito desde `handelDataRequest.go` si el handler maneja el `msg.PID` para su propia respuesta.

*   **Ejemplo (Añadiendo `get_info` / `mi_nuevo_recurso`):**

    ```go
    // backend/internal/websocket/handlers/handelDataRequest.go
    // ...
    switch requestData.Action {
    case "get_info":
        switch requestData.Resource {
        case "dashboard":
            // ... (código existente para dashboard)
            // ...
            go func(currentConn *customws.Connection[wsmodels.WsUserData], originalMsg types.ClientToServerMessage) {
                if err := HandleGetDashboardInfo(currentConn, originalMsg); err != nil {
                    logger.Errorf("HANDLER_DATA", "Error en goroutine HandleGetDashboardInfo para UserID %d, PID %s: %v", currentConn.ID, originalMsg.PID, err)
                }
            }(conn, msg)
            return nil

        case "mi_nuevo_recurso": // <<< NUEVO CASE
            // Asumimos que HandleGetMiNuevoRecurso enviará un data_event y no requiere ACK desde aquí
            // o que HandleGetMiNuevoRecurso es rápido y envía su propia "respuesta" que actúa como ACK.
            // Si HandleGetMiNuevoRecurso es lento y envía un data_event asíncrono, considera el patrón de ACK + goroutine:
            /*
            if msg.PID != "" {
                ackPayload := types.AckPayload{AcknowledgedPID: msg.PID, Status: "processing_mi_nuevo_recurso"}
                ackMsg := types.ServerToClientMessage{ PID: conn.Manager().Callbacks().GeneratePID(), Type: types.MessageTypeServerAck, Payload: ackPayload }
                if err := conn.SendMessage(ackMsg); err != nil { logger.Warnf("HANDLER_DATA", "Error ACK: %v", err) }
            }
            go func(c *customws.Connection[wsmodels.WsUserData], m types.ClientToServerMessage) {
                 if err := HandleGetMiNuevoRecurso(c, m); err != nil { logger.Errorf("HANDLER_DATA", "Error goroutine: %v", err) }
            }(conn, msg)
            return nil
            */
            return HandleGetMiNuevoRecurso(conn, msg) // Llamada directa si es síncrono o maneja su propio ACK/respuesta

        default:
            return handleUnsupportedResource(conn, msg.PID, requestData.Action, requestData.Resource)
        }
    // ... (otros cases para acciones como get_list, send_message, etc.)
    }
    // ...
    ```

### 4. (Opcional) Añadir Consultas a la Base de Datos

**Archivo:** `backend/internal/db/queries/queries.go`

Si necesitas nuevas consultas SQL:
*   Añade la función de consulta aquí.
*   Recuerda usar el wrapper `queries.MeasureQueryWithResult` o `queries.MeasureQuery` para las métricas.

---

## Frontend (React Native / TypeScript)

### 1. (Opcional) Definir Nuevas Interfaces de Datos

**Archivo:** `alumni-app/src/lib/react-native-ws-hooks.tsx` (o un archivo de tipos dedicado)

Si el frontend necesita manejar nuevas estructuras de datos (para la solicitud o la respuesta), define las interfaces TypeScript correspondientes.

*   **Ejemplo (Interfaz para `MiNuevoRecursoDataPayload`):**

    ```typescript
    // alumni-app/src/lib/react-native-ws-hooks.tsx
    // ...
    export interface MiNuevoRecursoData {
        campo1: string;
        campo2: number;
        // ... más campos
    }
    ```

### 2. Crear o Modificar un Hook Personalizado

**Directorio:** `alumni-app/src/hook/nombreDelModulo/` (ej. `alumni-app/src/hook/useMiNuevoRecurso/`)

Crea un nuevo hook o modifica uno existente para encapsular la lógica de interacción con el WebSocket para esta nueva funcionalidad.

*   **Nombre del archivo:** `useWS[NombreModulo].ts` (ej. `useWSMiNuevoRecurso.ts`).
*   **Contenido del Hook:**
    *   Importar `useWSContext` de `../useContext/useWSContext` (o la ruta correcta) para acceder a `sendDataRequest`, `addMessageListener`, `removeMessageListener`.
    *   Manejar estados locales (ej. `data`, `loading`, `error`).
    *   Implementar una función para enviar la solicitud al backend (ej. `loadMiNuevoRecursoData`).
        *   Usar `sendDataRequest` con el `action` y `resource` correctos.
    *   Usar `useEffect` con `addMessageListener` y `removeMessageListener` para escuchar los `data_event` (o mensajes específicos) enviados por el backend y actualizar el estado.

*   **Ejemplo (Hook `useWSMiNuevoRecurso.ts`):**

    ```typescript
    // alumni-app/src/hook/useMiNuevoRecurso/useWSMiNuevoRecurso.ts
    import { useState, useCallback, useEffect, useRef } from 'react';
    import { useWSContext } from '../useContext/useWSContext';
    import { DataRequestPayload, ServerToClientMessage, MiNuevoRecursoData } from '@/lib/react-native-ws-hooks';

    export const useWSMiNuevoRecurso = () => {
        const {
            sendDataRequest,
            connectionState,
            addMessageListener,
            removeMessageListener
        } = useWSContext();

        // Estados para UI y control
        const [recursoData, setRecursoData] = useState<MiNuevoRecursoData | null>(null);
        const [loading, setLoading] = useState(false);
        const [error, setError] = useState<string | null>(null);
        
        // Ref para control de estado de carga sin causar re-renders
        const loadingRef = useRef(false);

        const loadRecursoData = useCallback(async (params?: Record<string, any>) => {
            // Validación de conexión
            if (connectionState !== 'connected') {
                console.warn('[useWSMiNuevoRecurso] No conectado. Esperando conexión...');
                setError('No conectado al WebSocket');
                setLoading(false);
                loadingRef.current = false;
                return;
            }

            // Validación de función de envío
            if (!sendDataRequest) {
                console.warn('[useWSMiNuevoRecurso] sendDataRequest no disponible');
                setError('Función de envío no disponible');
                setLoading(false);
                loadingRef.current = false;
                return;
            }

            // Prevención de cargas múltiples
            if (loadingRef.current) {
                console.log('[useWSMiNuevoRecurso] Carga ya en progreso');
                return;
            }

            // Inicio de carga
            setLoading(true);
            loadingRef.current = true;
            setError(null);

            try {
                const payload: DataRequestPayload = {
                    action: 'get_info',
                    resource: 'mi_nuevo_recurso',
                    data: params || {},
                };
                console.log('[useWSMiNuevoRecurso] Enviando solicitud...', payload);
                await sendDataRequest(payload);

                // Timeout de seguridad para evitar estados de carga infinitos
                const timeoutId = setTimeout(() => {
                    if (loadingRef.current) {
                        console.warn('[useWSMiNuevoRecurso] Timeout al cargar datos');
                        setLoading(false);
                        loadingRef.current = false;
                        setError('Timeout al cargar los datos');
                    }
                }, 10000);

                return () => clearTimeout(timeoutId);
            } catch (err: any) {
                console.error('[useWSMiNuevoRecurso] Error solicitando datos:', err);
                setError(err.message || 'Error al solicitar datos');
                setLoading(false);
                loadingRef.current = false;
            }
        }, [sendDataRequest, connectionState]); // Removida la dependencia de loading

        useEffect(() => {
            if (!addMessageListener || !removeMessageListener) return;

            const handleMessage = (message: ServerToClientMessage) => {
                if (
                    message.type === 'data_event' &&
                    message.payload?.origin === 'get_info_mi_nuevo_recurso' &&
                    message.payload?.mi_nuevo_recurso
                ) {
                    console.log('[useWSMiNuevoRecurso] Datos recibidos:', message.payload.mi_nuevo_recurso);
                    setRecursoData(message.payload.mi_nuevo_recurso as MiNuevoRecursoData);
                    setLoading(false);
                    loadingRef.current = false;
                    setError(null);
                }
            };

            addMessageListener(handleMessage);
            return () => removeMessageListener(handleMessage);
        }, [addMessageListener, removeMessageListener]); // Removida la dependencia de setRecursoData

        return {
            recursoData,
            loading,
            error,
            loadRecursoData,
        };
    };
    ```

### 3. Integrar el Hook en el Componente de la Vista

**Archivo:** El componente de React Native donde se necesita la nueva funcionalidad (ej. `alumni-app/src/views/algunModulo/MiVista.tsx`).

*   Importar y usar el nuevo hook.
*   Llamar a la función de carga de datos (ej. `loadRecursoData`) desde un `useEffect` (ej. cuando el componente se monta o cuando la conexión WS está lista).
*   Renderizar la UI basándose en los estados `recursoData`, `loading`, y `error` del hook.

*   **Ejemplo (En `MiVista.tsx`):**

    ```typescript
    // alumni-app/src/views/algunModulo/MiVista.tsx
    import React, { useEffect, useCallback } from 'react';
    import { View, Text, Button, ActivityIndicator } from 'react-native';
    import { useWSMiNuevoRecurso } from '@/hook/useMiNuevoRecurso/useWSMiNuevoRecurso';
    import { useWSContext } from '@/hook/useContext/useWSContext';

    const MiVista: React.FC = () => {
        const { recursoData, loading, error, loadRecursoData } = useWSMiNuevoRecurso();
        const { connectionState } = useWSContext();

        // Memoizar la función de recarga para evitar recreaciones innecesarias
        const handleRefresh = useCallback(() => {
            if (connectionState === 'connected') {
                loadRecursoData();
            }
        }, [connectionState, loadRecursoData]);

        useEffect(() => {
            if (connectionState === 'connected') {
                console.log('[MiVista] Conectado. Cargando datos...');
                loadRecursoData();
            }
        }, [connectionState, loadRecursoData]);

        if (loading) return <ActivityIndicator size="large" />;
        if (error) return (
            <View>
                <Text>Error: {error}</Text>
                <Button title="Reintentar" onPress={handleRefresh} />
            </View>
        );
        if (!recursoData) return (
            <View>
                <Text>No hay datos disponibles.</Text>
                <Button title="Cargar" onPress={handleRefresh} />
            </View>
        );

        return (
            <View>
                <Text>Datos de Mi Nuevo Recurso:</Text>
                <Text>Campo1: {recursoData.campo1}</Text>
                <Text>Campo2: {recursoData.campo2}</Text>
                <Button title="Actualizar" onPress={handleRefresh} />
            </View>
        );
    };

    export default MiVista;
    ```

### 4. (Opcional) Exportar el Hook desde el Agregador Principal

**Archivo:** `alumni-app/src/lib/react-native-ws-hooks.tsx`

Si deseas que el nuevo hook sea fácilmente importable junto con los otros hooks de WS, puedes reexportarlo desde este archivo.

*   **Ejemplo:**

    ```typescript
    // alumni-app/src/lib/react-native-ws-hooks.tsx
    // ... (otras importaciones y reexportaciones)
    export { useWSMiNuevoRecurso } from '../hook/useMiNuevoRecurso/useWSMiNuevoRecurso'; // Nueva reexportación

    export default {
        // ... (otros hooks)
        useWSMiNuevoRecurso, // Añadir al default export
    };
    ```

---

## Consideraciones Importantes y Prevención de Problemas Comunes

### 1. Manejo de Tipos de Mensajes

Es crucial verificar el tipo exacto de mensaje que el servidor envía. Los tipos comunes son:

```typescript
type MessageType = 
  | 'data_event'    // Para eventos de datos generales
  | 'chat_list'     // Para listas de chat
  | 'chat_message'  // Para mensajes individuales
  | 'server_ack'    // Para confirmaciones del servidor
  | 'error'         // Para mensajes de error
```

**Ejemplo de manejo correcto:**
```typescript
const handleMessage = (message: ServerToClientMessage) => {
  // Verificar el tipo exacto del mensaje
  if (message.type === 'chat_list' && Array.isArray(message.payload)) {
    // Procesar lista de chats
  } else if (message.type === 'data_event' && message.payload?.origin === 'get_info_dashboard') {
    // Procesar datos del dashboard
  }
};
```

### 2. Prevención de Loops Infinitos

Para evitar loops infinitos en los hooks:

1. **Usar useRef para control de estado:**
```typescript
const loadingRef = useRef(false);

const loadData = useCallback(async () => {
  if (loadingRef.current) return;
  loadingRef.current = true;
  // ... resto del código
}, [/* dependencias */]);
```

2. **Manejar timeouts:**
```typescript
const timeoutId = setTimeout(() => {
  if (loadingRef.current) {
    setLoading(false);
    loadingRef.current = false;
    setError('Timeout al cargar datos');
  }
}, 10000);

return () => clearTimeout(timeoutId);
```

3. **Optimizar dependencias:**
```typescript
// ❌ Evitar
useEffect(() => {
  // ... código
}, [loading]); // Puede causar loops

// ✅ Preferir
useEffect(() => {
  // ... código
}, [connectionState]); // Solo dependencias necesarias
```

### 3. Manejo de Estados de Carga y Error

Implementar estados visuales claros:

```typescript
const LoadingIndicator = () => (
  <View style={styles.loadingContainer}>
    <ActivityIndicator size="large" color={themeColors.text} />
    <Text style={styles.loadingText}>Cargando...</Text>
  </View>
);

const ErrorMessage = () => (
  <View style={styles.errorContainer}>
    <Text style={styles.errorText}>{error}</Text>
  </View>
);

const EmptyList = () => (
  <View style={styles.emptyContainer}>
    <Text style={styles.emptyText}>No hay datos disponibles</Text>
  </View>
);
```

### 4. Validación de Datos

Siempre validar los datos recibidos:

```typescript
if (message.type === 'chat_list' && Array.isArray(message.payload)) {
  const chatList: ChatInfo[] = message.payload.map((chat: any) => ({
    chatId: chat.chatId,
    otherUserId: chat.otherUserId,
    otherUserName: chat.otherUserName || 'Usuario',
    // ... otros campos con valores por defecto
  }));
}
```

### 5. Logging y Depuración

Implementar logging consistente:

```typescript
// En el hook
console.log('[useWSHook] Enviando solicitud...', payload);
console.log('[useWSHook] Datos recibidos:', data);

// En el componente
console.log('[ComponentName] Estado actualizado:', { loading, error, data });
```

### 6. Manejo de Reconexión

Implementar lógica de reconexión:

```typescript
useEffect(() => {
  if (connectionState === 'connected') {
    console.log('[ComponentName] Conectado. Cargando datos...');
    loadData();
  } else if (connectionState === 'disconnected') {
    console.log('[ComponentName] Desconectado. Esperando reconexión...');
    setError('Desconectado del servidor');
  }
}, [connectionState, loadData]);
```

### 7. Limpieza de Recursos

Siempre limpiar recursos y listeners:

```typescript
useEffect(() => {
  addMessageListener(handleMessage);
  return () => {
    removeMessageListener(handleMessage);
    console.log('[ComponentName] Listener eliminado');
  };
}, [addMessageListener, removeMessageListener, handleMessage]);
```

Siguiendo estas prácticas, se pueden evitar problemas comunes como:
- Loops infinitos
- Estados de carga bloqueados
- Pérdida de memoria por listeners no limpiados
- Errores por tipos de mensajes incorrectos
- Estados de UI inconsistentes

Siguiendo estos pasos, podrás integrar nuevas funcionalidades WebSocket de manera estructurada y consistente con la arquitectura existente del proyecto.

``` 