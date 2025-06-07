package queries

import (
	"database/sql"
	"fmt"

	"github.com/davidM20/micro-service-backend-go.git/internal/models"
	"github.com/davidM20/micro-service-backend-go.git/internal/websocket/wsmodels"
)

/*
Package queries proporciona un lugar centralizado para toda la lógica de acceso a la base de datos.
Este archivo contiene funciones para interactuar con las tablas de la base de datos.

NORMAS Y DIRECTRICES PARA ESTE ARCHIVO:

1. CONEXIÓN A LA BASE DE DATOS:
  - La variable global `DB *sql.DB` se inicializa en el arranque.
  - NO pasar el puntero de conexión a la base de datos como argumento a las funciones.
  - Todas las funciones de consulta dentro de este paquete deben usar la variable global `DB` directamente.

2. REUTILIZACIÓN Y RESPONSABILIDAD DEL CÓDIGO:
  - Antes de añadir una nueva función, comprueba si una existente puede ser reutilizada o generalizada.
  - Cada función debe tener una única responsabilidad, claramente definida (p. ej., obtener datos de usuario, insertar un mensaje).
  - Mantén las funciones concisas y enfocadas.

3. DOCUMENTACIÓN:
  - Documenta cada nueva función y tipo.
  - Los comentarios deben explicar el propósito de la función, sus parámetros y lo que devuelve.
  - Explica cualquier lógica compleja o comportamiento no obvio.

4. MANEJO DE ERRORES:
  - Comprueba siempre los errores devueltos por `DB.Query`, `DB.QueryRow`, `DB.Exec` y `rows.Scan`.
  - Utiliza `fmt.Errorf("contexto: %w", err)` para envolver los errores, proporcionando contexto sin perder el error original.
  - Maneja `sql.ErrNoRows` específicamente cuando se espera que una consulta a veces no devuelva resultados (p. ej., `GetUserBy...`).

5. CONVENCIONES DE NOMENCLATURA:
  - Sigue las convenciones de nomenclatura idiomáticas de Go (p. ej., `CamelCase` para identificadores exportados).
  - Usa nombres descriptivos para las funciones (p. ej., `GetUserBySessionToken`, `CreateMessage`).

6. CONSTANTES:
  - Para campos de estado o IDs de tipo (p. ej., estado del mensaje), define constantes en la parte superior del archivo.
  - Usa estas constantes en lugar de números mágicos para mejorar la legibilidad y el mantenimiento.

7. MANEJO DE COLUMNAS ANULABLES:
  - Usa `sql.NullString`, `sql.NullInt64`, `sql.NullTime`, etc., para columnas de la base de datos que pueden ser NULL.
  - Comprueba siempre el campo `Valid` antes de acceder al valor de un tipo anulable.

8. SEGURIDAD:
  - Para prevenir la inyección de SQL, SIEMPRE usa consultas parametrizadas con `?` como marcadores de posición.
  - NUNCA construyas consultas concatenando cadenas con entradas proporcionadas por el usuario.

9. AÑADIR NUEVAS CONSULTAS:
  - Agrupa las funciones relacionadas (p. ej., todas las consultas relacionadas con el usuario, todas las relacionadas con los mensajes).
  - Considera las implicaciones de rendimiento. Usa `JOIN`s con criterio y añade cláusulas `LIMIT` donde sea aplicable.
  - Asegúrate de que tu consulta devuelva solo las columnas necesarias.
*/

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
		INSERT INTO WorkExperience (Id, PersonId, Company, Position, StartDate, EndDate, Description, CountryId, IsCurrentJob)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
		Company = VALUES(Company),
		Position = VALUES(Position),
		StartDate = VALUES(StartDate),
		EndDate = VALUES(EndDate),
		Description = VALUES(Description),
		CountryId = VALUES(CountryId),
		IsCurrentJob = VALUES(IsCurrentJob)
	`

	_, err := db.Exec(query,
		experience.Id,
		experience.PersonId,
		experience.Company,
		experience.Position,
		experience.StartDate,
		experience.EndDate,
		experience.Description,
		experience.CountryId,
		experience.IsCurrentJob,
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
		INSERT INTO Project (Id, PersonID, Title, Role, Description, Company, Document, ProjectStatus, StartDate, ExpectedEndDate, IsOngoing)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
		Title = VALUES(Title),
		Role = VALUES(Role),
		Description = VALUES(Description),
		Company = VALUES(Company),
		Document = VALUES(Document),
		ProjectStatus = VALUES(ProjectStatus),
		StartDate = VALUES(StartDate),
		ExpectedEndDate = VALUES(ExpectedEndDate),
		IsOngoing = VALUES(IsOngoing)
	`

	_, err := db.Exec(query,
		project.Id,
		project.PersonID,
		project.Title,
		project.Role,
		project.Description,
		project.Company,
		project.Document,
		project.ProjectStatus,
		project.StartDate,
		project.ExpectedEndDate,
		project.IsOngoing,
	)
	if err != nil {
		return fmt.Errorf("error al establecer proyecto: %w", err)
	}
	return nil
}

