-- Índices para la tabla PasswordReset

-- Índice compuesto para validar códigos no usados y vigentes
CREATE INDEX idx_reset_valid_code ON PasswordReset(Code, Used, ExpiresAt);

-- Índice para limpiar códigos expirados (mantenimiento)
CREATE INDEX idx_reset_expired ON PasswordReset(ExpiresAt); 