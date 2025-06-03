


Okay, voy a listar todos los archivos dentro de la carpeta `backend/internal/websocket` y sus subcarpetas (`handlers`, `admin`, `services`, `wsmodels`, `auth`). Luego, revisaré si existen archivos correspondientes a las APIs de WebSocket que identificamos como necesarias para el dashboard de administración y listaré los faltantes con sus estructuras de comunicación WS.

Primero, listaré los archivos en `backend/internal/websocket/handlers`.

Ran tool
Ahora listaré los archivos en `backend/internal/websocket/admin`.

Ran tool
Ahora listaré los archivos en `backend/internal/websocket/services`.

Ran tool
Ahora listaré los archivos en `backend/internal/websocket/wsmodels`.

Ran tool
Ahora listaré los archivos en `backend/internal/websocket/auth`.

Ran tool
He listado todos los archivos en las carpetas especificadas. Ahora revisaré si existen archivos relacionados con las APIs de administración que identificamos previamente y crearé la lista de faltantes.

Archivos existentes que podrían estar relacionados con las APIs de administración:

-   `backend/internal/websocket/admin/admin.go`: Este archivo con un nombre genérico "admin" y un tamaño considerable (805 líneas) podría contener la implementación de *varias* funcionalidades de administración, incluyendo potencialmente las del dashboard, gestión de carreras, sedes, etc. Es muy probable que gran parte de la lógica necesaria ya esté aquí.
-   `backend/internal/websocket/handlers/handelDataRequest.go`: Este handler genérico es donde probablemente se enrutan las solicitudes con `action` y `resource`. Debería contener la lógica para llamar a los servicios de administración adecuados.
-   `backend/internal/websocket/wsmodels/types.go`: Este archivo debe contener las definiciones de las estructuras de datos (payloads) utilizadas para la comunicación WS, tanto del cliente al servidor como viceversa. Deberíamos verificar si ya existen tipos para las solicitudes y respuestas de administración.

**Revisión más detallada y determinación de faltantes:**

Dada la existencia de un archivo `admin.go` grande, es muy probable que muchas de las funcionalidades identificadas ya estén implementadas allí. Para confirmar qué falta, necesitaría leer el contenido de `admin.go` y `handelDataRequest.go` para ver cómo manejan las solicitudes con `resource` como "dashboard", "degree", "campus", "event", "user", "role", "users_online" y las acciones correspondientes ("get_list", "get_info", "save", "delete", "update").

Sin embargo, basándome en los nombres de archivo y la arquitectura, puedo inferir lo siguiente:

-   **Dashboard General:** Es probable que la lógica para `action: "get_info", resource: "dashboard"` esté en `admin.go`, llamada desde `handelDataRequest.go`. Necesitaríamos confirmar si existe la implementación de la lógica para obtener todos los datos requeridos para el dashboard.
-   **Gestión de Carreras (`resource: "degree"`):** La lógica para "get_list", "save", "delete" probablemente debería estar en `admin.go`. Necesitaríamos confirmar si existen funciones específicas para manejar `resource: "degree"`.
-   **Gestión de Sedes (`resource: "campus"`):** Similar a carreras, la lógica para "get_list", "save", "delete" probablemente esté en `admin.go`. Necesitaríamos confirmar si existen funciones específicas para manejar `resource: "campus"`.
-   **Gestión de Eventos (`resource: "event"`):** Lo mismo aplica aquí; la lógica para "get_list", "save", "delete" debería estar en `admin.go`. Necesitaríamos confirmar si existen funciones para `resource: "event"`.
-   **Gestión de Roles de Usuarios (`resource: "user_role"` o "user" para lista, `resource: "role"` para lista):** La lógica para "get_list" de usuarios y roles, y "update" de user_role, debería estar en `admin.go`. Necesitaríamos confirmar si existen estas funciones y cómo se estructuran las solicitudes y respuestas.
-   **Usuarios en Línea (`resource: "users_online"`):** La lógica para obtener la lista de usuarios en línea probablemente esté en `admin.go` o interactúe con el `presence_service.go`. La emisión de `presence_event` ya existe según tu documentación inicial.

