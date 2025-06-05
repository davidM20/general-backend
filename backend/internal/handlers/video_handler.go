/*
 * =====================================================================================
 * HANDLER PARA LA SUBIDA Y STREAMING DE ARCHIVOS DE VIDEO
 * =====================================================================================
 *
 * DESCRIPCIÓN GENERAL:
 * --------------------
 * Este archivo maneja todas las solicitudes HTTP relacionadas con videos, incluyendo
 * la subida de nuevos videos y el servicio de manifiestos y segmentos HLS para streaming.
 * Interactúa con VideoUploadService para la lógica de negocio (subida, transcodificación),
 * con el paquete de queries para el acceso a la base de datos, y con cloudclient para
 * la interacción con Google Cloud Storage (GCS).
 *
 * =====================================================================================
 * GUÍA DE MANTENIMIENTO Y EXTENSIÓN
 * =====================================================================================
 *
 * ESTRUCTURA Y DEPENDENCIAS:
 * --------------------------
 * - VideoHandler: Contiene referencias a VideoUploadService, *sql.DB (para queries directas
 *   como GetMultimediaByContentID) y *config.Config (para JWT secret, etc.).
 * - NewVideoHandler: Constructor que inicializa el handler con sus dependencias.
 *
 * FLUJO DE SECUENCIA DE FUNCIONES PÚBLICAS:
 * -----------------------------------------
 * 1. UploadVideo (POST /api/v1/videos/upload):
 *    a. Extrae `userID` del contexto JWT (ruta protegida).
 *    b. Parsea el formulario `multipart/form-data` para obtener el archivo "video".
 *    c. Valida el archivo (tamaño máximo) y lo pasa a `h.videoService.ProcessAndUploadVideo`.
 *       - El servicio se encarga de: subir el original a GCS, guardar el registro
 *         inicial en `Multimedia` con estado "uploaded", y disparar la
 *         transcodificación asíncrona (actualmente simulada).
 *    d. Responde con `202 Accepted` y los detalles de la subida inicial (ContentID, URL original).
 *
 * 2. StreamVideoMasterPlaylist (GET /api/v1/videos/stream/{contentID}/master.m3u8?token=<jwt>):
 *    a. Extrae `contentID` de la ruta URL.
 *    b. Extrae y valida el token JWT del query parameter "token".
 *    c. Llama a `queries.GetMultimediaByContentID(h.db, contentID)` para obtener los detalles del video,
 *       incluyendo los paths de los manifiestos HLS y el estado de procesamiento.
 *    d. Verifica que el `ProcessingStatus` sea "completed". Si no, devuelve error (ej. 409 Conflict).
 *    e. Construye dinámicamente el manifiesto HLS maestro (`#EXTM3U`):
 *       - Incluye las directivas `#EXT-X-STREAM-INF` para cada variante disponible (1080p, 720p, 480p)
 *         si sus paths de manifiesto son válidos en la BD.
 *       - Los paths a los manifiestos de calidad en el maestro son relativos (ej. "1080p/playlist.m3u8").
 *         La función `extractRelativePathFromGCS` ayuda a obtener estos paths relativos.
 *    f. Establece el `Content-Type` a `application/vnd.apple.mpegurl` y sirve el manifiesto.
 *
 * 3. StreamVideoVariant (GET /api/v1/videos/stream/{contentID}/{quality}/{fileName}?token=<jwt>):
 *    a. Extrae `contentID`, `quality` (ej. "1080p"), y `fileName` (ej. "playlist.m3u8" o "segment001.ts")
 *       de la ruta URL.
 *    b. Extrae y valida el token JWT del query parameter "token".
 *    c. Construye la ruta completa al objeto en GCS (ej. "videos/{contentID}/{quality}/{fileName}").
 *    d. Llama a `cloudclient.DownloadFile` para descargar el archivo de GCS.
 *       - Se asume que `DownloadFile` devuelve `([]byte, error)`.
 *    e. Determina el `Content-Type` apropiado (`application/vnd.apple.mpegurl` para .m3u8,
 *       `video/MP2T` para .ts).
 *    f. Establece los headers (`Content-Type`, `Content-Length`, `Access-Control-Allow-Origin`).
 *    g. Escribe los bytes del archivo en `http.ResponseWriter` (usando `bytes.NewReader`).
 *
 * REGLAS Y CONSIDERACIONES PARA FUTUROS CAMBIOS:
 * ---------------------------------------------
 * 1.  AUTENTICACIÓN: Las rutas de streaming usan un token JWT en el query param ("token").
 *     Esto es diferente de las rutas de API protegidas estándar que usan el header `Authorization`.
 *     Mantener esta consistencia o documentar cualquier cambio.
 *
 * 2.  EXTRACCIÓN DE PARÁMETROS URL: La extracción actual de `contentID`, `quality`, `fileName`
 *     de `r.URL.Path` mediante `strings.Split` es funcional pero podría ser más robusta
 *     si se usaran las capacidades de extracción de parámetros de `gorilla/mux` de forma más directa
 *     (requeriría que las rutas se definan con `mux.Vars(r)` en mente, lo cual se hace
 *     automáticamente por `gorilla/mux` si las rutas tienen placeholders como `{contentID}`).
 *     Actualmente, las rutas en `api_routes.go` están definidas con esos placeholders,
 *     así que idealmente se usaría `mux.Vars(r)["paramName"]`.
 *
 * 3.  DEPENDENCIA DE VideoUploadService: Este handler delega la lógica compleja de subida y
 *     el inicio de la transcodificación al `VideoUploadService`. No se debe duplicar esa lógica aquí.
 *
 * 4.  ACCESO A BD: Para obtener el estado y los paths HLS, se usa `queries.GetMultimediaByContentID`.
 *     Asegurar que los modelos y queries estén sincronizados con la estructura de la tabla `Multimedia`.
 *
 * 5.  INTERACCIÓN CON GCS (cloudclient): El streaming actúa como un proxy, descargando archivos
 *     de GCS y sirviéndolos. La implementación actual asume que `cloudclient.DownloadFile`
 *     devuelve `([]byte, error)`. Si esta firma cambia (ej. a `io.ReadCloser`), se deberá ajustar
 *     `StreamVideoVariant` (manejo de `defer Close()`, `io.Copy` directo, y obtención de
 *     `Content-Length` que sería más compleja).
 *
 * 6.  TRANSCODIFICACIÓN ASÍNCRONA: El streaming solo funciona para videos cuyo `ProcessingStatus`
 *     es "completed". El handler `StreamVideoMasterPlaylist` verifica esto. Considerar si se
 *     necesitan mecanismos más sofisticados para informar al cliente sobre el progreso.
 *
 * 7.  PATHS HLS: La función `extractRelativePathFromGCS` es crucial para construir correctamente
 *     el manifiesto maestro con paths relativos a las variantes. Asegurar que su lógica siga
 *     siendo válida si cambian los formatos de `HLSManifestBaseURL` o los paths de los manifiestos
 *     de calidad almacenados en la BD.
 *
 * 8.  HEADERS HTTP: Prestar atención a los `Content-Type` correctos para HLS, y a `Content-Length`
 *     y `Access-Control-Allow-Origin` para el correcto funcionamiento con reproductores de video.
 *
 * 9.  MANEJO DE ERRORES: Usar `logger` para registrar errores y devolver respuestas HTTP
 *     apropiadas al cliente (400, 401, 403, 404, 409, 500 según el caso).
 */

