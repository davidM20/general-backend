package queries

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/davidM20/micro-service-backend-go.git/internal/models"
	"github.com/davidM20/micro-service-backend-go.git/internal/websocket/wsmodels"
	"github.com/davidM20/micro-service-backend-go.git/pkg/logger"
)

// UpdateUserProfile actualiza dinámicamente los campos del perfil de un usuario.
// Construye la consulta SQL basándose en los campos no nulos del payload.
func UpdateUserProfile(personID int64, payload models.UpdateProfilePayload) error {
	var setClauses []string
	var args []interface{}
	argID := 1

	if payload.FirstName != nil {
		setClauses = append(setClauses, fmt.Sprintf("FirstName = $%d", argID))
		args = append(args, *payload.FirstName)
		argID++
	}
	if payload.LastName != nil {
		setClauses = append(setClauses, fmt.Sprintf("LastName = $%d", argID))
		args = append(args, *payload.LastName)
		argID++
	}
	if payload.UserName != nil {
		setClauses = append(setClauses, fmt.Sprintf("UserName = $%d", argID))
		args = append(args, *payload.UserName)
		argID++
	}
	if payload.Phone != nil {
		setClauses = append(setClauses, fmt.Sprintf("Phone = $%d", argID))
		args = append(args, *payload.Phone)
		argID++
	}
	if payload.Sex != nil {
		setClauses = append(setClauses, fmt.Sprintf("Sex = $%d", argID))
		args = append(args, *payload.Sex)
		argID++
	}
	if payload.Birthdate != nil {
		parsedDate, err := time.Parse("2006-01-02", *payload.Birthdate)
		if err != nil {
			return fmt.Errorf("formato de fecha de nacimiento inválido: %w", err)
		}
		setClauses = append(setClauses, fmt.Sprintf("Birthdate = $%d", argID))
		args = append(args, parsedDate)
		argID++
	}
	if payload.NationalityID != nil {
		setClauses = append(setClauses, fmt.Sprintf("NationalityId = $%d", argID))
		args = append(args, *payload.NationalityID)
		argID++
	}
	if payload.Summary != nil {
		setClauses = append(setClauses, fmt.Sprintf("Summary = $%d", argID))
		args = append(args, *payload.Summary)
		argID++
	}
	if payload.Address != nil {
		setClauses = append(setClauses, fmt.Sprintf("Address = $%d", argID))
		args = append(args, *payload.Address)
		argID++
	}
	if payload.Github != nil {
		setClauses = append(setClauses, fmt.Sprintf("Github = $%d", argID))
		args = append(args, *payload.Github)
		argID++
	}
	if payload.Linkedin != nil {
		setClauses = append(setClauses, fmt.Sprintf("Linkedin = $%d", argID))
		args = append(args, *payload.Linkedin)
		argID++
	}
	if payload.CompanyName != nil {
		setClauses = append(setClauses, fmt.Sprintf("CompanyName = $%d", argID))
		args = append(args, *payload.CompanyName)
		argID++
	}
	if payload.Picture != nil {
		setClauses = append(setClauses, fmt.Sprintf("Picture = $%d", argID))
		args = append(args, *payload.Picture)
		argID++
	}
	if payload.Email != nil {
		setClauses = append(setClauses, fmt.Sprintf("Email = $%d", argID))
		args = append(args, *payload.Email)
		argID++
	}
	if payload.ContactEmail != nil {
		setClauses = append(setClauses, fmt.Sprintf("ContactEmail = $%d", argID))
		args = append(args, *payload.ContactEmail)
		argID++
	}
	if payload.Twitter != nil {
		setClauses = append(setClauses, fmt.Sprintf("Twitter = $%d", argID))
		args = append(args, *payload.Twitter)
		argID++
	}
	if payload.Facebook != nil {
		setClauses = append(setClauses, fmt.Sprintf("Facebook = $%d", argID))
		args = append(args, *payload.Facebook)
		argID++
	}
	if payload.DocId != nil {
		setClauses = append(setClauses, fmt.Sprintf("DocId = $%d", argID))
		args = append(args, *payload.DocId)
		argID++
	}
	if payload.DegreeId != nil {
		setClauses = append(setClauses, fmt.Sprintf("DegreeId = $%d", argID))
		args = append(args, *payload.DegreeId)
		argID++
	}
	if payload.UniversityId != nil {
		setClauses = append(setClauses, fmt.Sprintf("UniversityId = $%d", argID))
		args = append(args, *payload.UniversityId)
		argID++
	}
	if payload.Sector != nil {
		setClauses = append(setClauses, fmt.Sprintf("Sector = $%d", argID))
		args = append(args, *payload.Sector)
		argID++
	}
	if payload.Location != nil {
		setClauses = append(setClauses, fmt.Sprintf("Location = $%d", argID))
		args = append(args, *payload.Location)
		argID++
	}
	if payload.FoundationYear != nil {
		setClauses = append(setClauses, fmt.Sprintf("FoundationYear = $%d", argID))
		args = append(args, *payload.FoundationYear)
		argID++
	}
	if payload.EmployeeCount != nil {
		setClauses = append(setClauses, fmt.Sprintf("EmployeeCount = $%d", argID))
		args = append(args, *payload.EmployeeCount)
		argID++
	}

	if len(setClauses) == 0 {
		return errors.New("no se proporcionaron campos para actualizar")
	}

	query := fmt.Sprintf("UPDATE User SET %s WHERE id = $%d", strings.Join(setClauses, ", "), argID)
	args = append(args, personID)

	_, err := DB.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("error al ejecutar la actualización del perfil: %w", err)
	}

	return nil
}

