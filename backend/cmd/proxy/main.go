package main

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/davidM20/micro-service-backend-go.git/internal/config"
	"github.com/davidM20/micro-service-backend-go.git/pkg/logger"
	"github.com/joho/godotenv"
	"github.com/koding/websocketproxy"
)

// Wrapper para logging de respuestas
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	startTime  time.Time
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Implementar http.Hijacker para soporte de WebSocket
func (rw *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hijacker, ok := rw.ResponseWriter.(http.Hijacker); ok {
		return hijacker.Hijack()
	}
	return nil, nil, fmt.Errorf("ResponseWriter no implementa http.Hijacker")
}

// corsMiddleware agrega headers CORS para permitir todo
func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Permitir todos los orígenes
		w.Header().Set("Access-Control-Allow-Origin", "*")

		// Permitir todos los métodos
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH, HEAD")

		// Permitir headers específicos (incluyendo los necesarios para WebSocket)
		w.Header().Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept, Authorization, Cache-Control, X-File-Name, Sec-WebSocket-Key, Sec-WebSocket-Version, Sec-WebSocket-Extensions, Sec-WebSocket-Accept, Sec-WebSocket-Protocol, Connection, Upgrade")

		// Permitir credentials
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		// Headers para caché de preflight
		w.Header().Set("Access-Control-Max-Age", "3600")

		// Manejar preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Continuar con el handler original
		next(w, r)
	}
}

func main() {
	// Cargar .env (opcional)
	err := godotenv.Load()
	if err != nil {
		logger.Warn("CONFIG", "Could not load .env file. Using environment variables directly.")
	}

	// Cargar configuración
	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Errorf("CONFIG", "Failed to load configuration: %v", err)
		return
	}

	// Parsear URLs de destino
	apiURL, err := url.Parse(fmt.Sprintf("http://localhost:%s", cfg.ApiPort))
	if err != nil {
		logger.Errorf("CONFIG", "Failed to parse API URL: %v", err)
		return
	}

	wsURL, err := url.Parse(fmt.Sprintf("ws://localhost:%s", cfg.WsPort))
	if err != nil {
		logger.Errorf("CONFIG", "Failed to parse WebSocket URL: %v", err)
		return
	}

	// Crear proxies inversos
	apiProxy := httputil.NewSingleHostReverseProxy(apiURL)
	wsProxy := websocketproxy.NewProxy(wsURL)

	// Modificar el director del proxy API
	apiProxy.Director = func(req *http.Request) {
		req.URL.Scheme = apiURL.Scheme
		req.URL.Host = apiURL.Host
		req.Host = apiURL.Host
		logger.Infof("PROXY_DIRECTOR", "Authorization Header: %s", req.Header.Get("Authorization"))
	}

	// Definir el manejador principal del proxy con CORS
	http.HandleFunc("/", corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()

		// Wrapper para capturar el código de estado
		rw := &responseWriter{
			ResponseWriter: w,
			statusCode:     200, // Default
			startTime:      startTime,
		}

		if strings.HasPrefix(r.URL.Path, "/api/") {
			logger.Infof("PROXY", "→ API: %s %s", r.Method, r.URL.Path)
			apiProxy.ServeHTTP(rw, r)
			duration := time.Since(startTime)
			logger.ProxyLog(r.Method, r.URL.Path, apiURL.String(), fmt.Sprintf("%d", rw.statusCode), duration)
		} else if strings.HasPrefix(r.URL.Path, "/ws") {
			logger.Infof("PROXY", "→ WebSocket: %s %s", r.Method, r.URL.Path)
			wsProxy.ServeHTTP(rw, r)
			duration := time.Since(startTime)
			logger.ProxyLog(r.Method, r.URL.Path, wsURL.String(), "101", duration) // WebSocket upgrade
		} else {
			http.NotFound(rw, r)
			duration := time.Since(startTime)
			logger.Warnf("PROXY", "Path not found: %s", r.URL.Path)
			logger.ProxyLog(r.Method, r.URL.Path, "NOT_FOUND", "404", duration)
		}
	}))

	// Iniciar el servidor proxy
	serverAddr := cfg.ProxyPort
	logger.Successf("PROXY", "🚀 Reverse Proxy server starting on port %s with CORS enabled", serverAddr)
	logger.Infof("PROXY", "📡 API proxy: http://localhost:%s/api/* → %s", serverAddr, apiURL)
	logger.Infof("PROXY", "🔌 WebSocket proxy: http://localhost:%s/ws → %s", serverAddr, wsURL)

	if err := http.ListenAndServe(":"+serverAddr, nil); err != nil {
		logger.Errorf("PROXY", "Failed to start proxy server: %v", err)
	}
}
