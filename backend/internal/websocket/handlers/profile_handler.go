package handlers

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/davidM20/micro-service-backend-go.git/internal/websocket/services"
	"github.com/davidM20/micro-service-backend-go.git/internal/websocket/wsmodels"
	"github.com/davidM20/micro-service-backend-go.git/pkg/customws"
	"github.com/davidM20/micro-service-backend-go.git/pkg/customws/types"
	"github.com/davidM20/micro-service-backend-go.git/pkg/logger"
)

// HandleGetMyProfile maneja la solicitud para obtener el perfil del propio usuario.
func HandleGetMyProfile(conn *customws.Connection[wsmodels.WsUserData], msg types.ClientToServerMessage) error {
	logger.Infof("HANDLER_PROFILE", "Usuario %d solicitó su propio perfil. PID: %s", conn.ID, msg.PID)

	profileData, err := services.GetUserProfileData(conn.ID, conn.ID, conn.Manager())
	if err != nil {
		logger.Errorf("HANDLER_PROFILE", "Error obteniendo perfil para user %d: %v", conn.ID, err)
		conn.SendErrorNotification(msg.PID, 500, "Error al obtener tu perfil: "+err.Error())
		return err
	}

	responseMsg := types.ServerToClientMessage{
		PID:     msg.PID, // Responder al mismo PID si el cliente lo envió
		Type:    types.MessageTypeMyProfileData,
		Payload: profileData,
	}
	if msg.PID == "" {
		responseMsg.PID = conn.Manager().Callbacks().GeneratePID()
	}

	if err := conn.SendMessage(responseMsg); err != nil {
		logger.Errorf("HANDLER_PROFILE", "Error enviando datos de perfil a user %d: %v", conn.ID, err)
		return err
	}

	logger.Successf("HANDLER_PROFILE", "Datos de perfil enviados a user %d. PID respuesta: %s", conn.ID, responseMsg.PID)
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

// TODO: Implementar manejadores para perfiles
// - HandleUpdateMyProfile (para campos principales del User)
// - HandleUpdateProfileSection (para añadir/editar/eliminar items de Educación, Experiencia, Skills etc.)
