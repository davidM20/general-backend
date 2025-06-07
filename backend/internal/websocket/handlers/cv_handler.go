// REGLAS Y CONTEXTO PARA EL MANEJO DE CV

// CONTEXTO:
// Este archivo contiene los handlers de WebSocket para las operaciones relacionadas con el Currículum Vitae (CV) del usuario.
// Permite a los usuarios crear, actualizar y obtener secciones de su CV (habilidades, idiomas, experiencia laboral, certificaciones, proyectos).
// La información del CV se almacena en la base de datos y se transmite a través de mensajes WebSocket.

// REGLAS Y GUÍA:
// 1. Interacción con Servicios: Los handlers deben interactuar con la base de datos a través del servicio `cv_service` para mantener la separación de responsabilidades.
// 2. Deserialización de Payload: Cada handler que recibe datos del cliente debe deserializar el `msg.Payload` a la estructura de `models` correspondiente antes de pasarla al servicio.
// 3. Serialización de Respuesta: Los handlers que envían datos del CV al cliente (ej. HandleGetCV) deben serializar la respuesta utilizando las estructuras definidas en `wsmodels`.
// 4. Manejo de Errores: Implementar manejo de errores adecuado, registrando los errores en el servidor y enviando notificaciones de error significativas al cliente usando `conn.SendErrorNotification`.
// 5. ID de Usuario: Asegurarse de asociar el ID del usuario (`conn.ID`) a las operaciones de creación/actualización en la base de datos.
// 6. Tipos de Datos: Ser consciente de la diferencia entre las estructuras de datos usadas para la base de datos (`models`) y las usadas para la comunicación WebSocket (`wsmodels`). Asegurar el mapeo correcto entre ellas, especialmente al obtener datos de la DB (manejo de tipos nulos como `sql.NullTime`, `sql.NullInt64`, `sql.NullString`).
// 7. Logging: Usar el logger para registrar las acciones importantes y los errores.
// 8. PID: Incluir el PID del mensaje original del cliente en las notificaciones de error y respuestas si es posible para permitir al cliente correlacionar mensajes.

package handlers

import (
	"encoding/json"

	"github.com/davidM20/micro-service-backend-go.git/internal/db"
	"github.com/davidM20/micro-service-backend-go.git/internal/models"
	"github.com/davidM20/micro-service-backend-go.git/internal/websocket/services"
	"github.com/davidM20/micro-service-backend-go.git/internal/websocket/wsmodels"
	"github.com/davidM20/micro-service-backend-go.git/pkg/customws"
	"github.com/davidM20/micro-service-backend-go.git/pkg/customws/types"
	"github.com/davidM20/micro-service-backend-go.git/pkg/logger"
)

// HandleSetSkill maneja la solicitud para establecer una habilidad
func HandleSetSkill(conn *customws.Connection[wsmodels.WsUserData], msg types.ClientToServerMessage) error {
	logger.Infof("CV_HANDLER", "Estableciendo habilidad para UserID %d. PID: %s", conn.ID, msg.PID)

	var requestData struct {
		Data models.Skills `json:"data"`
	}
	payloadBytes, _ := json.Marshal(msg.Payload)
	if err := json.Unmarshal(payloadBytes, &requestData); err != nil {
		logger.Warnf("CV_HANDLER", "Error al decodificar payload de habilidad: %v", err)
		conn.SendErrorNotification(msg.PID, 400, "Payload de habilidad inválido.")
		return nil
	}
	skill := requestData.Data

	if skill.Skill == "" || skill.Level == "" {
		logger.Warnf("CV_HANDLER", "Validación fallida para set_skill: campos vacíos. UserID: %d", conn.ID)
		conn.SendErrorNotification(msg.PID, 400, "Los campos 'Skill' y 'Level' no pueden estar vacíos.")
		return nil
	}

	// Establecer el ID del usuario
	skill.PersonId = conn.ID

	// Usar el servicio de CV
	dbConn := db.GetDB()
	cvService := services.NewCVService(dbConn)

	if err := cvService.SetSkill(&skill); err != nil {
		logger.Errorf("CV_HANDLER", "Error al establecer habilidad: %v", err)
		conn.SendErrorNotification(msg.PID, 500, "Error interno al establecer habilidad.")
		return nil
	}

	// Enviar confirmación
	responseMsg := types.ServerToClientMessage{
		PID:        conn.Manager().Callbacks().GeneratePID(),
		Type:       "set_skill_success",
		FromUserID: 0,
		Payload:    map[string]interface{}{"status": "success"},
	}

	if err := conn.SendMessage(responseMsg); err != nil {
		logger.Errorf("CV_HANDLER", "Error al enviar confirmación de habilidad: %v", err)
	}

	return nil
}

