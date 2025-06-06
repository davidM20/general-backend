package queries

import (
	"database/sql"
	"fmt"

	"github.com/davidM20/micro-service-backend-go.git/internal/models"
	"github.com/davidM20/micro-service-backend-go.git/internal/websocket/wsmodels"
)

// SetSkill agrega o actualiza una habilidad en el CV del usuario
func SetSkill(db *sql.DB, skill *models.Skills) error {
	query := `
		INSERT INTO Skills (PersonId, Skill, Level)
		VALUES (?, ?, ?)
		ON DUPLICATE KEY UPDATE
		Skill = VALUES(Skill),
		Level = VALUES(Level)
	`

	_, err := db.Exec(query, skill.PersonId, skill.Skill, skill.Level)
	if err != nil {
		return fmt.Errorf("error al establecer habilidad: %w", err)
	}
	return nil
}

// SetLanguage agrega o actualiza un idioma en el CV del usuario
func SetLanguage(db *sql.DB, language *models.Languages) error {
	query := `
		INSERT INTO Languages (PersonId, Language, Level)
		VALUES (?, ?, ?)
		ON DUPLICATE KEY UPDATE
		Language = VALUES(Language),
		Level = VALUES(Level)
	`

	_, err := db.Exec(query, language.PersonId, language.Language, language.Level)
	if err != nil {
		return fmt.Errorf("error al establecer idioma: %w", err)
	}
	return nil
}

// SetWorkExperience agrega o actualiza una experiencia laboral en el CV del usuario
func SetWorkExperience(db *sql.DB, experience *models.WorkExperience) error {
	query := `
		INSERT INTO WorkExperience (PersonId, Company, Position, StartDate, EndDate, Description, CountryId)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
		Company = VALUES(Company),
		Position = VALUES(Position),
		StartDate = VALUES(StartDate),
		EndDate = VALUES(EndDate),
		Description = VALUES(Description),
		CountryId = VALUES(CountryId)
	`

	_, err := db.Exec(query,
		experience.PersonId,
		experience.Company,
		experience.Position,
		experience.StartDate,
		experience.EndDate,
		experience.Description,
		experience.CountryId,
	)
	if err != nil {
		return fmt.Errorf("error al establecer experiencia laboral: %w", err)
	}
	return nil
}

// SetCertification agrega o actualiza una certificación en el CV del usuario
func SetCertification(db *sql.DB, certification *models.Certifications) error {
	query := `
		INSERT INTO Certifications (PersonId, Certification, Institution, DateObtained)
		VALUES (?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
		Certification = VALUES(Certification),
		Institution = VALUES(Institution),
		DateObtained = VALUES(DateObtained)
	`

	_, err := db.Exec(query,
		certification.PersonId,
		certification.Certification,
		certification.Institution,
		certification.DateObtained,
	)
	if err != nil {
		return fmt.Errorf("error al establecer certificación: %w", err)
	}
	return nil
}

// SetProject agrega o actualiza un proyecto en el CV del usuario
func SetProject(db *sql.DB, project *models.Project) error {
	query := `
		INSERT INTO Project (PersonID, Title, Role, Description, Company, Document, ProjectStatus, StartDate, ExpectedEndDate)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
		Title = VALUES(Title),
		Role = VALUES(Role),
		Description = VALUES(Description),
		Company = VALUES(Company),
		Document = VALUES(Document),
		ProjectStatus = VALUES(ProjectStatus),
		StartDate = VALUES(StartDate),
		ExpectedEndDate = VALUES(ExpectedEndDate)
	`

	_, err := db.Exec(query,
		project.PersonID,
		project.Title,
		project.Role,
		project.Description,
		project.Company,
		project.Document,
		project.ProjectStatus,
		project.StartDate,
		project.ExpectedEndDate,
	)
	if err != nil {
		return fmt.Errorf("error al establecer proyecto: %w", err)
	}
	return nil
}

