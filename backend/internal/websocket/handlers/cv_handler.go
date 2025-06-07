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
	"database/sql"
	"encoding/json"
	"time"

	"github.com/davidM20/micro-service-backend-go.git/internal/db"
	"github.com/davidM20/micro-service-backend-go.git/internal/models"
	"github.com/davidM20/micro-service-backend-go.git/internal/websocket/services"
	"github.com/davidM20/micro-service-backend-go.git/internal/websocket/wsmodels"
	"github.com/davidM20/micro-service-backend-go.git/pkg/customws"
	"github.com/davidM20/micro-service-backend-go.git/pkg/customws/types"
	"github.com/davidM20/micro-service-backend-go.git/pkg/logger"
)

type RequestData[T any] struct {
	Data T `json:"data"`
}

// Payloads para la deserialización de datos del cliente
type EducationPayload struct {
	Id                  int64  `json:"id"`
	Institution         string `json:"institution"`
	Degree              string `json:"degree"`
	Campus              string `json:"campus"`
	GraduationDate      string `json:"graduationDate,omitempty"`
	IsCurrentlyStudying bool   `json:"isCurrentlyStudying"`
}

type WorkExperiencePayload struct {
	Id           int64  `json:"id"`
	Company      string `json:"company"`
	Position     string `json:"position"`
	StartDate    string `json:"startDate,omitempty"`
	EndDate      string `json:"endDate,omitempty"`
	Description  string `json:"description"`
	IsCurrentJob bool   `json:"isCurrentJob"`
}

type ProjectPayload struct {
	Id              int64  `json:"id"`
	Title           string `json:"title"`
	Role            string `json:"role"`
	Description     string `json:"description"`
	Company         string `json:"company"`
	Document        string `json:"document"`
	ProjectStatus   string `json:"projectStatus"`
	StartDate       string `json:"startDate,omitempty"`
	ExpectedEndDate string `json:"expectedEndDate,omitempty"`
	IsOngoing       bool   `json:"isOngoing"`
}

type SkillPayload struct {
	Id    int64  `json:"id"`
	Skill string `json:"skill"`
	Level string `json:"level"`
}

type LanguagePayload struct {
	Id       int64  `json:"id"`
	Language string `json:"language"`
	Level    string `json:"level"`
}

type CertificationPayload struct {
	Id            int64  `json:"id"`
	Certification string `json:"certification"`
	Institution   string `json:"institution"`
	DateObtained  string `json:"dateObtained,omitempty"`
}

// --- Handlers ---

