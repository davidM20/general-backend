package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
)

// Colores ANSI para los logs
const (
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Blue   = "\033[34m"
	Purple = "\033[35m"
	Cyan   = "\033[36m"
	White  = "\033[37m"
	Reset  = "\033[0m"
	Bold   = "\033[1m"
)

// Servicio representa un microservicio
type Service struct {
	Name      string
	Path      string
	Color     string
	Cmd       *exec.Cmd
	Port      string
	BuildPath string
}

func main() {
	fmt.Printf("%s%s🚀 Backend Microservices Development Tool%s\n", Bold, Cyan, Reset)
	fmt.Printf("%s================================%s\n\n", Cyan, Reset)

	// Definir los servicios
	services := []Service{
		{
			Name:      "API",
			Path:      "cmd/api",
			BuildPath: "./cmd/api/main.go",
			Color:     Green,
			Port:      "8081",
		},
		{
			Name:      "WebSocket",
			Path:      "cmd/websocket",
			BuildPath: "./cmd/websocket/main.go",
			Color:     Yellow,
			Port:      "8082",
		},
		{
			Name:      "Proxy",
			Path:      "cmd/proxy",
			BuildPath: "./cmd/proxy/main.go",
			Color:     Blue,
			Port:      "8080",
		},
	}

	// Crear contexto cancelable
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Configurar manejo de señales
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// WaitGroup para sincronizar goroutines
	var wg sync.WaitGroup

	// Compilar servicios
	fmt.Printf("%s%s🔨 Compilando servicios...%s\n", Bold, Purple, Reset)
	for i := range services {
		if !buildService(&services[i]) {
			log.Fatalf("Error compilando servicio %s", services[i].Name)
		}
	}
	fmt.Printf("%s%s✅ Todos los servicios compilados exitosamente%s\n\n", Bold, Green, Reset)

	// Iniciar servicios
	fmt.Printf("%s%s🚀 Iniciando servicios...%s\n", Bold, Cyan, Reset)
	for i := range services {
		wg.Add(1)
		go func(service *Service) {
			defer wg.Done()
			runService(ctx, service)
		}(&services[i])
		time.Sleep(500 * time.Millisecond) // Pequeña pausa entre inicios
	}

	// Mostrar información de estado
	fmt.Printf("\n%s%s📊 Estado de los servicios:%s\n", Bold, Cyan, Reset)
	for _, service := range services {
		fmt.Printf("%s[%s]%s Ejecutándose en puerto %s\n",
			service.Color, service.Name, Reset, service.Port)
	}
	fmt.Printf("\n%s%s💡 Presiona Ctrl+C para detener todos los servicios%s\n\n", Bold, White, Reset)

	// Esperar señal de terminación
	<-sigChan
	fmt.Printf("\n%s%s🛑 Deteniendo servicios...%s\n", Bold, Red, Reset)

	// Cancelar contexto para detener todos los servicios
	cancel()

	// Esperar a que todos los servicios terminen
	wg.Wait()

	fmt.Printf("%s%s✅ Todos los servicios detenidos%s\n", Bold, Green, Reset)
}

// buildService compila un servicio
func buildService(service *Service) bool {
	fmt.Printf("%s[BUILD]%s Compilando %s...\n", Purple, Reset, service.Name)

	cmd := exec.Command("go", "build", "-o",
		fmt.Sprintf("./bin/%s", strings.ToLower(service.Name)),
		service.BuildPath)

	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("%s[ERROR]%s Error compilando %s: %v\n%s\n",
			Red, Reset, service.Name, err, string(output))
		return false
	}

	fmt.Printf("%s[BUILD]%s ✅ %s compilado exitosamente\n", Purple, Reset, service.Name)
	return true
}

// runService ejecuta un servicio y captura sus logs
func runService(ctx context.Context, service *Service) {
	binaryPath := fmt.Sprintf("./bin/%s", strings.ToLower(service.Name))

	// Crear comando con contexto
	cmd := exec.CommandContext(ctx, binaryPath)
	service.Cmd = cmd

	// Configurar pipes para stdout y stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		logWithPrefix(service, fmt.Sprintf("Error creando stdout pipe: %v", err), true)
		return
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		logWithPrefix(service, fmt.Sprintf("Error creando stderr pipe: %v", err), true)
		return
	}

	// Iniciar el proceso
	if err := cmd.Start(); err != nil {
		logWithPrefix(service, fmt.Sprintf("Error iniciando servicio: %v", err), true)
		return
	}

	logWithPrefix(service, "🚀 Servicio iniciado", false)

	// Leer stdout en goroutine separada
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			logWithPrefix(service, scanner.Text(), false)
		}
	}()

	// Leer stderr en goroutine separada
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			logWithPrefix(service, scanner.Text(), true)
		}
	}()

	// Esperar a que termine el proceso
	if err := cmd.Wait(); err != nil {
		// Solo loggear error si no fue por cancelación del contexto
		if ctx.Err() == nil {
			logWithPrefix(service, fmt.Sprintf("Proceso terminó con error: %v", err), true)
		}
	}

	logWithPrefix(service, "⏹️  Servicio detenido", false)
}

// logWithPrefix agrega un prefijo coloreado a los logs
func logWithPrefix(service *Service, message string, isError bool) {
	timestamp := time.Now().Format("15:04:05")
	prefix := fmt.Sprintf("%s[%s]%s", service.Color, service.Name, Reset)

	// Detectar si realmente es un error basándose en el contenido del mensaje
	isActualError := isError && (strings.Contains(strings.ToLower(message), "error") ||
		strings.Contains(strings.ToLower(message), "fatal") ||
		strings.Contains(strings.ToLower(message), "panic"))

	// Detectar mensajes de warning
	isWarning := strings.Contains(strings.ToLower(message), "warning") ||
		strings.Contains(strings.ToLower(message), "warn")

	if isActualError {
		fmt.Printf("%s %s %s[ERROR]%s %s\n", timestamp, prefix, Red, Reset, message)
	} else if isWarning {
		fmt.Printf("%s %s %s[WARN]%s %s\n", timestamp, prefix, Yellow, Reset, message)
	} else {
		fmt.Printf("%s %s %s\n", timestamp, prefix, message)
	}
}