// SetEducation inserta o actualiza la educación de una persona.
func SetEducation(db *sql.DB, education *models.Education) error {
	query := `
        INSERT INTO Education (Id, PersonId, Institution, Degree, Campus, GraduationDate, IsCurrentlyStudying)
        VALUES (?, ?, ?, ?, ?, ?, ?)
        ON DUPLICATE KEY UPDATE
        Institution = VALUES(Institution),
        Degree = VALUES(Degree),
        Campus = VALUES(Campus),
        GraduationDate = VALUES(GraduationDate),
        IsCurrentlyStudying = VALUES(IsCurrentlyStudying)
    `
	_, err := db.Exec(query,
		education.Id,
		education.PersonId,
		education.Institution,
		education.Degree,
		education.Campus,
		education.GraduationDate,
		education.IsCurrentlyStudying,
	)
	if err != nil {
		return fmt.Errorf("error al insertar/actualizar educación: %w", err)
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
		SELECT w.Id, w.PersonId, w.Company, w.Position, w.StartDate, w.EndDate, w.Description, w.CountryId, n.CountryName, w.IsCurrentJob
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
		var countryName sql.NullString
		if err := experienceRows.Scan(
			&exp.Id,
			&exp.PersonId,
			&exp.Company,
			&exp.Position,
			&exp.StartDate,
			&exp.EndDate,
			&exp.Description,
			&exp.CountryId,
			&countryName,
			&exp.IsCurrentJob,
		); err != nil {
			return nil, fmt.Errorf("error al escanear experiencia laboral: %w", err)
		}
		expItem := wsmodels.WorkExperienceItem{
			ID:           exp.Id,
			Company:      exp.Company,
			Position:     exp.Position,
			StartDate:    formatNullTimeToString(exp.StartDate, "2006-01-02"),
			EndDate:      formatNullTimeToString(exp.EndDate, "2006-01-02"),
			Description:  exp.Description.String,
			CountryID:    safeNullInt64(exp.CountryId),
			CountryName:  countryName.String,
			IsCurrentJob: exp.IsCurrentJob.Bool,
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
	projectsQuery := `
		SELECT Id, PersonID, Title, Role, Description, Company, Document, ProjectStatus, StartDate, ExpectedEndDate, IsOngoing
		FROM Project
		WHERE PersonID = ?
	`
	projectsRows, err := db.Query(projectsQuery, personId)
	if err != nil {
		return nil, fmt.Errorf("error al obtener proyectos: %w", err)
	}
	defer projectsRows.Close()

	for projectsRows.Next() {
		var project models.Project
		if err := projectsRows.Scan(
			&project.Id,
			&project.PersonID,
			&project.Title,
			&project.Role,
			&project.Description,
			&project.Company,
			&project.Document,
			&project.ProjectStatus,
			&project.StartDate,
			&project.ExpectedEndDate,
			&project.IsOngoing,
		); err != nil {
			return nil, fmt.Errorf("error al escanear proyecto: %w", err)
		}
		projectItem := wsmodels.ProjectItem{
			ID:            project.Id,
			Title:         project.Title,
			Role:          project.Role.String,
			Description:   project.Description.String,
			Company:       project.Company.String,
			Document:      project.Document.String,
			ProjectStatus: project.ProjectStatus.String,
			IsOngoing:     project.IsOngoing.Bool,
		}
		if project.StartDate.Valid {
			projectItem.StartDate = project.StartDate.Time.Format("2006-01-02")
		}
		if project.ExpectedEndDate.Valid {
			projectItem.ExpectedEndDate = project.ExpectedEndDate.Time.Format("2006-01-02")
		}

		cv.Projects = append(cv.Projects, projectItem)
	}

	// Obtener educación
	educationQuery := `
		SELECT e.Id, e.PersonId, e.Institution, e.Degree, e.Campus, e.GraduationDate, e.CountryId, n.CountryName, e.IsCurrentlyStudying
		FROM Education e
		LEFT JOIN Nationality n ON e.CountryId = n.Id
		WHERE e.PersonId = ?
	`
	educationRows, err := db.Query(educationQuery, personId)
	if err != nil {
		return nil, fmt.Errorf("error al obtener educación: %w", err)
	}
	defer educationRows.Close()

	for educationRows.Next() {
		var edu models.Education
		var countryName sql.NullString
		if err := educationRows.Scan(
			&edu.Id,
			&edu.PersonId,
			&edu.Institution,
			&edu.Degree,
			&edu.Campus,
			&edu.GraduationDate,
			&edu.CountryId,
			&countryName,
			&edu.IsCurrentlyStudying,
		); err != nil {
			return nil, fmt.Errorf("error al escanear educación: %w", err)
		}
		eduItem := wsmodels.EducationItem{
			ID:                  edu.Id,
			Institution:         edu.Institution,
			Degree:              edu.Degree,
			Campus:              edu.Campus.String,
			GraduationDate:      formatNullTimeToString(edu.GraduationDate, "2006-01-02"),
			CountryID:           safeNullInt64(edu.CountryId),
			CountryName:         countryName.String,
			IsCurrentlyStudying: edu.IsCurrentlyStudying.Bool,
		}
		cv.Education = append(cv.Education, eduItem)
	}

	return cv, nil
}