package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/davidM20/micro-service-backend-go.git/internal/auth"
	"github.com/davidM20/micro-service-backend-go.git/internal/config"
	"github.com/davidM20/micro-service-backend-go.git/internal/db/queries"
	"github.com/davidM20/micro-service-backend-go.git/internal/middleware"
	"github.com/davidM20/micro-service-backend-go.git/internal/services"
	"github.com/davidM20/micro-service-backend-go.git/pkg/cloudclient"
	"github.com/davidM20/micro-service-backend-go.git/pkg/logger"
	gcsErrors "google.golang.org/api/googleapi"
)

/*
 * ===================================================
 * HANDLER PARA LA SUBIDA DE ARCHIVOS DE VIDEO
 * ===================================================
 */

// VideoHandler maneja las solicitudes de subida y streaming de video.
type VideoHandler struct {
	videoService *services.VideoUploadService
	db           *sql.DB
	cfg          *config.Config
}

// NewVideoHandler crea una nueva instancia de VideoHandler.
func NewVideoHandler(videoService *services.VideoUploadService, db *sql.DB, cfg *config.Config) *VideoHandler {
	return &VideoHandler{videoService: videoService, db: db, cfg: cfg}
}

// UploadVideo es el método que maneja la petición POST para subir un archivo de video.
func (h *VideoHandler) UploadVideo(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDContextKey).(int64)
	if !ok || userID == 0 {
		logger.Warn("UploadVideo.Auth", "No se pudo obtener userID del contexto o es inválido.")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Usuario no autenticado o ID de usuario inválido."})
		return
	}

	// Parsear el formulario multipart, limitando el tamaño total.
	// El servicio VideoUploadService.MaxVideoSize (500MB) es el límite real del archivo.
	// Aquí ponemos un límite ligeramente superior para el request completo.
	if err := r.ParseMultipartForm(services.MaxVideoSize + (10 * 1024 * 1024)); err != nil { // Ej: 510MB
		logger.Errorf("UploadVideo.ParseForm", "Error parseando multipart form: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Solicitud inválida o demasiado grande: " + err.Error()})
		return
	}

	file, handler, err := r.FormFile("video") // "video" es el nombre del campo en el form-data
	if err != nil {
		logger.Errorf("UploadVideo.FormFile", "Error obteniendo el archivo 'video' del formulario: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Error al recibir el archivo de video: " + err.Error()})
		return
	}
	defer file.Close()

	logger.Infof("UploadVideo", "Recibida solicitud de subida de video del usuario %d, archivo: %s, tamaño: %d", userID, handler.Filename, handler.Size)

	uploadDetails, err := h.videoService.ProcessAndUploadVideo(r.Context(), userID, file, handler)
	if err != nil {
		logger.Errorf("UploadVideo.ServiceCall", "Error procesando el video para el usuario %d: %v", userID, err)
		w.Header().Set("Content-Type", "application/json")
		// El código de estado podría depender del tipo de error (ej. BadRequest por tipo no soportado)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Error al procesar el archivo de video: " + err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted) // 202 Accepted, ya que el procesamiento (transcodificación) es asíncrono
	json.NewEncoder(w).Encode(uploadDetails)
}

