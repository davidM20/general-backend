package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// ContextKey es un tipo para usar como clave en el contexto de la petición
type ContextKey string

// UserIDKey es la clave para almacenar el UserID en el contexto
const UserIDKey ContextKey = "userID"

// Claims define la estructura de los claims del JWT
type Claims struct {
	UserID int64 `json:"userId"`
	RoleID int64 `json:"roleId"`
	jwt.RegisteredClaims
}

// GenerateJWT genera un nuevo token JWT para un usuario.
func GenerateJWT(userID int64, roleID int64, secretKey []byte, expirationTime time.Duration) (string, error) {
	expiration := time.Now().Add(expirationTime)
	claims := &Claims{
		UserID: userID,
		RoleID: roleID, // Corregido: Usar int64 directamente
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiration),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "backend-connect",         // Puedes cambiar el emisor
			Subject:   fmt.Sprintf("%d", userID), // El sujeto suele ser el ID del usuario
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", fmt.Errorf("error signing token: %w", err)
	}

	return tokenString, nil
}

// ValidateJWT valida un token JWT y devuelve los claims si es válido.
func ValidateJWT(tokenString string, secretKey []byte) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// Asegurarse de que el método de firma es el esperado
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return secretKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("error parsing token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}