// HandleSetLanguage maneja la solicitud para establecer un idioma
func HandleSetLanguage(conn *customws.Connection[wsmodels.WsUserData], msg types.ClientToServerMessage) error {
	logger.Infof("CV_HANDLER", "Estableciendo idioma para UserID %d. PID: %s", conn.ID, msg.PID)

	var requestData struct {
		Data models.Languages `json:"data"`
	}
	payloadBytes, _ := json.Marshal(msg.Payload)
	if err := json.Unmarshal(payloadBytes, &requestData); err != nil {
		logger.Warnf("CV_HANDLER", "Error al decodificar payload de idioma: %v", err)
		conn.SendErrorNotification(msg.PID, 400, "Payload de idioma inválido.")
		return nil
	}
	language := requestData.Data

	if language.Language == "" || language.Level == "" {
		logger.Warnf("CV_HANDLER", "Validación fallida para set_language: campos vacíos. UserID: %d", conn.ID)
		conn.SendErrorNotification(msg.PID, 400, "Los campos 'Language' y 'Level' no pueden estar vacíos.")
		return nil
	}

	// Establecer el ID del usuario
	language.PersonId = conn.ID

	// Usar el servicio de CV
	dbConn := db.GetDB()
	cvService := services.NewCVService(dbConn)

	if err := cvService.SetLanguage(&language); err != nil {
		logger.Errorf("CV_HANDLER", "Error al establecer idioma: %v", err)
		conn.SendErrorNotification(msg.PID, 500, "Error interno al establecer idioma.")
		return nil
	}

	// Enviar confirmación
	responseMsg := types.ServerToClientMessage{
		PID:        conn.Manager().Callbacks().GeneratePID(),
		Type:       "set_language_success",
		FromUserID: 0,
		Payload:    map[string]interface{}{"status": "success"},
	}

	if err := conn.SendMessage(responseMsg); err != nil {
		logger.Errorf("CV_HANDLER", "Error al enviar confirmación de idioma: %v", err)
	}

	return nil
}

// HandleSetWorkExperience maneja la solicitud para establecer una experiencia laboral
func HandleSetWorkExperience(conn *customws.Connection[wsmodels.WsUserData], msg types.ClientToServerMessage) error {
	logger.Infof("CV_HANDLER", "Estableciendo experiencia laboral para UserID %d. PID: %s", conn.ID, msg.PID)

	var requestData struct {
		Data models.WorkExperience `json:"data"`
	}
	payloadBytes, _ := json.Marshal(msg.Payload)
	if err := json.Unmarshal(payloadBytes, &requestData); err != nil {
		logger.Warnf("CV_HANDLER", "Error al decodificar payload de experiencia laboral: %v", err)
		conn.SendErrorNotification(msg.PID, 400, "Payload de experiencia laboral inválido.")
		return nil
	}
	experience := requestData.Data

	if experience.Company == "" || experience.Position == "" {
		logger.Warnf("CV_HANDLER", "Validación fallida para set_work_experience: campos vacíos. UserID: %d", conn.ID)
		conn.SendErrorNotification(msg.PID, 400, "Los campos 'Company' y 'Position' no pueden estar vacíos.")
		return nil
	}

	// Establecer el ID del usuario
	experience.PersonId = conn.ID

	// Usar el servicio de CV
	dbConn := db.GetDB()
	cvService := services.NewCVService(dbConn)

	if err := cvService.SetWorkExperience(&experience); err != nil {
		logger.Errorf("CV_HANDLER", "Error al establecer experiencia laboral: %v", err)
		conn.SendErrorNotification(msg.PID, 500, "Error interno al establecer experiencia laboral.")
		return nil
	}

	// Enviar confirmación
	responseMsg := types.ServerToClientMessage{
		PID:        conn.Manager().Callbacks().GeneratePID(),
		Type:       "set_work_experience_success",
		FromUserID: 0,
		Payload:    map[string]interface{}{"status": "success"},
	}

	if err := conn.SendMessage(responseMsg); err != nil {
		logger.Errorf("CV_HANDLER", "Error al enviar confirmación de experiencia laboral: %v", err)
	}

	return nil
}

