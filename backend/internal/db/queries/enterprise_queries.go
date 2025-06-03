package queries

import (
	"database/sql"
	"fmt"

	"github.com/davidM20/micro-service-backend-go.git/internal/models"
	"github.com/davidM20/micro-service-backend-go.git/pkg/logger"
)

// CheckEmailExists verifica si ya existe un usuario con el email proporcionado
func CheckEmailExists(db *sql.DB, email string) (bool, error) {
	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM User WHERE Email = ?)"

	err := db.QueryRow(query, email).Scan(&exists)
	if err != nil {
		logger.Errorf("ENTERPRISE_QUERY", "Error checking email existence: %v", err)
		return false, fmt.Errorf("error verificando email: %w", err)
	}

	return exists, nil
}

// CheckRIFExists verifica si ya existe una empresa con el RIF proporcionado
func CheckRIFExists(db *sql.DB, rif string) (bool, error) {
	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM User WHERE RIF = ?)"

	err := db.QueryRow(query, rif).Scan(&exists)
	if err != nil {
		logger.Errorf("ENTERPRISE_QUERY", "Error checking RIF existence: %v", err)
		return false, fmt.Errorf("error verificando RIF: %w", err)
	}

	return exists, nil
}

// RegisterEnterprise registra una nueva empresa en la tabla User
func RegisterEnterprise(db *sql.DB, enterprise *models.EnterpriseRegistration) (int64, error) {
	// Constantes para roles y estados
	const enterpriseRoleId = 9 // Rol para empresas
	const defaultStatusId = 1  // Estado por defecto

	query := `
		INSERT INTO User (
			CompanyName, RIF, Sector, FirstName, Email, Phone, Password,
			Location, RoleId, StatusAuthorizedId
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := db.Exec(
		query,
		enterprise.CompanyName,
		enterprise.RIF,
		enterprise.Sector,
		enterprise.FirstName,
		enterprise.Email,
		enterprise.Phone,
		enterprise.Password, // Debería recibir el hash, no la contraseña en texto plano
		enterprise.Location,
		enterpriseRoleId,
		defaultStatusId,
	)

	if err != nil {
		logger.Errorf("ENTERPRISE_QUERY", "Error registering enterprise: %v", err)
		return 0, fmt.Errorf("error registrando empresa: %w", err)
	}

	userId, err := result.LastInsertId()
	if err != nil {
		logger.Errorf("ENTERPRISE_QUERY", "Error getting last insert ID: %v", err)
		return 0, fmt.Errorf("error obteniendo ID: %w", err)
	}

	return userId, nil
}

// Funciones para futura implementación:

// GetEnterpriseById obtiene los datos de una empresa por su ID
// func GetEnterpriseById(db *sql.DB, id int64) (*models.Enterprise, error) {
//    // Implementación futura
// }

// UpdateEnterpriseProfile actualiza el perfil de una empresa
// func UpdateEnterpriseProfile(db *sql.DB, enterprise *models.Enterprise) error {
//    // Implementación futura
// }

// ListEnterprises lista empresas con filtros opcionales
// func ListEnterprises(db *sql.DB, filters map[string]interface{}, limit, offset int) ([]models.Enterprise, error) {
//    // Implementación futura
// }
