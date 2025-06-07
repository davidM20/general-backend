package handlers

import (
	"encoding/json"

	"github.com/davidM20/micro-service-backend-go.git/internal/websocket/wsmodels"
	"github.com/davidM20/micro-service-backend-go.git/pkg/customws"
	"github.com/davidM20/micro-service-backend-go.git/pkg/customws/types"
	"github.com/davidM20/micro-service-backend-go.git/pkg/logger"
)

// HandleGetMyProfile maneja la solicitud para obtener el perfil del usuario autenticado.
func HandleGetMyProfile(conn *customws.Connection[wsmodels.WsUserData], msg types.ClientToServerMessage) error {
	logger.Infof("PROFILE_HANDLER", "Solicitud para obtener mi perfil recibida de UserID: %d. PID: %s", conn.ID, msg.PID)

	// Aquí llamarías a un servicio que obtiene los datos del perfil del usuario
	// Por ahora, enviaremos una respuesta de demostración.
	// Ejemplo: profile, err := services.ProfileService.GetByID(conn.ID)

	responsePayload := map[string]interface{}{
		"message": "Funcionalidad para obtener mi perfil aún no implementada.",
		"userID":  conn.ID,
	}

	responseMsg := types.ServerToClientMessage{
		PID:     conn.Manager().Callbacks().GeneratePID(),
		Type:    "my_profile_data",
		Payload: responsePayload,
	}

	return conn.SendMessage(responseMsg)
}

// HandleUpdateMyProfile maneja la solicitud para actualizar el perfil del usuario.
func HandleUpdateMyProfile(conn *customws.Connection[wsmodels.WsUserData], msg types.ClientToServerMessage) error {
	logger.Infof("PROFILE_HANDLER", "Solicitud para actualizar mi perfil recibida de UserID: %d. PID: %s", conn.ID, msg.PID)

	var requestData map[string]interface{}
	payloadBytes, _ := json.Marshal(msg.Payload)
	if err := json.Unmarshal(payloadBytes, &requestData); err != nil {
		logger.Warnf("PROFILE_HANDLER", "Error al decodificar payload de actualización de perfil: %v", err)
		conn.SendErrorNotification(msg.PID, 400, "Payload de actualización inválido.")
		return nil
	}

	// Aquí llamarías a un servicio que actualiza los datos.
	// Ejemplo: err := services.ProfileService.Update(conn.ID, requestData)

	responsePayload := map[string]interface{}{
		"message":      "Perfil actualizado con éxito (simulado).",
		"dataReceived": requestData,
	}

	responseMsg := types.ServerToClientMessage{
		PID:     conn.Manager().Callbacks().GeneratePID(),
		Type:    "update_my_profile_success",
		Payload: responsePayload,
	}

	return conn.SendMessage(responseMsg)
}

// HandleGetUserProfile maneja la solicitud para obtener el perfil de otro usuario.
func HandleGetUserProfile(conn *customws.Connection[wsmodels.WsUserData], msg types.ClientToServerMessage) error {
	logger.Infof("PROFILE_HANDLER", "Solicitud para ver perfil de usuario recibida de UserID: %d. PID: %s", conn.ID, msg.PID)

	var requestData struct {
		UserID   *int    `json:"userId"`
		Username *string `json:"username"`
	}
	payloadBytes, _ := json.Marshal(msg.Payload)
	if err := json.Unmarshal(payloadBytes, &requestData); err != nil {
		logger.Warnf("PROFILE_HANDLER", "Error al decodificar payload de visualización de perfil: %v", err)
		conn.SendErrorNotification(msg.PID, 400, "Payload de visualización inválido.")
		return nil
	}

	if requestData.UserID == nil && requestData.Username == nil {
		conn.SendErrorNotification(msg.PID, 400, "Debe proporcionar 'userId' o 'username'.")
		return nil
	}

	// Aquí llamarías a un servicio que busca el perfil por ID o username.
	// Ejemplo: profile, err := services.ProfileService.GetByCriteria(requestData)

	responsePayload := map[string]interface{}{
		"message":        "Funcionalidad para obtener perfil de usuario aún no implementada.",
		"searchCriteria": requestData,
	}

	responseMsg := types.ServerToClientMessage{
		PID:     conn.Manager().Callbacks().GeneratePID(),
		Type:    "user_profile_data",
		Payload: responsePayload,
	}

	return conn.SendMessage(responseMsg)
}

// TODO: Implementar manejadores para perfiles
// - HandleUpdateMyProfile (para campos principales del User)
// - HandleUpdateProfileSection (para añadir/editar/eliminar items de Educación, Experiencia, Skills etc.)