// HandleSetCertification maneja la solicitud para establecer una certificación
func HandleSetCertification(conn *customws.Connection[wsmodels.WsUserData], msg types.ClientToServerMessage) error {
	logger.Infof("CV_HANDLER", "Estableciendo certificación para UserID %d. PID: %s", conn.ID, msg.PID)

	var requestData struct {
		Data models.Certifications `json:"data"`
	}
	payloadBytes, _ := json.Marshal(msg.Payload)
	if err := json.Unmarshal(payloadBytes, &requestData); err != nil {
		logger.Warnf("CV_HANDLER", "Error al decodificar payload de certificación: %v", err)
		conn.SendErrorNotification(msg.PID, 400, "Payload de certificación inválido.")
		return nil
	}
	certification := requestData.Data

	if certification.Certification == "" || certification.Institution == "" {
		logger.Warnf("CV_HANDLER", "Validación fallida para set_certification: campos vacíos. UserID: %d", conn.ID)
		conn.SendErrorNotification(msg.PID, 400, "Los campos 'Certification' y 'Institution' no pueden estar vacíos.")
		return nil
	}

	// Establecer el ID del usuario
	certification.PersonId = conn.ID

	// Usar el servicio de CV
	dbConn := db.GetDB()
	cvService := services.NewCVService(dbConn)

	if err := cvService.SetCertification(&certification); err != nil {
		logger.Errorf("CV_HANDLER", "Error al establecer certificación: %v", err)
		conn.SendErrorNotification(msg.PID, 500, "Error interno al establecer certificación.")
		return nil
	}

	// Enviar confirmación
	responseMsg := types.ServerToClientMessage{
		PID:        conn.Manager().Callbacks().GeneratePID(),
		Type:       "set_certification_success",
		FromUserID: 0,
		Payload:    map[string]interface{}{"status": "success"},
	}

	if err := conn.SendMessage(responseMsg); err != nil {
		logger.Errorf("CV_HANDLER", "Error al enviar confirmación de certificación: %v", err)
	}

	return nil
}

// HandleSetProject maneja la solicitud para establecer un proyecto
func HandleSetProject(conn *customws.Connection[wsmodels.WsUserData], msg types.ClientToServerMessage) error {
	logger.Infof("CV_HANDLER", "Estableciendo proyecto para UserID %d. PID: %s", conn.ID, msg.PID)

	var requestData struct {
		Data models.Project `json:"data"`
	}
	payloadBytes, _ := json.Marshal(msg.Payload)
	if err := json.Unmarshal(payloadBytes, &requestData); err != nil {
		logger.Warnf("CV_HANDLER", "Error al decodificar payload de proyecto: %v", err)
		conn.SendErrorNotification(msg.PID, 400, "Payload de proyecto inválido.")
		return nil
	}
	project := requestData.Data

	if project.Title == "" || project.Role == "" {
		logger.Warnf("CV_HANDLER", "Validación fallida para set_project: campos vacíos. UserID: %d", conn.ID)
		conn.SendErrorNotification(msg.PID, 400, "Los campos 'Title' y 'Role' no pueden estar vacíos.")
		return nil
	}

	// Establecer el ID del usuario
	project.PersonID = conn.ID

	// Usar el servicio de CV
	dbConn := db.GetDB()
	cvService := services.NewCVService(dbConn)

	if err := cvService.SetProject(&project); err != nil {
		logger.Errorf("CV_HANDLER", "Error al establecer proyecto: %v", err)
		conn.SendErrorNotification(msg.PID, 500, "Error interno al establecer proyecto.")
		return nil
	}

	// Enviar confirmación
	responseMsg := types.ServerToClientMessage{
		PID:        conn.Manager().Callbacks().GeneratePID(),
		Type:       "set_project_success",
		FromUserID: 0,
		Payload:    map[string]interface{}{"status": "success"},
	}

	if err := conn.SendMessage(responseMsg); err != nil {
		logger.Errorf("CV_HANDLER", "Error al enviar confirmación de proyecto: %v", err)
	}

	return nil
}