// GetUserProfile recupera la información pública de un perfil de usuario.
func GetUserProfile(userID int64) (*models.UserProfile, error) {
	query := `
        SELECT
            Id, FirstName, LastName, UserName, Email, ContactEmail, Twitter, Facebook,
            Phone, Sex, DocId, NationalityId, Birthdate, Picture, DegreeId, UniversityId,
            RoleId, StatusAuthorizedId, Summary, Address, Github, Linkedin, RIF,
            Sector, CompanyName, Location, FoundationYear, EmployeeCount, CreatedAt, UpdatedAt
        FROM User WHERE Id = ?
    `

	result, err := MeasureQueryWithResult(func() (interface{}, error) {
		var profile models.UserProfile
		var (
			contactEmail, twitter, facebook, phone, sex, docId, summary, address, github, linkedin, rif, sector, companyName, location sql.NullString
			nationalityId                                                                                                              sql.NullInt32
			birthdate                                                                                                                  sql.NullTime
			degreeId, universityId                                                                                                     sql.NullInt64
			foundationYear, employeeCount                                                                                              sql.NullInt32
		)

		err := DB.QueryRow(query, userID).Scan(
			&profile.Id, &profile.FirstName, &profile.LastName, &profile.UserName, &profile.Email, &contactEmail, &twitter, &facebook,
			&phone, &sex, &docId, &nationalityId, &birthdate, &profile.Picture, &degreeId, &universityId,
			&profile.RoleId, &profile.StatusAuthorizedId, &summary, &address, &github, &linkedin, &rif,
			&sector, &companyName, &location, &foundationYear, &employeeCount, &profile.CreatedAt, &profile.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		// Asignar valores desde tipos Null a tipos base
		profile.ContactEmail = contactEmail.String
		profile.Twitter = twitter.String
		profile.Facebook = facebook.String
		profile.Phone = phone.String
		profile.Sex = sex.String
		profile.DocId = docId.String
		profile.Summary = summary.String
		profile.Address = address.String
		profile.Github = github.String
		profile.Linkedin = linkedin.String
		profile.RIF = rif.String
		profile.Sector = sector.String
		profile.CompanyName = companyName.String
		profile.Location = location.String

		if nationalityId.Valid {
			val := int32(nationalityId.Int32)
			profile.NationalityId = &val
		}
		if birthdate.Valid {
			profile.Birthdate = &birthdate.Time
		}
		if degreeId.Valid {
			val := degreeId.Int64
			profile.DegreeId = &val
		}
		if universityId.Valid {
			val := universityId.Int64
			profile.UniversityId = &val
		}
		if foundationYear.Valid {
			val := int(foundationYear.Int32)
			profile.FoundationYear = &val
		}
		if employeeCount.Valid {
			val := int(employeeCount.Int32)
			profile.EmployeeCount = &val
		}

		return &profile, nil
	})

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("usuario con ID %d no encontrado", userID)
		}
		return nil, fmt.Errorf("error al obtener el perfil de usuario: %w", err)
	}

	return result.(*models.UserProfile), nil
}

