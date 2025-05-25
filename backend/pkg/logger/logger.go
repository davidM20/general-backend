package logger

import (
	"fmt"
	"log"
	"time"
)

// Códigos de color ANSI
const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorPurple = "\033[35m"
	ColorCyan   = "\033[36m"
	ColorWhite  = "\033[37m"
	ColorGray   = "\033[90m"

	// Colores de fondo
	BgRed    = "\033[41m"
	BgGreen  = "\033[42m"
	BgYellow = "\033[43m"
	BgBlue   = "\033[44m"
	BgPurple = "\033[45m"
	BgCyan   = "\033[46m"

	// Estilos
	Bold      = "\033[1m"
	Underline = "\033[4m"
)

// LogLevel representa el nivel de log
type LogLevel int

const (
	INFO LogLevel = iota
	WARN
	ERROR
	SUCCESS
	DEBUG
)

// getColorForLevel devuelve el color apropiado para cada nivel
func getColorForLevel(level LogLevel) string {
	switch level {
	case INFO:
		return ColorBlue
	case WARN:
		return ColorYellow
	case ERROR:
		return ColorRed
	case SUCCESS:
		return ColorGreen
	case DEBUG:
		return ColorGray
	default:
		return ColorWhite
	}
}

// getLevelText devuelve el texto del nivel
func getLevelText(level LogLevel) string {
	switch level {
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	case SUCCESS:
		return "SUCCESS"
	case DEBUG:
		return "DEBUG"
	default:
		return "LOG"
	}
}

// formatLog formatea un mensaje de log con colores
func formatLog(level LogLevel, component, message string) string {
	now := time.Now()
	timestamp := now.Format("2006/01/02 15:04:05")

	levelColor := getColorForLevel(level)
	levelText := getLevelText(level)

	// Formato: [TIMESTAMP] [LEVEL] [COMPONENT] MESSAGE
	return fmt.Sprintf("%s[%s]%s %s[%s]%s %s[%s]%s %s",
		ColorGray, timestamp, ColorReset,
		levelColor, levelText, ColorReset,
		ColorCyan, component, ColorReset,
		message)
}

// Info logs an info message
func Info(component, message string) {
	log.Println(formatLog(INFO, component, message))
}

// Warn logs a warning message
func Warn(component, message string) {
	log.Println(formatLog(WARN, component, message))
}

// Error logs an error message
func Error(component, message string) {
	log.Println(formatLog(ERROR, component, message))
}

// Success logs a success message
func Success(component, message string) {
	log.Println(formatLog(SUCCESS, component, message))
}

// Debug logs a debug message
func Debug(component, message string) {
	log.Println(formatLog(DEBUG, component, message))
}

// Infof logs a formatted info message
func Infof(component, format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	Info(component, message)
}

// Warnf logs a formatted warning message
func Warnf(component, format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	Warn(component, message)
}

// Errorf logs a formatted error message
func Errorf(component, format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	Error(component, message)
}

// Successf logs a formatted success message
func Successf(component, format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	Success(component, message)
}

// Debugf logs a formatted debug message
func Debugf(component, format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	Debug(component, message)
}

// ProxyLog logs proxy-specific messages with method and status colors
func ProxyLog(method, path, target, status string, duration time.Duration) {
	var statusColor string
	switch {
	case status[0] == '2': // 2xx
		statusColor = ColorGreen
	case status[0] == '3': // 3xx
		statusColor = ColorYellow
	case status[0] == '4': // 4xx
		statusColor = ColorRed
	case status[0] == '5': // 5xx
		statusColor = BgRed + ColorWhite
	default:
		statusColor = ColorWhite
	}

	var methodColor string
	switch method {
	case "GET":
		methodColor = ColorBlue
	case "POST":
		methodColor = ColorGreen
	case "PUT":
		methodColor = ColorYellow
	case "DELETE":
		methodColor = ColorRed
	case "PATCH":
		methodColor = ColorPurple
	default:
		methodColor = ColorWhite
	}

	message := fmt.Sprintf("%s%s%s %s → %s%s %s%s%s %s[%v]%s",
		methodColor, method, ColorReset,
		path,
		ColorCyan, target, ColorReset,
		statusColor, status, ColorReset,
		ColorGray, duration.Round(time.Millisecond), ColorReset)

	log.Println(formatLog(INFO, "PROXY", message))
}
