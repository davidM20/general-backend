package queries

import (
	"database/sql"
	"fmt"

	"github.com/davidM20/micro-service-backend-go.git/internal/models"
)

// SearchAll realiza una búsqueda combinada de usuarios y empresas.
// Busca en los campos `UserName`, `FirstName`, `LastName` y `CompanyName`.
//
// Parámetros:
//   - db: Conexión a la base de datos.
//   - searchTerm: Término de búsqueda.
//   - limit: Número máximo de resultados a devolver.
//   - offset: Número de resultados a omitir.
//
// Retorna:
//   - Una lista de usuarios (`[]models.User`) que coinciden con el término de búsqueda.
//   - Un error si la consulta falla.
func SearchAll(db *sql.DB, searchTerm string, limit, offset int) ([]models.User, error) {
	query := `
        SELECT
            u.Id,
            u.FirstName,
            u.LastName,
            CASE
                WHEN u.RoleId = 3 THEN u.CompanyName
                ELSE u.UserName
            END,
            u.Picture,
            u.Summary,
            u.RoleId
        FROM User u
        WHERE
            (u.RoleId IN (1, 2) AND (
                u.UserName LIKE ? OR
                u.FirstName LIKE ? OR
                u.LastName LIKE ?
            )) OR
            (u.RoleId = 3 AND (
                u.CompanyName LIKE ? OR
                u.Sector LIKE ?
            ))
        LIMIT ? OFFSET ?;
    `

	likeTerm := "%" + searchTerm + "%"
	rows, err := db.Query(query, likeTerm, likeTerm, likeTerm, likeTerm, likeTerm, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("error al ejecutar la consulta de búsqueda 'all': %w", err)
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User

		err := rows.Scan(
			&user.Id, &user.FirstName, &user.LastName, &user.UserName, &user.Picture, &user.Summary,
			&user.RoleId,
		)
		if err != nil {
			return nil, fmt.Errorf("error al escanear la fila de búsqueda 'all': %w", err)
		}
		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error durante la iteración de filas de búsqueda 'all': %w", err)
	}

	return users, nil
}