// HandleSetEducation maneja la solicitud para establecer una educación
func HandleSetEducation(conn *customws.Connection[wsmodels.WsUserData], msg types.ClientToServerMessage) error {
	logger.Infof("CV_HANDLER", "Estableciendo educación para UserID %d. PID: %s", conn.ID, msg.PID)

	var requestData struct {
		Data models.Education `json:"data"`
	}
	payloadBytes, _ := json.Marshal(msg.Payload)
	if err := json.Unmarshal(payloadBytes, &requestData); err != nil {
		logger.Warnf("CV_HANDLER", "Error al decodificar payload de educación: %v", err)
		conn.SendErrorNotification(msg.PID, 400, "Payload de educación inválido.")
		return nil
	}
	education := requestData.Data

	if education.Institution == "" || education.Degree == "" {
		logger.Warnf("CV_HANDLER", "Validación fallida para set_education: campos vacíos. UserID: %d", conn.ID)
		conn.SendErrorNotification(msg.PID, 400, "Los campos 'Institution' y 'Degree' no pueden estar vacíos.")
		return nil
	}

	// Establecer el ID del usuario
	education.PersonId = conn.ID

	// Usar el servicio de CV
	dbConn := db.GetDB()
	cvService := services.NewCVService(dbConn)

	if err := cvService.SetEducation(&education); err != nil {
		logger.Errorf("CV_HANDLER", "Error al establecer educación: %v", err)
		conn.SendErrorNotification(msg.PID, 500, "Error interno al establecer educación.")
		return nil
	}

	// Enviar confirmación
	responseMsg := types.ServerToClientMessage{
		PID:        conn.Manager().Callbacks().GeneratePID(),
		Type:       "set_education_success",
		FromUserID: 0,
		Payload:    map[string]interface{}{"status": "success", "action": "set_education"},
	}

	if err := conn.SendMessage(responseMsg); err != nil {
		logger.Errorf("CV_HANDLER", "Error al enviar confirmación de educación: %v", err)
	}

	return nil
}

// HandleGetCV maneja la solicitud para obtener el CV completo
func HandleGetCV(conn *customws.Connection[wsmodels.WsUserData], msg types.ClientToServerMessage) error {
	logger.Infof("CV_HANDLER", "Obteniendo CV para UserID %d. PID: %s", conn.ID, msg.PID)

	// Usar el servicio de CV
	dbConn := db.GetDB()
	cvService := services.NewCVService(dbConn)

	cv, err := cvService.GetCV(conn.ID)
	if err != nil {
		logger.Errorf("CV_HANDLER", "Error al obtener CV: %v", err)
		conn.SendErrorNotification(msg.PID, 500, "Error interno al obtener CV.")
		return nil
	}

	// Enviar el CV
	responseMsg := types.ServerToClientMessage{
		PID:        conn.Manager().Callbacks().GeneratePID(),
		Type:       "cv_data",
		FromUserID: 0,
		Payload:    cv,
	}

	if err := conn.SendMessage(responseMsg); err != nil {
		logger.Errorf("CV_HANDLER", "Error al enviar CV: %v", err)
	}

	return nil
}