**Lista de APIs WS Faltantes y Estructura JSON (Potencial):**

Dado que no podemos saber con certeza qué *no* está implementado sin leer el código, esta lista se basa en la *suposición* de que cada "resource" de administración que identificamos en el frontend debería tener un manejo explícito en el backend. Es posible que algunas ya estén implementadas dentro del archivo `admin.go`.

Para cada recurso, se necesitaría un handler (probablemente dentro de `handelDataRequest.go` que enrute a `admin.go`) y funciones de servicio correspondientes (dentro de `admin.go` o un nuevo servicio de administración).

Aquí están las estructuras de comunicación WS para los recursos de administración, asumiendo que no existen o necesitan ser verificadas/implementadas:

**1. Recurso: `dashboard`**

*   **Funcionalidad:** Obtener métricas y datos para el dashboard general.
*   **Solicitud (frontend -> backend):**
    ```json
    {
        "action": "get_info",
        "resource": "dashboard",
        "data": {}
    }
    ```
*   **Respuesta (backend -> frontend):**
    ```json
    {
        "type": "data_event",
        "payload": {
            "origin": "get-info-dashboard",
            "dashboard": {
                "activeUsers": 0, // int
                "totalRegisteredUsers": 0, // int
                "administrativeUsers": 0, // int
                "businessAccounts": 0, // int
                "alumniStudents": 0, // int
                "averageUsageTime": "", // string (ej: "2h 30m")
                "usersByCampus": [ // Array de objetos
                    {"users": 0, "name": ""},
                    // ...
                ],
                "monthlyActivity": { // Objeto
                    "labels": [], // Array de strings (ej: ["Ene", "Feb"])
                    "data": [] // Array de ints
                }
            }
        }
    }
    ```
*   **Archivos Potenciales para Implementar:** `backend/internal/websocket/admin/admin.go`, `backend/internal/websocket/handlers/handelDataRequest.go`, `backend/internal/db/queries/queries.go` (nuevas consultas para obtener los datos).

**2. Recurso: `degree`**

*   **Funcionalidades:** Listar, Agregar, Editar, Eliminar carreras.
*   **Solicitud Lista (frontend -> backend):**
    ```json
    {
        "action": "get_list",
        "resource": "degree",
        "data": {}
    }
    ```
*   **Respuesta Lista (backend -> frontend):**
    ```json
    {
        "type": "data_event", // o "degree_list"
        "payload": {
            "origin": "degree-list", // si es data_event
            "degrees": [
                {"degreeId": 0, "degreeName": "", "universityId": 0, "descriptions": "", "code": ""},
                // ...
            ]
        }
    }
    ```
*   **Solicitud Guardar (frontend -> backend):**
    ```json
    {
        "action": "save", // o "create_or_update"
        "resource": "degree",
        "data": {
            "degreeId": null, // int64 o null para agregar
            "degreeName": "", // string
            "universityId": 0, // int64
            "descriptions": "", // string
            "code": "" // string
        }
    }
    ```
*   **Respuesta Guardar (backend -> frontend):**
    ```json
    {
        "type": "server_ack",
        "pid": "PID_de_la_solicitud",
        "status": "success", // o "error"
        "payload": {
            "degree": {"degreeId": 0, "degreeName": "", "universityId": 0, "descriptions": "", "code": ""} // Objeto de la carrera guardada, o mensaje de error
        }
    }
    ```
*   **Solicitud Eliminar (frontend -> backend):**
    ```json
    {
        "action": "delete",
        "resource": "degree",
        "data": {
            "degreeId": 0 // int64
        }
    }
    ```
