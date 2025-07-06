package queries

import (
	"fmt"

	"github.com/davidM20/micro-service-backend-go.git/internal/models"
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

// SearchAll realiza una búsqueda combinada de usuarios y empresas.
// Busca en los campos `UserName`, `FirstName`, `LastName` y `CompanyName`.
//
// Parámetros:
//   - currentUserID: ID del usuario actual.
//   - searchTerm: Término de búsqueda.
//   - limit: Número máximo de resultados a devolver.
//   - offset: Número de resultados a omitir.
//
// Retorna:
//   - Una lista de usuarios (`[]models.User`) que coinciden con el término de búsqueda.
//   - Un error si la consulta falla.
func SearchAll(currentUserID int64, searchTerm string, limit, offset int) ([]models.User, error) {
	query := `
	SELECT
		u.Id,
		u.FirstName,
		u.LastName,
		u.UserName,
		u.Picture,
		u.Summary,
		u.RoleId,
		u.CompanyName,
		u.Sector,
		u.Location,
		e.Institution AS UniversityName,
		e.Degree AS DegreeName,
		c.ChatId
	FROM User u
	LEFT JOIN (
		-- Subconsulta para obtener solo la educación más reciente por usuario
		SELECT PersonId, Institution, Degree
		FROM Education
		WHERE (PersonId, GraduationDate) IN (
			SELECT PersonId, MAX(GraduationDate)
			FROM Education
			GROUP BY PersonId
		)
	) e ON u.Id = e.PersonId
	LEFT JOIN Contact c ON ((c.User1Id = ? AND c.User2Id = u.Id) OR (c.User1Id = u.Id AND c.User2Id = ?)) AND c.Status = 'accepted'
	WHERE
		u.Id != ? AND
		(
			(u.RoleId IN (1, 2) AND (
				u.UserName LIKE ? OR
				u.FirstName LIKE ? OR
				u.LastName LIKE ? OR
				e.Institution LIKE ? OR
				e.Degree LIKE ?
			)) OR
			(u.RoleId = 3 AND (
				u.CompanyName LIKE ? OR
				u.Sector LIKE ?
			))
		)
	LIMIT ? OFFSET ?;
`

	likeTerm := "%" + searchTerm + "%"
	rows, err := DB.Query(query, currentUserID, currentUserID, currentUserID, likeTerm, likeTerm, likeTerm, likeTerm, likeTerm, likeTerm, likeTerm, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("error al ejecutar la consulta de búsqueda 'all': %w", err)
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User

		err := rows.Scan(
			&user.Id,
			&user.FirstName,
			&user.LastName,
			&user.UserName,
			&user.Picture,
			&user.Summary,
			&user.RoleId,
			&user.CompanyName,
			&user.Sector,
			&user.Location,
			&user.UniversityName,
			&user.DegreeName,
			&user.ChatId,
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
