package middlewares

import (
	"fmt"
	"mgo-gin/app/model"
	"net/http"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

// Create the JWT key used to create the signature
var jwtKey = []byte("uit_secret_key")

type Claims struct {
	Id       string `json:"_id"`
	Username string `json:"username"`
	Roles    string `json:"roles"`
	Email    string `json:"email"`

	jwt.StandardClaims
}

func GenerateJWTToken(user model.ResponseUser) string {
	claims := jwt.MapClaims{
		"_id":      user.Id.Hex(),
		"username": user.Username,
		"password": user.Password,
		"roles":    user.Roles,
		"email":    user.Email,
		"aud":      "user",
		"iss":      "uit",
		"exp":      time.Now().Add(time.Hour * 24 * 30).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, _ := token.SignedString([]byte("uit_secret_key"))
	return signedToken
}

func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "required.authorization"})
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "El formato del encabezado de autorización debe ser Bearer {token}"})
			c.Abort()
			return
		}

		tokenString := parts[1]
		claims := &Claims{} // Usamos nuestra struct Claims

		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			// Valida el algoritmo de firma
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("método de firma inesperado: %v", token.Header["alg"])
			}
			return jwtKey, nil
		})

		if err != nil {
			if err == jwt.ErrSignatureInvalid {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Firma de token inválida"})
				c.Abort()
				return
			}
			c.JSON(http.StatusBadRequest, gin.H{"error": "Token inválido: " + err.Error()}) // Podría ser un token malformado o expirado
			c.Abort()
			return
		}

		if !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token inválido"})
			c.Abort()
			return
		}

		// Poner las claims del usuario en el contexto para uso posterior en los handlers
		c.Set("user_id", claims.Id) // ID es string (hex)
		c.Set("username", claims.Username)
		c.Set("email", claims.Email) // Asegúrate que 'email' esté en tus claims
		c.Set("roles", claims.Roles)
		// También podrías guardar el struct Claims completo:
		// c.Set("user_claims", claims)

		c.Next()
	}
}
