-- Índices para la tabla User

-- Índice para búsquedas por username (muy común en logins/consultas)
CREATE INDEX idx_user_username ON User(UserName);

-- Índice para búsquedas por nombre y apellido (búsquedas de personas)
CREATE INDEX idx_user_name ON User(FirstName, LastName);

-- Índice para filtrar usuarios por rol y estado (muy común en consultas de administración)
CREATE INDEX idx_user_role_status ON User(RoleId, StatusAuthorizedId);

-- Índice para búsquedas académicas (filtrar por universidad/carrera)
CREATE INDEX idx_user_academic ON User(UniversityId, DegreeId);

-- Índice para ordenamiento por fecha de creación (para listados recientes)
CREATE INDEX idx_user_created ON User(CreatedAt);

-- Índice para búsquedas de empresas por sector
CREATE INDEX idx_company_sector ON User(Sector, RoleId);

-- Índice para búsquedas por ubicación
CREATE INDEX idx_user_location ON User(Location);