package auth

import (
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// ContextKey es un tipo para usar como clave en el contexto de la petición
type ContextKey string

// UserIDKey es la clave para almacenar el UserID en el contexto
const UserIDKey ContextKey = "userID"

// ClaimsKey es la clave para almacenar el objeto de claims completo en el contexto.
const ClaimsKey ContextKey = "claims"

// Claims define la estructura de los claims del JWT
type Claims struct {
	UserID int64 `json:"userId"`
	RoleID int64 `json:"roleId"`
	jwt.RegisteredClaims
	ID string `json:"jti"` // "jti" (JWT ID) claim; ver RFC 7519, sección 4.1.7
}

// GenerateJWT genera un nuevo token JWT para un usuario.
func GenerateJWT(userID int64, roleID int64, secretKey []byte, expirationTime time.Duration) (string, int, error) {
	expiration := time.Now().Add(expirationTime)
	tokenID := 1 // Usar 1 como ID de token por defecto.
	claims := &Claims{
		UserID: userID,
		RoleID: roleID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiration),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "backend-connect",         // Puedes cambiar el emisor
			Subject:   fmt.Sprintf("%d", userID), // El sujeto suele ser el ID del usuario
			ID:        strconv.Itoa(tokenID),     // Asignar el ID al claim "jti" como string
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", 0, fmt.Errorf("error signing token: %w", err)
	}

	return tokenString, tokenID, nil
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