// GetCV obtiene todo el CV de un usuario
func GetCV(db *sql.DB, personId int64) (*wsmodels.CurriculumVitae, error) {
	cv := &wsmodels.CurriculumVitae{}

	// Obtener habilidades
	skillsQuery := `SELECT Id, PersonId, Skill, Level FROM Skills WHERE PersonId = ?`
	skillsRows, err := db.Query(skillsQuery, personId)
	if err != nil {
		return nil, fmt.Errorf("error al obtener habilidades: %w", err)
	}
	defer skillsRows.Close()

	for skillsRows.Next() {
		var skill models.Skills
		if err := skillsRows.Scan(&skill.Id, &skill.PersonId, &skill.Skill, &skill.Level); err != nil {
			return nil, fmt.Errorf("error al escanear habilidad: %w", err)
		}
		skillItem := wsmodels.SkillItem{
			ID:    skill.Id,
			Skill: skill.Skill,
			Level: skill.Level,
		}
		cv.Skills = append(cv.Skills, skillItem)
	}

	// Obtener idiomas
	languagesQuery := `SELECT Id, PersonId, Language, Level FROM Languages WHERE PersonId = ?`
	languagesRows, err := db.Query(languagesQuery, personId)
	if err != nil {
		return nil, fmt.Errorf("error al obtener idiomas: %w", err)
	}
	defer languagesRows.Close()

	for languagesRows.Next() {
		var language models.Languages
		if err := languagesRows.Scan(&language.Id, &language.PersonId, &language.Language, &language.Level); err != nil {
			return nil, fmt.Errorf("error al escanear idioma: %w", err)
		}
		languageItem := wsmodels.LanguageItem{
			ID:       language.Id,
			Language: language.Language,
			Level:    language.Level,
		}
		cv.Languages = append(cv.Languages, languageItem)
	}

	// Obtener experiencia laboral
	experienceQuery := `
		SELECT w.Id, w.PersonId, w.Company, w.Position, w.StartDate, w.EndDate, w.Description, w.CountryId, n.CountryName
		FROM WorkExperience w
		LEFT JOIN Nationality n ON w.CountryId = n.Id
		WHERE w.PersonId = ?
	`
	experienceRows, err := db.Query(experienceQuery, personId)
	if err != nil {
		return nil, fmt.Errorf("error al obtener experiencia laboral: %w", err)
	}
	defer experienceRows.Close()

	for experienceRows.Next() {
		var exp models.WorkExperience
		var countryName sql.NullString // Usar sql.NullString para CountryName ya que LEFT JOIN puede resultar en NULL
		if err := experienceRows.Scan(
			&exp.Id, &exp.PersonId, &exp.Company, &exp.Position,
			&exp.StartDate, &exp.EndDate, &exp.Description,
			&exp.CountryId, &countryName,
		); err != nil {
			return nil, fmt.Errorf("error al escanear experiencia laboral: %w", err)
		}

		// Mapear de models.WorkExperience a wsmodels.WorkExperienceItem
		expItem := wsmodels.WorkExperienceItem{
			ID:          exp.Id,
			Company:     exp.Company,
			Position:    exp.Position,
			Description: exp.Description,
		}

		if exp.StartDate.Valid {
			expItem.StartDate = exp.StartDate.Time.Format("2006-01-02")
		}

		if exp.EndDate.Valid {
			expItem.EndDate = exp.EndDate.Time.Format("2006-01-02")
		}

		if exp.CountryId.Valid {
			expItem.CountryID = exp.CountryId.Int64
		}

		if countryName.Valid {
			expItem.CountryName = countryName.String
		}

		cv.Experience = append(cv.Experience, expItem)
	}

	// Obtener certificaciones
	certificationsQuery := `SELECT Id, PersonId, Certification, Institution, DateObtained FROM Certifications WHERE PersonId = ?`
	certificationsRows, err := db.Query(certificationsQuery, personId)
	if err != nil {
		return nil, fmt.Errorf("error al obtener certificaciones: %w", err)
	}
	defer certificationsRows.Close()

	for certificationsRows.Next() {
		var cert models.Certifications
		if err := certificationsRows.Scan(
			&cert.Id, &cert.PersonId, &cert.Certification,
			&cert.Institution, &cert.DateObtained,
		); err != nil {
			return nil, fmt.Errorf("error al escanear certificación: %w", err)
		}

		// Mapear de models.Certifications a wsmodels.CertificationItem
		certItem := wsmodels.CertificationItem{
			ID:            cert.Id,
			Certification: cert.Certification,
			Institution:   cert.Institution,
		}

		if cert.DateObtained.Valid {
			certItem.DateObtained = cert.DateObtained.Time.Format("2006-01-02")
		}

		cv.Certifications = append(cv.Certifications, certItem)
	}

	// Obtener proyectos
	projectsQuery := `SELECT Id, PersonID, Title, Role, Description, Company, Document, ProjectStatus, StartDate, ExpectedEndDate FROM Project WHERE PersonID = ?`
	projectsRows, err := db.Query(projectsQuery, personId)
	if err != nil {
		return nil, fmt.Errorf("error al obtener proyectos: %w", err)
	}
	defer projectsRows.Close()

	for projectsRows.Next() {
		var project models.Project
		if err := projectsRows.Scan(
			&project.Id, &project.PersonID, &project.Title, &project.Role,
			&project.Description, &project.Company, &project.Document,
			&project.ProjectStatus, &project.StartDate, &project.ExpectedEndDate,
		); err != nil {
			return nil, fmt.Errorf("error al escanear proyecto: %w", err)
		}

		// Mapear de models.Project a wsmodels.ProjectItem
		projectItem := wsmodels.ProjectItem{
			ID:            project.Id,
			Title:         project.Title,
			Role:          project.Role,
			Description:   project.Description,
			Company:       project.Company,
			Document:      project.Document,
			ProjectStatus: project.ProjectStatus,
		}

		if project.StartDate.Valid {
			projectItem.StartDate = project.StartDate.Time.Format("2006-01-02")
		}

		if project.ExpectedEndDate.Valid {
			projectItem.ExpectedEndDate = project.ExpectedEndDate.Time.Format("2006-01-02")
		}

		cv.Projects = append(cv.Projects, projectItem)
	}

	return cv, nil
}
