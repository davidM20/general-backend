package queries

import (
	"database/sql"
	"fmt"

	"github.com/davidM20/micro-service-backend-go.git/internal/models"
)

// GetCompanyProfile recupera la información de perfil de una empresa por su ID.
func GetCompanyProfile(userID int64) (*models.CompanyProfile, error) {
	query := `
        SELECT
            Id, CompanyName, Email, ContactEmail, RIF, Sector, Location, Address,
            FoundationYear, EmployeeCount, Summary, Phone, Github, Linkedin, Twitter, Facebook,
            Picture, RoleId, StatusAuthorizedId, CreatedAt, UpdatedAt
        FROM User WHERE Id = ? AND RoleId = 3
    `
	var profile models.CompanyProfile
	var contactEmail, rif, sector, location, address, summary, phone, github, linkedin, twitter, facebook, picture sql.NullString
	var foundationYear, employeeCount sql.NullInt32

	err := DB.QueryRow(query, userID).Scan(
		&profile.Id, &profile.CompanyName, &profile.Email, &contactEmail, &rif, &sector, &location, &address,
		&foundationYear, &employeeCount, &summary, &phone, &github, &linkedin, &twitter, &facebook,
		&picture, &profile.RoleId, &profile.StatusAuthorizedId, &profile.CreatedAt, &profile.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("empresa con ID %d no encontrada o no es una empresa", userID)
		}
		return nil, fmt.Errorf("error al obtener perfil de empresa: %w", err)
	}

	// Solo asignar valores cuando son válidos (no NULL)
	if contactEmail.Valid {
		profile.ContactEmail = contactEmail.String
	}
	if rif.Valid {
		profile.RIF = rif.String
	}
	if sector.Valid {
		profile.Sector = sector.String
	}
	if location.Valid {
		profile.Location = location.String
	}
	if address.Valid {
		profile.Address = address.String
	}
	if summary.Valid {
		profile.Summary = summary.String
	}
	if phone.Valid {
		profile.Phone = phone.String
	}
	if github.Valid {
		profile.Github = github.String
	}
	if linkedin.Valid {
		profile.Linkedin = linkedin.String
	}
	if twitter.Valid {
		profile.Twitter = twitter.String
	}
	if facebook.Valid {
		profile.Facebook = facebook.String
	}
	if picture.Valid {
		profile.Picture = picture.String
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
}

// GetEventsForCompany recupera la lista de eventos para una empresa.
func GetEventsForCompany(companyID int64) ([]models.CompanyEvent, error) {
	query := `
        SELECT Id, Title, Description, EventDate, Location, ImageURL, CreatedAt, UpdatedAt
        FROM CommunityEvent WHERE CreatedByUserID = ? ORDER BY EventDate DESC
    `
	rows, err := DB.Query(query, companyID)
	if err != nil {
		return nil, fmt.Errorf("error al obtener eventos de la empresa: %w", err)
	}
	defer rows.Close()

	var events []models.CompanyEvent
	for rows.Next() {
		var event models.CompanyEvent
		var imageUrl sql.NullString
		var eventDate sql.NullTime
		var location sql.NullString
		if err := rows.Scan(&event.ID, &event.Title, &event.Description, &eventDate, &location, &imageUrl, &event.CreatedAt, &event.UpdatedAt); err != nil {
			return nil, fmt.Errorf("error al escanear evento: %w", err)
		}
		event.EventDate = models.NullTime{NullTime: eventDate}

		if imageUrl.Valid {
			event.ImageURL = imageUrl.String
		}

		if location.Valid {
			event.Location = location.String
		}

		events = append(events, event)
	}
	return events, nil
}

// GetUserRoleByID recupera únicamente el RoleId de un usuario por su ID.
func GetUserRoleByID(userID int64) (int, error) {
	var roleId int
	query := "SELECT RoleId FROM User WHERE Id = ?"
	err := DB.QueryRow(query, userID).Scan(&roleId)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, fmt.Errorf("usuario con ID %d no encontrado", userID)
		}
		return 0, fmt.Errorf("error al obtener el rol del usuario: %w", err)
	}
	return roleId, nil
}

// GetUserIDByRIF recupera el ID de un usuario empresa por su RIF.
func GetUserIDByRIF(rif string) (int64, error) {
	var userID int64
	query := "SELECT Id FROM User WHERE RIF = ? AND RoleId = 3"
	err := DB.QueryRow(query, rif).Scan(&userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, fmt.Errorf("empresa con RIF %s no encontrada", rif)
		}
		return 0, fmt.Errorf("error al obtener empresa por RIF: %w", err)
	}
	return userID, nil
}

// GetUserIDByCompanyName recupera el ID de un usuario empresa por su nombre.
// Dado que el nombre no es único, devuelve el primero que encuentra.
func GetUserIDByCompanyName(companyName string) (int64, error) {
	var userID int64
	query := "SELECT Id FROM User WHERE CompanyName = ? AND RoleId = 3 LIMIT 1"
	err := DB.QueryRow(query, companyName).Scan(&userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, fmt.Errorf("empresa con nombre '%s' no encontrada", companyName)
		}
		return 0, fmt.Errorf("error al obtener empresa por nombre: %w", err)
	}
	return userID, nil
}
