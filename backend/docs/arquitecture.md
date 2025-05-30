```
backend/
├── cmd/websocket/
│   └── main.go                 # Punto de entrada del servidor
├── internal/
│   ├── config/
│   │   └── config.go          # Configuración de la aplicación
│   ├── db/
│   │   ├── db.go              # Conexión y DDL de BD
│   │   └── queries/
│   │       └── queries.go     # Consultas a la base de datos
│   ├── models/
│   │   └── models.go          # Modelos de datos de BD
│   └── websocket/
│       ├── auth/
│       │   └── auth.go        # Autenticación WebSocket
│       ├── handlers/
│       │   ├── chat_handler.go
│       │   ├── profile_handler.go
│       │   └── notification_handler.go
│       ├── services/
│       │   ├── chat_service.go
│       │   ├── presence_service.go
│       │   ├── notification_service.go
│       │   └── profile_service.go
│       ├── wsmodels/
│       │   └── types.go       # DTOs específicos de WebSocket
│       ├── callbacks.go       # Callbacks de customws
│       └── router.go          # Enrutador de mensajes
├── pkg/
│   ├── customws/              # Paquete de gestión WebSocket
│   │   ├── customws.go
│   │   └── types/
│   │       └── types.go
│   └── logger/
│       └── logger.go          # Sistema de logging
```