// GetUserProfileByID recupera un perfil de búsqueda simplificado para un usuario.
func GetUserProfileByID(db *sql.DB, userID int64) (*models.SearchResultProfile, error) {
	query := `
        SELECT
            u.Id, u.FirstName, u.LastName, u.Picture, u.Summary, u.Location, u.RoleId,
            (SELECT SUM(rr.PointsRP) FROM ReputationReview rr WHERE rr.RevieweeId = u.Id) AS TotalReputation,
            (SELECT AVG(rr.Rating) FROM ReputationReview rr WHERE rr.RevieweeId = u.Id) AS AverageRating,
            (SELECT e.Degree FROM Education e WHERE e.PersonId = u.Id ORDER BY e.GraduationDate DESC LIMIT 1) AS Career,
            (SELECT SUM(DATEDIFF(IF(we.IsCurrentJob, CURDATE(), we.EndDate), we.StartDate)) / 365.25 FROM WorkExperience we WHERE we.PersonId = u.Id) AS YearsOfExperience
        FROM User u
        WHERE u.Id = ?
    `
	profile := &models.SearchResultProfile{}
	err := db.QueryRow(query, userID).Scan(
		&profile.ID, &profile.FirstName, &profile.LastName, &profile.Picture,
		&profile.Summary, &profile.Location, &profile.RoleId, &profile.TotalReputation, &profile.AverageRating,
		&profile.Career, &profile.YearsOfExperience,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("usuario con ID %d no encontrado para el perfil de búsqueda", userID)
		}
		logger.Errorf("QUERIES", "Error escaneando SearchResultProfile para el usuario %d: %v", userID, err)
		return nil, err
	}

	return profile, nil
}

// GetEducationForUser recupera la lista de educación de un usuario.
func GetEducationForUser(userID int64) ([]wsmodels.EducationItem, error) {
	query := `
		SELECT e.Id, e.PersonId, e.Institution, e.Degree, e.Campus, e.GraduationDate, e.CountryId, n.CountryName, e.IsCurrentlyStudying
		FROM Education e
		LEFT JOIN Nationality n ON e.CountryId = n.Id
		WHERE e.PersonId = ?
	`

	result, err := MeasureQueryWithResult(func() (interface{}, error) {
		rows, err := DB.Query(query, userID)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		var educations []wsmodels.EducationItem
		for rows.Next() {
			var edu models.Education
			if err := rows.Scan(&edu.Id, &edu.PersonId, &edu.Institution, &edu.Degree, &edu.Campus, &edu.GraduationDate, &edu.CountryId, &edu.CountryName, &edu.IsCurrentlyStudying); err != nil {
				return nil, err
			}
			educations = append(educations, wsmodels.EducationItem{
				ID:                  edu.Id,
				Institution:         edu.Institution,
				Degree:              edu.Degree,
				Campus:              edu.Campus.String,
				GraduationDate:      formatNullTimeToString(edu.GraduationDate, "2006-01-02"),
				CountryID:           safeNullInt64(edu.CountryId),
				CountryName:         edu.CountryName.String,
				IsCurrentlyStudying: edu.IsCurrentlyStudying.Bool,
			})
		}
		return educations, nil
	})

	if err != nil {
		return nil, fmt.Errorf("error al obtener la educación del usuario: %w", err)
	}

	return result.([]wsmodels.EducationItem), nil
}