*   **Respuesta Eliminar (backend -> frontend):**
    ```json
    {
        "type": "server_ack",
        "pid": "PID_de_la_solicitud",
        "status": "success", // o "error"
        "payload": null // o mensaje de error
    }
    ```
*   **Archivos Potenciales para Implementar:** `backend/internal/websocket/admin/admin.go`, `backend/internal/websocket/handlers/handelDataRequest.go`, `backend/internal/db/queries/queries.go`, `backend/internal/websocket/wsmodels/types.go` (definiciones de structs).

**3. Recurso: `campus`**

*   **Funcionalidades:** Listar, Agregar, Editar, Eliminar sedes.
*   **Solicitud Lista (frontend -> backend):**
    ```json
    {
        "action": "get_list",
        "resource": "campus",
        "data": {}
    }
    ```
*   **Respuesta Lista (backend -> frontend):**
    ```json
    {
        "type": "data_event", // o "campus_list"
        "payload": {
            "origin": "campus-list", // si es data_event
            "campuses": [
                {"campusId": 0, "name": "", "campus": ""},
                // ...
            ]
        }
    }
    ```
*   **Solicitud Guardar (frontend -> backend):**
    ```json
    {
        "action": "save", // o "create_or_update"
        "resource": "campus",
        "data": {
            "campusId": null, // int64 o null para agregar
            "name": "", // string (Nombre de la Universidad)
            "campus": "" // string (Nombre del Campus/Sede)
        }
    }
    ```
*   **Respuesta Guardar (backend -> frontend):**
    ```json
    {
        "type": "server_ack",
        "pid": "PID_de_la_solicitud",
        "status": "success", // o "error"
        "payload": {
            "campus": {"campusId": 0, "name": "", "campus": ""} // Objeto de la sede guardada, o mensaje de error
        }
    }
    ```
*   **Solicitud Eliminar (frontend -> backend):**
    ```json
    {
        "action": "delete",
        "resource": "campus",
        "data": {
            "campusId": 0 // int64
        }
    }
    ```
*   **Respuesta Eliminar (backend -> frontend):**
    ```json
    {
        "type": "server_ack",
        "pid": "PID_de_la_solicitud",
        "status": "success", // o "error"
        "payload": null // o mensaje de error
    }
    ```
*   **Archivos Potenciales para Implementar:** `backend/internal/websocket/admin/admin.go`, `backend/internal/websocket/handlers/handelDataRequest.go`, `backend/internal/db/queries/queries.go`, `backend/internal/websocket/wsmodels/types.go`.

**4. Recurso: `event`**

*   **Funcionalidades:** Listar, Agregar, Editar, Eliminar eventos.
*   **Solicitud Lista (frontend -> backend):**
    ```json
    {
        "action": "get_list",
        "resource": "event",
        "data": {}
    }
    ```
*   **Respuesta Lista (backend -> frontend):**
    ```json
    {
        "type": "data_event", // o "event_list"
        "payload": {
            "origin": "event-list", // si es data_event
            "events": [
                {"eventId": 0, "name": "", "date": "", "description": ""}, // Ejemplo básico, estructura completa depende de la implementación
                // ...
            ]
        }
    }
    ```
*   **Solicitud Guardar (frontend -> backend):**
    ```json
    {
        "action": "save", // o "create_or_update"
        "resource": "event",
        "data": {
            "eventId": null, // int64 o null para agregar
            "name": "", // string
            "date": "", // string (formato de fecha/hora)
            "description": "" // string
            // ... otros campos del evento
        }
    }
    ```
*   **Respuesta Guardar (backend -> frontend):**
    ```json
    {
        "type": "server_ack",
        "pid": "PID_de_la_solicitud",
        "status": "success", // o "error"
        "payload": {
            "event": {"eventId": 0, "name": "", ...} // Objeto del evento guardado, o mensaje de error
        }
    }
    ```
*   **Solicitud Eliminar (frontend -> backend):**
    ```json
    {
        "action": "delete",
        "resource": "event",
        "data": {
            "eventId": 0 // int64
        }
    }
    ```
