package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// contextKey es un tipo usado para las claves del contexto para evitar colisiones.
type contextKey string

// ContextKeyClaims es la clave usada para almacenar los claims del JWT en el contexto.
const ContextKeyClaims = contextKey("claims")

// Claims define la estructura de los claims del JWT
type Claims struct {
	UserID int64 `json:"user_id"`
	RoleID int   `json:"role_id"`
	jwt.RegisteredClaims
}

// GenerateJWT genera un nuevo token JWT para un usuario
func GenerateJWT(userID int64, roleID int, secretKey []byte) (string, error) {
	// Definir el tiempo de expiración del token (ej. 24 horas)
	expirationTime := time.Now().Add(24 * time.Hour)

	claims := &Claims{
		UserID: userID,
		RoleID: roleID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "backend-general-api", // Puedes cambiar el issuer
		},
	}

	// Crear el token con los claims y el método de firma HS256
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Firmar el token con la clave secreta
	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

// ValidateJWT valida un token JWT y devuelve los claims si es válido
func ValidateJWT(tokenString string, secretKey []byte) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// Verificar que el método de firma sea el esperado (HS256)
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return secretKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}