// GetWorkExperienceForUser recupera la lista de experiencia laboral de un usuario.
func GetWorkExperienceForUser(userID int64) ([]wsmodels.WorkExperienceItem, error) {
	query := `
		SELECT w.Id, w.PersonId, w.Company, w.Position, w.StartDate, w.EndDate, w.Description, w.CountryId, n.CountryName, w.IsCurrentJob
		FROM WorkExperience w
		LEFT JOIN Nationality n ON w.CountryId = n.Id
		WHERE w.PersonId = ?
	`

	result, err := MeasureQueryWithResult(func() (interface{}, error) {
		rows, err := DB.Query(query, userID)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		var experiences []wsmodels.WorkExperienceItem
		for rows.Next() {
			var exp models.WorkExperience
			if err := rows.Scan(&exp.Id, &exp.PersonId, &exp.Company, &exp.Position, &exp.StartDate, &exp.EndDate, &exp.Description, &exp.CountryId, &exp.CountryName, &exp.IsCurrentJob); err != nil {
				return nil, err
			}
			experiences = append(experiences, wsmodels.WorkExperienceItem{
				ID:           exp.Id,
				Company:      exp.Company,
				Position:     exp.Position,
				StartDate:    formatNullTimeToString(exp.StartDate, "2006-01-02"),
				EndDate:      formatNullTimeToString(exp.EndDate, "2006-01-02"),
				Description:  exp.Description.String,
				CountryID:    safeNullInt64(exp.CountryId),
				CountryName:  exp.CountryName.String,
				IsCurrentJob: exp.IsCurrentJob.Bool,
			})
		}
		return experiences, nil
	})

	if err != nil {
		return nil, fmt.Errorf("error al obtener la experiencia laboral del usuario: %w", err)
	}

	return result.([]wsmodels.WorkExperienceItem), nil
}

// GetCertificationsForUser recupera la lista de certificaciones de un usuario.
func GetCertificationsForUser(userID int64) ([]wsmodels.CertificationItem, error) {
	query := "SELECT Id, PersonId, Certification, Institution, DateObtained FROM Certifications WHERE PersonId = ?"

	result, err := MeasureQueryWithResult(func() (interface{}, error) {
		rows, err := DB.Query(query, userID)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		var certifications []wsmodels.CertificationItem
		for rows.Next() {
			var cert models.Certifications
			if err := rows.Scan(&cert.Id, &cert.PersonId, &cert.Certification, &cert.Institution, &cert.DateObtained); err != nil {
				return nil, err
			}
			certifications = append(certifications, wsmodels.CertificationItem{
				ID:            cert.Id,
				Certification: cert.Certification,
				Institution:   cert.Institution,
				DateObtained:  formatNullTimeToString(cert.DateObtained, "2006-01-02"),
			})
		}
		return certifications, nil
	})

	if err != nil {
		return nil, fmt.Errorf("error al obtener certificaciones: %w", err)
	}
	return result.([]wsmodels.CertificationItem), nil
}

// GetSkillsForUser recupera las habilidades de un usuario.
func GetSkillsForUser(userID int64) ([]models.Skills, error) {
	query := "SELECT Id, PersonId, Skill, Level FROM Skills WHERE PersonId = ?"

	result, err := MeasureQueryWithResult(func() (interface{}, error) {
		rows, err := DB.Query(query, userID)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		var skills []models.Skills
		for rows.Next() {
			var skill models.Skills
			if err := rows.Scan(&skill.Id, &skill.PersonId, &skill.Skill, &skill.Level); err != nil {
				return nil, err
			}
			skills = append(skills, skill)
		}
		return skills, nil
	})

	if err != nil {
		return nil, fmt.Errorf("error al obtener habilidades: %w", err)
	}

	return result.([]models.Skills), nil
}

