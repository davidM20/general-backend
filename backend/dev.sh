#!/bin/bash

# Script de desarrollo para Backend Microservices
# Este script compila y ejecuta todos los servicios con logs unificados

set -e  # Salir si hay algún error

echo "🚀 Backend Microservices Development Tool"
echo "========================================"
echo ""

# Colores para output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Verificar que go esté disponible
if ! command -v go &> /dev/null; then
    echo -e "${RED}Error: Go no está instalado o no está en PATH${NC}"
    exit 1
fi

# Crear directorio bin si no existe
mkdir -p bin

# Compilar herramienta de desarrollo
echo -e "${PURPLE}🔨 Compilando herramienta de desarrollo...${NC}"
go build -o ./bin/devtools ./cmd/devtools/main.go

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✅ Herramienta compilada exitosamente${NC}"
    echo ""
    
    # Ejecutar herramienta
    echo -e "${CYAN}🚀 Iniciando servicios...${NC}"
    ./bin/devtools
else
    echo -e "${RED}❌ Error compilando herramienta de desarrollo${NC}"
    exit 1
fi 