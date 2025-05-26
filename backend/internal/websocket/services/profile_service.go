package services

import (
	"database/sql"
	"fmt"

	// Necesario para convertir sql.NullTime a string
	"github.com/davidM20/micro-service-backend-go.git/internal/db/queries"
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

	userData, err := queries.GetUserFullProfileData(profileDB, userID)
	if err != nil {
		logger.Errorf("SERVICE_PROFILE", "Error obteniendo datos base del perfil para UserID %d: %v", userID, err)
		return nil, err
	}

	educationItemsDB, err := queries.GetEducationItemsForUser(profileDB, userID)
	if err != nil {
		logger.Warnf("SERVICE_PROFILE", "Error obteniendo items de educación para UserID %d: %v", userID, err)
		// No devolver error fatal, el perfil puede mostrarse sin esta sección
	}

	workExperienceItemsDB, err := queries.GetWorkExperienceItemsForUser(profileDB, userID)
	if err != nil {
		logger.Warnf("SERVICE_PROFILE", "Error obteniendo items de experiencia laboral para UserID %d: %v", userID, err)
	}

	certificationItemsDB, err := queries.GetCertificationItemsForUser(profileDB, userID)
	if err != nil {
		logger.Warnf("SERVICE_PROFILE", "Error obteniendo certificaciones para UserID %d: %v", userID, err)
	}

	skillItemsDB, err := queries.GetSkillItemsForUser(profileDB, userID)
	if err != nil {
		logger.Warnf("SERVICE_PROFILE", "Error obteniendo skills para UserID %d: %v", userID, err)
	}

	languageItemsDB, err := queries.GetLanguageItemsForUser(profileDB, userID)
	if err != nil {
		logger.Warnf("SERVICE_PROFILE", "Error obteniendo idiomas para UserID %d: %v", userID, err)
	}

	projectItemsDB, err := queries.GetProjectItemsForUser(profileDB, userID)
	if err != nil {
		logger.Warnf("SERVICE_PROFILE", "Error obteniendo proyectos para UserID %d: %v", userID, err)
	}

	// Convertir DB models a DTO wsmodels
	profileDto := &wsmodels.ProfileData{
		ID:                 userData.Id,
		FirstName:          userData.FirstName,
		LastName:           userData.LastName,
		UserName:           userData.UserName,
		Email:              userData.Email,
		Phone:              userData.Phone.String,
		Sex:                userData.Sex.String,
		DocId:              userData.DocId.String,
		NationalityId:      int(userData.NationalityId.Int32), // Asumiendo que NationalityId en User es sql.NullInt32
		NationalityName:    userData.NationalityName,          // Viene del JOIN en GetUserFullProfileData
		Birthdate:          formatNullTimeToString(userData.Birthdate, "2006-01-02"),
		Picture:            userData.Picture.String,
		DegreeName:         userData.DegreeName,         // Viene del JOIN
		UniversityName:     userData.UniversityName,     // Viene del JOIN
		RoleID:             userData.RoleId,             // RoleId en User es int
		RoleName:           userData.RoleName,           // Viene del JOIN
		StatusAuthorizedId: userData.StatusAuthorizedId, // StatusAuthorizedId en User es int
		Summary:            userData.Summary.String,
		Address:            userData.Address.String,
		Github:             userData.Github.String,
		Linkedin:           userData.Linkedin.String,
		CreatedAt:          userData.CreateAt,
		UpdatedAt:          userData.UpdateAt,
		Curriculum: wsmodels.CurriculumVitae{
			Education:      make([]wsmodels.EducationItem, 0),
			Experience:     make([]wsmodels.WorkExperienceItem, 0),
			Certifications: make([]wsmodels.CertificationItem, 0),
			Skills:         make([]wsmodels.SkillItem, 0),
			Languages:      make([]wsmodels.LanguageItem, 0),
			Projects:       make([]wsmodels.ProjectItem, 0),
		},
	}

	if userID == currentUserID {
		profileDto.IsOnline = true // El usuario que solicita su propio perfil siempre se considera "online" en este contexto
	} else if manager != nil {
		profileDto.IsOnline = manager.IsUserOnline(userID)
	}

	// Mapear Education
	for _, dbItem := range educationItemsDB {
		profileDto.Curriculum.Education = append(profileDto.Curriculum.Education, wsmodels.EducationItem{
			ID:             dbItem.Id,
			Institution:    dbItem.Institution,
			Degree:         dbItem.Degree,
			Campus:         dbItem.Campus,
			GraduationDate: formatNullTimeToString(dbItem.GraduationDate, "2006-01-02"),
			CountryID:      dbItem.CountryId,
			CountryName:    dbItem.CountryName, // Usar directamente el campo del modelo
		})
	}

	// Mapear WorkExperience
	for _, dbItem := range workExperienceItemsDB {
		profileDto.Curriculum.Experience = append(profileDto.Curriculum.Experience, wsmodels.WorkExperienceItem{
			ID:          dbItem.Id,
			Company:     dbItem.Company,
			Position:    dbItem.Position,
			StartDate:   formatNullTimeToString(dbItem.StartDate, "2006-01-02"),
			EndDate:     formatNullTimeToString(dbItem.EndDate, "2006-01-02"),
			Description: dbItem.Description,
			CountryID:   dbItem.CountryId,
			CountryName: dbItem.CountryName, // Usar directamente el campo del modelo
		})
	}

	// Mapear Certifications
	for _, dbItem := range certificationItemsDB {
		profileDto.Curriculum.Certifications = append(profileDto.Curriculum.Certifications, wsmodels.CertificationItem{
			ID:            dbItem.Id,
			Certification: dbItem.Certification,
			Institution:   dbItem.Institution,
			DateObtained:  formatNullTimeToString(dbItem.DateObtained, "2006-01-02"),
		})
	}

	// Mapear Skills
	for _, dbItem := range skillItemsDB {
		profileDto.Curriculum.Skills = append(profileDto.Curriculum.Skills, wsmodels.SkillItem{
			ID:    dbItem.Id,
			Skill: dbItem.Skill,
			Level: dbItem.Level,
		})
	}

	// Mapear Languages
	for _, dbItem := range languageItemsDB {
		profileDto.Curriculum.Languages = append(profileDto.Curriculum.Languages, wsmodels.LanguageItem{
			ID:       dbItem.Id,
			Language: dbItem.Language,
			Level:    dbItem.Level,
		})
	}

	// Mapear Projects
	for _, dbItem := range projectItemsDB {
		profileDto.Curriculum.Projects = append(profileDto.Curriculum.Projects, wsmodels.ProjectItem{
			ID:              dbItem.Id,
			Title:           dbItem.Title,
			Role:            dbItem.Role,
			Description:     dbItem.Description,
			Company:         dbItem.Company,
			Document:        dbItem.Document,
			ProjectStatus:   dbItem.ProjectStatus,
			StartDate:       formatNullTimeToString(dbItem.StartDate, "2006-01-02"),
			ExpectedEndDate: formatNullTimeToString(dbItem.ExpectedEndDate, "2006-01-02"),
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

// TODO: Implementar funciones del servicio de perfil
// - GetUserProfileData(userID int64, currentUserID int64, manager *customws.ConnectionManager[wsmodels.WsUserData]) (*wsmodels.ProfileData, error)
// - UpdateUserProfile(userID int64, updates map[string]interface{}) error
// - AddProfileSectionItem(userID int64, section string, itemData interface{}) error
// - UpdateProfileSectionItem(userID int64, section string, itemID int64, itemData interface{}) error
// - DeleteProfileSectionItem(userID int64, section string, itemID int64) error
