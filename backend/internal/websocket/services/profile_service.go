package services

import (
	"context"
	"database/sql"
	"fmt"

	// Necesario para convertir sql.NullTime a string
	"github.com/davidM20/micro-service-backend-go.git/internal/db/queries"
	"github.com/davidM20/micro-service-backend-go.git/internal/models"
	"github.com/davidM20/micro-service-backend-go.git/internal/websocket/wsmodels"
	"github.com/davidM20/micro-service-backend-go.git/pkg/customws"
	"github.com/davidM20/micro-service-backend-go.git/pkg/logger"
	"golang.org/x/sync/errgroup"
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

	g, _ := errgroup.WithContext(context.Background())
	var profileData wsmodels.ProfileData

	// 1. Obtener datos base del perfil
	g.Go(func() error {
		userData, err := queries.GetUserFullProfileData(userID)
		if err != nil {
			logger.Errorf("SERVICE_PROFILE", "Error obteniendo datos base para UserID %d: %v", userID, err)
			return err
		}
		profileData.ID = userData.Id
		profileData.FirstName = safeNullString(userData.FirstName)
		profileData.LastName = safeNullString(userData.LastName)
		profileData.UserName = userData.UserName
		profileData.Email = userData.Email
		profileData.Phone = safeNullString(userData.Phone)
		profileData.Sex = safeNullString(userData.Sex)
		profileData.DocId = safeNullString(userData.DocId)
		if userData.NationalityId.Valid {
			profileData.NationalityId = int(userData.NationalityId.Int32)
		}
		profileData.NationalityName = safeNullString(userData.NationalityName)
		if userData.Birthdate.Valid {
			profileData.Birthdate = userData.Birthdate.Time.Format("2006-01-02")
		}
		profileData.Picture = safeNullString(userData.Picture)
		profileData.DegreeName = safeNullString(userData.DegreeName)
		profileData.UniversityName = safeNullString(userData.UniversityName)
		profileData.RoleID = userData.RoleId
		profileData.RoleName = safeNullString(userData.RoleName)
		profileData.StatusAuthorizedId = userData.StatusAuthorizedId
		profileData.Summary = safeNullString(userData.Summary)
		profileData.Address = safeNullString(userData.Address)
		profileData.Github = safeNullString(userData.Github)
		profileData.Linkedin = safeNullString(userData.Linkedin)
		// profileData.CreatedAt = userData.CreatedAt // Campo no disponible en models.User
		// profileData.UpdatedAt = userData.UpdatedAt // Campo no disponible en models.User
		return nil
	})

	// 2. Obtener datos del currículum concurrentemente
	g.Go(func() error {
		items, err := queries.GetEducationForUser(userID)
		if err != nil {
			logger.Warnf("SERVICE_PROFILE", "Error en CV (Education) para UserID %d: %v", userID, err)
			return nil
		}
		profileData.Curriculum.Education = items
		return nil
	})
	g.Go(func() error {
		items, err := queries.GetWorkExperienceForUser(userID)
		if err != nil {
			logger.Warnf("SERVICE_PROFILE", "Error en CV (Experience) para UserID %d: %v", userID, err)
			return nil
		}
		profileData.Curriculum.Experience = items
		return nil
	})
	g.Go(func() error {
		items, err := queries.GetCertificationsForUser(userID)
		if err != nil {
			logger.Warnf("SERVICE_PROFILE", "Error en CV (Certs) para UserID %d: %v", userID, err)
			return nil
		}
		profileData.Curriculum.Certifications = items
		return nil
	})
	g.Go(func() error {
		items, err := queries.GetProjectsForUser(userID)
		if err != nil {
			logger.Warnf("SERVICE_PROFILE", "Error en CV (Projects) para UserID %d: %v", userID, err)
			return nil
		}
		profileData.Curriculum.Projects = items
		return nil
	})
	g.Go(func() error {
		items, err := queries.GetSkillsForUser(userID)
		if err != nil {
			logger.Warnf("SERVICE_PROFILE", "Error en CV (Skills) para UserID %d: %v", userID, err)
			return nil
		}
		for _, dbItem := range items {
			profileData.Curriculum.Skills = append(profileData.Curriculum.Skills, wsmodels.SkillItem{
				ID:    dbItem.Id,
				Skill: dbItem.Skill,
				Level: dbItem.Level,
			})
		}
		return nil
	})
	g.Go(func() error {
		items, err := queries.GetLanguagesForUser(userID)
		if err != nil {
			logger.Warnf("SERVICE_PROFILE", "Error en CV (Langs) para UserID %d: %v", userID, err)
			return nil
		}
		for _, dbItem := range items {
			profileData.Curriculum.Languages = append(profileData.Curriculum.Languages, wsmodels.LanguageItem{
				ID:       dbItem.Id,
				Language: dbItem.Language,
				Level:    dbItem.Level,
			})
		}
		return nil
	})

	// 3. Obtener estado de conexión
	if manager != nil {
		profileData.IsOnline = manager.IsUserOnline(userID)
	}

	// 4. Obtener estadísticas de reputación
	g.Go(func() error {
		stats, err := queries.GetReputationStatsByUserID(userID)
		if err != nil {
			logger.Warnf("SERVICE_PROFILE", "Error obteniendo stats de reputación para UserID %d: %v", userID, err)
			return nil // No es un error fatal
		}
		profileData.Reputation = stats
		return nil
	})

	// 5. Obtener lista de reseñas
	g.Go(func() error {
		reviewsDB, err := queries.GetReputationReviewsByUserID(userID)
		if err != nil {
			logger.Warnf("SERVICE_PROFILE", "Error obteniendo reseñas para UserID %d: %v", userID, err)
			return nil // No es un error fatal
		}

		reviewsWS := make([]wsmodels.ReputationReviewItem, 0, len(reviewsDB))
		for _, r := range reviewsDB {
			reviewsWS = append(reviewsWS, wsmodels.ReputationReviewItem{
				Rating:              safeNullFloat64(r.Rating),
				Comment:             safeNullString(r.Comment),
				ReviewerCompanyName: safeNullString(r.ReviewerCompanyName),
				ReviewerPicture:     safeNullString(r.ReviewerPicture),
				Id:                  r.Id,
			})
		}
		profileData.Reviews = reviewsWS
		return nil
	})

	if err := g.Wait(); err != nil {
		// Loguear el primer error que ocurrió en el grupo.
		logger.Errorf("SERVICE_PROFILE", "Error obteniendo datos del perfil para UserID %d en errgroup: %v", userID, err)
		return nil, err
	}

	// Asegurarse de que los slices no sean nulos para evitar `null` en JSON
	if profileData.Curriculum.Education == nil {
		profileData.Curriculum.Education = []wsmodels.EducationItem{}
	}
	if profileData.Curriculum.Experience == nil {
		profileData.Curriculum.Experience = []wsmodels.WorkExperienceItem{}
	}
	if profileData.Curriculum.Certifications == nil {
		profileData.Curriculum.Certifications = []wsmodels.CertificationItem{}
	}
	if profileData.Curriculum.Projects == nil {
		profileData.Curriculum.Projects = []wsmodels.ProjectItem{}
	}
	if profileData.Curriculum.Skills == nil {
		profileData.Curriculum.Skills = []wsmodels.SkillItem{}
	}
	if profileData.Curriculum.Languages == nil {
		profileData.Curriculum.Languages = []wsmodels.LanguageItem{}
	}
	if profileData.Reviews == nil {
		profileData.Reviews = []wsmodels.ReputationReviewItem{}
	}

	return &profileData, nil
}

// formatNullTimeToString convierte sql.NullTime a una cadena con el formato especificado.
// Devuelve una cadena vacía si NullTime no es válida.
func formatNullTimeToString(nt sql.NullTime, layout string) string {
	if nt.Valid {
		return nt.Time.Format(layout)
	}
	return ""
}

// safeNullString convierte sql.NullString a string, devolviendo "" si es nulo.
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

func safeNullFloat64(nf sql.NullFloat64) float64 {
	if nf.Valid {
		return nf.Float64
	}
	return 0.0
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
