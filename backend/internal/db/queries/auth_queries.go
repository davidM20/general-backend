package queries

import (
	"database/sql"
	"time"

	"github.com/davidM20/micro-service-backend-go.git/internal/models"
	"github.com/davidM20/micro-service-backend-go.git/pkg/logger"
)

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