// GetLanguagesForUser recupera los idiomas de un usuario.
func GetLanguagesForUser(userID int64) ([]models.Languages, error) {
	query := "SELECT Id, PersonId, Language, Level FROM Languages WHERE PersonId = ?"

	result, err := MeasureQueryWithResult(func() (interface{}, error) {
		rows, err := DB.Query(query, userID)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		var languages []models.Languages
		for rows.Next() {
			var lang models.Languages
			if err := rows.Scan(&lang.Id, &lang.PersonId, &lang.Language, &lang.Level); err != nil {
				return nil, err
			}
			languages = append(languages, lang)
		}
		return languages, nil
	})

	if err != nil {
		return nil, fmt.Errorf("error al obtener idiomas: %w", err)
	}
	return result.([]models.Languages), nil
}

// GetProjectsForUser recupera la lista de proyectos de un usuario.
func GetProjectsForUser(userID int64) ([]wsmodels.ProjectItem, error) {
	query := "SELECT Id, PersonID, Title, Role, Description, Company, Document, ProjectStatus, StartDate, ExpectedEndDate, IsOngoing FROM Project WHERE PersonID = ?"

	result, err := MeasureQueryWithResult(func() (interface{}, error) {
		rows, err := DB.Query(query, userID)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		var projects []wsmodels.ProjectItem
		for rows.Next() {
			var proj models.Project
			if err := rows.Scan(&proj.Id, &proj.PersonID, &proj.Title, &proj.Role, &proj.Description, &proj.Company, &proj.Document, &proj.ProjectStatus, &proj.StartDate, &proj.ExpectedEndDate, &proj.IsOngoing); err != nil {
				return nil, err
			}
			projects = append(projects, wsmodels.ProjectItem{
				ID:              proj.Id,
				Title:           proj.Title,
				Role:            proj.Role.String,
				Description:     proj.Description.String,
				Company:         proj.Company.String,
				Document:        proj.Document.String,
				ProjectStatus:   proj.ProjectStatus.String,
				StartDate:       formatNullTimeToString(proj.StartDate, "2006-01-02"),
				ExpectedEndDate: formatNullTimeToString(proj.ExpectedEndDate, "2006-01-02"),
				IsOngoing:       proj.IsOngoing.Bool,
			})
		}
		return projects, nil
	})

	if err != nil {
		return nil, fmt.Errorf("error al obtener los proyectos del usuario: %w", err)
	}

	return result.([]wsmodels.ProjectItem), nil
}

// GetDegreeByID recupera los detalles de un título universitario por su ID.
func GetDegreeByID(degreeID int64) (*models.Degree, error) {
	query := "SELECT Id, DegreeName, Descriptions, Code, UniversityId FROM Degree WHERE Id = ?"

	result, err := MeasureQueryWithResult(func() (interface{}, error) {
		var degree models.Degree
		err := DB.QueryRow(query, degreeID).Scan(&degree.Id, &degree.DegreeName, &degree.Descriptions, &degree.Code, &degree.UniversityId)
		if err != nil {
			return nil, err
		}
		return &degree, nil
	})

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No es un error si el usuario no tiene un título asociado
		}
		return nil, fmt.Errorf("error al obtener el título: %w", err)
	}

	return result.(*models.Degree), nil
}

// formatNullTimeToString convierte sql.NullTime a una cadena con el formato especificado.
// Devuelve una cadena vacía si NullTime no es válida.
func formatNullTimeToString(nt sql.NullTime, layout string) string {
	if nt.Valid {
		return nt.Time.Format(layout)
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
