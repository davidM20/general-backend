package services

import (
	"database/sql"

	"github.com/davidM20/micro-service-backend-go.git/internal/db/queries"
	"github.com/davidM20/micro-service-backend-go.git/internal/models"
	"github.com/davidM20/micro-service-backend-go.git/internal/websocket/wsmodels"
)

// CVService maneja la lógica de negocio relacionada con el CV
type CVService struct {
	db *sql.DB
}

// NewCVService crea una nueva instancia de CVService
func NewCVService(db *sql.DB) *CVService {
	return &CVService{db: db}
}

// SetSkill establece una habilidad en el CV
func (s *CVService) SetSkill(skill *models.Skills) error {
	return queries.SetSkill(s.db, skill)
}

// SetLanguage establece un idioma en el CV
func (s *CVService) SetLanguage(language *models.Languages) error {
	return queries.SetLanguage(s.db, language)
}

// SetWorkExperience establece una experiencia laboral en el CV
func (s *CVService) SetWorkExperience(experience *models.WorkExperience) error {
	return queries.SetWorkExperience(s.db, experience)
}

// SetCertification establece una certificación en el CV
func (s *CVService) SetCertification(certification *models.Certifications) error {
	return queries.SetCertification(s.db, certification)
}

// SetProject establece un proyecto en el CV
func (s *CVService) SetProject(project *models.Project) error {
	return queries.SetProject(s.db, project)
}

// SetEducation establece la educación de un usuario.
func (s *CVService) SetEducation(education *models.Education) error {
	return queries.SetEducation(s.db, education)
}

// GetCV obtiene todo el CV de un usuario y lo mapea a wsmodels
func (s *CVService) GetCV(personID int64) (*wsmodels.CurriculumVitae, error) {
	// Obtener todos los items de la base de datos (que usan models con sql.Null*)
	educationItemsDB, _ := queries.GetEducationItemsForUser(personID)
	workExperienceItemsDB, _ := queries.GetWorkExperienceItemsForUser(personID)
	certificationItemsDB, _ := queries.GetCertificationItemsForUser(personID)
	skillItemsDB, _ := queries.GetSkillItemsForUser(personID)
	languageItemsDB, _ := queries.GetLanguageItemsForUser(personID)
	projectItemsDB, _ := queries.GetProjectItemsForUser(personID)

	// Crear el objeto de respuesta del CV (wsmodels)
	cv := &wsmodels.CurriculumVitae{
		Education:      make([]wsmodels.EducationItem, 0),
		Experience:     make([]wsmodels.WorkExperienceItem, 0),
		Certifications: make([]wsmodels.CertificationItem, 0),
		Skills:         make([]wsmodels.SkillItem, 0),
		Languages:      make([]wsmodels.LanguageItem, 0),
		Projects:       make([]wsmodels.ProjectItem, 0),
	}

	// Mapear Education
	for _, dbItem := range educationItemsDB {
		cv.Education = append(cv.Education, wsmodels.EducationItem{
			ID:                  dbItem.Id,
			Institution:         dbItem.Institution,
			Degree:              dbItem.Degree,
			Campus:              dbItem.Campus.String,
			GraduationDate:      formatNullTimeToString(dbItem.GraduationDate, "2006-01-02"),
			CountryID:           safeNullInt64(dbItem.CountryId),
			CountryName:         dbItem.CountryName.String,
			IsCurrentlyStudying: dbItem.IsCurrentlyStudying.Bool,
		})
	}

	// Mapear WorkExperience
	for _, dbItem := range workExperienceItemsDB {
		cv.Experience = append(cv.Experience, wsmodels.WorkExperienceItem{
			ID:           dbItem.Id,
			Company:      dbItem.Company,
			Position:     dbItem.Position,
			StartDate:    formatNullTimeToString(dbItem.StartDate, "2006-01-02"),
			EndDate:      formatNullTimeToString(dbItem.EndDate, "2006-01-02"),
			Description:  dbItem.Description.String,
			CountryID:    safeNullInt64(dbItem.CountryId),
			CountryName:  dbItem.CountryName.String,
			IsCurrentJob: dbItem.IsCurrentJob.Bool,
		})
	}

	// Mapear Certifications
	for _, dbItem := range certificationItemsDB {
		cv.Certifications = append(cv.Certifications, wsmodels.CertificationItem{
			ID:            dbItem.Id,
			Certification: dbItem.Certification,
			Institution:   dbItem.Institution,
			DateObtained:  formatNullTimeToString(dbItem.DateObtained, "2006-01-02"),
		})
	}

	// Mapear Skills
	for _, dbItem := range skillItemsDB {
		cv.Skills = append(cv.Skills, wsmodels.SkillItem{
			ID:    dbItem.Id,
			Skill: dbItem.Skill,
			Level: dbItem.Level,
		})
	}

	// Mapear Languages
	for _, dbItem := range languageItemsDB {
		cv.Languages = append(cv.Languages, wsmodels.LanguageItem{
			ID:       dbItem.Id,
			Language: dbItem.Language,
			Level:    dbItem.Level,
		})
	}

	// Mapear Projects
	for _, dbItem := range projectItemsDB {
		cv.Projects = append(cv.Projects, wsmodels.ProjectItem{
			ID:              dbItem.Id,
			Title:           dbItem.Title,
			Role:            dbItem.Role.String,
			Description:     dbItem.Description.String,
			Company:         dbItem.Company.String,
			Document:        dbItem.Document.String,
			ProjectStatus:   dbItem.ProjectStatus.String,
			StartDate:       formatNullTimeToString(dbItem.StartDate, "2006-01-02"),
			ExpectedEndDate: formatNullTimeToString(dbItem.ExpectedEndDate, "2006-01-02"),
			IsOngoing:       dbItem.IsOngoing.Bool,
		})
	}

	return cv, nil
}
