package utils

import (
	"fmt"
	"strings"

	"github.com/dgrijalva/jwt-go"
)

type TokenClaims struct {
    ID       string `json:"_id"`
    Username string `json:"username"`
    Password string `json:"password"`
}

func ValidateToken(tokenString string) (*TokenClaims, error) {
    if tokenString == "" {
        return nil, fmt.Errorf("authorization token is required")
    }

    tokenParts := strings.Split(tokenString, "Bearer ")
    if len(tokenParts) != 2 {
        return nil, fmt.Errorf("invalid token format")
    }

    token, err := jwt.Parse(tokenParts[1], func(token *jwt.Token) (interface{}, error) {
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
        }
        return []byte("uit_secret_key"), nil
    })

    if err != nil {
        return nil, err
    }

    if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
        userClaims := &TokenClaims{
            ID:       claims["_id"].(string),
            Username: claims["username"].(string),
            Password: claims["password"].(string),
        }
        return userClaims, nil
    }

    return nil, fmt.Errorf("invalid token")
}