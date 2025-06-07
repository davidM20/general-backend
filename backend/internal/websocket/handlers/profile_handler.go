/*
PROFILE HANDLER: REGLAS Y ARQUITECTURA

Este archivo gestiona todas las acciones relacionadas con los perfiles de usuario y empresa
que se inician desde el cliente a través de WebSocket.

ARQUITECTURA DE FLUJO DE DATOS:
El flujo de una solicitud sigue estrictamente el patrón:
Handler -> Service -> Query

1. HANDLER (este archivo - `profile_handler.go`):
  - Su única responsabilidad es recibir el mensaje del WebSocket, validar la estructura básica
    del payload y llamar al servicio correspondiente.
  - NO debe contener lógica de negocio.
  - Se encarga de la comunicación directa con el cliente: deserializa las solicitudes,
    llama a los servicios y serializa las respuestas (o errores).
  - Usa `conn.SendMessage()` para respuestas exitosas y `conn.SendErrorNotification()` para errores.

2. SERVICE (`services/profile_service.go` o `services/company_service.go`):
  - Contiene TODA la lógica de negocio.
  - Orquesta las llamadas a las consultas (`queries`) para obtener los datos necesarios.
  - Puede realizar múltiples llamadas a la base de datos y combinar los resultados.
  - Realiza cálculos o transformaciones de datos (ej. calcular estadísticas).
  - Utiliza `errgroup` para ejecutar consultas concurrentes y mejorar el rendimiento.

3. QUERY (`queries/profile_queries.go` o `queries/company_queries.go`):
  - Contiene únicamente la lógica de acceso a la base de datos.
  - Cada función corresponde a una consulta SQL específica.
  - NO debe contener lógica de negocio, solo ejecutar la consulta y escanear los resultados
    en los modelos (`models`).

CÓMO AÑADIR UNA NUEVA ACCIÓN DE PERFIL (EJ: "follow_company"):
 1. Definir la estructura del mensaje en el frontend y la documentación.
 2. Handler: Añadir una nueva función `HandleFollowCompany` o un nuevo `case` en el router.
    Esta función valida el payload y llama al servicio.
 3. Service: Crear una función `FollowCompany(followerID, companyID int64)` en el servicio
    apropiado. Esta función implementará la lógica (ej. verificar que no se siga ya,
    llamar a la query para insertar el registro en la tabla de seguidores).
 4. Query: Crear una función `InsertFollower(followerID, companyID int64)` que ejecute
    el `INSERT INTO ...` en la base de datos.

Esta separación garantiza un código limpio, mantenible y escalable.
*/
package handlers

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/davidM20/micro-service-backend-go.git/internal/db/queries"
	"github.com/davidM20/micro-service-backend-go.git/internal/models"
	"github.com/davidM20/micro-service-backend-go.git/internal/websocket/services"
	"github.com/davidM20/micro-service-backend-go.git/internal/websocket/wsmodels"
	"github.com/davidM20/micro-service-backend-go.git/pkg/customws"
	"github.com/davidM20/micro-service-backend-go.git/pkg/customws/types"
	"github.com/davidM20/micro-service-backend-go.git/pkg/logger"
)

// HandleGetProfile maneja la solicitud para obtener el perfil del propio usuario.
func HandleGetProfile(conn *customws.Connection[wsmodels.WsUserData], msg types.ClientToServerMessage) error {
	logger.Infof("PROFILE_HANDLER", "Usuario %d solicitó su propio perfil. PID: %s", conn.ID, msg.PID)

	profileData, err := services.GetUserProfileData(conn.ID, conn.ID, conn.Manager())
	if err != nil {
		logger.Errorf("PROFILE_HANDLER", "Error obteniendo perfil para user %d: %v", conn.ID, err)
		conn.SendErrorNotification(msg.PID, 500, "Error al obtener tu perfil: "+err.Error())
		return err
	}

	responseMsg := types.ServerToClientMessage{
		PID:     msg.PID,
		Type:    "my_profile_data", // Tipo de mensaje para el perfil propio
		Payload: profileData,
	}
	if msg.PID == "" {
		responseMsg.PID = conn.Manager().Callbacks().GeneratePID()
	}

	if err := conn.SendMessage(responseMsg); err != nil {
		logger.Errorf("PROFILE_HANDLER", "Error enviando datos de perfil a user %d: %v", conn.ID, err)
		return err
	}

	logger.Successf("PROFILE_HANDLER", "Datos de perfil enviados a user %d. PID respuesta: %s", conn.ID, responseMsg.PID)
	return nil
}

