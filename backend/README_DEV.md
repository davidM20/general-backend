# 🚀 Herramienta de Desarrollo Backend

Esta herramienta te permite compilar y ejecutar todos los microservicios del backend de forma simultánea con logs unificados y coloreados.

## 🎯 Características

- ✅ **Compilación automática** de los 3 servicios (API, WebSocket, Proxy)
- ✅ **Logs unificados** con prefijos coloreados para cada servicio
- ✅ **Timestamps** en todos los logs
- ✅ **Manejo de señales** - Ctrl+C detiene todos los servicios
- ✅ **Detección de errores** durante compilación y ejecución
- ✅ **Colores distintivos** para cada servicio

## 🚀 Uso Rápido

### Opción 1: Script Bash (Recomendado)
```bash
./dev.sh
```

### Opción 2: Makefile
```bash
make dev
```

### Opción 3: Comando directo
```bash
go run ./cmd/devtools/main.go
```

## 📊 Servicios y Puertos

| Servicio  | Puerto | Color   | Descripción                    |
|-----------|--------|---------|--------------------------------|
| API       | 8081   | 🟢 Verde | Servidor API REST              |
| WebSocket | 8082   | 🟡 Amarillo | Servidor WebSocket             |
| Proxy     | 8080   | 🔵 Azul  | Proxy reverso con CORS         |

## 🎨 Formato de Logs

```
15:04:05 [API] 🚀 Servicio iniciado
15:04:05 [WebSocket] API Server starting on port 8081 (CORS handled by proxy)...
15:04:06 [Proxy] 🚀 Reverse Proxy server starting on port 8080 with CORS enabled
```

## 📋 Comandos Makefile Disponibles

```bash
make help           # Mostrar ayuda
make dev            # Ejecutar todos los servicios
make build          # Compilar todos los servicios
make install-deps   # Instalar dependencias
make run-api        # Ejecutar solo API
make run-ws         # Ejecutar solo WebSocket  
make run-proxy      # Ejecutar solo Proxy
make clean          # Limpiar binarios
make fmt            # Formatear código
make test           # Ejecutar tests
```

## 🛑 Detener Servicios

Para detener todos los servicios, simplemente presiona **Ctrl+C** en la terminal donde está ejecutándose la herramienta.

## 🔧 Solución de Problemas

### Error de compilación
Si algún servicio falla al compilar, revisa:
- ✅ Variables de entorno configuradas
- ✅ Dependencias instaladas: `make install-deps`
- ✅ Archivos `.env` configurados

### Puertos ocupados
Si hay conflictos de puertos:
```bash
# Verificar qué procesos usan los puertos
lsof -i :8080
lsof -i :8081  
lsof -i :8082

# Terminar procesos si es necesario
kill -9 <PID>
```

### Logs no aparecen
- ✅ Verifica que los servicios estén enviando logs a stdout/stderr
- ✅ Revisa la configuración de logging en cada servicio

## 🎯 Arquitectura CORS

```
Frontend (localhost:5173) 
    ↓ 
Proxy (localhost:8080) [MANEJA CORS] 
    ↓ 
├── API (localhost:8081) [SIN CORS]
└── WebSocket (localhost:8082) [SIN CORS]
```

**Importante**: Solo el Proxy maneja CORS para evitar duplicación de headers.

## 📝 Estructura de Archivos

```
backend/
├── cmd/
│   ├── devtools/     # 🛠️ Herramienta de desarrollo
│   ├── api/          # 🌐 Servidor API
│   ├── websocket/    # 🔌 Servidor WebSocket
│   └── proxy/        # 🔄 Proxy reverso
├── bin/              # 📦 Binarios compilados
├── dev.sh            # 🚀 Script de desarrollo
├── Makefile          # 🔧 Comandos Make
└── README_DEV.md     # 📚 Esta documentación
```

## 🚀 Para Empezar

1. **Instalar dependencias**:
   ```bash
   make install-deps
   ```

2. **Configurar variables de entorno**:
   ```bash
   cp .env_example .env
   # Editar .env con tus configuraciones
   ```

3. **Ejecutar en modo desarrollo**:
   ```bash
   ./dev.sh
   ```

¡Ya tienes todos los servicios ejecutándose con logs unificados! 🎉 