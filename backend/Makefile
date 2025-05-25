# Makefile para Backend Microservices

.PHONY: dev build clean help install-deps setup

# Variables
DEV_TOOL = ./bin/devtools
API_BIN = ./bin/api
WS_BIN = ./bin/websocket
PROXY_BIN = ./bin/proxy

# Comando por defecto
all: build

# Setup completo del entorno de desarrollo
setup:
	@echo "🔧 Configurando entorno de desarrollo..."
	@if [ ! -f .env ]; then \
		echo "📋 Creando archivo .env desde template..."; \
		cp .env_example_dev .env; \
		echo "✅ Archivo .env creado. Edítalo con tu configuración."; \
	else \
		echo "ℹ️  Archivo .env ya existe."; \
	fi
	@make install-deps
	@make build-devtools
	@echo ""
	@echo "🎉 ¡Entorno listo! Usa 'make dev' para iniciar todos los servicios"

# Instalar dependencias
install-deps:
	@echo "📦 Instalando dependencias..."
	go mod download
	go mod tidy
	@echo "✅ Dependencias instaladas"

# Compilar herramienta de desarrollo
build-devtools:
	@echo "🔨 Compilando herramienta de desarrollo..."
	go build -o $(DEV_TOOL) ./cmd/devtools/main.go
	@echo "✅ Herramienta de desarrollo compilada"

# Ejecutar todos los servicios en modo desarrollo
dev: build-devtools
	@echo "🚀 Iniciando todos los servicios en modo desarrollo..."
	$(DEV_TOOL)

# Compilar todos los servicios individualmente
build:
	@echo "🔨 Compilando todos los servicios..."
	go build -o $(API_BIN) ./cmd/api/main.go
	go build -o $(WS_BIN) ./cmd/websocket/main.go
	go build -o $(PROXY_BIN) ./cmd/proxy/main.go
	@echo "✅ Todos los servicios compilados"

# Ejecutar solo el API
run-api: 
	go build -o $(API_BIN) ./cmd/api/main.go && $(API_BIN)

# Ejecutar solo el WebSocket
run-ws:
	go build -o $(WS_BIN) ./cmd/websocket/main.go && $(WS_BIN)

# Ejecutar solo el Proxy
run-proxy:
	go build -o $(PROXY_BIN) ./cmd/proxy/main.go && $(PROXY_BIN)

# Limpiar binarios
clean:
	@echo "🧹 Limpiando binarios..."
	rm -f ./bin/*
	@echo "✅ Binarios eliminados"

# Verificar formato de código
fmt:
	@echo "🎨 Formateando código..."
	go fmt ./...
	@echo "✅ Código formateado"

# Ejecutar tests
test:
	@echo "🧪 Ejecutando tests..."
	go test ./...
	@echo "✅ Tests completados"

# Verificar puertos en uso
check-ports:
	@echo "🔍 Verificando puertos..."
	@echo "Puerto 8080 (Proxy):"
	@lsof -i :8080 || echo "  ✅ Libre"
	@echo "Puerto 8081 (API):"
	@lsof -i :8081 || echo "  ✅ Libre"
	@echo "Puerto 8082 (WebSocket):"
	@lsof -i :8082 || echo "  ✅ Libre"

# Mostrar ayuda
help:
	@echo "📚 Comandos disponibles:"
	@echo ""
	@echo "  make setup        - Configurar entorno de desarrollo completo"
	@echo "  make dev          - Ejecutar todos los servicios en modo desarrollo"
	@echo "  make build        - Compilar todos los servicios"
	@echo "  make install-deps - Instalar dependencias"
	@echo "  make run-api      - Ejecutar solo el servicio API"
	@echo "  make run-ws       - Ejecutar solo el servicio WebSocket"
	@echo "  make run-proxy    - Ejecutar solo el servicio Proxy"
	@echo "  make check-ports  - Verificar si los puertos están en uso"
	@echo "  make clean        - Limpiar binarios compilados"
	@echo "  make fmt          - Formatear código"
	@echo "  make test         - Ejecutar tests"
	@echo "  make help         - Mostrar esta ayuda"
	@echo ""
	@echo "🚀 Para empezar desde cero: make setup && make dev" 