// HandleUpdateProfile maneja la solicitud para actualizar datos del perfil del usuario.
func HandleUpdateProfile(conn *customws.Connection[wsmodels.WsUserData], msg types.ClientToServerMessage) error {
	logger.Infof("PROFILE_HANDLER", "Solicitud para actualizar perfil de UserID %d. PID: %s", conn.ID, msg.PID)

	var payload models.UpdateProfilePayload
	payloadBytes, err := json.Marshal(msg.Payload)
	if err != nil {
		conn.SendErrorNotification(msg.PID, 400, "Error decodificando payload (marshal): "+err.Error())
		return fmt.Errorf("error marshalling UpdateProfile payload: %w", err)
	}
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		conn.SendErrorNotification(msg.PID, 400, "Error decodificando payload (unmarshal): "+err.Error())
		return fmt.Errorf("error unmarshalling UpdateProfile payload: %w", err)
	}

	if err := services.UpdateUserProfile(conn.ID, payload); err != nil {
		logger.Errorf("PROFILE_HANDLER", "Error actualizando perfil para UserID %d: %v", conn.ID, err)
		conn.SendErrorNotification(msg.PID, 500, "Error interno al actualizar el perfil.")
		return err
	}

	// Notificar éxito
	conn.SendErrorNotification(msg.PID, 200, "Perfil actualizado con éxito.")
	logger.Successf("PROFILE_HANDLER", "Perfil actualizado con éxito para UserID %d.", conn.ID)

	// Opcional: Enviar el perfil actualizado de vuelta
	// go HandleGetProfile(conn, types.ClientToServerMessage{PID: ""}) // Sin PID para no generar ACK

	return nil
}

