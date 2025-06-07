package queries

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