// StreamVideoMasterPlaylist sirve el manifiesto HLS maestro para un video.
// La ruta esperada es /api/v1/videos/stream/{contentID}/master.m3u8?token=<jwt>
func (h *VideoHandler) StreamVideoMasterPlaylist(w http.ResponseWriter, r *http.Request) {
	// Extracción simple de contentID del path:
	// /api/v1/videos/stream/{contentID}/master.m3u8
	pathSegments := strings.Split(r.URL.Path, "/")
	var contentID string
	if len(pathSegments) >= 6 && pathSegments[1] == "api" && pathSegments[2] == "v1" && pathSegments[3] == "videos" && pathSegments[4] == "stream" && pathSegments[6] == "master.m3u8" {
		contentID = pathSegments[5]
	}

	if contentID == "" {
		logger.Warnf("StreamVideoMasterPlaylist.ExtractParam", "No se pudo extraer contentID del path: %s", r.URL.Path)
		http.Error(w, "ID de contenido inválido en la ruta.", http.StatusBadRequest)
		return
	}

	tokenStr := r.URL.Query().Get("token")
	if tokenStr == "" {
		logger.Warn("StreamVideoMasterPlaylist.Auth", "Token no proporcionado en query params.")
		http.Error(w, "Token de autorización requerido.", http.StatusUnauthorized)
		return
	}

	_, err := auth.ValidateJWT(tokenStr, []byte(h.cfg.JwtSecret))
	if err != nil {
		logger.Warnf("StreamVideoMasterPlaylist.Auth", "Token inválido para contentID %s: %v", contentID, err)
		http.Error(w, "Token de autorización inválido.", http.StatusUnauthorized)
		return
	}

	multimedia, err := queries.GetMultimediaByContentID(h.db, contentID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logger.Warnf("StreamVideoMasterPlaylist.DB", "Video no encontrado con contentID %s: %v", contentID, err)
			http.NotFound(w, r)
			return
		}
		logger.Errorf("StreamVideoMasterPlaylist.DB", "Error obteniendo video de DB para contentID %s: %v", contentID, err)
		http.Error(w, "Error interno del servidor al obtener información del video.", http.StatusInternalServerError)
		return
	}

	if !multimedia.ProcessingStatus.Valid || multimedia.ProcessingStatus.String != services.ProcessingStatusCompleted {
		status := "desconocido"
		if multimedia.ProcessingStatus.Valid {
			status = multimedia.ProcessingStatus.String
		}
		logger.Warnf("StreamVideoMasterPlaylist.Processing", "Video con contentID %s no está listo para streaming. Estado: %s", contentID, status)
		http.Error(w, fmt.Sprintf("El video aún se está procesando (estado: %s). Intente más tarde.", status), http.StatusConflict) // 409 Conflict
		return
	}

	var masterPlaylist strings.Builder
	masterPlaylist.WriteString("#EXTM3U\n")
	masterPlaylist.WriteString("#EXT-X-VERSION:3\n")

	addedVariant := false
	if multimedia.HLSManifest1080p.Valid && multimedia.HLSManifest1080p.String != "" {
		relativePath1080 := extractRelativePathFromGCS(multimedia.HLSManifest1080p.String, multimedia.HLSManifestBaseURL.String)
		if relativePath1080 != "" {
			masterPlaylist.WriteString(fmt.Sprintf("#EXT-X-STREAM-INF:BANDWIDTH=5000000,RESOLUTION=1920x1080,CODECS=\"avc1.640028,mp4a.40.2\"\n%s\n", relativePath1080))
			addedVariant = true
		}
	}
	if multimedia.HLSManifest720p.Valid && multimedia.HLSManifest720p.String != "" {
		relativePath720 := extractRelativePathFromGCS(multimedia.HLSManifest720p.String, multimedia.HLSManifestBaseURL.String)
		if relativePath720 != "" {
			masterPlaylist.WriteString(fmt.Sprintf("#EXT-X-STREAM-INF:BANDWIDTH=2800000,RESOLUTION=1280x720,CODECS=\"avc1.64001F,mp4a.40.2\"\n%s\n", relativePath720))
			addedVariant = true
		}
	}
	if multimedia.HLSManifest480p.Valid && multimedia.HLSManifest480p.String != "" {
		relativePath480 := extractRelativePathFromGCS(multimedia.HLSManifest480p.String, multimedia.HLSManifestBaseURL.String)
		if relativePath480 != "" {
			masterPlaylist.WriteString(fmt.Sprintf("#EXT-X-STREAM-INF:BANDWIDTH=1400000,RESOLUTION=854x480,CODECS=\"avc1.4D401E,mp4a.40.2\"\n%s\n", relativePath480))
			addedVariant = true
		}
	}

	if !addedVariant {
		logger.Warnf("StreamVideoMasterPlaylist.NoVariants", "No se encontraron variantes HLS válidas para contentID %s. Paths: 1080p: %s, 720p: %s, 480p: %s, Base: %s", contentID, multimedia.HLSManifest1080p.String, multimedia.HLSManifest720p.String, multimedia.HLSManifest480p.String, multimedia.HLSManifestBaseURL.String)
		http.Error(w, "No hay variantes de video disponibles para streaming.", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
	w.Header().Set("Access-Control-Allow-Origin", "*") // CORS
	fmt.Fprint(w, masterPlaylist.String())
	logger.Infof("StreamVideoMasterPlaylist", "Manifiesto maestro HLS servido para contentID %s", contentID)
}

// extractRelativePathFromGCS extrae, por ejemplo, "1080p/playlist.m3u8"
// de "gs://bucket-name/videos/uuid-content/1080p/playlist.m3u8" (fullGCSPath)
// y "gs://bucket-name/videos/uuid-content/" (baseGCSPath).
// O si los paths no incluyen el bucket y son solo "videos/uuid-content/1080p/playlist.m3u8"
func extractRelativePathFromGCS(fullGCSPath, baseGCSPath string) string {
	// Función interna para quitar el prefijo gs://bucket-name/
	stripGCSBucketPrefix := func(path string) string {
		if strings.HasPrefix(path, "gs://") {
			parts := strings.SplitN(path, "/", 4)
			if len(parts) == 4 {
				return parts[3] // Devuelve la parte después de gs://bucket-name/
			}
		}
		return path
	}

	normFullGCSPath := stripGCSBucketPrefix(fullGCSPath)
	normBaseGCSPath := stripGCSBucketPrefix(baseGCSPath)

	// Asegurar que el base path termine con / para una comparación de prefijo correcta
	if normBaseGCSPath != "" && !strings.HasSuffix(normBaseGCSPath, "/") {
		normBaseGCSPath += "/"
	}

	if normBaseGCSPath != "" && strings.HasPrefix(normFullGCSPath, normBaseGCSPath) {
		return strings.TrimPrefix(normFullGCSPath, normBaseGCSPath)
	} else {
		// Fallback: si baseGCSPath es vacío o no es un prefijo, intentar obtener los últimos dos segmentos
		// Esto es útil si fullGCSPath es algo como "1080p/playlist.m3u8" directamente,
		// o si baseGCSPath no coincide por alguna razón.
		parts := strings.Split(normFullGCSPath, "/")
		if len(parts) >= 2 {
			// Heurística: si el penúltimo segmento es una calidad conocida
			penultimate := parts[len(parts)-2]
			if penultimate == "1080p" || penultimate == "720p" || penultimate == "480p" {
				return strings.Join(parts[len(parts)-2:], "/")
			}
		}
		// Si todo falla, devuelve el path completo normalizado (sin gs://...) o el original si no tenía prefijo.
		logger.Warnf("extractRelativePathFromGCS", "No se pudo determinar path relativo de forma limpia. Full: '%s', Base: '%s', Result: '%s'", fullGCSPath, baseGCSPath, normFullGCSPath)
		return normFullGCSPath
	}
}

// StreamVideoVariant sirve un manifiesto de calidad HLS o un segmento de video.
// La ruta esperada es /api/v1/videos/stream/{contentID}/{quality}/{fileName}?token=<jwt>
func (h *VideoHandler) StreamVideoVariant(w http.ResponseWriter, r *http.Request) {
	// Path: /api/v1/videos/stream/{contentID}/{quality}/{fileName}
	pathSegments := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/v1/videos/stream/"), "/")
	if len(pathSegments) < 3 { // Esperamos contentID, quality, fileName (fileName puede tener '/')
		logger.Warnf("StreamVideoVariant.ExtractParam", "Path inválido, no se pudieron extraer parámetros: %s. Segments: %v", r.URL.Path, pathSegments)
		http.Error(w, "Ruta inválida para variante de video.", http.StatusBadRequest)
		return
	}
	contentID := pathSegments[0]
	quality := pathSegments[1]                      // "1080p", "720p", "480p"
	fileName := strings.Join(pathSegments[2:], "/") // Reconstruye el resto del path como fileName

	if contentID == "" || quality == "" || fileName == "" {
		logger.Warnf("StreamVideoVariant.ExtractParam", "Parámetros de path incompletos: contentID='%s', quality='%s', fileName='%s'", contentID, quality, fileName)
		http.Error(w, "Parámetros de ruta incompletos.", http.StatusBadRequest)
		return
	}

	tokenStr := r.URL.Query().Get("token")
	if tokenStr == "" {
		logger.Warn("StreamVideoVariant.Auth", "Token no proporcionado en query params.")
		http.Error(w, "Token de autorización requerido.", http.StatusUnauthorized)
		return
	}
	_, err := auth.ValidateJWT(tokenStr, []byte(h.cfg.JwtSecret))
	if err != nil {
		logger.Warnf("StreamVideoVariant.Auth", "Token inválido: %v", err)
		http.Error(w, "Token de autorización inválido.", http.StatusUnauthorized)
		return
	}

	// Los paths en GCS (según video_service.go) son como: videos/{contentID}/{quality}/playlist.m3u8 o videos/{contentID}/{quality}/segmentXXX.ts
	gcsObjectPath := fmt.Sprintf("videos/%s/%s/%s", contentID, quality, fileName)
	logger.Infof("StreamVideoVariant", "Solicitud para servir variante: GCS Path %s", gcsObjectPath)

	fileBytes, err := cloudclient.DownloadFile(r.Context(), gcsObjectPath) // Asumiendo que DownloadFile devuelve []byte, error
	if err != nil {
		// Intentar detectar error de "objeto no encontrado" de GCS
		var gae *gcsErrors.Error
		if errors.As(err, &gae) && gae.Code == http.StatusNotFound {
			logger.Warnf("StreamVideoVariant.GCS", "Archivo no encontrado en GCS: %s. API Error: %v", gcsObjectPath, gae)
			http.NotFound(w, r)
			return
		} else if strings.Contains(strings.ToLower(err.Error()), "object") && strings.Contains(strings.ToLower(err.Error()), "not found") {
			// Fallback para mensajes de error genéricos de "no encontrado"
			logger.Warnf("StreamVideoVariant.GCS", "Archivo no encontrado en GCS (string match): %s. Error: %v", gcsObjectPath, err)
			http.NotFound(w, r)
			return
		}

		logger.Errorf("StreamVideoVariant.GCS", "Error descargando archivo %s de GCS: %v", gcsObjectPath, err)
		http.Error(w, "Error interno al obtener el archivo de video.", http.StatusInternalServerError)
		return
	}

	contentType := "application/octet-stream" // Default
	if strings.HasSuffix(fileName, ".m3u8") {
		contentType = "application/vnd.apple.mpegurl"
	} else if strings.HasSuffix(fileName, ".ts") {
		contentType = "video/MP2T"
	}

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(fileBytes)))
	w.Header().Set("Access-Control-Allow-Origin", "*") // CORS

	_, copyErr := io.Copy(w, bytes.NewReader(fileBytes)) // Usar bytes.NewReader
	if copyErr != nil {
		logger.Errorf("StreamVideoVariant.Copy", "Error sirviendo archivo %s: %v", gcsObjectPath, copyErr)
	}
	logger.Infof("StreamVideoVariant", "Archivo %s servido (Content-Type: %s, Size: %d)", gcsObjectPath, contentType, len(fileBytes))
}