// HandleViewProfile maneja la solicitud para ver el perfil de otro usuario.
func HandleViewProfile(conn *customws.Connection[wsmodels.WsUserData], msg types.ClientToServerMessage) error {
	logger.Infof("PROFILE_HANDLER", "Usuario %d solicitó ver un perfil. PID: %s", conn.ID, msg.PID)

	type ViewProfilePayload struct {
		UserID      *int64  `json:"userId,omitempty"`
		RIF         *string `json:"rif,omitempty"`
		CompanyName *string `json:"companyName,omitempty"`
	}
	var payload ViewProfilePayload

	if msg.Payload == nil {
		conn.SendErrorNotification(msg.PID, 400, "Payload es requerido para ver un perfil.")
		return errors.New("payload vacío para ViewProfile")
	}

	payloadBytes, err := json.Marshal(msg.Payload)
	if err != nil {
		conn.SendErrorNotification(msg.PID, 400, "Error decodificando payload (marshal): "+err.Error())
		return fmt.Errorf("error marshalling ViewProfile payload: %w", err)
	}
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		conn.SendErrorNotification(msg.PID, 400, "Error decodificando payload (unmarshal): "+err.Error())
		return fmt.Errorf("error unmarshalling ViewProfile payload: %w", err)
	}

	var targetUserID int64

	// Determinar el ID del usuario a buscar
	if payload.UserID != nil {
		targetUserID = *payload.UserID
	} else if payload.RIF != nil {
		targetUserID, err = queries.GetUserIDByRIF(*payload.RIF)
		if err != nil {
			logger.Warnf("PROFILE_HANDLER", "No se encontró empresa con RIF '%s': %v", *payload.RIF, err)
			conn.SendErrorNotification(msg.PID, 404, "No se encontró una empresa con el RIF proporcionado.")
			return err
		}
	} else if payload.CompanyName != nil {
		targetUserID, err = queries.GetUserIDByCompanyName(*payload.CompanyName)
		if err != nil {
			logger.Warnf("PROFILE_HANDLER", "No se encontró empresa con nombre '%s': %v", *payload.CompanyName, err)
			conn.SendErrorNotification(msg.PID, 404, "No se encontró una empresa con el nombre proporcionado.")
			return err
		}
	} else {
		conn.SendErrorNotification(msg.PID, 400, "Debe proporcionar 'userId', 'rif' o 'companyName' para ver un perfil.")
		return errors.New("identificador de perfil no proporcionado en ViewProfile")
	}

	if targetUserID <= 0 {
		conn.SendErrorNotification(msg.PID, 400, "Identificador de perfil inválido.")
		return errors.New("identificador de perfil inválido en ViewProfile")
	}

	// Obtener el rol del usuario solicitado para determinar qué perfil cargar
	targetRoleID, err := queries.GetUserRoleByID(targetUserID)
	if err != nil {
		logger.Errorf("PROFILE_HANDLER", "Error obteniendo rol para TargetUserID %d: %v", targetUserID, err)
		conn.SendErrorNotification(msg.PID, 404, "El usuario solicitado no existe.")
		return err
	}

	var responsePayload interface{}

	if targetRoleID == 3 {
		// Lógica para perfil de empresa
		profile, err := services.GetCompleteCompanyProfile(targetUserID)
		if err != nil {
			logger.Errorf("PROFILE_HANDLER", "Error obteniendo perfil de empresa para TargetUserID %d: %v", targetUserID, err)
			conn.SendErrorNotification(msg.PID, 500, "Error al obtener el perfil de la empresa.")
			return err
		}
		responsePayload = profile
	} else {
		// Lógica para perfil de estudiante/egresado
		profile, err := services.GetCompleteProfile(targetUserID)
		if err != nil {
			logger.Errorf("PROFILE_HANDLER", "Error obteniendo perfil de usuario para TargetUserID %d: %v", targetUserID, err)
			conn.SendErrorNotification(msg.PID, 500, "Error al obtener el perfil del usuario.")
			return err
		}
		responsePayload = profile
	}

	responseMsg := types.ServerToClientMessage{
		PID:     msg.PID,
		Type:    "view_profile_success", // Tipo de mensaje genérico para la vista de perfil
		Payload: responsePayload,
	}

	if err := conn.SendMessage(responseMsg); err != nil {
		logger.Errorf("PROFILE_HANDLER", "Error enviando datos de perfil de %d a user %d: %v", targetUserID, conn.ID, err)
		return err
	}

	logger.Successf("PROFILE_HANDLER", "Datos de perfil de %d enviados a user %d. PID respuesta: %s", targetUserID, conn.ID, msg.PID)
	return nil
}