// HandleSetSkill maneja la solicitud para establecer una habilidad
func HandleSetSkill(conn *customws.Connection[wsmodels.WsUserData], msg types.ClientToServerMessage) error {
	logger.Infof("CV_HANDLER", "Estableciendo habilidad para UserID %d. PID: %s", conn.ID, msg.PID)

	var requestData RequestData[SkillPayload]
	payloadBytes, _ := json.Marshal(msg.Payload)
	if err := json.Unmarshal(payloadBytes, &requestData); err != nil {
		logger.Warnf("CV_HANDLER", "Error al decodificar payload de habilidad: %v", err)
		conn.SendErrorNotification(msg.PID, 400, "Payload de habilidad inválido.")
		return nil
	}
	skillPayload := requestData.Data

	if skillPayload.Skill == "" || skillPayload.Level == "" {
		logger.Warnf("CV_HANDLER", "Validación fallida para set_skill: campos vacíos. UserID: %d", conn.ID)
		conn.SendErrorNotification(msg.PID, 400, "Los campos 'Skill' y 'Level' no pueden estar vacíos.")
		return nil
	}

	skillModel := models.Skills{
		Id:       skillPayload.Id,
		PersonId: conn.ID,
		Skill:    skillPayload.Skill,
		Level:    skillPayload.Level,
	}

	dbConn := db.GetDB()
	cvService := services.NewCVService(dbConn)

	if err := cvService.SetSkill(&skillModel); err != nil {
		logger.Errorf("CV_HANDLER", "Error al establecer habilidad: %v", err)
		conn.SendErrorNotification(msg.PID, 500, "Error interno al establecer habilidad.")
		return nil
	}

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

	var requestData RequestData[LanguagePayload]
	payloadBytes, _ := json.Marshal(msg.Payload)
	if err := json.Unmarshal(payloadBytes, &requestData); err != nil {
		logger.Warnf("CV_HANDLER", "Error al decodificar payload de idioma: %v", err)
		conn.SendErrorNotification(msg.PID, 400, "Payload de idioma inválido.")
		return nil
	}
	languagePayload := requestData.Data

	if languagePayload.Language == "" || languagePayload.Level == "" {
		logger.Warnf("CV_HANDLER", "Validación fallida para set_language: campos vacíos. UserID: %d", conn.ID)
		conn.SendErrorNotification(msg.PID, 400, "Los campos 'Language' y 'Level' no pueden estar vacíos.")
		return nil
	}

	languageModel := models.Languages{
		Id:       languagePayload.Id,
		PersonId: conn.ID,
		Language: languagePayload.Language,
		Level:    languagePayload.Level,
	}

	dbConn := db.GetDB()
	cvService := services.NewCVService(dbConn)

	if err := cvService.SetLanguage(&languageModel); err != nil {
		logger.Errorf("CV_HANDLER", "Error al establecer idioma: %v", err)
		conn.SendErrorNotification(msg.PID, 500, "Error interno al establecer idioma.")
		return nil
	}

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

	var requestData RequestData[WorkExperiencePayload]
	payloadBytes, _ := json.Marshal(msg.Payload)
	if err := json.Unmarshal(payloadBytes, &requestData); err != nil {
		logger.Warnf("CV_HANDLER", "Error al decodificar payload de experiencia laboral: %v", err)
		conn.SendErrorNotification(msg.PID, 400, "Payload de experiencia laboral inválido.")
		return nil
	}
	experiencePayload := requestData.Data

	if experiencePayload.Company == "" || experiencePayload.Position == "" {
		logger.Warnf("CV_HANDLER", "Validación fallida para set_work_experience: campos vacíos. UserID: %d", conn.ID)
		conn.SendErrorNotification(msg.PID, 400, "Los campos 'Company' y 'Position' no pueden estar vacíos.")
		return nil
	}

	// Convertir payload a modelo de BD
	startDate, _ := time.Parse(time.RFC3339, experiencePayload.StartDate)
	endDate, _ := time.Parse(time.RFC3339, experiencePayload.EndDate)

	experienceModel := models.WorkExperience{
		Id:           experiencePayload.Id,
		PersonId:     conn.ID,
		Company:      experiencePayload.Company,
		Position:     experiencePayload.Position,
		Description:  sql.NullString{String: experiencePayload.Description, Valid: experiencePayload.Description != ""},
		StartDate:    sql.NullTime{Time: startDate, Valid: !startDate.IsZero()},
		EndDate:      sql.NullTime{Time: endDate, Valid: !endDate.IsZero() && !experiencePayload.IsCurrentJob},
		IsCurrentJob: sql.NullBool{Bool: experiencePayload.IsCurrentJob, Valid: true},
	}

	// Usar el servicio de CV
	dbConn := db.GetDB()
	cvService := services.NewCVService(dbConn)

	if err := cvService.SetWorkExperience(&experienceModel); err != nil {
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

	var requestData RequestData[CertificationPayload]
	payloadBytes, _ := json.Marshal(msg.Payload)
	if err := json.Unmarshal(payloadBytes, &requestData); err != nil {
		logger.Warnf("CV_HANDLER", "Error al decodificar payload de certificación: %v", err)
		conn.SendErrorNotification(msg.PID, 400, "Payload de certificación inválido.")
		return nil
	}
	certPayload := requestData.Data

	if certPayload.Certification == "" || certPayload.Institution == "" {
		logger.Warnf("CV_HANDLER", "Validación fallida para set_certification: campos vacíos. UserID: %d", conn.ID)
		conn.SendErrorNotification(msg.PID, 400, "Los campos 'Certification' y 'Institution' no pueden estar vacíos.")
		return nil
	}

	dateObtained, _ := time.Parse(time.RFC3339, certPayload.DateObtained)

	certificationModel := models.Certifications{
		Id:            certPayload.Id,
		PersonId:      conn.ID,
		Certification: certPayload.Certification,
		Institution:   certPayload.Institution,
		DateObtained:  sql.NullTime{Time: dateObtained, Valid: !dateObtained.IsZero()},
	}

	dbConn := db.GetDB()
	cvService := services.NewCVService(dbConn)

	if err := cvService.SetCertification(&certificationModel); err != nil {
		logger.Errorf("CV_HANDLER", "Error al establecer certificación: %v", err)
		conn.SendErrorNotification(msg.PID, 500, "Error interno al establecer certificación.")
		return nil
	}

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

	var requestData RequestData[ProjectPayload]
	payloadBytes, _ := json.Marshal(msg.Payload)
	if err := json.Unmarshal(payloadBytes, &requestData); err != nil {
		logger.Warnf("CV_HANDLER", "Error al decodificar payload de proyecto: %v", err)
		conn.SendErrorNotification(msg.PID, 400, "Payload de proyecto inválido.")
		return nil
	}
	projectPayload := requestData.Data

	if projectPayload.Title == "" || projectPayload.Role == "" {
		logger.Warnf("CV_HANDLER", "Validación fallida para set_project: campos vacíos. UserID: %d", conn.ID)
		conn.SendErrorNotification(msg.PID, 400, "Los campos 'Title' y 'Role' no pueden estar vacíos.")
		return nil
	}

	// Convertir payload a modelo de BD
	startDate, _ := time.Parse(time.RFC3339, projectPayload.StartDate)
	endDate, _ := time.Parse(time.RFC3339, projectPayload.ExpectedEndDate)

	projectModel := models.Project{
		Id:              projectPayload.Id,
		PersonID:        conn.ID,
		Title:           projectPayload.Title,
		Role:            sql.NullString{String: projectPayload.Role, Valid: projectPayload.Role != ""},
		Description:     sql.NullString{String: projectPayload.Description, Valid: projectPayload.Description != ""},
		Company:         sql.NullString{String: projectPayload.Company, Valid: projectPayload.Company != ""},
		Document:        sql.NullString{String: projectPayload.Document, Valid: projectPayload.Document != ""},
		ProjectStatus:   sql.NullString{String: projectPayload.ProjectStatus, Valid: projectPayload.ProjectStatus != ""},
		StartDate:       sql.NullTime{Time: startDate, Valid: !startDate.IsZero()},
		ExpectedEndDate: sql.NullTime{Time: endDate, Valid: !endDate.IsZero() && !projectPayload.IsOngoing},
		IsOngoing:       sql.NullBool{Bool: projectPayload.IsOngoing, Valid: true},
	}

	// Usar el servicio de CV
	dbConn := db.GetDB()
	cvService := services.NewCVService(dbConn)

	if err := cvService.SetProject(&projectModel); err != nil {
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

	var requestData RequestData[EducationPayload]
	payloadBytes, _ := json.Marshal(msg.Payload)
	if err := json.Unmarshal(payloadBytes, &requestData); err != nil {
		logger.Warnf("CV_HANDLER", "Error al decodificar payload de educación: %v", err)
		conn.SendErrorNotification(msg.PID, 400, "Payload de educación inválido.")
		return nil
	}
	educationPayload := requestData.Data

	if educationPayload.Institution == "" || educationPayload.Degree == "" {
		logger.Warnf("CV_HANDLER", "Validación fallida para set_education: campos vacíos. UserID: %d", conn.ID)
		conn.SendErrorNotification(msg.PID, 400, "Los campos 'Institution' y 'Degree' no pueden estar vacíos.")
		return nil
	}

	// Convertir payload a modelo de BD
	gradDate, _ := time.Parse(time.RFC3339, educationPayload.GraduationDate)

	educationModel := models.Education{
		Id:                  educationPayload.Id,
		PersonId:            conn.ID,
		Institution:         educationPayload.Institution,
		Degree:              educationPayload.Degree,
		Campus:              sql.NullString{String: educationPayload.Campus, Valid: educationPayload.Campus != ""},
		GraduationDate:      sql.NullTime{Time: gradDate, Valid: !gradDate.IsZero() && !educationPayload.IsCurrentlyStudying},
		IsCurrentlyStudying: sql.NullBool{Bool: educationPayload.IsCurrentlyStudying, Valid: true},
	}

	// Usar el servicio de CV
	dbConn := db.GetDB()
	cvService := services.NewCVService(dbConn)

	if err := cvService.SetEducation(&educationModel); err != nil {
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
