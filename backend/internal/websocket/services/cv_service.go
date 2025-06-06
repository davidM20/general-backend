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

// GetCV obtiene todo el CV de un usuario
func (s *CVService) GetCV(personId int64) (*wsmodels.CurriculumVitae, error) {
	return queries.GetCV(s.db, personId)
}
