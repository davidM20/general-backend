package services

import (
	"database/sql"
	"fmt"
	"time"

	// Necesario para convertir sql.NullTime a string
	"github.com/davidM20/micro-service-backend-go.git/internal/db/queries"
	"github.com/davidM20/micro-service-backend-go.git/internal/models"
	"github.com/davidM20/micro-service-backend-go.git/internal/websocket/wsmodels"
	"github.com/davidM20/micro-service-backend-go.git/pkg/customws"
	"github.com/davidM20/micro-service-backend-go.git/pkg/logger"
)

var profileDB *sql.DB

// InitializeProfileService inyecta la dependencia de la BD.
func InitializeProfileService(db *sql.DB) {
	profileDB = db
	logger.Info("SERVICE_PROFILE", "ProfileService inicializado con conexión a BD.")
}

// GetUserProfileData construye el wsmodels.ProfileData completo para un usuario.
// currentUserID es el ID del usuario que solicita el perfil (para determinar IsOnline si es el perfil de otro).
func GetUserProfileData(userID int64, currentUserID int64, manager *customws.ConnectionManager[wsmodels.WsUserData]) (*wsmodels.ProfileData, error) {
	if profileDB == nil {
		return nil, fmt.Errorf("ProfileService no inicializado")
	}

	userData, err := queries.GetUserFullProfileData(userID)
	if err != nil {
		logger.Errorf("SERVICE_PROFILE", "Error obteniendo datos base del perfil para UserID %d: %v", userID, err)
		return nil, err
	}

	userDTO := userData.ToUserDTO()

	profileDto := &wsmodels.ProfileData{
		ID:                 userDTO.Id,
		FirstName:          userDTO.FirstName,
		LastName:           userDTO.LastName,
		UserName:           userDTO.UserName,
		Email:              userDTO.Email,
		Phone:              userDTO.Phone,
		Sex:                userDTO.Sex,
		DocId:              userDTO.DocId,
		NationalityId:      userDTO.NationalityId,
		NationalityName:    safeNullString(userData.NationalityName),
		Birthdate:          userDTO.Birthdate,
		Picture:            userDTO.Picture,
		DegreeName:         safeNullString(userData.DegreeName),
		UniversityName:     safeNullString(userData.UniversityName),
		RoleID:             userDTO.RoleId,
		RoleName:           safeNullString(userData.RoleName),
		StatusAuthorizedId: userDTO.StatusAuthorizedId,
		Summary:            userDTO.Summary,
		Address:            userDTO.Address,
		Github:             userDTO.Github,
		Linkedin:           userDTO.Linkedin,
		CreatedAt:          time.Time{},
		UpdatedAt:          time.Time{},
		Curriculum: wsmodels.CurriculumVitae{
			// Inicializar slices vacíos para evitar `null` en JSON
			Education:      []wsmodels.EducationItem{},
			Experience:     []wsmodels.WorkExperienceItem{},
			Certifications: []wsmodels.CertificationItem{},
			Skills:         []wsmodels.SkillItem{},
			Languages:      []wsmodels.LanguageItem{},
			Projects:       []wsmodels.ProjectItem{},
		},
	}

	if userID == currentUserID {
		profileDto.IsOnline = true
	} else if manager != nil {
		profileDto.IsOnline = manager.IsUserOnline(userID)
	}

	// Las queries ahora devuelven wsmodels directamente.
	profileDto.Curriculum.Education, _ = queries.GetEducationForUser(userID)
	profileDto.Curriculum.Experience, _ = queries.GetWorkExperienceForUser(userID)
	profileDto.Curriculum.Certifications, _ = queries.GetCertificationsForUser(userID)
	profileDto.Curriculum.Projects, _ = queries.GetProjectsForUser(userID)

	// Skills y Languages todavía devuelven `models`, por lo que requieren mapeo.
	skillItemsDB, _ := queries.GetSkillsForUser(userID)
	for _, dbItem := range skillItemsDB {
		profileDto.Curriculum.Skills = append(profileDto.Curriculum.Skills, wsmodels.SkillItem{
			ID:    dbItem.Id,
			Skill: dbItem.Skill,
			Level: dbItem.Level,
		})
	}

	languageItemsDB, _ := queries.GetLanguagesForUser(userID)
	for _, dbItem := range languageItemsDB {
		profileDto.Curriculum.Languages = append(profileDto.Curriculum.Languages, wsmodels.LanguageItem{
			ID:       dbItem.Id,
			Language: dbItem.Language,
			Level:    dbItem.Level,
		})
	}

	return profileDto, nil
}

// formatNullTimeToString convierte sql.NullTime a una cadena con el formato especificado.
// Devuelve una cadena vacía si NullTime no es válida.
func formatNullTimeToString(nt sql.NullTime, layout string) string {
	if nt.Valid {
		return nt.Time.Format(layout)
	}
	return ""
}

// Helper function to safely get string from sql.NullString
func safeNullString(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return ""
}

// Helper function to safely get int64 from sql.NullInt64, returning 0 if not valid.
func safeNullInt64(ni sql.NullInt64) int64 {
	if ni.Valid {
		return ni.Int64
	}
	return 0 // O el valor por defecto que consideres apropiado si NULL
}

// UpdateUserProfile llama a la capa de base de datos para actualizar el perfil de un usuario.
func UpdateUserProfile(personID int64, payload models.UpdateProfilePayload) error {
	return queries.UpdateUserProfile(personID, payload)
}

// GetCompleteProfile reúne toda la información del perfil de un usuario de forma concurrente.
func GetCompleteProfile(userID int64) (*wsmodels.ProfileData, error) {
	// Reutilizamos GetUserProfileData que ya hace todo el trabajo de forma eficiente.
	return GetUserProfileData(userID, userID, nil)
}

// TODO: Implementar funciones del servicio de perfil
// - GetUserProfileData(userID int64, currentUserID int64, manager *customws.ConnectionManager[wsmodels.WsUserData]) (*wsmodels.ProfileData, error)
// - UpdateUserProfile(userID int64, updates map[string]interface{}) error
// - AddProfileSectionItem(userID int64, section string, itemData interface{}) error
// - UpdateProfileSectionItem(userID int64, section string, itemID int64, itemData interface{}) error
// - DeleteProfileSectionItem(userID int64, section string, itemID int64) error