*   **Respuesta Eliminar (backend -> frontend):**
    ```json
    {
        "type": "server_ack",
        "pid": "PID_de_la_solicitud",
        "status": "success", // o "error"
        "payload": null // o mensaje de error
    }
    ```
*   **Archivos Potenciales para Implementar:** `backend/internal/websocket/admin/admin.go`, `backend/internal/websocket/handlers/handelDataRequest.go`, `backend/internal/db/queries/queries.go`, `backend/internal/websocket/wsmodels/types.go`.

**5. Recurso: `user` (para lista de usuarios con roles) y `role` (para lista de roles)**

*   **Funcionalidades:** Listar usuarios con sus roles, Listar roles, Actualizar rol de usuario.
*   **Solicitud Lista Usuarios (frontend -> backend):**
    ```json
    {
        "action": "get_list",
        "resource": "user", // o "users_with_roles"
        "data": {
            "search_query": "", // string (opcional)
            "role_filter": null // int64 o null (opcional)
        }
    }
    ```
*   **Respuesta Lista Usuarios (backend -> frontend):**
    ```json
    {
        "type": "data_event", // o "user_list"
        "payload": {
            "origin": "user-list", // si es data_event
            "users": [
                {"id": 0, "name": "", "surname": "", "email": "", "cedula": "", "photo": "", "roleId": 0}, // Ejemplo básico
                // ...
            ]
        }
    }
    ```
*   **Solicitud Lista Roles (frontend -> backend):**
    ```json
    {
        "action": "get_list",
        "resource": "role",
        "data": {}
    }
    ```
*   **Respuesta Lista Roles (backend -> frontend):**
    ```json
    {
        "type": "data_event", // o "role_list"
        "payload": {
            "origin": "role-list", // si es data_event
            "roles": [
                {"id": 0, "label": ""},
                // ...
            ]
        }
    }
    ```
*   **Solicitud Actualizar Rol (frontend -> backend):**
    ```json
    {
        "action": "update", // o "assign_role"
        "resource": "user_role", // o "user"
        "data": {
            "userId": 0, // int64
            "roleId": 0 // int64
        }
    }
    ```
*   **Respuesta Actualizar Rol (backend -> frontend):**
    ```json
    {
        "type": "server_ack",
        "pid": "PID_de_la_solicitud",
        "status": "success", // o "error"
        "payload": null // o mensaje de error
    }
    ```
*   **Archivos Potenciales para Implementar:** `backend/internal/websocket/admin/admin.go`, `backend/internal/websocket/handlers/handelDataRequest.go`, `backend/internal/db/queries/queries.go`, `backend/internal/websocket/wsmodels/types.go`.

**6. Recurso: `users_online`**

*   **Funcionalidad:** Listar usuarios actualmente en línea.
*   **Solicitud Lista (frontend -> backend):**
    ```json
    {
        "action": "get_list",
        "resource": "users_online", // o "presence"
        "data": {}
    }
    ```
*   **Respuesta Lista (backend -> frontend):**
    ```json
    {
        "type": "data_event", // o "users_online_list"
        "payload": {
            "origin": "users-online-list", // si es data_event
            "users": [
                {"id": 0, "name": "", "lastActivity": ""}, // Ejemplo básico, depende de qué info de "online" se almacene/exponga
                // ...
            ]
        }
    }
    ```
*   **Archivos Potenciales para Implementar:** `backend/internal/websocket/admin/admin.go` (o interaccionar con `presence_service.go`), `backend/internal/websocket/handlers/handelDataRequest.go`, `backend/internal/db/queries/queries.go` (si se necesita info adicional de la BD).

Para avanzar, el siguiente paso sería examinar el contenido de `backend/internal/websocket/admin/admin.go` y `backend/internal/websocket/handlers/handelDataRequest.go` para ver qué de estas APIs ya está implementado y qué necesita ser añadido o modificado.