// HandleGetUserProfile maneja la solicitud para obtener el perfil de otro usuario.
func HandleGetUserProfile(conn *customws.Connection[wsmodels.WsUserData], msg types.ClientToServerMessage) error {
	logger.Infof("HANDLER_PROFILE", "Usuario %d solicitó perfil de otro usuario. PID: %s", conn.ID, msg.PID)

	type GetUserProfilePayload struct {
		UserID int64 `json:"userId"`
	}
	var payload GetUserProfilePayload

	if msg.Payload == nil {
		conn.SendErrorNotification(msg.PID, 400, "Payload es requerido para obtener perfil de usuario.")
		return errors.New("payload vacío para GetUserProfile")
	}

	payloadBytes, err := json.Marshal(msg.Payload)
	if err != nil {
		conn.SendErrorNotification(msg.PID, 400, "Error decodificando payload (marshal): "+err.Error())
		return fmt.Errorf("error marshalling GetUserProfile payload: %w", err)
	}
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		conn.SendErrorNotification(msg.PID, 400, "Error decodificando payload (unmarshal): "+err.Error())
		return fmt.Errorf("error unmarshalling GetUserProfile payload: %w", err)
	}

	if payload.UserID <= 0 {
		conn.SendErrorNotification(msg.PID, 400, "userId inválido en payload.")
		return errors.New("userId inválido en GetUserProfile")
	}

	profileData, err := services.GetUserProfileData(payload.UserID, conn.ID, conn.Manager())
	if err != nil {
		logger.Errorf("HANDLER_PROFILE", "Error obteniendo perfil para TargetUserID %d (solicitado por UserID %d): %v", payload.UserID, conn.ID, err)
		conn.SendErrorNotification(msg.PID, 500, fmt.Sprintf("Error al obtener perfil de usuario %d: %s", payload.UserID, err.Error()))
		return err
	}

	responseMsg := types.ServerToClientMessage{
		PID:     msg.PID, // Responder al mismo PID
		Type:    types.MessageTypeUserProfileData,
		Payload: profileData,
	}
	if msg.PID == "" {
		responseMsg.PID = conn.Manager().Callbacks().GeneratePID()
	}

	if err := conn.SendMessage(responseMsg); err != nil {
		logger.Errorf("HANDLER_PROFILE", "Error enviando datos de perfil de %d a user %d: %v", payload.UserID, conn.ID, err)
		return err
	}

	logger.Successf("HANDLER_PROFILE", "Datos de perfil de %d enviados a user %d. PID respuesta: %s", payload.UserID, conn.ID, responseMsg.PID)
	return nil
}

func HandleMyProfileView(conn *customws.Connection[wsmodels.WsUserData], msg types.ClientToServerMessage) error {
	logger.Infof("PROFILE_HANDLER", "Usuario %d solicitó ver su propio perfil. PID: %s", conn.ID, msg.PID)

	// Verificar el rol del usuario para determinar qué perfil cargar
	if conn.UserData.RoleId == 3 {
		// Lógica para perfil de empresa
		profile, err := services.GetCompleteCompanyProfile(conn.ID)
		if err != nil {
			logger.Errorf("PROFILE_HANDLER", "Error obteniendo el perfil de empresa para el usuario %d: %v", conn.ID, err)
			errorMsg := fmt.Sprintf("No se pudo cargar tu perfil de empresa: %v", err)
			conn.SendErrorNotification(msg.PID, 500, errorMsg)
			return err
		}
		response := types.ServerToClientMessage{
			PID:     msg.PID,
			Type:    "view_my_profile_success",
			Payload: profile,
		}
		if err := conn.SendMessage(response); err != nil {
			logger.Errorf("PROFILE_HANDLER", "Fallo al enviar el perfil de empresa al usuario %d: %v", conn.ID, err)
			return err
		}
		logger.Successf("PROFILE_HANDLER", "Perfil de empresa enviado exitosamente al usuario %d. PID: %s", conn.ID, msg.PID)

	} else {
		// Lógica para perfil de estudiante/egresado
		profile, err := services.GetCompleteProfile(conn.ID)
		if err != nil {
			logger.Errorf("PROFILE_HANDLER", "Error obteniendo el perfil completo para el usuario %d: %v", conn.ID, err)
			errorMsg := fmt.Sprintf("No se pudo cargar tu perfil: %v", err)
			conn.SendErrorNotification(msg.PID, 500, errorMsg)
			return err
		}
		response := types.ServerToClientMessage{
			PID:     msg.PID,
			Type:    "view_my_profile_success",
			Payload: profile,
		}
		if err := conn.SendMessage(response); err != nil {
			logger.Errorf("PROFILE_HANDLER", "Fallo al enviar el perfil completo al usuario %d: %v", conn.ID, err)
			return err
		}
		logger.Successf("PROFILE_HANDLER", "Perfil completo enviado exitosamente al usuario %d. PID: %s", conn.ID, msg.PID)
	}

	return nil
}

// TODO: Implementar manejadores para perfiles
// - HandleUpdateProfileSection (para añadir/editar/eliminar items de Educación, Experiencia, Skills etc.)
