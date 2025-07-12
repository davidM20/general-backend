package queries

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/davidM20/micro-service-backend-go.git/internal/models"
)

// GetCompanyNameByID recupera el nombre de la empresa de un usuario por su ID.
// Devuelve el nombre de la empresa o un error si no se encuentra o el usuario no es una empresa.
func GetCompanyNameByID(userID int64) (string, error) {
	var companyName sql.NullString
	// Asumimos que el rol de empresa es 3.
	query := "SELECT CompanyName FROM User WHERE Id = ? AND RoleId = 3"

	err := DB.QueryRow(query, userID).Scan(&companyName)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("no se encontró una empresa con ID %d o el usuario no es una empresa", userID)
		}
		return "", fmt.Errorf("error al obtener el nombre de la empresa para el ID %d: %w", userID, err)
	}

	if !companyName.Valid || companyName.String == "" {
		return "", fmt.Errorf("la empresa con ID %d no tiene un nombre asignado", userID)
	}

	return companyName.String, nil
}

// GetUserNameByID recupera el nombre y apellido de un usuario por su ID.
func GetUserNameByID(userID int64) (string, string, error) {
	var firstName, lastName sql.NullString
	query := "SELECT FirstName, LastName FROM User WHERE Id = ?"
	err := DB.QueryRow(query, userID).Scan(&firstName, &lastName)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", "", fmt.Errorf("usuario con ID %d no encontrado", userID)
		}
		return "", "", fmt.Errorf("error al obtener el nombre del usuario para el ID %d: %w", userID, err)
	}

	// Devuelve el nombre y apellido, que pueden ser strings vacíos si son NULL en la BD.
	return firstName.String, lastName.String, nil
}

// BuildUpdateUserQuery construye dinámicamente una consulta SQL de actualización para la tabla User.
// Recibe un ID de usuario y un payload con los campos a actualizar.
// Devuelve la consulta SQL, una lista de argumentos y un error si ocurre.
func BuildUpdateUserQuery(userID int64, payload models.UpdateProfilePayload) (string, []interface{}, error) {
	var updates []string
	var args []interface{}

	val := reflect.ValueOf(payload)
	typ := val.Type()

	// Mapa de campos del struct a columnas de la BD
	fieldToColumn := map[string]string{
		"FirstName":      "FirstName",
		"LastName":       "LastName",
		"UserName":       "UserName",
		"Phone":          "Phone",
		"Sex":            "Sex",
		"Birthdate":      "Birthdate",
		"NationalityID":  "NationalityId",
		"Summary":        "Summary",
		"Address":        "Address",
		"Github":         "Github",
		"Linkedin":       "Linkedin",
		"CompanyName":    "CompanyName",
		"Picture":        "Picture",
		"Email":          "Email",
		"ContactEmail":   "ContactEmail",
		"Twitter":        "Twitter",
		"Facebook":       "Facebook",
		"DocId":          "DocId",
		"DegreeId":       "DegreeId",
		"UniversityId":   "UniversityId",
		"Sector":         "Sector",
		"Location":       "Location",
		"FoundationYear": "FoundationYear",
		"EmployeeCount":  "EmployeeCount",
	}

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldName := typ.Field(i).Name

		// Solo procesar campos que no son nil (es decir, que fueron proporcionados en el JSON)
		if !field.IsNil() {
			dbColumn, ok := fieldToColumn[fieldName]
			if !ok {
				continue // Ignorar campos no mapeados
			}

			updates = append(updates, fmt.Sprintf("%s = ?", dbColumn))

			// Manejar el tipo de dato específico
			switch fieldName {
			case "Birthdate":
				// Convertir string a time.Time
				dateStr := field.Elem().Interface().(string)
				t, err := time.Parse("2006-01-02", dateStr)
				if err != nil {
					return "", nil, fmt.Errorf("invalid date format for Birthdate: %w", err)
				}
				args = append(args, t)
			default:
				args = append(args, field.Elem().Interface())
			}
		}
	}

	if len(updates) == 0 {
		return "", nil, fmt.Errorf("no fields to update")
	}

	// Siempre actualizar el campo UpdatedAt
	updates = append(updates, "UpdatedAt = ?")
	args = append(args, time.Now())

	query := fmt.Sprintf("UPDATE User SET %s WHERE Id = ?", strings.Join(updates, ", "))
	args = append(args, userID)

	return query, args, nil
}
