package queries

import (
	"bytes"
	"database/sql"
	"fmt"

	"github.com/davidM20/micro-service-backend-go.git/internal/models"
	"github.com/davidM20/micro-service-backend-go.git/pkg/logger"
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

// UpdateEnterpriseProfile actualiza los campos de una empresa en la tabla User.
// La consulta se construye dinámicamente para actualizar solo los campos proporcionados.
func UpdateEnterpriseProfile(db *sql.DB, userID int64, data *models.EnterpriseProfileUpdate) error {
	var query bytes.Buffer
	query.WriteString("UPDATE User SET ")

	args := make([]interface{}, 0)
	fieldCount := 0

	// Función auxiliar para añadir campos a la consulta
	addField := func(fieldName string, value interface{}) {
		if fieldCount > 0 {
			query.WriteString(", ")
		}
		query.WriteString(fmt.Sprintf("%s = ?", fieldName))
		args = append(args, value)
		fieldCount++
	}

	if data.CompanyName != nil {
		addField("CompanyName", *data.CompanyName)
	}
	if data.ContactEmail != nil {
		addField("ContactEmail", *data.ContactEmail)
	}
	if data.Twitter != nil {
		addField("Twitter", *data.Twitter)
	}
	if data.Facebook != nil {
		addField("Facebook", *data.Facebook)
	}
	if data.Phone != nil {
		addField("Phone", *data.Phone)
	}
	if data.Picture != nil {
		addField("Picture", *data.Picture)
	}
	if data.Summary != nil {
		addField("Summary", *data.Summary)
	}
	if data.Address != nil {
		addField("Address", *data.Address)
	}
	if data.Github != nil {
		addField("Github", *data.Github)
	}
	if data.Linkedin != nil {
		addField("Linkedin", *data.Linkedin)
	}
	if data.Sector != nil {
		addField("Sector", *data.Sector)
	}
	if data.Location != nil {
		addField("Location", *data.Location)
	}
	if data.FoundationYear != nil {
		addField("FoundationYear", *data.FoundationYear)
	}
	if data.EmployeeCount != nil {
		addField("EmployeeCount", *data.EmployeeCount)
	}

	// Si no se proporcionó ningún campo para actualizar, no hacemos nada.
	if fieldCount == 0 {
		return nil // O un error específico si se prefiere
	}

	query.WriteString(" WHERE Id = ?")
	args = append(args, userID)

	stmt, err := db.Prepare(query.String())
	if err != nil {
		return fmt.Errorf("error preparing update statement: %w", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(args...)
	if err != nil {
		return fmt.Errorf("error executing update for user %d: %w", userID, err)
	}

	return nil
}

// Funciones para futura implementación:

// GetEnterpriseById obtiene los datos de una empresa por su ID
// func GetEnterpriseById(db *sql.DB, id int64) (*models.Enterprise, error) {
//    // Implementación futura
// }

// ListEnterprises lista empresas con filtros opcionales
// func ListEnterprises(db *sql.DB, filters map[string]interface{}, limit, offset int) ([]models.Enterprise, error) {
//    // Implementación futura
// }
