-- Tabla para almacenar códigos de restablecimiento de contraseña
CREATE TABLE IF NOT EXISTS PasswordReset (
    Id INT PRIMARY KEY AUTO_INCREMENT,
    UserID BIGINT NOT NULL,    -- Cambiado a BIGINT para coincidir con el tipo de User.Id
    Code VARCHAR(5) NOT NULL,  -- Código numérico de 5 dígitos
    ExpiresAt DATETIME NOT NULL,
    Used BOOLEAN NOT NULL DEFAULT FALSE,
    CreatedAt DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (UserID) REFERENCES User(Id) ON DELETE CASCADE
);

-- Índice para búsquedas rápidas por código
CREATE INDEX idx_password_reset_code ON PasswordReset(Code);

-- Índice para búsquedas por usuario
CREATE INDEX idx_password_reset_user ON PasswordReset(UserID); 