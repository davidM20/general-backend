package queries

import (
	"database/sql"
	"time"

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

// CheckUserExists verifica si ya existe un usuario con el mismo email o nombre de usuario
func CheckUserExists(db *sql.DB, email, username string) (bool, error) {
	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM User WHERE Email = ? OR UserName = ?)"

	result, err := MeasureQueryWithResult(func() (interface{}, error) {
		var e bool
		err := db.QueryRow(query, email, username).Scan(&e)
		return e, err
	})

	if err != nil {
		logger.Errorf("AUTH_QUERIES", "Error checking user existence for %s: %v", email, err)
		return false, err
	}

	exists = result.(bool)
	return exists, nil
}

// RegisterNewUser registra un nuevo usuario en el sistema
func RegisterNewUser(db *sql.DB, user models.RegistrationStep1, hashedPassword string, roleId, statusId int) (int64, error) {
	query := `
        INSERT INTO User (FirstName, LastName, UserName, Password, Email, Phone, RoleId, StatusAuthorizedId)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?)
    `

	result, err := MeasureQueryWithResult(func() (interface{}, error) {
		return db.Exec(
			query,
			user.FirstName,
			user.LastName,
			user.UserName,
			hashedPassword,
			user.Email,
			user.Phone,
			roleId,
			statusId,
		)
	})

	if err != nil {
		logger.Errorf("AUTH_QUERIES", "Error inserting user %s: %v", user.Email, err)
		return 0, err
	}

	sqlResult := result.(sql.Result)
	userId, err := sqlResult.LastInsertId()
	if err != nil {
		logger.Errorf("AUTH_QUERIES", "Error getting last insert ID for %s: %v", user.Email, err)
		return 0, err
	}

	return userId, nil
}

// CheckDocIdExists verifica si ya existe un usuario con el mismo documento de identidad
func CheckDocIdExists(db *sql.DB, docId string, userId int64) (bool, error) {
	query := "SELECT EXISTS(SELECT 1 FROM User WHERE DocId = ? AND Id != ?)"

	result, err := MeasureQueryWithResult(func() (interface{}, error) {
		var exists bool
		err := db.QueryRow(query, docId, userId).Scan(&exists)
		return exists, err
	})

	if err != nil {
		logger.Errorf("AUTH_QUERIES", "Error checking DocId existence for %s: %v", docId, err)
		return false, err
	}

	return result.(bool), nil
}

// UpdateUserStep2 actualiza la información del paso 2 del registro
func UpdateUserStep2(db *sql.DB, userId int64, docId string, nationalityId int) error {
	query := "UPDATE User SET DocId = ?, NationalityId = ? WHERE Id = ?"

	err := MeasureQuery(func() error {
		_, err := db.Exec(query, docId, nationalityId, userId)
		return err
	})

	if err != nil {
		logger.Errorf("AUTH_QUERIES", "Error updating user step 2 for UserID %d: %v", userId, err)
		return err
	}

	return nil
}

// UpdateUserStep3 actualiza la información del paso 3 del registro
func UpdateUserStep3(db *sql.DB, userId int64, sex string, birthdate time.Time, roleId, statusId int) error {
	query := "UPDATE User SET Sex = ?, Birthdate = ?, RoleId = ?, StatusAuthorizedId = ? WHERE Id = ?"

	err := MeasureQuery(func() error {
		_, err := db.Exec(query, sex, birthdate, roleId, statusId, userId)
		return err
	})

	if err != nil {
		logger.Errorf("AUTH_QUERIES", "Error updating user step 3 for UserID %d: %v", userId, err)
		return err
	}

	return nil
}

// GetUserByEmail obtiene la información de un usuario por su email
func GetUserByEmail(db *sql.DB, email string) (models.User, string, error) {
	var user models.User

	query := `
        SELECT
            Id, FirstName, LastName, UserName, Password, Email, Phone, Sex, DocId,
            NationalityId, Birthdate, Picture, DegreeId, UniversityId,
            RoleId, StatusAuthorizedId, Summary, Address, Github, Linkedin
        FROM User WHERE Email = ?
    `

	result, err := MeasureQueryWithResult(func() (interface{}, error) {
		var u models.User
		var pwd string
		err := db.QueryRow(query, email).Scan(
			&u.Id, &u.FirstName, &u.LastName, &u.UserName, &pwd, &u.Email,
			&u.Phone, &u.Sex, &u.DocId, &u.NationalityId, &u.Birthdate,
			&u.Picture, &u.DegreeId, &u.UniversityId, &u.RoleId,
			&u.StatusAuthorizedId, &u.Summary, &u.Address, &u.Github, &u.Linkedin,
		)
		return struct {
			User     models.User
			Password string
		}{User: u, Password: pwd}, err
	})

	if err != nil {
		if err == sql.ErrNoRows {
			return user, "", sql.ErrNoRows
		}
		logger.Errorf("AUTH_QUERIES", "Error getting user by email %s: %v", email, err)
		return user, "", err
	}

	resultStruct := result.(struct {
		User     models.User
		Password string
	})

	return resultStruct.User, resultStruct.Password, nil
}

// RegisterUserSession registra una nueva sesión para el usuario
func RegisterUserSession(db *sql.DB, userId int64, token, ip string, roleId int) error {
	query := `
		INSERT INTO Session (UserId, Tk, Ip, RoleId, TokenId)
		VALUES (?, ?, ?, ?, ?)
	`

	err := MeasureQuery(func() error {
		_, err := db.Exec(query, userId, token, ip, roleId, 0) // TokenId = 0 por ahora
		return err
	})

	if err != nil {
		logger.Errorf("AUTH_QUERIES", "Failed inserting session for UserID %d: %v", userId, err)
		return err
	}

	return nil
}
